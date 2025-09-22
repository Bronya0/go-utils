package container

import (
	"sort"
	"strconv"
	"sync"
	"testing"
)

// --- 单元测试 (Unit Tests) ---

func TestNewSet(t *testing.T) {
	s := NewSet(1, 2, 3)
	if s.concurrent {
		t.Error("NewSet should create a non-concurrent set")
	}
	if s.Len() != 3 {
		t.Errorf("Expected length 3, got %d", s.Len())
	}
	if !s.Contains(2) {
		t.Error("Set should contain 2")
	}
}

func TestNewConcurrentSet(t *testing.T) {
	s := NewConcurrentSet("a", "b")
	if !s.concurrent {
		t.Error("NewConcurrentSet should create a concurrent set")
	}
	if s.Len() != 2 {
		t.Errorf("Expected length 2, got %d", s.Len())
	}
	if !s.Contains("b") {
		t.Error("Set should contain 'b'")
	}
}

func TestAdd(t *testing.T) {
	s := NewSet[int]()
	s.Add(1)
	if !s.Contains(1) {
		t.Error("Failed to add single element")
	}
	s.Add(2, 3, 2) // Add multiple, including a duplicate
	if s.Len() != 3 {
		t.Errorf("Expected length 3, got %d", s.Len())
	}
	if !s.Contains(3) {
		t.Error("Failed to add multiple elements")
	}
}

func TestRemove(t *testing.T) {
	s := NewSet(1, 2, 3, 4)
	s.Remove(2)
	if s.Contains(2) {
		t.Error("Failed to remove single element")
	}
	s.Remove(3, 1, 5) // Remove multiple, including a non-existent element
	if s.Len() != 1 {
		t.Errorf("Expected length 1 after removal, got %d", s.Len())
	}
	if s.Contains(1) || s.Contains(3) {
		t.Error("Failed to remove multiple elements")
	}
	if !s.Contains(4) {
		t.Error("Element 4 should still be in the set")
	}
}

func TestContains(t *testing.T) {
	s := NewSet("apple", "banana")
	if !s.Contains("apple") {
		t.Error("'apple' should be in the set")
	}
	if s.Contains("cherry") {
		t.Error("'cherry' should not be in the set")
	}
}

func TestLen(t *testing.T) {
	s := NewSet[int]()
	if s.Len() != 0 {
		t.Errorf("Expected length 0 for new set, got %d", s.Len())
	}
	s.Add(1, 2)
	if s.Len() != 2 {
		t.Errorf("Expected length 2, got %d", s.Len())
	}
}

func TestIsEmpty(t *testing.T) {
	s := NewSet[int]()
	if !s.IsEmpty() {
		t.Error("New set should be empty")
	}
	s.Add(1)
	if s.IsEmpty() {
		t.Error("Set with one element should not be empty")
	}
	s.Remove(1)
	if !s.IsEmpty() {
		t.Error("Set should be empty after removing its only element")
	}
}

func TestClear(t *testing.T) {
	s := NewSet(1, 2, 3)
	s.Clear()
	if !s.IsEmpty() {
		t.Error("Set should be empty after Clear()")
	}
	if s.Len() != 0 {
		t.Error("Set length should be 0 after Clear()")
	}
}

func TestToSlice(t *testing.T) {
	s := NewSet("c", "a", "b")
	slice := s.ToSlice()
	if len(slice) != 3 {
		t.Fatalf("Expected slice length 3, got %d", len(slice))
	}

	// Sort to ensure consistent order for comparison
	sort.Strings(slice)
	expected := []string{"a", "b", "c"}
	for i := range slice {
		if slice[i] != expected[i] {
			t.Errorf("Expected slice %v, got %v", expected, slice)
			break
		}
	}
}

func TestClone(t *testing.T) {
	s1 := NewConcurrentSet(1, 2)
	s2 := s1.Clone()

	if !s2.concurrent {
		t.Error("Cloned set should be concurrent")
	}
	if !s1.Equal(s2) {
		t.Error("Cloned set should be equal to the original")
	}
	s2.Add(3)
	if s1.Contains(3) {
		t.Error("Changes to cloned set should not affect the original")
	}
}

func TestString(t *testing.T) {
	s := NewSet(3, 1, 2)
	expected := "Set{1, 2, 3}"
	if s.String() != expected {
		t.Errorf("Expected string '%s', got '%s'", expected, s.String())
	}
}

