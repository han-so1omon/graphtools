package structures

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
)

const (
	// RBTreeType names RBTree for use in API operations
	RBTreeType = "red-black tree"
	// DataNodeTag denotes the node as a data (non-nil) node
	DataNodeTag = "data"
	// NilNodeTag denotes the node as nil (i.e. leaf or parent of root)
	NilNodeTag = "nil"
)

var (
	// Tags holds the minimized versions of RBTree node relative tags
	Tags map[string]string = map[string]string{
		"root":   "r",
		"parent": "p",
		"lchild": "cl",
		"rchild": "cr",
	}

	// Colors holds the valid values for RBTree node color
	Colors map[string]int8 = map[string]int8{
		"nil":    -1,
		"black":  0,
		"red":    1,
		"orange": 2,
		"yellow": 3,
		"green":  4,
		"blue":   5,
		"purple": 6,
	}
)

// Color implements Data interface
type ColorData struct {
	Color  int8 `json:"color"`
	Height int  `json:"height"`
}

func (c ColorData) GetData() interface{} {
	return c
}

func (c ColorData) DeleteData() {
}

func ColorDataFromData(d Data) (ColorData, bool) {
	c, ok := d.(ColorData)
	return c, ok
}

type ColorDataError struct {
	msg string
	Err error
}

func (e *ColorDataError) Error() string {
	return fmt.Sprintf(e.msg+": %v", e.Err)
}

func (e *ColorDataError) Unwrap() error { return e.Err }

type NilNodeError struct {
	msg string
	Err error
}

func (e *NilNodeError) Error() string {
	return fmt.Sprintf(e.msg+": %v", e.Err)
}

func (e *NilNodeError) Unwrap() error { return e.Err }

type RootInsertError struct {
	Err error
}

func (e *RootInsertError) Error() string {
	return fmt.Sprintf("Cannot reinsert root node into RBTree: %v", e.Err)
}

func (e *RootInsertError) Unwrap() error { return e.Err }

type TagError struct {
	tag string
	Err error
}

func (e *TagError) Error() string {
	return fmt.Sprintf(
		"RBTree node tag is type %s. Must be %s, %s, %s, or %s: %v",
		e.tag,
		Tags["root"],
		Tags["parent"],
		Tags["lchild"],
		Tags["rchild"],
		e.Err,
	)
}

func (e *TagError) Unwrap() error { return e.Err }

type NodeTypeTagError struct {
	nodeTag string
	Err     error
}

func (e *NodeTypeTagError) Error() string {
	return fmt.Sprintf(
		"NodeTypeTag is type %s. Must be %s or %s: %v",
		e.nodeTag,
		DataNodeTag,
		NilNodeTag,
		e.Err,
	)
}

func (e *NodeTypeTagError) Unwrap() error { return e.Err }

type rbIDDistributor struct {
	// dataNodeCount distributes positive ID values to data nodes
	dataNodeCount int
	// nilNodeCount distributes negative ID values to nil nodes
	nilNodeCount int
}

func (r *rbIDDistributor) GetID(nodeTypeTag string) int {
	if nodeTypeTag == DataNodeTag {
		return r.dataNodeCount
	} else {
		return r.nilNodeCount
	}
}

//TODO function to re-assign node heights during insertion and deletion
type RBTree struct {
	Root  *Node  `json:"root"`
	Graph *Graph `json:"graph"`
	Type  string `json:"type"`

	idDistributor     IDDistributor
	Height            int `json:"height"`
	numMaxHeightNodes int

	lock    *sync.Mutex
	updated chan struct{}
	cancel  context.CancelFunc
	ctx     context.Context
}

func (t *RBTree) String() string {
	var b strings.Builder
	fmt.Fprintf(&b, "\n+ + + + +RBTree+ + + + +\n")
	fmt.Fprintf(&b, "Type: %s\n", t.Type)
	fmt.Fprintf(&b, "Root: %d\n", t.Root.ID)
	fmt.Fprintf(&b, "Height: %d\n", t.Height)
	b.WriteString(t.Graph.String())
	fmt.Fprintf(&b, "+ + + + + + + + + + + + +\n")
	return b.String()
}

