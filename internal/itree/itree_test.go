package itree_test

import (
	"fmt"
	"slices"
	"testing"

	"github.com/danroc/geoblock/internal/itree"
)

type ComparableInt int

func (t ComparableInt) Compare(other ComparableInt) int {
	return int(t - other)
}

type set[E comparable] map[E]bool

func newSet[E comparable]() set[E] {
	return make(map[E]bool)
}

func (s set[E]) add(e ...E) {
	for _, v := range e {
		s[v] = true
	}
}

func (s set[E]) contains(e E) bool {
	_, ok := s[e]
	return ok
}

func (s set[E]) equal(other set[E]) bool {
	for k := range s {
		if !other.contains(k) {
			return false
		}
	}
	for k := range other {
		if !s.contains(k) {
			return false
		}
	}
	return true
}

func TestQuery(t *testing.T) {
	tree := itree.NewITree[ComparableInt, int]()

	// Default cases
	//
	// 1: [------]
	// 2:          [------------]
	// 3:                [------------]
	// 4:                               [------]
	// 5: [------------------------------------]
	//    01 02 03 04 05 06 07 08 09 10 11 12 13
	tree.Insert(itree.NewInterval[ComparableInt](1, 3), 1)
	tree.Insert(itree.NewInterval[ComparableInt](4, 8), 2)
	tree.Insert(itree.NewInterval[ComparableInt](6, 10), 3)
	tree.Insert(itree.NewInterval[ComparableInt](11, 13), 4)
	tree.Insert(itree.NewInterval[ComparableInt](1, 13), 5)

	// Cases to trigger rotations
	tree.Insert(itree.NewInterval[ComparableInt](1, 1), 6)
	tree.Insert(itree.NewInterval[ComparableInt](1, 1), 7)
	tree.Insert(itree.NewInterval[ComparableInt](3, 3), 8)
	tree.Insert(itree.NewInterval[ComparableInt](3, 3), 9)
	tree.Insert(itree.NewInterval[ComparableInt](3, 3), 10)

	tests := []struct {
		key     ComparableInt
		matches []int
	}{
		{0, []int{}},
		{1, []int{1, 5, 6, 7}},
		{2, []int{1, 5}},
		{3, []int{1, 5, 8, 9, 10}},
		{4, []int{2, 5}},
		{5, []int{2, 5}},
		{6, []int{2, 3, 5}},
		{7, []int{2, 3, 5}},
		{8, []int{2, 3, 5}},
		{9, []int{3, 5}},
		{10, []int{3, 5}},
		{11, []int{4, 5}},
		{12, []int{4, 5}},
		{13, []int{4, 5}},
		{14, []int{}},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("Query(%d)", test.key), func(t *testing.T) {
			matches := tree.Query(test.key)
			got := newSet[int]()
			got.add(matches...)

			want := newSet[int]()
			want.add(test.matches...)

			if !want.equal(got) {
				t.Errorf("expected %v, got %v", test.matches, matches)
			}
		})
	}
}

func TestQueryDuplicate(t *testing.T) {
	tree := itree.NewITree[ComparableInt, int]()
	tree.Insert(itree.NewInterval[ComparableInt](1, 2), 1)
	tree.Insert(itree.NewInterval[ComparableInt](1, 2), 1)

	tree.Insert(itree.NewInterval[ComparableInt](3, 5), 2)
	tree.Insert(itree.NewInterval[ComparableInt](2, 5), 2)

	tests := []struct {
		key     ComparableInt
		matches []int
	}{
		{0, []int{}},
		{1, []int{1, 1}},
		{2, []int{1, 2, 1}},
		{3, []int{2, 2}},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("Query(%d)", test.key), func(t *testing.T) {
			matches := tree.Query(test.key)
			if !slices.Equal(test.matches, matches) {
				t.Errorf("expected %v, got %v", test.matches, matches)
			}
		})
	}
}

// Benchmark tests

func BenchmarkInsert(b *testing.B) {
	sizes := []int{10, 100, 1000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("size_%d", size), func(b *testing.B) {
			// Setup tree once outside the benchmark loop
			tree := itree.NewITree[ComparableInt, int]()
			for j := range size {
				low := ComparableInt(j * 2)
				high := ComparableInt(j*2 + 10)
				tree.Insert(itree.NewInterval(low, high), 0)
			}

			b.ResetTimer()
			for i := range b.N {
				// Benchmark single insert with varying intervals
				low := ComparableInt(size*2 + i)
				high := ComparableInt(size*2 + i + 10)
				tree.Insert(itree.NewInterval(low, high), 0)
			}
		})
	}
}

