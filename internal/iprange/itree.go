package iprange

import (
	"fmt"
)

// Interval represents an IP range.
type Interval struct {
	Start int
	End   int
}

// Contains returns true if the interval contains the given IP address.
// Otherwise, it returns false.
func (i Interval) Contains(x int) bool {
	return i.Start <= x && x <= i.End
}

// Node represents a node in the interval tree.
type Node struct {
	Value    int
	Left     *Node
	Right    *Node
	Interval Interval
	maxEnd   int
	height   int
}

// NewNode creates a new node with the given interval.
func NewNode(i Interval, v int) *Node {
	return &Node{
		Interval: i,
		Value:    v,
		maxEnd:   i.End,
	}
}

// Height returns the height of the node.
func (n *Node) Height() int {
	if n == nil {
		return -1
	}
	return n.height
}

func (n *Node) MaxEnd() int {
	if n == nil {
		return -1
	}
	return n.maxEnd
}

func (n *Node) updateNode() {
	n.maxEnd = max(n.Interval.End, max(n.Left.MaxEnd(), n.Right.MaxEnd()))
	n.height = 1 + max(n.Left.Height(), n.Right.Height())
}

func (n *Node) rotateLeft() *Node {
	x := n.Right
	n.Right = x.Left
	x.Left = n

	n.updateNode()
	x.updateNode()

	return x
}

func (n *Node) rotateRight() *Node {
	x := n.Left
	n.Left = x.Right
	x.Right = n

	n.updateNode()
	x.updateNode()

	return x
}

func (n *Node) balanceFactor() int {
	return n.Left.Height() - n.Right.Height()
}

func (n *Node) balance() *Node {
	n.updateNode()
	if bf := n.balanceFactor(); bf > 1 {
		if n.Left.balanceFactor() < 0 {
			n.Left = n.Left.rotateLeft()
		}
		return n.rotateRight()
	} else if bf < -1 {
		if n.Right.balanceFactor() > 0 {
			n.Right = n.Right.rotateRight()
		}
		return n.rotateLeft()
	}
	return n
}

func insert(n *Node, i Interval, x int) *Node {
	if n == nil {
		return NewNode(i, x)
	}

	if i.Start <= n.Interval.Start {
		n.Left = insert(n.Left, i, x)
	} else {
		n.Right = insert(n.Right, i, x)
	}
	return n.balance()
}

// ITree represents an interval tree.
type ITree struct {
	root *Node
}

// NewITree creates a new interval tree.
func NewITree() *ITree {
	return &ITree{}
}

// Insert adds an interval to the interval tree.
func (t *ITree) Insert(interval Interval, value int) {
	t.root = insert(t.root, interval, value)
}

func printNode(n *Node) {
	if n == nil {
		print("nil")
		return
	}

	fmt.Print("<")
	printNode(n.Left)
	fmt.Printf(", [%d,%d]:%d, ", n.Interval.Start, n.Interval.End, n.Value)
	printNode(n.Right)
	fmt.Print(">")
}

// Print prints the interval tree.
func (t *ITree) Print() {
	printNode(t.root)
	fmt.Println()
}

// Height returns the height of the interval tree.
func (t *ITree) Height() int {
	return t.root.Height()
}

func query(n *Node, x int) []int {
	if n == nil || x > n.MaxEnd() {
		return nil
	}

	var r []int
	if n.Interval.Contains(x) {
		r = append(r, n.Value)
	}
	if x > n.Interval.Start {
		r = append(r, query(n.Right, x)...)
	}
	return append(r, query(n.Left, x)...)
}

func (t *ITree) Query(x int) []int {
	return query(t.root, x)
}
