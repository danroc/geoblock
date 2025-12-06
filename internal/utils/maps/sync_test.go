package maps_test

import (
	"sync"
	"testing"

	"github.com/danroc/geoblock/internal/utils/maps"
)

func TestNewSyncMap(t *testing.T) {
	t.Parallel()

	m := maps.NewSyncMap[string, int]()

	if m == nil {
		t.Fatal("NewSyncMap() returned nil")
	}

	// Verify map is empty
	if _, ok := m.Get("key"); ok {
		t.Error("Expected new map to be empty")
	}
}

func TestSyncMapGet(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		setup     func(*maps.SyncMap[string, int])
		key       string
		wantValue int
		wantOk    bool
	}{
		{
			name:      "non-existent key",
			setup:     func(m *maps.SyncMap[string, int]) {},
			key:       "missing",
			wantValue: 0,
			wantOk:    false,
		},
		{
			name: "existing key",
			setup: func(m *maps.SyncMap[string, int]) {
				m.Set("key1", 42)
			},
			key:       "key1",
			wantValue: 42,
			wantOk:    true,
		},
		{
			name: "zero value",
			setup: func(m *maps.SyncMap[string, int]) {
				m.Set("zero", 0)
			},
			key:       "zero",
			wantValue: 0,
			wantOk:    true,
		},
		{
			name: "multiple keys",
			setup: func(m *maps.SyncMap[string, int]) {
				m.Set("a", 1)
				m.Set("b", 2)
				m.Set("c", 3)
			},
			key:       "b",
			wantValue: 2,
			wantOk:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := maps.NewSyncMap[string, int]()
			tt.setup(m)

			gotValue, gotOk := m.Get(tt.key)

			if gotValue != tt.wantValue {
				t.Errorf("Get() value = %v, want %v", gotValue, tt.wantValue)
			}
			if gotOk != tt.wantOk {
				t.Errorf("Get() ok = %v, want %v", gotOk, tt.wantOk)
			}
		})
	}
}

