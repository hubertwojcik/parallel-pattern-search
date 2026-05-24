package automaton

// Testy jednostkowe automatu Aho-Corasick.
// Używamy wbudowanego pakietu "testing" — zero zewnętrznych zależności.
// Każdy test sprawdza jeden konkretny przypadek, żeby łatwo zlokalizować błąd.

import (
	"sort"
	"testing"
)

// helper: sortuje matches żeby porównanie było deterministyczne
// (automat zwraca dopasowania w kolejności napotkania, ale testy
// nie powinny zależeć od kolejności gdy sprawdzamy zbiór)
func sortMatches(matches []Match) {
	sort.Slice(matches, func(i, j int) bool {
		if matches[i].Position != matches[j].Position {
			return matches[i].Position < matches[j].Position
		}
		return matches[i].PatternID < matches[j].PatternID
	})
}

// TestSinglePattern — najprostszy przypadek: jeden wzorzec, jedno dopasowanie
func TestSinglePattern(t *testing.T) {
	a := Build([]string{"he"})
	matches := a.Search([]byte("he"))

	if len(matches) != 1 {
		t.Fatalf("oczekiwano 1 dopasowania, got %d", len(matches))
	}
	if matches[0].PatternID != 0 {
		t.Errorf("oczekiwano PatternID=0, got %d", matches[0].PatternID)
	}
	// Pos to indeks ostatniego bajtu dopasowania (inclusive)
	if matches[0].Position != 1 {
		t.Errorf("oczekiwano Pos=1, got %d", matches[0].Position)
	}
}

// TestMultiplePatterns — klasyczny przykład z literatury Aho-Corasick:
// wzorce "he", "she", "his", "hers" w tekście "ushers"
func TestMultiplePatterns(t *testing.T) {
	patterns := []string{"he", "she", "his", "hers"}
	a := Build(patterns)
	matches := a.Search([]byte("ushers"))
	sortMatches(matches)

	// "ushers" zawiera: "she" (pos 3), "he" (pos 3), "hers" (pos 5)
	// Uwaga: "she" i "he" kończą się na tej samej pozycji (3)
	if len(matches) != 3 {
		t.Fatalf("oczekiwano 3 dopasowań, got %d: %v", len(matches), matches)
	}
}

// TestNoMatch — tekst który nie zawiera żadnego wzorca
func TestNoMatch(t *testing.T) {
	a := Build([]string{"xyz"})
	matches := a.Search([]byte("abcdef"))

	if len(matches) != 0 {
		t.Errorf("oczekiwano 0 dopasowań, got %d", len(matches))
	}
}

// TestOverlapping — wzorce nakładające się na siebie
func TestOverlapping(t *testing.T) {
	// "aa" w "aaa" powinno dać 2 dopasowania: na pos 1 i pos 2
	a := Build([]string{"aa"})
	matches := a.Search([]byte("aaa"))

	if len(matches) != 2 {
		t.Fatalf("oczekiwano 2 dopasowań, got %d", len(matches))
	}
}

// TestEmptyText — brzegowy przypadek: pusty tekst
func TestEmptyText(t *testing.T) {
	a := Build([]string{"abc"})
	matches := a.Search([]byte(""))

	if len(matches) != 0 {
		t.Errorf("oczekiwano 0 dopasowań dla pustego tekstu, got %d", len(matches))
	}
}

// TestPatternAtEnd — wzorzec na końcu tekstu
func TestPatternAtEnd(t *testing.T) {
	a := Build([]string{"end"})
	matches := a.Search([]byte("the end"))

	if len(matches) != 1 {
		t.Fatalf("oczekiwano 1 dopasowania, got %d", len(matches))
	}
	// "end" kończy się na pozycji 6 (ostatni bajt tekstu)
	if matches[0].Position != 6 {
		t.Errorf("oczekiwano Pos=6, got %d", matches[0].Position)
	}
}

// TestCaseSensitive — automat pracuje na bajtach, wielkość liter ma znaczenie
func TestCaseSensitive(t *testing.T) {
	a := Build([]string{"ABC"})
	matches := a.Search([]byte("abc ABC"))

	if len(matches) != 1 {
		t.Fatalf("oczekiwano 1 dopasowania (tylko 'ABC'), got %d", len(matches))
	}
}