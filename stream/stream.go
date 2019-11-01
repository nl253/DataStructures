package stream

import (
	"fmt"
	"strings"
	"sync"

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

func New() *Stream {
	return &Stream{
		lk:     &sync.Mutex{},
		buf:    list.New(),
		lks:    list.New(),
		closed: false,
	}
}

func (s *Stream) Map(f func(x interface{}) interface{}) *Stream {
	s.lk.Lock()
	newS := New()
	go func() {
		for !s.closed {
			newS.PushBack(f(s.Pull()))
		}
		s.lk.Unlock()
	}()
	return newS
}

func (s *Stream) Filter(f func(x interface{}) bool) *Stream {
	s.lk.Lock()
	newS := New()
	go func() {
		for {
			x := s.Pull()
			if x == EndMarker {
				break
			}
			if f(x) {
				newS.PushBack(x)
			}
		}
		newS.Close()
		s.lk.Unlock()
	}()
	return newS
}

func (s *Stream) Connect(other *Stream) *Stream {
	s.lk.Lock()
	other.lk.Lock()
	go func() {
		for {
			x := s.Pull()
			if x == EndMarker {
				break
			}
			other.PushBack(x)
		}
		other.lk.Unlock()
		other.Close()
		s.lk.Unlock()
	}()
	return other
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
	s.lks.ForEachParallel(func(l interface{}, u uint) {
		s.PushBack(EndMarker)
		l.(*sync.Mutex).Unlock()
	})
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
	s.lk.Lock()
	lk := sync.RWMutex{}
	xs := make([]string, s.buf.Size())
	s.buf.ForEachParallel(func(i interface{}, u uint) {
		lk.RLock()
		for u >= uint(len(xs)) {
			lk.RUnlock()
			lk.Lock()
			xs = append(xs, "")
			lk.Unlock()
			lk.RLock()
		}
		switch i.(type) {
		case fmt.Stringer:
			xs[u] = i.(fmt.Stringer).String()
		default:
			xs[u] = fmt.Sprintf("%v", i)
		}
		lk.RUnlock()
	})
	s.lk.Unlock()
	return fmt.Sprintf("| %s |", strings.Join(xs, " < "))
}
