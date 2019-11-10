package stream

import (
	"fmt"
	"io"
	"math/rand"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/nl253/DataStructures/list"
)

type Stream struct {
	closed bool
	bufLk  *sync.Mutex
	lksLk  *sync.Mutex
	buf    *list.ConcurrentList
	lks    *list.ConcurrentList
}

type streamEnd struct{}

var EndMarker = &streamEnd{}

func (e *streamEnd) String() string {
	return "<END OF STREAM>"
}

func (e *streamEnd) Eq(x interface{}) bool {
	switch x.(type) {
	case *streamEnd:
		return x.(*streamEnd) == e
	default:
		return false
	}
}

func New(xs ...interface{}) *Stream {
	s := &Stream{
		bufLk:  &sync.Mutex{},
		lksLk:  &sync.Mutex{},
		buf:    list.New(),
		lks:    list.New(),
		closed: false,
	}
	for _, x := range xs {
		s.PushBack(x)
	}
	return s
}

func Pipeline(ss ...*Stream) *Stream {
	acc := New()
	for _, s := range ss {
		acc.PushBack(s)
	}
	acc.Close()
	return acc.Reduce(New(), func(acc interface{}, other interface{}) interface{} {
		return acc.(*Stream).Pipe(other.(*Stream))
	}).(*Stream)
}

func Generate(n int, m int, step int, f func(n int) interface{}) *Stream {
	s := New()
	go func() {
		for start := n; start < m; start += step {
			s.PushBack(f(start))
		}
		s.Close()
	}()
	return s
}

func FromSlice(xs []interface{}) *Stream { return New(xs...) }

func FromSliceSubslice(s []interface{}, n uint) *Stream {
	m := int(n)
	return Generate(0, len(s), int(n), func(state int) interface{} { return s[state : state+m] })
}

func FromFile(filePath string) *Stream {
	file, e := os.Open(filePath)
	if e != nil {
		panic(fmt.Sprintf("[ERROR] failed to open file - %s", e.Error()))
	}
	s := New()
	buf := make([]byte, 1, 1)
	off := int64(0)
	go func() {
		for _, err := file.ReadAt(buf, off); err != io.EOF; _, err = file.ReadAt(buf, off) {
			s.PushBack(buf[0])
			off++
		}
		s.Close()
	}()
	return s
}

func FromFileSplit(filePath string, delim byte) *Stream {
	s := FromFile(filePath)
	newS := New()
	go func() {
		sb := strings.Builder{}
		for !s.Closed() {
			if b := s.Pull().(byte); b == delim {
				newS.PushBack(sb.String())
				sb = strings.Builder{}
			} else {
				sb.WriteByte(b)
			}
		}
		s.Close()
	}()
	return newS
}

func FromFileLines(filePath string) *Stream {
	return FromFileSplit(filePath, '\n')
}

func FromStr(s string) *Stream {
	return Generate(0, len(s), 1, func(n int) interface{} { return s[n] })
}

func FromStrSubstr(s string, n uint) *Stream {
	m := int(n)
	return Generate(0, len(s), int(n), func(state int) interface{} { return s[state : state+m] })
}

func Range(lowBound int, upBound int, step int) *Stream {
	return Generate(lowBound, upBound, step, func(n int) interface{} { return n })
}

func Linear(lowBound float64, upBound float64, n uint) *Stream {
	s := New()
	go func() {
		step := (upBound - lowBound) / float64(n)
		for start := lowBound; start < upBound; start += step {
			s.PushBack(start)
		}
		s.Close()
	}()
	return s
}

func Nats(n uint) *Stream {
	s := New()
	go func() {
		for start := uint(0); start < n; start++ {
			s.PushBack(start)
		}
		s.Close()
	}()
	return s
}

func Ints(lowBound int, upBound int) *Stream {
	return Range(lowBound, upBound, 1)
}

func RandInts(min int, max int, n uint) *Stream {
	r := max - min
	return Nats(n).Map(func(_ interface{}) interface{} {
		return min + rand.Intn(r)
	})
}

func RandBytes(n uint) *Stream {
	return Nats(n).Map(func(x interface{}) interface{} {
		return byte(rand.Intn(256))
	})
}

