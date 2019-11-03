package DataStructures

type ICollection interface {
    New() *ICollection
    Clone() *ICollection
}

type ISized interface {
    ICollection
    Size() uint
    Empty() bool
}

type IIterable interface {
    ISized
    Find(func(interface{}, uint) bool) (interface{}, int)
    Remove(func(interface{}, uint) bool) (interface{}, int)
    Contains(interface{}) bool
    Nth(uint) interface{}
}

type IList interface {
    IIterable
    Append(x interface{})
    Prepend(x interface{})
    Tail() *IList
    PopFront() interface{}
    PeekFront() interface{}
    Map(func(interface{}) interface{}) *IList
    ForEach(func(interface{}))
    Reduce(interface{}, func(interface{}, interface{})) interface{}
}
