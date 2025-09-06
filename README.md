# SafeUtils
Golang安全的操作函数

## 使用
```go
// go get github.com/Bronya0/SafeUtils
import "github.com/Bronya0/SafeUtils"
```

## 函数说明

### archiveutil 包
- UnzipSafe: 安全解压ZIP文件，防御路径遍历、解压炸弹等攻击

### fileutil 包
- SaveFile: 安全保存上传文件，包含文件类型和哈希值校验
- HashReader: 对io.Reader进行流式哈希计算
- HashBytes: 对字节切片进行流式哈希计算
- HashFile: 对文件进行流式哈希计算

### strutil 包
- JoinStr: 高效字符串拼接，使用strings.Builder避免中间对象过多
