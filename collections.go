package DataStructures

import (
	"fmt"
	"io"
	"time"
)

type ICloneable interface {
	Clone() *IObject
}

type IEq interface {
	Eq() bool
}

type IObject interface {
	fmt.Stringer
	ICloneable
	IEq
	New() *IObject
}

type IContainer interface {
	Clear() *IContainer
	Empty() bool
	Find(func(interface{}, uint) bool) (interface{}, int)
	Remove(func(interface{}, uint) bool) (interface{}, int)
	Contains(interface{}) bool
	Filter(func(interface{}) bool) *IIterable
}

type IIterable interface {
	Size() uint
	Slice(uint, uint) *IIterable
	Nth(uint) interface{}
	ForEach(func(interface{})) *IIterable
	Flatten() *IIterable
	Map(func(interface{}) interface{}) *IList
	Reduce(interface{}, func(interface{}, interface{})) interface{}
}

type IStack interface {
	PushFront(interface{})
	PopFront() interface{}
	PeekFront() interface{}
	Tail() *IStack
}

type IQueue interface {
	PushBack(interface{})
	PopBack() interface{}
	PeekBack() interface{}
	Init() *IQueue
}

type IList interface {
	IContainer
	IIterable
	IStack
	IQueue
}

type IStream interface {
	IObject
	Tap(func(interface{})) *IStream
	Map(func(interface{}) interface{}) *IStream
	Throttle(time.Duration) *IStream
	Spike(uint, time.Duration) *IStream
	Log(io.Writer, string) *IStream
	Printf(string) *IStream
	Println(string) *IStream
}
