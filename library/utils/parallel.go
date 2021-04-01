package utils

import (
	"runtime"
	"sync"
)

var numParallelInstance = runtime.NumCPU() * 5

//ApplyParallel is apply each funcs parallelly
func ApplyParallel(start, end int, fnc func(start, end int)) {
	chunks := divideToChunk(start, end)
	var wg sync.WaitGroup
	for _, chunk := range chunks {
		wg.Add(1)
		funcChunk := chunk
		go func() {
			defer wg.Done()
			fnc(funcChunk["start"], funcChunk["end"])
		}()
	}
	wg.Wait()
}

func divideToChunk(start, end int) []map[string]int {
	chunks := make([]map[string]int, 0)
	allCount := end - start
	dividePoint := allCount / numParallelInstance
	if dividePoint < 1 {
		dividePoint = allCount
	}
	countStart := start
	countEnd := dividePoint
	for countStart < end {
		chunk := make(map[string]int, 0)
		chunk["start"] = countStart
		chunk["end"] = countEnd
		chunks = append(chunks, chunk)
		countStart += dividePoint
		countEnd += dividePoint
		if countEnd > end {
			countEnd = end
		}
	}
	return chunks
}
