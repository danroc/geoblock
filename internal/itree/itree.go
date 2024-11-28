package itree

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

// Height returns the height of the node.
func (n *Node[K, V]) Height() int {
	if n == nil {
		return -1
	}
	return n.height
}

// Max returns the maximum value between the `max` property of a node and
// another value.
func (n *Node[K, V]) Max(other K) K {
	if n == nil {
		return other
	}
	return Max(n.max, other)
}

// updateNode updates the `max` and `height` properties of the node.
func (n *Node[K, V]) updateNode() {
	n.max = n.left.Max(n.right.Max(n.interval.High))
	n.height = 1 + max(n.left.Height(), n.right.Height())
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
	return n.left.Height() - n.right.Height()
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
	// the value, there are no intervals to query.
	if node == nil || node.max.Compare(key) < 0 {
		return nil
	}

	var results []V

	// Even if the current interval contains the key
	if node.interval.Contains(key) {
		results = append(results, node.value)
	}

	// If the key is less than then or equal to the low value of the interval,
	// we know that it can only be in the left subtree, so the right subtree
	// can be ignored.
	if key.Compare(node.interval.Low) > 0 {
		results = append(results, query(node.right, key)...)
	}

	// The left subtree is always queried since it can contain intervals that
	// cover any range in the ]-∞, node.max] interval.
	return append(results, query(node.left, key)...)
}
