package automaton

type PFACTable struct {
	Goto      []int32
	HasOutput []int32
	NumStates int
}

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
