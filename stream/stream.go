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
	lk     *sync.Mutex
	buf    *list.ConcurrentList
	lks    *list.ConcurrentList
}

type streamEnd struct{}

var EndMarker = &streamEnd{}

func New(xs ...interface{}) *Stream {
	s := &Stream{
		lk:     &sync.Mutex{},
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
	for start := n; start < m; start += step {
		s.PushBack(f(start))
	}
	return s
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
		for {
			x := s.Pull()
			if x == EndMarker {
				newS.Close()
				return
			}
			newS.PushBack(f(x))
		}
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
		for {
			x := s.Pull()
			if x == EndMarker {
				newS.Close()
				return
			}
			newS.PushBack(x)
		}
	}()
	return newS
}

func (s *Stream) Broadcast(fst *Stream, ss ...*Stream) {
	go func() {
		for {
			x := s.Pull()
			if x == EndMarker {
				fst.Close()
				for _, ssEl := range ss {
					ssEl.Close()
				}
				return
			}
			fst.PushBack(x)
			for _, ssEl := range ss {
				ssEl.PushBack(x)
			}
		}
	}()
}

func (s *Stream) Reduce(init interface{}, f func(x interface{}, y interface{}) interface{}) *Stream {
	newS := New()
	go func() {
		for {
			x := s.Pull()
			if x == EndMarker {
				newS.PushBack(init)
				newS.Close()
				return
			}
			init = f(init, x)
		}
	}()
	return newS
}

func (s *Stream) Scan(init interface{}, f func(x interface{}, y interface{}) interface{}) *Stream {
	newS := New()
	go func() {
		for {
			x := s.Pull()
			if x == EndMarker {
				newS.Close()
				return
			}
			init = f(init, x)
			newS.PushBack(init)
		}
	}()
	return newS
}

func (s *Stream) Count(init interface{}, f func(x interface{}, y interface{}) interface{}) uint {
	return s.Reduce(0, func(x interface{}, y interface{}) interface{} { return x.(uint) + 1 }).Pull().(uint)
}

func (s *Stream) Sum(init interface{}, f func(x interface{}, y interface{}) interface{}) int {
	return s.Reduce(0, func(x interface{}, y interface{}) interface{} { return x.(int) + y.(int) }).Pull().(int)
}

func (s *Stream) Filter(f func(x interface{}) bool) *Stream {
	newS := New()
	go func() {
		for {
			x := s.Pull()
			if x == EndMarker {
				newS.Close()
				return
			}
			if f(x) {
				newS.PushBack(x)
			}
		}
	}()
	return newS
}

func (s *Stream) TakeUntil(f func(x interface{}) bool) *Stream {
	newS := New()
	go func() {
		for {
			x := s.Pull()
			if x == EndMarker || !f(x) {
				newS.Close()
				return
			}
			newS.PushBack(x)
		}
	}()
	return newS
}

func (s *Stream) TakeWhile(f func(x interface{}) bool) *Stream {
	return s.TakeUntil(func(x interface{}) bool { return !f(x) })
}

func (s *Stream) Connect(other *Stream) *Stream {
	go func() {
		for {
			x := s.Pull()
			if x == EndMarker {
				other.Close()
				return
			}
			other.PushBack(x)
		}
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
	s.lk.Lock()
	s.buf.Prepend(t)
	if !s.lks.Empty() {
		l := s.lks.PopFront()
		s.lk.Unlock()
		l.(*sync.Mutex).Unlock()
	} else {
		s.lk.Unlock()
	}
}

func (s *Stream) PushBack(x interface{}) {
	s.lk.Lock()
	s.buf.Append(x)
	if !s.lks.Empty() {
		l := s.lks.PopFront()
		s.lk.Unlock()
		l.(*sync.Mutex).Unlock()
	} else {
		s.lk.Unlock()
	}
}

func (s *Stream) Pull() interface{} {
	s.lk.Lock()
	if s.buf.Empty() {
		if s.closed {
			s.lk.Unlock()
			return EndMarker
		}
		l := &sync.Mutex{}
		l.Lock()
		s.lks.Append(l)
		s.lk.Unlock()
		l.Lock()
		l.Unlock()
		s.lk.Lock()
	}
	front := s.buf.PopFront()
	s.lk.Unlock()
	return front
}

func (s *Stream) Peek() interface{} {
	s.lk.Lock()
	if s.buf.Empty() {
		if s.closed {
			s.lk.Unlock()
			return EndMarker
		}
		l := &sync.Mutex{}
		l.Lock()
		s.lks.Append(l)
		s.lk.Unlock()
		l.Lock()
		l.Unlock()
		s.lk.Lock()
	}
	front := s.buf.PeekFront()
	s.lk.Unlock()
	return front
}

func (s *Stream) Close() {
	s.lk.Lock()
	s.closed = true
	s.lk.Unlock()
	s.lks.ForEachParallel(func(l interface{}, _ uint) {
		s.PushBack(EndMarker)
	})
	s.lks.ForEachParallel(func(l interface{}, _ uint) {
		l.(*sync.Mutex).Unlock()
	})
}

func (s *Stream) Clear() {
	s.lk.Lock()
	s.buf.Clear()
	s.lk.Unlock()
}

func (s *Stream) Drain() *list.ConcurrentList {
	xs := list.New()
	for {
		x := s.Pull()
		if x == EndMarker {
			return xs
		}
		xs.Append(x)
	}
}

func (s *Stream) Skip(n uint) {
	for i := uint(0); i < n; i++ {
		s.Pull()
	}
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
	s.lk.Lock()
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
	s.lk.Unlock()
	return fmt.Sprintf("|%s|", strings.Join(parts, " < "))
}
