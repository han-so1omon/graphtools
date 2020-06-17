package structures

import (
	"fmt"
	"log"
	"os"
	"reflect"
	"testing"
)

func TestMain(m *testing.M) {
	code := m.Run()
	os.Exit(code)
}

func TestGraph(t *testing.T) {
	log.Printf("Testing generalized graph")

	t.Run("Set graph nodes", func(t *testing.T) {
		g := NewGraph(100)
		if !g.IsEmpty() {
			t.Fatalf("Graph does not start off in an empty state")
		}

		// Check basic data mocking
		n1 := NewNode()
		data1 := mockData{21}
		// Add node to graph using setNodeHelper
		g.setNodeHelper(n1, -1, -1, -1, -1, data1)
		gHasN1 := g.HasNodeWithID(-1)
		if !gHasN1 {
			t.Fatalf(fmt.Sprintf("Could not detect node %d in graph", -1))
		}
		data1FromNode, ok := mockDataFromData(n1.Extra)
		if !ok || !reflect.DeepEqual(data1, data1FromNode) {
			t.Fatalf(fmt.Sprintf("Could not retrieve data correctly from node %d", -1))
		}
		// Check retrieving node from graph
		n1FromGraph, err := g.GetNodeByID(-1)
		if err != nil || !reflect.DeepEqual(n1FromGraph, n1) {
			t.Fatalf(fmt.Sprintf("Could not retrieve node %d correctly from graph", -1))
		}

		// Add second node to graph with different mock data
		n2 := NewNode()
		data2 := mockData{42}
		// Use SetNode this time
		g.SetNode(n2, -2, -2, -2, -2, data2)

		data1FromGraph, ok := mockDataFromData(n1FromGraph.Extra)
		if !ok || !reflect.DeepEqual(data1, data1FromGraph) {
			t.Fatalf(fmt.Sprintf("Could not retrieve data correctly from graph for node %d", -2))
		}

		// Check retrieving second node from graph
		n2FromGraph, err := g.GetNodeByID(-2)
		if err != nil || !reflect.DeepEqual(n2FromGraph, n2) {
			t.Fatalf(fmt.Sprintf("Could not retrieve node %d correctly from graph", -2))
		}
		data2FromGraph, ok := mockDataFromData(n2FromGraph.Extra)
		if !ok || !reflect.DeepEqual(data2, data2FromGraph) {
			t.Fatalf(fmt.Sprintf("Could not retrieve data correctly from graph for node %d", -2))
		}

		// Remove node 1 with RemoveNode
		g.RemoveNode(n1)
		_, err = g.GetNodeByID(-1)
		_, ok = err.(NoNodeError)
		if !ok {
			t.Fatalf("Node not removed from graph properly")
		}

		// Remove node 2 with RemoveNodeByID
		g.RemoveNodeByID(-2)
		_, err = g.GetNodeByID(-2)
		_, ok = err.(NoNodeError)
		if !ok {
			t.Fatalf("Node not removed from graph properly by ID")
		}

		if !g.IsEmpty() {
			t.Fatalf("Graph nodes not cleared correctly")
		}

		// Set 100 nodes with SetNodeByID
		data3 := mockData{42}
		for i := 0; i < 100; i++ {
			g.SetNodeByID(i, i, i, 0, data3)
		}
		for i := 0; i < 100; i++ {
			_, err := g.GetNodeByID(i)
			if err != nil {
				t.Fatalf(fmt.Sprintf("Could not retrieve node %d correctly from graph", i))
			}
			if g.NumNodes != 100 {
				t.Fatalf(fmt.Sprintf("Graph should have %d nodes", 100))
			}
		}

		// Remove 50 nodes with RemoveNode
		for i := 0; i < 50; i++ {
			n, _ := g.GetNodeByID(i)
			g.RemoveNode(n)
			_, err := g.GetNodeByID(i)
			_, ok = err.(NoNodeError)
			if !ok {
				t.Fatalf("Node not removed from graph properly")
			}
		}

		// Remove remaining 50 nodes with RemoveNodeByID
		for i := 50; i < 100; i++ {
			g.RemoveNodeByID(i)
			_, err := g.GetNodeByID(i)
			_, ok = err.(NoNodeError)
			if !ok {
				t.Fatalf("Node not removed from graph properly")
			}
		}

		if !g.IsEmpty() {
			t.Fatalf("Graph should be empty")
		}
	})

	t.Run("Graph node edges", func(t *testing.T) {
		g := NewGraph(100)
		// Set 100 nodes with SetNodeByID
		data := mockData{42}
		for i := 0; i < 100; i++ {
			g.SetNodeByID(i, i, i, 0, data)
		}

		n1, _ := g.GetNodeByID(10)
		n2, _ := g.GetNodeByID(30)
		err := g.setEdgeHelper2(n1, n2, 12, "parent", "child")
		if err != nil {
			t.Fatalf(fmt.Sprintf("Could not add edge from %d to %d", n1.ID, n2.ID))
		}
		checkEdge(t, g, n1, n2, 12, "parent", "child")

		n3, _ := g.GetNodeByID(50)
		err = g.setEdgeHelper(n1, n3, 21, "parent", "child", true)
		if err != nil {
			t.Fatalf(fmt.Sprintf("Could not add bidirectional edge from %d to %d", n1.ID, n3.ID))
		}
		checkEdge(t, g, n1, n3, 21, "parent", "child")
		checkEdge(t, g, n3, n1, 21, "child", "parent")

		n4, _ := g.GetNodeByID(70)
		err = g.SetEdge(n2, n4, 33, "big", "ups", true)
		if err != nil {
			t.Fatalf(fmt.Sprintf("Could not add bidirectional edge from %d to %d", n2.ID, n4.ID))
		}
		checkEdge(t, g, n2, n4, 33, "big", "ups")
		checkEdge(t, g, n4, n2, 33, "ups", "big")

		// Try setting edge with new node in place of existing edge
		n5, _ := g.GetNodeByID(75)
		err = g.SetEdge(n2, n5, 31, "big", "ups", true)
		if err != nil {
			t.Fatalf(fmt.Sprintf("Could not add bidirectional edge from %d to %d", n2.ID, n5.ID))
		}
		checkEdge(t, g, n2, n5, 31, "big", "ups")
		checkEdge(t, g, n5, n2, 31, "ups", "big")

		err = g.SetEdgeByNodeID(n2.ID, n3.ID, 77, "bad", "motha", true)
		if err != nil {
			t.Fatalf(fmt.Sprintf("Could not add bidirectional edge from %d to %d", n2.ID, n3.ID))
		}
		checkEdge(t, g, n2, n3, 77, "bad", "motha")
		checkEdge(t, g, n3, n2, 77, "motha", "bad")

		err = g.removeEdgeHelper2(n1, n2)
		if err != nil {
			t.Fatalf(
				fmt.Sprintf("Could not successfully remove edge from %d to %d",
					n1.ID, n2.ID,
				),
			)
		}
		checkRemoveEdge(t, g, n1, n2)

		err = g.removeEdgeHelper(n1, n3, true)
		if err != nil {
			t.Fatalf(
				fmt.Sprintf("Could not successfully remove bidirectional edge from %d to %d",
					n1.ID, n3.ID,
				),
			)
		}
		checkRemoveEdge(t, g, n1, n3)
		checkRemoveEdge(t, g, n3, n1)

		err = g.RemoveEdge(n2, n4, true)
		if err != nil {
			t.Fatalf(
				fmt.Sprintf("Could not successfully remove bidirectional edge from %d to %d",
					n1.ID, n3.ID,
				),
			)
		}
		checkRemoveEdge(t, g, n2, n4)
		checkRemoveEdge(t, g, n4, n2)

		err = g.RemoveEdgeByNodeID(n2.ID, n3.ID, true)
		if err != nil {
			t.Fatalf(
				fmt.Sprintf("Could not successfully remove bidirectional edge from %d to %d",
					n1.ID, n3.ID,
				),
			)
		}
		checkRemoveEdge(t, g, n2, n3)
		checkRemoveEdge(t, g, n3, n2)
	})

	t.Run("Graph node relatives", func(t *testing.T) {
		g := NewGraph(100)
		// Set 100 nodes with SetNodeByID
		data := mockData{42}
		for i := 0; i < 100; i++ {
			g.SetNodeByID(i, i, i, 0, data)
		}

		n1, _ := g.GetNodeByID(10)
		n2, _ := g.GetNodeByID(20)
		g.SetEdgeByNodeID(n1.ID, n2.ID, 12, "parent", "child", false)
		hasChild := g.HasRelative(n1, "child")
		if !hasChild {
			t.Fatalf(
				fmt.Sprintf(
					"Node %d should have relative with tag %s",
					n1.ID, "child",
				),
			)
		}

		n3, _ := g.GetNodeByID(50)
		g.SetEdgeByNodeID(n1.ID, n3.ID, 12, "biggus", "diccus", true)
		n3FromEdge, err := g.GetRelative(n1, "diccus")
		if err != nil {
			t.Fatalf(
				fmt.Sprintf(
					"Could not get node %d from edge %s on node %d",
					n3.ID, "diccus", n1.ID,
				),
			)
		}
		if !reflect.DeepEqual(n3FromEdge, n3) {
			t.Fatalf(
				fmt.Sprintf(
					"Node %d does not match node %d from edge %s on node %d",
					n3.ID, n3FromEdge.ID, "diccus", n1.ID,
				),
			)
		}

		n1FromEdge, err := g.GetRelativeByID(n3.ID, "biggus")
		if err != nil {
			t.Fatalf(
				fmt.Sprintf(
					"Could not get node %d from edge %s on node %d",
					n3.ID, "biggus", n1.ID,
				),
			)
		}
		if !reflect.DeepEqual(n1FromEdge, n1) {
			t.Fatalf(
				fmt.Sprintf(
					"Node %d does not match node %d from edge %s on node %d",
					n1.ID, n1FromEdge.ID, "biggus", n3.ID,
				),
			)
		}
	})

	t.Run("Random unidirectional graph", func(t *testing.T) {
		//TODO after RandomUnidirectionalGraph() is rewritten
	})
	//func RandomUnidirectionalGraph(n, e, x, y int, w float64) *Graph {
	//func grid2Nodes(idx, x int) (int, int) {
	//func edge2Nodes(idx, n int) (int, int) {
	//func nodes2Edge(n1, n2, n int) int {
	//func min(n1, n2 int) int {
	//func max(n1, n2 int) int {
	//func order(n1, n2 int) (int, int) {

	fmt.Println()
}

