package uid

import (
	"bufio"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"io"
	"math"
	"math/bits"
	"sync"
	"time"
)

// --- 核心类型和常量 ---

// ulidTag 是一个16字节的通用唯一词法可排序标识符。
type ulidTag [16]byte

const (
	// encodedSize 是文本编码后ULID的长度 (26个字符)。
	encodedSize = 26
	// encoding 是ULID字符串中使用的Base32编码字母表。
	encoding = "0123456789ABCDEFGHJKMNPQRSTVWXYZ"
)

var (
	// ErrBigTime 在使用大于 MaxTime 的时间戳构造ULID时返回。
	ErrBigTime = errors.New("ulidTag: time too big")
	// ErrMonotonicOverflow 在单调熵源递增时发生溢出时返回。
	ErrMonotonicOverflow = errors.New("ulidTag: monotonic entropy overflow")
	// maxTime 是可以在ULID中表示的最大Unix毫秒时间。
	maxTime = uint64(281474976710655) // 0xFFFFFFFFFFFF
)

// --- 公共 API：这就是您需要调用的函数 ---

var (
	// 使用 sync.Once 确保全局熵源只被初始化一次。
	initOnce sync.Once
	// 这是我们唯一的、全局的、线程安全的熵源。
	defaultSecureEntropy io.Reader
)

// initializeDefaultEntropy 使用密码学安全的熵源 `crypto/rand` 来创建
// 一个线程安全的单调读取器。
func initializeDefaultEntropy() {
	// 使用密码学安全的 `crypto/rand` 作为熵的基础来源。
	// 第二个参数 `0` 会使其使用一个安全的默认增量值 (math.MaxUint32)。
	source := monotonic(rand.Reader, 0)

	// 使用互斥锁包装，使其可以安全地在多个goroutine中并发使用。
	defaultSecureEntropy = &lockedMonotonicReader{monotonicReader: source}
}

// NewULID 生成一个新的、安全的、基于当前时间的 ulidTag 字符串。
// 这个函数是并发安全的，可以直接在多个 Goroutine 中调用。
// 它内部处理了所有关于熵和时间戳的细节。
func NewULID() string {
	// 确保我们的熵源已经被安全地初始化了。
	// Do 方法会保证 initializeDefaultEntropy() 函数在全局范围内只执行一次。
	initOnce.Do(initializeDefaultEntropy)

	// 使用 mustNew，它在极少数情况下（如 crypto/rand 读取失败）会直接 panic。
	// 这符合“即拿即用”的理念，因为这种情况通常意味着系统存在严重问题。
	id := mustNew(timestamp(time.Now().UTC()), defaultSecureEntropy)
	return id.String()
}

// --- 以下是支持生成逻辑所需的内部实现，从原始代码中提取和修改 ---

// mustNew 是 newUlid 的一个便捷函数，它在失败时会 panic 而不是返回错误。
func mustNew(ms uint64, entropy io.Reader) ulidTag {
	id, err := newUlid(ms, entropy)
	if err != nil {
		panic(err)
	}
	return id
}

// newUlid 使用给定的Unix毫秒时间戳和熵源返回一个ULID。
func newUlid(ms uint64, entropy io.Reader) (id ulidTag, err error) {
	if ms > maxTime {
		return id, ErrBigTime
	}

	// 设置时间部分 (前6个字节)
	id[0] = byte(ms >> 40)
	id[1] = byte(ms >> 32)
	id[2] = byte(ms >> 24)
	id[3] = byte(ms >> 16)
	id[4] = byte(ms >> 8)
	id[5] = byte(ms)

	// 从熵源填充随机部分 (后10个字节)
	if mr, ok := entropy.(monotonicReader); ok {
		err = mr.MonotonicRead(ms, id[6:])
	} else {
		_, err = io.ReadFull(entropy, id[6:])
	}

	return id, err
}

