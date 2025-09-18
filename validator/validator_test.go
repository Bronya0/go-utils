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

func BenchmarkIsMD5(b *testing.B) {
	// 一个合法的MD5哈希值
	md5 := "d41d8cd98f00b204e9800998ecf8427e"
	for i := 0; i < b.N; i++ {
		IsMD5(md5)
	}
}

func BenchmarkIsSHA1(b *testing.B) {
	// 一个合法的SHA1哈希值
	sha1 := "da39a3ee5e6b4b0d3255bfef95601890afd80709"
	for i := 0; i < b.N; i++ {
		IsSHA1(sha1)
	}
}

func BenchmarkIsSHA256(b *testing.B) {
	// 一个合法的SHA256哈希值
	sha256 := "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
	for i := 0; i < b.N; i++ {
		IsSHA256(sha256)
	}
}

func BenchmarkIsSHA512(b *testing.B) {
	// 一个合法的SHA512哈希值
	sha512 := "cf83e1357eefb8bdf1542850d66d8007d620e4050b5715dc83f4a921d36ce9ce47d0d13c5d85f2b0ff8318d2877eec2f63b931bd47417a81a538327af927da3e"
	for i := 0; i < b.N; i++ {
		IsSHA512(sha512)
	}
}
