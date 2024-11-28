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

	tree.Insert(itree.NewInterval[ComparableInt](3, 5), 1)
	tree.Insert(itree.NewInterval[ComparableInt](3, 5), 1)

	want := []int{1, 1}
}
