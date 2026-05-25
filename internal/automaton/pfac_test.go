package automaton

import "testing"

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

func TestPFACDeadStateAtRoot(t *testing.T) {
	pfac := BuildPFAC([]string{"abc"})
	idx := 0*AlphabetSize + int('x')
	if pfac.Goto[idx] != -1 {
		t.Errorf("PFAC: Goto[root]['x'] = %d, want -1", pfac.Goto[idx])
	}
}

func TestPFACNumStates(t *testing.T) {
	pfac := BuildPFAC([]string{"ab", "ac"})
	if pfac.NumStates != 4 {
		t.Errorf("want 4 states, got %d", pfac.NumStates)
	}
}

func TestPFACHasOutput(t *testing.T) {
	pfac := BuildPFAC([]string{"ab"})
	if pfac.HasOutput[2] != 1 {
		t.Errorf("state 2 should be accepting")
	}
	if pfac.HasOutput[0] != 0 || pfac.HasOutput[1] != 0 {
		t.Errorf("root and intermediate states should not be accepting")
	}
}

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
