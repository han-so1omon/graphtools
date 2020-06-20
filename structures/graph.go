package structures

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"
)

// NoNodeError states that the requested node does not exist
type NoNodeError struct {
	v   int
	Err error
}

// Error serves the error message for NoNodeError
func (e *NoNodeError) Error() string {
	return fmt.Sprintf("No graph node with value %d: %v", int(e.v), e.Err)
}

func (e *NoNodeError) Unwrap() error { return e.Err }

// NoEdgeError states that the requested edge does not exist
// n1 and n2 are the IDs of the near and far nodes
type NoEdgeError struct {
	msg string
	Err error
}

// Error serves the error message for NoEdgeError
func (e *NoEdgeError) Error() string {
	return fmt.Sprintf(e.msg+": %v", e.Err)
}

func (e *NoEdgeError) Unwrap() error { return e.Err }

// EdgeWeightError states that edge value is not valid
type EdgeWeightError struct {
	w   float64
	Err error
}

// Error serves the error message for EdgeWeightError
func (e *EdgeWeightError) Error() string {
	return fmt.Sprintf("Invalid edge value %f: %v", e.w, e.Err)
}

func (e *EdgeWeightError) Unwrap() error { return e.Err }

// Comparable interface defines an orderable type
// Ordering is used for efficient storage and sorting
type Comparable interface {
	// Compare returns >0 if element is greater than requested element
	//				   <0 if less than requested element
	//				   0 if elements are equal
	Compare(Comparable) int8
	// GetValue returns the integer equivalent value of a Comparable
	GetValue() int
}

const (
	// GenericGraphManagerType names the generic graph manager for use in API
	// operations
	GenericGraphManagerType = "generic"
)

type GraphDisplayManager interface {
	Updated() <-chan struct{}
	OnUpdate()
	Done()
	Lock()
	Unlock()
}

type GenericGraphManager struct {
	Graph *Graph `json:"graph"`
	Type  string `json:"type"`

	// Control structures
	updated chan struct{}
	cancel  context.CancelFunc
	ctx     context.Context
	lock    *sync.Mutex
	isDone  bool
}

func LoadCSV(ctx context.Context, cancel context.CancelFunc, csvText string) (*GenericGraphManager, error) {
	//mgr := NewGenericGraphManager(ctx, cancel, maxEdgeWeight)
	var (
		mgr           *GenericGraphManager
		numNodes      int
		numEdges      int
		maxEdgeWeight float64
	)
	reader := csv.NewReader(strings.NewReader(csvText))
	reader.FieldsPerRecord = -1
	lineNum := 0
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		if lineNum == 0 {
			numNodes, err = strconv.Atoi(record[0])
			if err != nil {
				return nil, err
			}
			numEdges, err = strconv.Atoi(record[1])
			if err != nil {
				return nil, err
			}
			maxEdgeWeight, err = strconv.ParseFloat(record[2], 64)
			if err != nil {
				return nil, err
			}
			mgr = NewGenericGraphManager(ctx, cancel, maxEdgeWeight)
		} else if lineNum <= numNodes {
			id, err := strconv.Atoi(record[0])
			if err != nil {
				return nil, err
			}
			x, err := strconv.ParseFloat(record[1], 64)
			if err != nil {
				return nil, err
			}
			y, err := strconv.ParseFloat(record[2], 64)
			if err != nil {
				return nil, err
			}
			z, err := strconv.ParseFloat(record[3], 64)
			if err != nil {
				return nil, err
			}

			data := ColorData{
				Color:  Colors["orange"],
				Type:   DataNodeTag,
				Height: 0,
			}
			mgr.Graph.SetNodeByID(id, x, y, z, data)
		} else if lineNum <= numNodes+numEdges {
			n1, err := strconv.Atoi(record[0])
			if err != nil {
				return nil, err
			}
			n2, err := strconv.Atoi(record[1])
			if err != nil {
				return nil, err
			}
			w, err := strconv.ParseFloat(record[2], 64)
			if err != nil {
				return nil, err
			}
			mgr.Graph.SetEdgeByNodeID(n1, n2, w, "n", "n", false)
			//func (g *Graph) SetEdgeByNodeID(n1, n2 int, w float64, t1, t2 string, bidirectional bool) error {
		} else {
			break // done with CSV per specification
		}

		lineNum++
	}

	return mgr, nil
}

func NewGenericGraphManager(
	ctx context.Context,
	cancel context.CancelFunc,
	maxEdgeWeight float64,
) *GenericGraphManager {
	mgr := new(GenericGraphManager)
	mgr.Graph = NewGraph(maxEdgeWeight)
	mgr.Type = GenericGraphManagerType
	mgr.lock = &sync.Mutex{}
	mgr.updated = make(chan struct{})
	mgr.cancel = cancel
	mgr.ctx = ctx

	return mgr
}

