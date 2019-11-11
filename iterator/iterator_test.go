package iterator

import (
	"io/ioutil"
	"testing"

	ut "github.com/nl253/Testing"
)

var fIter = ut.Test("Iterator")

const N uint = 10000

func TestIterator_Range(t *testing.T) {
	should := fIter("Range", t)
	should("make iter of ints in range [min, max)", 0, func() interface{} {
		it := Ints()
		return it.Pull()
	})
	should("make iter of ints in range [min, max)", 1, func() interface{} {
		it := Ints()
		it.Skip()
		return it.Pull()
	})
	should("make iter of ints in range [min, max)", 2, func() interface{} {
		it := Ints()
		it.SkipN(2)
		return it.Pull()
	})
}

func TestIterator_Slice(t *testing.T) {
	should := fIter("Take", t)
	should("slice [, 0) should give EndOfIteration", EndOfIteration, func() interface{} {
		return Ints().Take(0).Pull()
	})
	should("slice [, 1) should give 0", 0, func() interface{} {
		return Ints().Take(1).Pull()
	})
	should("slice [, 1).Skip() should give EndOfIteration", EndOfIteration, func() interface{} {
		return Ints().Println().Take(1).Skip().Pull()
	})
	should("slice [, 1).Skip() should give EndOfIteration", 2, func() interface{} {
		return Ints().Take(3).SkipN(2).Pull()
	})
}

func TestIterator_Sum(t *testing.T) {
	should := fIter("Sum", t)
	should("make iter of ints in range [min, max)", 0, func() interface{} {
		return Ints().Take(1).Sum()
	})
	should("make iter of ints in range [min, max)", 0+1+2, func() interface{} {
		return Ints().Take(3).Sum()
	})
}

func TestIterator_ForEach(t *testing.T) {
	should := fIter("ForEach", t)
	should("sum", 0, func() interface{} {
		return Ints().Take(0).Close().Map(func(i interface{}) interface{} {
			return 1
		}).Sum()
	})
}

func TestIterator_Map(t *testing.T) {
	should := fIter("Map", t)
	should("id should not modify values", 0, func() interface{} {
		return Ints().Map(func(x interface{}) interface{} {
			return x
		}).Map(func(x interface{}) interface{} {
			return x
		}).Take(1).Println().Pull()
	})
	should("id should not modify values", EndOfIteration, func() interface{} {
		return Ints().Map(func(x interface{}) interface{} {
			return x
		}).Map(func(x interface{}) interface{} {
			return x
		}).Take(1).Skip().Println().Pull()
	})
	should("id should not modify values", 1, func() interface{} {
		return Ints().Map(func(x interface{}) interface{} {
			return x
		}).Map(func(x interface{}) interface{} {
			return x
		}).Take(2).Skip().Println().Pull()
	})
	should("make iter of ints in range [min, max)", 0+10+1+10, func() interface{} {
		return Ints().Map(func(x interface{}) interface{} { return x.(int) + 10 }).Take(2).Println().Sum()
	})
}

func TestIterator_FromFile(t *testing.T) {
	should := fIter("FromFile", t)
	fileName := "/home/mx/go/src/github.com/nl253/DataStructures/iterator/iterator_test.go"
	should("iter over file bytes", []interface{}{byte('p'), byte('a'), byte('c'), byte('k'), byte('a'), byte('g'), byte('e')}, func() interface{} {
		return FromFile(fileName).Take(7).PullN(7)
	})
	bytes, _ := ioutil.ReadFile(fileName)
	should("iter over file bytes", bytes, func() interface{} {
		slice := FromFile(fileName).PullAll().ToSlice()
		slice = slice[:len(slice)-1]
		n := len(slice)
		bs := make([]byte, n, n)
		for i := 0; i < n; i++ {
			bs[i] = slice[i].(byte)
		}
		return bs
	})
}

func TestIterator_FromStr(t *testing.T) {
	should := fIter("FromStr", t)
	should("iter over str bytes", []interface{}{byte('p'), byte('a'), byte('c'), byte('k'), byte('a'), byte('g'), byte('e')}, func() interface{} {
		return FromStr("package").Take(7).PullN(7)
	})
	should("iter over empty str bytes", []interface{}{}, func() interface{} {
		return FromStr("package").Take(0).PullN(0)
	})
	should("iter over empty str bytes", []interface{}{}, func() interface{} {
		return FromStr("").Take(0).PullN(0)
	})
}
