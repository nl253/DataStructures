package stream

import "testing"

const N uint = 1000000

func BenchmarkStream_Pull(b *testing.B) {
	n := N
	s := RandInts(0, 100, n)
	for i := 0; i < 1000; i++ {
		s.Pull()
	}
}

func BenchmarkStream_Peek(b *testing.B) {
	n := N
	s := RandInts(0, 100, n)
	for i := 0; i < 1000; i++ {
		s.PeekFront()
	}
}

func BenchmarkStream_PushBack(b *testing.B) {
	n := N
	s := New()
	for i := uint(0); i < n; i++ {
		s.PushBack(i)
	}
}

func BenchmarkStream_PushFront(b *testing.B) {
	n := N
	s := New()
	for i := uint(0); i < n; i++ {
		s.PushFront(i)
	}
}
