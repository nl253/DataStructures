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

func (s Stream) PushFront(t interface{}) {
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

func (s Stream) PushBack(x interface{}) {
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

func (s Stream) Pull() interface{} {
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

func (s Stream) Peek() interface{} {
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
