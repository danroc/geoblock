package stats

import (
	"math"
	"sync"

	"github.com/danroc/geoblock/internal/utils/maps"
)

// Histogram represents a histogram metric.
type Histogram struct {
	mu      sync.RWMutex
	buckets maps.OrderedMap[float64, uint64]
	sum     float64
	count   uint64
}

// Bucket represents a single bucket in the histogram.
type Bucket struct {
	UpperBound float64
	Count      uint64
}

// NewHistogram returns a new Histogram with the given bucket upper bounds.
// A +Inf bucket is automatically added if not present.
func NewHistogram(buckets []float64) *Histogram {
	h := &Histogram{
		buckets: maps.NewOrdered[float64, uint64](),
	}

	hasInf := false
	for _, b := range buckets {
		h.buckets.Set(b, 0)
		if math.IsInf(b, 1) {
			hasInf = true
		}
	}

	if !hasInf {
		h.buckets.Set(math.Inf(1), 0)
	}

	return h
}

// Observe records a new observation in the histogram.
func (h *Histogram) Observe(value float64) {
	h.mu.Lock()
	defer h.mu.Unlock()

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
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.sum
}

// Count returns the total number of observations.
func (h *Histogram) Count() uint64 {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.count
}

// Buckets returns a snapshot of the histogram's buckets.
func (h *Histogram) Buckets() []Bucket {
	h.mu.RLock()
	defer h.mu.RUnlock()

	result := make([]Bucket, 0, h.buckets.Len())
	h.buckets.Range(func(upper float64, count uint64) bool {
		result = append(result, Bucket{
			UpperBound: upper,
			Count:      count,
		})
		return true
	})
	return result
}

// LinearBuckets creates a slice of linearly spaced bucket upper bounds.
func LinearBuckets(start, width float64, count int) []float64 {
	buckets := make([]float64, count)
	for i := range count {
		buckets[i] = start + float64(i)*width
	}
	return buckets
}
