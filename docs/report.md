# Sprawozdanie — Równoległe wyszukiwanie wzorców w tekście (Aho–Corasick / PFAC)

**Autor:** Hubert Wójcik  
**Data:** maj 2026  
**Język implementacji:** Go 1.25

---

## 1. Podział pracy

Projekt realizowany jednoosobowo.

| Zadanie | Autor |
|---|---|
| Algorytm sekwencyjny Aho-Corasick | Hubert Wójcik |
| Implementacja OpenMP (goroutines) | Hubert Wójcik |
| Implementacja MPI (go-mpi) | Hubert Wójcik |
| Kernel PFAC / OpenCL (GPU) | Hubert Wójcik |
| Skrypty benchmarkowe i wykresy | Hubert Wójcik |
| Dokumentacja | Hubert Wójcik |

---

## 2. Konfiguracja sprzętowa

Wszystkie testy przeprowadzono na jednej maszynie:

| Parametr | Wartość |
|---|---|
| Model | Apple MacBook Pro (M4 Pro) |
| CPU | Apple M4 Pro — 12 rdzeni (10 wydajnych + 2 energooszczędne) |
| GPU | Apple M4 Pro GPU — 20 rdzeni (unified memory) |
| RAM | 24 GB (unified memory — wspólna dla CPU i GPU) |
| System operacyjny | macOS 26.4 (Darwin 25.4.0) |
| OpenCL | Apple OpenCL framework (wbudowany w macOS) |
| OpenMPI | 5.0.9 (Homebrew) |
| Go | 1.25.5 |

Architektura unified memory oznacza, że CPU i GPU współdzielą tę samą pamięć fizyczną. Transfer danych CPU→GPU i GPU→CPU nie wymaga kopiowania przez szynę PCIe — to istotna zaleta dla PFAC, gdzie cały tekst musi trafić na GPU.

---

## 3. Opis algorytmów i implementacji

### 3.1 Algorytm Aho-Corasick (sekwencyjny i CPU-równoległy)

Aho-Corasick buduje automat skończony offline z zestawu wzorców. Składa się z trzech faz:

1. **Budowanie drzewa trie** — każdy wzorzec wstawiany jako ścieżka od korzenia
2. **Obliczenie dowiązań awaryjnych (failure links)** — BFS po stanach; gdy automat nie może kontynuować, przeskakuje do najdłuższego sufiksu który jest prefiksem jakiegoś wzorca
3. **Wyszukiwanie** — jeden liniowy przebieg po tekście, O(n) niezależnie od liczby wzorców

Kluczowy fragment — obliczanie failure links:

```go
for len(queue) > 0 {
    r := queue[0]
    queue = queue[1:]
    for c := 0; c < AlphabetSize; c++ {
        s := a.Goto[r][c]
        if s == -1 {
            a.Goto[r][c] = a.Goto[a.Fail[r]][c]  // uzupełnienie goto przez failure
            continue
        }
        a.Fail[s] = a.Goto[a.Fail[r]][c]
        a.Output[s] = append(a.Output[s], a.Output[a.Fail[s]]...)
        queue = append(queue, s)
    }
}
```

### 3.2 Implementacja OpenMP (goroutines)

Tekst dzielony jest na `W` równych fragmentów. Każda goroutine przetwarza swój fragment niezależnie. Problem graniczny: wzorzec może zaczynać się pod koniec jednego fragmentu i kończyć w kolejnym. Rozwiązanie — każdy fragment jest przedłużony o `maxPatternLen-1` bajtów za swoją granicę (obszar overlap), a po przeszukaniu filtrowane są dopasowania których **pozycja startowa** leży w obrębie fragmentu:

```go
matchStart := absPos - a.PatternLens[m.PatternID] + 1
if matchStart < end {
    adjusted = append(adjusted, ...)
}
```

Użycie `absPos < end` (pozycja końcowa) byłoby błędem — gubiłoby dopasowania kończące się w obszarze overlap.

### 3.3 Implementacja MPI

