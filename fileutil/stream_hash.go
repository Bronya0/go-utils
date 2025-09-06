package fileutil

import (
	"bytes"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"fmt"
	"hash"
	"io"
	"os"
)

const (
	MD5    = "md5"
	SHA1   = "sha1"
	SHA256 = "sha256"
	SHA512 = "sha512"
)

// 根据算法名返回 hash.Hash
func getHashFunc(alg string) (hash.Hash, error) {
	switch alg {
	case MD5:
		return md5.New(), nil
	case SHA1:
		return sha1.New(), nil
	case SHA256:
		return sha256.New(), nil
	case SHA512:
		return sha512.New(), nil
	default:
		return nil, fmt.Errorf("unsupported hash algorithm: %s", alg)
	}
}

// 流式哈希，alg为"md5"|"sha1"|"sha256"|"sha512"
func HashReader(r io.Reader, alg string) (string, error) {
	h, err := getHashFunc(alg)
	if err != nil {
		return "", err
	}
	buf := make([]byte, 32*1024)
	if _, err := io.CopyBuffer(h, r, buf); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

// 针对 []byte 流式哈希。alg为"md5"|"sha1"|"sha256"|"sha512"
func HashBytes(data []byte, alg string) (string, error) {
	return HashReader(bytes.NewReader(data), alg)
}

// 针对 文件 流式哈希。alg为"md5"|"sha1"|"sha256"|"sha512"
func HashFile(path string, alg string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	return HashReader(f, alg)
}
