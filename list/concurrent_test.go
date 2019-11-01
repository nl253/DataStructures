package list

import (
	"testing"

	ut "github.com/nl253/Testing"
)

var fCon = ut.Test("ConcurrentList")

func TestConcurrentList_MapParallelInPlace(t *testing.T) {
	should := fCon("MapParallelInPlace", t)
	should("apply func to each item and modify list in place", New(2, 3, 4), func() interface{} {
		xs := New(1, 2, 3)
		xs.MapParallelInPlace(func(x interface{}, idx uint) interface{} {
			return x.(int) + 1
		})
		return xs
	})
	should("do nothing for empty lists", New(), func() interface{} {
		xs := New()
		xs.MapParallelInPlace(func(x interface{}, idx uint) interface{} {
			return x.(int) + 1
		})
		return xs
	})
}

func TestConcurrentList_Append(t *testing.T) {
	should := fCon("Append", t)
	should("add item to back", true, func() interface{} {
		xs := New()
		xs.Append(0)
		return xs.Eq(New(0))
	})
	should("add item to back", New(1, 0), func() interface{} {
		xs := New(1)
		xs.Append(0)
		return xs
	})
}

func TestConcurrentList_Prepend(t *testing.T) {
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
}

func TestConcurrentList_PopFront(t *testing.T) {
	should := fCon("PopFront", t)
	should("remove 0th item", New(), func() interface{} {
		xs := New(0)
		xs.PopFront()
		return xs
	})
}

func TestConcurrentList_Range(t *testing.T) {
	should := fCon("Range", t)
	should("generate list of ints in bounds", New(0, 1), func() interface{} {
		return Range(0, 2)
	})
	should("generate empty list for equal bounds", New(), func() interface{} {
		return Range(0, 0)
	})
}

func TestConcurrentList_Generate(t *testing.T) {
	should := fCon("Generate", t)
	should("generate list of ints in bounds when using id fLookup", New(0, 1), func() interface{} {
		return Generate(0, 2, func(i int) interface{} {
			return i
		})
	})
	should("generate empty list for equal bounds", New(), func() interface{} {
		return Generate(0, 0, func(i int) interface{} {
			return i
		})
	})
}

func TestConcurrentList_Tail(t *testing.T) {
	should := fCon("Tail", t)
	should("return all but 0th elements", New(), func() interface{} {
		return New(0).Tail()
	})
}

func TestConcurrentList_PeekFront(t *testing.T) {
	should := fCon("PeekFront", t)
	should("return 0th element", 0, func() interface{} {
		return New(0).PeekFront()
	})
}

func TestConcurrentList_PeekBack(t *testing.T) {
	should := fCon("PeekBack", t)
	should("return 0th element", 0, func() interface{} {
		return New(0).PeekBack()

	})
}

func TestConcurrentList_Empty(t *testing.T) {
	should := fCon("Empty", t)
	should("be true for empty list", true, func() interface{} {
		return New().Empty()
	})
	should("be false for non-empty list", false, func() interface{} {
		return New(1).Empty()
	})
}

func TestConcurrentList_Size(t *testing.T) {
	should := fCon("Size", t)
	should("be 0 for empty list", uint(0), func() interface{} {
		return New().Size()
	})
	should("be 1 for list with 1 item", uint(1), func() interface{} {
		return New(1).Size()
	})
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
		xs.Remove(func(i interface{}, u uint) bool {
			return u == 0
		})
		return xs
	})
	should("do nothing when item not in the list (check val)", New(1), func() interface{} {
		xs := New(1)
		xs.Remove(func(i interface{}, u uint) bool {
			return i.(int) == 2
		})
		return xs
	})
	should("return removed item", 1, func() interface{} {
		xs := New(1)
		x, _ := xs.Remove(func(i interface{}, u uint) bool {
			return u == 0
		})
		return x
	})
	should("return index of removed item", 0, func() interface{} {
		xs := New(1)
		_, idx := xs.Remove(func(i interface{}, u uint) bool {
			return i.(int) == 1
		})
		return idx
	})
}

func TestConcurrentList_Find(t *testing.T) {
	should := fCon("Find", t)
	should("return -1 as index when not found (check idx)", -1, func() interface{} {
		xs := New(1)
		_, idx := xs.Find(func(i interface{}, u uint) bool {
			return u == 1
		})
		return idx
	})
	should("return -1 as index when not found (check val)", -1, func() interface{} {
		xs := New(1)
		_, idx := xs.Find(func(i interface{}, u uint) bool {
			return i.(int) == 2
		})
		return idx
	})
	should("return -1 as index when empty (check idx)", -1, func() interface{} {
		xs := New()
		_, idx := xs.Find(func(i interface{}, u uint) bool {
			return u == 0
		})
		return idx
	})
	should("return -1 as index when empty (check val)", -1, func() interface{} {
		xs := New()
		_, idx := xs.Find(func(i interface{}, u uint) bool {
			return i.(int) == 0
		})
		return idx
	})
	should("return index when found (check val)", 0, func() interface{} {
		xs := New(1)
		_, idx := xs.Find(func(i interface{}, u uint) bool {
			return i.(int) == 1
		})
		return idx
	})
	should("return index when found (check idx)", 0, func() interface{} {
		xs := New(1)
		_, idx := xs.Find(func(i interface{}, u uint) bool {
			return u == 0
		})
		return idx
	})
}

func TestConcurrentList_Nth(t *testing.T) {
	should := fCon("Nth", t)
	should("return val", 1, func() interface{} {
		return New(1).Nth(0)
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
