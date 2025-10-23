package fileutil

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
)

// setupTestFS 为单元测试创建一个标准的临时文件/目录结构。
// 它接受 *testing.T 并使用 t.Fatalf 报告错误。
// 返回根目录的路径和一个清理函数。
// 目录结构:
// - /
//   - empty_dir/
//   - sub_dir/
//   - nested_dir/
//   - deep_file.txt (内容: "deep")
//   - sub_file.txt (内容: "sub")
//   - empty_file.txt (内容: "")
//   - regular_file.txt (内容: "hello world")
//   - unwritable_file.txt (权限 0444, 内容: "readonly")
func setupTestFS(t *testing.T) (string, func()) {
	// 创建临时根目录
	rootDir, err := os.MkdirTemp("", "fileutil_test_*")
	if err != nil {
		t.Fatalf("无法创建临时目录: %v", err)
	}

	// 创建目录结构
	emptyDir := filepath.Join(rootDir, "empty_dir")
	subDir := filepath.Join(rootDir, "sub_dir")
	nestedDir := filepath.Join(subDir, "nested_dir")
	if err := os.MkdirAll(nestedDir, 0755); err != nil {
		t.Fatalf("创建目录结构失败: %v", err)
	}
	if err := os.Mkdir(emptyDir, 0755); err != nil {
		t.Fatalf("创建空目录失败: %v", err)
	}

	// 创建文件
	filesToCreate := map[string]string{
		filepath.Join(rootDir, "empty_file.txt"):      "",
		filepath.Join(rootDir, "regular_file.txt"):    "hello world",
		filepath.Join(rootDir, "unwritable_file.txt"): "readonly",
		filepath.Join(subDir, "sub_file.txt"):         "sub",
		filepath.Join(nestedDir, "deep_file.txt"):     "deep",
	}

	for path, content := range filesToCreate {
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("创建文件 %s 失败: %v", path, err)
		}
	}

	// 设置特殊权限
	if err := os.Chmod(filepath.Join(rootDir, "unwritable_file.txt"), 0444); err != nil {
		t.Fatalf("设置文件权限失败: %v", err)
	}

	// 返回根目录路径和清理函数
	cleanup := func() {
		_ = os.RemoveAll(rootDir)
	}
	return rootDir, cleanup
}

// ***** 新增函数 *****
// setupBenchmarkFS 为基准测试创建一个标准的临时文件/目录结构。
// 它接受 *testing.B 并使用 b.Fatalf 报告错误。
// 这是 setupTestFS 的基准测试版本。
func setupBenchmarkFS(b *testing.B) (string, func()) {
	rootDir, err := os.MkdirTemp("", "fileutil_bench_*")
	if err != nil {
		b.Fatalf("无法创建临时目录: %v", err)
	}

	// 创建目录结构
	subDir := filepath.Join(rootDir, "sub_dir")
	nestedDir := filepath.Join(subDir, "nested_dir")
	if err := os.MkdirAll(nestedDir, 0755); err != nil {
		b.Fatalf("创建目录结构失败: %v", err)
	}

	// 创建文件
	filesToCreate := map[string]string{
		filepath.Join(rootDir, "regular_file.txt"): "hello world",
		filepath.Join(subDir, "sub_file.txt"):      "sub",
		filepath.Join(nestedDir, "deep_file.txt"):  "deep",
	}

	for path, content := range filesToCreate {
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			b.Fatalf("创建文件 %s 失败: %v", path, err)
		}
	}

	cleanup := func() {
		_ = os.RemoveAll(rootDir)
	}
	return rootDir, cleanup
}

// --- Tests
func TestExists(t *testing.T) {
	rootDir, cleanup := setupTestFS(t)
	defer cleanup()

	if !Exists(rootDir) {
		t.Errorf("Exists(%q) = false; want true", rootDir)
	}
	if !Exists(filepath.Join(rootDir, "regular_file.txt")) {
		t.Error("Exists(regular_file.txt) = false; want true")
	}
	if Exists(filepath.Join(rootDir, "non_existent_file.txt")) {
		t.Error("Exists(non_existent_file.txt) = true; want false")
	}
}

func TestIsDir(t *testing.T) {
	rootDir, cleanup := setupTestFS(t)
	defer cleanup()

	if !IsDir(rootDir) {
		t.Errorf("IsDir(%q) = false; want true", rootDir)
	}
	if IsDir(filepath.Join(rootDir, "regular_file.txt")) {
		t.Error("IsDir(regular_file.txt) = true; want false")
	}
	if IsDir(filepath.Join(rootDir, "non_existent_dir")) {
		t.Error("IsDir(non_existent_dir) = true; want false")
	}
}