func BenchmarkInsertSequential(b *testing.B) {
	b.Run("sequential_inserts", func(b *testing.B) {
		for b.Loop() {
			tree := itree.NewITree[ComparableInt, int]()
			for j := range 100 {
				low := ComparableInt(j)
				high := ComparableInt(j + 1)
				tree.Insert(itree.NewInterval(low, high), j)
			}
		}
	})
}

func BenchmarkInsertSameLow(b *testing.B) {
	b.Run("same_low_inserts", func(b *testing.B) {
		for b.Loop() {
			tree := itree.NewITree[ComparableInt, int]()
			for j := range 50 {
				low := ComparableInt(1)
				high := ComparableInt(j + 1)
				tree.Insert(itree.NewInterval(low, high), 0)
			}
		}
	})
}

func BenchmarkInsertSameHigh(b *testing.B) {
	b.Run("same_high_inserts", func(b *testing.B) {
		for b.Loop() {
			tree := itree.NewITree[ComparableInt, int]()
			for j := range 50 {
				low := ComparableInt(j)
				high := ComparableInt(100)
				tree.Insert(itree.NewInterval(low, high), 0)
			}
		}
	})
}

func BenchmarkQuery(b *testing.B) {
	sizes := []int{10, 100, 1000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("size_%d", size), func(b *testing.B) {
			// Setup
			tree := itree.NewITree[ComparableInt, int]()
			for j := range size {
				low := ComparableInt(j * 2)
				high := ComparableInt(j*2 + 10)
				tree.Insert(itree.NewInterval(low, high), j)
			}

			b.ResetTimer()
			for b.Loop() {
				// Query a key that will hit many intervals
				key := ComparableInt(size)
				_ = tree.Query(key)
			}
		})
	}
}

func BenchmarkQueryHitRate(b *testing.B) {
	tree := itree.NewITree[ComparableInt, int]()
	const treeSize = 100

	// Setup tree with overlapping intervals
	for j := 0; j < treeSize; j++ {
		low := ComparableInt(j)
		high := ComparableInt(j + 50)
		tree.Insert(itree.NewInterval(low, high), j)
	}

	b.Run("high_hit_rate", func(b *testing.B) {
		for b.Loop() {
			// Query keys that will match many intervals
			key := ComparableInt(50)
			_ = tree.Query(key)
		}
	})

	b.Run("low_hit_rate", func(b *testing.B) {
		for b.Loop() {
			// Query keys that will match few or no intervals
			key := ComparableInt(treeSize + 1000)
			_ = tree.Query(key)
		}
	})
}

func BenchmarkQueryEmpty(b *testing.B) {
	tree := itree.NewITree[ComparableInt, int]()

	b.ResetTimer()
	for b.Loop() {
		_ = tree.Query(ComparableInt(1))
	}
}

func BenchmarkMixedWorkload(b *testing.B) {
	b.Run("mixed_insert_query", func(b *testing.B) {
		for b.Loop() {
			tree := itree.NewITree[ComparableInt, int]()

			// Mixed workload: 70% queries, 30% inserts
			for j := 0; j < 100; j++ {
				if j%10 < 3 { // 30% inserts
					low := ComparableInt(j)
					high := ComparableInt(j + 10)
					tree.Insert(itree.NewInterval(low, high), j)
				} else { // 70% queries
					key := ComparableInt(j / 2)
					_ = tree.Query(key)
				}
			}
		}
	})
}

func BenchmarkLargeIntervals(b *testing.B) {
	b.Run("large_overlapping_intervals", func(b *testing.B) {
		for b.Loop() {
			tree := itree.NewITree[ComparableInt, int]()

			// Insert large overlapping intervals
			for j := 0; j < 50; j++ {
				low := ComparableInt(j)
				high := ComparableInt(j + 200) // Large intervals
				tree.Insert(itree.NewInterval(low, high), j)
			}

			// Query in the middle where all intervals overlap
			_ = tree.Query(ComparableInt(100))
		}
	})
}
