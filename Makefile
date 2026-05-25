# Makefile — buduje wszystkie targety i zarządza danymi testowymi.
# Używamy go build zamiast CMake — Go nie potrzebuje systemu budowania.

.PHONY: all clean data bench seq goroutines mpi

# CGo musi widzieć nagłówki OpenMPI — na Apple Silicon Homebrew instaluje do /opt/homebrew
CGO_FLAGS = CGO_CFLAGS="-I/opt/homebrew/include" CGO_LDFLAGS="-L/opt/homebrew/lib"

# Buduje wszystkie cztery implementacje
all:
	go build -o aho_seq    ./cmd/aho_seq
	go build -o aho_omp    ./cmd/aho_goroutines
	$(CGO_FLAGS) go build -o aho_mpi    ./cmd/aho_mpi
	go build -o pfac_ocl   ./cmd/pfac_ocl

# Tylko baseline (przydatne na wczesnym etapie)
seq:
	go build -o aho_seq ./cmd/aho_seq

goroutines:
	go build -o aho_omp ./cmd/aho_goroutines

mpi:
	$(CGO_FLAGS) go build -o aho_mpi ./cmd/aho_mpi

pfac:
	go build -tags opencl -o pfac_ocl ./cmd/pfac_ocl

# Generuje dane testowe we wszystkich rozmiarach używanych w benchmarkach
data:
	mkdir -p data
	python3 scripts/gen_text.py     --size 10mb  --out data/text_10mb.txt
	python3 scripts/gen_text.py     --size 100mb --out data/text_100mb.txt
	python3 scripts/gen_text.py     --size 500mb --out data/text_500mb.txt
	python3 scripts/gen_text.py     --size 1gb   --out data/text_1gb.txt
	python3 scripts/gen_patterns.py --count 10   --out data/patterns_10.txt
	python3 scripts/gen_patterns.py --count 100  --out data/patterns_100.txt
	python3 scripts/gen_patterns.py --count 500  --out data/patterns_500.txt
	python3 scripts/gen_patterns.py --count 1000 --out data/patterns_1000.txt

# Usuwa binaria (nie dane — generowanie 1gb zajmuje czas)
clean:
	rm -f aho_seq aho_omp aho_mpi pfac_ocl