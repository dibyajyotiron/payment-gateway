package services

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
)

type CipherText struct {
	secret []byte
}

func NewCipherText(secretKey string) *CipherText {
	return &CipherText{
		secret: []byte(secretKey),
	}
}

// masks data using AES and secret
func (c *CipherText) MaskData(data []byte) (string, error) {
	block, err := aes.NewCipher(c.secret)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher block: %v", err)
	}

	// Use AES-GCM for authenticated encryption
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %v", err)
	}

	// Generate a nonce (IV) of the appropriate size for AES-GCM
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %v", err)
	}

	// Encrypt the data and append the nonce to the beginning of the ciphertext
	ciphertext := aesGCM.Seal(nonce, nonce, data, nil)

	// Return the encrypted data as a base64-encoded string
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// unmasks data using base64
func (c *CipherText) UnmaskData(maskedData string) ([]byte, error) {
	// Decode the base64-encoded data
	data, err := base64.StdEncoding.DecodeString(maskedData)
	if err != nil {
		return nil, fmt.Errorf("failed to decode data: %v", err)
	}

	// Create a new AES cipher block
	block, err := aes.NewCipher(c.secret)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher block: %v", err)
	}

	// Use AES-GCM for decryption
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %v", err)
	}

	// Split the nonce and ciphertext
	nonceSize := aesGCM.NonceSize()
	if len(data) < nonceSize {
		return nil, fmt.Errorf("invalid encrypted data")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]

	// Decrypt the data
	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt data: %v", err)
	}

	return plaintext, nil
}