// Updated will return a channel that receives whenever the graph is decided to
// be updated
func (g *GenericGraphManager) Updated() <-chan struct{} {
	return g.updated
}

// OnUpdate is useful to be called when the graph is decided to be updated.
// It is the prerogative of graph owners (i.e. end-users, accompanying
// structures, or algorithms) to call OnUpdate()
func (g *GenericGraphManager) OnUpdate() {
	if !g.isDone {
		g.updated <- struct{}{}
	}
}

// Done is useful to be called when the graph is decided to be done
// It is the prerogative of graph owners (i.e. end-users, accompanying
// structures, or algorithms) to call Done()
func (g *GenericGraphManager) Done() {
	//g.cancel()
	close(g.updated)
	g.isDone = true
}

// Lock is useful to be called when the graph needs to be accessed as an atomic
// structure
func (g *GenericGraphManager) Lock() {
	g.lock.Lock()
}

// Unlock removees the graph from the atomic locked state
func (g *GenericGraphManager) Unlock() {
	g.lock.Unlock()
}

// Graph is the generalized node-edge data structure to power the algorithms and
// structures built on the graphtools library
type Graph struct {
	NumNodes      int     `json:"numNodes"`
	NumEdges      int     `json:"numEdges"`
	MaxEdgeWeight float64 `json:"maxEdgeWeight"`
	Nodes         []*Node `json:"nodes"`

	// Control structures
	Lock *sync.Mutex `json:"-"`
}

// String does a pretty print of the current graph
func (g *Graph) String() string {
	var b strings.Builder
	fmt.Fprintf(&b, "-.-.-.-GRAPH-.-.-.-.-\n")
	for _, n := range g.Nodes {
		b.WriteString(n.String())
	}
	fmt.Fprintf(&b, "-.-.-.-.-.-.-.-.-.-.-\n")
	return b.String()
}

// NewGraph creates a new graph structure with a maximum edge weight value
func NewGraph(maxEdgeWeight float64) *Graph {
	rand.Seed(time.Now().UTC().UnixNano())
	g := new(Graph)

	g.Lock = &sync.Mutex{}

	g.MaxEdgeWeight = maxEdgeWeight

	return g
}

// IsEmpty returns whether or not the graph is empty
func (g *Graph) IsEmpty() bool {
	return g.NumNodes == 0
}

// HasNodeWithID returns whether the graph has a node with the specified ID
func (g *Graph) HasNodeWithID(id int) bool {
	for _, n := range g.Nodes {
		if n.ID == id {
			return true
		}
	}
	return false
}

// GetNodeByID returns the node with the specified ID
func (g *Graph) GetNodeByID(id int) (*Node, error) {
	for _, n := range g.Nodes {
		if n.ID == id {
			return n, nil
		}
	}
	return nil, &NoNodeError{id, nil}
}

// setNodeHelper is a non-blocking version of SetNode so that it can be called
// internally without blocking issues
func (g *Graph) setNodeHelper(n *Node, id int, x, y, z float64, extra Data) {
	n.ID = id
	n.Coords = Point{
		X: x,
		Y: y,
		Z: z,
	}
	n.Extra = extra

	// Case adding new node
	if !g.HasNodeWithID(id) {
		g.Nodes = append(g.Nodes, n)
		g.NumNodes++
	}
}

// SetNode sets node n with ID == id or adds a node with this ID if one does not
// exist in the graph
func (g *Graph) SetNode(n *Node, id int, x, y, z float64, extra Data) {
	g.Lock.Lock()
	defer g.Lock.Unlock()

	g.setNodeHelper(n, id, x, y, z, extra)
}

// SetNodeByID sets node with ID == id or adds a node with this ID if one does
// not exist in the graph
func (g *Graph) SetNodeByID(id int, x, y, z float64, extra Data) (*Node, error) {
	n, err := g.GetNodeByID(id)
	var errCheck *NoNodeError
	if errors.As(err, &errCheck) {
		n = NewNode()
	} else if err != nil {
		return nil, fmt.Errorf("setting node by id: %w", err)
	}

	g.SetNode(n, id, x, y, z, extra)

	return n, nil
}

