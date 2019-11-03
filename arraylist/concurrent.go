package arraylist

import (
	"fmt"
	"math/rand"
	"strings"
	"sync"
)

type (
	ConcurrentArrayList struct {
		xs []interface{}
		lk *sync.RWMutex
	}
)

func New(xs ...interface{}) *ConcurrentArrayList {
	newList := &ConcurrentArrayList{
		xs: make([]interface{}, 0),
		lk: &sync.RWMutex{},
	}
	for _, x := range xs {
		newList.Append(x)
	}
	return newList
}

func Range(n int, m int, step int) *ConcurrentArrayList {
	xs := New()
	for i := n; i < m; i += step {
		xs.Append(i)
	}
	return xs
}

func Generate(n int, m int, f func(int) interface{}) *ConcurrentArrayList {
	xs := New()
	for i := n; i < m; i++ {
		xs.Append(f(i))
	}
	return xs
}

func Ints(min int, max int, n uint) *ConcurrentArrayList {
	r := max - min
	return Generate(0, int(n), func(_ int) interface{} { return min + rand.Intn(r) })
}

func Floats(n uint) *ConcurrentArrayList {
	return Generate(0, int(n), func(_ int) interface{} { return rand.Float64() })
}

func Bytes(n uint) *ConcurrentArrayList {
	return Generate(0, int(n), func(_ int) interface{} { return byte(rand.Intn(256)) })
}

func Chars(n uint) *ConcurrentArrayList {
	return Generate(0, int(n), func(_ int) interface{} { return rand.Int31() })
}

func (xs *ConcurrentArrayList) Tail() *ConcurrentArrayList {
	xs.lk.RLock()
	tail := &ConcurrentArrayList{
		xs: xs.xs[1:],
		lk: &sync.RWMutex{},
	}
	xs.lk.RUnlock()
	return tail
}

func (xs *ConcurrentArrayList) Append(val interface{}) {
	xs.lk.Lock()
	xs.xs = append(xs.xs, val)
	xs.lk.Unlock()
}

func (xs *ConcurrentArrayList) Prepend(val interface{}) {
	xs.lk.Lock()
	newXS := make([]interface{}, len(xs.xs)+1)
	newXS[0] = val
	for i := 0; i < len(xs.xs); i++ {
		newXS[i+1] = xs.xs[i]
	}
	xs.xs = newXS
	xs.lk.Unlock()
}

func (xs *ConcurrentArrayList) PeekFront() interface{} {
	xs.lk.RLock()
	front := xs.xs[0]
	xs.lk.RUnlock()
	return front
}

func (xs *ConcurrentArrayList) PeekBack() interface{} {
	xs.lk.RLock()
	back := xs.xs[len(xs.xs)-1]
	xs.lk.RUnlock()
	return back
}

func (xs *ConcurrentArrayList) PopFront() interface{} {
	xs.lk.Lock()
	x := xs.xs[0]
	xs.xs = xs.xs[1:]
	xs.lk.Unlock()
	return x
}

func (xs *ConcurrentArrayList) Remove(pred func(interface{}, uint) bool) (interface{}, int) {
	xs.lk.Lock()
	defer xs.lk.Unlock()
	for idx, focus := range xs.xs {
		if pred(focus, uint(idx)) {
			newXS := make([]interface{}, len(xs.xs)-1)
			i := 0
			for ; i < idx; i++ {
				newXS[i] = xs.xs[i]
			}
			end := len(xs.xs)
			for ; i+1 < end; i++ {
				newXS[i] = xs.xs[i+1]
			}
		}
	}
	return nil, -1
}

func (xs *ConcurrentArrayList) RemoveAt(n uint) (interface{}, int) {
	return xs.Remove(func(_ interface{}, idx uint) bool {
		return idx == n
	})
}

func (xs *ConcurrentArrayList) RemoveVal(x interface{}) int {
	_, i := xs.Remove(func(val interface{}, _ uint) bool {
		return val == x
	})
	return i
}

