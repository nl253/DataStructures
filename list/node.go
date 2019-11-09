package list

import "fmt"

type node struct {
	val  interface{}
	next *node
}

func (n *node) Eq(x interface{}) bool {
	switch x.(type) {
	case *node:
		return n == x.(*node)
	default:
		return false
	}
}

func (n *node) String() string {
	return fmt.Sprintf("node %v ->", n.val)
}
