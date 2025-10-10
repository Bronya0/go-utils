package uid

import (
	"sync"
	"testing"
)

// BenchmarkNode_Generate 测试在高并发情况下的ID生成性能。
func BenchmarkNode_Generate(b *testing.B) {
	// 使用 worker ID 1 初始化一个节点。
	node, err := NewSnowflakeNode(1)
	if err != nil {
		b.Fatalf("创建节点失败: %v", err)
	}

	// b.RunParallel 会创建多个 goroutine 并发执行测试。
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = node.NewID()
		}
	})
}

// BenchmarkNode_Generate_NoContention 测试在单线程无竞争情况下的ID生成性能。
func BenchmarkNode_Generate_NoContention(b *testing.B) {
	node, err := NewSnowflakeNode(1)
	if err != nil {
		b.Fatalf("创建节点失败: %v", err)
	}

	// b.N 是由测试框架动态调整的循环次数。
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = node.NewID()
	}
}

// TestNode_Generate_Uniqueness 测试单线程生成的ID是否唯一。
func TestNode_Generate_Uniqueness(t *testing.T) {
	node, err := NewSnowflakeNode(1)
	if err != nil {
		t.Fatalf("创建节点失败: %v", err)
	}

	const numIDs = 1000000 // 生成一百万个ID进行测试
	ids := make(map[int64]bool, numIDs)
	for i := 0; i < numIDs; i++ {
		id := node.NewID()
		if ids[id] {
			// 如果发现重复ID，则测试失败。
			t.Fatalf("生成了重复的ID: %d", id)
		}
		ids[id] = true
	}
}

// TestNode_Generate_Concurrency_Uniqueness 测试多线程并发生成的ID是否唯一。
func TestNode_Generate_Concurrency_Uniqueness(t *testing.T) {
	node, err := NewSnowflakeNode(1)
	if err != nil {
		t.Fatalf("创建节点失败: %v", err)
	}

	const numGoRoutines = 100     // 并发协程数
	const idsPerGoRoutine = 10000 // 每个协程生成的ID数
	totalIDs := numGoRoutines * idsPerGoRoutine

	var wg sync.WaitGroup
	idChan := make(chan int64, totalIDs)

	for i := 0; i < numGoRoutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < idsPerGoRoutine; j++ {
				idChan <- node.NewID()
			}
		}()
	}

	wg.Wait()
	close(idChan)

	// 检查是否有重复ID。
	ids := make(map[int64]bool, totalIDs)
	for id := range idChan {
		if ids[id] {
			t.Fatalf("并发场景下生成了重复的ID: %d", id)
		}
		ids[id] = true
	}
}

// TestParseID 测试ID解析功能是否正确。
func TestParseID(t *testing.T) {
	node, err := NewSnowflakeNode(123) // 使用一个特定的worker ID
	if err != nil {
		t.Fatalf("创建节点失败: %v", err)
	}

	id := node.NewID()
	timestamp, workerID, sequence := ParseSnowflakeID(id)

	if workerID != 123 {
		t.Errorf("解析出的 workerID 不正确，期望 %d, 得到 %d", 123, workerID)
	}

	t.Logf("生成的ID: %d", id)
	t.Logf("解析结果 -> 时间戳: %d, Worker ID: %d, 序列号: %d", timestamp, workerID, sequence)
}
