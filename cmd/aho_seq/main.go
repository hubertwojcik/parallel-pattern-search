package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/hubertwojcik/parallel-pattern-search/internal/automaton"
)

func main(){
	patternsFile := flag.String("patterns", "", "plik z wzorcami (jeden per linia)")

	textFile := flag.String("text", "", "plik tekstowy")

	flag.Parse()

	if *patternsFile == "" || *textFile == "" {
		fmt.Fprintln(os.Stderr, "użycie: aho_seq --patterns FILE --text FILE")
		os.Exit(1)
	}

	patterns, err := loadLines(*patternsFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "blad czytania wzorcow: %v\n", err)
		os.Exit(1)
	}

	text, err := os.ReadFile(*textFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "blad czytania tekstu: %v\n", err)
		os.Exit(1)
	}

	buildStart := time.Now()
	a := automaton.Build(patterns)
	buildDur := time.Since(buildStart)

	searchStart := time.Now()
	matches := a.Search(text)
	searchDur := time.Since(searchStart)

	gb := float64(len(text)) / 1e9
	throughput := gb / searchDur.Seconds()


    fmt.Printf("wzorce:      %d\n", len(patterns))
    fmt.Printf("tekst:       %.2f MB\n", float64(len(text))/1e6)
    fmt.Printf("build:       %v\n", buildDur)
    fmt.Printf("search:      %v\n", searchDur)
    fmt.Printf("dopasowania: %d\n", len(matches))
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

	if start < len(data){
		lines = append(lines, string(data[start:]))
	}
	
	return lines, nil
}
