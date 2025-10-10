package uid

import (
	"fmt"
	"regexp"
	"sort"
	"sync"
	"testing"
	"time"
)

// --- 单元测试 (Unit Tests) ---
// go test -v
// go test -bench=.

// TestGenerate_Format 验证生成的ULID是否符合预期的格式。
// 1. 长度为26。
// 2. 只包含ULID的Base32字符集。
func TestGenerate_Format(t *testing.T) {
	id := NewULID()
	fmt.Println(id)

	// 1. 检查长度
	if len(id) != encodedSize {
		t.Errorf("NewID() aULID长度错误，期望 %d, 得到 %d", encodedSize, len(id))
	}

	// 2. 使用正则表达式检查字符集
	// 表达式 ^[0-9A-HJKMNP-TV-Z]{26}$ 确保字符串由26个指定的字符组成
	validChars := "^[" + regexp.QuoteMeta(encoding) + "]{26}$"
	match, err := regexp.MatchString(validChars, id)
	if err != nil {
		t.Fatalf("正则表达式匹配失败: %v", err)
	}
	if !match {
		t.Errorf("NewID() 生成的ULID '%s' 包含无效字符", id)
	}
}

// TestGenerate_Uniqueness 验证在大量生成时ULID的唯一性。
func TestGenerate_Uniqueness(t *testing.T) {
	const numIDs = 100000
	idSet := make(map[string]struct{}, numIDs)

	for i := 0; i < numIDs; i++ {
		id := NewULID()
		if _, exists := idSet[id]; exists {
			t.Fatalf("生成了重复的ULID: %s", id)
		}
		idSet[id] = struct{}{}
	}

	if len(idSet) != numIDs {
		t.Errorf("生成的唯一ULID数量与预期不符，期望 %d, 得到 %d", numIDs, len(idSet))
	}
}

// TestGenerate_Monotonicity 验证生成的ULID是否按字典序单调递增。
// 这是ULID的一个核心特性。
func TestGenerate_Monotonicity(t *testing.T) {
	const numIDs = 1000
	ids := make([]string, numIDs)
	for i := 0; i < numIDs; i++ {
		ids[i] = NewULID()
		// 在两次生成之间引入微小的延迟，以测试跨毫秒和同一毫秒内的情况
		time.Sleep(time.Microsecond)
	}

	// 检查原始切片是否已经排序
	isSorted := sort.StringsAreSorted(ids)
	if !isSorted {
		// 为了更好地调试，我们可以创建一个排序后的副本进行比较
		sortedIDs := make([]string, numIDs)
		copy(sortedIDs, ids)
		sort.Strings(sortedIDs)

		for i := 0; i < numIDs; i++ {
			if ids[i] != sortedIDs[i] {
				t.Errorf("单调性检查失败在索引 %d: 期望 %s, 得到 %s", i, sortedIDs[i], ids[i])
				// 找到第一个错误点后就停止，避免大量输出
				t.Fatalf("ULID序列不是单调递增的。")
			}
		}
	}
}

// TestGenerate_Concurrency 并发安全测试。
// 这个测试的目的是在开启Go的竞争检测器(-race)时，不会报告任何问题。
func TestGenerate_Concurrency(t *testing.T) {
	const numGoroutines = 50
	const idsPerGoroutine = 2000
	totalIDs := numGoroutines * idsPerGoroutine

	var wg sync.WaitGroup
	idChan := make(chan string, totalIDs)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < idsPerGoroutine; j++ {
				idChan <- NewULID()
			}
		}()
	}

	wg.Wait()
	close(idChan)

	// 除了竞争检测，我们还可以顺便检查并发生成下的唯一性
	idSet := make(map[string]struct{}, totalIDs)
	for id := range idChan {
		if _, exists := idSet[id]; exists {
			t.Errorf("并发生成时出现重复ULID: %s", id)
		}
		idSet[id] = struct{}{}
	}

	if len(idSet) != totalIDs {
		t.Errorf("并发生成的唯一ULID数量与预期不符，期望 %d, 得到 %d", totalIDs, len(idSet))
	}
}

// --- 基准测试/压测 (Benchmarks) ---

// BenchmarkGenerate_Simple 单线程（串行）生成性能测试
func BenchmarkGenerate_Simple(b *testing.B) {
	b.ReportAllocs() // 报告内存分配情况
	for i := 0; i < b.N; i++ {
		NewULID()
	}
}

// BenchmarkGenerate_Concurrent 并发生成性能测试
func BenchmarkGenerate_Concurrent(b *testing.B) {
	b.ReportAllocs()
	// b.RunParallel会创建多个goroutine，并将b.N分配给它们。
	// 这是Go语言中进行并行基准测试的标准方法。
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			NewULID()
		}
	})
}
