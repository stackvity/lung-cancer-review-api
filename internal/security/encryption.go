// internal/security/encryption.go
package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"

	"go.uber.org/zap"
)

// var logger *zap.Logger // Logger instance

func init() {
	logger = zap.L().Named("security")
}

// Encrypt encrypts data using AES-GCM.
//
// Key Management: The `key` parameter should be obtained from a secure key management
// system (e.g., HashiCorp Vault, AWS KMS, Google Cloud KMS, Azure Key Vault).
// *NEVER* hardcode keys or store them directly in environment variables in a production environment.
// The current implementation, reading from an environment variable, is *only* acceptable for
// local development and testing, and *must* be replaced with a proper secrets management solution
// before deployment.  THIS IS A CRITICAL SECURITY REQUIREMENT.
//
// Future Optimization (Out of Scope): For very large files, consider using io.Reader and io.Writer
// to avoid loading the entire file into memory.
func Encrypt(key []byte, data []byte) ([]byte, error) {
	const operation = "security.Encrypt"

	// Input Validation: Key length
	switch len(key) {
	case 16, 24, 32: // Valid key sizes for AES (128, 192, or 256 bits)
	default:
		logger.Error("Invalid key size", zap.String("operation", operation), zap.Int("key_size", len(key)))
		return nil, fmt.Errorf("invalid key size: %d. Key size must be 16, 24, or 32 bytes for AES-128, AES-192, or AES-256", len(key))
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		logger.Error("Failed to create new AES cipher", zap.String("operation", operation), zap.Error(err))
		return nil, fmt.Errorf("creating AES cipher: %w", err) // Wrap error
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		logger.Error("Failed to create GCM", zap.String("operation", operation), zap.Error(err))
		return nil, fmt.Errorf("creating GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		logger.Error("Failed to generate nonce", zap.String("operation", operation), zap.Error(err))
		return nil, fmt.Errorf("generating nonce: %w", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	logger.Debug("Data encrypted successfully", zap.String("operation", operation)) // Log at Debug level
	return ciphertext, nil
}

// Decrypt decrypts data using AES-GCM.
//
// Key Management: (Same critical note as for Encrypt applies here).
func Decrypt(key []byte, ciphertext []byte) ([]byte, error) {
	const operation = "security.Decrypt"

	// Input Validation: Key length
	switch len(key) {
	case 16, 24, 32: // Valid key sizes
	default:
		logger.Error("Invalid key size", zap.String("operation", operation), zap.Int("key_size", len(key)))
		return nil, fmt.Errorf("invalid key size: %d. Key size must be 16, 24, or 32 bytes", len(key))
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		logger.Error("Failed to create AES cipher", zap.String("operation", operation), zap.Error(err))
		return nil, fmt.Errorf("creating AES cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		logger.Error("Failed to create GCM", zap.String("operation", operation), zap.Error(err))
		return nil, fmt.Errorf("creating GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		logger.Error("Invalid ciphertext size", zap.String("operation", operation), zap.Int("size", len(ciphertext)), zap.Int("nonceSize", nonceSize))
		return nil, errors.New("invalid ciphertext size, size is smaller than nonce size") // Use standard errors.New
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		logger.Error("Failed to decrypt data", zap.String("operation", operation), zap.Error(err))
		return nil, fmt.Errorf("decrypting data: %w", err)
	}

	logger.Debug("Data decrypted successfully", zap.String("operation", operation)) // Log at Debug level
	return plaintext, nil
}

// EncryptAndEncode encrypts data and then encodes it using Base64 URL-safe encoding.
func EncryptAndEncode(key []byte, data []byte) (string, error) {
	ciphertext, err := Encrypt(key, data)
	if err != nil {
		return "", err // No need to wrap, Encrypt already wraps.
	}
	return base64.URLEncoding.EncodeToString(ciphertext), nil // URL-safe encoding
}

// DecodeAndDecrypt decodes Base64 URL-safe encoded data and then decrypts it.
func DecodeAndDecrypt(key []byte, encodedCiphertext string) ([]byte, error) {
	ciphertext, err := base64.URLEncoding.DecodeString(encodedCiphertext)
	if err != nil {
		return nil, fmt.Errorf("base64 decoding: %w", err)
	}
	return Decrypt(key, ciphertext) // No need to wrap, Decrypt already wraps.
}
