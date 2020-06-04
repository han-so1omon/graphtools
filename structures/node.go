package structures

import (
	"fmt"
	"strings"
)

type Point struct {
	X int `json:"x"`
	Y int `json:"y"`
}

//FIXME updated type of Edges to map[string]*Edge from map[*Node]*Edge
type Node struct {
	Value  int     `json:"value"`
	Coords Point   `json:"coords"`
	Edges  []*Edge `json:"edges"`
}

func (n *Node) String() string {
	var b strings.Builder
	fmt.Fprintf(&b, ".....NODE %d : (%d,%d).....\n", n.Value, n.Coords.X, n.Coords.Y)
	for _, e := range n.Edges {
		fmt.Fprintf(&b, "\t%d <= {%d} => %d\n", e.Nodes[0], e.Value, e.Nodes[1])
	}
	fmt.Fprintf(&b, "...........................\n")
	return b.String()
}

func NewNode() *Node {
	n := new(Node)
	return n
}

func (n *Node) Compare(c Comparable) int8 {
	if n.GetValue() > c.GetValue() {
		return 1
	} else if n.GetValue() < c.GetValue() {
		return -1
	}
	return 0
}

func (n *Node) GetValue() int {
	return n.Value
}

func (n *Node) AddEdge(e *Edge) {
	n.Edges = append(n.Edges, e)
}

func OrderNodes(n1 *Node, n2 *Node) (*Node, *Node) {
	cmp := n1.Compare(n2)
	if cmp == 1 || cmp == 0 {
		return n1, n2
	}
	return n2, n1

}
