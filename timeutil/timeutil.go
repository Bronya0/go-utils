package timeutil

import (
	"fmt"
	"time"
)

// DefaultLayout 定义了最常用的时间格式
const DefaultLayout = "2006-01-02 15:04:05"
const DefaultLayout2 = "2006/01/02 15:04:05"
const DefaultLayout3 = "20060102150405"

const DateLayout = "2006-01-02"
const DateLayout2 = "2006.01.02"
const DateLayout3 = "20060102"

const DateLayout4 = "2006/01/02"

// DateLayout5 "02/01/2006" 这个格式是有歧义的。
// 根据 Go 的解析规则，它代表 日/月/年。
// 但在美国，这个格式通常表示 月/日/年。
const DateLayout5 = "02/01/2006"

const OnlyTime = "15:04:05"
const OnlyHourMinute = "15:04"

// commonLayouts 存储了一系列常见的时间格式，用于自动解析
var commonLayouts = []string{
	DefaultLayout,
	DefaultLayout2,
	DefaultLayout3,
	DateLayout,
	DateLayout2,
	DateLayout3,
	DateLayout4,
	//DateLayout5,
	time.RFC3339,   // "2006-01-02T15:04:05Z07:00"
	time.RFC822,    // "02 Jan 06 15:04 MST"
	OnlyTime,       // 仅时间
	OnlyHourMinute, // 小时:分钟
}

// =================================================================================
// 1. 获取当前时间戳
// =================================================================================

// NowSeconds 秒级时间戳 (int64)
func NowSeconds() int64 { return time.Now().Unix() }

// NowMillis 毫秒级时间戳 (int64)
func NowMillis() int64 { return time.Now().UnixMilli() }

// NowMicro 微秒级时间戳 (int64)
func NowMicro() int64 { return time.Now().UnixMicro() }

// NowNanos 纳秒级时间戳 (int64)
func NowNanos() int64 { return time.Now().UnixNano() }

// =================================================================================
// 2. 时间类型转换 (字符串/时间戳 <--> time.Time)
// =================================================================================

// Format 将 time.Time 格式化为 "2006-01-02 15:04:05" 格式
func Format(t time.Time) string {
	return t.Format(DefaultLayout)
}

// FormatWithLayout 自定义格式的format为字符串
func FormatWithLayout(t time.Time, layout string) string {
	return t.Format(layout)
}

// FromSeconds 将秒级时间戳转换为 time.Time (使用本地时区)
func FromSeconds(sec int64) time.Time {
	return time.Unix(sec, 0)
}

// FromMillis 将毫秒级时间戳转换为 time.Time (使用本地时区)
func FromMillis(msec int64) time.Time {
	return time.Unix(msec/1000, (msec%1000)*1000000)
}

// ParseString 使用指定的布局和本地时区解析时间字符串
func ParseString(layout, value string) (time.Time, error) {
	return time.ParseInLocation(layout, value, time.Local)
}

// ParseStringAuto 自动尝试多种常见格式来解析时间字符串 (使用本地时区)
func ParseStringAuto(value string) (time.Time, error) {
	for _, layout := range commonLayouts {
		if t, err := time.ParseInLocation(layout, value, time.Local); err == nil {
			return t, nil
		}
	}
	// 如果所有格式都失败，返回一个明确的错误
	return time.Time{}, fmt.Errorf("unable to parse time: %s", value)
}

// =================================================================================
// 3. 关键时间点计算 (通用函数)
// =================================================================================

// DayStart 某一天零点
func DayStart(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}

// DayEnd 某一天 23:59:59.999999999
func DayEnd(t time.Time) time.Time {
	return DayStart(t).Add(24*time.Hour - time.Nanosecond)
}

// WeekStart 获取指定时间所在周的开始时间 (周一, 00:00:00)
func WeekStart(t time.Time) time.Time {
	startOfDay := DayStart(t)
	weekday := int(startOfDay.Weekday())
	// 在 Go 中，周日是 0，周一是 1...
	if weekday == 0 { // 如果是周日
		weekday = 7
	}
	// 计算需要往前推的天数
	offset := time.Duration(weekday-1) * 24 * time.Hour
	return startOfDay.Add(-offset)
}

// WeekEnd 获取指定时间所在周的结束时间 (周日, 23:59:59...)
func WeekEnd(t time.Time) time.Time {
	return WeekStart(t).AddDate(0, 0, 7).Add(-time.Nanosecond)
}

// MonthStart 某月第一天零点
func MonthStart(t time.Time) time.Time {
	y, m, _ := t.Date()
	return time.Date(y, m, 1, 0, 0, 0, 0, t.Location())
}

// MonthEnd 某月最后一天 23:59:59.999999999
func MonthEnd(t time.Time) time.Time {
	return MonthStart(t).AddDate(0, 1, 0).Add(-time.Nanosecond)
}

// YearStart 某年1月1日零点
func YearStart(t time.Time) time.Time {
	y, _, _ := t.Date()
	return time.Date(y, time.January, 1, 0, 0, 0, 0, t.Location())
}

// YearEnd 某年12月31日 23:59:59.999999999
func YearEnd(t time.Time) time.Time {
	return YearStart(t).AddDate(1, 0, 0).Add(-time.Nanosecond)
}

// --- 快捷函数 ---

func TodayStart() time.Time { return DayStart(time.Now()) }
func TodayEnd() time.Time   { return DayEnd(time.Now()) }

func YesterdayStart() time.Time { return DayStart(time.Now().AddDate(0, 0, -1)) }
func YesterdayEnd() time.Time   { return DayEnd(time.Now().AddDate(0, 0, -1)) }
func TomorrowStart() time.Time  { return DayStart(time.Now().AddDate(0, 0, 1)) }
func TomorrowEnd() time.Time    { return DayEnd(time.Now().AddDate(0, 0, 1)) }

func ThisWeekStart() time.Time { return WeekStart(time.Now()) }
func ThisWeekEnd() time.Time   { return WeekEnd(time.Now()) }

func ThisMonthStart() time.Time { return MonthStart(time.Now()) }
func ThisMonthEnd() time.Time   { return MonthEnd(time.Now()) }

func ThisYearStart() time.Time { return YearStart(time.Now()) }
func ThisYearEnd() time.Time   { return YearEnd(time.Now()) }
func LastYearStart() time.Time { return YearStart(time.Now().AddDate(-1, 0, 0)) }
func LastYearEnd() time.Time   { return YearEnd(time.Now().AddDate(-1, 0, 0)) }
