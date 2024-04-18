package set

type Set[E comparable] map[E]bool

func NewSet[E comparable]() Set[E] {
	return make(Set[E])
}

func (s Set[E]) Add(value E) {
	s[value] = true
}

func (s Set[E]) Contains(value E) bool {
	return s[value]
}

func (s Set[E]) Remove(value E) {
	delete(s, value)
}

func (s Set[E]) Cardinality() int {
	return len(s)
}

// Values returns a slice of all elements in the set.
func (s Set[E]) Values() []E {
	values := make([]E, 0, len(s))
	for k := range s {
		values = append(values, k)
	}
	return values
}

// Clear removes all elements from the set.
func (s Set[E]) Clear() {
	for k := range s {
		delete(s, k)
	}
}

// Union returns a new set with elements that are in either `s` or `other`.
func (s Set[E]) Union(other Set[E]) Set[E] {
	result := NewSet[E]()
	for k := range s {
		result.Add(k)
	}
	for k := range other {
		result.Add(k)
	}
	return result
}

// Intersection returns a new set with elements that are in both `s` and
// `other`.
func (s Set[E]) Intersection(other Set[E]) Set[E] {
	result := NewSet[E]()
	for k := range s {
		if other.Contains(k) {
			result.Add(k)
		}
	}
	return result
}

// Difference returns a new set with elements that are in `s` but not in
// `other`.
func (s Set[E]) Difference(other Set[E]) Set[E] {
	result := NewSet[E]()
	for k := range s {
		if !other.Contains(k) {
			result.Add(k)
		}
	}
	return result
}

func (s Set[E]) SymmetricDifference(other Set[E]) Set[E] {
	result := NewSet[E]()
	for k := range s {
		if !other.Contains(k) {
			result.Add(k)
		}
	}
	for k := range other {
		if !s.Contains(k) {
			result.Add(k)
		}
	}
	return result
}

func (s Set[E]) IsSubsetOf(other Set[E]) bool {
	for k := range s {
		if !other.Contains(k) {
			return false
		}
	}
	return true
}

func (s Set[E]) IsSupersetOf(other Set[E]) bool {
	return other.IsSubsetOf(s)
}
