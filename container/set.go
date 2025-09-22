package container

import (
	"fmt"
	"sort"
	"strings"
	"sync"
)

// Set 是一个通用的集合（Set）数据结构。
// 它可以被配置为线程安全的（并发的）或非线程安全的。
//
// 必须使用 New() 或 NewConcurrent() 来创建实例。
//
//   - New(): 创建一个非线程安全的集合，适用于单 goroutine 环境，性能更高。
//   - NewConcurrent(): 创建一个线程安全的集合，适用于多 goroutine 环境。
//
// --- 示例 1: 基本操作 (非并发安全) ---
//
//	fmt.Println("--- String Set Example ---")
//	s1 := New("apple", "banana", "cherry")
//	s1.Add("apple", "date")
//	fmt.Printf("s1: %s, Length: %d\n", s1, s1.Len())
//
//	s1.Remove("banana")
//	fmt.Printf("After removing 'banana': %s\n", s1)
//	fmt.Printf("Does s1 contain 'apple'? %t\n", s1.Contains("apple"))
//	fmt.Printf("Does s1 contain 'banana'? %t\n", s1.Contains("banana"))
//
// --- 示例 2: 集合运算 (并发安全) ---
//
//	fmt.Println("\n--- Integer Set Operations (Concurrent) ---")
//	setA := NewConcurrent(1, 2, 3, 4, 5)
//	setB := NewConcurrent(4, 5, 6, 7, 8)
//	fmt.Printf("Set A: %s\n", setA)
//	fmt.Printf("Set B: %s\n", setB)
//
//	union := setA.Union(setB) // 结果集也会是并发安全的
//	fmt.Printf("Union (A ∪ B): %s\n", union)
//
//	intersection := setA.Intersection(setB)
//	fmt.Printf("Intersection (A ∩ B): %s\n", intersection)
type Set[T comparable] struct {
	mu         sync.RWMutex
	items      map[T]struct{}
	concurrent bool
}

// NewSet 非线程安全的集合，可以传入初始值。
func NewSet[T comparable](values ...T) *Set[T] {
	s := &Set[T]{
		items:      make(map[T]struct{}, len(values)),
		concurrent: false,
	}
	if len(values) > 0 {
		s.Add(values...)
	}
	return s
}

// NewConcurrentSet 线程安全的集合，可以传入初始值。
func NewConcurrentSet[T comparable](values ...T) *Set[T] {
	s := &Set[T]{
		items:      make(map[T]struct{}, len(values)),
		concurrent: true,
	}
	if len(values) > 0 {
		s.Add(values...)
	}
	return s
}

// Add 向集合中添加一个或多个元素。
func (s *Set[T]) Add(values ...T) {
	if s.concurrent {
		s.mu.Lock()
		defer s.mu.Unlock()
	}
	for _, value := range values {
		s.items[value] = struct{}{}
	}
}

// Remove 从集合中移除一个或多个元素。
func (s *Set[T]) Remove(values ...T) {
	if s.concurrent {
		s.mu.Lock()
		defer s.mu.Unlock()
	}
	for _, value := range values {
		delete(s.items, value)
	}
}

// Contains 检查集合中是否存在指定的元素。
func (s *Set[T]) Contains(value T) bool {
	if s.concurrent {
		s.mu.RLock()
		defer s.mu.RUnlock()
	}
	_, exists := s.items[value]
	return exists
}

// Len 返回集合中的元素数量。
func (s *Set[T]) Len() int {
	if s.concurrent {
		s.mu.RLock()
		defer s.mu.RUnlock()
	}
	return len(s.items)
}

// IsEmpty 检查集合是否为空。
func (s *Set[T]) IsEmpty() bool {
	return s.Len() == 0
}

// Clear 清空集合中的所有元素。
func (s *Set[T]) Clear() {
	if s.concurrent {
		s.mu.Lock()
		defer s.mu.Unlock()
	}
	s.items = make(map[T]struct{})
}

// ToSlice 将集合中的元素转换为一个切片。
// 注意：由于 map 的无序性，切片中元素的顺序是不确定的。
func (s *Set[T]) ToSlice() []T {
	if s.concurrent {
		s.mu.RLock()
		defer s.mu.RUnlock()
	}
	slice := make([]T, 0, len(s.items))
	for item := range s.items {
		slice = append(slice, item)
	}
	return slice
}

// Clone 创建并返回当前集合的一个浅拷贝。
// 克隆的集合将保持与原集合相同的并发设置。
func (s *Set[T]) Clone() *Set[T] {
	if s.concurrent {
		s.mu.RLock()
		defer s.mu.RUnlock()
	}
	clone := &Set[T]{
		items:      make(map[T]struct{}, len(s.items)),
		concurrent: s.concurrent,
	}
	for item := range s.items {
		clone.items[item] = struct{}{}
	}
	return clone
}

