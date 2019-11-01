package list

import (
	"testing"

	ut "github.com/nl253/Testing"
)

var fLookup = ut.Mod("LookupList")

func TestLookupList_MapParallelInPlace(t *testing.T) {
	should := fLookup("MapParallelInPlace")
	should("apply func to each item and modify list in place")(ut.Case{
		Expected: true,
		F: func() interface{} {
			xs := NewLookup(1, 2, 3)
			xs.MapParallelInPlace(func(x interface{}, idx uint) interface{} {
				return x.(int) + 1
			})
			return xs.Eq(NewLookup(2, 3, 4))
		},
	})(t)
	should("do nothing for empty lists")(ut.Case{
		Expected: true,
		F: func() interface{} {
			xs := NewLookup()
			xs.MapParallelInPlace(func(x interface{}, idx uint) interface{} {
				return x.(int) + 1
			})
			return xs.Empty()
		},
	})(t)
}

func TestLookupList_Append(t *testing.T) {
	should := fLookup("Append")
	should("add item to back")(ut.Case{
		Expected: true,
		F: func() interface{} {
			xs := NewLookup()
			xs.Append(0)
			return xs.Eq(NewLookup(0))
		},
	})(t)
	should("add item to back")(ut.Case{
		Expected: true,
		F: func() interface{} {
			xs := NewLookup(1)
			xs.Append(0)
			return xs.Eq(NewLookup(1, 0))
		},
	})(t)
}

func TestLookupList_Prepend(t *testing.T) {
	should := fLookup("Prepend")
	should("add item to front")(ut.Case{
		Expected: true,
		F: func() interface{} {
			xs := NewLookup()
			xs.Prepend(0)
			return xs.Eq(NewLookup(0))
		},
	})(t)
	should("add item to front")(ut.Case{
		Expected: true,
		F: func() interface{} {
			xs := NewLookup(1)
			xs.Prepend(0)
			return xs.Eq(NewLookup(0, 1))
		},
	})(t)
}

func TestLookupList_PopFront(t *testing.T) {
	should := fLookup("PopFront")
	should("remove 0th item")(ut.Case{
		Expected: true,
		F: func() interface{} {
			xs := NewLookup(0)
			xs.PopFront()
			return xs.Eq(NewLookup())
		},
	})(t)
}

func TestLookupList_RangeLookup(t *testing.T) {
	should := fLookup("RangeLookup")
	should("generateLookup list of ints in bounds")(ut.Case{
		Expected: true,
		F: func() interface{} {
			return RangeLookup(0, 2).Eq(NewLookup(0, 1))
		},
	})(t)
	should("generateLookup empty list for equal bounds")(ut.Case{
		Expected: true,
		F: func() interface{} {
			return RangeLookup(0, 0).Eq(NewLookup())
		},
	})(t)
}

func TestLookupList_GenerateLookup(t *testing.T) {
	should := fLookup("GenerateLookup")
	should("generateLookup list of ints in bounds when using id fLookup")(ut.Case{
		Expected: true,
		F: func() interface{} {
			return GenerateLookup(0, 2, func(i int) interface{} {
				return i
			}).Eq(NewLookup(0, 1))
		},
	})(t)
	should("generateLookup empty list for equal bounds")(ut.Case{
		Expected: true,
		F: func() interface{} {
			return GenerateLookup(0, 0, func(i int) interface{} {
				return i
			}).Eq(NewLookup())
		},
	})(t)
}

func TestLookupList_Tail(t *testing.T) {
	should := fLookup("Tail")
	should("return all but 0th elements")(ut.Case{
		Expected: true,
		F: func() interface{} {
			return NewLookup(0).Tail().Eq(NewLookup())
		},
	})(t)
}

func TestLookupList_PeekFront(t *testing.T) {
	should := fLookup("PeekFront")
	should("return 0th element")(ut.Case{
		Expected: 0,
		F: func() interface{} {
			return NewLookup(0).PeekFront()
		},
	})(t)
}