func TestEach(t *testing.T) {
	s := NewConcurrentSet(1, 2, 3)
	sum := 0
	s.Each(func(item int) bool {
		sum += item
		return true
	})
	if sum != 6 {
		t.Errorf("Expected sum 6 from Each, got %d", sum)
	}

	// Test early exit
	count := 0
	s.Each(func(item int) bool {
		count++
		return item != 2 // Stop when item is 2
	})
	if count > 2 {
		t.Errorf("Each did not exit early, count was %d", count)
	}
}

func TestEqual(t *testing.T) {
	s1 := NewSet(1, 2)
	s2 := NewSet(2, 1)
	s3 := NewConcurrentSet(1, 2)
	s4 := NewSet(1, 2, 3)

	if !s1.Equal(s2) {
		t.Error("s1 and s2 should be equal")
	}
	if !s1.Equal(s3) {
		t.Error("s1 and s3 should be equal (concurrent vs non-concurrent)")
	}
	if s1.Equal(s4) {
		t.Error("s1 and s4 should not be equal")
	}
}

func TestSetOperations(t *testing.T) {
	s1 := NewSet(1, 2, 3, 4)
	s2 := NewConcurrentSet(3, 4, 5, 6)

	// Union
	union := s1.Union(s2)
	assertSetElements(t, union, "Union", 1, 2, 3, 4, 5, 6)
	if !union.concurrent {
		t.Error("Union of a concurrent set should be concurrent")
	}

	// Intersection
	intersection := s1.Intersection(s2)
	assertSetElements(t, intersection, "Intersection", 3, 4)
	if !intersection.concurrent {
		t.Error("Intersection of a concurrent set should be concurrent")
	}

	// Difference
	diff := s1.Difference(s2)
	assertSetElements(t, diff, "Difference", 1, 2)

	// Symmetric Difference
	symDiff := s1.SymmetricDifference(s2)
	assertSetElements(t, symDiff, "SymmetricDifference", 1, 2, 5, 6)
}

func TestSubsets(t *testing.T) {
	superset := NewConcurrentSet(1, 2, 3, 4)
	subset := NewSet(2, 3)
	disjoint := NewSet(5, 6)

	if !subset.IsSubset(superset) {
		t.Error("subset should be a subset of superset")
	}
	if superset.IsSubset(subset) {
		t.Error("superset should not be a subset of subset")
	}
	if !superset.IsSuperset(subset) {
		t.Error("superset should be a superset of subset")
	}
	if subset.IsSuperset(superset) {
		t.Error("subset should not be a superset of superset")
	}
	if disjoint.IsSubset(superset) {
		t.Error("disjoint set should not be a subset")
	}
}

// --- 并发安全测试 (Concurrency Tests) ---

// TestConcurrentAddRemoveContains tests basic thread safety of Add, Remove, and Contains.
func TestConcurrentAddRemoveContains(t *testing.T) {
	s := NewConcurrentSet[int]()
	var wg sync.WaitGroup
	numRoutines := 50
	itemsPerRoutine := 100

	// Concurrent Add
	for i := 0; i < numRoutines; i++ {
		wg.Add(1)
		go func(start int) {
			defer wg.Done()
			for j := 0; j < itemsPerRoutine; j++ {
				s.Add(start + j)
			}
		}(i * itemsPerRoutine)
	}
	wg.Wait()

	expectedLen := numRoutines * itemsPerRoutine
	if s.Len() != expectedLen {
		t.Fatalf("Expected length %d after concurrent Add, got %d", expectedLen, s.Len())
	}

	// Concurrent Contains
	for i := 0; i < expectedLen; i++ {
		wg.Add(1)
		go func(val int) {
			defer wg.Done()
			if !s.Contains(val) {
				t.Errorf("Set should contain %d after concurrent adds", val)
			}
		}(i)
	}
	wg.Wait()

	// Concurrent Remove
	for i := 0; i < numRoutines; i++ {
		wg.Add(1)
		go func(start int) {
			defer wg.Done()
			for j := 0; j < itemsPerRoutine; j++ {
				s.Remove(start + j)
			}
		}(i * itemsPerRoutine)
	}
	wg.Wait()

	if !s.IsEmpty() {
		t.Errorf("Set should be empty after concurrent Remove, but length is %d", s.Len())
	}
}

