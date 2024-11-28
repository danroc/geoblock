// Package itree provides an interval tree implementation.
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
