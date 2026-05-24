# parallel-pattern-search

A high-performance multi-pattern text search engine implementing the Aho-Corasick algorithm across four computing paradigms: sequential baseline, goroutines (shared-memory), MPI (distributed-memory), and PFAC (Parallel Failureless Aho-Corasick) on GPU via OpenCL. The project targets large text files and produces throughput benchmarks (GB/s) with scaling analysis and a structured CPU vs GPU comparison report.

Developed and benchmarked on Apple M4 Pro using the unified memory architecture via OpenCL.

---

## Algorithm Overview

### Aho-Corasick (CPU)
Aho-Corasick builds a finite automaton from a set of patterns offline (trie + failure links), then scans the input text in a single O(n) pass — matching all patterns simultaneously regardless of pattern count.

```
Offline:  patterns → trie → failure links → flat transition table
Online:   text[0..n] → automaton traversal → (pattern_id, position) matches
```

### PFAC (GPU — OpenCL)
PFAC eliminates failure links entirely. Each GPU work-item starts independently from a different text offset and walks the automaton forward. No synchronization is needed between threads — trivially parallel.

```
work-item i:  text[i], text[i+1], text[i+2], ... → match or dead state
```

---

## Implementations

| Target | Parallelism | Boundary Handling |
|---|---|---|
| `aho_seq` | Sequential | — |
| `aho_goroutines` | Goroutines — one chunk per goroutine (OpenMP equiv) | Overlap region of size `max_pattern_len - 1` at each chunk boundary |
| `aho_mpi` | MPI — one block per process (go-mpi) | Tail exchange between adjacent ranks |
| `pfac_ocl` | OpenCL — one work-item per text offset | Inherent in PFAC: each offset is independent |

---

## Tech Stack

| Category | Tools |
|---|---|
| Language | Go |
| CPU Parallelism | goroutines + `sync`/`errgroup` (OpenMP equiv), `go-mpi` (MPI equiv) |
| GPU Parallelism | `go-opencl` + OpenCL kernel (Apple M4 Pro) |
| Build System | `go build` + Makefile |
| Benchmarking | Custom harness (`scripts/benchmark.sh`), `time` package |
| Visualization | Python 3, matplotlib |

---

## Repository Structure

```
.
├── cmd/
│   ├── aho_seq/
│   │   └── main.go          # Sequential baseline
│   ├── aho_goroutines/
│   │   └── main.go          # Goroutines implementation (OpenMP equiv)
│   ├── aho_mpi/
│   │   └── main.go          # MPI implementation (go-mpi)
│   └── pfac_ocl/
│       └── main.go          # OpenCL host code
├── internal/
│   └── automaton/
│       └── automaton.go     # Shared automaton data structures
├── kernels/
│   └── pfac.cl              # PFAC OpenCL kernel (OpenCL C)
├── scripts/
│   ├── gen_text.py          # Synthetic text file generator
│   ├── gen_patterns.py      # Pattern set generator
│   ├── benchmark.sh         # Unified benchmarking driver
│   └── plot_results.py      # matplotlib plot generator
├── benchmarks/
│   ├── results/             # CSV output from benchmark runs
│   └── plots/               # PNG + PDF comparison plots
├── tests/
│   └── fixtures/            # Sample input fixtures for unit tests
├── go.mod
├── go.sum
├── Makefile
└── docs/
    ├── build.md
    ├── usage.md
    └── report.md
```

---

## Prerequisites

- macOS with Apple Silicon (M-series) — required for OpenCL target
- [Go](https://go.dev/) >= 1.22
- [OpenMPI](https://www.open-mpi.org/) — `brew install open-mpi`
- Python >= 3.9 — `pip install matplotlib`
- OpenCL — provided by Apple's system frameworks (no install needed)

---

## Build

```bash
# Build all targets
make all

# Or individually
go build ./cmd/aho_seq
go build ./cmd/aho_goroutines
go build ./cmd/aho_mpi
go build ./cmd/pfac_ocl
```

---

## Quick Start

```bash
# Generate benchmark data (requires Python)
make data

# Run sequential search
./aho_seq --patterns data/patterns_100.txt --text data/text_100mb.txt

# Run with goroutines (10 goroutines)
./aho_goroutines --patterns data/patterns_100.txt --text data/text_100mb.txt --workers 10

# Run with MPI (8 processes)
mpirun -np 8 ./aho_mpi --patterns data/patterns_100.txt --text data/text_100mb.txt

# Run PFAC on GPU
./pfac_ocl --patterns data/patterns_100.txt --text data/text_100mb.txt

# Run full benchmark suite
make bench

# Generate comparison plots
python3 scripts/plot_results.py
```

---

## Benchmarking

The benchmark harness runs all four implementations under identical conditions across:

- **Pattern count scaling**: 10 / 100 / 500 / 1000 patterns, fixed 100 MB text
- **Text length scaling**: 10 MB / 100 MB / 500 MB / 1 GB, fixed 100 patterns
- **Worker/process scaling** (goroutines + MPI): 1 / 2 / 4 / 8 / 10

Each configuration runs 5 times; mean ± stddev is recorded to `benchmarks/results/`.

Results are visualized as four plots in `benchmarks/plots/`:
1. Peak throughput (GB/s) per implementation — bar chart
2. Speedup vs worker/process count — line chart
3. Throughput vs pattern count
4. Throughput vs text size

---

## Milestones

| # | Milestone | Description |
|---|---|---|
| 1 | Foundation & Build System | go.mod setup, sequential Aho-Corasick baseline, test data generation |
| 2 | CPU Parallelism — Goroutines | Goroutine-level chunking, parallel search, boundary handling |
| 3 | CPU Parallelism — MPI | Process-level distribution via go-mpi, boundary exchange |
| 4 | GPU Parallelism — OpenCL/PFAC | PFAC kernel, automaton memory layout, M4 Pro optimization |
| 5 | Benchmarking & Report | Throughput harness, scaling experiments, plots, final report |

---

## License

MIT
