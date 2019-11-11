package iterator

import (
	"fmt"
	"io"
	"math/rand"
	"os"
	"strings"
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
			switch state {
			case EndOfIteration:
				return state
			default:
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
	return New(nil, func(_ interface{}) interface{} {
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

func FromFileSplit(filePath string, delim byte) *Iterator {
	file, e := os.Open(filePath)
	if e != nil {
		panic(fmt.Sprintf("[ERROR] failed to open file - %s", e.Error()))
	}
	buf := make([]byte, 1, 1)
	off := int64(0)
	return New(nil, func(_ interface{}) interface{} {
		sb := strings.Builder{}
		for {
			_, err := file.ReadAt(buf, off)
			if err == io.EOF {
				return sb.String()
			}
			if err != nil {
				panic(fmt.Sprintf("[ERROR] failed to read file %s at offset %d - %s", filePath, off, err.Error()))
			}
			off++
			if buf[0] == delim {
				return sb.String()
			}
			sb.WriteByte(buf[0])
		}
	})
}

func FromFileLines(filePath string) *Iterator {
	return FromFileSplit(filePath, '\n')
}

func FromStr(s string) *Iterator {
	i := 0
	return New(nil, func(_ interface{}) interface{} {
		if i < len(s) {
			tmp := s[i]
			i++
			return tmp
		}
		return EndOfIteration
	})
}

func Range(initState int, step int) *Iterator {
	return New(initState-step, func(i interface{}) interface{} { return i.(int) + step })
}

func Nats() *Iterator {
	return New(uint(0), func(i interface{}) interface{} { return i.(uint) + 1 }).Map(func(n interface{}) interface{} {
		return n.(uint) - 1
	})
}

func Ints() *Iterator {
	return Range(0, 1)
}

func FromClojure(f func() interface{}) *Iterator {
	return New(nil, func(_ interface{}) interface{} { return f() })
}

func RandF64s() *Iterator {
	return FromClojure(func() interface{} { return rand.Float64() })
}

func RandF32s() *Iterator {
	return FromClojure(func() interface{} { return rand.Float32() })
}

func Repeat(x interface{}) *Iterator {
	return New(x, func(y interface{}) interface{} { return y })
}

func (iter *Iterator) Take(n uint) *Iterator {
	iter.lk.Lock()
	defer iter.lk.Unlock()
	return New(iter.state, func(state interface{}) interface{} {
		if n == 0 {
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

func (iter *Iterator) pull() interface{} {
	iter.state = iter.f(iter.state)
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
	for i := uint(0); i < n && iter.state != EndOfIteration; i++ {
		iter.state = iter.f(iter.state)
	}
	return iter
}

func (iter *Iterator) Map(f func(interface{}) interface{}) *Iterator {
	iter.lk.Lock()
	defer iter.lk.Unlock()
	return New(iter.state, func(state interface{}) interface{} { return f(iter.Pull()) })
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
		for !f(focus) && focus != EndOfIteration {
			focus = iter.Pull()
		}
		return focus
	})
}

func (iter *Iterator) ReduceN(init interface{}, f func(interface{}, interface{}) interface{}, n uint) interface{} {
	iter.lk.Lock()
	defer iter.lk.Unlock()
	acc := init
	var x interface{}
	for i := uint(0); i < n; i++ {
		if x = iter.pull(); x == EndOfIteration {
			break
		}
		acc = f(acc, x)
	}
	return acc
}

func (iter *Iterator) ReduceAll(init interface{}, f func(interface{}, interface{}) interface{}) interface{} {
	iter.lk.Lock()
	defer iter.lk.Unlock()
	acc := init
	for x := iter.pull(); x != EndOfIteration; x = iter.pull() {
		acc = f(acc, x)
	}
	return acc
}

func (iter *Iterator) Sum() int {
	return iter.ReduceAll(0, func(x interface{}, y interface{}) interface{} { return x.(int) + y.(int) }).(int)
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
