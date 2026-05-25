package automaton

import "testing"

// TestPFACMatchesAC sprawdza że BuildPFAC i Build dają identyczną liczbę
// dopasowań — to najważniejsza gwarancja poprawności trie.
func TestPFACMatchesAC(t *testing.T) {
	patterns := []string{"he", "she", "his", "hers"}
	texts := []string{
		"ushers",
		"ushers and she said his hers",
		"",
		"xyz",
		"hehehehehers",
	}

	ac := Build(patterns)
	pfac := BuildPFAC(patterns)

	for _, text := range texts {
		tb := []byte(text)
		acCount := len(ac.Search(tb))
		pfacCount := pfacSearch(pfac, tb)
		if acCount != pfacCount {
			t.Errorf("text=%q: AC=%d PFAC=%d", text, acCount, pfacCount)
		}
	}
}

// TestPFACDeadStateAtRoot sprawdza że w PFAC znaki spoza wzorców dają -1 z korzenia.
// W AC korzeń ma pętlę do siebie — w PFAC nie.
func TestPFACDeadStateAtRoot(t *testing.T) {
	pfac := BuildPFAC([]string{"abc"})
	// 'x' nie zaczyna żadnego wzorca — z korzenia powinien dać -1
	idx := 0*AlphabetSize + int('x')
	if pfac.Goto[idx] != -1 {
		t.Errorf("PFAC: Goto[root]['x'] = %d, want -1", pfac.Goto[idx])
	}
}

// TestPFACNumStates sprawdza że liczba stanów zgadza się z rozmiarem trie.
func TestPFACNumStates(t *testing.T) {
	// "ab", "ac" — korzeń + 'a' + 'b' + 'c' = 4 stany
	pfac := BuildPFAC([]string{"ab", "ac"})
	if pfac.NumStates != 4 {
		t.Errorf("want 4 states, got %d", pfac.NumStates)
	}
}

// TestPFACHasOutput sprawdza że stany akceptujące są poprawnie oznaczone.
func TestPFACHasOutput(t *testing.T) {
	pfac := BuildPFAC([]string{"ab"})
	// stan 0=root, 1='a', 2='ab' (accepting)
	if pfac.HasOutput[2] != 1 {
		t.Errorf("state 2 should be accepting")
	}
	if pfac.HasOutput[0] != 0 || pfac.HasOutput[1] != 0 {
		t.Errorf("root and intermediate states should not be accepting")
	}
}

// pfacSearch symuluje PFAC na CPU — używane tylko w testach.
func pfacSearch(t *PFACTable, text []byte) int {
	count := 0
	for i := range text {
		state := 0
		for j := i; j < len(text); j++ {
			c := int(text[j])
			next := int(t.Goto[state*AlphabetSize+c])
			if next < 0 {
				break
			}
			state = next
			if t.HasOutput[state] == 1 {
				count++
			}
		}
	}
	return count
}
