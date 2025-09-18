package container

import (
	"fmt"
	"sort"
	"sync"
	"testing"
)

// assertEqual is a helper function to simplify checking for equality in tests.
func assertEqual[T comparable](t *testing.T, a, b T) {
	t.Helper()
	if a != b {
		t.Errorf("Expected %v, but got %v", b, a)
	}
}

// assertPanic is a helper to check if a function panics.
// (Not used in this case, but good practice to have)
func assertPanic(t *testing.T, f func()) {
	t.Helper()
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()
	f()
}

// TestNewSet tests the constructors New() and NewWithValues().
func TestNewSet(t *testing.T) {
	s := New[int]()
	if s == nil {
		t.Fatal("New() returned nil")
	}
	assertEqual(t, s.Len(), 0)

	s2 := NewWithValues(1, 2, 3, 2) // duplicates should be ignored
	assertEqual(t, s2.Len(), 3)
	if !s2.Contains(1) || !s2.Contains(2) || !s2.Contains(3) {
		t.Errorf("NewWithValues() did not initialize with correct elements")
	}
}

// TestSetBasicOperations tests Add, Remove, Contains, Len, IsEmpty, and Clear.
func TestSetBasicOperations(t *testing.T) {
	s := New[string]()

	// Test IsEmpty on new set
	assertEqual(t, s.IsEmpty(), true)

	// Test Add
	s.Add("apple")
	s.Add("banana")
	assertEqual(t, s.Len(), 2)
	assertEqual(t, s.Contains("apple"), true)
	assertEqual(t, s.IsEmpty(), false)

	// Test adding a duplicate
	s.Add("apple")
	assertEqual(t, s.Len(), 2)

	// Test Remove
	s.Remove("apple")
	assertEqual(t, s.Len(), 1)
	assertEqual(t, s.Contains("apple"), false)

	// Test removing a non-existent element
	s.Remove("grape")
	assertEqual(t, s.Len(), 1)

	// Test Clear
	s.Clear()
	assertEqual(t, s.Len(), 0)
	assertEqual(t, s.IsEmpty(), true)
	assertEqual(t, s.Contains("banana"), false)
}

// TestToSliceAndClone tests ToSlice and Clone methods.
func TestToSliceAndClone(t *testing.T) {
	s := NewWithValues("a", "b", "c")

	// Test ToSlice
	slice := s.ToSlice()
	assertEqual(t, len(slice), 3)
	// Sort to have a predictable order for comparison
	sort.Strings(slice)
	expectedSlice := []string{"a", "b", "c"}
	if fmt.Sprintf("%v", slice) != fmt.Sprintf("%v", expectedSlice) {
		t.Errorf("ToSlice() returned %v, expected %v", slice, expectedSlice)
	}

	// Test Clone
	clone := s.Clone()
	if !s.Equal(clone) {
		t.Errorf("Clone should be equal to the original set")
	}
	// Modify original, clone should not be affected
	s.Add("d")
	if s.Equal(clone) {
		t.Errorf("Clone should not be affected by changes to the original set")
	}
}

// TestSetEquality tests the Equal method.
func TestSetEquality(t *testing.T) {
	s1 := NewWithValues(1, 2, 3)
	s2 := NewWithValues(3, 2, 1)
	s3 := NewWithValues(1, 2, 4)
	s4 := NewWithValues(1, 2)

	assertEqual(t, s1.Equal(s2), true)
	assertEqual(t, s1.Equal(s3), false)
	assertEqual(t, s1.Equal(s4), false)
	assertEqual(t, s4.Equal(s1), false)
	assertEqual(t, s1.Equal(s1), true)
}

// TestSetAlgebra tests Union, Intersection, Difference, and SymmetricDifference.
func TestSetAlgebra(t *testing.T) {
	s1 := NewWithValues(1, 2, 3, 4)
	s2 := NewWithValues(3, 4, 5, 6)

	// Union
	union := s1.Union(s2)
	expectedUnion := NewWithValues(1, 2, 3, 4, 5, 6)
	if !union.Equal(expectedUnion) {
		t.Errorf("Union failed. Expected %v, got %v", expectedUnion, union)
	}

	// Intersection
	intersection := s1.Intersection(s2)
	expectedIntersection := NewWithValues(3, 4)
	if !intersection.Equal(expectedIntersection) {
		t.Errorf("Intersection failed. Expected %v, got %v", expectedIntersection, intersection)
	}

	// Difference (s1 - s2)
	difference := s1.Difference(s2)
	expectedDifference := NewWithValues(1, 2)
	if !difference.Equal(expectedDifference) {
		t.Errorf("Difference failed. Expected %v, got %v", expectedDifference, difference)
	}

	// Symmetric Difference
	symDifference := s1.SymmetricDifference(s2)
	expectedSymDifference := NewWithValues(1, 2, 5, 6)
	if !symDifference.Equal(expectedSymDifference) {
		t.Errorf("SymmetricDifference failed. Expected %v, got %v", expectedSymDifference, symDifference)
	}
}

