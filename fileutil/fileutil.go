package fileutil

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

// Exists 判断文件/目录是否存在。
// 注意，权限不足时也认为文件不存在，保守策略
func Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// IsDir 判断路径是否为目录
func IsDir(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// IsFile 判断路径是否为普通文件
func IsFile(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

// ListDir 非递归读取目录。返回文件列表和目录列表
func ListDir(dirPath string) ([]string, []string, error) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, nil, err
	}

	var files []string
	var dirs []string

	for _, entry := range entries {
		fullPath := filepath.Join(dirPath, entry.Name())
		if entry.IsDir() {
			dirs = append(dirs, fullPath)
		} else {
			files = append(files, fullPath)
		}
	}

	return files, dirs, nil
}

// ListDirRecursively 递归遍历目录。返回找到的所有文件的列表。
func ListDirRecursively(root string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		// 只添加文件，并跳过目录
		if !d.IsDir() {
			files = append(files, path)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return files, nil
}

// FileSize 获取文件大小（字节数）
func FileSize(path string) int64 {
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		return 0
	}
	return info.Size()
}

// FileMode 获取文件权限（unix风格）
func FileMode(path string) os.FileMode {
	info, err := os.Stat(path)
	if err != nil {
		return 0
	}
	return info.Mode()
}

// IsReadable 判断文件是否可读
func IsReadable(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	_ = f.Close()
	return true
}

// IsWritable 判断文件是否可写
func IsWritable(path string) bool {
	// 首先，文件必须存在
	info, err := os.Stat(path)
	if err != nil {
		// 不存在或无法访问，都算不可写
		return false
	}
	// 不能是目录
	if info.IsDir() {
		return false
	}

	// 尝试以只写模式打开已存在的文件
	f, err := os.OpenFile(path, os.O_WRONLY, 0)
	if err != nil {
		return false
	}
	f.Close()
	return true
}

// IsEmpty 检查文件或目录是否为空。
// 对于文件，检查其大小是否为 0。
// 对于目录，检查其中是否没有任何条目。
func IsEmpty(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		return false, err
	}

	if info.IsDir() {
		d, err := os.Open(path)
		if err != nil {
			return false, err
		}
		defer d.Close()

		// 读取目录中的一个条目，如果为空，则返回 io.EOF
		_, err = d.Readdir(1)
		if err == io.EOF {
			return true, nil
		}
		return false, err
	}

	// 对于文件，检查大小
	return info.Size() == 0, nil
}

// SafeRename 原子替换文件 (Linux/Unix)
func SafeRename(src, dst string) error {
	return os.Rename(src, dst) // Unix 下是原子操作，Windows 不是
}

// CopyFile 高性能拷贝文件 (支持大文件、零拷贝)
func CopyFile(src, dst string) error {
	sourceInfo, err := os.Stat(src)
	if err != nil {
		return err
	}
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	// 使用 sendfile/零拷贝 (Linux/macOS)
	if _, err := io.Copy(destFile, sourceFile); err != nil {
		return err
	}

	// 复制文件权限
	if err := os.Chmod(dst, sourceInfo.Mode()); err != nil {
		return err
	}

	return destFile.Sync()
}

// RemoveAllFiles 删除目录下所有文件但保留目录本身
func RemoveAllFiles(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if err := os.RemoveAll(filepath.Join(dir, entry.Name())); err != nil {
			return err
		}
	}
	return nil
}

// IsSameFile 判断两个路径是否是同一个文件 (inode)
func IsSameFile(path1, path2 string) bool {
	info1, err1 := os.Stat(path1)
	info2, err2 := os.Stat(path2)
	if err1 != nil || err2 != nil {
		return false
	}
	return os.SameFile(info1, info2)
}

// EnsureDir 确保目录存在（不存在则创建）
func EnsureDir(dir string, perm os.FileMode) error {
	if Exists(dir) {
		if !IsDir(dir) {
			return errors.New("path exists but not a directory: " + dir)
		}
		return nil
	}
	return os.MkdirAll(dir, perm)
}

// DirSize 计算目录总大小
func DirSize(path string) (int64, error) {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size, err
}

// CopyDir 递归复制目录
func CopyDir(srcPath string, dstPath string, mode os.FileMode) error {
	if mode == 0 {
		mode = 0755
	}
	srcInfo, err := os.Stat(srcPath)
	if err != nil {
		return fmt.Errorf("failed to get source directory info: %w", err)
	}

	if !srcInfo.IsDir() {
		return fmt.Errorf("source path is not a directory: %s", srcPath)
	}

	err = os.MkdirAll(dstPath, mode)
	if err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	entries, err := os.ReadDir(srcPath)
	if err != nil {
		return fmt.Errorf("failed to read source directory: %w", err)
	}

	for _, entry := range entries {
		srcDir := filepath.Join(srcPath, entry.Name())
		dstDir := filepath.Join(dstPath, entry.Name())

		if entry.IsDir() {
			err := CopyDir(srcDir, dstDir, mode)
			if err != nil {
				return err
			}
		} else {
			err := CopyFile(srcDir, dstDir)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
