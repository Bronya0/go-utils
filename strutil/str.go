package strutil

import (
	"strings"
)

// JoinStr 高效拼接字符串，内部使用 strings.Builder 避免中间对象过多。
// 推荐在频繁拼接场景下替代 "+" 或 fmt.Sprintf。
func JoinStr(parts ...string) string {
	var sb strings.Builder
	// 预分配，避免反复扩容
	totalLen := 0
	for _, p := range parts {
		totalLen += len(p)
	}
	sb.Grow(totalLen)

	for _, p := range parts {
		sb.WriteString(p)
	}
	return sb.String()
}
