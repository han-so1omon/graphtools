package structures

import (
	"context"
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

func TestRBTree(t *testing.T) {
	log.Printf("Testing RBTree")
	ctx, cancel := context.WithCancel(context.Background())

	t.Run("New RBTree and node data", func(t *testing.T) {
		tree := NewRBTree(ctx, cancel)
		// Check initial conditions
		if tree.Height != 0 {
			t.Fatalf("Tree must begin with height 0")
		}
		if tree.numMaxHeightNodes != 1 {
			t.Fatalf("Tree must begin with root node at height 0")
		}

		// Try to add child with correct tags
		err := tree.putNode(tree.Root, Tags["lchild"], DataNodeTag, Colors["red"])
		if err != nil {
			t.Fatalf("Could not add data child to tree")
		}

		// Try to add child with incorrect tags
		var ok bool
		err = tree.putNode(tree.Root, "faketag", DataNodeTag, Colors["red"])
		_, ok = err.(*TagError)
		if !ok {
			t.Fatalf("putNode with a fake tag should fail with DataError")
		}

		// Try to add child with incorrect node types
		err = tree.putNode(tree.Root, Tags["lchild"], "fakenodetype", Colors["red"])
		_, ok = err.(*NodeTypeTagError)
		if !ok {
			t.Fatalf("putNode with a node type tag should fail with NodeTypeTagError")
		}

		// Try to add second root node
		err = tree.putNode(nil, Tags["root"], DataNodeTag, Colors["red"])
		_, ok = err.(*RootInsertError)
		if !ok {
			t.Fatalf("putNode fail with RootInsertError when attempting to add second root node")
		}

		// Try to add second node with nil parent
		err = tree.putNode(nil, Tags["lchild"], DataNodeTag, Colors["red"])
		_, ok = err.(*NilNodeError)
		if !ok {
			t.Fatalf("putNode fail with NilNodeError when attempting to add second root node")
		}

		isNil, ok := tree.NodeIsNil(tree.Root)
		if isNil || !ok {
			t.Fatalf("Root node should not be nil")
		}
		rootColor, ok := tree.NodeColor(tree.Root)
		if rootColor != Colors["black"] || !ok {
			t.Fatalf("Root node as such should be black")
		}
	})

	t.Run("RBTree node getters and setters", func(t *testing.T) {
		tree := NewRBTree(ctx, cancel)
		tree.putNode(tree.Root, Tags["lchild"], DataNodeTag, Colors["red"])
		tree.putNode(tree.Root, Tags["rchild"], DataNodeTag, Colors["red"])

		parent, err := tree.GetParent(tree.Root)
		if err != nil {
			t.Fatalf("Could not get parent of root node")
		}
		isNil, ok := tree.NodeIsNil(parent)
		if !isNil || !ok {
			t.Fatalf("Parent of root should be nil type node")
		}

		n2, err := tree.GetLChild(tree.Root)
		if err != nil {
			t.Fatalf("Could not get lchild of root node")
		}

		// root is set as rchild to parent by default
		root, err := tree.GetRChild(parent)
		if err != nil || !reflect.DeepEqual(root, tree.Root) {
			t.Fatalf("Could not get root node from nil parent of root")
		}

		parentAgain, err := tree.GetGrandParent(n2)
		if err != nil || !reflect.DeepEqual(parent, parentAgain) {
			t.Fatalf("Could not get root parent as grandparent of root lchild")
		}

		n3, err := tree.GetSibling(n2)
		if err != nil {
			t.Fatalf("Could not get sibling node")
		}

		parentAgain2, err := tree.GetGrandParent(n3)
		if err != nil || !reflect.DeepEqual(parent, parentAgain2) {
			t.Fatalf("Could not get root parent as grandparent of root rchild")
		}

		mockNode := NewNode()
		err = tree.setLChild(n3, mockNode, true)

		n4, err := tree.GetUncle(mockNode)
		if err != nil || !reflect.DeepEqual(n2, n4) {
			t.Fatalf(fmt.Sprintf("Could not get %d as uncle of %d", n2.ID, n4.ID))
		}

		err = tree.setColor(n3, Colors["black"])
		n3ColorData, ok := ColorDataFromData(n3.Extra)
		if !ok || n3ColorData.Color != Colors["black"] {
			t.Fatalf(fmt.Sprintf("Color of %d should be %d", n3.ID, Colors["black"]))
		}
	})

	t.Run("RBTree left rotations", func(t *testing.T) {
		// Set up tree (level 0)
		tree := NewRBTree(ctx, cancel)
		n1 := tree.Root

		// Set up level 1
		tree.putNode(tree.Root, Tags["lchild"], DataNodeTag, Colors["red"])
		tree.putNode(tree.Root, Tags["rchild"], DataNodeTag, Colors["red"])
		n11, _ := tree.GetLChild(tree.Root)
		n12, _ := tree.GetRChild(tree.Root)

		// Set up level 2
		tree.putNode(n11, Tags["lchild"], DataNodeTag, Colors["black"])
		tree.putNode(n11, Tags["rchild"], DataNodeTag, Colors["black"])
		tree.putNode(n12, Tags["lchild"], DataNodeTag, Colors["black"])
		tree.putNode(n12, Tags["rchild"], DataNodeTag, Colors["black"])

		// Try rotation on level 0
		//FIXME node number naming
		p, _ := tree.GetParent(n1)
		n2pTag, _, _ := tree.Graph.GetEdgeTags(n1, p.ID)
		n2, err := tree.GetRChild(n1)
		n21, err := tree.GetLChild(n2)

		err = tree.rotateLeft(n1)
		if err != nil {
			t.Fatalf(fmt.Sprintf("Error rotating left on %d", n1.ID))
		}

		checkRotateLeft(t, tree, n1, p, n2, n21, n2pTag)
		if !reflect.DeepEqual(tree.Root, n12) {
			t.Fatalf(fmt.Sprintf("New root should be %d after rotation", n12.ID))
		}
	})

	t.Run("RBTree right rotations", func(t *testing.T) {
		// Set up tree (level 0)
		tree := NewRBTree(ctx, cancel)
		n1 := tree.Root

		// Set up level 1
		tree.putNode(tree.Root, Tags["lchild"], DataNodeTag, Colors["red"])
		tree.putNode(tree.Root, Tags["rchild"], DataNodeTag, Colors["red"])
		n11, _ := tree.GetLChild(tree.Root)
		n12, _ := tree.GetRChild(tree.Root)

		// Set up level 2
		tree.putNode(n11, Tags["lchild"], DataNodeTag, Colors["black"])
		tree.putNode(n11, Tags["rchild"], DataNodeTag, Colors["black"])
		tree.putNode(n12, Tags["lchild"], DataNodeTag, Colors["black"])
		tree.putNode(n12, Tags["rchild"], DataNodeTag, Colors["black"])

		// Try rotation on level 0
		//FIXME node number naming
		p, _ := tree.GetParent(n1)
		n2pTag, _, _ := tree.Graph.GetEdgeTags(n1, p.ID)
		n2, err := tree.GetLChild(n1)
		n21, err := tree.GetRChild(n2)

		err = tree.rotateRight(n1)
		checkRotateRight(t, tree, n1, p, n2, n21, n2pTag)
		if err != nil {
			t.Fatalf(fmt.Sprintf("Error rotating right on %d", n1.ID))
		}

		if !reflect.DeepEqual(tree.Root, n11) {
			t.Fatalf(fmt.Sprintf("New root should be %d after rotation", n11.ID))
		}
	})

	t.Run("RBTree insertion case 4-2", func(t *testing.T) {
		tree := newMockRBTree(ctx, cancel, t)
		n, _ := tree.GetLChild(tree.Root)
		n, _ = tree.GetLChild(n)
		p, _ := tree.GetParent(n)
		gp, _ := tree.GetParent(p)
		ggp, _ := tree.GetParent(gp)

		p2, _ := tree.GetRChild(p)
		gp2ggpTag, _, err := tree.Graph.GetEdgeTags(gp, ggp.ID)

		err = tree.insertCase4Step2(n)
		if err != nil {
			t.Fatalf(fmt.Sprintf("Could not successfully do insertCase4Step2 on %d", n.ID))
		}

		checkRotateRight(t, tree, gp, ggp, p, p2, gp2ggpTag)

		pData, _ := ColorDataFromData(p.Extra)
		gpData, _ := ColorDataFromData(gp.Extra)

		if pData.Color != Colors["black"] {
			t.Fatalf(fmt.Sprintf("Parent %d color must be black after insertCase4Step2", p.ID))
		}
		if gpData.Color != Colors["red"] {
			t.Fatalf(fmt.Sprintf("Grandparent %d color must be red after insertCase4Step2", p.ID))
		}
	})

	t.Run("RBTree insertion case 4", func(t *testing.T) {
		tree := newMockRBTree(ctx, cancel, t)
		treeCopy := newMockRBTree(ctx, cancel, t)
		if !reflect.DeepEqual(tree.Graph.String(), treeCopy.Graph.String()) {
			t.Fatalf("tree and treeCopy must be equal RBTree representations")
		}

		n, _ := tree.GetLChild(tree.Root)
		n, _ = tree.GetRChild(n)
		p, _ := tree.GetParent(n)
		//gp, _ := tree.GetParent(p)

		err := tree.insertCase4(n)
		if err != nil {
			t.Fatalf(fmt.Sprintf("Could not successfully do insertCase4 on %d", n.ID))
		}

		n, _ = treeCopy.GetLChild(treeCopy.Root)
		n, _ = treeCopy.GetRChild(n)
		p, _ = treeCopy.GetParent(n)

		treeCopy.rotateLeft(p)
		nLeft, _ := tree.GetLChild(n)
		treeCopy.insertCase4Step2(nLeft)
		if !reflect.DeepEqual(tree.Graph.String(), treeCopy.Graph.String()) {
			t.Fatalf("tree and treeCopy must be equal after insertCase4")
		}
	})

	t.Run("RBTree insertion case 2", func(t *testing.T) {
		//trivial
	})
	t.Run("RBTree insertion case 1", func(t *testing.T) {
		// nearly trivial
		tree := newMockRBTree(ctx, cancel, t)
		n, _ := tree.GetLChild(tree.Root)
		err := tree.insertCase1(n)
		nData, _ := ColorDataFromData(n.Extra)
		if err != nil {
			t.Fatalf(fmt.Sprintf("Could not successfully do insertCase1 on %d", n.ID))
		}
		if nData.Color != Colors["black"] {
			t.Fatalf(fmt.Sprintf("tree insertCase1 failed on node %d", n.ID))
		}

		n, _ = tree.GetRChild(n)
		err = tree.insertCase1(n)
		nData, _ = ColorDataFromData(n.Extra)
		if err != nil {
			t.Fatalf(fmt.Sprintf("Could not successfully do insertCase1 on %d", n.ID))
		}
		if nData.Color != Colors["black"] {
			t.Fatalf(fmt.Sprintf("tree insertCase1 failed on node %d", n.ID))
		}
	})

	t.Run("RBTree insertion case 1", func(t *testing.T) {
		fmt.Println("TODO")
		// insertCase3 is recursive with insertRepairTree
		// insertCase3 is also pretty trivial, so skip debugging it by itself
	})
	//func (t *RBTree) Insert(root *Node, n *Node) error {
	//func (t *RBTree) insertRecurse(root *Node, n *Node) error {
	//func (t *RBTree) insertRepairTree(n *Node) error {
	//func (t *RBTree) insertCase3(n *Node) error {

	t.Run("RBTree replaceNode", func(t *testing.T) {
		tree := newMockRBTree(ctx, cancel, t)
		n := tree.Root
		p, _ := tree.GetParent(n)
		n1, _ := tree.GetLChild(n)
		err := tree.replaceNode(n, n1)
		if err != nil {
			t.Fatalf(fmt.Sprintf("Could not successfully do replaceNode on %d", n.ID))
		}
		_, _, err = tree.Graph.GetEdgeTags(n, p.ID)
		_, ok := err.(*NoEdgeError)
		if !ok {
			t.Fatalf(fmt.Sprintf("No edge should exist between %d and %d", n.ID, p.ID))
		}

		n12pTag, _, err := tree.Graph.GetEdgeTags(n1, p.ID)
		if err != nil || n12pTag != Tags["rchild"] {
			t.Fatalf(fmt.Sprintf("Node %d should be %s of node %d", n1.ID, n12pTag, p.ID))
		}
	})

	t.Run("RBTree deleteCase6", func(t *testing.T) {
		tree := newMockRBTree(ctx, cancel, t)
		n := tree.Root
		n, _ = tree.GetLChild(n)
		n1, _ := tree.GetLChild(n)
		tree.deleteCase6(n1)

		treeCopy := newMockRBTree(ctx, cancel, t)
		n = treeCopy.Root
		n, _ = tree.GetLChild(n)
		n1, _ = treeCopy.GetLChild(n)
		s1, _ := treeCopy.GetSibling(n1)
		s11, _ := treeCopy.GetLChild(s1)
		nData, _ := ColorDataFromData(n.Extra)
		treeCopy.setColor(s1, nData.Color)
		treeCopy.setColor(n, Colors["black"])
		treeCopy.setColor(s11, Colors["black"])
		treeCopy.rotateLeft(n)
		if !reflect.DeepEqual(tree.Graph.String(), treeCopy.Graph.String()) {
			t.Fatalf("tree and treeCopy must be equal after deleteCase6")
		}
	})
	//func (t *RBTree) deleteCase6(n *Node) error {

	t.Run("RBTree deleteCase5", func(t *testing.T) {
		//tree := newMockRBTree(ctx, cancel, t)
		fmt.Println("TODO")
	})
	//func (t *RBTree) deleteCase5(n *Node) error {

	t.Run("RBTree deleteCase4", func(t *testing.T) {
		//tree := newMockRBTree(ctx, cancel, t)
		fmt.Println("TODO")
	})
	//func (t *RBTree) deleteCase4(n *Node) error {

	t.Run("RBTree deleteCase3", func(t *testing.T) {
		//tree := newMockRBTree(ctx, cancel, t)
		fmt.Println("TODO")
	})
	//func (t *RBTree) deleteCase3(n *Node) error {

	t.Run("RBTree deleteCase2", func(t *testing.T) {
		//tree := newMockRBTree(ctx, cancel, t)
		fmt.Println("TODO")
	})
	//func (t *RBTree) deleteCase2(n *Node) error {

	t.Run("RBTree deleteCase1", func(t *testing.T) {
		//tree := newMockRBTree(ctx, cancel, t)
		fmt.Println("TODO")
	})
	//func (t *RBTree) deleteCase1(n *Node) error {

	t.Run("RBTree DeleteOneChild", func(t *testing.T) {
		//tree := newMockRBTree(ctx, cancel, t)
		fmt.Println("TODO")
	})

	fmt.Println()
}

func newMockRBTree(ctx context.Context, cancel context.CancelFunc, t *testing.T) *RBTree {
	t.Helper()

	tree := NewRBTree(ctx, cancel)
	tree.GetParent(tree.Root)

	// Set up level 1
	tree.putNode(tree.Root, Tags["lchild"], DataNodeTag, Colors["red"])
	tree.putNode(tree.Root, Tags["rchild"], DataNodeTag, Colors["red"])
	n11, err := tree.GetLChild(tree.Root)
	if err != nil {
		t.Fatalf("Unable to create mock RBTree")
	}
	n12, err := tree.GetRChild(tree.Root)
	if err != nil {
		t.Fatalf("Unable to create mock RBTree")
	}

	// Set up level 2
	tree.putNode(n11, Tags["lchild"], DataNodeTag, Colors["black"])
	tree.putNode(n11, Tags["rchild"], DataNodeTag, Colors["black"])
	tree.putNode(n12, Tags["lchild"], DataNodeTag, Colors["black"])
	tree.putNode(n12, Tags["rchild"], DataNodeTag, Colors["black"])

	n111, err := tree.GetLChild(n11)
	if err != nil {
		t.Fatalf("Unable to create mock RBTree")
	}
	n112, err := tree.GetRChild(n11)
	if err != nil {
		t.Fatalf("Unable to create mock RBTree")
	}

	n121, err := tree.GetLChild(n12)
	if err != nil {
		t.Fatalf("Unable to create mock RBTree")
	}
	n122, err := tree.GetRChild(n12)
	if err != nil {
		t.Fatalf("Unable to create mock RBTree")
	}

	// Set up level 3
	tree.putNode(n111, Tags["lchild"], DataNodeTag, Colors["red"])
	tree.putNode(n111, Tags["rchild"], DataNodeTag, Colors["red"])
	tree.putNode(n112, Tags["lchild"], DataNodeTag, Colors["red"])
	tree.putNode(n112, Tags["rchild"], DataNodeTag, Colors["red"])
	tree.putNode(n121, Tags["lchild"], DataNodeTag, Colors["red"])
	tree.putNode(n121, Tags["rchild"], DataNodeTag, Colors["red"])
	tree.putNode(n122, Tags["lchild"], DataNodeTag, Colors["red"])
	tree.putNode(n122, Tags["rchild"], DataNodeTag, Colors["red"])

	return tree
}

func checkRotateLeft(t *testing.T, tree *RBTree, n, p, n2, n21 *Node, n2pTag string) {
	var err error
	var n2After *Node
	if n2pTag == Tags["rchild"] {
		n2After, err = tree.GetRChild(p)
	} else {
		n2After, err = tree.GetLChild(p)
	}
	if err != nil {
		t.Fatalf(fmt.Sprintf("Could not get %s of %d", n2pTag, n.ID))
	}

	if !reflect.DeepEqual(n2After, n2) {
		t.Fatalf(fmt.Sprintf("New root should be %d after rotation", n2.ID))
	}

	n21After, _ := tree.GetLChild(n2)
	if !reflect.DeepEqual(n21After, n) {
		t.Fatalf(fmt.Sprintf("%d should be lchild of %d after rotation", n.ID, n2.ID))
	}

	n212After, _ := tree.GetRChild(n)
	if !reflect.DeepEqual(n212After, n21) {
		t.Fatalf(fmt.Sprintf("%d should be rchild of %d after rotation", n21.ID, n.ID))
	}
}

func checkRotateRight(t *testing.T, tree *RBTree, n, p, n1, n12 *Node, n2pTag string) {
	var err error
	var n1After *Node
	if n2pTag == Tags["rchild"] {
		n1After, err = tree.GetRChild(p)
	} else {
		n1After, err = tree.GetLChild(p)
	}
	if err != nil {
		t.Fatalf(fmt.Sprintf("Could not get edge %s of %d", n2pTag, n.ID))
	}

	if !reflect.DeepEqual(n1After, n1) {
		t.Fatalf(fmt.Sprintf("New root should be %d after rotation", n1.ID))
	}

	n12After, _ := tree.GetRChild(n1)
	if !reflect.DeepEqual(n12After, n) {
		t.Fatalf(fmt.Sprintf("%d should be rchild of %d after rotation", n.ID, n1.ID))
	}

	n121After, _ := tree.GetLChild(n)
	if !reflect.DeepEqual(n121After, n12) {
		t.Fatalf(fmt.Sprintf("%d should be lchild of %d after rotation", n12.ID, n.ID))
	}
}
