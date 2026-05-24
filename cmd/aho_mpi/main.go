package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/hubertwojcik/parallel-pattern-search/internal/automaton"
	mpi "github.com/sbromberger/gompi"
)

func main() {

	mpi.Start(false)
	defer mpi.Stop()

comm := mpi.NewCommunicator(nil)

	rank := comm.Rank()
	size := comm.Size()

	patternsFile := flag.String("patterns", "", "file with patterns")
	textFile     := flag.String("text", "", "text file")
	flag.Parse()

	if *patternsFile == "" || *textFile == "" {
		if rank == 0 {
			fmt.Fprintln(os.Stderr, "usage: mpirun -np N aho_mpi --patterns FILE --text FILE")
		}
		os.Exit(1)
	}

	
	
	
	patterns, err := loadLines(*patternsFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "rank %d: error reading patterns: %v\n", rank, err)
		os.Exit(1)
	}

	
	
	
	text, err := os.ReadFile(*textFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "rank %d: error reading text: %v\n", rank, err)
		os.Exit(1)
	}

	
	buildStart := time.Now()
	a := automaton.Build(patterns)
	buildDur := time.Since(buildStart)

	
	maxPatLen := 0
	for _, p := range patterns {
		if len(p) > maxPatLen {
			maxPatLen = len(p)
		}
	}
	overlap := maxPatLen - 1
	if overlap < 0 {
		overlap = 0
	}

	
	n := len(text)
	chunkSize := n / size

	start := rank * chunkSize
	end := start + chunkSize
	if rank == size-1 {
		end = n 
	}

	
	searchEnd := end + overlap
	if searchEnd > n {
		searchEnd = n
	}

	chunk := text[start:searchEnd]

	
	searchStart := time.Now()
	rawMatches := a.Search(chunk)
	searchDur := time.Since(searchStart)

	
	localCount := 0
	for _, m := range rawMatches {
		absPos := start + m.Position
		patLen := a.PatternLens[m.PatternID]
		absStart := absPos - patLen + 1

		if absStart >= start && absStart < end {
			localCount++
		}
	}

	
	
	localSlice := []int64{int64(localCount)}
	totalSlice := []int64{0}
	comm.ReduceInt64s(totalSlice, localSlice, mpi.OpSum, 0)


	
	if rank == 0 {
		totalCount := int(totalSlice[0])
		gb := float64(n) / 1e9
		throughput := gb / searchDur.Seconds()

		fmt.Printf("patterns:      %d\n", len(patterns))
		fmt.Printf("texts:       %.2f MB\n", float64(n)/1e6)
		fmt.Printf("processes:     %d\n", size)
		fmt.Printf("build:       %v\n", buildDur)
		fmt.Printf("search:      %v\n", searchDur)
		fmt.Printf("matches: %d\n", totalCount)
		fmt.Printf("throughput:  %.3f GB/s\n", throughput)
	}
}

func loadLines(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var lines []string
	start := 0
	for i, b := range data {
		if b == '\n' {
			line := string(data[start:i])
			if len(line) > 0 {
				lines = append(lines, line)
			}
			start = i + 1
		}
	}
	if start < len(data) {
		lines = append(lines, string(data[start:]))
	}
	return lines, nil
}