package list

import (
	"fmt"
	"strings"
	"sync"
)

type (
	ConcurrentList struct {
		fst  *node
		lst  *node
		size uint
		lk   *sync.RWMutex
	}
)

func New(xs ...interface{}) *ConcurrentList {
	newList := &ConcurrentList{
		fst:  nil,
		lst:  nil,
		size: 0,
		lk:   &sync.RWMutex{},
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
	xs.lk.RLock()
	defer xs.lk.RUnlock()
	return &ConcurrentList{
		fst:  xs.fst.next,
		lst:  xs.lst,
		size: xs.size - 1,
		lk:   &sync.RWMutex{},
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
	xs.lk.RLock()
	front := xs.fst.val
	xs.lk.RUnlock()
	return front
}

func (xs *ConcurrentList) PeekBack() interface{} {
	xs.lk.RLock()
	back := xs.lst.val
	xs.lk.RUnlock()
	return back
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
	xs.lk.RLock()
	defer xs.lk.RUnlock()
	for focus := xs.fst; focus != nil; focus = focus.next {
		if pred(focus.val, idx) {
			return focus.val, int(idx)
		}
		idx++
	}
	return nil, -1
}

func (xs *ConcurrentList) ForEachParallel(f func(interface{}, uint)) {
	xs.lk.RLock()
	var idx uint32 = 0
	wg := &sync.WaitGroup{}
	wg.Add(int(xs.size))
	for focus := xs.fst; focus != nil; focus = focus.next {
		go func(val interface{}, idx uint32) {
			f(val, uint(idx))
			wg.Done()
		}(focus.val, idx)
		idx++
	}
	xs.lk.RUnlock()
	wg.Wait()
}

func (xs *ConcurrentList) MapParallelInPlace(f func(interface{}, uint) interface{}) {
	var idx uint32 = 0
	xs.lk.Lock()
	wg := sync.WaitGroup{}
	wg.Add(int(xs.size))
	for focus := xs.fst; focus != nil; focus = focus.next {
		go func(focus *node, idx uint32) {
			focus.val = f(focus.val, uint(idx))
			wg.Done()
		}(focus, idx)
		idx++
	}
	wg.Wait()
	xs.lk.Unlock()
}

func (xs *ConcurrentList) Nth(n uint) interface{} {
	if n == 0 {
		return xs.PeekFront()
	}
	xs.lk.RLock()
	defer xs.lk.RUnlock()
	focus := xs.fst.next
	for idx := uint(1); idx < n; idx++ {
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
	xs.lk.RLock()
	empty := xs.size == 0
	xs.lk.RUnlock()
	return empty
}

func (xs *ConcurrentList) Size() uint {
	xs.lk.RLock()
	size := xs.size
	xs.lk.RUnlock()
	return size
}

func (xs *ConcurrentList) Reduce(init interface{}, f func(x interface{}, y interface{}, idx uint) interface{}) interface{} {
	var idx uint = 0
	xs.lk.RLock()
	defer xs.lk.RUnlock()
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
		if xs.Size() != ys.Size() {
			return false
		}
		xs.lk.RLock()
		defer xs.lk.RUnlock()
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
	lk := sync.RWMutex{}
	n := xs.Size()
	s := make([]string, n)
	xs.ForEachParallel(func(x interface{}, idx uint) {
		lk.RLock()
		if idx >= uint(len(s)) {
			lk.RUnlock()
			lk.Lock()
			s = append(s, "")
			lk.Unlock()
			lk.RLock()
		}
		switch x.(type) {
		case fmt.Stringer:
			s[idx] = x.(fmt.Stringer).String()
		default:
			s[idx] = fmt.Sprintf("%v", x)
		}
		lk.RUnlock()
	})
	return fmt.Sprintf("[%s]", strings.Join(s, " "))
}
