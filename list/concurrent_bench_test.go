package list

import "testing"

const N uint = 1000000

func BenchmarkConcurrentList_PopFront(b *testing.B) {
	n := N
	s := Range(0, int(n), 1)
	for i := uint(0); i < n; i++ {
		s.PopFront()
	}
}

func BenchmarkConcurrentList_PeekFront(b *testing.B) {
	n := N
	s := Ints(0, 100, n)
	for i := 0; i < 1000; i++ {
		s.PeekFront()
	}
}

func BenchmarkConcurrentList_Append(b *testing.B) {
	n := N
	s := New()
	for i := uint(0); i < n; i++ {
		s.Append(i)
	}
}

func BenchmarkConcurrentList_Prepend(b *testing.B) {
	n := N
	s := New()
	for i := uint(0); i < n; i++ {
		s.Prepend(i)
	}
}

func BenchmarkConcurrentList_Nth(b *testing.B) {
	n := N
	s := Range(0, int(n), 1)
	for i := 0; i < 1000; i++ {
		s.Nth(n - 1)
	}
}
