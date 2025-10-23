package metrics

import (
	"sort"

	"github.com/danroc/geoblock/internal/utils/maps"
)

// Histogram represents a histogram metric.
type Histogram struct {
	buckets maps.OrderedMap[float64, uint64]
	sum     float64
	count   uint64
}

// NewHistogram returns a new Histogram with the given bucket upper bounds.
// Buckets must be sorted in ascending order.
func NewHistogram(buckets []float64) *Histogram {
	h := &Histogram{
		buckets: maps.NewOrdered[float64, uint64](),
	}

	// Make a copy of the original slice, and sort it to ensure it is in
	// ascending order.
	sorted := make([]float64, len(buckets))
	copy(sorted, buckets)
	sort.Sort(sort.Reverse(sort.Float64Slice(sorted)))

	for _, b := range sorted {
		h.buckets.Set(b, 0)
	}
	return h
}

// Observe records a new observation in the histogram.
func (h *Histogram) Observe(value float64) {
	h.sum += value
	h.count++

	h.buckets.Range(func(upper float64, count uint64) bool {
		if value <= upper {
			h.buckets.Set(upper, count+1)
		}
		return true
	})
}

// Sum returns the total of all observed values.
func (h *Histogram) Sum() float64 {
	return h.sum
}

// Count returns the total number of observations.
func (h *Histogram) Count() uint64 {
	return h.count
}

// LinearBuckets creates a slice of linearly spaced bucket upper bounds.
func LinearBuckets(start, width float64, count int) []float64 {
	buckets := make([]float64, count)
	for i := range count {
		buckets[i] = start + float64(i)*width
	}
	return buckets
}
