package uid

import (
	"fmt"
	"sync"
	"time"
)

const (
	// workerIDBits 表示 worker id 所占的比特数。
	// 默认值为 10，意味着最多支持 1024 个节点。
	workerIDBits uint8 = 10

	// sequenceBits 表示每个节点每毫秒生成的序列号所占的比特数。
	// 默认值为 12，意味着每个节点每毫秒最多可以生成 4096 个 ID。
	sequenceBits uint8 = 12

	// workerIDShift 是 worker id 的左移位数。
	// 值为 sequenceBits。
	workerIDShift = sequenceBits

	// timestampShift 是时间戳的左移位数。
	// 值为 workerIDBits + sequenceBits。
	timestampShift = workerIDBits + sequenceBits

	// sequenceMask 是序列号的掩码，用于防止序列号溢出。
	// -1 左移 sequenceBits 位，然后取反，即为序列号的最大值。
	sequenceMask = -1 ^ (-1 << sequenceBits)

	// maxWorkerID 是 worker id 的最大值。
	// -1 左移 workerIDBits 位，然后取反，即为 worker id 的最大值。
	maxWorkerID = -1 ^ (-1 << workerIDBits)
)

// Epoch 是雪花ID算法的起始时间戳（毫秒）。
// 这个值一旦确定，就不能再更改。2025-10-01 00:00:00
var Epoch int64 = 1759248000000

// SnowflakeNode 代表一个雪花ID生成器节点。
type SnowflakeNode struct {
	mu            sync.Mutex // 互斥锁，保证并发安全
	lastTimestamp int64      // 上次生成ID时的时间戳（毫秒）
	workerID      int64      // Worker ID
	sequence      int64      // 序列号
}

// NewSnowflakeNode 使用给定的 worker id 创建一个新的雪花ID节点。
//
// 重要提示：worker id 在您的整个分布式系统中必须是唯一的！
// 您需要自己管理 worker id 的分配，确保不同的节点使用不同的 worker id。
// worker id 的取值范围是 [0, 1023]。
func NewSnowflakeNode(workerID int64) (*SnowflakeNode, error) {
	if workerID < 0 || workerID > maxWorkerID {
		return nil, fmt.Errorf("worker ID %d must be between 0 and %d", workerID, maxWorkerID)
	}
	return &SnowflakeNode{
		workerID: workerID,
	}, nil
}

// NewID 生成一个唯一的、单调递增的雪花ID。
func (n *SnowflakeNode) NewID() int64 {
	n.mu.Lock()
	defer n.mu.Unlock()

	// 获取当前的毫秒级时间戳。
	now := time.Now().UnixNano() / 1e6

	// --- 处理时钟回拨 ---
	// 如果当前时间小于上一次记录的时间戳，说明发生了时钟回拨。
	if now < n.lastTimestamp {
		// 时钟回拨是分布式系统中的严重问题。传统的雪花算法会在此处报错。
		// 但为了追求高可用性，我们选择等待，直到时钟追平上一次的时间戳。
		// 这会造成当前goroutine的阻塞，但能保证ID的单调递增性。
		time.Sleep(time.Duration(n.lastTimestamp-now) * time.Millisecond)
		now = time.Now().UnixNano() / 1e6 // 再次获取当前时间
	}

	// 如果在同一毫秒内生成ID。
	if now == n.lastTimestamp {
		// 序列号加1，并与序列号掩码进行位与运算，防止溢出。
		n.sequence = (n.sequence + 1) & sequenceMask
		// 如果序列号溢出（即达到4096），则必须等待下一个毫秒。
		if n.sequence == 0 {
			// 自旋等待，直到进入下一个毫秒。
			for now <= n.lastTimestamp {
				now = time.Now().UnixNano() / 1e6
			}
		}
	} else {
		// 如果是新的毫秒，则重置序列号为0。
		n.sequence = 0
	}

	// 更新最后时间戳。
	n.lastTimestamp = now

	// 组装雪花ID。
	// 1. (时间戳 - Epoch) 左移 timestampShift 位
	// 2. worker id 左移 workerIDShift 位
	// 3. 或上序列号
	id := ((now - Epoch) << timestampShift) |
		(n.workerID << workerIDShift) |
		(n.sequence)

	return id
}

// ParseSnowflakeID 从一个雪花ID中解析出时间戳、Worker ID和序列号。
// 这对于调试和验证ID非常有用。
func ParseSnowflakeID(id int64) (timestamp int64, workerID int64, sequence int64) {
	timestamp = (id >> timestampShift) + Epoch
	workerID = (id >> workerIDShift) & maxWorkerID
	sequence = id & sequenceMask
	return
}