func TestIsFile(t *testing.T) {
	rootDir, cleanup := setupTestFS(t)
	defer cleanup()

	if !IsFile(filepath.Join(rootDir, "regular_file.txt")) {
		t.Error("IsFile(regular_file.txt) = false; want true")
	}
	if IsFile(rootDir) {
		t.Errorf("IsFile(%q) = true; want false", rootDir)
	}
	if IsFile(filepath.Join(rootDir, "non_existent_file.txt")) {
		t.Error("IsFile(non_existent_file.txt) = true; want false")
	}
}

func TestListDir(t *testing.T) {
	rootDir, cleanup := setupTestFS(t)
	defer cleanup()

	files, dirs, err := ListDir(rootDir)
	if err != nil {
		t.Fatalf("ListDir() error = %v", err)
	}

	// 排序以确保测试结果一致性
	sort.Strings(files)
	sort.Strings(dirs)

	expectedFiles := []string{
		filepath.Join(rootDir, "empty_file.txt"),
		filepath.Join(rootDir, "regular_file.txt"),
		filepath.Join(rootDir, "unwritable_file.txt"),
	}
	expectedDirs := []string{
		filepath.Join(rootDir, "empty_dir"),
		filepath.Join(rootDir, "sub_dir"),
	}

	if !reflect.DeepEqual(files, expectedFiles) {
		t.Errorf("ListDir() files = %v, want %v", files, expectedFiles)
	}
	if !reflect.DeepEqual(dirs, expectedDirs) {
		t.Errorf("ListDir() dirs = %v, want %v", dirs, expectedDirs)
	}
}

func TestListDirRecursively(t *testing.T) {
	rootDir, cleanup := setupTestFS(t)
	defer cleanup()

	files, err := ListDirRecursively(rootDir)
	if err != nil {
		t.Fatalf("ListDirRecursively() error = %v", err)
	}
	sort.Strings(files)

	expectedFiles := []string{
		filepath.Join(rootDir, "empty_file.txt"),
		filepath.Join(rootDir, "regular_file.txt"),
		filepath.Join(rootDir, "sub_dir", "nested_dir", "deep_file.txt"),
		filepath.Join(rootDir, "sub_dir", "sub_file.txt"),
		filepath.Join(rootDir, "unwritable_file.txt"),
	}
	sort.Strings(expectedFiles) // 确保期望结果也是排序的

	if !reflect.DeepEqual(files, expectedFiles) {
		t.Errorf("ListDirRecursively() = %v, want %v", files, expectedFiles)
	}
}

func TestFileSize(t *testing.T) {
	rootDir, cleanup := setupTestFS(t)
	defer cleanup()

	size := FileSize(filepath.Join(rootDir, "regular_file.txt"))
	if size != 11 {
		t.Errorf("FileSize(regular_file.txt) = %d; want 11", size)
	}
	size = FileSize(filepath.Join(rootDir, "empty_file.txt"))
	if size != 0 {
		t.Errorf("FileSize(empty_file.txt) = %d; want 0", size)
	}
	size = FileSize(rootDir) // 目录大小应为0
	if size != 0 {
		t.Errorf("FileSize(dir) = %d; want 0", size)
	}
}

func TestFileMode(t *testing.T) {
	rootDir, cleanup := setupTestFS(t)
	defer cleanup()

	mode := FileMode(filepath.Join(rootDir, "unwritable_file.txt"))
	if mode != 0444 {
		t.Errorf("FileMode(unwritable_file.txt) = %v; want %v", mode, os.FileMode(0444))
	}
}

func TestIsReadable(t *testing.T) {
	rootDir, cleanup := setupTestFS(t)
	defer cleanup()

	if !IsReadable(filepath.Join(rootDir, "regular_file.txt")) {
		t.Error("IsReadable(regular_file.txt) = false; want true")
	}
}

func TestIsWritable(t *testing.T) {
	rootDir, cleanup := setupTestFS(t)
	defer cleanup()

	if !IsWritable(filepath.Join(rootDir, "regular_file.txt")) {
		t.Error("IsWritable(regular_file.txt) = false; want true")
	}
	if IsWritable(filepath.Join(rootDir, "unwritable_file.txt")) {
		t.Error("IsWritable(unwritable_file.txt) = true; want false")
	}
}

func TestIsEmpty(t *testing.T) {
	rootDir, cleanup := setupTestFS(t)
	defer cleanup()

	// 测试空文件
	empty, err := IsEmpty(filepath.Join(rootDir, "empty_file.txt"))
	if err != nil || !empty {
		t.Errorf("IsEmpty(empty_file.txt) = %v, %v; want true, nil", empty, err)
	}
	// 测试非空文件
	empty, err = IsEmpty(filepath.Join(rootDir, "regular_file.txt"))
	if err != nil || empty {
		t.Errorf("IsEmpty(regular_file.txt) = %v, %v; want false, nil", empty, err)
	}
	// 测试空目录
	empty, err = IsEmpty(filepath.Join(rootDir, "empty_dir"))
	if err != nil || !empty {
		t.Errorf("IsEmpty(empty_dir) = %v, %v; want true, nil", empty, err)
	}
	// 测试非空目录
	empty, err = IsEmpty(filepath.Join(rootDir, "sub_dir"))
	if err != nil || empty {
		t.Errorf("IsEmpty(sub_dir) = %v, %v; want false, nil", empty, err)
	}
}