func (xs *ConcurrentArrayList) Find(pred func(interface{}, uint) bool) (interface{}, int) {
	xs.lk.RLock()
	defer xs.lk.RUnlock()
	for idx, focus := range xs.xs {
		if pred(focus, uint(idx)) {
			return focus, idx
		}
		idx++
	}
	return nil, -1
}

func (xs *ConcurrentArrayList) FindVal(pred func(interface{}, uint) bool) interface{} {
	find, _ := xs.Find(pred)
	return find
}

func (xs *ConcurrentArrayList) FindIdx(pred func(interface{}, uint) bool) int {
	_, idx := xs.Find(pred)
	return idx
}

// // FIXME
// func (xs *ConcurrentArrayList) SubList(n uint, m uint) *ConcurrentArrayList {
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
// 	return &ConcurrentArrayList{
// 		fst:  fst,
// 		lst:  lst,
// 		size: m - n,
// 		lk:   &sync.RWMutex{},
// 	}
// }

func (xs *ConcurrentArrayList) Contains(x interface{}) bool {
	_, idx := xs.Find(func(y interface{}, _ uint) bool {
		return y == x
	})
	return idx >= 0
}

func (xs *ConcurrentArrayList) IdxOf(x interface{}) int {
	_, idx := xs.Find(func(y interface{}, _ uint) bool {
		return y == x
	})
	return idx
}

func Repeat(x interface{}, n uint) *ConcurrentArrayList {
	return Generate(0, int(n), func(_ int) interface{} {
		return x
	})
}

func (xs *ConcurrentArrayList) ForEachParallel(f func(interface{}, uint)) {
	xs.lk.RLock()
	wg := &sync.WaitGroup{}
	wg.Add(len(xs.xs))
	for idx, val := range xs.xs {
		go func(val interface{}, idx int) {
			f(val, uint(idx))
			wg.Done()
		}(val, idx)
	}
	xs.lk.RUnlock()
	wg.Wait()
}

func (xs *ConcurrentArrayList) ForEach(f func(interface{}, uint)) {
	xs.lk.RLock()
	for idx, val := range xs.xs {
		f(val, uint(idx))
		idx++
	}
	xs.lk.RUnlock()
}

func (xs *ConcurrentArrayList) MapParallelInPlace(f func(interface{}, uint) interface{}) {
	xs.lk.Lock()
	wg := sync.WaitGroup{}
	wg.Add(len(xs.xs))
	for idx, val := range xs.xs {
		go func(val interface{}, idx int) {
			xs.xs[idx] = f(val, uint(idx))
			wg.Done()
		}(val, idx)
		idx++
	}
	wg.Wait()
	xs.lk.Unlock()
}

func (xs *ConcurrentArrayList) MapInPlace(f func(interface{}, uint) interface{}) {
	xs.lk.Lock()
	for idx, val := range xs.xs {
		xs.xs[idx] = f(val, uint(idx))
	}
	xs.lk.Unlock()
}

func (xs *ConcurrentArrayList) Map(f func(interface{}, uint) interface{}) *ConcurrentArrayList {
	xs.lk.RLock()
	defer xs.lk.RUnlock()
	newXS := make([]interface{}, len(xs.xs))
	for idx, val := range xs.xs {
		newXS[idx] = f(val, uint(idx))
	}
	return &ConcurrentArrayList{
		xs: newXS,
		lk: &sync.RWMutex{},
	}
}

func (xs *ConcurrentArrayList) MapParallel(f func(interface{}, uint) interface{}) *ConcurrentArrayList {
	xs.lk.RLock()
	defer xs.lk.RUnlock()
	wg := &sync.WaitGroup{}
	wg.Add(len(xs.xs))
	newXS := make([]interface{}, len(xs.xs))
	for idx, val := range xs.xs {
		go func(idx uint, v interface{}) {
			newXS[idx] = f(v, idx)
			wg.Done()
		}(uint(idx), val)
	}
	wg.Wait()
	return &ConcurrentArrayList{
		xs: newXS,
		lk: &sync.RWMutex{},
	}
}

func (xs *ConcurrentArrayList) Clone() *ConcurrentArrayList {
	return xs.Map(func(i interface{}, _ uint) interface{} {
		return i
	})
}

