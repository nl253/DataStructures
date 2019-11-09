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
		return ss.Pull() == 1 && ss.Pull() == 2 && ss.Pull() == 3 && ss.Pull() == EndMarker && ss.Pull() == EndMarker && ss.BufEmpty()
	})
	should("after close, all pulls result in EndMarker", true, func() interface{} {
		ss := New().Close()
		return ss.Pull() == EndMarker && ss.Pull() == EndMarker
	})
	for i := 0; i < 10; i++ {
		should("not freeze the runtime", true, func() interface{} {
			ss := New()
			go func() {
				for j := 0; j < 10; j++ {
					go func() {
						ss.PushBack(rand.Int())
					}()
				}
			}()
			ss.Close()
			for j := 0; j < 10; j++ {
				if ss.Pull() != EndMarker {
					fmt.Printf("expected non-end-marker but got endmarker\n")
					return false
				}
			}
			for j := 0; j < 10; j++ {
				if x := ss.Pull(); x != EndMarker {
					fmt.Printf("expected end-marker but got %v :: %T\n", x, x)
					return false
				}
			}
			return true
		})
	}
}

func TestStream_Range(t *testing.T) {
	should := fStream("Range", t)
	should("range puts ints from [min, max) to stream", list.New(1, 2, 3), func() interface{} { return Range(1, 4, 1).PullAll() })
	should("range with upper = lower gives an empty stream", list.New(), func() interface{} { return Range(0, 0, 1).PullAll() })
	should("generate valid stream", true, func() interface{} { return isValid(Range(1, 4, 1)) })
}

func TestStream_Drain(t *testing.T) {
	should := fStream("BufDrain", t)
	should("drain internal buffer to list", list.New(1, 4, 1), func() interface{} { return New(1, 4, 1).BufDrain() })
	should("drain empty internal buffer to empty list", list.New(), func() interface{} { return New().BufDrain() })
}

func TestStream_PullAll(t *testing.T) {
	should := fStream("PullAll", t)
	should("pull all", list.New(1, 2, 3), func() interface{} { return Range(1, 4, 1).PullAll() })
	should("pull all empty", list.New(), func() interface{} { return Range(0, 0, 1).PullAll() })
}

func TestStream_Count(t *testing.T) {
	should := fStream("Count", t)
	should("count elems", uint(10), func() interface{} { return Natural(10).Count() })
	should("count elems", uint(0), func() interface{} { return Natural(0).Count() })
}

func TestStream_Concat(t *testing.T) {
	should := fStream("Concat", t)
	should("concat elems", "abc", func() interface{} { return New("a", "b", "c").Close().Concat() })
	should("concat elems", "", func() interface{} { return RandStrs(10, 20, 0).Close().Concat() })
}

func TestStream_Sum(t *testing.T) {
	should := fStream("Sum", t)
	should("sum many elems", float64(0+1+2), func() interface{} { return Ints(0, 3).Sum() })
	should("sum 0 elems", float64(0), func() interface{} { return Ints(0, 0).Sum() })
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

func TestStream_RandF32s(t *testing.T) {
	should := fStream("RandF32s", t)
	should("make stream of f32", true, func() interface{} {
		return RandF32s(10).PullAll().All(func(x interface{}) bool {
			switch x.(type) {
			case float32:
				return true
			default:
				return false
			}
		})
	})
	should("map 0 elements", list.New(), func() interface{} {
		return RandF32s(0).PullAll()
	})
	should("generate valid stream", true, func() interface{} {
		return isValid(RandF32s(100))
	})
}

func TestStream_RandF64s(t *testing.T) {
	should := fStream("RandF64s", t)
	should("make stream of f64", true, func() interface{} {
		return RandF64s(10).PullAll().All(func(x interface{}) bool {
			switch x.(type) {
			case float64:
				return true
			default:
				return false
			}
		})
	})
	should("map 0 elements", list.New(), func() interface{} {
		return RandF64s(0).PullAll()
	})
	should("generate valid stream", true, func() interface{} {
		return isValid(RandF64s(100))
	})
}

func TestStream_RandInts(t *testing.T) {
	should := fStream("RandInts", t)
	should("make stream of ints", true, func() interface{} {
		return RandInts(-10, 10, 10).PullAll().All(func(x interface{}) bool {
			switch x.(type) {
			case int:
				return true
			default:
				return false
			}
		})
	})
	should("map 0 elements", list.New(), func() interface{} {
		return RandInts(-100, 100, 0).PullAll()
	})
	should("generate valid stream", true, func() interface{} {
		return isValid(RandInts(0, 100, 10))
	})
}

func TestStream_RandBytes(t *testing.T) {
	should := fStream("RandBytes", t)
	should("make stream of bytes", true, func() interface{} {
		return RandBytes(10).PullAll().All(func(x interface{}) bool {
			switch x.(type) {
			case byte:
				return true
			default:
				return false
			}
		})
	})
	should("map 0 elements", list.New(), func() interface{} {
		return RandBytes(0).PullAll()
	})
	should("generate valid stream", true, func() interface{} {
		return isValid(RandBytes(10))
	})
}

func TestStream_FromStr(t *testing.T) {
	should := fStream("FromStr", t)
	should("make stream of bytes", true, func() interface{} {
		return FromStr("abc").PullAll().All(func(x interface{}) bool {
			switch x.(type) {
			case byte:
				return true
			default:
				return false
			}
		})
	})
	should("'abc' makes stream of 3 bytes", uint(3), func() interface{} {
		return FromStr("abc").Count()
	})
	should("'abc' makes stream of bytes", list.New(byte('a'), byte('b'), byte('c')), func() interface{} {
		return FromStr("abc").PullAll()
	})
	should("map 0 elements", list.New(), func() interface{} {
		return FromStr("").PullAll()
	})
	should("generate valid stream", true, func() interface{} {
		return isValid(FromStr("abc")) && isValid(FromStr(""))
	})
}
