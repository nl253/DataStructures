package stream

import (
	"testing"

	"github.com/nl253/DataStructures/list"
	ut "github.com/nl253/Testing"
)

var fStream = ut.Test("Stream")

func isValid(xs *Stream) bool {
	// if xs.size < 0 {
	// 	fmt.Printf("size was < 0\n")
	// 	return false
	// }
	// if xs.bufLk == nil {
	// 	fmt.Printf("lock was nil\n")
	// 	return false
	// }
	// if xs.fst == nil {
	// 	if xs.lst != nil {
	// 		fmt.Printf("fst was nil so expected lst to be nil as well but was %v :: %T\n", xs.lst, xs.lst)
	// 		return false
	// 	}
	// 	if xs.size != 0 {
	// 		fmt.Printf("fst was nil so expected size to be 0 but was %d\n", xs.size)
	// 		return false
	// 	}
	// }
	// if xs.lst == nil {
	// 	if xs.fst != nil {
	// 		fmt.Printf("lst was nil so expected fst to be nil as well but was %v :: %T\n", xs.fst, xs.fst)
	// 		return false
	// 	}
	// 	if xs.size != 0 {
	// 		fmt.Printf("lst was nil so expected size to be 0 but was %d\n", xs.size)
	// 		return false
	// 	}
	// }
	// if xs.size == 1 {
	// 	if xs.lst != xs.fst {
	// 		fmt.Printf("list of length 0 should have fst and lst point to the same node but fst pointed to %v while lst pointed to %v\n", xs.fst, xs.lst)
	// 		return false
	// 	}
	// }
	//
	// for focus := xs.fst; focus != nil; focus = focus.next {
	// 	if focus.val == nil {
	// 		fmt.Println("val should never be nil on nodes but was")
	// 		return false
	// 	}
	// }
	return true
}

func TestStream_Concurrency(t *testing.T) {
	should := fStream("general concurrency", t)
	should("not freeze the runtime", true, func() interface{} {
		ss := New(1, 2, 3).Close()
		return ss.Pull() == 1 && ss.Pull() == 2 && ss.Pull() == 3 && ss.Pull() == EndMarker && ss.Pull() == EndMarker && ss.Empty()
	})
	should("after close, all pulls result in EndMarker", true, func() interface{} {
		ss := New().Close()
		return ss.Pull() == EndMarker && ss.Pull() == EndMarker
	})
}

func TestStream_Generate(t *testing.T) {
	should := fStream("Generate", t)
	should("generate", list.New(1, 2, 3), func() interface{} { return Range(1, 4, 1).PullAll() })
	should("generate empty", list.New(), func() interface{} { return Range(0, 0, 1).PullAll() })
}

func TestStream_Range(t *testing.T) {
	should := fStream("Generate", t)
	should("range", list.New(1, 2, 3), func() interface{} { return Range(1, 4, 1).PullAll() })
	should("range empty", list.New(), func() interface{} { return Range(0, 0, 1).PullAll() })
}

func TestStream_PullAll(t *testing.T) {
	should := fStream("PullAll", t)
	should("pull all", list.New(1, 2, 3), func() interface{} { return Range(1, 4, 1).PullAll() })
	should("pull all empty", list.New(), func() interface{} { return Range(0, 0, 1).PullAll() })
}

func TestStream_Count(t *testing.T) {
	should := fStream("Count", t)
	should("count elems", uint(10), func() interface{} {
		return Range(0, 10, 1).Close().Count()
	})
}

func TestStream_Sum(t *testing.T) {
	should := fStream("Sum", t)
	should("sum many elems", 0+1+2, func() interface{} { return Range(0, 3, 1).Sum() })
	should("sum 0 elems", 0, func() interface{} { return Range(0, 0, 1).Sum() })
}

func TestStream_Map(t *testing.T) {
	should := fStream("map", t)
	should("map many elements", list.New(1, 2, 3), func() interface{} {
		return Range(0, 3, 1).Map(func(x interface{}) interface{} { return x.(int) + 1 }).PullAll()
	})
	should("map 0 elements", list.New(), func() interface{} {
		return Range(0, 0, 1).Map(func(x interface{}) interface{} { return x.(int) + 1 }).PullAll()
	})
}