func RandStrs(min uint, max uint, n uint) *Stream {
	return Nats(n).Map(func(_ interface{}) interface{} {
		sb := strings.Builder{}
		end := uint(rand.Intn(int(max)))
		for i := min; i < end; i++ {
			sb.WriteRune(33 + rand.Int31n(95))
		}
		return sb.String()
	})
}

func RandF64s(n int) *Stream {
	return Generate(0, n, 1, func(x int) interface{} { return rand.Float64() })
}

func RandF32s(n int) *Stream {
	return Generate(0, n, 1, func(x int) interface{} { return rand.Float32() })
}

func Replicate(n uint, xs ...interface{}) *Stream {
	return Repeat(n, FromSlice(xs)).Flatten()
}

func Repeat(n uint, x interface{}) *Stream {
	return Generate(0, int(n), 1, func(n int) interface{} {
		return x
	})
}

func Tick(freq time.Duration, n uint, x interface{}) *Stream {
	return Emit(freq, n, func(_ uint) interface{} { return x })
}

func Emit(freq time.Duration, count uint, f func(n uint) interface{}) *Stream {
	return Nats(count).Map(func(x interface{}) interface{} {
		time.Sleep(freq)
		return f(x.(uint))
	})
}

func (s *Stream) PushFront(t interface{}) {
	s.bufLk.Lock()
	s.buf.PushFront(t)
	s.lksLk.Lock()
	if !s.lks.Empty() {
		l := s.lks.PopFront()
		l.(*sync.Mutex).Unlock()
	}
	s.bufLk.Unlock()
	s.lksLk.Unlock()
}

func (s *Stream) PushBack(x interface{}) {
	s.bufLk.Lock()
	s.buf.PushBack(x)
	s.lksLk.Lock()
	if !s.lks.Empty() {
		s.lks.PopFront().(*sync.Mutex).Unlock()
	}
	s.lksLk.Unlock()
	s.bufLk.Unlock()
}

func (s *Stream) Pull() interface{} {
	for {
		s.bufLk.Lock()
		if s.buf.Empty() {
			if s.closed {
				s.bufLk.Unlock()
				return EndMarker
			}
			s.bufLk.Unlock()
			l := &sync.Mutex{}
			l.Lock()
			s.lksLk.Lock()
			s.lks.PushBack(l)
			s.lksLk.Unlock()
			l.Lock()
			l.Unlock()
			continue
		}
		front := s.buf.PopFront()
		s.bufLk.Unlock()
		return front
	}
}

func (s *Stream) PullN(n uint) []interface{} {
	xs := make([]interface{}, n)
	for i := uint(0); i < n; i++ {
		xs[i] = s.Pull()
	}
	return xs
}

func (s *Stream) PullAll() *list.ConcurrentList {
	return s.Reduce(list.New(), func(acc interface{}, x interface{}) interface{} {
		acc.(*list.ConcurrentList).PushBack(x)
		return acc
	}).(*list.ConcurrentList)
}

func (s *Stream) PeekFront() interface{} {
	for {
		s.bufLk.Lock()
		if s.buf.Empty() {
			if s.closed {
				s.bufLk.Unlock()
				return EndMarker
			}
			l := &sync.Mutex{}
			l.Lock()
			s.lksLk.Lock()
			s.lks.PushBack(l)
			s.lksLk.Unlock()
			s.bufLk.Unlock()
			l.Lock()
			l.Unlock()
			continue
		}
		front := s.buf.PeekFront()
		s.bufLk.Unlock()
		return front
	}
}

func (s *Stream) Close() *Stream {
	s.closed = true
	s.lksLk.Lock()
	s.bufLk.Lock()
	s.lks.ForEach(func(l interface{}, _ uint) {
		l.(*sync.Mutex).Unlock()
	})
	s.bufLk.Unlock()
	s.lksLk.Unlock()
	return s
}

func (s *Stream) forEach(f func(x interface{})) {
	for x := s.Pull(); x != EndMarker; x = s.Pull() {
		f(x)
	}
}

func (s *Stream) Map(f func(x interface{}) interface{}) *Stream {
	newS := New()
	go func() {
		s.forEach(func(x interface{}) { newS.PushBack(f(x)) })
		newS.Close()
	}()
	return newS
}

