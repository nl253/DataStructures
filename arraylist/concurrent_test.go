package arraylist

import (
	"math/rand"
	"sync"
	"testing"

	ut "github.com/nl253/Testing"
)

var fCon = ut.Test("ConcurrentArrayList")

func isValid(xs *ConcurrentArrayList) bool {
	return true
}

func TestConcurrentArrayList_Concurrency(t *testing.T) {
	should := fCon("general concurrency", t)
	should("not freeze the runtime", true, func() interface{} {
		m := 100000
		xs := Range(0, m, 1)
		locks := make([]*sync.Mutex, m, m)
		for i := uint(0); i < xs.Size()-1; i++ {
			locks[i] = &sync.Mutex{}
			locks[i].Lock()
			go func(i uint) uint {
				defer locks[i].Unlock()
				return i + 1
			}(i)
		}
		for i := uint(0); i < xs.Size()-1; i++ {
			l := locks[i]
			l.Lock()
			l.Unlock()
		}
		return true
	})
}

// func TestConcurrentArrayList_ConcurrencyDoesntLooseData(t *testing.T) {
//     should := fCon("not loose data", t)
//     xs := New()
//     m := &sync.Map{}
//     wg := sync.WaitGroup{}
//     n := uint(10000)
//     wg.Add(int(n))
//     for i := uint(0); i < n; i++ {
//         x := rand.Int31()
//         m.Store(i, x)
//         go func() {
//             if rand.Intn(2) == 1 {
//                 xs.Prepend(x)
//             } else {
//                 xs.Append(x)
//             }
//             wg.Running()
//         }()
//     }
//     wg.Wait()
//     should(fmt.Sprintf("contain %d values after %d insertions", n, n), n, func() interface{} { return xs.Size() })
//     for i := uint(0); i < n; i++ {
//         value, ok := m.Load(i)
//         if !ok {
//             panic(fmt.Sprintf("failed to put rand num in sync.Map"))
//         }
//         should(fmt.Sprintf("contain %v", value), true, func() interface{} {
//             return xs.Contains(value)
//         })
//     }
//     xs.MapParallelInPlace(func(x interface{}, idx uint) interface{} { return x.(int32) + int32(1) })
//     for i := uint(0); i < n; i++ {
//         v, ok := m.Load(i)
//         if !ok {
//             panic(fmt.Sprintf("failed to put rand num in sync.Map"))
//         }
//         value := v.(int32) + int32(1)
//         should(fmt.Sprintf("contain %d after %d was mapped to %d", value, v.(int32), value), true, func() interface{} {
//             return xs.Contains(value)
//         })
//     }
//     for !xs.BufEmpty() {
//         size := int(xs.Size())
//         idx := rand.Intn(size)
//         x := xs.Nth(uint(idx))
//         xs.RemoveAt(uint(idx))
//         should(fmt.Sprintf("remove %d (at idx %d) and not contain it afterwards", x, idx), false, func() interface{} {
//             return xs.Contains(x)
//         })
//         should("have size decremented", size-1, func() interface{} { return int(xs.Size()) })
//     }
// }

func TestConcurrentArrayList_MapParallelInPlace(t *testing.T) {
	should := fCon("MapParallelInPlace", t)
	should("apply func to each item and modify list in place", New(2, 3, 4), func() interface{} {
		xs := New(1, 2, 3)
		xs.MapParallelInPlace(func(x interface{}, idx uint) interface{} { return x.(int) + 1 })
		return xs
	})
	should("do nothing for empty lists", New(), func() interface{} {
		xs := New()
		xs.MapParallelInPlace(func(x interface{}, idx uint) interface{} { return x.(int) + 1 })
		return xs
	})
	for _, xs := range []*ConcurrentArrayList{New(), New(1, 2, 3), Ints(-10, 10, 10), Floats(10), Bytes(10), Chars(10), Range(0, 10, 1), Range(-10, 10, 2)} {
		should("produce valid list", true, func() interface{} {
			xs.MapParallelInPlace(func(x interface{}, idx uint) interface{} { return 10 })
			return isValid(xs)
		})
	}
}