Każdy proces MPI czyta cały plik tekstowy lokalnie i przetwarza swój blok `[rank*chunkSize, (rank+1)*chunkSize)` z takim samym mechanizmem overlap jak w goroutines. Na końcu wszystkie procesy redukują liczniki dopasowań do procesu 0:

```go
localSlice := []int64{int64(localCount)}
totalSlice := []int64{0}
comm.ReduceInt64s(totalSlice, localSlice, mpi.OpSum, 0)
```

Brak komunikacji podczas wyszukiwania (tylko jeden `Reduce` na końcu) daje doskonałą skalowalność.

### 3.4 PFAC / OpenCL (GPU)

PFAC (Parallel Failureless Aho-Corasick) eliminuje failure links. Automat jest budowany jako samo drzewo trie — brakujące przejścia dają -1 (martwy stan), nie pętlę do korzenia jak w AC.

Każdy wątek GPU (`work-item`) startuje od innego offsetu tekstu i samodzielnie przesuwa się do przodu po trie:

```c
__kernel void pfac(
    __global const uchar *text, const int text_len,
    __global const int *goto_table, __global const int *has_output,
    const int alphabet_size, __global volatile int *total_matches
) {
    int gid = get_global_id(0);
    if (gid >= text_len) return;
    int state = 0;
    for (int pos = gid; pos < text_len; pos++) {
        int next = goto_table[state * alphabet_size + (int)text[pos]];
        if (next < 0) break;
        state = next;
        if (has_output[state]) atomic_inc(total_matches);
    }
}
```

Brak synchronizacji między wątkami — każdy wątek zatrzymuje się niezależnie przy pierwszym nierozpoznanym znaku. Dzięki temu GPU może uruchomić miliony wątków równocześnie (jeden na każdy bajt tekstu).

---

## 4. Testy wydajnościowe

Każdy pomiar to najlepszy wynik z 3 powtórzeń. Tekst: losowe słowa ASCII, wzorce: podzbiory słownika.

### 4.1 Skalowanie liczby wątków / procesów

Tekst: 10 MB, wzorce: 100, platforma: Apple M4 Pro.

#### Goroutines

| Wątki | Przepustowość (GB/s) | Przyspieszenie S(p) | Efektywność E(p) |
|---|---|---|---|
| 1 | 0.093 | 1.00 | 1.00 |
| 2 | 0.165 | 1.77 | 0.89 |
| 4 | 0.242 | 2.60 | 0.65 |
| 8 | 0.321 | 3.45 | 0.43 |

#### MPI

| Procesy | Przepustowość (GB/s) | Przyspieszenie S(p) | Efektywność E(p) |
|---|---|---|---|
| 1 | 0.117 | 1.00 | 1.00 |
| 2 | 0.234 | 2.00 | 1.00 |
| 4 | 0.474 | 4.05 | 1.01 |

MPI osiąga niemal liniowe przyspieszenie — każdy proces pracuje na niezależnym fragmencie bez dzielenia stanu, a jedyna komunikacja to jeden `Reduce` po zakończeniu wyszukiwania. Efektywność >1 przy 4 procesach wynika z efektu cache: mniejszy fragment lepiej mieści się w L2/L3.

![Przyspieszenie](../benchmarks/plots/speedup.png)

![Efektywność](../benchmarks/plots/efficiency.png)

### 4.2 Porównanie najlepszych wyników

Tekst: 100 MB, wzorce: 100.

| Implementacja | Konfiguracja | Przepustowość (GB/s) | Przyspieszenie vs seq |
|---|---|---|---|
| Sekwencyjny | 1 wątek | 0.124 | 1.0× |
| Goroutines | 4 wątki | 0.292 | 2.4× |
| MPI | 4 procesy | 0.477 | 3.8× |
| PFAC/OpenCL | GPU | 4.241 | **34.2×** |

![Porównanie najlepszych wyników](../benchmarks/plots/comparison.png)

### 4.3 Skalowanie rozmiaru tekstu

