# parallel-pattern-search

A high-performance multi-pattern text search engine implementing the Aho-Corasick algorithm across four computing paradigms: sequential baseline, OpenMP (shared-memory), MPI (distributed-memory), and PFAC (Parallel Failureless Aho-Corasick) on GPU via OpenCL. The project targets large text files and produces throughput benchmarks (GB/s) with scaling analysis and a structured CPU vs GPU comparison report.

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
| `aho_omp` | OpenMP — one chunk per thread | Overlap region of size `max_pattern_len - 1` at each chunk boundary |
| `aho_mpi` | MPI — one block per process | Non-blocking `MPI_Isend`/`MPI_Irecv` tail exchange between adjacent ranks |
| `pfac_ocl` | OpenCL — one work-item per text offset | Inherent in PFAC: each offset is independent |

---

## Tech Stack

| Category | Tools |
|---|---|
| Language | C/C++ |
| CPU Parallelism | OpenMP, MPI (OpenMPI) |
| GPU Parallelism | OpenCL (Apple M4 Pro) |
| Build System | CMake |
| Benchmarking | Custom harness (`scripts/benchmark.sh`), `std::chrono` |
| Visualization | Python 3, matplotlib |
| CI/CD | GitHub Actions |

---

## Repository Structure

```
.
├── src/
│   ├── aho_seq.cpp          # Sequential baseline
│   ├── aho_omp.cpp          # OpenMP implementation
│   ├── aho_mpi.cpp          # MPI implementation
│   └── pfac_ocl.cpp         # OpenCL host code
├── kernels/
│   └── pfac.cl              # PFAC OpenCL kernel
├── include/
│   └── automaton.h          # Shared automaton data structures
├── scripts/
│   ├── gen_text.py          # Synthetic text file generator
│   ├── gen_patterns.py      # Pattern set generator
│   ├── benchmark.sh         # Unified benchmarking driver
│   ├── parse_results.py     # (unused at runtime — benchmark helper)
│   └── plot_results.py      # matplotlib plot generator
├── benchmarks/
│   ├── results/             # CSV output from benchmark runs
│   └── plots/               # PNG + PDF comparison plots
├── config/                  # (reserved for future config files)
├── tests/
│   └── fixtures/            # Sample ARF/input fixtures for unit tests
├── cmake/                   # CMake presets
├── CMakeLists.txt
├── .github/
│   └── workflows/
│       └── build.yml        # CI: build all targets
└── docs/
    ├── build.md
    ├── usage.md
    ├── report.md
    └── opencl-optimization.md
```

---

## Prerequisites

- macOS with Apple Silicon (M-series) — required for OpenCL target
- [CMake](https://cmake.org/) >= 3.20
- [OpenMPI](https://www.open-mpi.org/) — `brew install open-mpi`
- Clang with OpenMP support — `brew install libomp`
- Python >= 3.9 — `pip install matplotlib`
- OpenCL — provided by Apple's system frameworks (no install needed)

---

## Build

```bash
# Configure (release mode)
cmake --preset release

# Build all targets
cmake --build build/release

# Targets produced:
#   build/release/aho_seq
#   build/release/aho_omp
#   build/release/aho_mpi
#   build/release/pfac_ocl
```

See `docs/build.md` for detailed build instructions and troubleshooting.

---

## Quick Start

```bash
# Generate benchmark data (requires Python)
make data

# Run a single search (sequential)
./build/release/aho_seq --patterns data/patterns_100.txt --text data/text_100mb.txt

# Run with OpenMP (10 threads)
./build/release/aho_omp --patterns data/patterns_100.txt --text data/text_100mb.txt --threads 10

# Run with MPI (8 processes)
mpirun -np 8 ./build/release/aho_mpi --patterns data/patterns_100.txt --text data/text_100mb.txt

# Run PFAC on GPU
./build/release/pfac_ocl --patterns data/patterns_100.txt --text data/text_100mb.txt

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
- **Thread/process scaling** (OpenMP + MPI): 1 / 2 / 4 / 8 / 10 cores

Each configuration runs 5 times; mean ± stddev is recorded to `benchmarks/results/`.

Results are visualized as four plots in `benchmarks/plots/`:
1. Peak throughput (GB/s) per implementation — bar chart
2. Speedup vs thread/process count — line chart
3. Throughput vs pattern count
4. Throughput vs text size

---

## Milestones

| # | Milestone | Description |
|---|---|---|
| 1 | Foundation & Build System | CMake setup, sequential Aho-Corasick baseline, test data generation |
| 2 | CPU Parallelism — OpenMP | Thread-level chunking, parallel search, boundary handling |
| 3 | CPU Parallelism — MPI | Process-level distribution, inter-process boundary exchange |
| 4 | GPU Parallelism — OpenCL/PFAC | PFAC kernel, automaton memory layout, M4 Pro optimization |
| 5 | Benchmarking & Report | Throughput harness, scaling experiments, plots, final report |

---

## License

MIT
