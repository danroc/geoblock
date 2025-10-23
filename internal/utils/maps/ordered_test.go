package maps_test

import (
	"testing"

	"github.com/danroc/geoblock/internal/utils/maps"
)

func TestOrderedMap_SetAndGet(t *testing.T) {
	m := maps.NewOrdered[string, int]()

	m.Set("a", 1)
	m.Set("b", 2)

	if got, ok := m.Get("a"); !ok || got != 1 {
		t.Errorf("Get(a) = %v, %v; want 1, true", got, ok)
	}
	if got, ok := m.Get("b"); !ok || got != 2 {
		t.Errorf("Get(b) = %v, %v; want 2, true", got, ok)
	}

	// Nonexistent key
	if _, ok := m.Get("x"); ok {
		t.Errorf("Get(x) = _, true; want false")
	}
}

func TestOrderedMap_OverwritePreservesOrder(t *testing.T) {
	m := maps.NewOrdered[string, int]()

	m.Set("a", 1)
	m.Set("b", 2)
	m.Set("a", 3) // overwrite existing key

	keys := m.Keys()
	wantKeys := []string{"a", "b"}
	if len(keys) != len(wantKeys) {
		t.Fatalf("Keys length = %d, want %d", len(keys), len(wantKeys))
	}
	for i, k := range wantKeys {
		if keys[i] != k {
			t.Errorf("Keys[%d] = %q, want %q", i, keys[i], k)
		}
	}

	if got, _ := m.Get("a"); got != 3 {
		t.Errorf("Get(a) = %d, want 3", got)
	}
}

func TestOrderedMap_LenAndEmpty(t *testing.T) {
	m := maps.NewOrdered[string, int]()
	if !m.Empty() {
		t.Errorf("Empty() = false; want true")
	}
	if m.Len() != 0 {
		t.Errorf("Len() = %d; want 0", m.Len())
	}

	m.Set("a", 1)
	m.Set("b", 2)
	if m.Empty() {
		t.Errorf("Empty() = true; want false")
	}
	if m.Len() != 2 {
		t.Errorf("Len() = %d; want 2", m.Len())
	}
}

func TestOrderedMap_Has(t *testing.T) {
	m := maps.NewOrdered[string, int]()
	m.Set("foo", 42)

	if !m.Has("foo") {
		t.Errorf("Has(foo) = false; want true")
	}
	if m.Has("bar") {
		t.Errorf("Has(bar) = true; want false")
	}
}

func TestOrderedMap_KeysAndValues(t *testing.T) {
	m := maps.NewOrdered[string, int]()
	m.Set("x", 10)
	m.Set("y", 20)
	m.Set("z", 30)

	wantKeys := []string{"x", "y", "z"}
	gotKeys := m.Keys()
	for i := range wantKeys {
		if gotKeys[i] != wantKeys[i] {
			t.Errorf("Keys[%d] = %q, want %q", i, gotKeys[i], wantKeys[i])
		}
	}

	wantVals := []int{10, 20, 30}
	gotVals := m.Values()
	for i := range wantVals {
		if gotVals[i] != wantVals[i] {
			t.Errorf("Values[%d] = %d, want %d", i, gotVals[i], wantVals[i])
		}
	}
}

func TestOrderedMap_RangeStopsEarly(t *testing.T) {
	m := maps.NewOrdered[int, string]()
	m.Set(1, "a")
	m.Set(2, "b")
	m.Set(3, "c")

	collected := []int{}
	m.Range(func(k int, _ string) bool {
		collected = append(collected, k)
		return k != 2 // Stop at 2
	})

	want := []int{1, 2}
	if len(collected) != len(want) {
		t.Fatalf(
			"Range collected %d items, want %d",
			len(collected),
			len(want),
		)
	}
	for i := range want {
		if collected[i] != want[i] {
			t.Errorf("Range[%d] = %v, want %v", i, collected[i], want[i])
		}
	}
}

func TestOrderedMap_KeysAndValuesAreCopies(t *testing.T) {
	m := maps.NewOrdered[string, int]()
	m.Set("a", 1)
	keys := m.Keys()
	values := m.Values()

	keys[0] = "mutated"
	values[0] = 999

	// Original map should be unaffected
	if k := m.Keys()[0]; k != "a" {
		t.Errorf("Keys modified external slice: got %q, want %q", k, "a")
	}
	if v := m.Values()[0]; v != 1 {
		t.Errorf("Values modified external slice: got %d, want %d", v, 1)
	}
}

func BenchmarkOrderedMap_Set(b *testing.B) {
	m := maps.NewOrdered[int, int]()

	for i := 0; b.Loop(); i++ {
		m.Set(i%10_000, i) // Overwrite some keys to simulate reuse
	}
}

func BenchmarkOrderedMap_Get(b *testing.B) {
	m := maps.NewOrdered[int, int]()
	for i := range 10_000 {
		m.Set(i, i)
	}

	for i := 0; b.Loop(); i++ {
		_, _ = m.Get(i % 10_000)
	}
}

func BenchmarkOrderedMap_Range(b *testing.B) {
	m := maps.NewOrdered[int, int]()
	for i := range 10_000 {
		m.Set(i, i)
	}

	for b.Loop() {
		m.Range(func(_, _ int) bool {
			return true
		})
	}
}

func BenchmarkOrderedMap_Keys(b *testing.B) {
	m := maps.NewOrdered[int, int]()
	for i := range 10_000 {
		m.Set(i, i)
	}

	for b.Loop() {
		_ = m.Keys()
	}
}

func BenchmarkOrderedMap_Values(b *testing.B) {
	m := maps.NewOrdered[int, int]()
	for i := range 10_000 {
		m.Set(i, i)
	}

	for b.Loop() {
		_ = m.Values()
	}
}
