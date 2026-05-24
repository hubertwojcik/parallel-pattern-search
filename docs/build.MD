# Build Instructions

## Prerequisites

- macOS with Apple Silicon (M-series)
- Go >= 1.22 — [golang.org](https://go.dev/)
- Python >= 3.9 — do generatorów danych testowych

## Build

```bash
# Tylko sequential baseline (Milestone 1)
make seq

# Wszystkie targety (po ukończeniu wszystkich milestoneów)
make all
```

## Dane testowe

```bash
make data
```

Generuje pliki w `data/`:

- teksty: 10 MB, 100 MB, 500 MB, 1 GB
- wzorce: 10, 100, 500, 1000

## Uruchomienie

```bash
./aho_seq --patterns data/patterns_100.txt --text data/text_100mb.txt
```

## Testy jednostkowe

```bash
go test ./...
```
