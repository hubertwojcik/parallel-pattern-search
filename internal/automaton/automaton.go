package automaton

const AlphabetSize = 256

type State = int

type Match struct {
	PatternID int
	Position  int
}

type Automaton struct {
	Goto [][]State
	Fail []State
	Output [][]int
	PatternLens []int	
} 


func Build(patterns []string) *Automaton {
	a := &Automaton{}
	a.PatternLens = make([]int, len(patterns))
	
	for i, p := range patterns {
		a.PatternLens[i] = len(p)
	}
	
	// 1. Budowanie drzewa
	a.addState()

	for id, pat := range patterns {
		cur := 0
		for i:=0; i < len(pat); i++ {
			c := int(pat[i])
			
			if a.Goto[cur][c] == -1{
				next := a.addState()
				a.Goto[cur][c] = next
			}
			cur = a.Goto[cur][c]
		}
		a.Output[cur] = append(a.Output[cur], id)
	}

	// 2. Kompletacja root
	for c := 0; c < AlphabetSize; c++ {
		if a.Goto[0][c] == -1 {
			a.Goto[0][c] =0
		}
	}

	// 3. Obliczaenie BFS
	queue := make([]State, 0, len(a.Goto))

	for c :=0; c < AlphabetSize; c++ {
		s := a.Goto[0][c]
		if s != 0 {
			a.Fail[s] = 0
			queue = append(queue, s)
		}
	}
	
	for len(queue) > 0 {
		r := queue[0]
		queue = queue[1:]
		
		for c := 0; c< AlphabetSize; c++{
			s := a.Goto[r][c]
			if s == -1 {
				a.Goto[r][c] = a.Goto[a.Fail[r]][c]
				continue
			}
			a.Fail[s] = a.Goto[a.Fail[r]][c]
			a.Output[s] = append(a.Output[s], a.Output[a.Fail[s]]...)
			queue = append(queue, s)
		}
 	}

	return a
}

func (a *Automaton) Search(text []byte) []Match {
	var matches []Match
	cur := 0

	for i, b := range text {
		cur = a.Goto[cur][int(b)]

for _, id := range a.Output[cur] {
	matches = append(matches, Match{
		PatternID:id,
		Position: i,
	})
}

	}
	return matches
	
}

func (a *Automaton) addState() State {
	idx := len(a.Goto)

	row := make([]State, AlphabetSize)
	for i:= range row {
		row[i] = -1
	}

	a.Goto=append(a.Goto, row)
	a.Fail = append(a.Fail, 0)
	a.Output = append(a.Output, nil)
	return idx
}