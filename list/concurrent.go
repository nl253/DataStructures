package list

import (
	"fmt"
	"math/rand"
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

func Range(n int, m int, step int) *ConcurrentList {
	xs := New()
	for i := n; i < m; i += step {
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

func Ints(min int, max int, n uint) *ConcurrentList {
	r := max - min
	return Generate(0, int(n), func(_ int) interface{} { return min + rand.Intn(r) })
}

func Floats(n uint) *ConcurrentList {
	return Generate(0, int(n), func(_ int) interface{} { return rand.Float64() })
}

func Bytes(n uint) *ConcurrentList {
	return Generate(0, int(n), func(_ int) interface{} { return byte(rand.Intn(256)) })
}

func Chars(n uint) *ConcurrentList {
	return Generate(0, int(n), func(_ int) interface{} { return rand.Int31() })
}

func (xs *ConcurrentList) Tail() *ConcurrentList {
	xs.lk.RLock()
	if xs.size == 1 {
		xs.lk.RUnlock()
		return New()
	}
	tail := &ConcurrentList{
		fst:  xs.fst.next,
		lst:  xs.lst,
		size: xs.size - 1,
		lk:   &sync.RWMutex{},
	}
	xs.lk.RUnlock()
	return tail
}

func (xs *ConcurrentList) Append(val interface{}) {
	newNode := &node{val: val, next: nil}
	xs.lk.Lock()
	if xs.size == 0 {
		xs.fst = newNode
		xs.lst = newNode
	} else {
		xs.lst.next = newNode
		xs.lst = newNode
	}
	xs.size++
	xs.lk.Unlock()
}

func (xs *ConcurrentList) Prepend(val interface{}) {
	xs.lk.Lock()
	if xs.size == 0 {
		newNode := &node{val: val, next: nil}
		xs.fst = newNode
		xs.lst = newNode
	} else {
		xs.fst = &node{val: val, next: xs.fst}
	}
	xs.size++
	xs.lk.Unlock()
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

func (xs *ConcurrentList) RemoveAt(n uint) (interface{}, int) {
	return xs.Remove(func(_ interface{}, idx uint) bool {
		return idx == n
	})
}

func (xs *ConcurrentList) RemoveVal(x interface{}) int {
	_, i := xs.Remove(func(val interface{}, _ uint) bool {
		return val == x
	})
	return i
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

func (xs *ConcurrentList) FindVal(pred func(interface{}, uint) bool) interface{} {
	find, _ := xs.Find(pred)
	return find
}

func (xs *ConcurrentList) FindIdx(pred func(interface{}, uint) bool) int {
	_, idx := xs.Find(pred)
	return idx
}

// // FIXME
// func (xs *ConcurrentList) SubList(n uint, m uint) *ConcurrentList {
// 	if n == m {
// 		return New()
// 	} else if m-n == 1 {
// 		return New(xs.Nth(n))
// 	}
// 	xs.lk.Lock()
// 	fst := xs.fst
// 	i := uint(0)
// 	for ; i < n; i++ {
// 		fst = fst.next
// 	}
// 	lst := fst
// 	for ; i < m-1; i++ {
// 		lst = lst.next
// 	}
// 	defer xs.lk.Unlock()
// 	return &ConcurrentList{
// 		fst:  fst,
// 		lst:  lst,
// 		size: m - n,
// 		lk:   &sync.RWMutex{},
// 	}
// }

func (xs *ConcurrentList) Contains(x interface{}) bool {
	_, idx := xs.Find(func(y interface{}, _ uint) bool {
		return y == x
	})
	return idx >= 0
}

func (xs *ConcurrentList) IdxOf(x interface{}) int {
	_, idx := xs.Find(func(y interface{}, _ uint) bool {
		return y == x
	})
	return idx
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

func (xs *ConcurrentList) ForEach(f func(interface{}, uint)) {
	var idx uint32 = 0
	xs.lk.RLock()
	for focus := xs.fst; focus != nil; focus = focus.next {
		f(focus.val, uint(idx))
		idx++
	}
	xs.lk.RUnlock()
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

func (xs *ConcurrentList) MapInPlace(f func(interface{}, uint) interface{}) {
	var idx uint32 = 0
	xs.lk.Lock()
	for focus := xs.fst; focus != nil; focus = focus.next {
		focus.val = f(focus.val, uint(idx))
		idx++
	}
	xs.lk.Unlock()
}

func (xs *ConcurrentList) Map(f func(interface{}, uint) interface{}) *ConcurrentList {
	xs.lk.RLock()
	defer xs.lk.RUnlock()
	if xs.size == 0 {
		return New()
	}
	lst := &node{val: f(xs.fst.val, 0)}
	fst := lst
	var idx uint = 1
	for focus := xs.fst.next; focus != nil; focus = focus.next {
		lst.next = &node{val: f(focus.val, idx)}
		if focus.next != nil {
			lst = lst.next
			idx++
		}
	}
	return &ConcurrentList{
		fst:  fst,
		lst:  lst,
		size: xs.size,
		lk:   &sync.RWMutex{},
	}
}

func (xs *ConcurrentList) MapParallel(f func(interface{}, uint) interface{}) *ConcurrentList {
	xs.lk.RLock()
	defer xs.lk.RUnlock()
	if xs.size == 0 {
		return New()
	}
	lst := &node{val: f(xs.fst.val, 0)}
	fst := lst
	wg := &sync.WaitGroup{}
	wg.Add(int(xs.size) - 1)
	var idx uint = 1
	for focus := xs.fst.next; focus != nil; focus = focus.next {
		lst.next = &node{}
		go func(lst *node, v interface{}) {
			lst.val = f(v, idx)
			wg.Done()
		}(lst.next, focus.val)
		if focus.next != nil {
			lst = lst.next
			idx++
		}
	}
	wg.Wait()
	return &ConcurrentList{
		fst:  fst,
		lst:  lst,
		size: xs.size,
		lk:   &sync.RWMutex{},
	}
}

func (xs *ConcurrentList) Clone() *ConcurrentList {
	return xs.Map(func(i interface{}, _ uint) interface{} {
		return i
	})
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

func (xs *ConcurrentList) Rotate() *ConcurrentList {
	xs.lk.RLock()
	defer xs.lk.RUnlock()
	newList := New()
	for focus := xs.fst; focus != nil; focus = focus.next {
		newList.Prepend(focus.val)
	}
	return newList
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

func (xs *ConcurrentList) All(pred func(z interface{}) bool) bool {
	return !xs.Any(func(z interface{}) bool { return !pred(z) })
}

func (xs *ConcurrentList) Any(pred func(z interface{}) bool) bool {
	return xs.Reduce(false, func(x interface{}, y interface{}, idx uint) interface{} {
		return x.(bool) || pred(y)
	}).(bool)
}

func (xs *ConcurrentList) Join(delim string) string {
	s := xs.ToSlice()
	n := len(s)
	parts := make([]string, n, n)
	for idx, x := range s {
		parts[idx] = x.(string)
	}
	return strings.Join(parts, delim)
}

func (xs *ConcurrentList) ToSlice() []interface{} {
	var idx uint = 0
	xs.lk.RLock()
	n := xs.size
	parts := make([]interface{}, n, n)
	for focus := xs.fst; focus != nil; focus = focus.next {
		parts[idx] = focus.val
		idx++
	}
	xs.lk.RUnlock()
	return parts
}

func (xs *ConcurrentList) String() string {
	var idx uint = 0
	wg := &sync.WaitGroup{}
	xs.lk.RLock()
	n := xs.size
	parts := make([]string, n, n)
	wg.Add(int(n))
	for focus := xs.fst; focus != nil; focus = focus.next {
		go func(val interface{}, idx uint) {
			switch val.(type) {
			case fmt.Stringer:
				parts[idx] = val.(fmt.Stringer).String()
			default:
				parts[idx] = fmt.Sprintf("%v", val)
			}
			wg.Done()
		}(focus.val, idx)
		idx++
	}
	xs.lk.RUnlock()
	wg.Wait()
	return fmt.Sprintf("[%s]", strings.Join(parts, " "))
}

func (xs *ConcurrentList) Filter(pred func(x interface{}) bool) *ConcurrentList {
	newXS := New()
	xs.ForEach(func(x interface{}, u uint) {
		if pred(x) {
			newXS.Append(x)
		}
	})
	return newXS
}

func (xs *ConcurrentList) TakeWhile(pred func(x interface{}) bool) *ConcurrentList {
	newXS := New()
	ok := true
	xs.ForEach(func(x interface{}, u uint) {
		if !ok {
			return
		}
		if ok = pred(x); ok {
			newXS.Append(x)
		}
	})
	return newXS
}

func (xs *ConcurrentList) TakeUntil(pred func(x interface{}) bool) *ConcurrentList {
	return xs.TakeWhile(func(x interface{}) bool {
		return !pred(x)
	})
}
