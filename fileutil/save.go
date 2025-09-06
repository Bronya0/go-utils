package fileutil

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
)

// SaveFile 保存上传的文件
// 增加了严格的安全校验
// 参数：
// fileHeader *multipart.FileHeader: 上传的文件
// dstPath string: 文件保存的目标路径
// fileType: 文件类型, 如 "application/zip"，可以为空，表示不进行文件类型校验
// expectedHash string: 预期的文件的哈希值，用于严格校验，为空表示不进行校验
func SaveFile(fileHeader *multipart.FileHeader, dstPath, fileType, expectedHash string) error {
	src, err := fileHeader.Open()
	if err != nil {
		return fmt.Errorf("打开上传的文件失败: %w", err)
	}
	defer src.Close()

	// 创建一个临时文件来存储上传的内容
	tempFile, err := os.CreateTemp("", "upload-*.tmp")
	if err != nil {
		return fmt.Errorf("创建临时文件失败: %w", err)
	}
	defer tempFile.Close()
	defer os.Remove(tempFile.Name()) // 确保在函数结束时删除临时文件

	// 将上传文件的内容写入临时文件，同时计算哈希值
	hasher := sha256.New()
	// MultiWriter可以同时向文件和哈希器写入数据
	writer := io.MultiWriter(tempFile, hasher)

	if _, err := io.Copy(writer, src); err != nil {
		return fmt.Errorf("写入临时文件失败: %w", err)
	}

	// 1. 服务端哈希校验
	if expectedHash != "" {
		actualHash := hex.EncodeToString(hasher.Sum(nil))
		if actualHash != expectedHash {
			return fmt.Errorf("文件哈希值不匹配。预期: %s, 实际: %s", expectedHash, actualHash)
		}
	}

	// 将文件指针移回开头，以便进行文件类型检测
	if _, err := tempFile.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("无法重置临时文件指针: %w", err)
	}

	// 2. 文件类型校验 (Magic Number)
	if fileType != "" {
		// 读取文件的前512个字节来判断MIME类型
		buffer := make([]byte, 512)
		n, err := tempFile.Read(buffer)
		if err != nil && err != io.EOF {
			return fmt.Errorf("读取文件头失败: %w", err)
		}
		t := http.DetectContentType(buffer[:n])
		// 校验其是否为 application/zip
		if t != fileType {
			return fmt.Errorf("无效的文件类型。预期: %s, 实际: %s", fileType, t)
		}
	}

	// 3. 安全地保存文件
	// 只有所有校验通过后，才将临时文件移动到最终位置
	if _, err := os.Stat(dstPath); err == nil {
		return fmt.Errorf("文件已存在：%s", dstPath)
	}

	// 关闭临时文件句柄，以便重命名操作
	tempFile.Close()
	if err := os.Rename(tempFile.Name(), dstPath); err != nil {
		return fmt.Errorf("移动文件到持久化存储目录失败: %w", err)
	}

	return nil
}
