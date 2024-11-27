package itree

import (
	"github.com/danroc/geoblock/internal/itree/cmp"
)

// Node represents a node in the interval tree.
type Node[C cmp.Comparable[C], V any] struct {
	left     *Node[C, V]
	right    *Node[C, V]
	interval Interval[C]
	value    V
	max      C
	height   int
}

// NewNode creates a new node with the given interval.
func NewNode[C cmp.Comparable[C], V any](i Interval[C], v V) *Node[C, V] {
	return &Node[C, V]{
		interval: i,
		value:    v,
		max:      i.High,
	}
}

// IsLeaf returns whether the node is a leaf.
func (n *Node[C, V]) IsLeaf() bool {
	return n.left == nil && n.right == nil
}

// Height returns the height of the node.
func (n *Node[C, V]) Height() int {
	if n == nil {
		return -1
	}
	return n.height
}

// Max returns the maximum value between the `max` property of a node and
// another value.
func (n *Node[C, V]) Max(o C) C {
	if n == nil {
		return o
	}
	return cmp.Max(n.max, o)
}

// updateNode updates the `max` and `height` properties of the node.
func (n *Node[C, V]) updateNode() {
	n.max = n.left.Max(n.right.Max(n.interval.High))
	n.height = 1 + max(n.left.Height(), n.right.Height())
}

// roteLeft rotates the node to the left.
func (n *Node[C, V]) rotateLeft() *Node[C, V] {
	x := n.right
	n.right = x.left
	x.left = n

	n.updateNode()
	x.updateNode()

	return x
}

// rotateRight rotates the node to the right.
func (n *Node[C, V]) rotateRight() *Node[C, V] {
	x := n.left
	n.left = x.right
	x.right = n

	n.updateNode()
	x.updateNode()

	return x
}

// balanceFactor returns the balance factor of the node.
func (n *Node[C, V]) balanceFactor() int {
	return n.left.Height() - n.right.Height()
}

// balance balances the node using the AVL algorithm.
func (n *Node[C, V]) balance() *Node[C, V] {
	n.updateNode()
	if bf := n.balanceFactor(); bf > 1 {
		if n.left.balanceFactor() < 0 {
			n.left = n.left.rotateLeft()
		}
		return n.rotateRight()
	} else if bf < -1 {
		if n.right.balanceFactor() > 0 {
			n.right = n.right.rotateRight()
		}
		return n.rotateLeft()
	}
	return n
}

// insert inserts an interval and its value into the interval tree.
func insert[C cmp.Comparable[C], V any](
	n *Node[C, V],
	i Interval[C],
	v V,
) *Node[C, V] {
	if n == nil {
		return NewNode(i, v)
	}

	if i.Low.Compare(n.interval.Low) <= 0 {
		n.left = insert(n.left, i, v)
	} else {
		n.right = insert(n.right, i, v)
	}
	return n.balance()
}

// ITree represents an interval tree.
type ITree[C cmp.Comparable[C], V any] struct {
	root *Node[C, V]
}

// NewITree creates a new interval tree.
func NewITree[C cmp.Comparable[C], V any]() *ITree[C, V] {
	return &ITree[C, V]{}
}

// Insert adds an interval to the interval tree.
func (t *ITree[C, V]) Insert(interval Interval[C], value V) {
	t.root = insert(t.root, interval, value)
}

// Height returns the height of the interval tree.
func (t *ITree[C, V]) Height() int {
	return t.root.Height()
}

func query[I Interval[C], C cmp.Comparable[C], V any](n *Node[C, V], x C) []V {
	if n == nil || n.max.Compare(x) < 0 {
		return nil
	}

	var r []V
	if n.interval.Contains(x) {
		r = append(r, n.value)
	}
	if x.Compare(n.interval.Low) > 0 {
		r = append(r, query(n.right, x)...)
	}
	return append(r, query(n.left, x)...)
}

// Query returns the values associated with the intervals that contain the
// given `x` value.
func (t *ITree[C, V]) Query(x C) []V {
	return query(t.root, x)
}