func NewRBTree(ctx context.Context, cancel context.CancelFunc) *RBTree {
	t := new(RBTree)
	t.lock = &sync.Mutex{}
	t.idDistributor = &rbIDDistributor{0, -1}

	t.updated = make(chan struct{})
	t.cancel = cancel
	t.ctx = ctx
	t.Graph = NewGraph(1.0)
	t.Type = RBTreeType

	t.putNode(nil, Tags["root"], NilNodeTag, Colors["black"])

	return t
}

// Updated will return a channel that receives whenever the graph is decided to
// be updated
func (t *RBTree) Updated() <-chan struct{} {
	return t.updated
}

// OnUpdate is useful to be called when the graph is decided to be updated.
// It is the prerogative of graph owners (i.e. end-users, accompanying
// structures, or algorithms) to call OnUpdate()
func (t *RBTree) OnUpdate() {
	t.updated <- struct{}{}
}

// Done is useful to be called when the graph is decided to be done
// It is the prerogative of graph owners (i.e. end-users, accompanying
// structures, or algorithms) to call Done()
func (t *RBTree) Done() {
	close(t.updated)
	t.cancel()
}

// Lock is useful to be called when the graph needs to be accessed as an atomic
// structure
func (t *RBTree) Lock() {
	t.lock.Lock()
}

// Unlock removees the graph from the atomic locked state
func (t *RBTree) Unlock() {
	t.lock.Unlock()
}

func (t *RBTree) NewNode(nodeTypeTag string) (*Node, error) {
	t.Lock()
	defer t.Unlock()
	var id int
	if nodeTypeTag == DataNodeTag {
		id = t.idDistributor.(*rbIDDistributor).dataNodeCount
		t.idDistributor.(*rbIDDistributor).dataNodeCount++
	} else {
		id = t.idDistributor.(*rbIDDistributor).nilNodeCount
		t.idDistributor.(*rbIDDistributor).nilNodeCount--
	}

	data := ColorData{
		Color:  Colors["red"],
		Height: 0,
	}
	x := id
	y := id
	z := 0
	n, err := t.Graph.SetNodeByID(id, x, y, z, data)
	if err != nil {
		return nil, err
	}

	return n, nil
}

