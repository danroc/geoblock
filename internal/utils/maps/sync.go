package maps

import (
	"sync"
)

// SyncMap is a thread-safe generic map implementation using RWMutex. It
// provides concurrent-safe access to a map with type-safe keys and values.
type SyncMap[K comparable, V any] struct {
	m  map[K]V
	mu sync.RWMutex
}

// NewSyncMap creates and returns a new empty SyncMap.
func NewSyncMap[K comparable, V any]() *SyncMap[K, V] {
	return &SyncMap[K, V]{
		m: make(map[K]V),
	}
}

// Get retrieves the value associated with the given key.
//
// Returns the value and true if the key exists, or the zero value and false if
// the key does not exist.
func (sm *SyncMap[K, V]) Get(key K) (V, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	val, ok := sm.m[key]
	return val, ok
}

// Has reports whether the key exists in the map.
func (sm *SyncMap[K, V]) Has(key K) bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	_, ok := sm.m[key]
	return ok
}

// Set associates the given value with the given key. If the key already
// exists, its value is replaced.
func (sm *SyncMap[K, V]) Set(key K, value V) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.m[key] = value
}

// GetOrSet returns the existing value for the key if present. Otherwise, it
// stores and returns the given value.
//
// The boolean return value indicates whether the key already existed (true) or
// was newly set (false).
func (sm *SyncMap[K, V]) GetOrSet(key K, value V) (V, bool) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if v, ok := sm.m[key]; ok {
		return v, true
	}

	sm.m[key] = value
	return value, false
}

// Delete removes the value associated with the given key.
func (sm *SyncMap[K, V]) Delete(key K) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	delete(sm.m, key)
}

// Range iterates over all key-value pairs in the map, calling the provided
// function for each pair.
//
// If the function returns false, the iteration stops.
func (sm *SyncMap[K, V]) Range(f func(key K, value V) bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	for k, v := range sm.m {
		if !f(k, v) {
			break
		}
	}
}