func TestSyncMapHas(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		setup func(*maps.SyncMap[string, string])
		key   string
		want  bool
	}{
		{
			name:  "non-existent key",
			setup: func(m *maps.SyncMap[string, string]) {},
			key:   "missing",
			want:  false,
		},
		{
			name: "existing key",
			setup: func(m *maps.SyncMap[string, string]) {
				m.Set("key1", "value1")
			},
			key:  "key1",
			want: true,
		},
		{
			name: "empty string value",
			setup: func(m *maps.SyncMap[string, string]) {
				m.Set("empty", "")
			},
			key:  "empty",
			want: true,
		},
		{
			name: "wrong key after set",
			setup: func(m *maps.SyncMap[string, string]) {
				m.Set("key1", "value1")
			},
			key:  "key2",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := maps.NewSyncMap[string, string]()
			tt.setup(m)

			got := m.Has(tt.key)

			if got != tt.want {
				t.Errorf("Has() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSyncMapSet(t *testing.T) {
	t.Parallel()

	m := maps.NewSyncMap[string, int]()

	// Set initial value
	m.Set("key1", 10)

	if val, ok := m.Get("key1"); !ok || val != 10 {
		t.Errorf("Expected key1=10, got %d, ok=%v", val, ok)
	}

	// Overwrite value
	m.Set("key1", 20)

	if val, ok := m.Get("key1"); !ok || val != 20 {
		t.Errorf("Expected key1=20 after overwrite, got %d, ok=%v", val, ok)
	}

	// Set multiple keys
	m.Set("key2", 30)
	m.Set("key3", 40)

	if val, ok := m.Get("key2"); !ok || val != 30 {
		t.Errorf("Expected key2=30, got %d, ok=%v", val, ok)
	}
	if val, ok := m.Get("key3"); !ok || val != 40 {
		t.Errorf("Expected key3=40, got %d, ok=%v", val, ok)
	}
}

func TestSyncMapGetOrSet(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		setup       func(*maps.SyncMap[string, int])
		key         string
		setValue    int
		wantValue   int
		wantExisted bool
	}{
		{
			name:        "new key",
			setup:       func(m *maps.SyncMap[string, int]) {},
			key:         "new",
			setValue:    100,
			wantValue:   100,
			wantExisted: false,
		},
		{
			name: "existing key",
			setup: func(m *maps.SyncMap[string, int]) {
				m.Set("existing", 50)
			},
			key:         "existing",
			setValue:    100,
			wantValue:   50,
			wantExisted: true,
		},
		{
			name: "existing key with zero value",
			setup: func(m *maps.SyncMap[string, int]) {
				m.Set("zero", 0)
			},
			key:         "zero",
			setValue:    100,
			wantValue:   0,
			wantExisted: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := maps.NewSyncMap[string, int]()
			tt.setup(m)

			gotValue, gotExisted := m.GetOrSet(tt.key, tt.setValue)

			if gotValue != tt.wantValue {
				t.Errorf(
					"GetOrSet() value = %v, want %v",
					gotValue,
					tt.wantValue,
				)
			}
			if gotExisted != tt.wantExisted {
				t.Errorf(
					"GetOrSet() existed = %v, want %v",
					gotExisted,
					tt.wantExisted,
				)
			}

			// Verify value is actually in map
			storedValue, ok := m.Get(tt.key)
			if !ok {
				t.Errorf(
					"Expected key %q to exist in map after GetOrSet",
					tt.key,
				)
			}
			if storedValue != tt.wantValue {
				t.Errorf(
					"Expected stored value %d, got %d",
					tt.wantValue,
					storedValue,
				)
			}
		})
	}
}

func TestSyncMapConcurrentGet(t *testing.T) {
	t.Parallel()

	m := maps.NewSyncMap[int, int]()

	// Pre-populate map
	for i := 0; i < 100; i++ {
		m.Set(i, i*10)
	}

	const numGoroutines = 50
	const readsPerGoroutine = 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	errors := make(chan error, numGoroutines*readsPerGoroutine)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < readsPerGoroutine; j++ {
				key := j % 100
				val, ok := m.Get(key)
				if !ok {
					errors <- nil
					return
				}
				if val != key*10 {
					errors <- nil
					return
				}
			}
		}()
	}

	wg.Wait()
	close(errors)

	if len(errors) > 0 {
		t.Error("Concurrent Get operations failed")
	}
}

func TestSyncMapConcurrentSet(t *testing.T) {
	t.Parallel()

	m := maps.NewSyncMap[int, int]()

	const numGoroutines = 50
	const setsPerGoroutine = 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		goroutineID := i
		go func() {
			defer wg.Done()
			for j := 0; j < setsPerGoroutine; j++ {
				key := goroutineID*setsPerGoroutine + j
				m.Set(key, key*2)
			}
		}()
	}

	wg.Wait()

	// Verify all values were set correctly
	for i := 0; i < numGoroutines*setsPerGoroutine; i++ {
		val, ok := m.Get(i)
		if !ok {
			t.Errorf("Expected key %d to exist", i)
		}
		if val != i*2 {
			t.Errorf("Expected value %d, got %d", i*2, val)
		}
	}
}

