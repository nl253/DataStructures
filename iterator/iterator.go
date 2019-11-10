package iterator

import (
	"fmt"
	"io"
	"math/rand"
	"os"
	"sync"

	"github.com/nl253/DataStructures/list"
)

type Iterator struct {
	lk    *sync.Mutex
	f     func(interface{}) interface{}
	state interface{}
}

type finished struct{}

var EndOfIteration = &finished{}

func New(initState interface{}, f func(interface{}) interface{}) *Iterator {
	return &Iterator{
		lk: &sync.Mutex{},
		f: func(state interface{}) interface{} {
			if state == EndOfIteration {
				return state
			} else {
				return f(state)
			}
		},
		state: initState,
	}
}

func FromFile(filePath string) *Iterator {
	file, e := os.Open(filePath)
	if e != nil {
		panic(fmt.Sprintf("[ERROR] failed to open file - %s", e.Error()))
	}
	buf := make([]byte, 1, 1)
	off := int64(0)
	_, err := file.ReadAt(buf, off)
	if err == io.EOF {
		return Repeat(EndOfIteration)
	}
	if err != nil {
		panic(fmt.Sprintf("[ERROR] failed to read file %s at offset %d - %s", filePath, off, err.Error()))
	}
	off++
	return New(buf[0], func(_ interface{}) interface{} {
		_, err := file.ReadAt(buf, off)
		if err == io.EOF {
			return EndOfIteration
		}
		if err != nil {
			panic(fmt.Sprintf("[ERROR] failed to read file %s at offset %d - %s", filePath, off, err.Error()))
		}
		off++
		return buf[0]
	})
}

// func FromFileSplit(filePath string, delim byte) *Iterator {
//     it := FromFile(filePath)
// }

// func FromFileLines(filePath string) *Iterator {
// 	return FileSplit(filePath, '\n')
// }

func FromStr(s string) *Iterator {
	i := 1
	return New(s[0], func(_ interface{}) interface{} {
		if i < len(s) {
			tmp := s[i]
			i++
			return tmp
		}
		return EndOfIteration
	})
}

func Range(initState int, step int) *Iterator {
	return New(initState, func(i interface{}) interface{} { return i.(int) + step })
}

func Nats() *Iterator {
	return New(0, func(i interface{}) interface{} { return i.(int) + 1 })
}

func Ints() *Iterator {
	return Range(0, 1)
}

func FromClojure(f func() interface{}) *Iterator {
	return New(f(), func(_ interface{}) interface{} { return rand.Float64() })
}

func RandF64s() *Iterator {
	return New(rand.Float64(), func(_ interface{}) interface{} { return rand.Float64() })
}

func RandF32s() *Iterator {
	return New(rand.Float32(), func(_ interface{}) interface{} { return rand.Float32() })
}

func Repeat(x interface{}) *Iterator {
	return New(x, func(y interface{}) interface{} { return y })
}

func (iter *Iterator) Take(n uint) *Iterator {
	iter.lk.Lock()
	defer iter.lk.Unlock()
	if n == 0 {
		return Repeat(EndOfIteration)
	}
	return New(iter.pull(), func(state interface{}) interface{} {
		if n <= 1 {
			return EndOfIteration
		} else {
			n--
			return iter.Pull()
		}
	})
}

func (iter *Iterator) Peek() interface{} {
	iter.lk.Lock()
	defer iter.lk.Unlock()
	return iter.state
}

func (iter *Iterator) Close() *Iterator {
	iter.lk.Lock()
	defer iter.lk.Unlock()
	iter.state = EndOfIteration
	iter.f = func(_ interface{}) interface{} {
		return EndOfIteration
	}
	return iter
}

func (iter *Iterator) Empty() bool {
	iter.lk.Lock()
	defer iter.lk.Unlock()
	return iter.state == EndOfIteration
}

func (iter *Iterator) Pull() interface{} {
	iter.lk.Lock()
	defer iter.lk.Unlock()
	return iter.pull()
}

func (iter *Iterator) tryAdvance() bool {
	ok := iter.state != EndOfIteration
	if ok {
		iter.state = iter.f(iter.state)
	}
	return ok
}