// String 实现了 fmt.Stringer 接口，用于打印集合内容。
// 为了输出稳定，会对元素进行排序。
func (s *Set[T]) String() string {
	slice := s.ToSlice()
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
// 对于并发安全的集合，在遍历前回先复制键列表，以避免在回调函数中修改集合时产生死锁。
func (s *Set[T]) Each(f func(item T) bool) {
	if s.concurrent {
		s.mu.RLock()
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
	} else {
		// 非并发集合，直接遍历
		for item := range s.items {
			if !f(item) {
				break
			}
		}
	}
}

// Equal 检查两个集合是否相等（包含完全相同的元素）。
func (s *Set[T]) Equal(other *Set[T]) bool {
	if s.concurrent {
		s.mu.RLock()
		defer s.mu.RUnlock()
	}
	if other.concurrent {
		other.mu.RLock()
		defer other.mu.RUnlock()
	}

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

// Union 并集。返回一个新集合，包含两个集合中的所有元素。
// 如果任一输入集合是并发安全的，则结果集也是并发安全的。
func (s *Set[T]) Union(other *Set[T]) *Set[T] {
	if s.concurrent {
		s.mu.RLock()
		defer s.mu.RUnlock()
	}
	if other.concurrent {
		other.mu.RLock()
		defer other.mu.RUnlock()
	}

	result := &Set[T]{
		items:      make(map[T]struct{}, len(s.items)+len(other.items)),
		concurrent: s.concurrent || other.concurrent,
	}

	for item := range s.items {
		result.items[item] = struct{}{}
	}
	for item := range other.items {
		result.items[item] = struct{}{}
	}
	return result
}

// Intersection 交集。返回一个新集合，包含同时存在于两个集合中的元素。
// 如果任一输入集合是并发安全的，则结果集也是并发安全的。
func (s *Set[T]) Intersection(other *Set[T]) *Set[T] {
	if s.concurrent {
		s.mu.RLock()
		defer s.mu.RUnlock()
	}
	if other.concurrent {
		other.mu.RLock()
		defer other.mu.RUnlock()
	}

	result := &Set[T]{
		concurrent: s.concurrent || other.concurrent,
	}

	if len(s.items) == 0 || len(other.items) == 0 {
		result.items = make(map[T]struct{})
		return result
	}

	var smaller, larger *Set[T]
	if len(s.items) < len(other.items) {
		smaller, larger = s, other
	} else {
		smaller, larger = other, s
	}

	result.items = make(map[T]struct{}, len(smaller.items))
	for item := range smaller.items {
		if _, exists := larger.items[item]; exists {
			result.items[item] = struct{}{}
		}
	}
	return result
}

// Difference 差集。返回一个新集合，包含在当前集合中但不在另一个集合中的元素 (S - Other)。
// 如果任一输入集合是并发安全的，则结果集也是并发安全的。
func (s *Set[T]) Difference(other *Set[T]) *Set[T] {
	if s.concurrent {
		s.mu.RLock()
		defer s.mu.RUnlock()
	}
	if other.concurrent {
		other.mu.RLock()
		defer other.mu.RUnlock()
	}

	result := &Set[T]{
		items:      make(map[T]struct{}, len(s.items)),
		concurrent: s.concurrent || other.concurrent,
	}

	for item := range s.items {
		if _, exists := other.items[item]; !exists {
			result.items[item] = struct{}{}
		}
	}
	return result
}

// SymmetricDifference 对称差集。返回一个新集合，包含只存在于其中一个集合中的元素。
// 如果任一输入集合是并发安全的，则结果集也是并发安全的。
func (s *Set[T]) SymmetricDifference(other *Set[T]) *Set[T] {
	if s.concurrent {
		s.mu.RLock()
		defer s.mu.RUnlock()
	}
	if other.concurrent {
		other.mu.RLock()
		defer other.mu.RUnlock()
	}

	result := &Set[T]{
		items:      make(map[T]struct{}, len(s.items)+len(other.items)),
		concurrent: s.concurrent || other.concurrent,
	}

	for item := range s.items {
		if _, exists := other.items[item]; !exists {
			result.items[item] = struct{}{}
		}
	}
	for item := range other.items {
		if _, exists := s.items[item]; !exists {
			result.items[item] = struct{}{}
		}
	}
	return result
}

// IsSubset 检查当前集合是否是另一个集合的子集。
func (s *Set[T]) IsSubset(other *Set[T]) bool {
	if s.concurrent {
		s.mu.RLock()
		defer s.mu.RUnlock()
	}
	if other.concurrent {
		other.mu.RLock()
		defer other.mu.RUnlock()
	}

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