// String 返回ULID的词法可排序字符串编码（26个字符）。
func (id ulidTag) String() string {
	ulid := make([]byte, encodedSize)
	// 10 字节时间戳
	ulid[0] = encoding[(id[0]&224)>>5]
	ulid[1] = encoding[id[0]&31]
	ulid[2] = encoding[(id[1]&248)>>3]
	ulid[3] = encoding[((id[1]&7)<<2)|((id[2]&192)>>6)]
	ulid[4] = encoding[(id[2]&62)>>1]
	ulid[5] = encoding[((id[2]&1)<<4)|((id[3]&240)>>4)]
	ulid[6] = encoding[((id[3]&15)<<1)|((id[4]&128)>>7)]
	ulid[7] = encoding[(id[4]&124)>>2]
	ulid[8] = encoding[((id[4]&3)<<3)|((id[5]&224)>>5)]
	ulid[9] = encoding[id[5]&31]
	// 16 字节熵
	ulid[10] = encoding[(id[6]&248)>>3]
	ulid[11] = encoding[((id[6]&7)<<2)|((id[7]&192)>>6)]
	ulid[12] = encoding[(id[7]&62)>>1]
	ulid[13] = encoding[((id[7]&1)<<4)|((id[8]&240)>>4)]
	ulid[14] = encoding[((id[8]&15)<<1)|((id[9]&128)>>7)]
	ulid[15] = encoding[(id[9]&124)>>2]
	ulid[16] = encoding[((id[9]&3)<<3)|((id[10]&224)>>5)]
	ulid[17] = encoding[id[10]&31]
	ulid[18] = encoding[(id[11]&248)>>3]
	ulid[19] = encoding[((id[11]&7)<<2)|((id[12]&192)>>6)]
	ulid[20] = encoding[(id[12]&62)>>1]
	ulid[21] = encoding[((id[12]&1)<<4)|((id[13]&240)>>4)]
	ulid[22] = encoding[((id[13]&15)<<1)|((id[14]&128)>>7)]
	ulid[23] = encoding[(id[14]&124)>>2]
	ulid[24] = encoding[((id[14]&3)<<3)|((id[15]&224)>>5)]
	ulid[25] = encoding[id[15]&31]

	return string(ulid)
}

// timestamp 将 time.Time 转换为 Unix 毫秒。
func timestamp(t time.Time) uint64 {
	return uint64(t.UnixNano()) / uint64(time.Millisecond)
}

// --- 单调熵 (monotonic Entropy) 相关的内部逻辑 ---

// monotonicReader 是一个接口，它应该为相同的毫秒参数产生单调递增的熵。
type monotonicReader interface {
	io.Reader
	MonotonicRead(ms uint64, p []byte) error
}

// MonotonicEntropy 提供单调熵。
type MonotonicEntropy struct {
	io.Reader
	ms      uint64
	inc     uint64
	entropy [10]byte
	rand    [8]byte
	isZero  bool
}

// monotonic 返回一个熵源，该熵源产生严格递增的熵字节。
func monotonic(entropy io.Reader, inc uint64) *MonotonicEntropy {
	m := MonotonicEntropy{
		Reader: bufio.NewReader(entropy),
		inc:    inc,
		isZero: true,
	}
	if m.inc == 0 {
		m.inc = math.MaxUint32
	}
	return &m
}

// MonotonicRead 实现 monotonicReader 接口。
func (m *MonotonicEntropy) MonotonicRead(ms uint64, p []byte) (err error) {
	if !m.isZero && m.ms == ms {
		err = m.increment()
		copy(p, m.entropy[:])
	} else {
		_, err = io.ReadFull(m.Reader, p)
		if err == nil {
			m.ms = ms
			copy(m.entropy[:], p)
			m.isZero = false
		}
	}
	return err
}

func (m *MonotonicEntropy) increment() error {
	inc, err := m.random()
	if err != nil {
		return err
	}

	// 以大端序递增10字节（80位）的熵
	i := len(m.entropy) - 1
	var carry uint64
	for i >= 2 {
		val := uint64(m.entropy[i]) + (inc & 0xFF) + carry
		m.entropy[i] = byte(val)
		carry = val >> 8
		inc >>= 8
		i--
		if inc == 0 && carry == 0 {
			break
		}
	}

	if i < 2 && (inc > 0 || carry > 0) {
		val := uint64(binary.BigEndian.Uint16(m.entropy[:2])) + inc + carry
		if val > math.MaxUint16 {
			return ErrMonotonicOverflow
		}
		binary.BigEndian.PutUint16(m.entropy[:2], uint16(val))
	}

	return nil
}

func (m *MonotonicEntropy) random() (inc uint64, err error) {
	if m.inc <= 1 {
		return 1, nil
	}
	bitLen := bits.Len64(m.inc - 1)
	byteLen := uint(bitLen+7) / 8
	mask := byte((1 << (uint(bitLen%8) + 8)) - 1)
	if bitLen%8 == 0 {
		mask = 0xFF
	}

	for {
		if _, err = io.ReadFull(m.Reader, m.rand[:byteLen]); err != nil {
			return 0, err
		}
		m.rand[0] &= mask
		inc = binary.LittleEndian.Uint64(m.rand[:])
		if inc < m.inc {
			break
		}
	}
	return 1 + inc, nil
}

// lockedMonotonicReader 使用 sync.Mutex 包装 monotonicReader 以实现并发安全。
type lockedMonotonicReader struct {
	mu sync.Mutex
	monotonicReader
}

// MonotonicRead 同步对被包装的 monotonicReader 的调用。
func (r *lockedMonotonicReader) MonotonicRead(ms uint64, p []byte) (err error) {
	r.mu.Lock()
	err = r.monotonicReader.MonotonicRead(ms, p)
	r.mu.Unlock()
	return err
}
