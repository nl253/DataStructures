package DataStructures

import (
	"fmt"
)

type Cloneable interface {
	Clone() *Cloneable
}

type Equatable interface {
	Eq() bool
}

type IObject interface {
	fmt.Stringer
	Cloneable
	Equatable
	New() *IObject
}
