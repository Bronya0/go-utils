package container

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"unsafe"
)

// Set 是一个通用的集合（Set）数据结构。
// 它可以被配置为线程安全的（并发的）或非线程安全的。
//
// 必须使用 NewSet() 或 NewConcurrentSet() 来创建实例。
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
func (s *Set[T]) Each(f func(item T) bool) {
	var keys []T
	if s.concurrent {
		s.mu.RLock()
		keys = make([]T, 0, len(s.items))
		for k := range s.items {
			keys = append(keys, k)
		}
		s.mu.RUnlock()
	}

	if keys != nil {
		for _, k := range keys {
			if !f(k) {
				break
			}
		}
	} else {
		for item := range s.items {
			if !f(item) {
				break
			}
		}
	}
}

// lockBoth a helper function to lock two sets in a consistent order.
func lockBoth[T comparable](s1, s2 *Set[T]) {
	p1 := unsafe.Pointer(s1)
	p2 := unsafe.Pointer(s2)
	if uintptr(p1) < uintptr(p2) {
		s1.mu.RLock()
		s2.mu.RLock()
	} else {
		s2.mu.RLock()
		s1.mu.RLock()
	}
}

// unlockBoth a helper function to unlock two sets in the reverse order of locking.
func unlockBoth[T comparable](s1, s2 *Set[T]) {
	p1 := unsafe.Pointer(s1)
	p2 := unsafe.Pointer(s2)
	if uintptr(p1) < uintptr(p2) {
		s2.mu.RUnlock()
		s1.mu.RUnlock()
	} else {
		s1.mu.RUnlock()
		s2.mu.RUnlock()
	}
}

// Equal 检查两个集合是否相等。
func (s *Set[T]) Equal(other *Set[T]) bool {
	if s == other {
		return true
	}

	if s.concurrent && other.concurrent {
		lockBoth(s, other)
		defer unlockBoth(s, other)
	} else if s.concurrent {
		s.mu.RLock()
		defer s.mu.RUnlock()
	} else if other.concurrent {
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

// Union 并集。
func (s *Set[T]) Union(other *Set[T]) *Set[T] {
	if s.concurrent && other.concurrent {
		lockBoth(s, other)
		defer unlockBoth(s, other)
	} else if s.concurrent {
		s.mu.RLock()
		defer s.mu.RUnlock()
	} else if other.concurrent {
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

// Intersection 交集。
func (s *Set[T]) Intersection(other *Set[T]) *Set[T] {
	if s.concurrent && other.concurrent {
		lockBoth(s, other)
		defer unlockBoth(s, other)
	} else if s.concurrent {
		s.mu.RLock()
		defer s.mu.RUnlock()
	} else if other.concurrent {
		other.mu.RLock()
		defer other.mu.RUnlock()
	}

	result := &Set[T]{
		concurrent: s.concurrent || other.concurrent,
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

// Difference 差集。
func (s *Set[T]) Difference(other *Set[T]) *Set[T] {
	if s.concurrent && other.concurrent {
		lockBoth(s, other)
		defer unlockBoth(s, other)
	} else if s.concurrent {
		s.mu.RLock()
		defer s.mu.RUnlock()
	} else if other.concurrent {
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

// SymmetricDifference 对称差集。
func (s *Set[T]) SymmetricDifference(other *Set[T]) *Set[T] {
	if s.concurrent && other.concurrent {
		lockBoth(s, other)
		defer unlockBoth(s, other)
	} else if s.concurrent {
		s.mu.RLock()
		defer s.mu.RUnlock()
	} else if other.concurrent {
		other.mu.RLock()
		defer other.mu.RUnlock()
	}

	result := &Set[T]{
		items:      make(map[T]struct{}),
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
	if s == other {
		return true
	}

	if s.concurrent && other.concurrent {
		lockBoth(s, other)
		defer unlockBoth(s, other)
	} else if s.concurrent {
		s.mu.RLock()
		defer s.mu.RUnlock()
	} else if other.concurrent {
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
	if s == other {
		return true
	}

	if s.concurrent && other.concurrent {
		lockBoth(s, other)
		defer unlockBoth(s, other)
	} else if s.concurrent {
		s.mu.RLock()
		defer s.mu.RUnlock()
	} else if other.concurrent {
		other.mu.RLock()
		defer other.mu.RUnlock()
	}

	if len(other.items) > len(s.items) {
		return false
	}

	for item := range other.items {
		if _, exists := s.items[item]; !exists {
			return false
		}
	}
	return true
}
