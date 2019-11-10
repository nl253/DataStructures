package list

import (
	"fmt"
	"math/rand"
	"sync"
	"testing"

	ut "github.com/nl253/Testing"
)

const (
	MANY uint = 1000
	FEW  uint = 10
)

var fCon = ut.Test("ConcurrentList")

func isValid(xs *ConcurrentList) bool {
	if xs.size < 0 {
		fmt.Printf("size was < 0\n")
		return false
	}
	if xs.lk == nil {
		fmt.Printf("lock was nil\n")
		return false
	}
	if xs.fst == nil {
		if xs.lst != nil {
			fmt.Printf("fst was nil so expected lst to be nil as well but was %v :: %T\n", xs.lst, xs.lst)
			return false
		}
		if xs.size != 0 {
			fmt.Printf("fst was nil so expected size to be 0 but was %d\n", xs.size)
			return false
		}
	}
	if xs.lst == nil {
		if xs.fst != nil {
			fmt.Printf("lst was nil so expected fst to be nil as well but was %v :: %T\n", xs.fst, xs.fst)
			return false
		}
		if xs.size != 0 {
			fmt.Printf("lst was nil so expected size to be 0 but was %d\n", xs.size)
			return false
		}
	}
	if xs.size == 1 {
		if xs.lst != xs.fst {
			fmt.Printf("list of length 0 should have fst and lst point to the same node but fst pointed to %v while lst pointed to %v\n", xs.fst, xs.lst)
			return false
		}
	}

	for focus := xs.fst; focus != nil; focus = focus.next {
		if focus.val == nil {
			fmt.Println("val should never be nil on nodes but was")
			return false
		}
	}
	return true
}

