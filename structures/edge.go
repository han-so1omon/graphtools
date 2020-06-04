package structures

import (
	"fmt"
	"strings"
)

//FIXME updated type of Nodes to map[string]*Node from []*Node
type Edge struct {
	Value int   `json:"value"`
	Nodes []int `json:"nodevals"`
}

func (e *Edge) String() string {
	var b strings.Builder
	fmt.Fprintf(&b, "-----EDGE %d-----\n", e.Value)
	fmt.Fprintf(&b, "\t%d <=> %d\n", e.Nodes[0], e.Nodes[1])
	fmt.Fprintf(&b, "-----------------\n")
	return b.String()
}

func NewEdge() *Edge {
	e := new(Edge)
	e.Nodes = make([]int, 2)
	return e
}

func (e *Edge) Compare(c Comparable) int8 {
	if e.GetValue() > c.GetValue() {
		return 1
	} else if e.GetValue() < c.GetValue() {
		return -1
	} else {
		return 0
	}
}

func (e *Edge) GetValue() int {
	return e.Value
}

func (e *Edge) AddNodes(n1, n2 *Node) {
	e.Nodes[0] = n1.Value
	e.Nodes[1] = n2.Value
}

func OrderEdges(e1 *Edge, e2 *Edge) (*Edge, *Edge) {
	cmp := e1.Compare(e2)
	if cmp == 1 || cmp == 0 {
		return e1, e2
	}
	return e2, e1
}