func (t *RBTree) putNode(parent *Node, tag, nodeTypeTag string, color int8) error {
	if color != Colors["red"] && color != Colors["black"] && color != Colors["nil"] {
		return &ColorDataError{fmt.Sprintf("Color %d is not a valid color", color), nil}
	}

	var err error
	// Handle root insert
	if parent == nil && tag == Tags["root"] {
		if t.Root != nil {
			return &RootInsertError{nil}
		}
		id := t.idDistributor.(*rbIDDistributor).dataNodeCount
		t.idDistributor.(*rbIDDistributor).dataNodeCount++
		x := id
		y := id
		z := 0
		data := ColorData{
			Color:  color,
			Height: 0,
		}
		n, err := t.Graph.SetNodeByID(id, x, y, z, data)
		if err != nil {
			return err
		}

		t.Root = n
		t.Height = 0
		t.numMaxHeightNodes = 1

		// Set nil node as parent of root
		id = t.idDistributor.(*rbIDDistributor).nilNodeCount
		t.idDistributor.(*rbIDDistributor).nilNodeCount--
		x = id
		y = id
		z = 0
		data = ColorData{
			Color:  Colors["nil"],
			Height: -1,
		}
		p, err := t.Graph.SetNodeByID(id, x, y, z, data)
		err = t.setRChild(p, n, true)
		if err != nil {
			return &NilNodeError{"Problem setting nil parent of root node", err}
		}

		// Set nil nodes as children of root
		id = t.idDistributor.(*rbIDDistributor).nilNodeCount
		t.idDistributor.(*rbIDDistributor).nilNodeCount--
		x = id
		y = id
		z = 0
		data = ColorData{
			Color:  Colors["nil"],
			Height: 1,
		}
		rc, err := t.Graph.SetNodeByID(id, x, y, z, data)
		err = t.setRChild(n, rc, true)
		if err != nil {
			return &NilNodeError{"Problem setting right child of root node", err}
		}
		id = t.idDistributor.(*rbIDDistributor).nilNodeCount
		t.idDistributor.(*rbIDDistributor).nilNodeCount--
		x = id
		y = id
		z = 0
		data = ColorData{
			Color:  Colors["nil"],
			Height: 1,
		}
		lc, err := t.Graph.SetNodeByID(id, x, y, z, data)
		err = t.setLChild(n, lc, true)
		if err != nil {
			return &NilNodeError{"Problem setting left child of root node", err}
		}

		return nil
	} else if parent == nil {
		return &NilNodeError{"Cannot insert node with nil parent unless node is root", nil}
	}

	// Get parent data to determine height
	parentData, ok := ColorDataFromData(parent.Extra)
	if !ok {
		return &DataError{nil}
	}

	// Set data node or nil node
	if tag != Tags["rchild"] && tag != Tags["lchild"] {
		return &TagError{tag, nil}
	}

	if nodeTypeTag != DataNodeTag && nodeTypeTag != NilNodeTag {
		return &NodeTypeTagError{nodeTypeTag, nil}
	}

	var id, x, y, z, height int
	var data ColorData
	var n *Node
	if nodeTypeTag == DataNodeTag { // Handle data node
		id = t.idDistributor.(*rbIDDistributor).dataNodeCount
		x = id
		y = id
		z = 0
	} else { // Handle nil node
		id = t.idDistributor.(*rbIDDistributor).nilNodeCount
		x = id
		y = id
		z = 0
	}
	height = parentData.Height + 1
	data = ColorData{
		Color:  color,
		Height: height,
	}

	n, err = t.Graph.SetNodeByID(id, x, y, z, data)
	if err != nil {
		return err
	}

	if height > t.Height {
		t.Height = height
		t.numMaxHeightNodes = 1
	} else if height == t.Height {
		t.numMaxHeightNodes++
	}
	if tag == Tags["rchild"] {
		t.setRChild(parent, n, true)
	} else if tag == Tags["lchild"] {
		t.setLChild(parent, n, true)
	}

	// Increment appropriate idDistributor node count
	// Must wait to the end so that node count is only incremented when a node
	// has been successfully created
	if nodeTypeTag == DataNodeTag {
		t.idDistributor.(*rbIDDistributor).dataNodeCount++
	} else {
		t.idDistributor.(*rbIDDistributor).nilNodeCount--
	}
	return nil
}

func (t *RBTree) removeNilNode(n *Node) {
	t.Graph.RemoveNode(n)
}

func (t *RBTree) NodeIsNil(n *Node) (bool, bool) {
	c, ok := ColorDataFromData(n.Extra)
	if !ok {
		return false, ok
	}
	return c.Color == Colors["nil"], ok
}

func (t *RBTree) NodeColor(n *Node) (int8, bool) {
	c, ok := ColorDataFromData(n.Extra)
	if !ok {
		return 0, ok
	}
	return c.Color, ok
}

func (t *RBTree) GetParent(n *Node) (*Node, error) {
	return t.Graph.GetRelative(n, Tags["parent"])
}

func (t *RBTree) GetGrandParent(n *Node) (*Node, error) {
	p, err := t.GetParent(n)
	if err != nil {
		return nil, err
	}

	return t.GetParent(p)
}

func (t *RBTree) GetRChild(n *Node) (*Node, error) {
	return t.Graph.GetRelative(n, Tags["rchild"])
}

func (t *RBTree) GetLChild(n *Node) (*Node, error) {
	return t.Graph.GetRelative(n, Tags["lchild"])
}

