package fileutil

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// ZipFiles 安全地将多个源文件压缩到一个目标 ZIP 文件中。
//
// destZipPath: 目标 ZIP 文件的路径。如果文件已存在，它将被覆盖。
// srcFilePaths: 一个或多个源文件的路径（可变参数）。
//
// 该函数具有以下特点：
// 1. 安全性：只将文件的基本名（如 "report.docx"）存入 ZIP，避免泄露本地文件结构。
// 2. 高效性：使用流式处理（io.Copy），内存占用低，适合处理大文件。
// 3. 健壮性：对输入进行校验，确保只处理文件，不处理目录，并正确管理资源。
func ZipFiles(destZipPath string, srcFilePaths ...string) (rerr error) {
	if destZipPath == "" {
		return errors.New("目标 zip 路径不能为空")
	}
	if len(srcFilePaths) == 0 {
		return errors.New("至少提供一个源文件")
	}

	out, err := os.OpenFile(destZipPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("创建 zip 文件失败: %w", err)
	}
	defer func() {
		if cerr := out.Close(); cerr != nil && rerr == nil {
			rerr = cerr
		}
	}()

	zw := zip.NewWriter(out)
	defer func() {
		if cerr := zw.Close(); cerr != nil && rerr == nil {
			rerr = cerr
		}
	}()

	buf := make([]byte, 256*1024) // 提前分配缓冲区，避免多次分配

	for _, filePath := range srcFilePaths {
		info, err := os.Stat(filePath)
		if err != nil {
			return fmt.Errorf("无法访问文件 '%s': %w", filePath, err)
		}
		if info.IsDir() {
			return fmt.Errorf("'%s' 是目录，请使用 ZipDir", filePath)
		}

		src, err := os.Open(filePath)
		if err != nil {
			return fmt.Errorf("打开文件 '%s' 失败: %w", filePath, err)
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			src.Close()
			return fmt.Errorf("创建 header 失败: %w", err)
		}

		header.Name = filepath.Base(filePath)
		header.Method = zip.Deflate

		// 使用自定义压缩等级（可调整性能 vs 压缩比）
		writer, err := zw.CreateHeader(header)
		if err != nil {
			src.Close()
			return fmt.Errorf("创建 zip 条目失败: %w", err)
		}

		if _, err := io.CopyBuffer(writer, src, buf); err != nil {
			src.Close()
			return fmt.Errorf("写入 zip 内容失败: %w", err)
		}
		src.Close()
	}

	return nil
}