// RemoveNode removes a node from a graph and deletes the metadata from that node.
// It is guaranteed to complete even in the event of errors
func (g *Graph) RemoveNode(n1 *Node) {
	g.Lock.Lock()
	defer g.Lock.Unlock()

	// Delete all edges to and from this node
	for _, n2 := range g.Nodes {
		g.removeEdgeHelper(n2, n1, true)
	}
	for i, n := range g.Nodes {
		if n.ID == n1.ID {
			if i == len(g.Nodes)-1 {
				g.Nodes = g.Nodes[:i]
			} else {
				g.Nodes = append(g.Nodes[:i], g.Nodes[i+1:]...)
			}
			g.NumNodes--
		}
	}

	// Delete node data
	n1.Extra.DeleteData()
}

// RemoveNodeByID removes a node from a graph and deletes the metadata from that
// node. If the node exists in the graph, it is guaranteed to complete even in the
// event of errors
func (g *Graph) RemoveNodeByID(n1 int) (*Node, error) {
	n, err := g.GetNodeByID(n1)
	if err != nil {
		return nil, fmt.Errorf("remove error by id: %w", err)
	}

	g.RemoveNode(n)
	return n, nil
}

// HasRelative returns whether the graph node has a relative with the specified
// tag
func (g *Graph) HasRelative(n *Node, tag string) bool {
	for _, e := range n.Edges {
		if e.Nodes[1].Tag == tag {
			return true
		}
	}
	return false
}

// GetRelative returns the relative node with the specified tag
func (g *Graph) GetRelative(n *Node, tag string) (*Node, error) {
	for _, e := range n.Edges {
		if e.Nodes[1].Tag == tag {
			return e.Nodes[1].Node, nil
		}
	}
	return nil, &NoEdgeError{fmt.Sprintf("No relative from %d with tag %s", n.ID, tag), nil}
}

// GetRelativeByID returns the relative node with the specified tag
func (g *Graph) GetRelativeByID(n1 int, tag string) (*Node, error) {
	n, err := g.GetNodeByID(n1)
	if err != nil {
		return nil, err
	}

	return g.GetRelative(n, tag)
}

// GetEdge returns the edge from n1 to n2
func (g *Graph) GetEdge(n1 *Node, n2 int) (*Edge, error) {
	for _, e := range n1.Edges {
		if e.Nodes[1].ID == n2 {
			return e, nil
		}
	}

	return nil, &NoEdgeError{fmt.Sprintf("No edge from %d to %d", n1.ID, n2), nil}
}

// GetEdgeByNodeID returns the edge from n1 to n2
func (g *Graph) GetEdgeByNodeID(n1, n2 int) (*Edge, error) {
	n, err := g.GetNodeByID(n1)
	if err != nil {
		return nil, err
	}

	return g.GetEdge(n, n2)
}

// GetEdgeTags returns the tags on the edge from n1 to n2
func (g *Graph) GetEdgeTags(n1 *Node, n2 int) (string, string, error) {
	e, err := g.GetEdge(n1, n2)
	if err != nil {
		return "", "", err
	}

	return e.Nodes[0].Tag, e.Nodes[1].Tag, nil
}

// GetEdgeTagsByNodeID returns the tags on the edge from n1 to n2
func (g *Graph) GetEdgeTagsByNodeID(n1, n2 int) (string, string, error) {
	n, err := g.GetNodeByID(n1)
	if err != nil {
		return "", "", err
	}

	return g.GetEdgeTags(n, n2)
}

// setEdgeHelper2 sets edge values in one direction
func (g *Graph) setEdgeHelper2(n1, n2 *Node, w float64, t1, t2 string) error {
	if w > g.MaxEdgeWeight {
		return &EdgeWeightError{w, nil}
	}

	newEdge := false
	var e *Edge
	e, err := g.GetEdge(n1, n2.ID)
	var errCheck *NoEdgeError
	if errors.As(err, &errCheck) {
		e = NewEdge()
		e.AddNodes(n1, n2, t1, t2)
		newEdge = true
	} else if err != nil {
		return err
	}

	e.Weight = w

	if newEdge {
		n1.Edges = append(n1.Edges, e)
		g.NumEdges++
	}

	return nil
}

// setEdgeHelper is a non-blocking version of SetEdge so that it can be called
// internally without blocking issues
func (g *Graph) setEdgeHelper(n1, n2 *Node, w float64, t1, t2 string, bidirectional bool) error {
	err := g.setEdgeHelper2(n1, n2, w, t1, t2)
	if err != nil {
		return err
	}

	if bidirectional {
		err = g.setEdgeHelper2(n2, n1, w, t2, t1)
		if err != nil {
			return err
		}
	}

	return nil
}