func (s *Stream) Stringify() *Stream {
	return s.Map(func(x interface{}) interface{} {
		switch x.(type) {
		case fmt.Stringer:
			return x.(fmt.Stringer).String()
		default:
			return fmt.Sprintf("%v", x)
		}
	})
}

func (s *Stream) ForEach(f func(x interface{})) *Stream {
	return s.Map(func(x interface{}) interface{} {
		f(x)
		return x
	})
}

func (s *Stream) Consume() {
	s.forEach(func(x interface{}) {})
}

func (s *Stream) Log(writer io.Writer, format string) *Stream {
	return s.ForEach(func(x interface{}) {
		if _, err := fmt.Fprintf(writer, format, x); err != nil {
			if writer != os.Stdout {
				fmt.Printf("[ERROR] %s", err.Error())
			}
			_, _ = fmt.Fprintf(writer, "[ERROR] %s", err.Error())
		}
	})
}

func (s *Stream) Printf(format string) *Stream {
	return s.Log(os.Stdout, format)
}

func (s *Stream) Println() *Stream {
	return s.Printf("%v\n")
}

func (s *Stream) Sprintf(format string) *Stream {
	return s.Map(func(x interface{}) interface{} { return fmt.Sprintf(format, x.(string)) })
}

func (s *Stream) Throttle(d time.Duration) *Stream {
	return s.Map(func(x interface{}) interface{} {
		time.Sleep(d)
		return x
	})
}

func (s *Stream) Spike(n uint, d time.Duration) *Stream {
	newS := New()
	go func() {
		for !s.Closed() {
			time.Sleep(d)
			i := uint(0)
			for x := s.Pull(); i < n && x != EndMarker; x = s.Pull() {
				newS.PushBack(x)
				i++
			}
		}
		newS.Close()
	}()
	return newS
}

func (s *Stream) Delay(d time.Duration) *Stream {
	newS := New()
	go func() {
		time.Sleep(d)
		s.forEach(func(x interface{}) { newS.PushBack(x) })
		newS.Close()
	}()
	return newS
}

func (s *Stream) Broadcast(ss ...*Stream) *Stream {
	return s.ForEach(func(x interface{}) {
		for _, s := range ss {
			s.PushBack(x)
		}
	})
}

func (s *Stream) Tee(n uint) []*Stream {
	ss := make([]*Stream, n)
	for i := uint(0); i < n; i++ {
		ss[i] = New()
	}
	s.Broadcast(ss...)
	return ss
}

func (s *Stream) Filter(f func(x interface{}) bool) *Stream {
	newS := New()
	go func() {
		s.forEach(func(x interface{}) {
			if f(x) {
				newS.PushBack(x)
			}
		})
		newS.Close()
	}()
	return newS
}

func (s *Stream) Flatten() *Stream {
	newS := New()
	go func() {
		s.forEach(func(innerStream interface{}) { innerStream.(*Stream).Pipe(newS) })
		newS.Close()
	}()
	return newS
}

func (s *Stream) FlattenDeep() *Stream {
	newS := New()
	go func() {
		s.forEach(func(x interface{}) {
			switch x.(type) {
			case *Stream:
				x.(*Stream).FlattenDeep().ForEach(func(y interface{}) { newS.PushBack(y) })
			default:
				newS.PushBack(x)
			}
		})
		newS.Close()
	}()
	return newS
}

func (s *Stream) FlatMap(f func(x interface{}) *Stream) *Stream {
	return s.Map(func(x interface{}) interface{} { return f(x) }).FlattenDeep()
}

func (s *Stream) TakeUntil(f func(x interface{}) bool) *Stream {
	newS := New()
	go func() {
		for x := s.Pull(); x != EndMarker && f(x); x = s.Pull() {
			newS.PushBack(x)
		}
		newS.Close()
	}()
	return newS
}

func (s *Stream) TakeWhile(f func(x interface{}) bool) *Stream {
	return s.TakeUntil(func(x interface{}) bool { return !f(x) })
}

func (s *Stream) Pipe(other *Stream) *Stream {
	go func() {
		s.forEach(func(x interface{}) { other.PushBack(x) })
		other.Close()
	}()
	return other
}

func (s *Stream) Closed() bool { return s.PeekFront() == EndMarker }

