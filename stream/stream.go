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

func From(xs []interface{}) *Stream { return New(xs...).Close() }

func Generate(n int, m int, step int, f func(n int) interface{}) *Stream {
	s := New()
	for start := n; start < m; start += step {
		s.PushBack(f(start))
	}
	return s.Close()
}

func Range(n int, m int, step int) *Stream {
	return Generate(n, m, step, func(x int) interface{} { return x })
}

func Ints(min int, max int, n int) *Stream {
	r := max - min
	return Generate(0, n, 1, func(x int) interface{} { return min + rand.Intn(r) })
}

func Bytes(n int) *Stream {
	return Generate(0, n, 1, func(x int) interface{} { return byte(rand.Intn(256)) })
}

func Floats(n int) *Stream {
	return Generate(0, n, 1, func(x int) interface{} { return rand.Float64() })
}

func (s *Stream) Map(f func(x interface{}) interface{}) *Stream {
	newS := New()
	go func() {
		for x := s.Pull(); x != EndMarker; x = s.Pull() {
			newS.PushBack(f(x))
		}
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

func (s *Stream) Delay(d time.Duration) *Stream {
	newS := New()
	go func() {
		time.Sleep(d)
		for x := s.Pull(); x != EndMarker; x = s.Pull() {
			newS.PushBack(x)
		}
		newS.Close()
	}()
	return newS
}

func (s *Stream) Broadcast(fst *Stream, ss ...*Stream) {
	go func() {
		for x := s.Pull(); x != EndMarker; x = s.Pull() {
			fst.PushBack(x)
			for _, ssEl := range ss {
				ssEl.PushBack(x)
			}
		}
		fst.Close()
		for _, ssEl := range ss {
			ssEl.Close()
		}
	}()
}

func (s *Stream) Reduce(init interface{}, f func(x interface{}, y interface{}) interface{}) *Stream {
	newS := New()
	go func() {
		for x := s.Pull(); x != EndMarker; x = s.Pull() {
			init = f(init, x)
		}
		newS.PushBack(init)
		newS.Close()
	}()
	return newS
}

func (s *Stream) Scan(init interface{}, f func(x interface{}, y interface{}) interface{}) *Stream {
	newS := New()
	go func() {
		for x := s.Pull(); x != EndMarker; x = s.Pull() {
			init = f(init, x)
			newS.PushBack(init)
		}
		newS.Close()
	}()
	return newS
}

func (s *Stream) Count() uint {
	return s.Reduce(uint(0), func(x interface{}, y interface{}) interface{} { return x.(uint) + 1 }).Pull().(uint)
}

func (s *Stream) Sum() int {
	return s.Reduce(0, func(x interface{}, y interface{}) interface{} { return x.(int) + y.(int) }).Pull().(int)
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

func (s *Stream) Connect(other *Stream) *Stream {
	go func() {
		for x := s.Pull(); x != EndMarker; x = s.Pull() {
			other.PushBack(x)
		}
		other.Close()
	}()
	return other
}

func Pipeline(s *Stream, ss ...*Stream) *Stream {
	acc := s
	end := uint(len(ss))
	for start := uint(0); start < end; start++ {
		acc = acc.Connect(ss[start])
	}
	return acc
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
		l := s.lks.PopFront()
		l.(*sync.Mutex).Unlock()
	}
	s.lksLk.Unlock()
	s.bufLk.Unlock()
}

func (s *Stream) BufSize() uint {
	s.bufLk.Lock()
	size := s.buf.Size()
	s.bufLk.Unlock()
	return size
}

func (s *Stream) Empty() bool {
	s.bufLk.Lock()
	isEmpty := s.buf.Size() == 0 || s.buf.All(func(z interface{}) bool { return z == EndMarker })
	s.bufLk.Unlock()
	return isEmpty
}

func (s *Stream) Pull() interface{} {
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
		s.bufLk.Lock()
	}
	front := s.buf.PopFront()
	s.bufLk.Unlock()
	return front
}

func (s *Stream) PullN(n uint) []interface{} {
	xs := make([]interface{}, n)
	s.bufLk.Lock()
	s.lksLk.Lock()
	defer s.bufLk.Unlock()
	defer s.lksLk.Unlock()
	for i := uint(0); i < n; i++ {
		if s.buf.Empty() {
			if s.closed {
				xs[i] = EndMarker
				continue
			}
			l := &sync.Mutex{}
			l.Lock()
			s.lks.Append(l)
			l.Lock()
			l.Unlock()
		}
		xs[i] = s.buf.PopFront()
	}
	return xs
}

func (s *Stream) PullAll() *list.ConcurrentList {
	xs := list.New()
	for x := s.Pull(); x != EndMarker; x = s.Pull() {
		xs.Append(x)
	}
	return xs
}

func (s *Stream) Peek() interface{} {
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
		s.bufLk.Lock()
	}
	front := s.buf.PeekFront()
	s.bufLk.Unlock()
	return front
}

func (s *Stream) Close() *Stream {
	s.closed = true
	s.lksLk.Lock()
	s.bufLk.Lock()
	s.lks.ForEachParallel(func(l interface{}, _ uint) {
		s.buf.Append(EndMarker)
		l.(*sync.Mutex).Unlock()
	})
	s.bufLk.Unlock()
	s.lksLk.Unlock()
	return s
}

func (s *Stream) Clear() {
	s.bufLk.Lock()
	s.buf.Clear()
	s.bufLk.Unlock()
}

func (s *Stream) Drain() *list.ConcurrentList {
	s.bufLk.Lock()
	defer s.bufLk.Unlock()
	saveList := s.buf.TakeWhile(func(x interface{}) bool { return x != EndMarker })
	s.buf = list.New()
	return saveList
}

func (s *Stream) Skip(n uint) *Stream {
	newS := New()
	go func() {
		for i := uint(0); i < n; i++ {
			s.Pull()
		}
		for x := s.Pull(); x != EndMarker; x = s.Pull() {
			newS.PushBack(x)
		}
		newS.Close()
	}()
	return newS
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
	pLk := &sync.RWMutex{}
	s.bufLk.Lock()
	s.buf.ForEachParallel(func(x interface{}, idx uint) {
		go func(x interface{}, idx int) {
			pLk.Lock()
			for idx >= len(parts) {
				parts = append(parts, "")
			}
			pLk.Unlock()
			switch x.(type) {
			case fmt.Stringer:
				parts[idx] = x.(fmt.Stringer).String()
			default:
				parts[idx] = fmt.Sprintf("%v", x)
			}
		}(x, int(idx))
		idx++
	})
	s.bufLk.Unlock()
	return fmt.Sprintf("|%s|", strings.Join(parts, " < "))
}
