# Golang Safe Utils
Golang安全的操作函数

## 使用
```go
// go get github.com/Bronya0/go-utils
// 按需导入
import(
    "github.com/Bronya0/go-utils/ziputil"
    "github.com/Bronya0/go-utils/fileutil"
    "github.com/Bronya0/go-utils/strutil"
)
```

## 函数说明

### ziputil 包
- `UnzipSafe`: 安全解压ZIP文件，防御路径遍历、解压炸弹等攻击

### fileutil 包
- `SaveFile`: 安全保存上传文件，包含文件类型和哈希值校验
- `HashReader`: 对io.Reader进行流式哈希计算
- `HashBytes`: 对字节切片进行流式哈希计算
- `HashFile`: 对文件进行流式哈希计算

### strutil 包
- `JoinStr`: 高效字符串拼接，使用strings.Builder避免中间对象过多

### container 包
- `NewSet`: 创建set集合。提供了完善的操作API

### validator 包
- 提供了常用的高性能、准确的校验函数