func TestSafeRename(t *testing.T) {
	rootDir, cleanup := setupTestFS(t)
	defer cleanup()

	src := filepath.Join(rootDir, "regular_file.txt")
	dst := filepath.Join(rootDir, "renamed_file.txt")

	if err := SafeRename(src, dst); err != nil {
		t.Fatalf("SafeRename() error = %v", err)
	}

	if Exists(src) {
		t.Error("源文件在重命名后仍然存在")
	}
	if !Exists(dst) {
		t.Error("目标文件在重命名后不存在")
	}
}

func TestCopyFile(t *testing.T) {
	rootDir, cleanup := setupTestFS(t)
	defer cleanup()

	src := filepath.Join(rootDir, "regular_file.txt")
	dst := filepath.Join(rootDir, "copied_file.txt")

	if err := CopyFile(src, dst); err != nil {
		t.Fatalf("CopyFile() error = %v", err)
	}

	srcContent, _ := os.ReadFile(src)
	dstContent, _ := os.ReadFile(dst)
	if !bytes.Equal(srcContent, dstContent) {
		t.Error("复制后文件内容不匹配")
	}

	srcInfo, _ := os.Stat(src)
	dstInfo, _ := os.Stat(dst)
	if srcInfo.Mode() != dstInfo.Mode() {
		t.Error("复制后文件权限不匹配")
	}
}

func TestRemoveAllFiles(t *testing.T) {
	rootDir, cleanup := setupTestFS(t)
	defer cleanup()

	targetDir := filepath.Join(rootDir, "sub_dir")
	if err := RemoveAllFiles(targetDir); err != nil {
		t.Fatalf("RemoveAllFiles() error = %v", err)
	}

	empty, err := IsEmpty(targetDir)
	if err != nil || !empty {
		t.Errorf("RemoveAllFiles() 后目录不为空, empty=%v, err=%v", empty, err)
	}
}

func TestIsSameFile(t *testing.T) {
	rootDir, cleanup := setupTestFS(t)
	defer cleanup()

	path1 := filepath.Join(rootDir, "regular_file.txt")
	path2 := filepath.Join(rootDir, "sub_dir", "sub_file.txt")

	if !IsSameFile(path1, path1) {
		t.Error("IsSameFile(path, path) = false; want true")
	}
	if IsSameFile(path1, path2) {
		t.Error("IsSameFile(path1, path2) = true; want false")
	}
}

func TestEnsureDir(t *testing.T) {
	rootDir, cleanup := setupTestFS(t)
	defer cleanup()

	// 测试已存在的目录
	if err := EnsureDir(rootDir, 0755); err != nil {
		t.Errorf("EnsureDir(existing_dir) error = %v", err)
	}
	// 测试创建新目录
	newDir := filepath.Join(rootDir, "new_dir_to_create")
	if err := EnsureDir(newDir, 0755); err != nil {
		t.Errorf("EnsureDir(new_dir) error = %v", err)
	}
	if !IsDir(newDir) {
		t.Error("EnsureDir() 未能创建新目录")
	}
	// 测试路径被文件占用的情况
	filePath := filepath.Join(rootDir, "regular_file.txt")
	if err := EnsureDir(filePath, 0755); err == nil {
		t.Error("EnsureDir(path_is_file) 应该返回错误但没有")
	}
}

func TestDirSize(t *testing.T) {
	rootDir, cleanup := setupTestFS(t)
	defer cleanup()

	subDir := filepath.Join(rootDir, "sub_dir")
	size, err := DirSize(subDir)
	if err != nil {
		t.Fatalf("DirSize() error = %v", err)
	}
	// "sub" (3) + "deep" (4) = 7
	expectedSize := int64(7)
	if size != expectedSize {
		t.Errorf("DirSize() = %d; want %d", size, expectedSize)
	}
}

func TestCopyDir(t *testing.T) {
	rootDir, cleanup := setupTestFS(t)
	defer cleanup()

	srcDir := filepath.Join(rootDir, "sub_dir")
	dstDir := filepath.Join(rootDir, "sub_dir_copy")

	if err := CopyDir(srcDir, dstDir, 0755); err != nil {
		t.Fatalf("CopyDir() error = %v", err)
	}

	// 比较两个目录的大小
	srcSize, _ := DirSize(srcDir)
	dstSize, _ := DirSize(dstDir)
	if srcSize != dstSize {
		t.Errorf("复制后目录大小不匹配: src=%d, dst=%d", srcSize, dstSize)
	}

	// 比较文件列表
	srcFiles, _ := ListDirRecursively(srcDir)
	dstFiles, _ := ListDirRecursively(dstDir)
	if len(srcFiles) != len(dstFiles) {
		t.Errorf("复制后文件数量不匹配: src=%d, dst=%d", len(srcFiles), len(dstFiles))
	}
}

