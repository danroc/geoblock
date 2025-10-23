package maps

// OrderedMap is a map that preserves the insertion order of keys.
type OrderedMap[K comparable, V any] struct {
	keys []K
	data map[K]V
}

// NewOrdered returns a new, empty OrderedMap.
func NewOrdered[K comparable, V any]() OrderedMap[K, V] {
	return OrderedMap[K, V]{
		data: make(map[K]V),
	}
}

// Get returns the value for key, along with a boolean indicating whether it exists.
func (m OrderedMap[K, V]) Get(key K) (V, bool) {
	v, ok := m.data[key]
	return v, ok
}

// Len returns the number of elements in the map.
func (m OrderedMap[K, V]) Len() int {
	return len(m.keys)
}

// Empty reports whether the map has no elements.
func (m OrderedMap[K, V]) Empty() bool {
	return len(m.keys) == 0
}

// Has reports whether key is present in the map.
func (m OrderedMap[K, V]) Has(key K) bool {
	_, ok := m.data[key]
	return ok
}

// Set sets the value for key, preserving insertion order for new keys.
func (m *OrderedMap[K, V]) Set(key K, value V) {
	if _, exists := m.data[key]; !exists {
		m.keys = append(m.keys, key)
	}
	m.data[key] = value
}

// Keys returns a copy of the keys in insertion order.
func (m OrderedMap[K, V]) Keys() []K {
	keys := make([]K, len(m.keys))
	copy(keys, m.keys)
	return keys
}

// Values returns the values in insertion order.
func (m OrderedMap[K, V]) Values() []V {
	values := make([]V, len(m.keys))
	for i, k := range m.keys {
		values[i] = m.data[k]
	}
	return values
}

// Range calls fn for each key/value pair in insertion order.
// If fn returns false, iteration stops.
func (m OrderedMap[K, V]) Range(fn func(K, V) bool) {
	for _, k := range m.keys {
		if !fn(k, m.data[k]) {
			break
		}
	}
}