func (xs *ConcurrentArrayList) Nth(n uint) interface{} {
	xs.lk.RLock()
	defer xs.lk.RUnlock()
	return xs.xs[n]
}

func (xs *ConcurrentArrayList) Clear() {
	xs.lk.Lock()
	defer xs.lk.Unlock()
	xs.xs = make([]interface{}, 0)
}

func (xs *ConcurrentArrayList) Empty() bool {
	xs.lk.RLock()
	empty := len(xs.xs) == 0
	xs.lk.RUnlock()
	return empty
}

func (xs *ConcurrentArrayList) Size() uint {
	xs.lk.RLock()
	size := len(xs.xs)
	xs.lk.RUnlock()
	return uint(size)
}

func (xs *ConcurrentArrayList) Rotate() *ConcurrentArrayList {
	xs.lk.RLock()
	defer xs.lk.RUnlock()
	newList := New()
	for _, val := range xs.xs {
		newList.Prepend(val)
	}
	return newList
}

func (xs *ConcurrentArrayList) Reduce(init interface{}, f func(x interface{}, y interface{}, idx uint) interface{}) interface{} {
	xs.lk.RLock()
	defer xs.lk.RUnlock()
	for idx, val := range xs.xs {
		init = f(init, val, uint(idx))
		idx++
	}
	return init
}

func (xs *ConcurrentArrayList) Eq(_ys interface{}) bool {
	switch _ys.(type) {
	case *ConcurrentArrayList:
		ys := _ys.(*ConcurrentArrayList)
		if xs.Size() != ys.Size() {
			return false
		}
		xs.lk.RLock()
		defer xs.lk.RUnlock()
		end := len(xs.xs)
		for i := 0; i < end; i++ {
			if xs.xs[i] != ys.xs[i] {
				return false
			}
		}
		return true
	default:
		return false
	}
}

func (xs *ConcurrentArrayList) All(pred func(z interface{}) bool) bool {
	return !xs.Any(func(z interface{}) bool { return !pred(z) })
}

func (xs *ConcurrentArrayList) Any(pred func(z interface{}) bool) bool {
	return xs.Reduce(false, func(x interface{}, y interface{}, idx uint) interface{} {
		return x.(bool) || pred(y)
	}).(bool)
}

func (xs *ConcurrentArrayList) Join(delim string) string {
	s := xs.ToSlice()
	n := len(s)
	parts := make([]string, n, n)
	for idx, x := range s {
		parts[idx] = x.(string)
	}
	return strings.Join(parts, delim)
}

func (xs *ConcurrentArrayList) ToSlice() []interface{} {
	xs.lk.RLock()
	n := len(xs.xs)
	parts := make([]interface{}, n, n)
	for idx, focus := range xs.xs {
		parts[idx] = focus
		idx++
	}
	xs.lk.RUnlock()
	return parts
}

func (xs *ConcurrentArrayList) String() string {
	xs.lk.RLock()
	n := len(xs.xs)
	parts := make([]string, n, n)
	for idx, focus := range xs.xs {
		switch focus.(type) {
		case fmt.Stringer:
			parts[idx] = focus.(fmt.Stringer).String()
		default:
			parts[idx] = fmt.Sprintf("%v", focus)
		}
		idx++
	}
	xs.lk.RUnlock()
	return fmt.Sprintf("[%s]", strings.Join(parts, " "))
}

func (xs *ConcurrentArrayList) Filter(pred func(x interface{}) bool) *ConcurrentArrayList {
	newXS := New()
	xs.ForEach(func(x interface{}, u uint) {
		if pred(x) {
			newXS.Append(x)
		}
	})
	return newXS
}

func (xs *ConcurrentArrayList) TakeWhile(pred func(x interface{}) bool) *ConcurrentArrayList {
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

func (xs *ConcurrentArrayList) TakeUntil(pred func(x interface{}) bool) *ConcurrentArrayList {
	return xs.TakeWhile(func(x interface{}) bool {
		return !pred(x)
	})
}