func (t *RBTree) GetSibling(n *Node) (*Node, error) {
	// No sibling if n is root
	if n == t.Root {
		return nil, &NilNodeError{"Root node has no siblings", nil}
	}

	p, err := t.GetParent(n)
	if err != nil {
		return nil, err
	}

	// Check node tag
	e, err := t.Graph.GetEdge(n, p.ID)
	if err != nil {
		return nil, err
	}
	if e.Nodes[0].Tag == Tags["rchild"] {
		return t.GetLChild(p)
	} else {
		return t.GetRChild(p)
	}
}

func (t *RBTree) GetUncle(n *Node) (*Node, error) {
	p, err := t.GetParent(n)
	if err != nil {
		return nil, err
	}

	return t.GetSibling(p)
}

func (t *RBTree) setRChild(np, nrc *Node, bidirectional bool) error {
	return t.Graph.SetEdge(np, nrc, 1.0, Tags["parent"], Tags["rchild"], bidirectional)
}

func (t *RBTree) setLChild(np, nrc *Node, bidirectional bool) error {
	return t.Graph.SetEdge(np, nrc, 1.0, Tags["parent"], Tags["lchild"], bidirectional)
}

func (t *RBTree) setColor(n *Node, color int8) error {
	c, ok := ColorDataFromData(n.Extra)
	if !ok {
		return &DataError{nil}
	}
	c.Color = color
	t.Graph.SetNode(n, n.ID, n.Coords.X, n.Coords.Y, n.Coords.Z, c)
	return nil
}

func (t *RBTree) rotateLeft(n *Node) error {
	nnew, err := t.GetRChild(n)
	if err != nil {
		return err
	}
	p, err := t.GetParent(n)
	if err != nil {
		return err
	}
	n2pTag, _, err := t.Graph.GetEdgeTags(n, p.ID)

	isNil, ok := t.NodeIsNil(nnew)
	if !ok {
		return &DataError{nil}
	}

	if isNil {
		return &NilNodeError{fmt.Sprintf("%d has no edge %s", n.ID, Tags["rchild"]), nil}
	}

	nnewLeft, err := t.GetLChild(nnew)
	if err != nil {
		return err
	}

	// Remove existing edges
	err = t.Graph.RemoveEdge(p, n, true)
	if err != nil {
		return err
	}
	err = t.Graph.RemoveEdge(n, nnew, true)
	if err != nil {
		return err
	}
	err = t.Graph.RemoveEdge(nnew, nnewLeft, true)
	if err != nil {
		return err
	}

	// Set new edges
	err = t.setRChild(n, nnewLeft, true)
	if err != nil {
		return err
	}
	err = t.setLChild(nnew, n, true)
	if err != nil {
		return err
	}

	if n2pTag == Tags["lchild"] {
		err = t.setLChild(p, nnew, true)
		if err != nil {
			return err
		}
	} else if n2pTag == Tags["rchild"] {
		err = t.setRChild(p, nnew, true)
		if err != nil {
			return err
		}
	}

	if n.ID == t.Root.ID {
		t.Root = nnew
	}

	return nil
}

func (t *RBTree) rotateRight(n *Node) error {
	nnew, err := t.GetLChild(n)
	if err != nil {
		return err
	}
	p, err := t.GetParent(n)
	if err != nil {
		return err
	}
	n2pTag, _, err := t.Graph.GetEdgeTags(n, p.ID)

	isNil, ok := t.NodeIsNil(nnew)
	if !ok {
		return &DataError{nil}
	}

	if isNil {
		return &NilNodeError{fmt.Sprintf("%d has no edge %s", n.ID, Tags["rchild"]), nil}
	}

	nnewRight, err := t.GetRChild(nnew)
	if err != nil {
		return err
	}

	// Remove existing edges
	err = t.Graph.RemoveEdge(p, n, true)
	if err != nil {
		return err
	}
	err = t.Graph.RemoveEdge(n, nnew, true)
	if err != nil {
		return err
	}
	err = t.Graph.RemoveEdge(nnew, nnewRight, true)
	if err != nil {
		return err
	}

	// Set new edges
	err = t.setLChild(n, nnewRight, true)
	if err != nil {
		return err
	}
	err = t.setRChild(nnew, n, true)
	if err != nil {
		return err
	}

	if n2pTag == Tags["lchild"] {
		err = t.setLChild(p, nnew, true)
		if err != nil {
			return err
		}
	} else if n2pTag == Tags["rchild"] {
		err = t.setRChild(p, nnew, true)
		if err != nil {
			return err
		}
	}

	if n.ID == t.Root.ID {
		t.Root = nnew
	}

	return nil

}

