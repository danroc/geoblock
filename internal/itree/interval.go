// Package itree provides an interval tree implementation.
package itree

import "github.com/danroc/geoblock/internal/itree/cmp"

// Interval represents the `[Low, High]` interval (inclusive).
type Interval[C cmp.Comparable[C]] struct {
	Low  C
	High C
}

// NewInterval creates a new interval with the given low and high values.
func NewInterval[C cmp.Comparable[C]](low, high C) Interval[C] {
	return Interval[C]{Low: low, High: high}
}

// Contains returns whether the interval contains the given value.
func (i Interval[C]) Contains(v C) bool {
	return i.Low.Compare(v) <= 0 && v.Compare(i.High) <= 0
}
