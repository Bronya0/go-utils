package validator

import (
	"encoding/json"
	"net"
	"net/mail"
	"net/url"
	"regexp"
	"strconv"
	"time"
	"unicode/utf8"
)

// 我们将所有需要用到的正则表达式在包初始化时编译一次。
var (
	// chineseMobileRegexp 校验中国大陆手机号的正则表达式。
	chineseMobileRegexp = regexp.MustCompile(`^1[3-9]\d{9}$`)

	// internationalPhoneRegexp 校验通用国际电话号码（E.164标准）的正则表达式。
	internationalPhoneRegexp = regexp.MustCompile(`^\+[1-9]\d{1,14}$`)

	// uuidRegexp 校验UUID的正则表达式。
	uuidRegexp = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[1-5][0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}$`)

	// chineseIDCardRegexp 校验中国18位身份证号格式的正则表达式。
	chineseIDCardRegexp = regexp.MustCompile(`^\d{17}(\d|X)$`)

	// md5Regexp 校验MD5哈希值的正则表达式 (32个十六进制字符)。
	md5Regexp = regexp.MustCompile(`^[a-fA-F0-9]{32}$`)

	// sha1Regexp 校验SHA1哈希值的正则表达式 (40个十六进制字符)。
	sha1Regexp = regexp.MustCompile(`^[a-fA-F0-9]{40}$`)

	// sha256Regexp 校验SHA256哈希值的正则表达式 (64个十六进制字符)。
	sha256Regexp = regexp.MustCompile(`^[a-fA-F0-9]{64}$`)

	// sha512Regexp 校验SHA512哈希值的正则表达式 (128个十六进制字符)。
	sha512Regexp = regexp.MustCompile(`^[a-fA-F0-9]{128}$`)
)

// IsIPv4 校验字符串是否为合法的 IPv4 地址。
// @param ip string: 待校验的字符串。
// @return bool: 如果是合法的 IPv4 地址则返回 true，否则返回 false。
func IsIPv4(ip string) bool {
	parsedIP := net.ParseIP(ip)
	// net.ParseIP 能解析 IPv4 和 IPv6，通过 To4() != nil 来确认是 IPv4。
	// 同时要排除 ::ffff:127.0.0.1 这种 IPv4 映射的 IPv6 地址。
	return parsedIP != nil && parsedIP.To4() != nil && parsedIP.To16() != nil && parsedIP.To4().String() == ip
}

// IsIPv6 校验字符串是否为合法的 IPv6 地址。
// @param ip string: 待校验的字符串。
// @return bool: 如果是合法的 IPv6 地址则返回 true，否则返回 false。
func IsIPv6(ip string) bool {
	parsedIP := net.ParseIP(ip)
	// 如果 To4() 为 nil 且 To16() 不为 nil，则说明是 IPv6 地址。
	return parsedIP != nil && parsedIP.To4() == nil && parsedIP.To16() != nil
}

// IsIP 校验字符串是否为合法的 IP 地址（包括 IPv4 和 IPv6）。
// @param ip string: 待校验的字符串。
// @return bool: 如果是合法的 IP 地址则返回 true，否则返回 false。
func IsIP(ip string) bool {
	return net.ParseIP(ip) != nil
}

// IsPort 校验字符串是否为合法的端口号（0-65535）。
// @param portStr string: 待校验的端口字符串。
// @return bool: 如果是合法的端口号则返回 true，否则返回 false。
func IsPort(portStr string) bool {
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return false
	}
	return port >= 0 && port <= 65535
}

// IsURL 校验字符串是否为合法的 URL。
// 要求包含 scheme (如 "http") 和 host。
// @param urlStr string: 待校验的 URL 字符串。
// @return bool: 如果是合法的 URL 则返回 true，否则返回 false。
func IsURL(urlStr string) bool {
	u, err := url.Parse(urlStr)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return false
	}
	return true
}

// IsURI 校验字符串是否为合法的 URI。
// URI的要求比URL更宽泛，例如 "mailto:user@example.com" 是合法的URI但不是标准的URL。
// @param uriStr string: 待校验的 URI 字符串。
// @return bool: 如果是合法的 URI 则返回 true，否则返回 false。
func IsURI(uriStr string) bool {
	_, err := url.ParseRequestURI(uriStr)
	return err == nil
}

// IsEmail 校验字符串是否为合法的邮箱地址。
// 使用标准库 `net/mail`，它遵循 RFC 5322 标准，非常准确。
// @param email string: 待校验的邮箱字符串。
// @return bool: 如果是合法的邮箱地址则返回 true，否则返回 false。
func IsEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

// IsChineseMobile 校验字符串是否为合法的中国大陆手机号。
// 规则：以1开头的11位数字，第二位为3-9。
// @param mobile string: 待校验的手机号字符串。
// @return bool: 如果是合法的中国大陆手机号则返回 true，否则返回 false。
func IsChineseMobile(mobile string) bool {
	return chineseMobileRegexp.MatchString(mobile)
}

// IsInternationalPhone 校验字符串是否为通用的国际电话号码格式（E.164 标准）。
// 规则：以+号开头，后跟1到15位数字。
// @param phone string: 待校验的国际电话号码字符串。
// @return bool: 如果是合法的国际电话号码则返回 true，否则返回 false。
func IsInternationalPhone(phone string) bool {
	return internationalPhoneRegexp.MatchString(phone)
}

// IsGender 校验字符串是否为合法的性别标识。
// @param gender string: 待校验的性别字符串，接受 "male", "female", "unknown"。
// @return bool: 如果是则返回 true，否则返回 false。
func IsGender(gender string) bool {
	switch gender {
	case "male", "female", "unknown":
		return true
	default:
		return false
	}
}

// IsStringLengthInRange 校验字符串长度（兼容中英文数字混合）是否在指定范围内。
// 注意：这里计算的是字符数（rune），而不是字节数（byte）。
// @param s string: 待校验的字符串。
// @param min int: 最小长度（包含）。
// @param max int: 最大长度（包含）。
// @return bool: 如果长度在范围内则返回 true，否则返回 false。
func IsStringLengthInRange(s string, min, max int) bool {
	length := utf8.RuneCountInString(s)
	return length >= min && length <= max
}

// IsDateTime 校验字符串是否符合指定的日期时间格式。
// @param dtStr string: 待校验的日期时间字符串。
// @param layout string: Go语言的时间格式化模板，例如 "2006-01-02 15:04:05"。
// @return bool: 如果符合指定格式则返回 true，否则返回 false。
func IsDateTime(dtStr, layout string) bool {
	_, err := time.Parse(layout, dtStr)
	return err == nil
}

// IsUUID 校验字符串是否为合法的 UUID。
// @param uuid string: 待校验的 UUID 字符串。
// @return bool: 如果是合法的 UUID 则返回 true，否则返回 false。
func IsUUID(uuid string) bool {
	return uuidRegexp.MatchString(uuid)
}

// IsJSON 校验字符串是否为合法的 JSON 格式。
// @param jsonStr string: 待校验的 JSON 字符串。
// @return bool: 如果是合法的 JSON 字符串则返回 true，否则返回 false。
func IsJSON(jsonStr string) bool {
	return json.Valid([]byte(jsonStr))
}

// IsPasswordStrong 校验密码强度。
// 规则：长度至少8位，并且必须包含大写字母、小写字母、数字和特殊字符中的至少三种。
// @param password string: 待校验的密码字符串。
// @return bool: 如果密码强度足够则返回 true，否则返回 false。
func IsPasswordStrong(password string) bool {
	if utf8.RuneCountInString(password) < 8 {
		return false
	}
	level := 0
	if ok, _ := regexp.MatchString(`[a-z]`, password); ok {
		level++
	}
	if ok, _ := regexp.MatchString(`[A-Z]`, password); ok {
		level++
	}
	if ok, _ := regexp.MatchString(`[0-9]`, password); ok {
		level++
	}
	if ok, _ := regexp.MatchString(`[\W_]`, password); ok { // \W 匹配非单词字符
		level++
	}
	return level >= 3
}

// IsMAC 校验字符串是否为合法的 MAC 地址。
// @param mac string: 待校验的 MAC 地址字符串。
// @return bool: 如果是合法的 MAC 地址则返回 true，否则返回 false。
func IsMAC(mac string) bool {
	_, err := net.ParseMAC(mac)
	return err == nil
}

// IsChineseIDCard 校验字符串是否为合法的中国大陆18位身份证号。
// 该函数会校验格式和最后的校验码，但不会校验地址码和日期的真实有效性。
// @param id string: 待校验的身份证号字符串。
// @return bool: 如果是合法的身份证号则返回 true，否则返回 false。
func IsChineseIDCard(id string) bool {
	// 1. 格式校验
	if !chineseIDCardRegexp.MatchString(id) {
		return false
	}

	// 2. 校验码校验
	var (
		// ISO 7064:1983, MOD 11-2
		weight = []int{7, 9, 10, 5, 8, 4, 2, 1, 6, 3, 7, 9, 10, 5, 8, 4, 2}
		check  = []byte{'1', '0', 'X', '9', '8', '7', '6', '5', '4', '3', '2'}
		sum    = 0
	)

	for i := 0; i < 17; i++ {
		// '0' 的 ASCII 码是 48
		sum += int(id[i]-'0') * weight[i]
	}

	return check[sum%11] == id[17]
}

// IsMD5 校验字符串是否为合法的MD5哈希值。
// 规则：32个十六进制字符 (不区分大小写)。
// @param s string: 待校验的字符串。
// @return bool: 如果是合法的MD5哈希值则返回 true，否则返回 false。
func IsMD5(s string) bool {
	return md5Regexp.MatchString(s)
}

// IsSHA1 校验字符串是否为合法的SHA1哈希值。
// 规则：40个十六进制字符 (不区分大小写)。
// @param s string: 待校验的字符串。
// @return bool: 如果是合法的SHA1哈希值则返回 true，否则返回 false。
func IsSHA1(s string) bool {
	return sha1Regexp.MatchString(s)
}

// IsSHA256 校验字符串是否为合法的SHA256哈希值。
// 规则：64个十六进制字符 (不区分大小写)。
// @param s string: 待校验的字符串。
// @return bool: 如果是合法的SHA256哈希值则返回 true，否则返回 false。
func IsSHA256(s string) bool {
	return sha256Regexp.MatchString(s)
}

// IsSHA512 校验字符串是否为合法的SHA512哈希值。
// 规则：128个十六进制字符 (不区分大小写)。
// @param s string: 待校验的字符串。
// @return bool: 如果是合法的SHA512哈希值则返回 true，否则返回 false。
func IsSHA512(s string) bool {
	return sha512Regexp.MatchString(s)
}
