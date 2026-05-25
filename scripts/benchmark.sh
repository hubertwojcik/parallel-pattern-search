#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
RESULTS="$ROOT/benchmarks/results"
DATA="$ROOT/data"
RUNS=3

mkdir -p "$RESULTS" "$DATA"

run() {
    local bin="$1"; shift
    local best=0
    for _ in $(seq 1 "$RUNS"); do
        val=$("$ROOT/$bin" "$@" 2>/dev/null | grep throughput | awk '{print $2}')
        val=${val:-0}
        best=$(awk "BEGIN{print ($val > $best) ? $val : $best}")
    done
    echo "$best"
}

run_mpi() {
    local np="$1"; shift
    local best=0
    for _ in $(seq 1 "$RUNS"); do
        val=$(mpirun -np "$np" "$ROOT/aho_mpi" "$@" 2>/dev/null | grep throughput | awk '{print $2}')
        val=${val:-0}
        best=$(awk "BEGIN{print ($val > $best) ? $val : $best}")
    done
    echo "$best"
}

echo "=== Generowanie danych ==="
for size in 10mb 100mb; do
    f="$DATA/text_${size}.txt"
    [[ -f "$f" ]] || python3 "$ROOT/scripts/gen_text.py" --size "$size" --out "$f"
done
for n in 10 100 500 1000; do
    f="$DATA/patterns_${n}.txt"
    [[ -f "$f" ]] || python3 "$ROOT/scripts/gen_patterns.py" --count "$n" --out "$f"
done

echo "=== Skalowanie wątków/procesów (10MB, 100 wzorców) ==="
OUT="$RESULTS/scaling_workers.csv"
echo "impl,workers,text_mb,patterns,throughput_gbs" > "$OUT"

t=$(run aho_seq --patterns "$DATA/patterns_100.txt" --text "$DATA/text_10mb.txt")
echo "seq,1,10,100,$t" >> "$OUT"

for w in 1 2 4 8; do
    t=$(run aho_goroutines --patterns "$DATA/patterns_100.txt" --text "$DATA/text_10mb.txt" --workers "$w")
    echo "goroutines,$w,10,100,$t" >> "$OUT"
done

for np in 1 2 4; do
    t=$(run_mpi "$np" --patterns "$DATA/patterns_100.txt" --text "$DATA/text_10mb.txt")
    echo "mpi,$np,10,100,$t" >> "$OUT"
done

t=$(run pfac_ocl --patterns "$DATA/patterns_100.txt" --text "$DATA/text_10mb.txt")
echo "pfac,1,10,100,$t" >> "$OUT"

echo "=== Skalowanie rozmiaru tekstu (100 wzorców, 4 wątki) ==="
OUT="$RESULTS/scaling_text.csv"
echo "impl,workers,text_mb,patterns,throughput_gbs" > "$OUT"

for size_mb in 10 100; do
    txt="$DATA/text_${size_mb}mb.txt"
    t=$(run aho_seq --patterns "$DATA/patterns_100.txt" --text "$txt")
    echo "seq,1,$size_mb,100,$t" >> "$OUT"

    t=$(run aho_goroutines --patterns "$DATA/patterns_100.txt" --text "$txt" --workers 4)
    echo "goroutines,4,$size_mb,100,$t" >> "$OUT"

    t=$(run_mpi 4 --patterns "$DATA/patterns_100.txt" --text "$txt")
    echo "mpi,4,$size_mb,100,$t" >> "$OUT"

    t=$(run pfac_ocl --patterns "$DATA/patterns_100.txt" --text "$txt")
    echo "pfac,1,$size_mb,100,$t" >> "$OUT"
done

echo "=== Skalowanie liczby wzorców (10MB, 4 wątki) ==="
OUT="$RESULTS/scaling_patterns.csv"
echo "impl,workers,text_mb,patterns,throughput_gbs" > "$OUT"

for n in 10 100 500 1000; do
    pat="$DATA/patterns_${n}.txt"
    t=$(run aho_seq --patterns "$pat" --text "$DATA/text_10mb.txt")
    echo "seq,1,10,$n,$t" >> "$OUT"

    t=$(run aho_goroutines --patterns "$pat" --text "$DATA/text_10mb.txt" --workers 4)
    echo "goroutines,4,10,$n,$t" >> "$OUT"

    t=$(run_mpi 4 --patterns "$pat" --text "$DATA/text_10mb.txt")
    echo "mpi,4,10,$n,$t" >> "$OUT"

    t=$(run pfac_ocl --patterns "$pat" --text "$DATA/text_10mb.txt")
    echo "pfac,1,10,$n,$t" >> "$OUT"
done

echo "=== Gotowe. Wyniki w $RESULTS/ ==="