// TestConcurrentSetOperations_NoDeadlock verifies that the deadlock fix works.
func TestConcurrentSetOperations_NoDeadlock(t *testing.T) {
	s1 := NewConcurrentSet(1, 2, 3)
	s2 := NewConcurrentSet(3, 4, 5)

	var wg sync.WaitGroup

	// Test multiple operations that lock two sets
	operations := map[string]func(){
		"Union":               func() { s1.Union(s2); s2.Union(s1) },
		"Intersection":        func() { s1.Intersection(s2); s2.Intersection(s1) },
		"Difference":          func() { s1.Difference(s2); s2.Difference(s1) },
		"SymmetricDifference": func() { s1.SymmetricDifference(s2); s2.SymmetricDifference(s1) },
		"Equal":               func() { s1.Equal(s2); s2.Equal(s1) },
		"IsSubset":            func() { s1.IsSubset(s2); s2.IsSubset(s1) },
		"IsSuperset":          func() { s1.IsSuperset(s2); s2.IsSuperset(s1) },
	}

	for name, op := range operations {
		t.Run(name, func(t *testing.T) {
			done := make(chan bool)
			wg.Add(2)

			// Run the operations in opposite order in two goroutines
			go func() {
				defer wg.Done()
				op()
			}()
			go func() {
				defer wg.Done()
				op()
			}()

			go func() {
				wg.Wait()
				close(done)
			}()

			// This test will time out if a deadlock occurs
			select {
			case <-done:
				// Success, test finished
			}
		})
	}
}

// --- 基准测试 (Benchmarks) ---

var (
	numBenchmarkItems = 1000
	nonConcurrentSet  *Set[string]
	concurrentSet     *Set[string]
	otherSet          *Set[string] // For binary operations
)

// go测试文件（_test.go）中的 init 函数会不会随着整个项目的启动而运行
func init() {
	nonConcurrentSet = NewSet[string]()
	concurrentSet = NewConcurrentSet[string]()
	otherSet = NewSet[string]()
	for i := 0; i < numBenchmarkItems; i++ {
		item := "item" + strconv.Itoa(i)
		nonConcurrentSet.Add(item)
		concurrentSet.Add(item)
		if i%2 == 0 { // Let otherSet have half the items
			otherSet.Add(item)
		}
	}
}

// Add
func BenchmarkSet_Add(b *testing.B) {
	b.ReportAllocs() // 显式报告内存分配

	s := NewSet[int]()
	for i := 0; i < b.N; i++ {
		s.Add(i)
	}
}

func BenchmarkConcurrentSet_Add(b *testing.B) {
	b.ReportAllocs() // 显式报告内存分配

	s := NewConcurrentSet[int]()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			s.Add(i)
			i++
		}
	})
}

// Contains
func BenchmarkSet_Contains(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		nonConcurrentSet.Contains("item" + strconv.Itoa(i%numBenchmarkItems))
	}
}

func BenchmarkConcurrentSet_Contains(b *testing.B) {
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			concurrentSet.Contains("item" + strconv.Itoa(i%numBenchmarkItems))
			i++
		}
	})
}

// Remove
func BenchmarkSet_Remove(b *testing.B) {
	b.ReportAllocs()
	s := NewSet[string]()
	for i := 0; i < b.N; i++ {
		item := "item" + strconv.Itoa(i)
		s.Add(item)
		s.Remove(item)
	}
}

func BenchmarkConcurrentSet_Remove(b *testing.B) {
	b.ReportAllocs()
	s := NewConcurrentSet[string]()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			item := "item" + strconv.Itoa(i)
			s.Add(item)
			s.Remove(item)
			i++
		}
	})
}

// Union
func BenchmarkSet_Union(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		nonConcurrentSet.Union(otherSet)
	}
}

func BenchmarkConcurrentSet_Union(b *testing.B) {
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			concurrentSet.Union(otherSet)
		}
	})
}

// --- Helper Functions ---

// assertSetElements checks if a set contains exactly the given elements.
func assertSetElements[T comparable](t *testing.T, s *Set[T], opName string, elements ...T) {
	t.Helper()
	if s.Len() != len(elements) {
		t.Errorf("%s: Expected length %d, got %d. Set: %s", opName, len(elements), s.Len(), s.String())
		return
	}
	for _, el := range elements {
		if !s.Contains(el) {
			t.Errorf("%s: Set should contain element '%v', but it doesn't. Set: %s", opName, el, s.String())
		}
	}
}
