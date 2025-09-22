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

// TestNewSet tests the constructors NewSet() and NewSetWithValues().
func TestNewSet(t *testing.T) {
	s := NewSet[int]()
	if s == nil {
		t.Fatal("NewSet() returned nil")
	}
	assertEqual(t, s.Len(), 0)

	s2 := NewSet(1, 2, 3, 2) // duplicates should be ignored
	assertEqual(t, s2.Len(), 3)
	if !s2.Contains(1) || !s2.Contains(2) || !s2.Contains(3) {
		t.Errorf("NewSetWithValues() did not initialize with correct elements")
	}
}

// TestNewConcurrentSet tests the constructors TestNewConcurrentSet()
func TestNewConcurrentSet(t *testing.T) {
	s := NewSet[int]()
	if s == nil {
		t.Fatal("NewSet() returned nil")
	}
	assertEqual(t, s.Len(), 0)

	s2 := NewConcurrentSet(1, 2, 3, 2) // duplicates should be ignored
	assertEqual(t, s2.Len(), 3)
	if !s2.Contains(1) || !s2.Contains(2) || !s2.Contains(3) {
		t.Errorf("NewSetWithValues() did not initialize with correct elements")
	}
}

// TestSetBasicOperations tests Add, Remove, Contains, Len, IsEmpty, and Clear.
func TestSetBasicOperations(t *testing.T) {
	s := NewSet[string]()

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
	s := NewSet("a", "b", "c")
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
	s1 := NewSet(1, 2, 3)
	s2 := NewSet(3, 2, 1)
	s3 := NewSet(1, 2, 4)
	s4 := NewSet(1, 2)

	assertEqual(t, s1.Equal(s2), true)
	assertEqual(t, s1.Equal(s3), false)
	assertEqual(t, s1.Equal(s4), false)
	assertEqual(t, s4.Equal(s1), false)
	assertEqual(t, s1.Equal(s1), true)
}

// TestSetAlgebra tests Union, Intersection, Difference, and SymmetricDifference.
func TestSetAlgebra(t *testing.T) {
	s1 := NewSet(1, 2, 3, 4)
	s2 := NewSet(3, 4, 5, 6)

	// Union
	union := s1.Union(s2)
	expectedUnion := NewSet(1, 2, 3, 4, 5, 6)
	if !union.Equal(expectedUnion) {
		t.Errorf("Union failed. Expected %v, got %v", expectedUnion, union)
	}

	// Intersection
	intersection := s1.Intersection(s2)
	expectedIntersection := NewSet(3, 4)
	if !intersection.Equal(expectedIntersection) {
		t.Errorf("Intersection failed. Expected %v, got %v", expectedIntersection, intersection)
	}

	// Difference (s1 - s2)
	difference := s1.Difference(s2)
	expectedDifference := NewSet(1, 2)
	if !difference.Equal(expectedDifference) {
		t.Errorf("Difference failed. Expected %v, got %v", expectedDifference, difference)
	}

	// Symmetric Difference
	symDifference := s1.SymmetricDifference(s2)
	expectedSymDifference := NewSet(1, 2, 5, 6)
	if !symDifference.Equal(expectedSymDifference) {
		t.Errorf("SymmetricDifference failed. Expected %v, got %v", expectedSymDifference, symDifference)
	}
}

// TestSubsets tests IsSubset and IsSuperset.
func TestSubsets(t *testing.T) {
	s1 := NewSet(1, 2, 3, 4)
	s2 := NewSet(2, 3)
	s3 := NewSet(2, 3, 5)

	assertEqual(t, s2.IsSubset(s1), true)
	assertEqual(t, s1.IsSuperset(s2), true)

	assertEqual(t, s3.IsSubset(s1), false)
	assertEqual(t, s1.IsSuperset(s3), false)

	assertEqual(t, s1.IsSubset(s1), true) // A set is a subset of itself
	assertEqual(t, s1.IsSuperset(s1), true)
}

// TestEach tests the Each method for iteration.
func TestEach(t *testing.T) {
	s := NewSet(10, 20, 30)
	sum := 0
	s.Each(func(item int) bool {
		sum += item
		return true // continue iteration
	})
	assertEqual(t, sum, 60)

	// Test early exit
	count := 0
	stoppedAt20 := false
	s.Each(func(item int) bool {
		count++
		if item == 20 {
			stoppedAt20 = true
			return false // stop here
		}
		return true
	})
	if !stoppedAt20 {
		t.Errorf("Expected to stop at item 20, but didn't")
	}
}

// TestSetConcurrency is crucial for a thread-safe set.
// It runs multiple goroutines to add and remove items concurrently.
// Run with `go test -race` to detect race conditions.
func TestSetConcurrency(t *testing.T) {
	s := NewSet[int]()
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
	s := NewSet[int]()
	for i := 0; i < n; i++ {
		s.Add(i)
	}
	return s
}

func BenchmarkAdd(b *testing.B) {
	s := NewSet[int]()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Add(i)
	}
}

func BenchmarkAddUnSafe(b *testing.B) {
	s := NewSet[int]()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Add(i)
	}
}

