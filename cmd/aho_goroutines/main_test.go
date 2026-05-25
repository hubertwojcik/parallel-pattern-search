package main

import (
	"sort"
	"testing"

	"github.com/hubertwojcik/parallel-pattern-search/internal/automaton"
)

func sortMatches(ms []automaton.Match) {
	sort.Slice(ms, func(i, j int) bool {
		if ms[i].Position != ms[j].Position {
			return ms[i].Position < ms[j].Position
		}
		return ms[i].PatternID < ms[j].PatternID
	})
}

var (
	testPatterns = []string{"he", "she", "his", "hers"}
	testText     = []byte("ushers and she said his hers")
)

func seqMatches(a *automaton.Automaton, text []byte) []automaton.Match {
	ms := a.Search(text)
	sortMatches(ms)
	return ms
}

func parMatches(a *automaton.Automaton, text []byte, workers int) []automaton.Match {
	maxLen := 0
	for _, l := range a.PatternLens {
		if l > maxLen {
			maxLen = l
		}
	}
	overlap := maxLen - 1
	if overlap < 0 {
		overlap = 0
	}
	ms := parallelSearch(a, text, workers, overlap)
	sortMatches(ms)
	return ms
}

func TestMatchesEqualSeq(t *testing.T) {
	a := automaton.Build(testPatterns)
	want := seqMatches(a, testText)

	for _, w := range []int{2, 3, 4, 7} {
		if w > len(testText) {
			continue
		}
		got := parMatches(a, testText, w)
		if len(got) != len(want) {
			t.Errorf("workers=%d: got %d matches, want %d", w, len(got), len(want))
			continue
		}
		for i := range want {
			if got[i] != want[i] {
				t.Errorf("workers=%d: match[%d] got %+v, want %+v", w, i, got[i], want[i])
			}
		}
	}
}

func TestBoundaryMatch(t *testing.T) {
	text := []byte("his hers ok")
	a := automaton.Build(testPatterns)

	want := seqMatches(a, text)
	got := parMatches(a, text, 2)

	if len(got) != len(want) {
		t.Errorf("boundary: got %d matches, want %d\ngot:  %v\nwant: %v", len(got), len(want), got, want)
	}
}
