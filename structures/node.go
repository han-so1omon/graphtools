package structures

import (
	"fmt"
	"strings"
)

// Data is extra information required per node.
// It may be useful for graph organization, e.g. if making a balanced tree.
type Data interface {
	// GetData returns the appropriate representation of data
	GetData() interface{}
	// DeleteData deletes internal data. It is the responsibility of the
	// implementing type to handle any loose pointers to data that may exist
	DeleteData()
}

// DataError states that the requested data is formatted incorrectly
type DataError struct{}

func (e DataError) Error() string {
	return "Node: problem with data"
}

type IDDistributor interface {
	// GetID creates a new integer ID with the option to partition ID values by
	// type, denoted by a string parameter
	GetID(string) int
}

// Point is a cartesian demarcation of a node.
// It is useful for displaying a node
type Point struct {
	X int `json:"x"`
	Y int `json:"y"`
	Z int `json:"z"`
}

type Node struct {
	ID     int     `json:"id"`
	Extra  Data    `json:"extra"`
	Coords Point   `json:"coords"`
	Edges  []*Edge `json:"edges"`
}

func (n *Node) String() string {
	var b strings.Builder
	fmt.Fprintf(&b, ".....NODE %d : (%d,%d,%d).....\n", n.ID, n.Coords.X, n.Coords.Y, n.Coords.Z)
	for _, e := range n.Edges {
		fmt.Fprintf(
			&b,
			"\t%d <= {%s-%s: %f} => %d\n",
			e.Nodes[0].ID, e.Nodes[0].Tag, e.Nodes[1].Tag, e.Weight, e.Nodes[1].ID,
		)
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
	return n.ID
}

func (n *Node) AddEdge(e *Edge) {
	n.Edges = append(n.Edges, e)
}

func OrderNodes(n1 *Node, n2 *Node) (*Node, *Node) {
	cmp := n1.Compare(n2)
	if cmp <= 0 {
		return n1, n2
	}
	return n2, n1
}