func BenchmarkContains(b *testing.B) {
	s := createSetWithNItems(10000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Contains(5000) // Element that is always present
	}
}

func BenchmarkContainsUnSafe(b *testing.B) {
	s := createSetWithNItems(10000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Contains(5000) // Element that is always present
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

func BenchmarkRemoveUnSafe(b *testing.B) {
	// We need to re-populate the set on each run of the benchmark loop
	// to ensure we are always removing from a set of a similar size.
	b.StopTimer()
	s := createSetWithNItems(b.N)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		s.Remove(i)
	}
}

// --- 辅助函数 ---

// generateIntSet 创建一个包含指定数量整数的集合，用于测试。
// 元素从 0 到 size-1。
func generateIntSet(size int) *Set[int] {
	s := NewSet[int]()
	for i := 0; i < size; i++ {
		s.items[i] = struct{}{} // 使用无锁的方式快速填充
	}
	return s
}

// runSetOperationBenchmark 是一个通用的基准测试函数，用于测试需要两个集合作为输入的运算
func runSetOperationBenchmark(b *testing.B, operation func(s1, s2 *Set[int]) any) {
	sizes := []int{10000}
	overlaps := []float64{0.2, 0.8} // 测试 50%, 100% 重叠的情况

	for _, size := range sizes {
		for _, overlap := range overlaps {
			name := fmt.Sprintf("%d-%.0f%%", size, overlap*100)
			b.Run(name, func(b *testing.B) {
				// s1 包含 [0, size-1]
				s1 := generateIntSet(size)
				// s2 根据重叠度计算元素
				overlapCount := int(float64(size) * overlap)
				s2 := NewSet[int]()
				// 添加重叠部分: [0, overlapCount-1]
				for i := 0; i < overlapCount; i++ {
					s2.Add(i)
				}
				// 添加非重叠部分: [size, size + (size-overlapCount) - 1]
				for i := 0; i < size-overlapCount; i++ {
					s2.Add(size + i)
				}

				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					operation(s1, s2)
				}
			})
		}
	}
}

func BenchmarkEqual(b *testing.B) {
	sizes := []int{10000}
	for _, size := range sizes {
		s1 := generateIntSet(size)
		s2 := generateIntSet(size) // 与 s1 完全相同
		s3 := generateIntSet(size)
		s3.Add(size + 1) // 比 s1 多一个元素

		b.Run(fmt.Sprintf("True-%d", size), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				s1.Equal(s2)
			}
		})

		b.Run(fmt.Sprintf("False-%d", size), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				s1.Equal(s3)
			}
		})
	}
}

func BenchmarkUnion(b *testing.B) {
	runSetOperationBenchmark(b, func(s1, s2 *Set[int]) any {
		return s1.Union(s2)
	})
}

func BenchmarkUnionUnSafe(b *testing.B) {
	runSetOperationBenchmark(b, func(s1, s2 *Set[int]) any {
		return s1.Union(s2)
	})
}

func BenchmarkIntersection(b *testing.B) {
	runSetOperationBenchmark(b, func(s1, s2 *Set[int]) any {
		return s1.Intersection(s2)
	})
}

func BenchmarkDifference(b *testing.B) {
	runSetOperationBenchmark(b, func(s1, s2 *Set[int]) any {
		return s1.Difference(s2)
	})
}

func BenchmarkSymmetricDifference(b *testing.B) {
	runSetOperationBenchmark(b, func(s1, s2 *Set[int]) any {
		return s1.SymmetricDifference(s2)
	})
}

func BenchmarkIsSubset(b *testing.B) {
	sizes := []int{100, 1000, 10000}
	for _, size := range sizes {
		superSet := generateIntSet(size)   // 超集: [0, size-1]
		subSet := generateIntSet(size / 2) // 真子集: [0, size/2 - 1]
		notSubSet := generateIntSet(size / 2)
		notSubSet.Add(size + 1) // 包含一个超集里没有的元素

		b.Run(fmt.Sprintf("True-%d", size), func(b *testing.B) {
			var r bool
			for i := 0; i < b.N; i++ {
				r = subSet.IsSubset(superSet)
			}
			_ = r
		})

		b.Run(fmt.Sprintf("False-%d", size), func(b *testing.B) {
			var r bool
			for i := 0; i < b.N; i++ {
				r = notSubSet.IsSubset(superSet)
			}
			_ = r
		})
	}
}

func BenchmarkIsSuperset(b *testing.B) {
	sizes := []int{100, 1000, 10000}
	for _, size := range sizes {
		superSet := generateIntSet(size)   // 超集: [0, size-1]
		subSet := generateIntSet(size / 2) // 真子集: [0, size/2 - 1]

		b.Run(fmt.Sprintf("True-%d", size), func(b *testing.B) {
			var r bool
			for i := 0; i < b.N; i++ {
				r = superSet.IsSuperset(subSet)
			}
			_ = r
		})

		b.Run(fmt.Sprintf("False-%d", size), func(b *testing.B) {
			var r bool
			for i := 0; i < b.N; i++ {
				// IsSuperset 的 False case 是 IsSubset 的反向情况
				r = subSet.IsSuperset(superSet)
			}
			_ = r
		})
	}
}