func (t *RBTree) Insert(root *Node, n *Node) error {
	t.lock.Lock()
	defer t.lock.Unlock()
	err := t.insertRecurse(root, n)
	if err != nil {
		return fmt.Errorf("Insert: %w", err)
	}

	err = t.insertRepairTree(n)
	if err != nil {
		return err
	}

	root = n
	rootParent, err := t.GetParent(root)
	if err != nil {
		return fmt.Errorf("Insert: %w", err)
	}
	isNil, ok := t.NodeIsNil(rootParent)
	for ok && !isNil {
		root = rootParent
		rootParent, err = t.GetParent(root)
		if err != nil {
			return fmt.Errorf("Insert: %w", err)
		}
		isNil, ok = t.NodeIsNil(rootParent)
	}
	if !ok {
		return &DataError{nil}
	}

	t.Root = root
	return nil
}

func (t *RBTree) insertRecurse(root *Node, n *Node) error {
	isNil, ok := t.NodeIsNil(root)
	if !ok {
		return &DataError{nil}
	}

	var child *Node
	var err error
	var isLeft bool
	if !isNil {
		cmp := n.Compare(root)
		if cmp < 0 {
			child, err = t.GetLChild(root)
			isLeft = true
		} else {
			child, err = t.GetRChild(root)
			isLeft = false
		}

		if err != nil {
			return err
		}
		isNil, ok = t.NodeIsNil(child)
		if !ok {
			return &DataError{nil}
		}
		if !isNil {
			return t.insertRecurse(child, n)
		} else {
			if isLeft {
				err = t.setLChild(root, n, true)
			} else {
				err = t.setRChild(root, n, true)
			}
			if err != nil {
				return err
			}
			t.removeNilNode(child)
		}
	}

	err = t.setColor(n, Colors["red"])
	if err != nil {
		return err
	}
	err = t.putNode(n, Tags["lchild"], NilNodeTag, Colors["nil"])
	if err != nil {
		return err
	}
	t.putNode(n, Tags["rchild"], NilNodeTag, Colors["nil"])
	if err != nil {
		return err
	}

	return nil
}

func (t *RBTree) insertRepairTree(n *Node) error {
	p, err := t.GetParent(n)
	if err != nil { // node must have parent
		return err
	}
	parentData, ok := ColorDataFromData(p.Extra)
	if !ok {
		return &DataError{nil}
	}
	parentIsNil, ok := t.NodeIsNil(p)
	if !ok {
		return &DataError{nil}
	}
	if parentIsNil {
		return t.insertCase1(n)
	} else if parentData.Color == Colors["black"] {
		return t.insertCase2(n)
	}

	u, err := t.GetUncle(n)

	if err == nil {
		uncleData, ok := ColorDataFromData(u.Extra)
		if !ok {
			return &DataError{nil}
		}
		uncleIsNil, ok := t.NodeIsNil(u)
		if !ok {
			return &DataError{nil}
		}
		if !uncleIsNil && uncleData.Color == Colors["red"] {
			return t.insertCase3(n)
		}
	}

	// Proceed if err above was just NilNodeError because uncle doesn't exist
	// If not, then error was legitimate and return err
	var errCheck *NilNodeError
	errCheckOk := errors.As(err, &errCheck)
	if !errCheckOk {
		return err
	}
	return t.insertCase4(n)
}

func (t *RBTree) insertCase1(n *Node) error {
	err := t.setColor(n, Colors["black"])
	if err != nil {
		return err
	}
	return nil
}

