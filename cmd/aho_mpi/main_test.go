package main

import (
	"sort"
	"testing"

	"github.com/hubertwojcik/parallel-pattern-search/internal/automaton"
)

var (
	testPatterns = []string{"he", "she", "his", "hers"}
	testText     = []byte("ushers and she said his hers")
)

func seqCount(a *automaton.Automaton, text []byte) int {
	return len(a.Search(text))
}

func overlap(a *automaton.Automaton) int {
	max := 0
	for _, l := range a.PatternLens {
		if l > max {
			max = l
		}
	}
	if max == 0 {
		return 0
	}
	return max - 1
}

// TestChunksSumToSeq sprawdza że suma po wszystkich rankach == wynik sekwencyjny.
func TestChunksSumToSeq(t *testing.T) {
	a := automaton.Build(testPatterns)
	ov := overlap(a)
	want := seqCount(a, testText)

	for _, size := range []int{1, 2, 3, 4, 7} {
		total := 0
		for rank := 0; rank < size; rank++ {
			total += searchChunk(a, testText, rank, size, ov)
		}
		if total != want {
			t.Errorf("size=%d: got %d matches, want %d", size, total, want)
		}
	}
}

// TestBoundaryMatch sprawdza wzorzec przecinający granicę między procesami.
func TestBoundaryMatch(t *testing.T) {
	// "hers" zaczyna się w pierwszym chunka, kończy w drugim
	text := []byte("his hers end")
	a := automaton.Build(testPatterns)
	ov := overlap(a)
	want := seqCount(a, text)

	total := 0
	for rank := 0; rank < 2; rank++ {
		total += searchChunk(a, text, rank, 2, ov)
	}
	if total != want {
		t.Errorf("boundary: got %d, want %d", total, want)
	}
}

// TestNoDoubleCounting sprawdza że wzorzec na granicy nie jest liczony dwa razy.
func TestNoDoubleCounting(t *testing.T) {
	a := automaton.Build(testPatterns)
	ov := overlap(a)

	for _, size := range []int{2, 3, 4} {
		counts := make([]int, size)
		for rank := 0; rank < size; rank++ {
			counts[rank] = searchChunk(a, testText, rank, size, ov)
		}
		total := 0
		for _, c := range counts {
			total += c
		}
		want := seqCount(a, testText)
		if total != want {
			t.Errorf("double-count check size=%d: got %d, want %d, per-rank=%v",
				size, total, want, counts)
		}
	}
}

// TestSingleProcess sprawdza że jeden proces = wynik sekwencyjny.
func TestSingleProcess(t *testing.T) {
	a := automaton.Build(testPatterns)
	got := searchChunk(a, testText, 0, 1, overlap(a))
	want := seqCount(a, testText)
	if got != want {
		t.Errorf("single process: got %d, want %d", got, want)
	}
}

// TestDeterministic sprawdza że wyniki są powtarzalne.
func TestDeterministic(t *testing.T) {
	a := automaton.Build(testPatterns)
	ov := overlap(a)

	counts1 := make([]int, 4)
	counts2 := make([]int, 4)
	for rank := 0; rank < 4; rank++ {
		counts1[rank] = searchChunk(a, testText, rank, 4, ov)
		counts2[rank] = searchChunk(a, testText, rank, 4, ov)
	}
	if !sort.IntsAreSorted(append(counts1, counts2...)) {
		// tylko sprawdzamy że są identyczne
	}
	for i := range counts1 {
		if counts1[i] != counts2[i] {
			t.Errorf("rank %d: niepowtarzalny wynik %d vs %d", i, counts1[i], counts2[i])
		}
	}
}
