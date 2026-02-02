// Package service provides business logic services.
package service

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
)

// ErrInvalidCiphertext is returned when decryption fails due to invalid ciphertext.
var ErrInvalidCiphertext = errors.New("invalid ciphertext")

// CryptoService defines the interface for encryption/decryption operations.
type CryptoService interface {
	Encrypt(plaintext string) (string, error)
	Decrypt(ciphertext string) (string, error)
}

// cryptoService implements CryptoService using AES-256-GCM.
type cryptoService struct {
	key []byte
}

// NewCryptoService creates a new CryptoService instance.
// The key should be a base64-encoded 32-byte key for AES-256.
func NewCryptoService(base64Key string) (CryptoService, error) {
	if base64Key == "" {
		return nil, errors.New("encryption key is required")
	}

	key, err := base64.StdEncoding.DecodeString(base64Key)
	if err != nil {
		return nil, fmt.Errorf("failed to decode encryption key: %w", err)
	}

	if len(key) != 32 {
		return nil, fmt.Errorf("encryption key must be 32 bytes, got %d", len(key))
	}

	return &cryptoService{key: key}, nil
}

// Encrypt encrypts the plaintext using AES-256-GCM and returns a base64-encoded ciphertext.
func (s *cryptoService) Encrypt(plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}

	block, err := aes.NewCipher(s.key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// Generate random nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt and prepend nonce to ciphertext
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)

	// Return base64-encoded ciphertext
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decrypts the base64-encoded ciphertext and returns the plaintext.
func (s *cryptoService) Decrypt(ciphertext string) (string, error) {
	if ciphertext == "" {
		return "", nil
	}

	// Decode base64
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("failed to decode ciphertext: %w", err)
	}

	block, err := aes.NewCipher(s.key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// Extract nonce from the beginning of the ciphertext
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", ErrInvalidCiphertext
	}

	nonce, ciphertextBytes := data[:nonceSize], data[nonceSize:]

	// Decrypt
	plaintext, err := gcm.Open(nil, nonce, ciphertextBytes, nil)
	if err != nil {
		return "", ErrInvalidCiphertext
	}

	return string(plaintext), nil
}