func (t *RBTree) insertCase2(n *Node) error {
	return nil
}

func (t *RBTree) insertCase3(n *Node) error {
	p, err := t.GetParent(n)
	if err != nil {
		return err
	}
	u, err := t.GetUncle(n)
	if err != nil {
		return err
	}

	gp, err := t.GetGrandParent(n)
	if err != nil {
		return err
	}

	err = t.setColor(p, Colors["black"])
	if err != nil {
		return err
	}
	err = t.setColor(u, Colors["black"])
	if err != nil {
		return err
	}
	err = t.setColor(gp, Colors["red"])
	if err != nil {
		return err
	}
	err = t.insertRepairTree(gp)
	if err != nil {
		return err
	}

	return nil
}

func (t *RBTree) insertCase4(n *Node) error {
	p, err := t.GetParent(n)
	if err != nil {
		return err
	}

	gp, err := t.GetGrandParent(n)
	if err != nil {
		return err
	}

	n2pTag, _, err := t.Graph.GetEdgeTags(n, p.ID)
	if err != nil {
		return err
	}
	p2gpTag, _, err := t.Graph.GetEdgeTags(p, gp.ID)
	if err != nil {
		return err
	}

	if n2pTag == Tags["rchild"] && p2gpTag == Tags["lchild"] {
		t.rotateLeft(p)
		n, err = t.GetLChild(n)
		if err != nil {
			return err
		}
	} else if n2pTag == Tags["lchild"] && p2gpTag == Tags["rchild"] {
		err = t.rotateRight(p)
		if err != nil {
			return err
		}
		n, err = t.GetRChild(n)
		if err != nil {
			return err
		}
	}

	return t.insertCase4Step2(n)
}

func (t *RBTree) insertCase4Step2(n *Node) error {
	p, err := t.GetParent(n)
	if err != nil {
		return err
	}

	gp, err := t.GetGrandParent(n)
	if err != nil {
		return err
	}

	n2pTag, _, err := t.Graph.GetEdgeTags(n, p.ID)
	if err != nil {
		return err
	}
	_, _, err = t.Graph.GetEdgeTags(p, gp.ID)
	if err != nil {
		return err
	}

	if n2pTag == Tags["lchild"] {
		err = t.rotateRight(gp)
		if err != nil {
			return err
		}
	} else {
		err = t.rotateLeft(gp)
		if err != nil {
			return err
		}
	}

	err = t.setColor(p, Colors["black"])
	if err != nil {
		return err
	}
	err = t.setColor(gp, Colors["red"])
	if err != nil {
		return err
	}

	return nil
}

func (t *RBTree) replaceNode(n, child *Node) error {
	p, err := t.GetParent(n)
	if err != nil {
		return err
	}

	n2pTag, _, err := t.Graph.GetEdgeTags(n, p.ID)
	if err != nil {
		return err
	}
	if n2pTag == Tags["lchild"] {
		err = t.setLChild(p, child, true)
	} else {
		err = t.setRChild(p, child, true)
	}
	if err != nil {
		return err
	}

	err = t.Graph.RemoveEdge(p, n, true)
	if err != nil {
		return err
	}
	err = t.Graph.RemoveEdge(n, child, true)
	if err != nil {
		return err
	}

	if n.ID == t.Root.ID {
		t.Root = child
	}

	return nil
}

