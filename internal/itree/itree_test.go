package itree_test

import (
	"cmp"
	"fmt"
	"slices"
	"testing"

	"github.com/danroc/geoblock/internal/itree"
)

type ComparableInt int

func (t ComparableInt) Compare(other ComparableInt) int {
	return int(t - other)
}

func TestQuery(t *testing.T) {
	tree := itree.New[ComparableInt, int]()

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

	for _, tt := range tests {
		t.Run(fmt.Sprintf("Query(%d)", tt.key), func(t *testing.T) {
			got := slices.Clone(tree.Query(tt.key))
			slices.Sort(got)

			want := slices.Clone(tt.matches)
			slices.Sort(want)

			if !slices.Equal(got, want) {
				t.Errorf("expected %v, got %v", tt.matches, got)
			}
		})
	}
}

func TestQuery_Duplicate(t *testing.T) {
	tree := itree.New[ComparableInt, int]()
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

	for _, tt := range tests {
		t.Run(fmt.Sprintf("Query(%d)", tt.key), func(t *testing.T) {
			matches := tree.Query(tt.key)
			if !slices.Equal(tt.matches, matches) {
				t.Errorf("expected %v, got %v", tt.matches, matches)
			}
		})
	}
}

func TestRotations(t *testing.T) {
	t.Run("LeftRotation", func(t *testing.T) {
		// Insert in ascending order to trigger left rotations
		tree := itree.New[ComparableInt, int]()
		tree.Insert(itree.NewInterval[ComparableInt](1, 1), 1)
		tree.Insert(itree.NewInterval[ComparableInt](2, 2), 2)
		tree.Insert(itree.NewInterval[ComparableInt](3, 3), 3)

		var values []int
		tree.Traverse(func(_ itree.Interval[ComparableInt], value int) {
			values = append(values, value)
		})

		expected := []int{2, 1, 3}
		if !slices.Equal(expected, values) {
			t.Errorf("expected pre-order %v, got %v", expected, values)
		}
	})

	t.Run("RightRotation", func(t *testing.T) {
		// Insert in descending order to trigger right rotations
		tree := itree.New[ComparableInt, int]()
		tree.Insert(itree.NewInterval[ComparableInt](3, 3), 3)
		tree.Insert(itree.NewInterval[ComparableInt](2, 2), 2)
		tree.Insert(itree.NewInterval[ComparableInt](1, 1), 1)

		var values []int
		tree.Traverse(func(_ itree.Interval[ComparableInt], value int) {
			values = append(values, value)
		})

		expected := []int{2, 1, 3}
		if !slices.Equal(expected, values) {
			t.Errorf("expected pre-order %v, got %v", expected, values)
		}
	})

	t.Run("LeftRightRotation", func(t *testing.T) {
		// Insert to trigger left-right rotation
		tree := itree.New[ComparableInt, int]()
		tree.Insert(itree.NewInterval[ComparableInt](3, 3), 3)
		tree.Insert(itree.NewInterval[ComparableInt](1, 1), 1)
		tree.Insert(itree.NewInterval[ComparableInt](2, 2), 2)

		var values []int
		tree.Traverse(func(_ itree.Interval[ComparableInt], value int) {
			values = append(values, value)
		})

		expected := []int{2, 1, 3}
		if !slices.Equal(expected, values) {
			t.Errorf("expected pre-order %v, got %v", expected, values)
		}
	})

	t.Run("RightLeftRotation", func(t *testing.T) {
		// Insert to trigger right-left rotation
		tree := itree.New[ComparableInt, int]()
		tree.Insert(itree.NewInterval[ComparableInt](1, 1), 1)
		tree.Insert(itree.NewInterval[ComparableInt](3, 3), 3)
		tree.Insert(itree.NewInterval[ComparableInt](2, 2), 2)

		var values []int
		tree.Traverse(func(_ itree.Interval[ComparableInt], value int) {
			values = append(values, value)
		})

		expected := []int{2, 1, 3}
		if !slices.Equal(expected, values) {
			t.Errorf("expected pre-order %v, got %v", expected, values)
		}
	})

	t.Run("SameLowValueDifferentHigh", func(t *testing.T) {
		// Insert 3 intervals with the same low value
		tree := itree.New[ComparableInt, int]()
		tree.Insert(itree.NewInterval[ComparableInt](1, 1), 1)
		tree.Insert(itree.NewInterval[ComparableInt](1, 2), 2)
		tree.Insert(itree.NewInterval[ComparableInt](1, 3), 3)

		var values []int
		tree.Traverse(func(_ itree.Interval[ComparableInt], value int) {
			values = append(values, value)
		})

		expected := []int{2, 3, 1}
		if !slices.Equal(expected, values) {
			t.Errorf("expected pre-order %v, got %v", expected, values)
		}
	})
}