func TestLookupList_PeekBack(t *testing.T) {
	should := fLookup("PeekBack")
	should("return 0th element")(ut.Case{
		Expected: 0,
		F: func() interface{} {
			return NewLookup(0).PeekBack()
		},
	})(t)
}

func TestLookupList_Empty(t *testing.T) {
	should := fLookup("Empty")
	should("be true for empty list")(ut.Case{
		Expected: true,
		F: func() interface{} {
			return NewLookup().Empty()
		},
	})(t)
	should("be false for non-empty list")(ut.Case{
		Expected: false,
		F: func() interface{} {
			return NewLookup(1).Empty()
		},
	})(t)
}

func TestLookupList_Size(t *testing.T) {
	should := fLookup("Size")
	should("be 0 for empty list")(ut.Case{
		Expected: uint(0),
		F: func() interface{} {
			return NewLookup().Size()
		},
	})(t)
	should("be 1 for list with 1 item")(ut.Case{
		Expected: uint(1),
		F: func() interface{} {
			return NewLookup(1).Size()
		},
	})(t)
}

func TestLookupList_String(t *testing.T) {
	should := fLookup("String")
	should("be \"[]\" for empty list")(ut.Case{
		Expected: "[]",
		F: func() interface{} {
			return NewLookup().String()
		},
	})(t)
	should("be \"[1]\" for list with `1`")(ut.Case{
		Expected: "[1]",
		F: func() interface{} {
			return NewLookup(1).String()
		},
	})(t)
}

func TestLookupList_Clear(t *testing.T) {
	should := fLookup("Clear")
	should("do nothing to empty list")(ut.Case{
		Expected: true,
		F: func() interface{} {
			xs := NewLookup()
			sizeAtT1 := xs.Size()
			xs.Clear()
			sizeAtT2 := xs.Size()
			return sizeAtT1 == sizeAtT2
		},
	})(t)
	should("remove items from list and set size to 0")(ut.Case{
		Expected: true,
		F: func() interface{} {
			xs := NewLookup(1)
			sizeAtT1 := xs.Size()
			xs.Clear()
			sizeAtT2 := xs.Size()
			return sizeAtT1 != sizeAtT2 && sizeAtT2 == uint(0) && xs.Empty()
		},
	})(t)
}

func TestLookupList_Remove(t *testing.T) {
	should := fLookup("Remove")
	should("do nothing to empty list (check idx)")(ut.Case{
		Expected: uint(0),
		F: func() interface{} {
			xs := NewLookup()
			xs.Remove(func(i interface{}, u uint) bool {
				return u == 0
			})
			return xs.Size()
		},
	})(t)
	should("do nothing to empty list (check val)")(ut.Case{
		Expected: uint(0),
		F: func() interface{} {
			xs := NewLookup()
			xs.Remove(func(i interface{}, u uint) bool {
				return i == 1
			})
			return xs.Size()
		},
	})(t)
	should("remove item from list and decrement size (check val)")(ut.Case{
		Expected: uint(0),
		F: func() interface{} {
			xs := NewLookup(1)
			xs.Remove(func(i interface{}, u uint) bool {
				return i.(int) == 1
			})
			return xs.Size()
		},
	})(t)
	should("remove item from list and decrement size (check idx)")(ut.Case{
		Expected: uint(0),
		F: func() interface{} {
			xs := NewLookup(1)
			xs.Remove(func(i interface{}, u uint) bool {
				return u == 0
			})
			return xs.Size()
		},
	})(t)
	should("do nothing when item not in the list (check val)")(ut.Case{
		Expected: uint(1),
		F: func() interface{} {
			xs := NewLookup(1)
			xs.Remove(func(i interface{}, u uint) bool {
				return i.(int) == 2
			})
			return xs.Size()
		},
	})(t)
	should("return removed item")(ut.Case{
		Expected: 1,
		F: func() interface{} {
			xs := NewLookup(1)
			x, _ := xs.Remove(func(i interface{}, u uint) bool {
				return u == 0
			})
			return x
		},
	})(t)
	should("return index of removed item")(ut.Case{
		Expected: 0,
		F: func() interface{} {
			xs := NewLookup(1)
			_, idx := xs.Remove(func(i interface{}, u uint) bool {
				return i.(int) == 1
			})
			return idx
		},
	})(t)
}

