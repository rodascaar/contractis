package utils

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
)

// CalculateFileHash calcula el hash SHA256 de un archivo
func CalculateFileHash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", fmt.Errorf("error calculating hash: %w", err)
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}
