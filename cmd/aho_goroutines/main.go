package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/hubertwojcik/parallel-pattern-search/internal/automaton"
)


func parallelSearch(a *automaton.Automaton, text []byte, workers, overlap int) []automaton.Match {
	n := len(text)
	chunkSize := n / workers
	results := make([][]automaton.Match, workers)
	var wg sync.WaitGroup

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			start := idx * chunkSize
			end := start + chunkSize
			if idx == workers-1 {
				end = n
			}
			searchEnd := end + overlap
			if searchEnd > n {
				searchEnd = n
			}

			var adjusted []automaton.Match
			for _, m := range a.Search(text[start:searchEnd]) {
				absPos := start + m.Position
				// keep match if it STARTS before this chunk's end
				// (absPos is the end of the match, so start = absPos - patLen + 1)
				matchStart := absPos - a.PatternLens[m.PatternID] + 1
				if matchStart < end {
					adjusted = append(adjusted, automaton.Match{
						PatternID: m.PatternID,
						Position:  absPos,
					})
				}
			}
			results[idx] = adjusted
		}(i)
	}

	wg.Wait()
	var all []automaton.Match
	for _, r := range results {
		all = append(all, r...)
	}
	return all
}

func main() {
	patternsFile := flag.String("patterns", "", "file with patterns")
	textFile     := flag.String("text", "", "text file")
	workers      := flag.Int("workers", runtime.NumCPU(), "number of goroutines")
	flag.Parse()

	if *patternsFile == "" || *textFile == "" {
		fmt.Fprintln(os.Stderr, "usage: aho_goroutines --patterns FILE --text FILE [--workers N]")
		os.Exit(1)
	}

	patterns, err := loadLines(*patternsFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading patterns: %v\n", err)
		os.Exit(1)
	}

	text, err := os.ReadFile(*textFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading text: %v\n", err)
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
	w := *workers
	if w > n {
		w = n 
	}

	searchStart := time.Now()
	allMatches := parallelSearch(a, text, w, overlap)
	searchDur := time.Since(searchStart)

	total := len(allMatches)

	gb := float64(len(text)) / 1e9
	throughput := gb / searchDur.Seconds()

	fmt.Printf("patterns:    %d\n", len(patterns))
	fmt.Printf("text:        %.2f MB\n", float64(len(text))/1e6)
	fmt.Printf("workers:     %d\n", w)
	fmt.Printf("build:       %v\n", buildDur)
	fmt.Printf("search:      %v\n", searchDur)
	fmt.Printf("matches:     %d\n", total)
	fmt.Printf("throughput:  %.3f GB/s\n", throughput)
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