// SetEdge creates a new edge from n1 to n2 with weight w, tags t1 and t2. If
// bidirectional is true, then the reverse edge will also be created with the
// same weight and the tags reversed
func (g *Graph) SetEdge(n1, n2 *Node, w float64, t1, t2 string, bidirectional bool) error {
	g.Lock.Lock()
	defer g.Lock.Unlock()

	return g.setEdgeHelper(n1, n2, w, t1, t2, bidirectional)
}

// SetEdgeByNodeID creates a new edge from n1 to n2 with weight w, tags t1 and t2. If
// bidirectional is true, then the reverse edge will also be created with the
// same weight and the tags reversed
func (g *Graph) SetEdgeByNodeID(n1, n2 int, w float64, t1, t2 string, bidirectional bool) error {
	node1, err := g.GetNodeByID(n1)
	if err != nil {
		return err
	}
	node2, err := g.GetNodeByID(n2)
	if err != nil {
		return err
	}

	return g.SetEdge(node1, node2, w, t1, t2, bidirectional)
}

// removeEdgeHelper2 is a non-locking, unidirectional version of remove edge
func (g *Graph) removeEdgeHelper2(n1, n2 *Node) error {
	_, err := g.GetEdge(n1, n2.ID)
	var errCheck *NoEdgeError
	errCheckOk := errors.As(err, &errCheck)
	if err != nil && !errCheckOk {
		return err
	} else if errCheckOk {
		return nil
	}

	for i, e := range n1.Edges {
		if e.Nodes[1].ID == n2.ID {
			if i == len(n1.Edges)-1 {
				n1.Edges = n1.Edges[:i]
			} else {
				n1.Edges = append(n1.Edges[:i], n1.Edges[i+1:]...)
			}
			break
		}
	}

	return nil
}

// removeEdgeHelper is a non-locking version of remove edge, so that it can be
// called internally without blocking issues
// If bidirectional and both nodes exist, must process both cases even on error of first case
func (g *Graph) removeEdgeHelper(n1, n2 *Node, bidirectional bool) error {
	err := g.removeEdgeHelper2(n1, n2)
	if err != nil {
		return err
	}

	if bidirectional {
		return g.removeEdgeHelper2(n2, n1)
	}
	return nil
}

// RemoveEdge removes the edge between n1 and n2
func (g *Graph) RemoveEdge(n1, n2 *Node, bidirectional bool) error {
	g.Lock.Lock()
	defer g.Lock.Unlock()

	return g.removeEdgeHelper(n1, n2, bidirectional)
}

// RemoveEdgeByNodeID removes the edge between n1 and n2
func (g *Graph) RemoveEdgeByNodeID(n1, n2 int, bidirectional bool) error {
	node1, err := g.GetNodeByID(n1)
	if err != nil {
		return err
	}
	node2, err := g.GetNodeByID(n2)
	if err != nil {
		return err
	}

	return g.RemoveEdge(node1, node2, bidirectional)
}

// RandomUnidirectionalGraph creates a bidirectional graph
// with n nodes, e edges, and m max value of an edge
// with cartesian boundaries x and y
func RandomUnidirectionalGraph(
	ctx context.Context, cancel context.CancelFunc, n, e, x, y int, w float64) *GenericGraphManager {
	//TODO rewrite this with current graph-building tools
	g := NewGenericGraphManager(ctx, cancel, w)
	/*
		g.NumNodes = n
		g.NumEdges = e
		g.MaxEdgeWeight = w

		gridSize := x * y
		openGridSet := make([]int, gridSize)
		for i := range openGridSet {
			openGridSet[i] = i
		}
		// Create graph nodes
		for i := 0; i < n; i++ {
			g.Nodes = append(g.Nodes, NewNode())
			g.Nodes[i].ID = i

			gridNum := rand.Intn(len(openGridSet))
			gridIdx := openGridSet[gridNum]
			xVal, yVal := grid2Nodes(gridIdx, x)
			g.Nodes[i].Coords = Point{X: xVal, Y: yVal, Z: 0}

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
			edge.Weight = rand.Float64() * w
			// Get node values from edge index
			n1, n2 := edge2Nodes(edgeIdx, n)
			// Add edge to starting node
			g.Nodes[n1].AddEdge(edge)
			// Add nodes to edge
			edge.AddNodes(g.Nodes[n1], g.Nodes[n2], "sibling", "sibling")

			// Remove edge from open set
			if edgeNum == len(openEdgeSet) {
				openEdgeSet = openEdgeSet[:edgeNum]
			} else {
				openEdgeSet = append(openEdgeSet[:edgeNum], openEdgeSet[edgeNum+1:]...)
			}
		}
	*/

	return g
}

// ----- Helpers for random graph generation -----

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
