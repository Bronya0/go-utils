package container

import (
	"fmt"
	"sort"
	"strings"
	"sync"
)

// Set 是一个线程安全的、支持泛型的集合类型。
// 底层使用 map[T]struct{} 实现以获得最佳性能。
//
// // --- 示例 1: 基本操作 (string类型) ---
//
//	fmt.Println("--- String Set Example ---")
//	s1 := NewSetWithValues("apple", "banana", "cherry")
//	s1.Add("apple", "date")
//	fmt.Printf("s1: %s, Length: %d\n", s1, s1.Len())
//
//	s1.Remove("banana")
//	fmt.Printf("After removing 'banana': %s\n", s1)
//	fmt.Printf("Does s1 contain 'apple'? %t\n", s1.Contains("apple"))
//	fmt.Printf("Does s1 contain 'banana'? %t\n", s1.Contains("banana"))
//
//	// --- 示例 2: 集合运算 (int类型) ---
//	fmt.Println("\n--- Integer Set Operations ---")
//	setA := NewSetWithValues(1, 2, 3, 4, 5)
//	setB := NewSetWithValues(4, 5, 6, 7, 8)
//	fmt.Printf("Set A: %s\n", setA)
//	fmt.Printf("Set B: %s\n", setB)
//
//	union := setA.Union(setB)
//	fmt.Printf("Union (A ∪ B): %s\n", union)
//
//	intersection := setA.Intersection(setB)
//	fmt.Printf("Intersection (A ∩ B): %s\n", intersection)
//
//	difference := setA.Difference(setB)
//	fmt.Printf("Difference (A - B): %s\n", difference)
//
//	symDifference := setA.SymmetricDifference(setB)
//	fmt.Printf("Symmetric Difference (A Δ B): %s\n", symDifference)
//
//	// --- 示例 3: 子集和相等判断 ---
//	fmt.Println("\n--- Subset and Equality ---")
//	setC := NewSetWithValues(1, 2, 3)
//	setD := NewSetWithValues(1, 2, 3, 4, 5)
//	fmt.Printf("Set C: %s\n", setC)
//	fmt.Printf("Set D: %s\n", setD)
//	fmt.Printf("Is C a subset of A? %t\n", setC.IsSubset(setA))
//	fmt.Printf("Is A a superset of C? %t\n", setA.IsSuperset(setC))
//	fmt.Printf("Is A equal to D? %t\n", setA.Equal(setD))
//
//	// --- 示例 4: 迭代 ---
//	fmt.Println("\n--- Iteration using Each ---")
//	setA.Each(func(item int) bool {
//		fmt.Printf("Iterating item: %d\n", item)
//		if item == 3 {
//			fmt.Println("Stopping iteration at 3.")
//			return false // 中断遍历
//		}
//		return true // 继续遍历
//	})
type Set[T comparable] struct {
	mu    sync.RWMutex
	items map[T]struct{}
}

// NewSet 创建并返回一个新的空集合。
func NewSet[T comparable]() *Set[T] {
	return &Set[T]{
		items: make(map[T]struct{}),
	}
}

// NewSetWithValues 创建并返回一个包含初始值的新集合。
func NewSetWithValues[T comparable](values ...T) *Set[T] {
	s := NewSet[T]()
	s.Add(values...)
	return s
}

// Add 向集合中添加一个或多个元素。
func (s *Set[T]) Add(values ...T) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.AddUnSafe(values...)
}

// AddUnSafe 无锁添加元素
func (s *Set[T]) AddUnSafe(values ...T) {
	for _, value := range values {
		s.items[value] = struct{}{}
	}
}

// Remove 从集合中移除一个或多个元素。
func (s *Set[T]) Remove(values ...T) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.RemoveUnSafe(values...)
}

// RemoveUnSafe 无锁移除元素
func (s *Set[T]) RemoveUnSafe(values ...T) {
	for _, value := range values {
		delete(s.items, value)
	}
}

// Contains 检查集合中是否存在指定的元素。
func (s *Set[T]) Contains(value T) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.ContainsUnSafe(value)
}

// ContainsUnSafe 无锁检查元素
func (s *Set[T]) ContainsUnSafe(value T) bool {
	_, exists := s.items[value]
	return exists
}

// Len 返回集合中的元素数量。
func (s *Set[T]) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.LenUnSafe()
}

// LenUnSafe 无锁获取元素数量
func (s *Set[T]) LenUnSafe() int {
	return len(s.items)
}

// IsEmpty 检查集合是否为空。
func (s *Set[T]) IsEmpty() bool {
	return s.Len() == 0
}

// Clear 清空集合中的所有元素。
func (s *Set[T]) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items = make(map[T]struct{})
}

// ToSlice 将集合中的元素转换为一个切片。
// 注意：由于 map 的无序性，切片中元素的顺序是不确定的。
func (s *Set[T]) ToSlice() []T {
	s.mu.RLock()
	defer s.mu.RUnlock()
	slice := make([]T, 0, len(s.items))
	for item := range s.items {
		slice = append(slice, item)
	}
	return slice
}

// Clone 创建并返回当前集合的一个浅拷贝。
func (s *Set[T]) Clone() *Set[T] {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.CloneUnSafe()
}

// CloneUnSafe 无锁拷贝
func (s *Set[T]) CloneUnSafe() *Set[T] {
	clone := NewSet[T]()
	clone.items = make(map[T]struct{}, len(s.items))
	for item := range s.items {
		clone.items[item] = struct{}{}
	}
	return clone
}

