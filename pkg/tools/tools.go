package tools

import (
	"crypto/md5"
	"encoding/hex"
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

// GetMD5Hash вернет md5 хеш  текста.
func GetMD5Hash(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}
