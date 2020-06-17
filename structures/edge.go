package structures

import (
	"fmt"
	"strings"
)

// NodeRepr is a representation of a node. It is useful for packing Node-Edge
// structures into messages without cycles. It is also useful for quick lookup
// of node IDs and tags associated with an edge
type NodeRepr struct {
	Node *Node  `json:"-"` // Omit this from JSON to prevent cyclic structure
	ID   int    `json:"id"`
	Tag  string `json:"tag"`
}

// Edge is defined as the connection between two nodes
// Edge is uni-directional
// Weight holds the value of the connection, which may indicate difficulty
// or strength
// Nodes holds the representation of the connecting nodes
// The first node is referred to as near, while the second node is referred to
// as far
type Edge struct {
	Weight float64    `json:"weight"`
	Nodes  []NodeRepr `json:"noderepr"`
}

// String is the string representation of an edge. This is useful formatted
// printing
func (e *Edge) String() string {
	var b strings.Builder
	fmt.Fprintf(&b, "-----EDGE %f-----\n", e.Weight)
	fmt.Fprintf(&b, "\t%d <=> %d\n", e.Nodes[0].ID, e.Nodes[1].ID)
	fmt.Fprintf(&b, "-----------------\n")
	return b.String()
}

// NewEdge creates a blank new edge
func NewEdge() *Edge {
	e := new(Edge)
	e.Nodes = make([]NodeRepr, 2)
	return e
}

// Compare compares an edge with another comparable value
// Returns 1 if value is greater than comparable value
//		  -1 if value is less than comparable value
//		   0 if value is equal to comparable value
func (e *Edge) Compare(c Comparable) int8 {
	if e.GetValue() > c.GetValue() {
		return 1
	} else if e.GetValue() < c.GetValue() {
		return -1
	} else {
		return 0
	}
}

// Ordering valuation is by id of far node
func (e *Edge) GetValue() int {
	return e.Nodes[1].ID
}

// AddNodes adds n1 to the near node of edge e and adds
// n2 to the far node of edge e
func (e *Edge) AddNodes(n1, n2 *Node, t1, t2 string) {
	e.Nodes[0].Node = n1
	e.Nodes[0].ID = n1.ID
	e.Nodes[0].Tag = t1
	e.Nodes[1].Node = n2
	e.Nodes[1].ID = n2.ID
	e.Nodes[1].Tag = t2
}

// OrderEdges orders edges smallest to largest
func OrderEdges(e1 *Edge, e2 *Edge) (*Edge, *Edge) {
	cmp := e1.Compare(e2)
	if cmp <= 0 {
		return e1, e2
	}
	return e2, e1
}
