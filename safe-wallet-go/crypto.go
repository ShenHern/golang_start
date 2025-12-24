package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"

	"golang.org/x/crypto/pbkdf2"
)

const (
	saltSize   = 32
	nonceSize  = 12 // GCM standard nonce size
	keySize    = 32 // AES-256 key size
	iterations = 100000
)

// deriveKey derives an encryption key from a password using PBKDF2
func deriveKey(password string, salt []byte) []byte {
	return pbkdf2.Key([]byte(password), salt, iterations, keySize, sha256.New)
}

// EncryptData encrypts data using AES-GCM with a password-derived key
func EncryptData(data []byte, password string) ([]byte, error) {
	// Generate random salt
	salt := make([]byte, saltSize)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, err
	}

	// Derive key from password
	key := deriveKey(password, salt)

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Create GCM
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Generate random nonce
	nonce := make([]byte, nonceSize)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	// Encrypt and authenticate
	ciphertext := gcm.Seal(nil, nonce, data, nil)

	// Prepend salt and nonce to ciphertext
	encrypted := make([]byte, 0, len(salt)+len(nonce)+len(ciphertext))
	encrypted = append(encrypted, salt...)
	encrypted = append(encrypted, nonce...)
	encrypted = append(encrypted, ciphertext...)

	return encrypted, nil
}

// DecryptData decrypts data using AES-GCM with a password-derived key
func DecryptData(encrypted []byte, password string) ([]byte, error) {
	if len(encrypted) < saltSize+nonceSize {
		return nil, errors.New("encrypted data too short")
	}

	// Extract salt, nonce, and ciphertext
	salt := encrypted[:saltSize]
	nonce := encrypted[saltSize : saltSize+nonceSize]
	ciphertext := encrypted[saltSize+nonceSize:]

	// Derive key from password
	key := deriveKey(password, salt)

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Create GCM
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Decrypt and verify
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, errors.New("decryption failed: invalid password or corrupted data")
	}

	return plaintext, nil
}

// EncryptToBase64 encrypts data and returns it as a base64-encoded string
func EncryptToBase64(data []byte, password string) (string, error) {
	encrypted, err := EncryptData(data, password)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(encrypted), nil
}

// DecryptFromBase64 decrypts data from a base64-encoded string
func DecryptFromBase64(encryptedStr string, password string) ([]byte, error) {
	encrypted, err := base64.StdEncoding.DecodeString(encryptedStr)
	if err != nil {
		return nil, err
	}
	return DecryptData(encrypted, password)
}
