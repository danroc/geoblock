// Package maps provides utility functions for working with maps.
package maps

import (
	"cmp"
	"maps"
	"slices"
)

// Keys returns the keys of the given map in no particular order.
func Keys[M ~map[K]V, K comparable, V any](m M) []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// SortedKeys returns the keys of the given map sorted in ascending order.
func SortedKeys[M ~map[K]V, K cmp.Ordered, V any](m M) []K {
	keys := Keys(m)
	slices.Sort(keys)
	return keys
}

func Merge[M ~map[K]V, K comparable, V any](a, b M) M {
	out := make(M, len(a)+len(b))
	maps.Copy(out, a)
	maps.Copy(out, b)
	return out
}
