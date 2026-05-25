package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"
	"unsafe"

	cl "github.com/jgillich/go-opencl/cl"

	"github.com/hubertwojcik/parallel-pattern-search/internal/automaton"
)

func main() {
	patternsFile := flag.String("patterns", "", "file with patterns")
	textFile     := flag.String("text", "", "text file")
	kernelFile   := flag.String("kernel", "kernels/pfac.cl", "path to PFAC kernel")
	flag.Parse()

	if *patternsFile == "" || *textFile == "" {
		fmt.Fprintln(os.Stderr, "usage: pfac_ocl --patterns FILE --text FILE [--kernel FILE]")
		os.Exit(1)
	}

	patterns, err := loadLines(*patternsFile)
	if err != nil {
		log.Fatal(err)
	}
	text, err := os.ReadFile(*textFile)
	if err != nil {
		log.Fatal(err)
	}
	kernelSrc, err := os.ReadFile(*kernelFile)
	if err != nil {
		log.Fatalf("cannot read kernel %s: %v", *kernelFile, err)
	}

	// --- Budowanie automatu PFAC (CPU) ---
	buildStart := time.Now()
	table := automaton.BuildPFAC(patterns)
	buildDur := time.Since(buildStart)

	// --- Inicjalizacja OpenCL ---
	platforms, err := cl.GetPlatforms()
	if err != nil || len(platforms) == 0 {
		log.Fatal("no OpenCL platforms:", err)
	}

	var device *cl.Device
	for _, p := range platforms {
		devs, e := p.GetDevices(cl.DeviceTypeGPU)
		if e == nil && len(devs) > 0 {
			device = devs[0]
			break
		}
	}
	if device == nil {
		devs, e := platforms[0].GetDevices(cl.DeviceTypeAll)
		if e != nil || len(devs) == 0 {
			log.Fatal("no OpenCL devices found")
		}
		device = devs[0]
	}

	ctx, err := cl.CreateContext([]*cl.Device{device})
	if err != nil {
		log.Fatal("create context:", err)
	}

	queue, err := ctx.CreateCommandQueue(device, 0)
	if err != nil {
		log.Fatal("create queue:", err)
	}

	program, err := ctx.CreateProgramWithSource([]string{string(kernelSrc)})
	if err != nil {
		log.Fatal("create program:", err)
	}
	if err = program.BuildProgram(nil, ""); err != nil {
		log.Fatal("build program:", err)
	}

	kernel, err := program.CreateKernel("pfac")
	if err != nil {
		log.Fatal("create kernel:", err)
	}

	// --- Bufory GPU ---
	textBuf, err := ctx.CreateBuffer(cl.MemReadOnly|cl.MemCopyHostPtr, text)
	if err != nil {
		log.Fatal("text buffer:", err)
	}

	gotoBytes := int32SliceToBytes(table.Goto)
	gotoBuf, err := ctx.CreateBuffer(cl.MemReadOnly|cl.MemCopyHostPtr, gotoBytes)
	if err != nil {
		log.Fatal("goto buffer:", err)
	}

	outBytes := int32SliceToBytes(table.HasOutput)
	outBuf, err := ctx.CreateBuffer(cl.MemReadOnly|cl.MemCopyHostPtr, outBytes)
	if err != nil {
		log.Fatal("output buffer:", err)
	}

	var matchCount int32
	resultBuf, err := ctx.CreateBufferUnsafe(cl.MemReadWrite|cl.MemCopyHostPtr, 4, unsafe.Pointer(&matchCount))
	if err != nil {
		log.Fatal("result buffer:", err)
	}

	// --- Argumenty kernela ---
	textLen     := int32(len(text))
	alphabetSz  := int32(automaton.AlphabetSize)

	must(kernel.SetArgBuffer(0, textBuf))
	must(kernel.SetArgInt32(1, textLen))
	must(kernel.SetArgBuffer(2, gotoBuf))
	must(kernel.SetArgBuffer(3, outBuf))
	must(kernel.SetArgInt32(4, alphabetSz))
	must(kernel.SetArgBuffer(5, resultBuf))

	// --- Uruchomienie kernela ---
	searchStart := time.Now()
	if _, err = queue.EnqueueNDRangeKernel(kernel, nil, []int{len(text)}, nil, nil); err != nil {
		log.Fatal("enqueue kernel:", err)
	}
	if err = queue.Finish(); err != nil {
		log.Fatal("queue finish:", err)
	}
	searchDur := time.Since(searchStart)

	// --- Odczyt wyniku ---
	if _, err = queue.EnqueueReadBuffer(resultBuf, true, 0, 4, unsafe.Pointer(&matchCount), nil); err != nil {
		log.Fatal("read result:", err)
	}

	gb := float64(len(text)) / 1e9
	throughput := gb / searchDur.Seconds()

	fmt.Printf("patterns:    %d\n", len(patterns))
	fmt.Printf("text:        %.2f MB\n", float64(len(text))/1e6)
	fmt.Printf("PFAC states: %d\n", table.NumStates)
	fmt.Printf("build:       %v\n", buildDur)
	fmt.Printf("search:      %v\n", searchDur)
	fmt.Printf("matches:     %d\n", matchCount)
	fmt.Printf("throughput:  %.3f GB/s\n", throughput)
}

func int32SliceToBytes(s []int32) []byte {
	if len(s) == 0 {
		return nil
	}
	return unsafe.Slice((*byte)(unsafe.Pointer(&s[0])), len(s)*4)
}

func must(err error) {
	if err != nil {
		log.Fatal(err)
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
			if i > start {
				lines = append(lines, string(data[start:i]))
			}
			start = i + 1
		}
	}
	if start < len(data) {
		lines = append(lines, string(data[start:]))
	}
	return lines, nil
}