func TestInterval_Equal(t *testing.T) {
	tests := []struct {
		name string
		a    itree.Interval[ComparableInt]
		b    itree.Interval[ComparableInt]
		want bool
	}{
		{
			name: "equal intervals",
			a:    itree.NewInterval[ComparableInt](1, 5),
			b:    itree.NewInterval[ComparableInt](1, 5),
			want: true,
		},
		{
			name: "different low",
			a:    itree.NewInterval[ComparableInt](1, 5),
			b:    itree.NewInterval[ComparableInt](2, 5),
			want: false,
		},
		{
			name: "different high",
			a:    itree.NewInterval[ComparableInt](1, 5),
			b:    itree.NewInterval[ComparableInt](1, 6),
			want: false,
		},
		{
			name: "both different",
			a:    itree.NewInterval[ComparableInt](1, 5),
			b:    itree.NewInterval[ComparableInt](2, 6),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.a.Equal(tt.b); got != tt.want {
				t.Errorf("%v.Equal(%v) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestInterval_Compare(t *testing.T) {
	tests := []struct {
		name string
		a    itree.Interval[ComparableInt]
		b    itree.Interval[ComparableInt]
		want int
	}{
		{
			name: "equal intervals",
			a:    itree.NewInterval[ComparableInt](1, 5),
			b:    itree.NewInterval[ComparableInt](1, 5),
			want: 0,
		},
		{
			name: "a.Low < b.Low",
			a:    itree.NewInterval[ComparableInt](1, 5),
			b:    itree.NewInterval[ComparableInt](2, 5),
			want: -1,
		},
		{
			name: "a.Low > b.Low",
			a:    itree.NewInterval[ComparableInt](2, 5),
			b:    itree.NewInterval[ComparableInt](1, 5),
			want: 1,
		},
		{
			name: "same low, a.High < b.High",
			a:    itree.NewInterval[ComparableInt](1, 4),
			b:    itree.NewInterval[ComparableInt](1, 5),
			want: -1,
		},
		{
			name: "same low, a.High > b.High",
			a:    itree.NewInterval[ComparableInt](1, 6),
			b:    itree.NewInterval[ComparableInt](1, 5),
			want: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.a.Compare(tt.b); got != tt.want {
				t.Errorf("%v.Compare(%v) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func compareEntries(
	a, b itree.Entry[ComparableInt, int],
) int {
	if c := a.Interval.Compare(b.Interval); c != 0 {
		return c
	}
	return cmp.Compare(a.Value, b.Value)
}

func TestEntries(t *testing.T) {
	tree := itree.New[ComparableInt, int]()
	tree.Insert(itree.NewInterval[ComparableInt](1, 2), 10)
	tree.Insert(itree.NewInterval[ComparableInt](3, 4), 20)
	tree.Insert(itree.NewInterval[ComparableInt](5, 6), 30)

	if got := tree.Size(); got != 3 {
		t.Errorf("Size() = %d, want 3", got)
	}

	got := tree.Entries()
	slices.SortFunc(got, compareEntries)

	want := []itree.Entry[ComparableInt, int]{
		{
			Interval: itree.NewInterval[ComparableInt](1, 2),
			Value:    10,
		},
		{
			Interval: itree.NewInterval[ComparableInt](3, 4),
			Value:    20,
		},
		{
			Interval: itree.NewInterval[ComparableInt](5, 6),
			Value:    30,
		},
	}
	slices.SortFunc(want, compareEntries)

	if !slices.EqualFunc(got, want, func(a, b itree.Entry[ComparableInt, int]) bool {
		return compareEntries(a, b) == 0
	}) {
		t.Errorf("Entries() = %v, want %v", got, want)
	}
}

func TestSize(t *testing.T) {
	t.Run("empty tree", func(t *testing.T) {
		tree := itree.New[ComparableInt, int]()
		if got := tree.Size(); got != 0 {
			t.Errorf("Size() = %d, want 0", got)
		}
	})

	t.Run("single node", func(t *testing.T) {
		tree := itree.New[ComparableInt, int]()
		tree.Insert(itree.NewInterval[ComparableInt](1, 5), 10)
		if got := tree.Size(); got != 1 {
			t.Errorf("Size() = %d, want 1", got)
		}
	})

	t.Run("multiple nodes", func(t *testing.T) {
		tree := itree.New[ComparableInt, int]()
		tree.Insert(itree.NewInterval[ComparableInt](1, 5), 10)
		tree.Insert(itree.NewInterval[ComparableInt](2, 6), 20)
		tree.Insert(itree.NewInterval[ComparableInt](3, 7), 30)
		if got := tree.Size(); got != 3 {
			t.Errorf("Size() = %d, want 3", got)
		}
	})
}

func TestCompact(t *testing.T) {
	sumMerge := func(values []int) int {
		sum := 0
		for _, v := range values {
			sum += v
		}
		return sum
	}

	t.Run("empty tree", func(t *testing.T) {
		tree := itree.New[ComparableInt, int]()
		compacted := tree.Compacted(sumMerge)
		if got := compacted.Size(); got != 0 {
			t.Errorf("Size() = %d, want 0", got)
		}
	})

	t.Run("no duplicates", func(t *testing.T) {
		tree := itree.New[ComparableInt, int]()
		tree.Insert(itree.NewInterval[ComparableInt](1, 2), 10)
		tree.Insert(itree.NewInterval[ComparableInt](3, 4), 20)
		tree.Insert(itree.NewInterval[ComparableInt](5, 6), 30)

		compacted := tree.Compacted(sumMerge)

		results := make(map[int]int)
		compacted.Traverse(func(interval itree.Interval[ComparableInt], value int) {
			results[int(interval.Low)] = value
		})

		expected := map[int]int{1: 10, 3: 20, 5: 30}
		for k, v := range expected {
			if results[k] != v {
				t.Errorf("expected %d at key %d, got %d", v, k, results[k])
			}
		}
	})

	t.Run("with duplicates", func(t *testing.T) {
		tree := itree.New[ComparableInt, int]()
		tree.Insert(itree.NewInterval[ComparableInt](1, 2), 10)
		tree.Insert(itree.NewInterval[ComparableInt](1, 2), 20)
		tree.Insert(itree.NewInterval[ComparableInt](1, 2), 30)
		tree.Insert(itree.NewInterval[ComparableInt](3, 4), 5)

		compacted := tree.Compacted(sumMerge)

		if got := compacted.Size(); got != 2 {
			t.Errorf("Size() = %d, want 2", got)
		}

		var sum12, sum34 int
		compacted.Traverse(func(interval itree.Interval[ComparableInt], value int) {
			if interval.Equal(itree.NewInterval[ComparableInt](1, 2)) {
				sum12 = value
			}
			if interval.Equal(itree.NewInterval[ComparableInt](3, 4)) {
				sum34 = value
			}
		})

		if sum12 != 60 {
			t.Errorf("expected merged value 60 for [1,2], got %d", sum12)
		}
		if sum34 != 5 {
			t.Errorf("expected value 5 for [3,4], got %d", sum34)
		}
	})

	t.Run("query after compact", func(t *testing.T) {
		tree := itree.New[ComparableInt, int]()
		tree.Insert(itree.NewInterval[ComparableInt](1, 5), 10)
		tree.Insert(itree.NewInterval[ComparableInt](1, 5), 20)
		tree.Insert(itree.NewInterval[ComparableInt](3, 7), 100)

		compacted := tree.Compacted(sumMerge)

		// Query at point 4 should match both [1,5] (merged: 30) and [3,7] (100)
		results := compacted.Query(4)
		if len(results) != 2 {
			t.Errorf("expected 2 results, got %d", len(results))
		}

		got := slices.Clone(results)
		slices.Sort(got)

		want := []int{30, 100}
		if !slices.Equal(got, want) {
			t.Errorf("expected values %v, got %v", want, got)
		}
	})
}