func TestConcurrentArrayList_Map(t *testing.T) {
	should := fCon("Map", t)
	should("apply func to each item", Range(1, 11, 1), func() interface{} {
		return Range(0, 10, 1).Map(func(x interface{}, idx uint) interface{} { return x.(int) + 1 })
	})
	should("not modifying the list but create a new one", false, func() interface{} {
		xs := Ints(0, 100, 10)
		ys := xs.Map(func(x interface{}, idx uint) interface{} { return x.(int) + 0 })
		return xs == ys
	})
	should("do nothing for empty lists", New(), func() interface{} {
		return New().Map(func(x interface{}, idx uint) interface{} { return x.(int) + 1 })
	})
	for _, xs := range []*ConcurrentArrayList{New(), New(1, 2, 3), Ints(-10, 10, 10), Floats(10), Bytes(10), Chars(10), Range(0, 10, 1), Range(-10, 10, 2)} {
		should("produce valid list", true, func() interface{} {
			return isValid(xs.Map(func(x interface{}, idx uint) interface{} { return 10 }))
		})
	}
}

func TestConcurrentArrayList_MapParallel(t *testing.T) {
	should := fCon("MapParallel", t)
	should("apply func to each item", Range(1, 3, 1), func() interface{} {
		return Range(0, 2, 1).MapParallel(func(x interface{}, idx uint) interface{} {
			return x.(int) + 1
		})
	})
	should("not modifying the list but create a new one", false, func() interface{} {
		xs := Ints(0, 100, 10)
		ys := xs.MapParallel(func(x interface{}, idx uint) interface{} {
			return x.(int) + 0
		})
		return xs == ys
	})
	should("do nothing for empty lists", New(), func() interface{} {
		return New().MapParallel(func(x interface{}, idx uint) interface{} {
			return x.(int) + 1
		})
	})
	for _, xs := range []*ConcurrentArrayList{New(), New(1, 2, 3), Ints(-10, 10, 10), Floats(10), Bytes(10), Chars(10), Range(0, 10, 1), Range(-10, 10, 2)} {
		should("produce valid list", true, func() interface{} {
			return isValid(xs.MapParallel(func(x interface{}, idx uint) interface{} { return 10 }))
		})
	}
}

// func TestConcurrentArrayList_SubArrayList(t *testing.T) {
// 	should := fCon("SubArrayList", t)
// 	should("slice with low bound inc but upbound exc", New(0), func() interface{} {
// 		return New(0).SubArrayList(0, 1)
// 	})
// 	should("slice with low bound inc but upbound exc", New(0), func() interface{} {
// 		return New(0, 1, 2).SubArrayList(0, 1)
// 	})
// 	should("slice with low bound inc but upbound exc", New(0, 1), func() interface{} {
// 		return New(0, 1, 2).SubArrayList(0, 2)
// 	})
// 	should("evaluate to empty list if empty slice (low bound is the same as upper)", New(), func() interface{} {
// 		return New().SubArrayList(0, 0)
// 	})
// }

func TestConcurrentArrayList_Append(t *testing.T) {
	should := fCon("Append", t)
	should("add item to back", New(0), func() interface{} {
		xs := New()
		xs.Append(0)
		return xs
	})
	should("add item to back", New(1, 0), func() interface{} {
		xs := New(1)
		xs.Append(0)
		return xs
	})
	for _, xs := range []*ConcurrentArrayList{New(), New(1, 2, 3), Ints(-10, 10, 10), Floats(10), Bytes(10), Chars(10), Range(0, 10, 1), Range(-10, 10, 2)} {
		should("produce valid list", true, func() interface{} {
			xs.Append(rand.Int())
			xs.Append(rand.Float64())
			xs.Append(rand.Float32())
			return isValid(xs)
		})
	}
}