// func TestStream_ConcurrencyDoesntLooseData(t *testing.T) {
// 	should := fStream("not loose data", t)
// 	xs := New()
// 	m := &sync.Map{}
// 	wg := sync.WaitGroup{}
// 	n := uint(10000)
// 	wg.Add(int(n))
// 	for i := uint(0); i < n; i++ {
// 		x := rand.Int31()
// 		m.Store(i, x)
// 		go func() {
// 			if rand.Intn(2) == 1 {
// 				xs.Prepend(x)
// 			} else {
// 				xs.Append(x)
// 			}
// 			wg.Done()
// 		}()
// 	}
// 	wg.Wait()
// 	should(fmt.Sprintf("contain %d values after %d insertions", n, n), n, func() interface{} { return xs.Size() })
// 	for i := uint(0); i < n; i++ {
// 		value, ok := m.Load(i)
// 		if !ok {
// 			panic(fmt.Sprintf("failed to put rand num in sync.Map"))
// 		}
// 		should(fmt.Sprintf("contain %v", value), true, func() interface{} {
// 			return xs.Contains(value)
// 		})
// 	}
// 	xs.MapParallelInPlace(func(x interface{}, idx uint) interface{} { return x.(int32) + int32(1) })
// 	for i := uint(0); i < n; i++ {
// 		v, ok := m.Load(i)
// 		if !ok {
// 			panic(fmt.Sprintf("failed to put rand num in sync.Map"))
// 		}
// 		value := v.(int32) + int32(1)
// 		should(fmt.Sprintf("contain %d after %d was mapped to %d", value, v.(int32), value), true, func() interface{} {
// 			return xs.Contains(value)
// 		})
// 	}
// 	for !xs.Empty() {
// 		size := int(xs.Size())
// 		idx := rand.Intn(size)
// 		x := xs.Nth(uint(idx))
// 		xs.RemoveAt(uint(idx))
// 		should(fmt.Sprintf("remove %d (at idx %d) and not contain it afterwards", x, idx), false, func() interface{} {
// 			return xs.Contains(x)
// 		})
// 		should("have size decremented", size-1, func() interface{} { return int(xs.Size()) })
// 	}
// }

