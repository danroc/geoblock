// Package itree provides an interval tree implementation.
package itree

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

// Interval represents the `[Low, High]` interval (inclusive).
type Interval[C Comparable[C]] struct {
	Low  C
	High C
}

// NewInterval creates a new interval with the given low and high values.
func NewInterval[C Comparable[C]](low, high C) Interval[C] {
	return Interval[C]{Low: low, High: high}
}

// Contains returns whether the interval contains the given value.
func (i Interval[C]) Contains(v C) bool {
	return i.Low.Compare(v) <= 0 && v.Compare(i.High) <= 0
}
