package stream

import (
	"time"

	"Lisp-GO/datastructures/list"
)

const (
	BASE_W8_TM                      = 10 * time.Millisecond
	W8_DELTA                        = 10 * time.Millisecond
	W8_MAX_MULTIPLIER time.Duration = 100
)

type Stream struct {
	closed bool
	buf    *list.ConcurrentList
}

type streamEnd struct{}

var EndMarker = &streamEnd{}

func New() *Stream {
	return &Stream{buf: list.New(), closed: false}
}

func (s Stream) PushFront(t interface{}) {
	s.buf.Prepend(t)
}

func (s Stream) PushBack(x interface{}) {
	s.buf.Append(x)
}

func (s Stream) Pull() interface{} {
	var waited time.Duration = 0
	for s.buf.Empty() {
		if s.closed {
			return EndMarker
		} else {
			if waited > W8_MAX_MULTIPLIER {
				waited = W8_MAX_MULTIPLIER
			}
			time.Sleep(BASE_W8_TM + waited*W8_DELTA)
			waited++
		}
	}
	return s.buf.PopFront()
}

func (s Stream) Peek() interface{} {
	var waited time.Duration = 0
	for s.buf.Empty() {
		if s.closed {
			return EndMarker
		} else {
			if waited > W8_MAX_MULTIPLIER {
				waited = W8_MAX_MULTIPLIER
			}
			time.Sleep(BASE_W8_TM + waited*W8_DELTA)
			waited++
		}
	}
	return s.buf.PeekFront()
}

func (s *Stream) Close() {
	s.closed = true
}
