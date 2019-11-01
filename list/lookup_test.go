package list

import (
	"sync"
	"testing"

	ut "github.com/nl253/Testing"
)

var fLookup = ut.Test("LookupList")

func TestLookupList_Concurrency(t *testing.T) {
	should := fLookup("general concurrency", t)
	should("not freeze the runtime", true, func() interface{} {
		m := 100000
		xs := Range(0, m)
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

func TestLookupList_MapParallelInPlace(t *testing.T) {
	should := fLookup("MapParallelInPlace", t)
	should("apply func to each item and modify list in place", NewLookup(2, 3, 4), func() interface{} {
		xs := NewLookup(1, 2, 3)
		xs.MapParallelInPlace(func(x interface{}, idx uint) interface{} {
			return x.(int) + 1
		})
		return xs
	})
	should("do nothing for empty lists", NewLookup(), func() interface{} {
		xs := NewLookup()
		xs.MapParallelInPlace(func(x interface{}, idx uint) interface{} {
			return x.(int) + 1
		})
		return xs
	})
}

func TestLookupList_Append(t *testing.T) {
	should := fLookup("Append", t)
	should("add item to back", NewLookup(0), func() interface{} {
		xs := NewLookup()
		xs.Append(0)
		return xs
	})
	should("add item to back", NewLookup(1, 0), func() interface{} {
		xs := NewLookup(1)
		xs.Append(0)
		return xs
	})
}

func TestLookupList_Prepend(t *testing.T) {
	should := fLookup("Prepend", t)
	should("add item to front", NewLookup(0), func() interface{} {
		xs := NewLookup()
		xs.Prepend(0)
		return xs
	})
	should("add item to front", NewLookup(0, 1), func() interface{} {
		xs := NewLookup(1)
		xs.Prepend(0)
		return xs
	})
}

func TestLookupList_PopFront(t *testing.T) {
	should := fLookup("PopFront", t)
	should("remove 0th item", NewLookup(), func() interface{} {
		xs := NewLookup(0)
		xs.PopFront()
		return xs
	})
}

func TestLookupList_RangeLookup(t *testing.T) {
	should := fLookup("RangeLookup", t)
	should("generateLookup list of ints in bounds", NewLookup(0, 1), func() interface{} {
		return RangeLookup(0, 2)
	})
	should("generateLookup empty list for equal bounds", NewLookup(), func() interface{} {
		return RangeLookup(0, 0)
	})
}

func TestLookupList_GenerateLookup(t *testing.T) {
	should := fLookup("GenerateLookup", t)
	should("generateLookup list of ints in bounds when using id fLookup", NewLookup(0, 1), func() interface{} {
		return GenerateLookup(0, 2, func(i int) interface{} {
			return i
		})
	})
	should("generateLookup empty list for equal bounds", NewLookup(), func() interface{} {
		return GenerateLookup(0, 0, func(i int) interface{} {
			return i
		})
	})
}

func TestLookupList_Tail(t *testing.T) {
	should := fLookup("Tail", t)
	should("return all but 0th elements", NewLookup(), func() interface{} {
		return NewLookup(0).Tail()
	})
}

func TestLookupList_PeekFront(t *testing.T) {
	should := fLookup("PeekFront", t)
	should("return 0th element", 0, func() interface{} {
		return NewLookup(0).PeekFront()
	})
}

func TestLookupList_PeekBack(t *testing.T) {
	should := fLookup("PeekBack", t)
	should("return 0th element", 0, func() interface{} {
		return NewLookup(0).PeekBack()
	})
}

func TestLookupList_Empty(t *testing.T) {
	should := fLookup("Empty", t)
	should("be true for empty list", true, func() interface{} {
		return NewLookup().Empty()
	})
	should("be false for non-empty list", false, func() interface{} {
		return NewLookup(1).Empty()
	})
}

func TestLookupList_Size(t *testing.T) {
	should := fLookup("Size", t)
	should("be 0 for empty list", uint(0), func() interface{} {
		return NewLookup().Size()
	})
	should("be 1 for list with 1 item", uint(1), func() interface{} {
		return NewLookup(1).Size()
	})
}

func TestLookupList_String(t *testing.T) {
	should := fLookup("String", t)
	should("be \"[]\" for empty list", "[]", func() interface{} {
		return NewLookup().String()
	})
	should("be \"[1]\" for list with `1`", "[1]", func() interface{} {
		return NewLookup(1).String()
	})
}

func TestLookupList_Clear(t *testing.T) {
	should := fLookup("Clear", t)
	should("do nothing to empty list", NewLookup(), func() interface{} {
		xs := NewLookup()
		xs.Clear()
		return xs
	})
	should("remove items from list and set size to 0", NewLookup(), func() interface{} {
		xs := NewLookup(1)
		xs.Clear()
		return xs
	})
}

func TestLookupList_Remove(t *testing.T) {
	should := fLookup("Remove", t)
	should("do nothing to empty list (check idx)", NewLookup(), func() interface{} {
		xs := NewLookup()
		xs.Remove(func(i interface{}, u uint) bool {
			return u == 0
		})
		return xs
	})
	should("do nothing to empty list (check val)", NewLookup(), func() interface{} {
		xs := NewLookup()
		xs.Remove(func(i interface{}, u uint) bool {
			return i == 1
		})
		return xs
	})
	should("remove item from list and decrement size (check val)", NewLookup(), func() interface{} {
		xs := NewLookup(1)
		xs.Remove(func(i interface{}, u uint) bool {
			return i.(int) == 1
		})
		return xs
	})
	should("remove item from list and decrement size (check idx)", NewLookup(), func() interface{} {
		xs := NewLookup(1)
		xs.Remove(func(i interface{}, u uint) bool {
			return u == 0
		})
		return xs
	})
	should("do nothing when item not in the list (check val)", NewLookup(1), func() interface{} {
		xs := NewLookup(1)
		xs.Remove(func(i interface{}, u uint) bool {
			return i.(int) == 2
		})
		return xs
	})
	should("return removed item", 1, func() interface{} {
		xs := NewLookup(1)
		x, _ := xs.Remove(func(i interface{}, u uint) bool { return u == 0 })
		return x
	})
	should("return index of removed item", 0, func() interface{} {
		xs := NewLookup(1)
		_, idx := xs.Remove(func(i interface{}, u uint) bool {
			return i.(int) == 1
		})
		return idx

	})
}

func TestLookupList_Find(t *testing.T) {
	should := fLookup("Find", t)
	should("return false when not found (check idx)", false, func() interface{} {
		xs := NewLookup(1)
		_, ok := xs.Find(func(i interface{}, u uint) bool {
			return u == 1
		})
		return ok
	})
	should("return false when not found (check val)", false, func() interface{} {
		xs := NewLookup(1)
		_, ok := xs.Find(func(i interface{}, u uint) bool {
			return i.(int) == 2
		})
		return ok
	})
	should("return false when empty (check idx)", false, func() interface{} {
		xs := NewLookup()
		_, ok := xs.Find(func(i interface{}, u uint) bool {
			return u == 0
		})
		return ok
	})
	should("return false when empty (check val)", false, func() interface{} {
		xs := NewLookup()
		_, ok := xs.Find(func(i interface{}, u uint) bool {
			return i.(int) == 0
		})
		return ok
	})
	should("return true when found (check val)", true, func() interface{} {
		xs := NewLookup(1)
		_, ok := xs.Find(func(i interface{}, u uint) bool {
			return i.(int) == 1
		})
		return ok
	})
	should("return true when found (check idx)", true, func() interface{} {
		xs := NewLookup(1)
		_, idx := xs.Find(func(i interface{}, u uint) bool {
			return u == 0
		})
		return idx
	})
	should("reorder when found (check val)", true, func() interface{} {
		xs := NewLookup(0, 1)
		xs.Find(func(i interface{}, u uint) bool {
			return i.(int) == 1
		})
		return xs.Eq(NewLookup(1, 0))
	})
	should("reorder when found (check idx)", NewLookup(1, 0), func() interface{} {
		xs := NewLookup(0, 1)
		xs.Find(func(i interface{}, u uint) bool {
			return u == 1
		})
		return xs
	})
}

func TestLookupList_Nth(t *testing.T) {
	should := fLookup("Nth", t)
	should("return val", 1, func() interface{} {
		return NewLookup(1).Nth(0)
	})
}

func TestLookupList_Eq(t *testing.T) {
	should := fLookup("Eq", t)
	should("be true if both empty", true, func() interface{} {
		return NewLookup().Eq(NewLookup())
	})
	should("be false if only one is empty", false, func() interface{} {
		return NewLookup(1).Eq(NewLookup())
	})
	should("be false if only one is empty", false, func() interface{} {
		return NewLookup().Eq(NewLookup(1))
	})
	should("be true if every el is equal", true, func() interface{} {
		return NewLookup(1).Eq(NewLookup(1))
	})
	should("be not true if every el is not equal", false, func() interface{} {
		return NewLookup(2).Eq(NewLookup(1))
	})
}