// String 实现了 fmt.Stringer 接口，用于打印集合内容。
// 为了输出稳定，会对元素进行排序（如果可能）。
func (s *Set[T]) String() string {
	slice := s.ToSlice()
	// 为了稳定的输出，尝试对 slice 进行排序
	// 这里我们只处理几种常见类型，实际应用中可能需要更复杂的处理
	sort.Slice(slice, func(i, j int) bool {
		return fmt.Sprint(slice[i]) < fmt.Sprint(slice[j])
	})

	var builder strings.Builder
	builder.WriteString("Set{")
	for i, item := range slice {
		if i > 0 {
			builder.WriteString(", ")
		}
		builder.WriteString(fmt.Sprintf("%v", item))
	}
	builder.WriteString("}")
	return builder.String()
}

// Each 遍历集合中的所有元素，并对每个元素执行给定的函数。
// 如果函数返回 false，则停止遍历。
func (s *Set[T]) Each(f func(item T) bool) {
	s.mu.RLock()
	// 复制 keys 以避免在 f 中修改集合时产生死锁
	keys := make([]T, 0, len(s.items))
	for k := range s.items {
		keys = append(keys, k)
	}
	s.mu.RUnlock() // 尽早释放锁

	for _, k := range keys {
		if !f(k) {
			break
		}
	}
}

// Equal 检查两个集合是否相等（包含完全相同的元素）。
func (s *Set[T]) Equal(other *Set[T]) bool {
	s.mu.RLock()
	other.mu.RLock()
	defer s.mu.RUnlock()
	defer other.mu.RUnlock()

	if len(s.items) != len(other.items) {
		return false
	}

	for item := range s.items {
		if _, exists := other.items[item]; !exists {
			return false
		}
	}
	return true
}

// Union 并集。返回一个新集合，包含两个集合中的所有元素（并集）。
func (s *Set[T]) Union(other *Set[T]) *Set[T] {
	s.mu.RLock()
	other.mu.RLock()
	defer s.mu.RUnlock()
	defer other.mu.RUnlock()
	return s.UnionUnSafe(other)
}

// UnionUnSafe 无锁并集
func (s *Set[T]) UnionUnSafe(other *Set[T]) *Set[T] {
	// 确定较大和较小的集合
	var smaller, larger *Set[T]
	if len(s.items) < len(other.items) {
		smaller, larger = s, other
	} else {
		smaller, larger = other, s
	}

	// 快速处理较小集合为空的情况
	if len(smaller.items) == 0 {
		return larger.Clone()
	}

	resultItems := make(map[T]struct{}, len(larger.items)+len(smaller.items)) // 预估容量

	// 先复制较大集合
	for item := range larger.items {
		resultItems[item] = struct{}{}
	}
	// 再添加较小集合中独有的元素
	for item := range smaller.items {
		resultItems[item] = struct{}{}
	}
	return &Set[T]{items: resultItems}
}

// Intersection 交集。返回一个新集合，包含同时存在于两个集合中的元素（交集）。
func (s *Set[T]) Intersection(other *Set[T]) *Set[T] {
	s.mu.RLock()
	other.mu.RLock()
	defer s.mu.RUnlock()
	defer other.mu.RUnlock()

	result := NewSet[T]()
	// 如果任一集合为空，直接返回空集
	if len(s.items) == 0 || len(other.items) == 0 {
		return result
	}

	// 遍历较小的集合以提高性能
	var smaller, larger *Set[T]
	if len(s.items) < len(other.items) {
		smaller, larger = s, other
	} else {
		smaller, larger = other, s
	}
	// 预分配 map 容量
	result.items = make(map[T]struct{}, len(smaller.items))

	for item := range smaller.items {
		if _, exists := larger.items[item]; exists {
			result.items[item] = struct{}{}
		}
	}
	return result
}

// Difference 差集。返回一个新集合，包含在当前集合中但不在另一个集合中的元素（差集 S - Other）。
func (s *Set[T]) Difference(other *Set[T]) *Set[T] {
	result := NewSet[T]()
	s.mu.RLock()
	other.mu.RLock()
	defer s.mu.RUnlock()
	defer other.mu.RUnlock()

	result.items = make(map[T]struct{}, len(s.items))

	for item := range s.items {
		if _, exists := other.items[item]; !exists {
			result.items[item] = struct{}{}
		}
	}
	return result
}

// SymmetricDifference 对称差集。返回一个新集合，包含只存在于其中一个集合中的元素（对称差集）。
func (s *Set[T]) SymmetricDifference(other *Set[T]) *Set[T] {
	s.mu.RLock()
	other.mu.RLock()
	defer s.mu.RUnlock()
	defer other.mu.RUnlock()
	capacity := len(s.items) + len(other.items)
	resultItems := make(map[T]struct{}, capacity)
	// 1. 添加 s 中有，但 other 中没有的元素
	for item := range s.items {
		if _, exists := other.items[item]; !exists {
			resultItems[item] = struct{}{}
		}
	}

	// 2. 添加 other 中有，但 s 中没有的元素
	for item := range other.items {
		if _, exists := s.items[item]; !exists {
			resultItems[item] = struct{}{}
		}
	}

	return &Set[T]{items: resultItems}
}

// IsSubset 检查当前集合是否是另一个集合的子集。
func (s *Set[T]) IsSubset(other *Set[T]) bool {
	s.mu.RLock()
	other.mu.RLock()
	defer s.mu.RUnlock()
	defer other.mu.RUnlock()

	if len(s.items) > len(other.items) {
		return false
	}

	for item := range s.items {
		if _, exists := other.items[item]; !exists {
			return false
		}
	}
	return true
}

// IsSuperset 检查当前集合是否是另一个集合的超集。
func (s *Set[T]) IsSuperset(other *Set[T]) bool {
	return other.IsSubset(s)
}
