package list

import (
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	N_WORKERS      int32 = 300
	SLEEP_DURATION       = 1 * time.Millisecond
)

type (
	ConcurrentList struct {
		fst  *node
		lst  *node
		size uint
		lk   *sync.Mutex
	}
)

func New(xs ...interface{}) *ConcurrentList {
	newList := &ConcurrentList{
		fst:  nil,
		lst:  nil,
		size: 0,
		lk:   &sync.Mutex{},
	}
	for _, x := range xs {
		newList.Append(x)
	}
	return newList
}

func Range(n int, m int) *ConcurrentList {
	xs := New()
	for i := n; i < m; i++ {
		xs.Append(i)
	}
	return xs
}

func Generate(n int, m int, f func(int) interface{}) *ConcurrentList {
	xs := New()
	for i := n; i < m; i++ {
		xs.Append(f(i))
	}
	return xs
}

func (xs *ConcurrentList) Tail() *ConcurrentList {
	if xs.Size() == 1 {
		return New()
	}
	xs.lk.Lock()
	defer xs.lk.Unlock()
	return &ConcurrentList{
		fst:  xs.fst.next,
		lst:  xs.lst,
		size: xs.size - 1,
		lk:   &sync.Mutex{},
	}
}

func (xs *ConcurrentList) Append(val interface{}) {
	newNode := &node{val: val, next: nil}
	xs.lk.Lock()
	defer xs.lk.Unlock()
	if xs.size == 0 {
		xs.fst = newNode
		xs.lst = newNode
	} else {
		xs.lst.next = newNode
		xs.lst = newNode
	}
	xs.size++
}

func (xs *ConcurrentList) Prepend(val interface{}) {
	xs.lk.Lock()
	defer xs.lk.Unlock()
	if xs.size == 0 {
		newNode := &node{val: val, next: nil}
		xs.fst = newNode
		xs.lst = newNode
	} else {
		xs.fst = &node{val: val, next: xs.fst}
	}
	xs.size++
}

func (xs *ConcurrentList) PeekFront() interface{} {
	xs.lk.Lock()
	defer xs.lk.Unlock()
	return xs.fst.val
}

func (xs *ConcurrentList) PeekBack() interface{} {
	xs.lk.Lock()
	defer xs.lk.Unlock()
	return xs.lst.val
}

func (xs *ConcurrentList) PopFront() interface{} {
	xs.lk.Lock()
	x := xs.fst.val
	xs.fst = xs.fst.next
	xs.size--
	if xs.fst == nil {
		xs.lst = nil
	}
	xs.lk.Unlock()
	return x
}

func (xs *ConcurrentList) Remove(pred func(interface{}, uint) bool) (interface{}, int) {
	idx := uint(0)
	var parent *node = nil
	xs.lk.Lock()
	defer xs.lk.Unlock()
	for focus := xs.fst; focus != nil; focus = focus.next {
		if pred(focus.val, idx) {
			if parent == nil {
				xs.size--
				xs.fst = xs.fst.next
				if xs.fst == nil {
					xs.lst = nil
				}
			} else {
				xs.size--
				parent.next = focus.next
			}
			return focus.val, int(idx)
		}
		idx++
		parent = focus
	}
	return nil, -1
}

func (xs *ConcurrentList) Find(pred func(interface{}, uint) bool) (interface{}, int) {
	idx := uint(0)
	xs.lk.Lock()
	defer xs.lk.Unlock()
	for focus := xs.fst; focus != nil; focus = focus.next {
		if pred(focus.val, idx) {
			return focus.val, int(idx)
		}
		idx++
	}
	return nil, -1
}

