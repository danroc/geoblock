package set_test

import (
	"testing"

	"github.com/danroc/geoblock/pkg/set"
)

func TestNewSet(t *testing.T) {
	s := set.NewSet[int]()
	if !s.IsEmpty() {
		t.Error("Set should be empty")
	}
}

func TestAdd(t *testing.T) {
	s := set.NewSet[int]()
	s.Add(1)
	if !s.Contains(1) {
		t.Error("Set should contain 1 after adding 1")
	}
}

func TestRemove(t *testing.T) {
	s := set.NewSet[int]()
	s.Add(1)
	s.Remove(1)
	if s.Contains(1) {
		t.Error("Set should not contain 1 after removing 1")
	}
}

func TestCardinality(t *testing.T) {
	s := set.NewSet[int]()
	s.Add(1)
	s.Add(2)
	if s.Cardinality() != 2 {
		t.Error("Set should have cardinality 2")
	}
}

func TestElements(t *testing.T) {
	s := set.NewSet[int]()
	s.Add(1)
	s.Add(2)
	elements := s.Elements()
	if len(elements) != 2 ||
		(elements[0] != 1 && elements[0] != 2) ||
		(elements[1] != 1 && elements[1] != 2) {
		t.Error("Elements should return all elements in the set")
	}
}

func TestClear(t *testing.T) {
	s := set.NewSet[int]()
	s.Add(1)
	s.Add(2)
	s.Clear()
	if !s.IsEmpty() {
		t.Error("Set should be empty after Clear")
	}
}

func TestUnion(t *testing.T) {
	s1 := set.NewSet[int]()
	s1.Add(1)
	s1.Add(2)

	s2 := set.NewSet[int]()
	s2.Add(2)
	s2.Add(3)

	s3 := set.NewSet[int]()
	s3.Add(1)
	s3.Add(2)
	s3.Add(3)

	if !s1.Union(s2).Equal(s3) {
		t.Error("Union of s1 and s2 should be s3")
	}
}

func TestEqual(t *testing.T) {
	s1 := set.NewSet[int]()
	s1.Add(1)
	s1.Add(2)

	s2 := set.NewSet[int]()
	s2.Add(1)
	s2.Add(2)

	if !s1.Equal(s2) {
		t.Error("Sets should be equal")
	}

	s3 := set.NewSet[int]()
	s3.Add(1)
	s3.Add(3)

	if s1.Equal(s3) {
		t.Error("Sets should not be equal")
	}
}

func TestIsEmpty(t *testing.T) {
	s := set.NewSet[int]()
	if !s.IsEmpty() {
		t.Error("Set should be empty")
	}

	s.Add(1)
	if s.IsEmpty() {
		t.Error("Set should not be empty")
	}
}

func TestIsSubsetOf(t *testing.T) {
	s1 := set.NewSet[int]()
	s1.Add(1)
	s1.Add(2)

	s2 := set.NewSet[int]()
	s2.Add(1)
	s2.Add(2)
	s2.Add(3)

	if !s1.IsSubsetOf(s2) {
		t.Error("s1 should be a subset of s2")
	}

	if s2.IsSubsetOf(s1) {
		t.Error("s2 should not be a subset of s1")
	}
}

func TestIsSupersetOf(t *testing.T) {
	s1 := set.NewSet[int]()
	s1.Add(1)
	s1.Add(2)

	s2 := set.NewSet[int]()
	s2.Add(1)
	s2.Add(2)
	s2.Add(3)

	if !s2.IsSupersetOf(s1) {
		t.Error("s2 should be a superset of s1")
	}

	if s1.IsSupersetOf(s2) {
		t.Error("s1 should not be a superset of s2")
	}
}

func TestIntersection(t *testing.T) {
	s1 := set.NewSet[int]()
	s1.Add(1)
	s1.Add(2)

	s2 := set.NewSet[int]()
	s2.Add(2)
	s2.Add(3)

	s3 := set.NewSet[int]()
	s3.Add(2)

	if !s1.Intersection(s2).Equal(s3) {
		t.Error("Intersection of s1 and s2 should be s3")
	}
}

func TestDifference(t *testing.T) {
	s1 := set.NewSet[int]()
	s1.Add(1)
	s1.Add(2)

	s2 := set.NewSet[int]()
	s2.Add(2)
	s2.Add(3)

	s3 := set.NewSet[int]()
	s3.Add(1)

	if !s1.Difference(s2).Equal(s3) {
		t.Error("Difference of s1 and s2 should be s3")
	}
}

func TestSymmetricDifference(t *testing.T) {
	s1 := set.NewSet[int]()
	s1.Add(1)
	s1.Add(2)

	s2 := set.NewSet[int]()
	s2.Add(2)
	s2.Add(3)

	s3 := set.NewSet[int]()
	s3.Add(1)
	s3.Add(3)

	if !s1.SymmetricDifference(s2).Equal(s3) {
		t.Error("Symmetric difference of s1 and s2 should be s3")
	}
}
