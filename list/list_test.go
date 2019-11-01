package list

import (
	"testing"

	ut "github.com/nl253/Testing"
)

var funct = ut.Mod("ConcurrentList")

func TestConcurrentList_MapParallelInPlace(t *testing.T) {
	should := funct("MapParallelInPlace")
	should("apply func to each item and modify list in place")(ut.Case{
		Expected: true,
		F: func() interface{} {
			xs := New(1, 2, 3)
			xs.MapParallelInPlace(func(x interface{}, idx uint) interface{} {
				return x.(int) + 1
			})
			return xs.Eq(New(2, 3, 4))
		},
	})(t)
	should("do nothing for empty lists")(ut.Case{
		Expected: true,
		F: func() interface{} {
			xs := New()
			xs.MapParallelInPlace(func(x interface{}, idx uint) interface{} {
				return x.(int) + 1
			})
			return xs.Empty()
		},
	})(t)
}

func TestConcurrentList_Append(t *testing.T) {
	should := funct("Append")
	should("add item to back")(ut.Case{
		Expected: true,
		F: func() interface{} {
			xs := New()
			xs.Append(0)
			return xs.Eq(New(0))
		},
	})(t)
	should("add item to back")(ut.Case{
		Expected: true,
		F: func() interface{} {
			xs := New(1)
			xs.Append(0)
			return xs.Eq(New(1, 0))
		},
	})(t)
}

func TestConcurrentList_Prepend(t *testing.T) {
	should := funct("Prepend")
	should("add item to front")(ut.Case{
		Expected: true,
		F: func() interface{} {
			xs := New()
			xs.Prepend(0)
			return xs.Eq(New(0))
		},
	})(t)
	should("add item to front")(ut.Case{
		Expected: true,
		F: func() interface{} {
			xs := New(1)
			xs.Prepend(0)
			return xs.Eq(New(0, 1))
		},
	})(t)
}

func TestConcurrentList_PopFront(t *testing.T) {
	should := funct("PopFront")
	should("remove 0th item")(ut.Case{
		Expected: true,
		F: func() interface{} {
			xs := New(0)
			xs.PopFront()
			return xs.Eq(New())
		},
	})(t)
}

func TestConcurrentList_Range(t *testing.T) {
	should := funct("Range")
	should("generate list of ints in bounds")(ut.Case{
		Expected: true,
		F: func() interface{} {
			return Range(0, 2).Eq(New(0, 1))
		},
	})(t)
	should("generate empty list for equal bounds")(ut.Case{
		Expected: true,
		F: func() interface{} {
			return Range(0, 0).Eq(New())
		},
	})(t)
}

func TestConcurrentList_Generate(t *testing.T) {
	should := funct("Generate")
	should("generate list of ints in bounds when using id funct")(ut.Case{
		Expected: true,
		F: func() interface{} {
			return Generate(0, 2, func(i int) interface{} {
				return i
			}).Eq(New(0, 1))
		},
	})(t)
	should("generate empty list for equal bounds")(ut.Case{
		Expected: true,
		F: func() interface{} {
			return Generate(0, 0, func(i int) interface{} {
				return i
			}).Eq(New())
		},
	})(t)
}

func TestConcurrentList_Tail(t *testing.T) {
	should := funct("Tail")
	should("return all but 0th elements")(ut.Case{
		Expected: true,
		F: func() interface{} {
			return New(0).Tail().Eq(New())
		},
	})(t)
}

func TestConcurrentList_PeekFront(t *testing.T) {
	should := funct("PeekFront")
	should("return 0th element")(ut.Case{
		Expected: 0,
		F: func() interface{} {
			return New(0).PeekFront()
		},
	})(t)
}

func TestConcurrentList_PeekBack(t *testing.T) {
	should := funct("PeekBack")
	should("return 0th element")(ut.Case{
		Expected: 0,
		F: func() interface{} {
			return New(0).PeekBack()
		},
	})(t)
}

func TestConcurrentList_Empty(t *testing.T) {
	should := funct("Empty")
	should("be true for empty list")(ut.Case{
		Expected: true,
		F: func() interface{} {
			return New().Empty()
		},
	})(t)
	should("be false for non-empty list")(ut.Case{
		Expected: false,
		F: func() interface{} {
			return New(1).Empty()
		},
	})(t)
}

func TestConcurrentList_Size(t *testing.T) {
	should := funct("Size")
	should("be 0 for empty list")(ut.Case{
		Expected: uint(0),
		F: func() interface{} {
			return New().Size()
		},
	})(t)
	should("be 1 for list with 1 item")(ut.Case{
		Expected: uint(1),
		F: func() interface{} {
			return New(1).Size()
		},
	})(t)
}

func TestConcurrentList_String(t *testing.T) {
	should := funct("String")
	should("be \"[]\" for empty list")(ut.Case{
		Expected: "[]",
		F: func() interface{} {
			return New().String()
		},
	})(t)
	should("be \"[1]\" for list with `1`")(ut.Case{
		Expected: "[1]",
		F: func() interface{} {
			return New(1).String()
		},
	})(t)
}

