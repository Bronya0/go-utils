package validator

import (
	"testing"
)

// --- IP 地址相关基准测试 ---

func BenchmarkIsIPv4(b *testing.B) {
	// b.N 是由测试框架动态调整的循环次数
	for i := 0; i < b.N; i++ {
		IsIPv4("192.168.1.1")
	}
}

func BenchmarkIsIPv6(b *testing.B) {
	for i := 0; i < b.N; i++ {
		IsIPv6("2001:0db8:85a3:0000:0000:8a2e:0370:7334")
	}
}

func BenchmarkIsIP(b *testing.B) {
	for i := 0; i < b.N; i++ {
		IsIP("192.168.1.1")
	}
}

// --- 网络相关基准测试 ---
func BenchmarkIsPort(b *testing.B) {
	for i := 0; i < b.N; i++ {
		IsPort("8080")
	}
}

func BenchmarkIsURL(b *testing.B) {
	for i := 0; i < b.N; i++ {
		IsURL("https://www.example.com/path?query=value")
	}
}

func BenchmarkIsURI(b *testing.B) {
	for i := 0; i < b.N; i++ {
		IsURI("mailto:user@example.com")
	}
}

func BenchmarkIsMAC(b *testing.B) {
	for i := 0; i < b.N; i++ {
		IsMAC("00:00:5e:00:53:01")
	}
}

// --- 联系方式相关基准测试 ---
func BenchmarkIsEmail(b *testing.B) {
	for i := 0; i < b.N; i++ {
		IsEmail("test.user+alias@example.com")
	}
}

func BenchmarkIsChineseMobile(b *testing.B) {
	for i := 0; i < b.N; i++ {
		IsChineseMobile("13800138000")
	}
}

func BenchmarkIsInternationalPhone(b *testing.B) {
	for i := 0; i < b.N; i++ {
		IsInternationalPhone("+14155552671")
	}
}

// --- 字符串与格式基准测试 ---
func BenchmarkIsGender(b *testing.B) {
	for i := 0; i < b.N; i++ {
		IsGender("female")
	}
}

func BenchmarkIsStringLengthInRange(b *testing.B) {
	str := "你好世界hello world" // 16个字符
	for i := 0; i < b.N; i++ {
		IsStringLengthInRange(str, 10, 20)
	}
}

func BenchmarkIsDateTime(b *testing.B) {
	layout := "2006-01-02 15:04:05"
	dtStr := "2023-10-27 10:00:00"
	for i := 0; i < b.N; i++ {
		IsDateTime(dtStr, layout)
	}
}

func BenchmarkIsUUID(b *testing.B) {
	for i := 0; i < b.N; i++ {
		IsUUID("a4f4f7a3-4a1f-4a0b-8e1e-4b6e8b4e0e4b")
	}
}

func BenchmarkIsJSON(b *testing.B) {
	jsonStr := `{"name":"test", "age":25, "isStudent": true}`
	for i := 0; i < b.N; i++ {
		IsJSON(jsonStr)
	}
}

// --- 安全与身份标识基准测试 ---

func BenchmarkIsPasswordStrong(b *testing.B) {
	for i := 0; i < b.N; i++ {
		IsPasswordStrong("Str0ngP@ssw0rd!")
	}
}

func BenchmarkIsChineseIDCard(b *testing.B) {
	// 这是一个合法的身份证号码
	for i := 0; i < b.N; i++ {
		IsChineseIDCard("11010119900307221X")
	}
}