// func TestStream_MapParallelInPlace(t *testing.T) {
// 	should := fStream("MapParallelInPlace", t)
// 	should("apply func to each item and modify list in place", New(2, 3, 4), func() interface{} {
// 		xs := New(1, 2, 3)
// 		xs.MapParallelInPlace(func(x interface{}, idx uint) interface{} { return x.(int) + 1 })
// 		return xs
// 	})
// 	should("do nothing for empty lists", New(), func() interface{} {
// 		xs := New()
// 		xs.MapParallelInPlace(func(x interface{}, idx uint) interface{} { return x.(int) + 1 })
// 		return xs
// 	})
// 	for _, xs := range []*Stream{New(), New(1, 2, 3), Ints(-10, 10, 10), Floats(10), Bytes(10), Chars(10), Range(0, 10, 1), Range(-10, 10, 2)} {
// 		should("produce valid list", true, func() interface{} {
// 			xs.MapParallelInPlace(func(x interface{}, idx uint) interface{} { return 10 })
// 			return isValid(xs)
// 		})
// 	}
// }
//
// func TestStream_Map(t *testing.T) {
// 	should := fStream("Map", t)
// 	should("apply func to each item", Range(1, 11, 1), func() interface{} {
// 		return Range(0, 10, 1).Map(func(x interface{}, idx uint) interface{} { return x.(int) + 1 })
// 	})
// 	should("not modifying the list but create a new one", false, func() interface{} {
// 		xs := Ints(0, 100, 10)
// 		ys := xs.Map(func(x interface{}, idx uint) interface{} { return x.(int) + 0 })
// 		return xs == ys
// 	})
// 	should("do nothing for empty lists", New(), func() interface{} {
// 		return New().Map(func(x interface{}, idx uint) interface{} { return x.(int) + 1 })
// 	})
// 	for _, xs := range []*Stream{New(), New(1, 2, 3), Ints(-10, 10, 10), Floats(10), Bytes(10), Chars(10), Range(0, 10, 1), Range(-10, 10, 2)} {
// 		should("produce valid list", true, func() interface{} {
// 			return isValid(xs.Map(func(x interface{}, idx uint) interface{} { return 10 }))
// 		})
// 	}
// }
//
// func TestStream_MapParallel(t *testing.T) {
// 	should := fStream("MapParallel", t)
// 	should("apply func to each item", Range(1, 3, 1), func() interface{} {
// 		return Range(0, 2, 1).MapParallel(func(x interface{}, idx uint) interface{} {
// 			return x.(int) + 1
// 		})
// 	})
// 	should("not modifying the list but create a new one", false, func() interface{} {
// 		xs := Ints(0, 100, 10)
// 		ys := xs.MapParallel(func(x interface{}, idx uint) interface{} {
// 			return x.(int) + 0
// 		})
// 		return xs == ys
// 	})
// 	should("do nothing for empty lists", New(), func() interface{} {
// 		return New().MapParallel(func(x interface{}, idx uint) interface{} {
// 			return x.(int) + 1
// 		})
// 	})
// 	for _, xs := range []*Stream{New(), New(1, 2, 3), Ints(-10, 10, 10), Floats(10), Bytes(10), Chars(10), Range(0, 10, 1), Range(-10, 10, 2)} {
// 		should("produce valid list", true, func() interface{} {
// 			return isValid(xs.MapParallel(func(x interface{}, idx uint) interface{} { return 10 }))
// 		})
// 	}
// }
//
// // func TestStream_SubList(t *testing.T) {
// // 	should := fStream("SubList", t)
// // 	should("slice with low bound inc but upbound exc", New(0), func() interface{} {
// // 		return New(0).SubList(0, 1)
// // 	})
// // 	should("slice with low bound inc but upbound exc", New(0), func() interface{} {
// // 		return New(0, 1, 2).SubList(0, 1)
// // 	})
// // 	should("slice with low bound inc but upbound exc", New(0, 1), func() interface{} {
// // 		return New(0, 1, 2).SubList(0, 2)
// // 	})
// // 	should("evaluate to empty list if empty slice (low bound is the same as upper)", New(), func() interface{} {
// // 		return New().SubList(0, 0)
// // 	})
// // }
//
// func TestStream_Append(t *testing.T) {
// 	should := fStream("Append", t)
// 	should("add item to back", New(0), func() interface{} {
// 		xs := New()
// 		xs.Append(0)
// 		return xs
// 	})
// 	should("add item to back", New(1, 0), func() interface{} {
// 		xs := New(1)
// 		xs.Append(0)
// 		return xs
// 	})
// 	for _, xs := range []*Stream{New(), New(1, 2, 3), Ints(-10, 10, 10), Floats(10), Bytes(10), Chars(10), Range(0, 10, 1), Range(-10, 10, 2)} {
// 		should("produce valid list", true, func() interface{} {
// 			xs.Append(rand.Int())
// 			xs.Append(rand.Float64())
// 			xs.Append(rand.Float32())
// 			return isValid(xs)
// 		})
// 	}
// }
//
// func TestStream_Prepend(t *testing.T) {
// 	should := fStream("Prepend", t)
// 	should("add item to front", New(0), func() interface{} {
// 		xs := New()
// 		xs.Prepend(0)
// 		return xs
// 	})
// 	should("add item to front", New(0, 1), func() interface{} {
// 		xs := New(1)
// 		xs.Prepend(0)
// 		return xs
// 	})
// 	for _, xs := range []*Stream{New(), New(1, 2, 3), Ints(-10, 10, 10), Floats(10), Bytes(10), Chars(10), Range(0, 10, 1), Range(-10, 10, 2)} {
// 		should("produce valid list", true, func() interface{} {
// 			xs.Prepend(rand.Int())
// 			xs.Prepend(rand.Float64())
// 			xs.Prepend(rand.Float32())
// 			return isValid(xs)
// 		})
// 	}
// }
//
// func TestStream_PopFront(t *testing.T) {
// 	should := fStream("PopFront", t)
// 	should("remove 0th item", New(), func() interface{} {
// 		xs := New(0)
// 		xs.PopFront()
// 		return xs
// 	})
// 	for _, xs := range []*Stream{New(1, 2, 3), Ints(-10, 10, 10), Floats(10), Bytes(10), Chars(10), Range(0, 10, 1), Range(-10, 10, 2)} {
// 		should("produce valid list", true, func() interface{} {
// 			xs.PopFront()
// 			return isValid(xs)
// 		})
// 	}
// }
//
// func TestStream_Range(t *testing.T) {
// 	should := fStream("Range", t)
// 	should("generate list of ints in bounds", New(0, 1), func() interface{} {
// 		return Range(0, 2, 1)
// 	})
// 	should("generate empty list for equal bounds", New(), func() interface{} {
// 		return Range(0, 0, 1)
// 	})
// 	should("produce valid list", true, func() interface{} {
// 		return isValid(Range(0, 0, 1)) && isValid(Range(0, 1, 1))
// 	})
// }
//
// func TestStream_Generate(t *testing.T) {
// 	should := fStream("Generate", t)
// 	should("generate list of ints in bounds when using id fLookup", New(0, 1), func() interface{} {
// 		return Generate(0, 2, func(i int) interface{} { return i })
// 	})
// 	should("generate empty list for equal bounds", New(), func() interface{} {
// 		return Generate(0, 0, func(i int) interface{} { return i })
// 	})
// 	should("produce valid list", true, func() interface{} {
// 		return isValid(Generate(0, 0, func(i int) interface{} { return i })) &&
// 			isValid(Generate(0, 2, func(i int) interface{} { return i }))
// 	})
// }
//
// func TestStream_Tail(t *testing.T) {
// 	should := fStream("Tail", t)
// 	should("return all but 0th elements", New(), func() interface{} { return New(0).Tail() })
// }
//
// func TestStream_PeekFront(t *testing.T) {
// 	should := fStream("PeekFront", t)
// 	should("return 0th element", 0, func() interface{} { return New(0).PeekFront() })
// }
//
// func TestStream_PeekBack(t *testing.T) {
// 	should := fStream("PeekBack", t)
// 	should("return 0th element", 0, func() interface{} { return New(0).PeekBack() })
// }
//
// func TestStream_Empty(t *testing.T) {
// 	should := fStream("Empty", t)
// 	should("be true for empty list", true, func() interface{} { return New().Empty() })
// 	should("be false for non-empty list", false, func() interface{} { return New(1).Empty() })
// }
//
// func TestStream_Size(t *testing.T) {
// 	should := fStream("Size", t)
// 	should("be 0 for empty list", uint(0), func() interface{} { return New().Size() })
// 	should("be 1 for list with 1 item", uint(1), func() interface{} { return New(1).Size() })
// }
//
// func TestStream_String(t *testing.T) {
// 	should := fStream("String", t)
// 	should("be \"[]\" for empty list", "[]", func() interface{} {
// 		return New().String()
// 	})
// 	should("be \"[1]\" for list with `1`", "[1]", func() interface{} {
// 		return New(1).String()
// 	})
// }
//
// func TestStream_Clear(t *testing.T) {
// 	should := fStream("Clear", t)
// 	should("do nothing to empty list", New(), func() interface{} {
// 		xs := New()
// 		xs.Clear()
// 		return xs
// 	})
// 	should("remove items from list and set size to 0", New(), func() interface{} {
// 		xs := New(1)
// 		xs.Clear()
// 		return xs
// 	})
// 	should("produce valid list", true, func() interface{} {
// 		xs := New(1)
// 		xs.Clear()
// 		return isValid(xs)
// 	})
// 	should("produce valid list", true, func() interface{} {
// 		xs := New()
// 		xs.Clear()
// 		return isValid(xs)
// 	})
// }
//
// func TestStream_Remove(t *testing.T) {
// 	should := fStream("Remove", t)
// 	should("do nothing to empty list (check idx)", New(), func() interface{} {
// 		xs := New()
// 		xs.Remove(func(i interface{}, u uint) bool {
// 			return u == 0
// 		})
// 		return xs
// 	})
// 	should("do nothing to empty list (check val)", New(), func() interface{} {
// 		xs := New()
// 		xs.Remove(func(i interface{}, u uint) bool {
// 			return i == 1
// 		})
// 		return xs
// 	})
// 	should("remove item from list and decrement size (check val)", uint(0), func() interface{} {
// 		xs := New(1)
// 		xs.Remove(func(i interface{}, u uint) bool {
// 			return i.(int) == 1
// 		})
// 		return xs.Size()
// 	})
// 	should("remove item from list and decrement size (check idx)", New(), func() interface{} {
// 		xs := New(1)
// 		xs.Remove(func(i interface{}, u uint) bool { return u == 0 })
// 		return xs
// 	})
// 	should("do nothing when item not in the list (check val)", New(1), func() interface{} {
// 		xs := New(1)
// 		xs.Remove(func(i interface{}, u uint) bool { return i.(int) == 2 })
// 		return xs
// 	})
// 	should("return removed item", 1, func() interface{} {
// 		xs := New(1)
// 		x, _ := xs.Remove(func(i interface{}, u uint) bool { return u == 0 })
// 		return x
// 	})
// 	should("return index of removed item", 0, func() interface{} {
// 		xs := New(1)
// 		_, idx := xs.Remove(func(i interface{}, u uint) bool { return i.(int) == 1 })
// 		return idx
// 	})
// }
//
// func TestStream_Find(t *testing.T) {
// 	should := fStream("Find", t)
// 	should("return -1 as index when not found (check idx)", -1, func() interface{} {
// 		xs := New(1)
// 		_, idx := xs.Find(func(i interface{}, u uint) bool { return u == 1 })
// 		return idx
// 	})
// 	should("return -1 as index when not found (check val)", -1, func() interface{} {
// 		xs := New(1)
// 		_, idx := xs.Find(func(i interface{}, u uint) bool { return i.(int) == 2 })
// 		return idx
// 	})
// 	should("return -1 as index when empty (check idx)", -1, func() interface{} {
// 		xs := New()
// 		_, idx := xs.Find(func(i interface{}, u uint) bool { return u == 0 })
// 		return idx
// 	})
// 	should("return -1 as index when empty (check val)", -1, func() interface{} {
// 		xs := New()
// 		_, idx := xs.Find(func(i interface{}, u uint) bool { return i.(int) == 0 })
// 		return idx
// 	})
// 	should("return index when found (check val)", 0, func() interface{} {
// 		xs := New(1)
// 		_, idx := xs.Find(func(i interface{}, u uint) bool { return i.(int) == 1 })
// 		return idx
// 	})
// 	should("return index when found (check idx)", 0, func() interface{} {
// 		xs := New(1)
// 		_, idx := xs.Find(func(i interface{}, u uint) bool { return u == 0 })
// 		return idx
// 	})
// }
//
// func TestStream_Nth(t *testing.T) {
// 	should := fStream("Nth", t)
// 	should("return val", 1, func() interface{} { return New(1).Nth(0) })
// }
//
// func TestStream_Eq(t *testing.T) {
// 	should := fStream("Eq", t)
// 	should("be true if both empty", true, func() interface{} {
// 		return New().Eq(New())
// 	})
// 	should("be false if only one is empty", false, func() interface{} {
// 		return New(1).Eq(New())
// 	})
// 	should("be false if only one is empty", false, func() interface{} {
// 		return New().Eq(New(1))
// 	})
// 	should("be true if every el is equal", true, func() interface{} {
// 		return New(1).Eq(New(1))
// 	})
// 	should("be not true if every el is not equal", false, func() interface{} {
// 		return New(2).Eq(New(1))
// 	})
// }