// ZipDir — 递归压缩目录（安全+路径检查）
func ZipDir(folderPath, destPath string) (rerr error) {
	info, err := os.Stat(folderPath)
	if err != nil {
		return fmt.Errorf("无法访问目录 '%s': %w", folderPath, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("路径 '%s' 不是目录", folderPath)
	}

	out, err := os.OpenFile(destPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("创建目标文件失败: %w", err)
	}
	defer func() {
		if cerr := out.Close(); cerr != nil && rerr == nil {
			rerr = cerr
		}
	}()

	zw := zip.NewWriter(out)
	defer func() {
		if cerr := zw.Close(); cerr != nil && rerr == nil {
			rerr = cerr
		}
	}()

	buf := make([]byte, 256*1024)

	err = filepath.WalkDir(folderPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(folderPath, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(filepath.Clean(rel)) // 清洗路径
		if rel == "." {
			return nil
		}

		if d.IsDir() {
			_, err = zw.Create(rel + "/")
			return err
		}

		info, err := d.Info()
		if err != nil {
			return err
		}
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		header.Name = rel
		header.Method = zip.Deflate

		// 使用可配置压缩等级（性能优化点）
		header.Modified = info.ModTime()
		writer, err := zw.CreateHeader(header)
		if err != nil {
			return err
		}

		src, err := os.Open(path)
		if err != nil {
			return err
		}
		_, err = io.CopyBuffer(writer, src, buf)
		src.Close()
		return err
	})
	return err
}

// UnzipSafe 是一个经过安全加固的解压函数。
// 它能有效防御路径遍历（Zip Slip）、解压炸弹（Zip Bomb）、
// 符号链接攻击、不安全的文件权限以及非预期的文件类型（如管道、设备文件）。
//
// 参数:
//
//	source: zip 压缩包的文件路径。
//	destination: 解压目标目录。
//	maxSize: 允许解压的总大小上限（字节）。
//	maxFiles: 允许解压的文件数量上限。
func UnzipSafe(source, destination string, maxSize int64, maxFiles int) error {
	r, err := zip.OpenReader(source)
	if err != nil {
		return err
	}
	defer r.Close()

	// 确保目标目录存在，权限为 0755
	if err := os.MkdirAll(destination, 0755); err != nil {
		return err
	}

	var totalSize int64
	var fileCount int

	for _, f := range r.File {
		// [安全策略] 1. 检查文件数量是否超限
		fileCount++
		if fileCount > maxFiles {
			return fmt.Errorf("解压失败：文件数量超过限制 (%d)", maxFiles)
		}

		// [安全策略] 2. 预检查单个文件解压后的大小（基于头信息）
		// 防止单个文件就构成解压炸弹。
		if f.UncompressedSize64 > uint64(maxSize) {
			return fmt.Errorf("解压失败：文件 '%s' 的未压缩大小 (%d) 超过了总限制 (%d bytes)", f.Name, f.UncompressedSize64, maxSize)
		}

		// [安全策略] 3. 防御路径遍历（Zip Slip）攻击
		filePath := filepath.Join(destination, f.Name)
		// 清理目标路径，确保它是一个绝对且干净的路径
		cleanDest := filepath.Clean(filePath)
		// 检查清理后的路径是否仍然在预期的基础目录内
		if !strings.HasPrefix(cleanDest, filepath.Clean(destination)+string(os.PathSeparator)) && cleanDest != filepath.Clean(destination) {
			return fmt.Errorf("不安全的压缩文件路径: %s", f.Name)
		}

		// [安全策略] 4. 禁止解压符号链接，防止指向任意位置
		// f.Mode() 返回的是 zip 包头中记录的权限和模式位
		if f.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("检测到不安全的符号链接，已禁止: %s", f.Name)
		}

		// 处理目录
		if f.FileInfo().IsDir() {
			// [安全策略] 5. 为目录强制设置安全权限 (0755)
			if err := os.MkdirAll(filePath, 0755); err != nil {
				return err
			}
			continue
		}

		// [安全策略] 6. 只允许解压常规文件
		// 防止创建命名管道(FIFO)、套接字(Socket)、设备文件等特殊文件。
		if !f.Mode().IsRegular() {
			return fmt.Errorf("检测到不安全的文件类型 (非常规文件)，已禁止: %s", f.Name)
		}

		// 为文件创建父目录，同样使用安全权限
		if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
			return err
		}

		// 使用匿名函数 + defer 来确保文件句柄被正确关闭
		err = func() error {
			// [安全策略] 7. 为文件强制设置安全权限 (0644)
			// O_TRUNC: 如果文件已存在则清空
			outFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
			if err != nil {
				return err
			}
			defer outFile.Close()

			rc, err := f.Open()
			if err != nil {
				return err
			}
			defer rc.Close()

			// [安全策略] 8. 限制读取的数据量，防止头信息欺诈
			// 确保实际写入的总大小不会超过 maxSize。
			remainingSize := maxSize - totalSize
			limitedReader := io.LimitReader(rc, remainingSize+1) // 多读一个字节用于检测是否超限

			// [安全策略] 9. 使用 io.CopyN 精确控制写入量，并累加真实解压大小
			written, err := io.CopyN(outFile, limitedReader, remainingSize+1)
			if err != nil && err != io.EOF { // io.EOF 在这里是正常情况
				return err
			}

			if written > remainingSize {
				return fmt.Errorf("解压失败：解压后总大小超过限制 (%d bytes)", maxSize)
			}

			totalSize += written
			return nil
		}()

		if err != nil {
			return err
		}
	}

	return nil
}

// IsZipFile 函数
func IsZipFile(filepath string) (bool, error) {
	f, err := os.Open(filepath)
	if err != nil {
		return false, err
	}
	defer f.Close()

	buf := make([]byte, 4)
	if _, err := io.ReadFull(f, buf); err != nil {
		// 如果文件小于4字节，ReadFull会返回ErrUnexpectedEOF，这是正常情况
		if errors.Is(err, io.ErrUnexpectedEOF) || err == io.EOF {
			return false, nil
		}
		return false, err
	}

	return bytes.Equal(buf, []byte("PK\x03\x04")), nil
}
