package structures

import (
	"fmt"
	"log"
	"reflect"
	"testing"
)

/*
func TestMain(m *testing.M) {
	code := m.Run()
	os.Exit(code)
}
*/

func TestNode(t *testing.T) {
	log.Printf("Testing generalized node")
	// Create 2 new nodes and a bidirectional edge between them
	// with a parent-child relationship
	n1 := NewNode()
	n1.ID = 1
	n2 := NewNode()
	n2.ID = 2
	e1 := NewEdge()
	e1.AddNodes(n1, n2, "parent", "child")
	n1.AddEdge(e1)

	if n1.Edges[0].Nodes[1].ID != n2.ID {
		t.Fatalf(
			fmt.Sprintf("Node %d cannot reach node %d through edge",
				n1.ID,
				n2.ID,
			),
		)
	}

	if !reflect.DeepEqual(n1.Edges[0].Nodes[1].Node, n2) {
		t.Fatalf("Pointer to node %d unreachable through edge from node %d",
			n2.ID,
			n1.ID,
		)
	}

	// Test compare to greater ordered node
	cmpRes := n1.Compare(n2)
	if cmpRes >= 0 {
		t.Fatalf(
			fmt.Sprintf("Node %d should be ordered before node %d",
				n1.ID,
				n2.ID,
			),
		)
	}

	// Test compare to lesser ordered node
	cmpRes = n2.Compare(n1)
	if cmpRes <= 0 {
		t.Fatalf(
			fmt.Sprintf("Node %d should be ordered after node %d",
				n1.ID,
				n2.ID,
			),
		)
	}

	// Test compare to lesser ordered node
	cmpRes = n2.Compare(n2)
	if cmpRes != 0 {
		t.Fatalf(
			fmt.Sprintf("Node %d should be ordered same as node %d",
				n2.ID,
				n2.ID,
			),
		)
	}

	node1, node2 := OrderNodes(n1, n2)
	if !reflect.DeepEqual(n1, node1) || !reflect.DeepEqual(n2, node2) {
		t.Fatalf("Nodes ordered incorrectly")
	}
}
