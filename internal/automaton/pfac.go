package automaton

// PFACTable to spłaszczony, gotowy na GPU automat PFAC.
// Goto[state*AlphabetSize + c] = następny stan (-1 = martwy).
// HasOutput[state] = 1 jeśli w tym stanie kończy się jakiś wzorzec.
//
// Różnica względem AC: brak dowiązań awaryjnych (failure links).
// W korzeniu brakujące znaki dają -1, nie pętlę do korzenia.
// Dzięki temu każdy wątek GPU może niezależnie zatrzymać się przy
// pierwszym nierozpoznanym znaku.
type PFACTable struct {
	Goto      []int32
	HasOutput []int32
	NumStates int
}

// BuildPFAC buduje samo drzewo trie (bez failure links) dla algorytmu PFAC.
func BuildPFAC(patterns []string) *PFACTable {
	gotoRaw := [][]int32{make([]int32, AlphabetSize)}
	for i := range gotoRaw[0] {
		gotoRaw[0][i] = -1
	}
	hasOut := []bool{false}

	addState := func() int {
		row := make([]int32, AlphabetSize)
		for i := range row {
			row[i] = -1
		}
		gotoRaw = append(gotoRaw, row)
		hasOut = append(hasOut, false)
		return len(gotoRaw) - 1
	}

	for _, pat := range patterns {
		cur := 0
		for _, b := range []byte(pat) {
			c := int(b)
			if gotoRaw[cur][c] == -1 {
				gotoRaw[cur][c] = int32(addState())
			}
			cur = int(gotoRaw[cur][c])
		}
		hasOut[cur] = true
	}

	numStates := len(gotoRaw)
	flat := make([]int32, numStates*AlphabetSize)
	for s, row := range gotoRaw {
		copy(flat[s*AlphabetSize:], row)
	}

	outFlat := make([]int32, numStates)
	for s, has := range hasOut {
		if has {
			outFlat[s] = 1
		}
	}

	return &PFACTable{
		Goto:      flat,
		HasOutput: outFlat,
		NumStates: numStates,
	}
}
