package structures

import (
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"
)

type Comparable interface {
	Compare(Comparable) int8
	GetValue() int
}

//FIXME updated type of Nodes to map[string]*Node from []*Node
type Graph struct {
	NumNodes     int     `json:"numNodes"`
	NumEdges     int     `json:"numEdges"`
	MaxEdgeValue int     `json:"maxEdgeValue"`
	Nodes        []*Node `json:"nodes"`

	// Control structures
	Lock    *sync.Mutex   `json:"-"`
	Updated chan struct{} `json:"-"`
	Done    chan struct{} `json:"-"`
}

func (g *Graph) String() string {
	var b strings.Builder
	fmt.Fprintf(&b, "-.-.-.-GRAPH-.-.-.-.-\n")
	for _, n := range g.Nodes {
		b.WriteString(n.String())
	}
	fmt.Fprintf(&b, "-.-.-.-.-.-.-.-.-.-.-\n")
	return b.String()
}

func NewGraph() *Graph {
	rand.Seed(time.Now().UTC().UnixNano())
	g := new(Graph)
	return g
}

// RandomBidirectionalGraph creates a bidirectional graph
// with n nodes, e edges, and v max value of an edge
// with cartesian boundaries x and y
func RandomBidirectionalGraph(n, e, v, x, y int) *Graph {
	g := NewGraph()
	g.NumNodes = n
	g.NumEdges = e
	g.MaxEdgeValue = v

	// Create graph control structures
	g.Lock = &sync.Mutex{}
	g.Updated = make(chan struct{})
	g.Done = make(chan struct{})

	gridSize := x * y
	openGridSet := make([]int, gridSize)
	for i := range openGridSet {
		openGridSet[i] = i
	}
	// Create graph nodes
	for i := 0; i < n; i++ {
		g.Nodes = append(g.Nodes, NewNode())
		g.Nodes[i].Value = i

		gridNum := rand.Intn(len(openGridSet))
		gridIdx := openGridSet[gridNum]
		xVal, yVal := grid2Nodes(gridIdx, x)
		g.Nodes[i].Coords = Point{xVal, yVal}

		if gridNum == len(openGridSet) {
			openGridSet = openGridSet[:gridNum]
		} else {
			openGridSet = append(openGridSet[:gridNum], openGridSet[gridNum+1:]...)
		}
	}

	// Create edges
	// Create set of all available edges
	n2 := n * n
	openEdgeSet := make([]int, n2)
	for i := range openEdgeSet {
		openEdgeSet[i] = i
	}

	// Pick edge from open set of available edges and remove from open set
	// Add edge to nodes and add nodes to edge
	for i := 0; i < e; i++ {
		// Pick edge number from open set
		edgeNum := rand.Intn(len(openEdgeSet))
		// Translate edge number to edge index
		edgeIdx := openEdgeSet[edgeNum]
		// Create new edge
		edge := NewEdge()
		// Assign edge random value
		edge.Value = rand.Intn(v)
		// Get node values from edge index
		n1, n2 := edge2Nodes(edgeIdx, n)
		// Add edge to starting node
		g.Nodes[n1].AddEdge(edge)
		// Add nodes to edge
		edge.AddNodes(g.Nodes[n1], g.Nodes[n2])

		// Remove edge from open set
		if edgeNum == len(openEdgeSet) {
			openEdgeSet = openEdgeSet[:edgeNum]
		} else {
			openEdgeSet = append(openEdgeSet[:edgeNum], openEdgeSet[edgeNum+1:]...)
		}
	}

	return g
}

func grid2Nodes(idx, x int) (int, int) {
	return idx % x, idx / x
}

func edge2Nodes(idx, n int) (int, int) {
	return idx / n, idx % n
}

func nodes2Edge(n1, n2, n int) int {
	return n2*n + n2
}

func min(n1, n2 int) int {
	if n1 <= n2 {
		return n1
	}
	return n2
}

func max(n1, n2 int) int {
	if n1 == min(n1, n2) {
		return n2
	}
	return n1
}

func order(n1, n2 int) (int, int) {
	return min(n1, n2), max(n1, n2)
}
