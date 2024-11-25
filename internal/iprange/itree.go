package iprange

import (
	"fmt"
	"net"

	"github.com/danroc/geoblock/internal/utils/iputils"
)

// Interval represents an IP range.
type Interval struct {
	Start net.IP
	End   net.IP
}

func (i Interval) Equal(other Interval) bool {
	return iputils.CompareIP(i.Start, other.Start) == 0 &&
		iputils.CompareIP(i.End, other.End) == 0
}

// Node represents a node in the interval tree.
type Node struct {
	Interval Interval
	Left     *Node
	Right    *Node
	height   int
}

func NewNode(i Interval) *Node {
	return &Node{
		Interval: i,
	}
}

func (n *Node) Height() int {
	if n == nil {
		return -1
	}
	return n.height
}

func (n *Node) updateHeight() {
	n.height = 1 + max(n.Left.Height(), n.Right.Height())
}

func (n *Node) rotateLeft() *Node {
	x := n.Right
	n.Right = x.Left
	x.Left = n

	n.updateHeight()
	x.updateHeight()

	return x
}

func (n *Node) rotateRight() *Node {
	x := n.Left
	n.Left = x.Right
	x.Right = n

	n.updateHeight()
	x.updateHeight()

	return x
}

func (n *Node) balanceFactor() int {
	return n.Left.Height() - n.Right.Height()
}

func (n *Node) isLeftHeavy() bool {
	return n.balanceFactor() >= 1
}

func (n *Node) isRightHeavy() bool {
	return n.balanceFactor() <= -1
}

func (n *Node) isTooLeftHeavy() bool {
	return n.balanceFactor() > 1
}

func (n *Node) isTooRightHeavy() bool {
	return n.balanceFactor() < -1
}

func (n *Node) balance() *Node {
	if n.isTooLeftHeavy() {
		if n.Left.isRightHeavy() {
			n.Left = n.Left.rotateLeft()
		}
		return n.rotateRight()
	}

	if n.isTooRightHeavy() {
		if n.Right.isLeftHeavy() {
			n.Right = n.Right.rotateRight()
		}
		return n.rotateLeft()
	}

	return n
}

func insert(n *Node, i Interval) *Node {
	if n == nil {
		return NewNode(i)
	}

	if n.Interval.Equal(i) {
		return n
	}

	if iputils.CompareIP(i.Start, n.Interval.Start) < 0 {
		n.Left = insert(n.Left, i)
	} else {
		n.Right = insert(n.Right, i)
	}

	n.updateHeight()
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
func (t *ITree) Insert(interval Interval) {
	t.root = insert(t.root, interval)
}

func printNode(n *Node) {
	if n == nil {
		print("nil")
		return
	}

	fmt.Print("<")
	printNode(n.Left)
	fmt.Printf(", %s/%s, ", n.Interval.Start, n.Interval.End)
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
