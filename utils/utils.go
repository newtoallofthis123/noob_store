package utils

import (
	"crypto/sha256"
	"fmt"
)

func CalHash(content []byte) string {
	hash := sha256.New()
	hash.Write(content)
	return fmt.Sprintf("%x", hash.Sum(nil))
}