func TestSyncMapConcurrentMixed(t *testing.T) {
	t.Parallel()

	m := maps.NewSyncMap[string, int]()

	const numGoroutines = 30
	const opsPerGoroutine = 100

	var wg sync.WaitGroup

	// Writers
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		goroutineID := i
		go func() {
			defer wg.Done()
			for j := 0; j < opsPerGoroutine; j++ {
				key := string(rune('a' + (goroutineID % 26)))
				m.Set(key, goroutineID*j)
			}
		}()
	}

	// Readers
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < opsPerGoroutine; j++ {
				key := string(rune('a' + (j % 26)))
				m.Get(key)
			}
		}()
	}

	// Has checkers
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < opsPerGoroutine; j++ {
				key := string(rune('a' + (j % 26)))
				m.Has(key)
			}
		}()
	}

	// GetOrSet operations
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		goroutineID := i
		go func() {
			defer wg.Done()
			for j := 0; j < opsPerGoroutine; j++ {
				key := string(rune('a' + (j % 26)))
				m.GetOrSet(key, goroutineID*j)
			}
		}()
	}

	wg.Wait()

	// Verify map has entries (exact values are non-deterministic)
	hasAnyKey := false
	for i := 0; i < 26; i++ {
		key := string(rune('a' + i))
		if m.Has(key) {
			hasAnyKey = true
			break
		}
	}
	if !hasAnyKey {
		t.Error("Expected map to have at least one key after concurrent ops")
	}
}

func TestSyncMapDifferentTypes(t *testing.T) {
	t.Parallel()

	t.Run("string to struct", func(t *testing.T) {
		type Person struct {
			Name string
			Age  int
		}

		m := maps.NewSyncMap[string, Person]()

		p1 := Person{Name: "Alice", Age: 30}
		m.Set("alice", p1)

		got, ok := m.Get("alice")
		if !ok {
			t.Fatal("Expected key 'alice' to exist")
		}
		if got != p1 {
			t.Errorf("Expected %+v, got %+v", p1, got)
		}
	})

	t.Run("int to string", func(t *testing.T) {
		m := maps.NewSyncMap[int, string]()

		m.Set(1, "one")
		m.Set(2, "two")

		if val, ok := m.Get(1); !ok || val != "one" {
			t.Errorf("Expected 'one', got %q, ok=%v", val, ok)
		}
		if val, ok := m.Get(2); !ok || val != "two" {
			t.Errorf("Expected 'two', got %q, ok=%v", val, ok)
		}
	})

	t.Run("string to pointer", func(t *testing.T) {
		m := maps.NewSyncMap[string, *int]()

		val1 := 42
		m.Set("key1", &val1)

		got, ok := m.Get("key1")
		if !ok {
			t.Fatal("Expected key 'key1' to exist")
		}
		if *got != 42 {
			t.Errorf("Expected 42, got %d", *got)
		}
	})
}

func BenchmarkSyncMapGet(b *testing.B) {
	m := maps.NewSyncMap[int, int]()
	m.Set(1, 100)

	for b.Loop() {
		m.Get(1)
	}
}

func BenchmarkSyncMapSet(b *testing.B) {
	m := maps.NewSyncMap[int, int]()

	for i := 0; b.Loop(); i++ {
		m.Set(i%100, i)
	}
}

func BenchmarkSyncMapHas(b *testing.B) {
	m := maps.NewSyncMap[int, int]()
	m.Set(1, 100)

	for b.Loop() {
		m.Has(1)
	}
}

func BenchmarkSyncMapGetOrSet(b *testing.B) {
	m := maps.NewSyncMap[int, int]()

	for i := 0; b.Loop(); i++ {
		m.GetOrSet(i%100, i)
	}
}

func BenchmarkSyncMapConcurrentReads(b *testing.B) {
	m := maps.NewSyncMap[int, int]()
	for i := range 100 {
		m.Set(i, i*10)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			m.Get(i % 100)
			i++
		}
	})
}

func BenchmarkSyncMapConcurrentWrites(b *testing.B) {
	m := maps.NewSyncMap[int, int]()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			m.Set(i%100, i)
			i++
		}
	})
}

func BenchmarkSyncMapConcurrentMixed(b *testing.B) {
	m := maps.NewSyncMap[int, int]()
	for i := range 100 {
		m.Set(i, i*10)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			switch i % 4 {
			case 0:
				m.Get(i % 100)
			case 1:
				m.Set(i%100, i)
			case 2:
				m.Has(i % 100)
			case 3:
				m.GetOrSet(i%100, i)
			}
			i++
		}
	})
}
