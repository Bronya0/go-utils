package convert

import (
	"fmt"
	"math"
	"strings"
	"time"
)

// 常用字节单位（IEC标准, 1024 进制）
const (
	KB = 1024
	MB = KB * 1024
	GB = MB * 1024
	TB = GB * 1024
	PB = TB * 1024
)

// HumanBytes 将字节数格式化为可读字符串
// 例如: 1536 -> "1.50 KB"
func HumanBytes(size int64) string {
	if size < KB {
		return fmt.Sprintf("%d B", size)
	}
	unit := []string{"B", "KB", "MB", "GB", "TB", "PB"}
	i := int(math.Floor(math.Log(float64(size)) / math.Log(KB)))
	val := float64(size) / math.Pow(KB, float64(i))
	return fmt.Sprintf("%.2f %s", val, unit[i])
}

// ParseBytes 将带单位的字符串转为字节数
// 支持: B, KB, MB, GB, TB, PB (不区分大小写)
// 例如: "1.5GB" -> 1610612736
func ParseBytes(s string) (int64, error) {
	s = strings.TrimSpace(strings.ToUpper(s))
	units := map[string]int64{
		"B":  1,
		"KB": KB,
		"MB": MB,
		"GB": GB,
		"TB": TB,
		"PB": PB,
	}
	for k, v := range units {
		if strings.HasSuffix(s, k) {
			val := strings.TrimSuffix(s, k)
			var num float64
			if _, err := fmt.Sscanf(val, "%f", &num); err != nil {
				return 0, err
			}
			return int64(num * float64(v)), nil
		}
	}
	return 0, fmt.Errorf("invalid size string: %s", s)
}

// HumanBandwidth 将字节/秒 转换为 Mbps, Gbps 等
// 网络流量速率转换
// （适用于带宽、速率显示，如 Mbps、MB/s）
// 例如: 125000 -> "1.00 Mbps"  (≈125 KB/s)
func HumanBandwidth(bytesPerSec int64) string {
	bitsPerSec := bytesPerSec * 8
	unit := []string{"bps", "Kbps", "Mbps", "Gbps", "Tbps"}
	if bitsPerSec < 1000 {
		return fmt.Sprintf("%d bps", bitsPerSec)
	}
	i := int(math.Floor(math.Log10(float64(bitsPerSec)) / 3))
	val := float64(bitsPerSec) / math.Pow(1000, float64(i))
	return fmt.Sprintf("%.2f %s", val, unit[i])
}

// HumanDuration 将时间间隔格式化为可读字符串
// 例如:  90s -> "1m30s"
func HumanDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	if d < time.Minute {
		return fmt.Sprintf("%.0fs", d.Seconds())
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm%ds", int(d.Minutes()), int(d.Seconds())%60)
	}
	return fmt.Sprintf("%dh%dm%ds",
		int(d.Hours()),
		int(d.Minutes())%60,
		int(d.Seconds())%60,
	)
}
