package tools

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/exp/rand"
)

const (
	Mode0755 = 0o755
	Mode0750 = 0o750
	Mode0600 = 0o600
)

func init() {
	rand.Seed(uint64(time.Now().UnixNano()))
}

// RandomString генерирует строку заданой длины.
func RandomString(n uint) string {
	var letterRunes = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")
	b := make([]byte, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func SaveUploadedFile(data *[]byte, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), Mode0750); err != nil {
		return fmt.Errorf("failed create dir: %w", err)
	}

	buf := bytes.NewBuffer(*data)

	out, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed create file: %w", err)
	}

	_, err = io.Copy(out, buf)
	if err != nil {
		return fmt.Errorf("failed copy: %w", err)
	}

	err = out.Close()
	if err != nil {
		return fmt.Errorf("failed close file: %w", err)
	}

	return nil
}

func GetMD5Hash(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}
