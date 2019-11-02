package stream

import (
	"fmt"
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
	newS := New()
	go func() {
		for {
			x := s.Pull()
			if x == EndMarker {
				newS.Close()
				return
			}
			f(x)
			newS.PushBack(x)
		}
	}()
	return newS
}

func (s *Stream) Throttle(d time.Duration) *Stream {
	newS := New()
	go func() {
		for {
			x := s.Pull()
			if x == EndMarker {
				newS.Close()
				return
			}
			time.Sleep(d)
			newS.PushBack(x)
		}
	}()
	return newS
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
			for _, ssEl := range ss {
				fst.PushBack(x)
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

func (s *Stream) Count(init interface{}, f func(x interface{}, y interface{}) interface{}) uint {
	return s.Reduce(0, func(x interface{}, y interface{}) interface{} {
		return x.(uint) + 1
	}).Pull().(uint)
}

func (s *Stream) Sum(init interface{}, f func(x interface{}, y interface{}) interface{}) int {
	return s.Reduce(0, func(x interface{}, y interface{}) interface{} {
		return x.(int) + y.(int)
	}).Pull().(int)
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
	lk := sync.RWMutex{}
	s.lk.Lock()
	xs := make([]string, s.buf.Size())
	s.buf.ForEachParallel(func(x interface{}, idx uint) {
		lk.RLock()
		for idx >= uint(len(xs)) {
			lk.RUnlock()
			lk.Lock()
			xs = append(xs, "")
			lk.Unlock()
			lk.RLock()
		}
		switch x.(type) {
		case fmt.Stringer:
			xs[idx] = x.(fmt.Stringer).String()
		default:
			xs[idx] = fmt.Sprintf("%v", x)
		}
		lk.RUnlock()
	})
	s.lk.Unlock()
	return fmt.Sprintf("|%s|", strings.Join(xs, " < "))
}
