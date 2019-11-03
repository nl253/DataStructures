package stream

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/nl253/DataStructures/list"
	ut "github.com/nl253/Testing"
)

var fStream = ut.Test("Stream")

func all(xs []bool) bool {
	for _, x := range xs {
		if !x {
			return false
		}
	}
	return true
}

func isValid(xs *Stream) bool {
	if xs.closed && xs.buf.Empty() && xs.Pull() != EndMarker {
		fmt.Printf("closed streams should emit EndMarker\n")
		return false
	}
	time.Sleep(time.Second * 1)
	if xs.closed && !xs.lks.Empty() {
		fmt.Printf("goroutines still waiting\n")
		return false
	}
	nChildThreads := uint(10)
	concurrentAccessOk := make([]bool, nChildThreads, nChildThreads)
	for idx := uint(0); idx < nChildThreads; idx++ {
		go func(idx uint) {
			time.Sleep(time.Millisecond * time.Duration(rand.Intn(10)))
			xs.Pull()
			concurrentAccessOk[idx] = true
		}(idx)
	}
	time.Sleep(time.Second * 1)
	if !all(concurrentAccessOk) {
		fmt.Printf("concurrent access failed\n")
		return false
	}
	return true
}

func TestStream_Concurrency(t *testing.T) {
	should := fStream("general concurrency", t)
	should("not freeze the runtime", true, func() interface{} {
		ss := New(1, 2, 3).Close()
		return ss.Pull() == 1 && ss.Pull() == 2 && ss.Pull() == 3 && ss.Pull() == EndMarker && ss.Pull() == EndMarker && ss.Empty()
	})
	should("after close, all pulls result in EndMarker", true, func() interface{} {
		ss := New().Close()
		return ss.Pull() == EndMarker && ss.Pull() == EndMarker
	})
}

func TestStream_Range(t *testing.T) {
	should := fStream("Range", t)
	should("range", list.New(1, 2, 3), func() interface{} { return Range(1, 4, 1).PullAll() })
	should("range empty", list.New(), func() interface{} { return Range(0, 0, 1).PullAll() })
	should("generate valid stream", true, func() interface{} { return isValid(Range(1, 4, 1)) })
}

func TestStream_PullAll(t *testing.T) {
	should := fStream("PullAll", t)
	should("pull all", list.New(1, 2, 3), func() interface{} { return Range(1, 4, 1).PullAll() })
	should("pull all empty", list.New(), func() interface{} { return Range(0, 0, 1).PullAll() })
}

func TestStream_Count(t *testing.T) {
	should := fStream("Count", t)
	should("count elems", uint(10), func() interface{} {
		return Range(0, 10, 1).Close().Count()
	})
}

func TestStream_Sum(t *testing.T) {
	should := fStream("Sum", t)
	should("sum many elems", 0+1+2, func() interface{} { return Range(0, 3, 1).Sum() })
	should("sum 0 elems", 0, func() interface{} { return Range(0, 0, 1).Sum() })
}

func TestStream_Map(t *testing.T) {
	should := fStream("map", t)
	should("map many elements", list.New(1, 2, 3), func() interface{} {
		return Range(0, 3, 1).Map(func(x interface{}) interface{} { return x.(int) + 1 }).PullAll()
	})
	should("map 0 elements", list.New(), func() interface{} {
		return Range(0, 0, 1).Map(func(x interface{}) interface{} { return x.(int) + 1 }).PullAll()
	})
	should("generate valid stream", true, func() interface{} {
		return isValid(Range(0, 0, 1).Map(func(x interface{}) interface{} { return x.(int) + 1 }))
	})
}
