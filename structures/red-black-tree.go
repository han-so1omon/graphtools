package structures

import (
	"fmt"
	"sync"
)

const (
	DataNodeTag = "data"
	NilNodeTag  = "nil"
)

var (
	Tags map[string]string = map[string]string{
		"root":   "r",
		"parent": "p",
		"lchild": "cl",
		"rchild": "cr",
	}

	Colors map[string]int8 = map[string]int8{
		"red":   1,
		"black": -1,
		"nil":   0,
	}
)

// Color implements Data interface
type ColorData struct {
	Color  int8
	Height int
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
}

func (e ColorDataError) Error() string {
	return "RBTree: " + e.msg
}

type NilNodeError struct {
	msg string
}

func (e NilNodeError) Error() string {
	return "RBTree: " + e.msg
}

type RootInsertError struct{}

func (e RootInsertError) Error() string {
	return "RBTree: Cannot reinsert root node"
}

type TagError struct {
	tag string
}

func (e TagError) Error() string {
	return fmt.Sprintf(
		"RBTree: Tag is type %s. Must be %s, %s, %s, or %s",
		e.tag,
		Tags["root"],
		Tags["parent"],
		Tags["lchild"],
		Tags["rchild"],
	)
}

type NodeTypeTagError struct {
	nodeTag string
}

func (e NodeTypeTagError) Error() string {
	return fmt.Sprintf(
		"RBTree: NodeTypeTag is type %s. Must be %s or %s",
		e.nodeTag,
		DataNodeTag,
		NilNodeTag,
	)
}

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
	Root  *Node
	Graph *Graph

	idDistributor     IDDistributor
	Height            int `json:"height"`
	numMaxHeightNodes int

	Lock *sync.Mutex `json:"-"`
}

func NewRBTree() *RBTree {
	t := new(RBTree)
	t.Lock = &sync.Mutex{}
	t.idDistributor = &rbIDDistributor{0, -1}

	t.Graph = NewGraph(1.0)

	t.newNode(nil, Tags["root"], NilNodeTag, Colors["black"])

	return t
}

func (t *RBTree) newNode(parent *Node, tag, nodeTypeTag string, color int8) error {
	if color != Colors["red"] && color != Colors["black"] && color != Colors["nil"] {
		return ColorDataError{fmt.Sprintf("Color %d is not a valid color", color)}
	}

	var err error
	// Handle root insert
	if parent == nil && tag == Tags["root"] {
		if t.Root != nil {
			return RootInsertError{}
		}
		id := t.idDistributor.(*rbIDDistributor).dataNodeCount
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
			return NilNodeError{"Problem setting nil parent of root node. " + err.Error()}
		}

		t.idDistributor.(*rbIDDistributor).dataNodeCount++
		t.idDistributor.(*rbIDDistributor).nilNodeCount--
		return nil
	} else if parent == nil {
		return NilNodeError{"Cannot insert node with nil parent unless node is root"}
	}

	// Get parent data to determine height
	parentData, ok := ColorDataFromData(parent.Extra)
	if !ok {
		return DataError{}
	}

	// Set data node or nil node
	if tag != Tags["rchild"] && tag != Tags["lchild"] {
		return TagError{tag}
	}

	if nodeTypeTag != DataNodeTag && nodeTypeTag != NilNodeTag {
		return NodeTypeTagError{nodeTypeTag}
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
		return nil, NilNodeError{"Root node has no siblings"}
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
		return DataError{}
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
		return DataError{}
	}

	if isNil {
		return NilNodeError{fmt.Sprintf("%d has no edge %s", n.ID, Tags["rchild"])}
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
		return DataError{}
	}

	if isNil {
		return NilNodeError{fmt.Sprintf("%d has no edge %s", n.ID, Tags["rchild"])}
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
	t.Lock.Lock()
	defer t.Lock.Unlock()
	err := t.insertRecurse(root, n)
	if err != nil {
		return err
	}

	err = t.insertRepairTree(n)
	if err != nil {
		return err
	}

	root = n
	isNil, ok := t.NodeIsNil(root)
	for ok && !isNil {
		root, err = t.GetParent(root)
		if err != nil {
			return err
		}
		root = n
	}
	if !ok {
		return DataError{}
	}

	t.Root = root
	return nil
}