// --- Benchmarks ---

func BenchmarkExists(b *testing.B) {
	// ***** 修改点 *****
	rootDir, cleanup := setupBenchmarkFS(b)
	defer cleanup()
	path := filepath.Join(rootDir, "regular_file.txt")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Exists(path)
	}
}

func BenchmarkIsDir(b *testing.B) {
	// ***** 修改点 *****
	rootDir, cleanup := setupBenchmarkFS(b)
	defer cleanup()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		IsDir(rootDir)
	}
}

func BenchmarkListDir(b *testing.B) {
	rootDir, err := os.MkdirTemp("", "bench_list_dir")
	if err != nil {
		b.Fatal(err)
	}
	defer os.RemoveAll(rootDir)
	for i := 0; i < 100; i++ {
		if err := os.WriteFile(filepath.Join(rootDir, fmt.Sprintf("file%d.txt", i)), []byte("test"), 0644); err != nil {
			b.Fatalf("Failed to write file for benchmark setup: %v", err)
		}
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = ListDir(rootDir)
	}
}

func BenchmarkListDirRecursively(b *testing.B) {
	rootDir, err := os.MkdirTemp("", "bench_list_dir_rec")
	if err != nil {
		b.Fatal(err)
	}
	defer os.RemoveAll(rootDir)
	curDir := rootDir
	for i := 0; i < 10; i++ {
		curDir = filepath.Join(curDir, fmt.Sprintf("dir%d", i))
		if err := os.Mkdir(curDir, 0755); err != nil {
			b.Fatalf("Failed to create dir for benchmark setup: %v", err)
		}
		for j := 0; j < 10; j++ {
			if err := os.WriteFile(filepath.Join(curDir, fmt.Sprintf("file%d.txt", j)), []byte("test"), 0644); err != nil {
				b.Fatalf("Failed to write file for benchmark setup: %v", err)
			}
		}
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ListDirRecursively(rootDir)
	}
}

func BenchmarkFileSize(b *testing.B) {
	// ***** 修改点 *****
	rootDir, cleanup := setupBenchmarkFS(b)
	defer cleanup()
	path := filepath.Join(rootDir, "regular_file.txt")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		FileSize(path)
	}
}

func BenchmarkCopyFile(b *testing.B) {
	src, err := os.CreateTemp("", "bench_copy_src")
	if err != nil {
		b.Fatal(err)
	}
	defer os.Remove(src.Name())
	defer src.Close()

	data := make([]byte, 1024*1024) // 1MB 文件
	if _, err := src.Write(data); err != nil {
		b.Fatal(err)
	}

	dstPath := filepath.Join(os.TempDir(), "bench_copy_dst")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = CopyFile(src.Name(), dstPath)
		_ = os.Remove(dstPath) // 清理以便下次迭代
	}
}

func BenchmarkDirSize(b *testing.B) {
	// ***** 修改点 *****
	rootDir, cleanup := setupBenchmarkFS(b)
	defer cleanup()
	targetDir := filepath.Join(rootDir, "sub_dir")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = DirSize(targetDir)
	}
}

func BenchmarkEnsureDir(b *testing.B) {
	rootDir, err := os.MkdirTemp("", "bench_ensure_dir")
	if err != nil {
		b.Fatal(err)
	}
	defer os.RemoveAll(rootDir)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// 每次迭代都在一个干净的子目录中进行，以模拟“不存在则创建”的常见情况
		path := filepath.Join(rootDir, fmt.Sprintf("dir_%d", i))
		_ = EnsureDir(path, 0755)
	}
}

func BenchmarkCopyDir(b *testing.B) {
	// ***** 修改点 *****
	rootDir, cleanup := setupBenchmarkFS(b)
	defer cleanup()
	srcDir := filepath.Join(rootDir, "sub_dir")
	dstParent := filepath.Join(rootDir, "copy_dest_parent")
	if err := os.Mkdir(dstParent, 0755); err != nil {
		b.Fatalf("Failed to create parent dir for benchmark setup: %v", err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dstDir := filepath.Join(dstParent, fmt.Sprintf("copy_%d", i))
		_ = CopyDir(srcDir, dstDir, 0755)
		// 注意：这里没有清理目标目录，因为连续创建带递增编号的目录
		// 更符合现实中 backup 等场景。如果需要每次都清理，会增加 benchmark 的开销。
	}
}
