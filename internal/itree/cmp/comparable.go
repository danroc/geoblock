// Package cmp provides a generic comparable interface and related functions.
package cmp

// Comparable is an interface for types that can be compared.
type Comparable[V any] interface {
	Compare(other V) int
}

// Max returns the maximum value from the given comparable values.
func Max[C Comparable[C]](values ...C) C {
	m := values[0]
	for _, v := range values[1:] {
		if v.Compare(m) > 0 {
			m = v
		}
	}
	return m
}
