package stream

import (
	"fmt"
	"math/rand"
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

func From(xs []interface{}) *Stream { return New(xs...) }

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

func Range(n int, m int, step int) *Stream {
	return Generate(n, m, step, func(x int) interface{} { return x })
}

func Ints(min int, max int, n uint) *Stream {
	r := max - min
	return Generate(0, int(n), 1, func(x int) interface{} { return min + rand.Intn(r) })
}

func Bytes(n int) *Stream {
	return Generate(0, n, 1, func(x int) interface{} { return byte(rand.Intn(256)) })
}

func Floats(n int) *Stream {
	return Generate(0, n, 1, func(x int) interface{} { return rand.Float64() })
}

func Replicate(n uint, xs ...interface{}) *Stream {
	return Repeat(n, From(xs)).Flatten()
}

func Repeat(n uint, x interface{}) *Stream {
	return Generate(int(0), int(n), 1, func(n int) interface{} {
		return x
	})
}

func Tick(freq time.Duration, x interface{}) *Stream {
	return Emit(freq, func(_ uint) interface{} { return x })
}

func Emit(freq time.Duration, f func(n uint) interface{}) *Stream {
	s := New()
	go func() {
		for n := uint(0); ; n++ {
			s.PushBack(f(n))
			time.Sleep(freq)
		}
	}()
	return s
}

func (s *Stream) PushFront(t interface{}) {
	s.bufLk.Lock()
	s.buf.Prepend(t)
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
	s.buf.Append(x)
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
			s.lks.Append(l)
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
		acc.(*list.ConcurrentList).Append(x)
		return acc
	}).(*list.ConcurrentList)
}

func (s *Stream) Peek() interface{} {
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
			s.lks.Append(l)
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

func (s *Stream) ForEach(f func(x interface{})) {
	for x := s.Pull(); x != EndMarker; x = s.Pull() {
		f(x)
	}
}

func (s *Stream) Map(f func(x interface{}) interface{}) *Stream {
	newS := New()
	go func() {
		s.ForEach(func(x interface{}) { newS.PushBack(f(x)) })
		newS.Close()
	}()
	return newS
}

func (s *Stream) Tap(f func(x interface{})) *Stream {
	return s.Map(func(x interface{}) interface{} {
		f(x)
		return x
	})
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
	return Pipeline(New(EndMarker).Throttle(d), s)
}

func (s *Stream) Broadcast(ss ...*Stream) *Stream {
	return s.Map(func(x interface{}) interface{} {
		for _, s := range ss {
			s.PushBack(x)
		}
		return x
	})
}

func (s *Stream) Duplicate(n uint) []*Stream {
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
		for x := s.Pull(); x != EndMarker; x = s.Pull() {
			if f(x) {
				newS.PushBack(x)
			}
		}
		newS.Close()
	}()
	return newS
}

func (s *Stream) Flatten() *Stream {
	newS := New()
	go func() {
		s.ForEach(func(x interface{}) {
			x.(*Stream).ForEach(func(y interface{}) {
				newS.PushBack(y)
			})
		})
		newS.Close()
	}()
	return newS
}

func (s *Stream) FlattenDeep() *Stream {
	newS := New()
	go func() {
		s.ForEach(func(x interface{}) {
			switch x.(type) {
			case *Stream:
				x.(*Stream).ForEach(func(y interface{}) { s.PushBack(y) })
			default:
				newS.PushBack(x)
			}
		})
		newS.Close()
	}()
	return newS
}

func (s *Stream) FlatMap(f func(x interface{}) *Stream) *Stream {
	return s.Map(func(x interface{}) interface{} { return f(x) }).Flatten()
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

func (s *Stream) pipe(other *Stream) *Stream {
	go func() {
		s.ForEach(func(x interface{}) { other.PushBack(x) })
		other.Close()
	}()
	return other
}

func Pipeline(ss ...*Stream) *Stream {
	acc := New()
	for _, s := range ss {
		acc.PushBack(s)
	}
	return acc.Reduce(New(), func(acc interface{}, _other interface{}) interface{} {
		other := _other.(*Stream)
		go func() {
			acc.(*Stream).ForEach(func(x interface{}) { other.PushBack(x) })
			other.Close()
		}()
		return other
	}).(*Stream)
}

func (s *Stream) Closed() bool { return s.Peek() == EndMarker }

func (s *Stream) Take(n uint) *Stream {
	newS := New()
	go func() {
		var x interface{}
		for i := uint(0); i < n; i++ {
			x = s.Pull()
			if x == EndMarker {
				break
			}
			newS.PushBack(x)
		}
		newS.Close()
	}()
	return newS
}

func (s *Stream) Skip(n uint) *Stream {
	newS := New()
	go func() {
		s.Take(n).PullAll()
		s.ForEach(func(x interface{}) { newS.PushBack(x) })
		newS.Close()
	}()
	return newS
}

func (s *Stream) Reduce(init interface{}, f func(acc interface{}, x interface{}) interface{}) interface{} {
	s.ForEach(func(x interface{}) { init = f(init, x) })
	return init
}

func (s *Stream) Scan(init interface{}, f func(x interface{}, y interface{}) interface{}) *Stream {
	return s.Map(func(x interface{}) interface{} {
		init = f(init, x)
		return init
	})
}

func (s *Stream) Count() uint {
	return s.Reduce(uint(0), func(x interface{}, y interface{}) interface{} { return x.(uint) + 1 }).(uint)
}

func (s *Stream) Sum() int {
	return s.Reduce(0, func(x interface{}, y interface{}) interface{} { return x.(int) + y.(int) }).(int)
}

func (s *Stream) Concat() string {
	return s.Reduce("", func(x interface{}, y interface{}) interface{} { return x.(string) + y.(string) }).(string)
}

func (s *Stream) Join(delim string) string {
	str := s.Reduce("", func(x interface{}, y interface{}) interface{} { return x.(string) + delim }).(string)
	return str[:len(str)-len(delim)]
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
