package list

import "fmt"

type node struct {
	val  interface{}
	next *node
}

func (node *node) String() string {
	return fmt.Sprintf("node %v ->", node.val)
}
