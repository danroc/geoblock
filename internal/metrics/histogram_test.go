package metrics

import (
	"math"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestNewHistogram(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		buckets []float64
		want    []float64
	}{
		{
			"empty buckets",
			[]float64{},
			[]float64{},
		},
		{
			"single bucket",
			[]float64{1.0},
			[]float64{1.0},
		},
		{
			"sorted buckets",
			[]float64{1.0, 2.0, 3.0},
			[]float64{1.0, 2.0, 3.0},
		},
		{
			"unsorted buckets",
			[]float64{3.0, 1.0, 2.0},
			[]float64{3.0, 1.0, 2.0},
		},
		{
			"duplicate buckets",
			[]float64{1.0, 2.0, 1.0},
			[]float64{1.0, 2.0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewHistogram(tt.buckets)
			if h == nil {
				t.Fatal("NewHistogram() returned nil")
			}

			if got := h.Sum(); got != 0 {
				t.Errorf("Sum() = %v, want 0", got)
			}
			if got := h.Count(); got != 0 {
				t.Errorf("Count() = %v, want 0", got)
			}

			if diff := cmp.Diff(tt.want, h.buckets.Keys()); diff != "" {
				t.Errorf("bucket keys mismatch (-want +got):\n%s", diff)
			}

			for _, b := range h.buckets.Keys() {
				if c, _ := h.buckets.Get(b); c != 0 {
					t.Errorf("bucket %v count = %v, want 0", b, c)
				}
			}
		})
	}
}

func TestHistogramObserve(t *testing.T) {
	h := NewHistogram([]float64{1.0, 2.0, 5.0, 10.0})
	tests := []struct {
		name        string
		value       float64
		wantSum     float64
		wantCount   uint64
		wantBuckets map[float64]uint64
	}{
		{
			"below first bucket",
			0.5,
			0.5,
			1,
			map[float64]uint64{1.0: 1, 2.0: 1, 5.0: 1, 10.0: 1},
		},
		{
			"middle bucket",
			1.5,
			2.0,
			2,
			map[float64]uint64{1.0: 1, 2.0: 2, 5.0: 2, 10.0: 2},
		},
		{
			"larger bucket",
			7.0,
			9.0,
			3,
			map[float64]uint64{1.0: 1, 2.0: 2, 5.0: 2, 10.0: 3},
		},
		{
			"above all buckets",
			15.0,
			24.0,
			4,
			map[float64]uint64{1.0: 1, 2.0: 2, 5.0: 2, 10.0: 3},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h.Observe(tt.value)

			if got := h.Sum(); got != tt.wantSum {
				t.Errorf("Sum() = %v, want %v", got, tt.wantSum)
			}

			if got := h.Count(); got != tt.wantCount {
				t.Errorf("Count() = %v, want %v", got, tt.wantCount)
			}

			for b, want := range tt.wantBuckets {
				if got, ok := h.buckets.Get(b); !ok {
					t.Errorf("bucket %v not found", b)
				} else if got != want {
					t.Errorf("bucket %v count = %v, want %v", b, got, want)
				}
			}
		})
	}
}

func TestHistogramObserve_SameBucket(t *testing.T) {
	t.Parallel()

	h := NewHistogram([]float64{5.0, 10.0})
	for _, v := range []float64{1.0, 2.0, 3.0} {
		h.Observe(v)
	}

	if diff := math.Abs(h.Sum() - 6.0); diff > 1e-9 {
		t.Errorf("Sum() = %v, want 6.0", h.Sum())
	}
	if got := h.Count(); got != 3 {
		t.Errorf("Count() = %v, want 3", got)
	}

	expected := map[float64]uint64{5.0: 3, 10.0: 3}
	for b, want := range expected {
		if got, _ := h.buckets.Get(b); got != want {
			t.Errorf("bucket %v count = %v, want %v", b, got, want)
		}
	}
}

func TestHistogramObserve_ExactBoundaries(t *testing.T) {
	t.Parallel()

	h := NewHistogram([]float64{1.0, 2.0, 5.0})
	for _, v := range []float64{1.0, 2.0, 5.0} {
		h.Observe(v)
	}

	if got := h.Sum(); got != 8.0 {
		t.Errorf("Sum() = %v, want 8.0", got)
	}
	if got := h.Count(); got != 3 {
		t.Errorf("Count() = %v, want 3", got)
	}

	want := map[float64]uint64{1.0: 1, 2.0: 2, 5.0: 3}
	for b, w := range want {
		if got, _ := h.buckets.Get(b); got != w {
			t.Errorf("bucket %v count = %v, want %v", b, got, w)
		}
	}
}

func TestHistogramObserve_NegativeValues(t *testing.T) {
	t.Parallel()

	h := NewHistogram([]float64{-1.0, 0.0, 1.0})
	for _, v := range []float64{-2.0, -0.5, 0.5} {
		h.Observe(v)
	}

	if got := h.Sum(); got != -2.0 {
		t.Errorf("Sum() = %v, want -2.0", got)
	}
	if got := h.Count(); got != 3 {
		t.Errorf("Count() = %v, want 3", got)
	}

	want := map[float64]uint64{-1.0: 1, 0.0: 2, 1.0: 3}
	for b, w := range want {
		if got, _ := h.buckets.Get(b); got != w {
			t.Errorf("bucket %v count = %v, want %v", b, got, w)
		}
	}
}

func TestHistogramObserve_Infinity(t *testing.T) {
	t.Parallel()

	h := NewHistogram([]float64{1.0, math.Inf(1)})
	h.Observe(math.Inf(1))
	h.Observe(1000.0)

	if !math.IsInf(h.Sum(), 1) {
		t.Errorf("Sum() = %v, want infinity", h.Sum())
	}
}

func TestLinearBuckets(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		start float64
		width float64
		count int
		want  []float64
	}{
		{"zero count", 1.0, 1.0, 0, []float64{}},
		{"single", 1.0, 1.0, 1, []float64{1.0}},
		{"multiple", 1.0, 2.0, 5, []float64{1.0, 3.0, 5.0, 7.0, 9.0}},
		{"fractional", 0.0, 0.5, 4, []float64{0.0, 0.5, 1.0, 1.5}},
		{"negative start", -2.0, 1.0, 3, []float64{-2.0, -1.0, 0.0}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := LinearBuckets(tt.start, tt.width, tt.count)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("LinearBuckets mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestHistogram_DoesNotModifyOriginalBuckets(t *testing.T) {
	t.Parallel()

	orig := []float64{3.0, 1.0, 2.0}
	h := NewHistogram(orig)

	wantOrig := []float64{3.0, 1.0, 2.0}
	if diff := cmp.Diff(wantOrig, orig); diff != "" {
		t.Errorf("original buckets modified (-want +got):\n%s", diff)
	}

	if diff := cmp.Diff(wantOrig, h.buckets.Keys()); diff != "" {
		t.Errorf("histogram buckets mismatch (-want +got):\n%s", diff)
	}
}

func BenchmarkHistogramObserve(b *testing.B) {
	h := NewHistogram(LinearBuckets(0, 1, 100))

	for i := 0; b.Loop(); i++ {
		h.Observe(float64(i % 50))
	}
}
