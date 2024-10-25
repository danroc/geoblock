package utils

// Any returns true if the given function returns true for any of the given
// values. If the slice is empty, it returns false.
func Any[T any](values []T, f func(T) bool) bool {
	for _, v := range values {
		if f(v) {
			return true
		}
	}
	return false
}

// All returns true if the given function returns true for all of the given
// values. If the slice is empty, it returns true.
func All[T any](values []T, f func(T) bool) bool {
	for _, v := range values {
		if !f(v) {
			return false
		}
	}
	return true
}

// None returns true if the given function returns false for all of the given
// values. If the slice is empty, it returns true.
func None[T any](values []T, f func(T) bool) bool {
	return !Any(values, f)
}