Przepustowość rośnie wraz z rozmiarem tekstu dla GPU — GPU amortyzuje koszt transferu danych i kompilacji kernela. CPU pozostaje stabilne.

![Skalowanie rozmiaru tekstu](../benchmarks/plots/scaling_text.png)

### 4.4 Skalowanie liczby wzorców

Wszystkie implementacje tracą przepustowość przy rosnącej liczbie wzorców — większy automat (więcej stanów) gorzej korzysta z cache CPU. GPU jest bardziej odporne na ten efekt dzięki równoległemu dostępowi do pamięci.

| Wzorce | seq (GB/s) | goroutines 4× (GB/s) | MPI 4× (GB/s) | PFAC GPU (GB/s) |
|---|---|---|---|---|
| 10 | 0.270 | 0.730 | 1.028 | 1.569 |
| 100 | 0.131 | 0.261 | 0.501 | 1.382 |
| 500 | 0.091 | 0.173 | 0.342 | 1.152 |
| 1000 | 0.077 | 0.176 | 0.290 | 1.115 |

![Skalowanie liczby wzorców](../benchmarks/plots/scaling_patterns.png)

---

## 5. Wnioski

1. **MPI** daje najlepsze przyspieszenie spośród implementacji CPU — niemal liniowe (4.05× przy 4 procesach) dzięki braku komunikacji podczas wyszukiwania i efektowi lepszego cache.

2. **Goroutines** skaluje się gorzej niż MPI — współdzielona pamięć oznacza rywalizację o cache L3 przy odczycie tekstu przez wiele wątków.

3. **PFAC/OpenCL** dominuje przepustowością (4.241 GB/s na 100 MB, 34× szybciej niż seq). Kluczowa jest unified memory architektury M4 Pro — transfer CPU→GPU jest szybki i nie ogranicza wyników.

4. GPU szczególnie zyskuje na **większych plikach** — koszt inicjalizacji (kompilacja kernela, setup OpenCL) amortyzuje się przy dużych tekstach.

5. **Liczba wzorców** bardziej wpływa na CPU niż GPU — automat rośnie, ale wątki GPU i tak pracują równolegle; GPU degeneruje łagodniej.

---

## 6. Instrukcja obsługi

### Wymagania

- Go ≥ 1.22
- OpenMPI: `brew install open-mpi`
- OpenCL: wbudowany w macOS (brak dodatkowej instalacji)
- Python ≥ 3.9 + matplotlib: `pip install matplotlib`

### Budowanie

```bash
make all          # buduje aho_seq, aho_omp, aho_mpi
make pfac         # buduje pfac_ocl (wymaga OpenCL)
```

### Generowanie danych testowych

```bash
make data
# lub ręcznie:
python3 scripts/gen_text.py --size 100mb --out data/text_100mb.txt
python3 scripts/gen_patterns.py --count 100 --out data/patterns_100.txt
```

### Uruchamianie

```bash
# Sekwencyjny
./aho_seq --patterns data/patterns_100.txt --text data/text_100mb.txt

# Goroutines (OpenMP ekwiwalent)
./aho_omp --patterns data/patterns_100.txt --text data/text_100mb.txt --workers 8

# MPI
mpirun -np 4 ./aho_mpi --patterns data/patterns_100.txt --text data/text_100mb.txt

# PFAC GPU
./pfac_ocl --patterns data/patterns_100.txt --text data/text_100mb.txt
```

### Benchmarki i wykresy

```bash
bash scripts/benchmark.sh        # zapisuje CSV do benchmarks/results/
python3 scripts/plot_results.py  # generuje PNG do benchmarks/plots/
```

---

## 7. Testy jednostkowe

```bash
go test ./...
# lub z MPI:
CGO_CFLAGS="-I/opt/homebrew/include" CGO_LDFLAGS="-L/opt/homebrew/lib" go test ./...
```

Testy obejmują: budowanie automatu AC i PFAC, poprawność wyszukiwania sekwencyjnego, obsługę granic fragmentów w goroutines i MPI, brak podwójnego liczenia dopasowań na granicach.