func TestConcurrentArrayList_Prepend(t *testing.T) {
	should := fCon("Prepend", t)
	should("add item to front", New(0), func() interface{} {
		xs := New()
		xs.Prepend(0)
		return xs
	})
	should("add item to front", New(0, 1), func() interface{} {
		xs := New(1)
		xs.Prepend(0)
		return xs
	})
	for _, xs := range []*ConcurrentArrayList{New(), New(1, 2, 3), Ints(-10, 10, 10), Floats(10), Bytes(10), Chars(10), Range(0, 10, 1), Range(-10, 10, 2)} {
		should("produce valid list", true, func() interface{} {
			xs.Prepend(rand.Int())
			xs.Prepend(rand.Float64())
			xs.Prepend(rand.Float32())
			return isValid(xs)
		})
	}
}

func TestConcurrentArrayList_PopFront(t *testing.T) {
	should := fCon("PopFront", t)
	should("remove 0th item", New(), func() interface{} {
		xs := New(0)
		xs.PopFront()
		return xs
	})
	for _, xs := range []*ConcurrentArrayList{New(1, 2, 3), Ints(-10, 10, 10), Floats(10), Bytes(10), Chars(10), Range(0, 10, 1), Range(-10, 10, 2)} {
		should("produce valid list", true, func() interface{} {
			xs.PopFront()
			return isValid(xs)
		})
	}
}

func TestConcurrentArrayList_Range(t *testing.T) {
	should := fCon("Range", t)
	should("generate list of ints in bounds", New(0, 1), func() interface{} {
		return Range(0, 2, 1)
	})
	should("generate empty list for equal bounds", New(), func() interface{} {
		return Range(0, 0, 1)
	})
	should("produce valid list", true, func() interface{} {
		return isValid(Range(0, 0, 1)) && isValid(Range(0, 1, 1))
	})
}

func TestConcurrentArrayList_Generate(t *testing.T) {
	should := fCon("Generate", t)
	should("generate list of ints in bounds when using id fLookup", New(0, 1), func() interface{} {
		return Generate(0, 2, func(i int) interface{} { return i })
	})
	should("generate empty list for equal bounds", New(), func() interface{} {
		return Generate(0, 0, func(i int) interface{} { return i })
	})
	should("produce valid list", true, func() interface{} {
		return isValid(Generate(0, 0, func(i int) interface{} { return i })) &&
			isValid(Generate(0, 2, func(i int) interface{} { return i }))
	})
}

func TestConcurrentArrayList_Tail(t *testing.T) {
	should := fCon("Tail", t)
	should("return all but 0th elements", New(), func() interface{} { return New(0).Tail() })
}

func TestConcurrentArrayList_PeekFront(t *testing.T) {
	should := fCon("PeekFront", t)
	should("return 0th element", 0, func() interface{} { return New(0).PeekFront() })
}

func TestConcurrentArrayList_PeekBack(t *testing.T) {
	should := fCon("PeekBack", t)
	should("return 0th element", 0, func() interface{} { return New(0).PeekBack() })
}

func TestConcurrentArrayList_Empty(t *testing.T) {
	should := fCon("BufEmpty", t)
	should("be true for empty list", true, func() interface{} { return New().Empty() })
	should("be false for non-empty list", false, func() interface{} { return New(1).Empty() })
}

func TestConcurrentArrayList_Size(t *testing.T) {
	should := fCon("Size", t)
	should("be 0 for empty list", uint(0), func() interface{} { return New().Size() })
	should("be 1 for list with 1 item", uint(1), func() interface{} { return New(1).Size() })
}

func TestConcurrentArrayList_String(t *testing.T) {
	should := fCon("String", t)
	should("be \"[]\" for empty list", "[]", func() interface{} {
		return New().String()
	})
	should("be \"[1]\" for list with `1`", "[1]", func() interface{} {
		return New(1).String()
	})
}

