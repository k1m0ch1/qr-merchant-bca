package client

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
)

const (
	// Encryption key discovered from JavaScript analysis
	encryptionKey = "9C0XAVRJ6PQB86TVTAD6SK6XD01PSCIK"
)

// encryptPassword encrypts a password using AES-256-CBC encryption
// This matches the "encryptMessi" function from the JavaScript
func encryptPassword(plaintext string) (string, error) {
	// Convert key to bytes
	key := []byte(encryptionKey)

	// Create cipher block
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	// Add PKCS7 padding
	paddedPlaintext := pkcs7Pad([]byte(plaintext), aes.BlockSize)

	// Generate random IV
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", fmt.Errorf("failed to generate IV: %w", err)
	}

	// Encrypt using CBC mode
	ciphertext := make([]byte, len(paddedPlaintext))
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext, paddedPlaintext)

	// Prepend IV to ciphertext (standard practice for CBC mode)
	result := append(iv, ciphertext...)

	// Base64 encode the result
	return base64.StdEncoding.EncodeToString(result), nil
}

// pkcs7Pad adds PKCS7 padding to the data
func pkcs7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - (len(data) % blockSize)
	padText := make([]byte, padding)
	for i := range padText {
		padText[i] = byte(padding)
	}
	return append(data, padText...)
}