// TestSubsets tests IsSubset and IsSuperset.
func TestSubsets(t *testing.T) {
	s1 := NewWithValues(1, 2, 3, 4)
	s2 := NewWithValues(2, 3)
	s3 := NewWithValues(2, 3, 5)

	assertEqual(t, s2.IsSubset(s1), true)
	assertEqual(t, s1.IsSuperset(s2), true)

	assertEqual(t, s3.IsSubset(s1), false)
	assertEqual(t, s1.IsSuperset(s3), false)

	assertEqual(t, s1.IsSubset(s1), true) // A set is a subset of itself
	assertEqual(t, s1.IsSuperset(s1), true)
}

// TestEach tests the Each method for iteration.
func TestEach(t *testing.T) {
	s := NewWithValues(10, 20, 30)
	sum := 0
	s.Each(func(item int) bool {
		sum += item
		return true // continue iteration
	})
	assertEqual(t, sum, 60)

	// Test early exit
	count := 0
	s.Each(func(item int) bool {
		count++
		return item != 20 // stop when item is 20
	})
	if count > 2 { // Order is not guaranteed, but it should stop early
		t.Errorf("Each did not stop iteration early. Count: %d", count)
	}
}

// TestSetConcurrency is crucial for a thread-safe set.
// It runs multiple goroutines to add and remove items concurrently.
// Run with `go test -race` to detect race conditions.
func TestSetConcurrency(t *testing.T) {
	s := New[int]()
	var wg sync.WaitGroup
	numGoroutines := 100
	itemsPerGoroutine := 100

	// Concurrent additions
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(start int) {
			defer wg.Done()
			for j := 0; j < itemsPerGoroutine; j++ {
				s.Add(start + j)
			}
		}(i * itemsPerGoroutine)
	}
	wg.Wait()

	expectedLen := numGoroutines * itemsPerGoroutine
	assertEqual(t, s.Len(), expectedLen)

	// Concurrent removals
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(start int) {
			defer wg.Done()
			for j := 0; j < itemsPerGoroutine; j++ {
				s.Remove(start + j)
			}
		}(i * itemsPerGoroutine)
	}
	wg.Wait()

	assertEqual(t, s.Len(), 0)
}

// --- Benchmarks ---

// createSetWithNItems is a helper for benchmark setup.
func createSetWithNItems(n int) *Set[int] {
	s := New[int]()
	for i := 0; i < n; i++ {
		s.Add(i)
	}
	return s
}

func BenchmarkAdd(b *testing.B) {
	s := New[int]()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Add(i)
	}
}

func BenchmarkContainsHit(b *testing.B) {
	s := createSetWithNItems(10000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Contains(5000) // Element that is always present
	}
}

func BenchmarkContainsMiss(b *testing.B) {
	s := createSetWithNItems(10000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Contains(10001) // Element that is never present
	}
}

func BenchmarkRemove(b *testing.B) {
	// We need to re-populate the set on each run of the benchmark loop
	// to ensure we are always removing from a set of a similar size.
	b.StopTimer()
	s := createSetWithNItems(b.N)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		s.Remove(i)
	}
}

func BenchmarkUnion(b *testing.B) {
	s1 := createSetWithNItems(1000)
	s2 := createSetWithNItems(1000)
	// Offset s2's items so there's some overlap and some difference
	for i := 500; i < 1500; i++ {
		s2.Add(i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s1.Union(s2)
	}
}

func BenchmarkIntersection(b *testing.B) {
	s1 := createSetWithNItems(1000)
	s2 := createSetWithNItems(1000)
	// Offset s2's items
	for i := 500; i < 1500; i++ {
		s2.Add(i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s1.Intersection(s2)
	}
}

func BenchmarkDifference(b *testing.B) {
	s1 := createSetWithNItems(1000)
	s2 := createSetWithNItems(1000)
	for i := 500; i < 1500; i++ {
		s2.Add(i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s1.Difference(s2)
	}
}

func BenchmarkSymmetricDifference(b *testing.B) {
	s1 := createSetWithNItems(1000)
	s2 := createSetWithNItems(1000)
	for i := 500; i < 1500; i++ {
		s2.Add(i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s1.SymmetricDifference(s2)
	}
}