func (t *RBTree) DeleteOneChild(n *Node) error {
	t.Lock()
	defer t.Unlock()
	// Precondition: n has at most one non-leaf child
	var child *Node
	var err error

	rc, err := t.GetRChild(n)
	if err != nil {
		return fmt.Errorf("DeleteOneChild: %w", err)
	}
	rcIsNil, ok := t.NodeIsNil(rc)
	if !ok {
		return &DataError{nil}
	}
	lc, err := t.GetLChild(n)
	lcIsNil, ok := t.NodeIsNil(lc)
	if !ok {
		return &DataError{nil}
	}
	if err != nil {
		return fmt.Errorf("DeleteOneChild: %w", err)
	}

	if !rcIsNil && !lcIsNil {
		return &NilNodeError{fmt.Sprintf("Node %d must have at most one non-leaf child", n.ID), nil}
	} else if !rcIsNil {
		child = rc
	} else if !lcIsNil {
		child = lc
	} else {
		return &NilNodeError{fmt.Sprintf("Node %d must have at least one child", n.ID), nil}
	}

	err = t.replaceNode(n, child)
	if err != nil {
		return fmt.Errorf("DeleteOneChild: %w", err)
	}

	nodeData, ok := ColorDataFromData(n.Extra)
	if !ok {
		return &DataError{nil}
	}
	childData, ok := ColorDataFromData(child.Extra)
	if !ok {
		return &DataError{nil}
	}

	if nodeData.Color == Colors["black"] {
		if childData.Color == Colors["red"] {
			err = t.setColor(child, Colors["black"])
			if err != nil {
				return fmt.Errorf("DeleteOneChild: %w", err)
			}
		} else {
			t.deleteCase1(child)
		}
	}

	t.Graph.RemoveNode(n)

	return nil
}

func (t *RBTree) deleteCase1(n *Node) error {
	_, err := t.GetParent(n)
	// NoEdgeError is acceptable, as it means that the parent does not exist
	var errCheck *NoEdgeError
	if errors.As(err, &errCheck) {
		return nil
	} else if err != nil {
		// Make sure there is not an unexpected error
		return err
	}

	t.deleteCase2(n)
	return nil
}

func (t *RBTree) deleteCase2(n *Node) error {
	s, err := t.GetSibling(n)
	if err != nil {
		return err
	}

	p, err := t.GetParent(n)
	if err != nil {
		return err
	}

	sData, ok := ColorDataFromData(s.Extra)
	if !ok {
		return &DataError{nil}
	}
	if sData.Color == Colors["red"] {
		err = t.setColor(p, Colors["red"])
		if err != nil {
			return err
		}
		err = t.setColor(s, Colors["black"])
		if err != nil {
			return err
		}

		n2pTag, _, err := t.Graph.GetEdgeTags(n, p.ID)
		if err != nil {
			return err
		}
		if n2pTag == Tags["lchild"] {
			err = t.rotateLeft(p)
		} else {
			err = t.rotateRight(p)
		}
		if err != nil {
			return err
		}
	}

	t.deleteCase3(n)
	return nil
}

func (t *RBTree) deleteCase3(n *Node) error {
	s, err := t.GetSibling(n)
	if err != nil {
		return err
	}
	sData, ok := ColorDataFromData(s.Extra)
	if !ok {
		return &DataError{nil}
	}

	p, err := t.GetParent(n)
	if err != nil {
		return err
	}
	pData, ok := ColorDataFromData(p.Extra)
	if !ok {
		return &DataError{nil}
	}

	rc, err := t.GetRChild(n)
	if err != nil {
		return err
	}
	rcData, ok := ColorDataFromData(rc.Extra)
	if !ok {
		return &DataError{nil}
	}

	lc, err := t.GetLChild(n)
	if err != nil {
		return err
	}
	lcData, ok := ColorDataFromData(lc.Extra)
	if !ok {
		return &DataError{nil}
	}

	if (pData.Color == Colors["black"]) &&
		(sData.Color == Colors["black"]) &&
		(lcData.Color == Colors["black"]) &&
		(rcData.Color == Colors["black"]) {
		err = t.setColor(s, Colors["red"])
		if err != nil {
			return err
		}
		return t.deleteCase1(p)
	}

	return t.deleteCase4(n)
}

