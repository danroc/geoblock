// Package maps provides utility functions for working with maps.
package maps

import (
	"cmp"
	"slices"
)

// Keys returns the keys of the given map in no particular order.
func Keys[K comparable, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// SortedKeys returns the keys of the given map sorted in ascending order.
func SortedKeys[K cmp.Ordered, V any](m map[K]V) []K {
	keys := Keys(m)
	slices.Sort(keys)
	return keys
}