func TestConcurrentList_Clear(t *testing.T) {
	should := funct("Clear")
	should("do nothing to empty list")(ut.Case{
		Expected: true,
		F: func() interface{} {
			xs := New()
			sizeAtT1 := xs.Size()
			xs.Clear()
			sizeAtT2 := xs.Size()
			return sizeAtT1 == sizeAtT2
		},
	})(t)
	should("remove items from list and set size to 0")(ut.Case{
		Expected: true,
		F: func() interface{} {
			xs := New(1)
			sizeAtT1 := xs.Size()
			xs.Clear()
			sizeAtT2 := xs.Size()
			return sizeAtT1 != sizeAtT2 && sizeAtT2 == uint(0) && xs.Empty()
		},
	})(t)
}

func TestConcurrentList_Remove(t *testing.T) {
	should := funct("Remove")
	should("do nothing to empty list (check idx)")(ut.Case{
		Expected: uint(0),
		F: func() interface{} {
			xs := New()
			xs.Remove(func(i interface{}, u uint) bool {
				return u == 0
			})
			return xs.Size()
		},
	})(t)
	should("do nothing to empty list (check val)")(ut.Case{
		Expected: uint(0),
		F: func() interface{} {
			xs := New()
			xs.Remove(func(i interface{}, u uint) bool {
				return i == 1
			})
			return xs.Size()
		},
	})(t)
	should("remove item from list and decrement size (check val)")(ut.Case{
		Expected: uint(0),
		F: func() interface{} {
			xs := New(1)
			xs.Remove(func(i interface{}, u uint) bool {
				return i.(int) == 1
			})
			return xs.Size()
		},
	})(t)
	should("remove item from list and decrement size (check idx)")(ut.Case{
		Expected: uint(0),
		F: func() interface{} {
			xs := New(1)
			xs.Remove(func(i interface{}, u uint) bool {
				return u == 0
			})
			return xs.Size()
		},
	})(t)
	should("do nothing when item not in the list (check val)")(ut.Case{
		Expected: uint(1),
		F: func() interface{} {
			xs := New(1)
			xs.Remove(func(i interface{}, u uint) bool {
				return i.(int) == 2
			})
			return xs.Size()
		},
	})(t)
	should("return removed item")(ut.Case{
		Expected: 1,
		F: func() interface{} {
			xs := New(1)
			x, _ := xs.Remove(func(i interface{}, u uint) bool {
				return u == 0
			})
			return x
		},
	})(t)
	should("return index of removed item")(ut.Case{
		Expected: 0,
		F: func() interface{} {
			xs := New(1)
			_, idx := xs.Remove(func(i interface{}, u uint) bool {
				return i.(int) == 1
			})
			return idx
		},
	})(t)
}

func TestConcurrentList_Find(t *testing.T) {
	should := funct("Find")
	should("return -1 as index when not found (check idx)")(ut.Case{
		Expected: -1,
		F: func() interface{} {
			xs := New(1)
			_, idx := xs.Find(func(i interface{}, u uint) bool {
				return u == 1
			})
			return idx
		},
	})(t)
	should("return -1 as index when not found (check val)")(ut.Case{
		Expected: -1,
		F: func() interface{} {
			xs := New(1)
			_, idx := xs.Find(func(i interface{}, u uint) bool {
				return i.(int) == 2
			})
			return idx
		},
	})(t)
	should("return -1 as index when empty (check idx)")(ut.Case{
		Expected: -1,
		F: func() interface{} {
			xs := New()
			_, idx := xs.Find(func(i interface{}, u uint) bool {
				return u == 0
			})
			return idx
		},
	})(t)
	should("return -1 as index when empty (check val)")(ut.Case{
		Expected: -1,
		F: func() interface{} {
			xs := New()
			_, idx := xs.Find(func(i interface{}, u uint) bool {
				return i.(int) == 0
			})
			return idx
		},
	})(t)
	should("return index when found (check val)")(ut.Case{
		Expected: 0,
		F: func() interface{} {
			xs := New(1)
			_, idx := xs.Find(func(i interface{}, u uint) bool {
				return i.(int) == 1
			})
			return idx
		},
	})(t)
	should("return index when found (check idx)")(ut.Case{
		Expected: 0,
		F: func() interface{} {
			xs := New(1)
			_, idx := xs.Find(func(i interface{}, u uint) bool {
				return u == 0
			})
			return idx
		},
	})(t)
}

func TestConcurrentList_Nth(t *testing.T) {
	should := funct("Nth")
	should("return val")(ut.Case{
		Expected: 1,
		F: func() interface{} {
			return New(1).Nth(0)
		},
	})(t)
}

func TestConcurrentList_Eq(t *testing.T) {
	should := funct("Eq")
	should("be true if both empty")(ut.Case{
		Expected: true,
		F: func() interface{} {
			return New().Eq(New())
		},
	})(t)
	should("be false if only one is empty")(ut.Case{
		Expected: false,
		F: func() interface{} {
			return New(1).Eq(New())
		},
	})(t)
	should("be false if only one is empty")(ut.Case{
		Expected: false,
		F: func() interface{} {
			return New().Eq(New(1))
		},
	})(t)
	should("be true if every el is equal")(ut.Case{
		Expected: true,
		F: func() interface{} {
			return New(1).Eq(New(1))
		},
	})(t)
	should("be not true if every el is not equal")(ut.Case{
		Expected: false,
		F: func() interface{} {
			return New(2).Eq(New(1))
		},
	})(t)
}