func (s *Stream) Take(n uint) *Stream {
	newS := New()
	go func() {
		for i := uint(0); i < n; i++ {
			x := s.Pull()
			if x == EndMarker {
				break
			}
			newS.PushBack(x)
		}
		newS.Close()
	}()
	return newS
}

func (s *Stream) Skip() *Stream {
	return s.SkipN(1)
}

func (s *Stream) SkipN(n uint) *Stream {
	newS := New()
	go func() {
		s.Take(n).PullN(n)
		s.Pipe(newS)
	}()
	return newS
}

func (s *Stream) Reduce(init interface{}, f func(acc interface{}, x interface{}) interface{}) interface{} {
	s.forEach(func(x interface{}) { init = f(init, x) })
	return init
}

func (s *Stream) Scan(init interface{}, f func(x interface{}, y interface{}) interface{}) *Stream {
	return s.Map(func(x interface{}) interface{} {
		defer func() { init = f(init, x) }()
		return init
	})
}

func (s *Stream) Count() uint {
	return s.Reduce(uint(0), func(x interface{}, y interface{}) interface{} { return x.(uint) + 1 }).(uint)
}

func (s *Stream) Sum() float64 {
	switch s.PeekFront().(type) {
	case int:
		return float64(s.Reduce(0, func(x interface{}, y interface{}) interface{} { return x.(int) + y.(int) }).(int))
	case int64:
		return float64(s.Reduce(int64(0), func(x interface{}, y interface{}) interface{} { return x.(int64) + y.(int64) }).(int64))
	case int32:
		return float64(s.Reduce(int32(0), func(x interface{}, y interface{}) interface{} { return x.(int32) + y.(int32) }).(int32))
	case float32:
		return float64(s.Reduce(float32(0), func(x interface{}, y interface{}) interface{} { return x.(float32) + y.(float32) }).(float32))
	case float64:
		return s.Reduce(float64(0), func(x interface{}, y interface{}) interface{} { return x.(float64) + y.(float64) }).(float64)
	case uint:
		return float64(s.Reduce(uint(0), func(x interface{}, y interface{}) interface{} { return x.(uint) + y.(uint) }).(uint))
	default:
		return 0
	}
}

func (s *Stream) Concat() string {
	return s.Reduce("", func(x interface{}, y interface{}) interface{} { return x.(string) + y.(string) }).(string)
}

func (s *Stream) Join(delim string) string {
	str := s.Reduce("", func(x interface{}, y interface{}) interface{} { return x.(string) + delim + y.(string) }).(string)
	return str
}

func (s *Stream) BufSize() uint {
	s.bufLk.Lock()
	size := s.buf.Size()
	s.bufLk.Unlock()
	return size
}

func (s *Stream) BufEmpty() bool {
	s.bufLk.Lock()
	isEmpty := s.buf.Size() == 0 || s.buf.All(func(z interface{}) bool { return z == EndMarker })
	s.bufLk.Unlock()
	return isEmpty
}

func (s *Stream) BufClear() { s.BufDrain() }

func (s *Stream) BufDrain() *list.ConcurrentList {
	s.bufLk.Lock()
	saveList := s.buf.TakeWhile(func(x interface{}) bool { return x != EndMarker })
	s.buf.Clear()
	s.bufLk.Unlock()
	return saveList
}

func (s *Stream) Eq(x interface{}) bool {
	switch x.(type) {
	case *Stream:
		return s == x.(*Stream)
	default:
		return false
	}
}

func (s *Stream) Clone() *Stream {
	s.lksLk.Lock()
	s.bufLk.Lock()
	defer s.lksLk.Unlock()
	defer s.bufLk.Unlock()
	return &Stream{
		closed: s.closed,
		bufLk:  &sync.Mutex{},
		lksLk:  &sync.Mutex{},
		buf:    s.buf.Clone(),
		lks:    s.lks.Clone(),
	}
}

func (s *Stream) String() string {
	parts := make([]string, 0)
	s.bufLk.Lock()
	s.buf.ForEach(func(x interface{}, idx uint) {
		switch x.(type) {
		case fmt.Stringer:
			parts[idx] = x.(fmt.Stringer).String()
		default:
			parts[idx] = fmt.Sprintf("%v", x)
		}
	})
	s.bufLk.Unlock()
	return fmt.Sprintf("|%s|", strings.Join(parts, " < "))
}
