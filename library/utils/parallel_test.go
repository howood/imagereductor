package utils_test

import (
	"sync"
	"sync/atomic"
	"testing"

	"github.com/howood/imagereductor/library/utils"
)

func Test_ApplyParallel(t *testing.T) {
	t.Parallel()

	const start, end = 0, 1000
	var counter int64
	var mu sync.Mutex
	visited := make(map[int]struct{}, end-start)

	utils.ApplyParallel(start, end, func(s, e int) {
		for i := s; i < e; i++ {
			atomic.AddInt64(&counter, 1)
			mu.Lock()
			visited[i] = struct{}{}
			mu.Unlock()
		}
	})

	if int(counter) < end-start {
		t.Fatalf("ApplyParallel processed less than expected: got %d, want >= %d", counter, end-start)
	}
	for i := start; i < end; i++ {
		if _, ok := visited[i]; !ok {
			t.Fatalf("ApplyParallel missed index %d", i)
		}
	}
}

func Test_ApplyParallel_SmallRange(t *testing.T) {
	t.Parallel()

	var called int64
	utils.ApplyParallel(0, 1, func(_, _ int) {
		atomic.AddInt64(&called, 1)
	})
	if called == 0 {
		t.Fatal("ApplyParallel did not invoke callback for small range")
	}
}