type mockData struct {
	D int
}

func (m mockData) GetData() interface{} {
	return m
}

func (m mockData) DeleteData() {
}

func mockDataFromData(d Data) (mockData, bool) {
	m, ok := d.(mockData)
	return m, ok
}

func checkEdge(t *testing.T, g *Graph, n1, n2 *Node, w float64, t1, t2 string) {
	t.Helper()
	e12, err := g.GetEdge(n1, n2.ID)
	if err != nil {
		t.Fatalf(fmt.Sprintf("Could not retrieve edge from %d to %d", n1.ID, n2.ID))
	}
	eps := 0.0001
	if e12.Weight-w > eps {
		t.Fatalf(
			fmt.Sprintf("Edge weight from nodes %d to %d is %f but should be %f",
				n1.ID, n2.ID, e12.Weight, w,
			),
		)
	}
	if e12.Nodes[0].Tag != t1 || e12.Nodes[1].Tag != t2 {
		t.Fatalf(
			fmt.Sprintf("Edge tags for nodes %d to %d should be %s and %s",
				n1.ID, n2.ID, t1, t2),
		)
	}
	if !reflect.DeepEqual(e12.Nodes[0].Node, n1) || !reflect.DeepEqual(e12.Nodes[1].Node, n2) {
		t.Fatalf(
			fmt.Sprintf("Edge nodes for nodes %d to %d should be equivalent to actual nodes",
				n1.ID, n2.ID),
		)
	}
}

func checkRemoveEdge(t *testing.T, g *Graph, n1, n2 *Node) {
	t.Helper()
	_, err := g.GetEdge(n1, n2.ID)
	_, ok := err.(NoEdgeError)
	if !ok {
		t.Fatalf(fmt.Sprintf("Edge from %d to %d not removed from graph properly", n1.ID, n2.ID))
	}
}