func (t *RBTree) insertRecurse(root *Node, n *Node) error {
	isNil, ok := t.NodeIsNil(root)
	if !ok {
		return DataError{}
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
			return DataError{}
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

	t.newNode(n, Tags["lchild"], NilNodeTag, Colors["nil"])
	t.newNode(n, Tags["rchild"], NilNodeTag, Colors["nil"])

	return nil
}

func (t *RBTree) insertRepairTree(n *Node) error {
	p, err := t.GetParent(n)
	_, ok := err.(NoNodeError)
	if err != nil { // node must have parent
		return err
	}
	parentData, ok := ColorDataFromData(p.Extra)
	if !ok {
		return DataError{}
	}
	parentIsNil, ok := t.NodeIsNil(p)
	if !ok {
		return DataError{}
	}

	u, err := t.GetUncle(n)
	_, ok = err.(NilNodeError)
	// uncle may not exist, so proceed if we find NilNodeError for
	// uncle
	if err != nil && !ok {
		return err
	}
	uncleData, ok := ColorDataFromData(u.Extra)
	if !ok {
		return DataError{}
	}
	uncleIsNil, ok := t.NodeIsNil(u)
	if !ok {
		return DataError{}
	}

	if parentIsNil {
		return t.insertCase1(n)
	} else if parentData.Color == Colors["black"] {
		return t.insertCase2(n)
	} else if !uncleIsNil && uncleData.Color == Colors["red"] {
		return t.insertCase3(n)
	} else {
		return t.insertCase4(n)
	}
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

	if n.ID == t.Root.ID {
		t.Root = child
	}

	return nil
}

func (t *RBTree) DeleteOneChild(n *Node) error {
	// Precondition: n has at most one non-leaf child
	var child *Node
	var err error
	if t.Graph.HasRelative(n, Tags["lchild"]) {
		child, err = t.GetLChild(n)
		if err != nil {
			return err
		}
	} else if t.Graph.HasRelative(n, Tags["rchild"]) {
		child, err = t.GetRChild(n)
		if err != nil {
			return err
		}
	} else {
		return NilNodeError{fmt.Sprintf("Node %d must have at least one child", n.ID)}
	}

	err = t.replaceNode(n, child)
	if err != nil {
		return err
	}

	nodeData, ok := ColorDataFromData(n.Extra)
	if !ok {
		return DataError{}
	}
	childData, ok := ColorDataFromData(child.Extra)
	if !ok {
		return DataError{}
	}

	if nodeData.Color == Colors["black"] {
		if childData.Color == Colors["red"] {
			err = t.setColor(child, Colors["black"])
			if err != nil {
				return err
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
	_, ok := err.(NoEdgeError)
	// Make sure there is not an unexpected error
	if err != nil && !ok {
		return err
	} else if ok {
		return nil
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
		return DataError{}
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
		return DataError{}
	}

	p, err := t.GetParent(n)
	if err != nil {
		return err
	}
	pData, ok := ColorDataFromData(p.Extra)
	if !ok {
		return DataError{}
	}

	rc, err := t.GetRChild(n)
	if err != nil {
		return err
	}
	rcData, ok := ColorDataFromData(rc.Extra)
	if !ok {
		return DataError{}
	}

	lc, err := t.GetLChild(n)
	if err != nil {
		return err
	}
	lcData, ok := ColorDataFromData(lc.Extra)
	if !ok {
		return DataError{}
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
	} else {
		return t.deleteCase4(n)
	}
}

func (t *RBTree) deleteCase4(n *Node) error {
	s, err := t.GetSibling(n)
	if err != nil {
		return err
	}
	sData, ok := ColorDataFromData(s.Extra)
	if !ok {
		return DataError{}
	}

	p, err := t.GetParent(n)
	if err != nil {
		return err
	}
	pData, ok := ColorDataFromData(p.Extra)
	if !ok {
		return DataError{}
	}

	rc, err := t.GetRChild(n)
	if err != nil {
		return err
	}
	rcData, ok := ColorDataFromData(rc.Extra)
	if !ok {
		return DataError{}
	}

	lc, err := t.GetLChild(n)
	if err != nil {
		return err
	}
	lcData, ok := ColorDataFromData(lc.Extra)
	if !ok {
		return DataError{}
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
		return DataError{}
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
		return DataError{}
	}

	sLc, err := t.GetLChild(s)
	if err != nil {
		return err
	}
	sLcData, ok := ColorDataFromData(sLc.Extra)
	if !ok {
		return DataError{}
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
		return DataError{}
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
			return DataError{}
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
			return DataError{}
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
