package itree

// Comparable is an interface for types that can be compared.
type Comparable[V any] interface {
	Compare(other V) int
}

// Interval represents the `[Low, High]` interval (inclusive).
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
func NewNode[K Comparable[K], V any](
	interval Interval[K],
	value V,
) *Node[K, V] {
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

// maxOf returns the maximum value between the `max` property of the receiver
// and a given value.
func (n *Node[K, V]) maxOf(other K) K {
	if n == nil || other.Compare(n.max) > 0 {
		return other
	}
	return n.max
}

// updateNode updates the `max` and `height` properties of the node.
func (n *Node[K, V]) updateNode() {
	n.max = n.left.maxOf(n.right.maxOf(n.interval.High))
	n.height = 1 + max(n.left.getHeight(), n.right.getHeight())
}

// roteLeft rotates the node to the left.
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

// Query returns the values associated with the intervals that contain the
// given key.
func (t *ITree[K, V]) Query(key K) []V {
	return query(t.root, key)
}

func query[K Comparable[K], V any](
	node *Node[K, V],
	key K,
) []V {
	// If the maximum of all intervals from this node and below is less than
	// the key, there are no intervals to query.
	if node == nil || node.max.Compare(key) < 0 {
		return nil
	}

	var results []V

	// Even if the current interval contains the key, we still need to query
	// the subtrees since they can also contain intervals that cover the key.
	if node.interval.Contains(key) {
		results = append(results, node.value)
	}

	// After a re-balance, both the left and right children of a node can have
	// the same low value. In this case, we need to query both subtrees.
	//
	// However, if the key is less than the low value of the interval, we know
	// that it can only be in the left subtree, so the right subtree can be
	// ignored.
	if key.Compare(node.interval.Low) >= 0 {
		results = append(results, query(node.right, key)...)
	}

	// The left subtree is always queried since it can contain intervals that
	// cover any range in the ]-∞, node.max] interval.
	return append(results, query(node.left, key)...)
}
