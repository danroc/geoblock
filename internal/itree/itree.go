// Package itree provides an interval tree implementation.
package itree

import "slices"

// Comparable is an interface for types that can be compared.
type Comparable[V any] interface {
	Compare(other V) int
}

// Interval represents the [Low, High] interval (inclusive).
type Interval[V Comparable[V]] struct {
	Low  V
	High V
}

// NewInterval creates a new interval with the given low and high values.
func NewInterval[V Comparable[V]](low, high V) Interval[V] {
	return Interval[V]{Low: low, High: high}
}

// Contains returns whether the interval contains the given value.
func (i Interval[V]) Contains(value V) bool {
	return i.Low.Compare(value) <= 0 && value.Compare(i.High) <= 0
}

// Compare returns a negative value if i < other, zero if i == other, and a positive
// value if i > other. Intervals are ordered by their low value first, then by their
// high value.
func (i Interval[V]) Compare(other Interval[V]) int {
	if cmp := i.Low.Compare(other.Low); cmp != 0 {
		return cmp
	}
	return i.High.Compare(other.High)
}

// Equal returns whether the interval is equal to another interval.
func (i Interval[V]) Equal(other Interval[V]) bool {
	return i.Compare(other) == 0
}

// Node represents a node in the interval tree.
type Node[K Comparable[K], V any] struct {
	left     *Node[K, V]
	right    *Node[K, V]
	interval Interval[K]
	value    V
	max      K
	height   int
}

// NewNode creates a new node with the given interval.
func NewNode[K Comparable[K], V any](interval Interval[K], value V) *Node[K, V] {
	return &Node[K, V]{
		interval: interval,
		value:    value,
		max:      interval.High,
	}
}

// getHeight returns the height of the node.
func (n *Node[K, V]) getHeight() int {
	if n == nil {
		return -1
	}
	return n.height
}

// maxOf returns the maximum value between the max property of the receiver and a given
// value.
func (n *Node[K, V]) maxOf(other K) K {
	if n == nil || other.Compare(n.max) > 0 {
		return other
	}
	return n.max
}

// updateNode updates the max and height properties of the node.
func (n *Node[K, V]) updateNode() {
	n.max = n.left.maxOf(n.right.maxOf(n.interval.High))
	n.height = 1 + max(n.left.getHeight(), n.right.getHeight())
}

// rotateLeft rotates the node to the left.
func (n *Node[K, V]) rotateLeft() *Node[K, V] {
	x := n.right
	n.right = x.left
	x.left = n

	n.updateNode()
	x.updateNode()

	return x
}

// rotateRight rotates the node to the right.
func (n *Node[K, V]) rotateRight() *Node[K, V] {
	x := n.left
	n.left = x.right
	x.right = n

	n.updateNode()
	x.updateNode()

	return x
}

// balanceFactor returns the balance factor of the node.
func (n *Node[K, V]) balanceFactor() int {
	return n.left.getHeight() - n.right.getHeight()
}

// balance balances the node using the AVL algorithm.
func (n *Node[K, V]) balance() *Node[K, V] {
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
func insert[K Comparable[K], V any](
	node *Node[K, V],
	interval Interval[K],
	value V,
) *Node[K, V] {
	if node == nil {
		return NewNode(interval, value)
	}

	if interval.Low.Compare(node.interval.Low) <= 0 {
		node.left = insert(node.left, interval, value)
	} else {
		node.right = insert(node.right, interval, value)
	}
	return node.balance()
}

// ITree represents an interval tree.
type ITree[K Comparable[K], V any] struct {
	root *Node[K, V]
}

// NewITree creates a new interval tree.
func NewITree[K Comparable[K], V any]() *ITree[K, V] {
	return &ITree[K, V]{}
}

// Insert adds an interval to the interval tree.
func (t *ITree[K, V]) Insert(interval Interval[K], value V) {
	t.root = insert(t.root, interval, value)
}

// Query returns the values associated with the intervals that contain the given key.
func (t *ITree[K, V]) Query(key K) []V {
	return query(t.root, key, nil)
}

// Traverse walks the tree in pre-order (root, left, right) and calls the provided
// function for each node with its interval and value.
func (t *ITree[K, V]) Traverse(fn func(interval Interval[K], value V)) {
	traverse(t.root, fn)
}

// Size returns the number of nodes in the tree.
func (t *ITree[K, V]) Size() int {
	return size(t.root)
}

// size recursively counts the number of nodes in the subtree.
func size[K Comparable[K], V any](node *Node[K, V]) int {
	if node == nil {
		return 0
	}
	return 1 + size(node.left) + size(node.right)
}

// traverse is a helper function that performs pre-order traversal of the tree.
func traverse[K Comparable[K], V any](
	node *Node[K, V],
	fn func(interval Interval[K], value V),
) {
	if node == nil {
		return
	}

	fn(node.interval, node.value)
	traverse(node.left, fn)
	traverse(node.right, fn)
}

func query[K Comparable[K], V any](node *Node[K, V], key K, results []V) []V {
	// If the maximum of all intervals from this node and below is less than the key,
	// there are no intervals to query.
	if node == nil || node.max.Compare(key) < 0 {
		return results
	}

	// Even if the current interval contains the key, we still need to query the
	// subtrees since they can also contain intervals that cover the key.
	if node.interval.Contains(key) {
		results = append(results, node.value)
	}

	// After a rebalance, both the left and right children of a node can have the same
	// low value. In this case, we need to query both subtrees.
	//
	// However, if the key is less than the low value of the interval, we know that it
	// can only be in the left subtree, so the right subtree can be ignored.
	if key.Compare(node.interval.Low) >= 0 {
		results = query(node.right, key, results)
	}

	// The left subtree is always queried since it can contain intervals that cover any
	// range in the ]-∞, node.max] interval.
	return query(node.left, key, results)
}

// Entry represents an interval and its associated value in the tree.
type Entry[K Comparable[K], V any] struct {
	Interval Interval[K]
	Value    V
}

// Entries returns all entries in the tree as a slice of Entry structs.
func (t *ITree[K, V]) Entries() []Entry[K, V] {
	var entries []Entry[K, V]
	t.Traverse(func(interval Interval[K], value V) {
		entries = append(entries, Entry[K, V]{
			Interval: interval,
			Value:    value,
		})
	})
	return entries
}

// Compacted merges nodes with identical intervals using the provided merge function.
// Returns a new tree where each unique interval appears exactly once.
func (t *ITree[K, V]) Compacted(merge func([]V) V) *ITree[K, V] {
	entries := t.Entries()

	slices.SortFunc(entries, func(a, b Entry[K, V]) int {
		return a.Interval.Compare(b.Interval)
	})

	result := NewITree[K, V]()
	for i := 0; i < len(entries); {
		interval := entries[i].Interval
		values := []V{entries[i].Value}

		j := i + 1
		for j < len(entries) && entries[j].Interval.Equal(interval) {
			values = append(values, entries[j].Value)
			j++
		}

		result.Insert(interval, merge(values))
		i = j
	}

	return result
}
