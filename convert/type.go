package convert

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

// ============== 字符串 -- 数字 ================

// StrToInt 安全转换字符串到 int
func StrToInt(s string) (int, error) {
	i64, err := strconv.ParseInt(strings.TrimSpace(s), 10, 0)
	return int(i64), err
}

// MustStrToInt 忽略错误，失败返回 0
func MustStrToInt(s string) int {
	i, _ := StrToInt(s)
	return i
}

func StrToInt64(s string) (int64, error) {
	return strconv.ParseInt(strings.TrimSpace(s), 10, 64)
}

func StrToUint64(s string) (uint64, error) {
	return strconv.ParseUint(strings.TrimSpace(s), 10, 64)
}

func StrToFloat64(s string) (float64, error) {
	return strconv.ParseFloat(strings.TrimSpace(s), 64)
}

func IntToStr(i int) string {
	return strconv.Itoa(i)
}

func Int64ToStr(i int64) string {
	return strconv.FormatInt(i, 10)
}

// Float64ToStr (带精度控制)
func Float64ToStr(f float64, prec int) string {
	return strconv.FormatFloat(f, 'f', prec, 64)
}

// ============== 布尔 -- 其他 ================

// StrToBool 支持 "true/false/1/0/yes/no"
func StrToBool(s string) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "1", "true", "yes", "on":
		return true, nil
	case "0", "false", "no", "off":
		return false, nil
	}
	return false, fmt.Errorf("invalid bool string: %s", s)
}

func BoolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func IntToBool(i int) bool {
	return i != 0
}

// ============== 切片 -- map ================

// StringsToInts []string -> []int
func StringsToInts(arr []string) ([]int, error) {
	res := make([]int, len(arr))
	for i, v := range arr {
		n, err := StrToInt(v)
		if err != nil {
			return nil, err
		}
		res[i] = n
	}
	return res, nil
}

// IntsToStrings []int -> []string
func IntsToStrings(arr []int) []string {
	res := make([]string, len(arr))
	for i, v := range arr {
		res[i] = IntToStr(v)
	}
	return res
}

// MapStrAnyToJSON 转换 map[string]any -> JSON 字符串
func MapStrAnyToJSON(m map[string]any) (string, error) {
	b, err := json.Marshal(m)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// JSONToMapStrAny 转换 JSON -> map[string]any
func JSONToMapStrAny(s string) (map[string]any, error) {
	var m map[string]any
	err := json.Unmarshal([]byte(s), &m)
	return m, err
}

// ============== 通用 ================

// ToString 任意类型转 string
func ToString(v interface{}) string {
	switch val := v.(type) {
	case string:
		return val
	case int, int8, int16, int32, int64:
		return fmt.Sprintf("%d", val)
	case uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", val)
	case float32, float64:
		return fmt.Sprintf("%v", val)
	case bool:
		return strconv.FormatBool(val)
	case time.Time:
		return val.Format(time.RFC3339)
	default:
		b, _ := json.Marshal(val)
		return string(b)
	}
}

// ToInt64 任意类型转 int64
func ToInt64(v interface{}) (int64, error) {
	switch val := v.(type) {
	case int:
		return int64(val), nil
	case int64:
		return val, nil
	case uint64:
		if val > math.MaxInt64 {
			return 0, fmt.Errorf("overflow: %d", val)
		}
		return int64(val), nil
	case float64:
		return int64(val), nil
	case string:
		return StrToInt64(val)
	default:
		return 0, fmt.Errorf("unsupported type: %T", v)
	}
}