func (xs *ConcurrentList) ForEachParallel(f func(interface{}, uint)) {
	xs.lk.Lock()
	var idx uint32 = 0
	jobs := make([]*sync.Mutex, xs.size, xs.size)
	max := N_WORKERS
	sema := &max
	for focus := xs.fst; focus != nil; focus = focus.next {
		for atomic.LoadInt32(sema) <= int32(0) {
			time.Sleep(SLEEP_DURATION)
		}
		atomic.AddInt32(sema, int32(-1))
		m := &sync.Mutex{}
		m.Lock()
		jobs[idx] = m
		go func(val interface{}, idx uint32) {
			f(val, uint(idx))
			jobs[idx].Unlock()
			atomic.AddInt32(sema, int32(1))
		}(focus.val, idx)
		idx++
	}
	xs.lk.Unlock()
	for i := uint(0); i < xs.size; i++ {
		jobs[i].Lock()
		jobs[i].Unlock()
	}
}

func (xs *ConcurrentList) MapParallelInPlace(f func(interface{}, uint) interface{}) {
	var idx uint32 = 0
	xs.lk.Lock()
	jobs := make([]*sync.Mutex, xs.size, xs.size)
	max := N_WORKERS
	sema := &max
	for focus := xs.fst; focus != nil; focus = focus.next {
		for atomic.LoadInt32(sema) <= int32(0) {
			time.Sleep(SLEEP_DURATION)
		}
		atomic.AddInt32(sema, int32(-1))
		m := &sync.Mutex{}
		m.Lock()
		jobs[idx] = m
		go func(focus *node, idx uint32) {
			focus.val = f(focus.val, uint(idx))
			jobs[idx].Unlock()
			atomic.AddInt32(sema, int32(1))
		}(focus, idx)
		idx++
	}
	xs.lk.Unlock()
	for i := uint(0); i < xs.size; i++ {
		jobs[i].Lock()
		jobs[i].Unlock()
	}
}

func (xs *ConcurrentList) Nth(n uint) interface{} {
	if n == 0 {
		return xs.PeekFront()
	}
	xs.lk.Lock()
	defer xs.lk.Unlock()
	focus := xs.fst
	for idx := uint(0); idx < n; idx++ {
		focus = focus.next
	}
	return focus.val
}

func (xs *ConcurrentList) Clear() {
	xs.lk.Lock()
	defer xs.lk.Unlock()
	xs.size = 0
	xs.fst = nil
	xs.lst = nil
}

func (xs *ConcurrentList) Empty() bool {
	return xs.Size() == 0
}

func (xs *ConcurrentList) Size() uint {
	xs.lk.Lock()
	defer xs.lk.Unlock()
	return xs.size
}

func (xs *ConcurrentList) Reduce(init interface{}, f func(x interface{}, y interface{}, idx uint) interface{}) interface{} {
	var idx uint = 0
	xs.lk.Lock()
	defer xs.lk.Unlock()
	for focus := xs.fst; focus != nil; focus = focus.next {
		init = f(init, focus.val, idx)
		idx++
	}
	return init
}

func (xs *ConcurrentList) Eq(_ys interface{}) bool {
	switch _ys.(type) {
	case *ConcurrentList:
		ys := _ys.(*ConcurrentList)
		xs.lk.Lock()
		defer xs.lk.Unlock()
		focus := xs.fst
		focus2 := ys.fst
		for {
			if focus == nil {
				if focus2 == nil {
					break
				}
				return false
			}

			if focus2 == nil {
				if focus == nil {
					break
				}
				return false
			}

			if focus.val != focus2.val {
				return false
			}

			focus = focus.next
			focus2 = focus2.next
		}
		return true
	default:
		return false
	}
}

func (xs *ConcurrentList) String() string {
	n := xs.Size()
	s := make([]string, n, n)
	xs.ForEachParallel(func(x interface{}, idx uint) {
		switch x.(type) {
		case fmt.Stringer:
			s[idx] = x.(fmt.Stringer).String()
		default:
			s[idx] = fmt.Sprintf("%v", x)
		}
	})
	return fmt.Sprintf("[%s]", strings.Join(s, " "))
}
