package Channel

import "testing"

const N uint = 1000000

func generate() chan interface{} {
	c := make(chan interface{}, N+1)
	for i := uint(0); i < N; i++ {
		c <- i
	}
	return c
}

func BenchmarkChannel_PopFront(b *testing.B) {
	c := generate()
	for i := uint(0); i < N; i++ {
		<-c
	}
}

func BenchmarkChannel_Append(b *testing.B) {
	c := make(chan interface{}, N+1)
	for i := uint(0); i < N; i++ {
		c <- i
	}
}
