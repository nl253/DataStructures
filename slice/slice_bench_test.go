package slice

import "testing"

const N uint = 1000000

func none(x interface{}) {}

func BenchmarkSlice_PopFront(b *testing.B) {
	n := N
	s := make([]interface{}, n)
	for i := uint(0); i < n; i++ {
		s[0] = 0
		s = s[1:]
	}
}

func BenchmarkSlice_PeekFront(b *testing.B) {
	n := N
	s := make([]int, n, n)
	for i := 0; i < 1000; i++ {
		none(s[0])
	}
}

func BenchmarkSlice_Append(b *testing.B) {
	n := N
	s := make([]interface{}, 0)
	for i := uint(0); i < n; i++ {
		s = append(s, i)
	}
}

func BenchmarkSlice_Prepend(b *testing.B) {
	n := N
	s := make([]int, 0)
	for i := uint(0); i < n; i++ {
		s = append([]int{0}, s...)
	}
}

func BenchmarkSlice_Nth(b *testing.B) {
	n := N
	s := make([]int, n)
	for i := 0; i < 1000; i++ {
		none(s[n-1])
	}
}

func Benchmark_Index_Slice(b *testing.B) {
	n := uint(100000)
	s := make([]interface{}, n, n)
	for i := uint(0); i < n; i++ {
		none(s[0])
		s = s[1:]
	}
}