func (t *RBTree) deleteCase4(n *Node) error {
	s, err := t.GetSibling(n)
	if err != nil {
		return err
	}
	sData, ok := ColorDataFromData(s.Extra)
	if !ok {
		return &DataError{nil}
	}

	p, err := t.GetParent(n)
	if err != nil {
		return err
	}
	pData, ok := ColorDataFromData(p.Extra)
	if !ok {
		return &DataError{nil}
	}

	rc, err := t.GetRChild(n)
	if err != nil {
		return err
	}
	rcData, ok := ColorDataFromData(rc.Extra)
	if !ok {
		return &DataError{nil}
	}

	lc, err := t.GetLChild(n)
	if err != nil {
		return err
	}
	lcData, ok := ColorDataFromData(lc.Extra)
	if !ok {
		return &DataError{nil}
	}

	if (pData.Color == Colors["red"]) &&
		(sData.Color == Colors["black"]) &&
		(lcData.Color == Colors["black"]) &&
		(rcData.Color == Colors["black"]) {
		err = t.setColor(s, Colors["red"])
		if err != nil {
			return err
		}
		err = t.setColor(p, Colors["black"])
		if err != nil {
			return err
		}
	} else {
		return t.deleteCase5(n)
	}

	return nil
}

func (t *RBTree) deleteCase5(n *Node) error {
	s, err := t.GetSibling(n)
	if err != nil {
		return err
	}
	sData, ok := ColorDataFromData(s.Extra)
	if !ok {
		return &DataError{nil}
	}

	p, err := t.GetParent(n)
	if err != nil {
		return err
	}
	n2pTag, _, err := t.Graph.GetEdgeTags(n, p.ID)
	if err != nil {
		return err
	}

	sRc, err := t.GetRChild(s)
	if err != nil {
		return err
	}
	sRcData, ok := ColorDataFromData(sRc.Extra)
	if !ok {
		return &DataError{nil}
	}

	sLc, err := t.GetLChild(s)
	if err != nil {
		return err
	}
	sLcData, ok := ColorDataFromData(sLc.Extra)
	if !ok {
		return &DataError{nil}
	}

	if sData.Color == Colors["black"] {
		if (n2pTag == Tags["lchild"]) &&
			(sRcData.Color == Colors["black"]) &&
			(sLcData.Color == Colors["red"]) {
			err = t.setColor(s, Colors["red"])
			if err != nil {
				return err
			}
			err = t.setColor(sLc, Colors["black"])
			if err != nil {
				return err
			}
			t.rotateRight(s)
		} else if (n2pTag == Tags["rchild"]) &&
			(sRcData.Color == Colors["red"]) &&
			(sLcData.Color == Colors["black"]) {
			err = t.setColor(s, Colors["red"])
			if err != nil {
				return err
			}
			err = t.setColor(sRc, Colors["black"])
			if err != nil {
				return err
			}
			t.rotateLeft(s)
		}
	}
	return t.deleteCase6(n)
}

func (t *RBTree) deleteCase6(n *Node) error {
	s, err := t.GetSibling(n)
	if err != nil {
		return err
	}

	p, err := t.GetParent(n)
	if err != nil {
		return err
	}
	pData, ok := ColorDataFromData(p.Extra)
	if !ok {
		return &DataError{nil}
	}
	n2pTag, _, err := t.Graph.GetEdgeTags(n, p.ID)
	if err != nil {
		return err
	}

	// Set sibling color to parent color
	err = t.setColor(s, pData.Color)
	if err != nil {
		return err
	}
	// Set parent color to black
	err = t.setColor(p, Colors["black"])
	if err != nil {
		return err
	}

	if n2pTag == Tags["lchild"] {
		sRc, err := t.GetRChild(s)
		if err != nil {
			return err
		}
		_, ok = ColorDataFromData(sRc.Extra)
		if !ok {
			return &DataError{nil}
		}

		err = t.setColor(sRc, Colors["black"])
		if err != nil {
			return err
		}
		err = t.rotateLeft(p)
	} else {
		sLc, err := t.GetLChild(s)
		if err != nil {
			return err
		}
		_, ok = ColorDataFromData(sLc.Extra)
		if !ok {
			return &DataError{nil}
		}

		err = t.setColor(sLc, Colors["black"])
		if err != nil {
			return err
		}
		err = t.rotateRight(p)
	}
	if err != nil {
		return err
	}

	return nil
}
