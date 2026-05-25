package automaton

import (
	"sort"
	"testing"
)

func sortMatches(matches []Match) {
	sort.Slice(matches, func(i, j int) bool {
		if matches[i].Position != matches[j].Position {
			return matches[i].Position < matches[j].Position
		}
		return matches[i].PatternID < matches[j].PatternID
	})
}

func TestSinglePattern(t *testing.T) {
	a := Build([]string{"he"})
	matches := a.Search([]byte("he"))

	if len(matches) != 1 {
		t.Fatalf("oczekiwano 1 dopasowania, got %d", len(matches))
	}
	if matches[0].PatternID != 0 {
		t.Errorf("oczekiwano PatternID=0, got %d", matches[0].PatternID)
	}
	if matches[0].Position != 1 {
		t.Errorf("oczekiwano Pos=1, got %d", matches[0].Position)
	}
}

func TestMultiplePatterns(t *testing.T) {
	patterns := []string{"he", "she", "his", "hers"}
	a := Build(patterns)
	matches := a.Search([]byte("ushers"))
	sortMatches(matches)

	if len(matches) != 3 {
		t.Fatalf("oczekiwano 3 dopasowań, got %d: %v", len(matches), matches)
	}
}

func TestNoMatch(t *testing.T) {
	a := Build([]string{"xyz"})
	matches := a.Search([]byte("abcdef"))

	if len(matches) != 0 {
		t.Errorf("oczekiwano 0 dopasowań, got %d", len(matches))
	}
}

func TestOverlapping(t *testing.T) {
	a := Build([]string{"aa"})
	matches := a.Search([]byte("aaa"))

	if len(matches) != 2 {
		t.Fatalf("oczekiwano 2 dopasowań, got %d", len(matches))
	}
}

func TestEmptyText(t *testing.T) {
	a := Build([]string{"abc"})
	matches := a.Search([]byte(""))

	if len(matches) != 0 {
		t.Errorf("oczekiwano 0 dopasowań dla pustego tekstu, got %d", len(matches))
	}
}

func TestPatternAtEnd(t *testing.T) {
	a := Build([]string{"end"})
	matches := a.Search([]byte("the end"))

	if len(matches) != 1 {
		t.Fatalf("oczekiwano 1 dopasowania, got %d", len(matches))
	}
	if matches[0].Position != 6 {
		t.Errorf("oczekiwano Pos=6, got %d", matches[0].Position)
	}
}

func TestCaseSensitive(t *testing.T) {
	a := Build([]string{"ABC"})
	matches := a.Search([]byte("abc ABC"))

	if len(matches) != 1 {
		t.Fatalf("oczekiwano 1 dopasowania (tylko 'ABC'), got %d", len(matches))
	}
}
