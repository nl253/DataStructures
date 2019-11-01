package list

import (
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type (
	LookupList struct {
		fst  *node
		lst  *node
		size uint
		lk   *sync.Mutex
	}
)

func NewLookup(xs ...interface{}) *LookupList {
	newList := &LookupList{
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

func GenerateLookup(n int, m int, f func(int) interface{}) *LookupList {
	xs := NewLookup()
	for i := n; i < m; i++ {
		xs.Append(f(i))
	}
	return xs
}

func RangeLookup(n int, m int) *LookupList {
	xs := NewLookup()
	for i := n; i < m; i++ {
		xs.Append(i)
	}
	return xs
}

func (xs *LookupList) Tail() *LookupList {
	if xs.Size() == 1 {
		return NewLookup()
	}
	xs.lk.Lock()
	defer xs.lk.Unlock()
	return &LookupList{
		fst:  xs.fst.next,
		lst:  xs.lst,
		size: xs.size - 1,
		lk:   &sync.Mutex{},
	}
}

func (xs *LookupList) Append(val interface{}) {
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

func (xs *LookupList) Prepend(val interface{}) {
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

func (xs *LookupList) PeekFront() interface{} {
	xs.lk.Lock()
	defer xs.lk.Unlock()
	return xs.fst.val
}

func (xs *LookupList) PeekBack() interface{} {
	xs.lk.Lock()
	defer xs.lk.Unlock()
	return xs.lst.val
}

func (xs *LookupList) PopFront() interface{} {
	xs.lk.Lock()
	defer xs.lk.Unlock()
	x := xs.fst.val
	xs.fst = xs.fst.next
	xs.size--
	if xs.fst == nil {
		xs.lst = nil
	}
	return x
}

func (xs *LookupList) Find(pred func(interface{}, uint) bool) (interface{}, bool) {
	idx := uint(0)
	var parent *node = nil
	xs.lk.Lock()
	for focus := xs.fst; focus != nil; focus = focus.next {
		if pred(focus.val, idx) {
			if parent != nil {
				parent.next = focus.next
				xs.size--
				xs.lk.Unlock()
				xs.Prepend(focus.val)
				return focus.val, true
			}
			xs.lk.Unlock()
			return focus.val, true
		}
		idx++
		parent = focus
	}
	xs.lk.Unlock()
	return nil, false
}

func (xs *LookupList) Remove(pred func(interface{}, uint) bool) (interface{}, int) {
	idx := uint(0)
	var parent *node = nil
	xs.lk.Lock()
	defer xs.lk.Unlock()
	for focus := xs.fst; focus != nil; focus = focus.next {
		if pred(focus.val, idx) {
			if parent == nil {
				xs.fst = xs.fst.next
				if xs.fst == nil {
					xs.lst = nil
				}
			} else {
				parent.next = focus.next
			}
			xs.size--
			return focus.val, int(idx)
		}
		idx++
		parent = focus
	}
	return nil, -1
}

func (xs *LookupList) ForEachParallel(f func(interface{}, uint)) {
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

func (xs *LookupList) MapParallelInPlace(f func(interface{}, uint) interface{}) {
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
		go func(node *node, idx uint32) {
			node.val = f(node.val, uint(idx))
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

func (xs *LookupList) Nth(n uint) interface{} {
	if n == 0 {
		return xs.PeekFront()
	}
	xs.lk.Lock()
	element := xs.fst
	var parent *node
	for idx := uint(0); idx < n; idx++ {
		parent = element
		element = element.next
	}
	parent.next = element.next
	xs.lk.Unlock()
	xs.Prepend(element.val)
	return element.val
}

func (xs *LookupList) Clear() {
	xs.lk.Lock()
	defer xs.lk.Unlock()
	xs.size = 0
	xs.fst = nil
	xs.lst = nil
}

func (xs *LookupList) Empty() bool {
	return xs.Size() == 0
}

func (xs *LookupList) Size() uint {
	xs.lk.Lock()
	defer xs.lk.Unlock()
	return xs.size
}

func (xs *LookupList) String() string {
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

func (xs *LookupList) Eq(_ys interface{}) bool {
	switch _ys.(type) {
	case *LookupList:
		ys := _ys.(*LookupList)
		if xs.Size() != ys.Size() {
			return false
		}
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