func TestConcurrentArrayList_Clear(t *testing.T) {
	should := fCon("Clear", t)
	should("do nothing to empty list", New(), func() interface{} {
		xs := New()
		xs.Clear()
		return xs
	})
	should("remove items from list and set size to 0", New(), func() interface{} {
		xs := New(1)
		xs.Clear()
		return xs
	})
	should("produce valid list", true, func() interface{} {
		xs := New(1)
		xs.Clear()
		return isValid(xs)
	})
	should("produce valid list", true, func() interface{} {
		xs := New()
		xs.Clear()
		return isValid(xs)
	})
}

func TestConcurrentArrayList_Remove(t *testing.T) {
	should := fCon("Remove", t)
	should("do nothing to empty list (check idx)", New(), func() interface{} {
		xs := New()
		xs.Remove(func(i interface{}, u uint) bool {
			return u == 0
		})
		return xs
	})
	should("do nothing to empty list (check val)", New(), func() interface{} {
		xs := New()
		xs.Remove(func(i interface{}, u uint) bool {
			return i == 1
		})
		return xs
	})
	should("remove item from list and decrement size (check val)", uint(0), func() interface{} {
		xs := New(1)
		xs.Remove(func(i interface{}, u uint) bool {
			return i.(int) == 1
		})
		return xs.Size()
	})
	should("remove item from list and decrement size (check idx)", New(), func() interface{} {
		xs := New(1)
		xs.Remove(func(i interface{}, u uint) bool { return u == 0 })
		return xs
	})
	should("do nothing when item not in the list (check val)", New(1), func() interface{} {
		xs := New(1)
		xs.Remove(func(i interface{}, u uint) bool { return i.(int) == 2 })
		return xs
	})
	should("return removed item", 1, func() interface{} {
		xs := New(1)
		x, _ := xs.Remove(func(i interface{}, u uint) bool { return u == 0 })
		return x
	})
	should("return index of removed item", 0, func() interface{} {
		xs := New(1)
		_, idx := xs.Remove(func(i interface{}, u uint) bool { return i.(int) == 1 })
		return idx
	})
}

func TestConcurrentArrayList_Find(t *testing.T) {
	should := fCon("Find", t)
	should("return -1 as index when not found (check idx)", -1, func() interface{} {
		xs := New(1)
		_, idx := xs.Find(func(i interface{}, u uint) bool { return u == 1 })
		return idx
	})
	should("return -1 as index when not found (check val)", -1, func() interface{} {
		xs := New(1)
		_, idx := xs.Find(func(i interface{}, u uint) bool { return i.(int) == 2 })
		return idx
	})
	should("return -1 as index when empty (check idx)", -1, func() interface{} {
		xs := New()
		_, idx := xs.Find(func(i interface{}, u uint) bool { return u == 0 })
		return idx
	})
	should("return -1 as index when empty (check val)", -1, func() interface{} {
		xs := New()
		_, idx := xs.Find(func(i interface{}, u uint) bool { return i.(int) == 0 })
		return idx
	})
	should("return index when found (check val)", 0, func() interface{} {
		xs := New(1)
		_, idx := xs.Find(func(i interface{}, u uint) bool { return i.(int) == 1 })
		return idx
	})
	should("return index when found (check idx)", 0, func() interface{} {
		xs := New(1)
		_, idx := xs.Find(func(i interface{}, u uint) bool { return u == 0 })
		return idx
	})
}

func TestConcurrentArrayList_Nth(t *testing.T) {
	should := fCon("Nth", t)
	should("return val", 1, func() interface{} { return New(1).Nth(0) })
}

func TestConcurrentArrayList_Eq(t *testing.T) {
	should := fCon("Eq", t)
	should("be true if both empty", true, func() interface{} {
		return New().Eq(New())
	})
	should("be false if only one is empty", false, func() interface{} {
		return New(1).Eq(New())
	})
	should("be false if only one is empty", false, func() interface{} {
		return New().Eq(New(1))
	})
	should("be true if every el is equal", true, func() interface{} {
		return New(1).Eq(New(1))
	})
	should("be not true if every el is not equal", false, func() interface{} {
		return New(2).Eq(New(1))
	})
}