func TestConcurrentList_IsThreadSafe(t *testing.T) {
	should := fCon("general concurrency", t)
	should("not freeze the runtime", true, func() interface{} {
		m := int(MANY)
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

func TestConcurrencyDoesNotLooseData(t *testing.T) {
	should := fCon("not loose data", t)
	xs := New()
	m := &sync.Map{}
	wg := sync.WaitGroup{}
	n := MANY
	wg.Add(int(n))
	for i := uint(0); i < n; i++ {
		x := rand.Int31()
		m.Store(i, x)
		go func() {
			if rand.Intn(2) == 1 {
				xs.PushFront(x)
			} else {
				xs.PushBack(x)
			}
			wg.Done()
		}()
	}
	wg.Wait()
	should(fmt.Sprintf("contain %d values after %d insertions", n, n), n, func() interface{} { return xs.Size() })
	for i := uint(0); i < n; i++ {
		value, ok := m.Load(i)
		if !ok {
			panic(fmt.Sprintf("failed to put rand num in sync.Map"))
		}
		should(fmt.Sprintf("contain %v", value), true, func() interface{} {
			return xs.Contains(value)
		})
	}
	xs.MapParallelInPlace(func(x interface{}, idx uint) interface{} { return x.(int32) + int32(1) })
	for i := uint(0); i < n; i++ {
		v, ok := m.Load(i)
		if !ok {
			panic(fmt.Sprintf("failed to put rand num in sync.Map"))
		}
		value := v.(int32) + int32(1)
		should(fmt.Sprintf("contain %d after %d was mapped to %d", value, v.(int32), value), true, func() interface{} {
			return xs.Contains(value)
		})
	}
	for !xs.Empty() {
		size := int(xs.Size())
		idx := rand.Intn(size)
		x := xs.Nth(uint(idx))
		xs.RemoveAt(uint(idx))
		should(fmt.Sprintf("remove %d (at idx %d) and not contain it afterwards", x, idx), false, func() interface{} {
			return xs.Contains(x)
		})
		should("have size decremented", size-1, func() interface{} { return int(xs.Size()) })
	}
}

func TestConcurrentList_Map(t *testing.T) {
	should := fCon("Map", t)
	should("apply func to each item", Range(1, int(MANY)+1, 1), func() interface{} {
		return Range(0, int(MANY), 1).Map(func(x interface{}, idx uint) interface{} { return x.(int) + 1 })
	})
	should("not modifying the list but create a new one", false, func() interface{} {
		xs := Ints(0, int(MANY), MANY)
		ys := xs.Map(func(x interface{}, idx uint) interface{} { return x.(int) + 0 })
		return xs == ys
	})
	should("do nothing for empty lists", New(), func() interface{} {
		return New().Map(func(x interface{}, idx uint) interface{} { return x.(int) + 1 })
	})
	for _, xs := range []*ConcurrentList{New(), New(1, 2, 3), Ints(-int(MANY), int(MANY), MANY), Floats(MANY), Bytes(MANY), Nats(MANY), Chars(MANY), Range(0, int(MANY), 1), Range(-int(MANY), int(MANY), 2)} {
		should("make valid list", true, func() interface{} {
			return isValid(xs.Map(func(x interface{}, idx uint) interface{} { return rand.Int() }))
		})
	}
}

func TestConcurrentList_MapParallelInPlace(t *testing.T) {
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
	for _, xs := range []*ConcurrentList{New(), New(1, 2, 3), Ints(-int(MANY), int(MANY), MANY), Floats(MANY), Bytes(MANY), Nats(MANY), Chars(MANY), Range(0, int(MANY), 1), Range(-int(MANY), int(MANY), 2)} {
		should("make valid list", true, func() interface{} {
			xs.MapParallelInPlace(func(x interface{}, idx uint) interface{} { return rand.Int() })
			return isValid(xs)
		})
	}
}

func TestConcurrentList_MapParallel(t *testing.T) {
	should := fCon("MapParallel", t)
	should("apply func to each item", Range(1, 3, 1), func() interface{} {
		return Range(0, 2, 1).MapParallel(func(x interface{}, idx uint) interface{} {
			return x.(int) + 1
		})
	})
	should("not modifying the list but create a new one", false, func() interface{} {
		xs := Ints(0, int(MANY), FEW)
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
	for _, xs := range []*ConcurrentList{New(), New(1, 2, 3), Ints(-int(MANY), int(MANY), MANY), Floats(MANY), Bytes(MANY), Nats(MANY), Chars(MANY), Range(0, int(MANY), 1), Range(-int(MANY), int(MANY), 2)} {
		should("make valid list", true, func() interface{} {
			return isValid(xs.MapParallel(func(x interface{}, idx uint) interface{} { return rand.Int() }))
		})
	}
}

func TestConcurrentList_Slice(t *testing.T) {
	should := fCon("Slice", t)
	should("slice with lower bound inclusive but upper bound exclusive", New(0), func() interface{} {
		return New(0).Slice(0, 1)
	})
	should("slice with lower bound inclusive but upper bound exclusive", New(0), func() interface{} {
		return New(0, 1, 2).Slice(0, 1)
	})
	should("slice with lower bound inclusive but upper bound exclusive", New(0, 1), func() interface{} {
		return New(0, 1, 2).Slice(0, 2)
	})
	for _, n := range []uint{FEW, MANY, 0} {
		should("evaluate to empty list if empty slice (lower bound is the same as upper)", New(), func() interface{} {
			return New().Slice(n, n)
		})
	}
}

func TestConcurrentList_PushBack(t *testing.T) {
	should := fCon("PushBack", t)
	should("add item to back", New(0), func() interface{} {
		xs := New()
		xs.PushBack(0)
		return xs
	})
	should("add item to back", New(1, 0), func() interface{} {
		xs := New(1)
		xs.PushBack(0)
		return xs
	})
	for _, xs := range []*ConcurrentList{New(), New(1, 2, 3), Ints(-int(MANY), int(MANY), MANY), Floats(MANY), Bytes(MANY), Nats(MANY), Chars(MANY), Range(0, int(MANY), 1), Range(-int(MANY), int(MANY), 2)} {
		should("make valid list", true, func() interface{} {
			xs.PushBack(rand.Int())
			xs.PushBack(rand.Float64())
			xs.PushBack(rand.Float32())
			return isValid(xs)
		})
	}
}

func TestConcurrentList_PushFront(t *testing.T) {
	should := fCon("PushFront", t)
	should("add item to front", New(0), func() interface{} {
		xs := New()
		xs.PushFront(0)
		return xs
	})
	should("add item to front", New(0, 1), func() interface{} {
		xs := New(1)
		xs.PushFront(0)
		return xs
	})
	for _, xs := range []*ConcurrentList{New(), New(1, 2, 3), Ints(-int(MANY), int(MANY), MANY), Floats(MANY), Bytes(MANY), Nats(MANY), Chars(MANY), Range(0, int(MANY), 1), Range(-int(MANY), int(MANY), 2)} {
		should("make valid list", true, func() interface{} {
			xs.PushFront(rand.Int())
			xs.PushFront(rand.Float64())
			xs.PushFront(rand.Float32())
			return isValid(xs)
		})
	}
}

func TestConcurrentList_PopFront(t *testing.T) {
	should := fCon("PopFront", t)
	should("remove 0th item", New(), func() interface{} {
		xs := New(0)
		xs.PopFront()
		return xs
	})
	for _, xs := range []*ConcurrentList{New(1, 2, 3), Ints(-int(MANY), int(MANY), MANY), Floats(MANY), Bytes(MANY), Nats(MANY), Chars(MANY), Range(0, int(MANY), 1), Range(-int(MANY), int(MANY), 2)} {
		should("make valid list", true, func() interface{} {
			xs.PopFront()
			return isValid(xs)
		})
	}
}

func TestConcurrentList_Range(t *testing.T) {
	should := fCon("Range", t)
	should("generate list of ints in bounds", New(0, 1), func() interface{} {
		return Range(0, 2, 1)
	})
	for _, n := range []uint{FEW, MANY, 0} {
		should("generate empty list for equal bounds", New(), func() interface{} {
			return Range(0, 0, int(n))
		})
		should("make valid list", true, func() interface{} {
			return isValid(Range(0, int(n), 1))
		})
		should("make valid list", true, func() interface{} {
			return isValid(Range(0, int(n), int(FEW)))
		})
		should("make valid list", true, func() interface{} {
			return isValid(Range(int(n), -int(n), -int(FEW)))
		})
		should("make valid list", true, func() interface{} {
			return isValid(Range(int(n), int(n)*2, 1))
		})
		should("make valid list", true, func() interface{} {
			return isValid(Range(int(n), int(n)*2, int(n)))
		})
	}
}

func TestConcurrentList_Generate(t *testing.T) {
	should := fCon("Generate", t)
	should("generate list of ints in bounds when using id fLookup", New(0, 1), func() interface{} {
		return Generate(0, 2, func(i int) interface{} { return i })
	})
	should("generate empty list for equal bounds", New(), func() interface{} {
		return Generate(0, 0, func(i int) interface{} { return i })
	})
	should("make valid list", true, func() interface{} {
		return isValid(Generate(0, 0, func(i int) interface{} { return i })) &&
			isValid(Generate(0, 2, func(i int) interface{} { return i }))
	})
}

func TestConcurrentList_Ints(t *testing.T) {
	should := fCon("Ints", t)
	should("make [0, 0, 0]", New(0, 0, 0), func() interface{} {
		return Ints(0, 1, 3)
	})
	should("make [1, 1]", New(1, 1), func() interface{} {
		return Ints(1, 2, 2)
	})
	should("make empty list", New(), func() interface{} {
		return Ints(-int(MANY), int(MANY), 0)
	})
	should("make int list", true, func() interface{} {
		return Ints(-int(MANY), int(MANY), MANY).All(func(x interface{}) bool {
			switch x.(type) {
			case int:
				return true
			default:
				return false
			}
		})
	})
	for _, xs := range []*ConcurrentList{
		Ints(0, 1, 3),
		Ints(1, 2, 2),
		Ints(1, 2, FEW),
		Ints(0, int(MANY), 0),
		Ints(-int(MANY), int(MANY), MANY),
	} {
		should("make valid list", true, func() interface{} { return isValid(xs) })
	}
}

func TestConcurrentList_Chars(t *testing.T) {
	should := fCon("Chars", t)
	for _, n := range []uint{FEW, MANY, 0} {
		should("make list of chars", true, func() interface{} {
			return Chars(n).All(func(x interface{}) bool {
				switch x.(type) {
				case rune:
					return true
				default:
					return false
				}
			})
		})
		should(fmt.Sprintf("make list of %d chars", n), true, func() interface{} {
			return Chars(n).Size() == n
		})
		should("make valid list", true, func() interface{} { return isValid(Chars(n)) })
	}
}

func TestConcurrentList_Nats(t *testing.T) {
	should := fCon("Nats", t)
	for _, n := range []uint{FEW, MANY, 0} {
		should("make list of nats", true, func() interface{} {
			return Nats(n).All(func(x interface{}) bool {
				switch x.(type) {
				case uint:
					return true
				default:
					return false
				}
			})
		})
		should(fmt.Sprintf("make list of %d nats", n), true, func() interface{} {
			return Nats(n).Size() == n
		})
		should("make valid list", true, func() interface{} { return isValid(Nats(n)) })
	}
	should("make [0, 1, 2]", New(uint(0), uint(1), uint(2)), func() interface{} {
		return Nats(3)
	})
	should("make [0, 1]", New(uint(0), uint(1)), func() interface{} {
		return Nats(2)
	})
}

func TestConcurrentList_Floats(t *testing.T) {
	should := fCon("Floats", t)
	for _, n := range []uint{FEW, MANY, 0} {
		should("make list of floats", true, func() interface{} {
			return Floats(n).All(func(x interface{}) bool {
				switch x.(type) {
				case float64:
					return true
				default:
					return false
				}
			})
		})
		should(fmt.Sprintf("make list of %d floats", n), true, func() interface{} {
			return Floats(n).Size() == n
		})
		should("make valid list", true, func() interface{} { return isValid(Floats(n)) })
	}
}

func TestConcurrentList_Bytes(t *testing.T) {
	should := fCon("Bytes", t)
	for _, n := range []uint{FEW, MANY, 0} {
		should("make list of bytes", true, func() interface{} {
			return Bytes(n).All(func(x interface{}) bool {
				switch x.(type) {
				case byte:
					return true
				default:
					return false
				}
			})
		})
		should(fmt.Sprintf("make list of %d bytes", n), true, func() interface{} {
			return Bytes(n).Size() == n
		})
		should("make valid list", true, func() interface{} { return isValid(Bytes(n)) })
	}
}

func TestConcurrentList_Tail(t *testing.T) {
	should := fCon("Tail", t)
	should("return all but 0th elements", New(), func() interface{} { return New(0).Tail() })
}

func TestConcurrentList_PeekFront(t *testing.T) {
	should := fCon("PeekFront", t)
	should("return 0th element", 0, func() interface{} { return New(0).PeekFront() })
}

func TestConcurrentList_PeekBack(t *testing.T) {
	should := fCon("PeekBack", t)
	should("return 0th element", 0, func() interface{} { return New(0).PeekBack() })
}

func TestConcurrentList_Empty(t *testing.T) {
	should := fCon("Empty", t)
	should("be true for empty list", true, func() interface{} { return New().Empty() })
	for _, n := range []uint{FEW, MANY} {
		should("be false for non-empty list", false, func() interface{} { return Nats(n).Empty() })
	}
}

func TestConcurrentList_Size(t *testing.T) {
	should := fCon("Size", t)
	should("be 0 for empty list", uint(0), func() interface{} { return New().Size() })
	should("be 1 for list with 1 item", uint(1), func() interface{} { return New(1).Size() })
}

func TestConcurrentList_String(t *testing.T) {
	should := fCon("String", t)
	should("be \"[]\" for empty list", "[]", func() interface{} {
		return New().String()
	})
	should("be \"[1]\" for list with `1`", "[1]", func() interface{} {
		return New(1).String()
	})
}

func TestConcurrentList_Clear(t *testing.T) {
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
	should("make valid list", true, func() interface{} {
		xs := New(1)
		xs.Clear()
		return isValid(xs)
	})
	should("make valid list", true, func() interface{} {
		xs := New()
		xs.Clear()
		return isValid(xs)
	})
}

func TestConcurrentList_Remove(t *testing.T) {
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

func TestConcurrentList_Find(t *testing.T) {
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

func TestConcurrentList_New(t *testing.T) {
	should := fCon("New", t)
	should("make new list", []interface{}{0, 1, 2}, func() interface{} { return New(0, 1, 2).ToSlice() })
	should("make empty list", []interface{}{}, func() interface{} { return New().ToSlice() })
}

func TestConcurrentList_Nth(t *testing.T) {
	should := fCon("Nth", t)
	should("return val", 1, func() interface{} { return New(1).Nth(0) })
}

func TestConcurrentList_Reduce(t *testing.T) {
	should := fCon("Reduce", t)
	should("collect empty to empty slice", []interface{}{}, func() interface{} { return New().ToSlice() })
	should("collect to slice", New(uint(0), uint(1), uint(2)), func() interface{} {
		return Nats(3).Reduce(New(), func(x interface{}, y interface{}, _ uint) interface{} {
			x.(*ConcurrentList).PushBack(y)
			return x
		})
	})
}

func TestConcurrentList_Eq(t *testing.T) {
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