func TestLookupList_Find(t *testing.T) {
	should := fLookup("Find")
	should("return false when not found (check idx)")(ut.Case{
		Expected: false,
		F: func() interface{} {
			xs := NewLookup(1)
			_, ok := xs.Find(func(i interface{}, u uint) bool {
				return u == 1
			})
			return ok
		},
	})(t)
	should("return false when not found (check val)")(ut.Case{
		Expected: false,
		F: func() interface{} {
			xs := NewLookup(1)
			_, ok := xs.Find(func(i interface{}, u uint) bool {
				return i.(int) == 2
			})
			return ok
		},
	})(t)
	should("return false when empty (check idx)")(ut.Case{
		Expected: false,
		F: func() interface{} {
			xs := NewLookup()
			_, ok := xs.Find(func(i interface{}, u uint) bool {
				return u == 0
			})
			return ok
		},
	})(t)
	should("return false when empty (check val)")(ut.Case{
		Expected: false,
		F: func() interface{} {
			xs := NewLookup()
			_, ok := xs.Find(func(i interface{}, u uint) bool {
				return i.(int) == 0
			})
			return ok
		},
	})(t)
	should("return true when found (check val)")(ut.Case{
		Expected: true,
		F: func() interface{} {
			xs := NewLookup(1)
			_, ok := xs.Find(func(i interface{}, u uint) bool {
				return i.(int) == 1
			})
			return ok
		},
	})(t)
	should("return true when found (check idx)")(ut.Case{
		Expected: true,
		F: func() interface{} {
			xs := NewLookup(1)
			_, idx := xs.Find(func(i interface{}, u uint) bool {
				return u == 0
			})
			return idx
		},
	})(t)
	should("reorder when found (check val)")(ut.Case{
		Expected: true,
		F: func() interface{} {
			xs := NewLookup(0, 1)
			xs.Find(func(i interface{}, u uint) bool {
				return i.(int) == 1
			})
			return xs.Eq(NewLookup(1, 0))
		},
	})(t)
	should("reorder when found (check idx)")(ut.Case{
		Expected: true,
		F: func() interface{} {
			xs := NewLookup(0, 1)
			xs.Find(func(i interface{}, u uint) bool {
				return u == 1
			})
			return xs.Eq(NewLookup(1, 0))
		},
	})(t)
}

func TestLookupList_Nth(t *testing.T) {
	should := fLookup("Nth")
	should("return val")(ut.Case{
		Expected: 1,
		F: func() interface{} {
			return NewLookup(1).Nth(0)
		},
	})(t)
}

func TestLookupList_Eq(t *testing.T) {
	should := fLookup("Eq")
	should("be true if both empty")(ut.Case{
		Expected: true,
		F: func() interface{} {
			return NewLookup().Eq(NewLookup())
		},
	})(t)
	should("be false if only one is empty")(ut.Case{
		Expected: false,
		F: func() interface{} {
			return NewLookup(1).Eq(NewLookup())
		},
	})(t)
	should("be false if only one is empty")(ut.Case{
		Expected: false,
		F: func() interface{} {
			return NewLookup().Eq(NewLookup(1))
		},
	})(t)
	should("be true if every el is equal")(ut.Case{
		Expected: true,
		F: func() interface{} {
			return NewLookup(1).Eq(NewLookup(1))
		},
	})(t)
	should("be not true if every el is not equal")(ut.Case{
		Expected: false,
		F: func() interface{} {
			return NewLookup(2).Eq(NewLookup(1))
		},
	})(t)
}
