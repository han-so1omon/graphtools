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

func TestEdge(t *testing.T) {
	log.Printf("Testing generalized edge")
	e1 := NewEdge()
	n1 := NewNode()
	n1.ID = 1
	n2 := NewNode()
	n2.ID = 2

	// Test add nodes to edge
	e1.AddNodes(n1, n2, "parent", "child")
	if e1.Nodes[0].ID != e1.Nodes[0].Node.ID {
		t.Fatalf("Node representation ID must be equal to node ID")
	}
	if e1.Nodes[0].Tag != "parent" && e1.Nodes[1].Tag != "child" {
		t.Fatalf("Edge tags set incorrectly")
	}

	n3 := NewNode()
	n3.ID = 3
	e2 := NewEdge()
	e2.AddNodes(n1, n3, "parent", "child")

	// Test compare to lesser edge
	cmpRes := e1.Compare(e2)
	if cmpRes >= 0 {
		t.Fatalf(
			fmt.Sprintf("Edge connecting %d->%d should be ordered before edge connecting %d->%d",
				n1.ID, n2.ID,
				n1.ID, n3.ID,
			),
		)
	}

	// Test compare to greater edge
	cmpRes = e2.Compare(e1)
	if cmpRes <= 0 {
		t.Fatalf(
			fmt.Sprintf("Edge connecting %d->%d should be ordered after edge connecting %d->%d",
				n1.ID, n3.ID,
				n1.ID, n2.ID,
			),
		)
	}

	// Test compare to equal edge
	cmpRes = e2.Compare(e2)
	if cmpRes != 0 {
		t.Fatalf(
			fmt.Sprintf("Edge connecting %d->%d should be ordered same as edge connecting %d->%d",
				n1.ID, n3.ID,
				n1.ID, n3.ID,
			),
		)
	}

	edge1, edge2 := OrderEdges(e1, e2)
	if !reflect.DeepEqual(e1, edge1) || !reflect.DeepEqual(e2, edge2) {
		t.Fatalf("Edges ordered incorrectly")
	}
}
