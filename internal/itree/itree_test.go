package itree_test

import (
	"fmt"
	"testing"

	"github.com/danroc/geoblock/internal/itree"
)

type ComparableInt int

func (t ComparableInt) Compare(other ComparableInt) int {
	return int(t - other)
}

func TestQuery(t *testing.T) {
	tree := itree.NewITree[ComparableInt, int]()
	tree.Insert(itree.NewInterval[ComparableInt](1, 3), 1)
	tree.Insert(itree.NewInterval[ComparableInt](4, 8), 2)
	tree.Insert(itree.NewInterval[ComparableInt](6, 10), 3)
	tree.Insert(itree.NewInterval[ComparableInt](11, 13), 4)
	tree.Insert(itree.NewInterval[ComparableInt](1, 13), 5)

	tests := []struct {
		key     ComparableInt
		matches []int
	}{
		{0, []int{}},
		{1, []int{1, 5}},
		{2, []int{1, 5}},
		{3, []int{1, 5}},
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
			if len(matches) != len(test.matches) {
				t.Errorf("expected %v, got %v", test.matches, matches)
			}
			for i, match := range matches {
				if match != test.matches[i] {
					t.Errorf("expected %v, got %v", test.matches, matches)
				}
			}
		})
	}
}

func TestTraverse(t *testing.T) {
	tree := itree.NewITree[ComparableInt, int]()
	tree.Insert(itree.NewInterval[ComparableInt](41, 41), 1)
	tree.Insert(itree.NewInterval[ComparableInt](20, 20), 2)
	tree.Insert(itree.NewInterval[ComparableInt](11, 11), 3)
	tree.Insert(itree.NewInterval[ComparableInt](29, 29), 4)
	tree.Insert(itree.NewInterval[ComparableInt](26, 26), 5)
	tree.Insert(itree.NewInterval[ComparableInt](65, 65), 6)
	tree.Insert(itree.NewInterval[ComparableInt](50, 50), 7)
	tree.Insert(itree.NewInterval[ComparableInt](23, 23), 8)
	tree.Insert(itree.NewInterval[ComparableInt](55, 55), 9)

	for i, v := range []int{3, 8, 5, 4, 2, 1, 7, 9, 6} {
		t.Run(fmt.Sprintf("Traverse(%d)", i), func(t *testing.T) {
			j := 0
			tree.Traverse(func(k int, value int) bool {
				if k != j || value != v {
					t.Errorf("expected %d, got %d", v, value)
				}
				j++
				return true
			})
		})
	}
}