func (iter *Iterator) pull() interface{} {
	defer iter.tryAdvance()
	return iter.state
}

func (iter *Iterator) PullN(n uint) []interface{} {
	iter.lk.Lock()
	defer iter.lk.Unlock()
	result := make([]interface{}, n)
	for i := uint(0); i < n; i++ {
		result[i] = iter.pull()
	}
	return result
}

func (iter *Iterator) PullAll() *list.ConcurrentList {
	iter.lk.Lock()
	defer iter.lk.Unlock()
	result := list.New()
	for iter.state != EndOfIteration {
		result.PushBack(iter.pull())
	}
	return result
}

func (iter *Iterator) Consume() {
	iter.lk.Lock()
	defer iter.lk.Unlock()
	for iter.state != EndOfIteration {
		iter.skip()
	}
}

func (iter *Iterator) Skip() *Iterator {
	return iter.SkipN(1)
}

func (iter *Iterator) skip() *Iterator {
	return iter.skipN(1)
}

func (iter *Iterator) SkipN(n uint) *Iterator {
	iter.lk.Lock()
	defer iter.lk.Unlock()
	return iter.skipN(n)
}

func (iter *Iterator) skipN(n uint) *Iterator {
	for i := uint(0); i < n; i++ {
		if !iter.tryAdvance() {
			return iter
		}
	}
	return iter
}

func (iter *Iterator) Map(f func(interface{}) interface{}) *Iterator {
	iter.lk.Lock()
	defer iter.lk.Unlock()
	return New(f(iter.pull()), func(state interface{}) interface{} { return f(iter.Pull()) })
}

func (iter *Iterator) Log(writer io.Writer, format string) *Iterator {
	return iter.ForEach(func(x interface{}) {
		if _, err := fmt.Fprintf(writer, format, x); err != nil {
			if writer != os.Stdout {
				fmt.Printf("[ERROR] %s", err.Error())
			}
			_, _ = fmt.Fprintf(writer, "[ERROR] %s", err.Error())
		}
	})
}

func (iter *Iterator) Printf(format string) *Iterator {
	return iter.Log(os.Stdout, format)
}

func (iter *Iterator) Println() *Iterator {
	return iter.Printf("%v\n")
}

func (iter *Iterator) ForEach(f func(interface{})) *Iterator {
	return iter.Map(func(x interface{}) interface{} {
		if x != EndOfIteration {
			f(x)
		}
		return x
	})
}

func (iter *Iterator) Filter(f func(interface{}) bool) *Iterator {
	iter.lk.Lock()
	defer iter.lk.Unlock()
	return New(iter.state, func(state interface{}) interface{} {
		focus := state
		for !f(focus) {
			focus = iter.Pull()
		}
		return focus
	})
}

func (iter *Iterator) ReduceN(n uint, f func(interface{}, interface{}) interface{}) interface{} {
	iter.lk.Lock()
	defer iter.lk.Unlock()
	acc := iter.pull()
	for i := uint(1); i < n; i++ {
		acc = f(acc, iter.pull())
	}
	return acc
}

func (iter *Iterator) ReduceAll(f func(interface{}, interface{}) interface{}) interface{} {
	iter.lk.Lock()
	defer iter.lk.Unlock()
	acc := iter.pull()
	for {
		if x := iter.pull(); x != EndOfIteration {
			acc = f(acc, x)
		} else {
			return acc
		}
	}
}

func (iter *Iterator) Sum() int {
	return iter.ReduceAll(func(x interface{}, y interface{}) interface{} { return x.(int) + y.(int) }).(int)
}

func (iter *Iterator) Count() int {
	return iter.Map(func(x interface{}) interface{} { return 1 }).Sum()
}

func (iter *Iterator) Clone() *Iterator {
	iter.lk.Lock()
	defer iter.lk.Unlock()
	return New(iter.state, iter.f)
}

func (iter *Iterator) Eq(x interface{}) bool {
	switch x.(type) {
	case *Iterator:
		return iter == x.(*Iterator)
	default:
		return false
	}
}

func (iter *Iterator) String() string {
	return "Iterator"
}
