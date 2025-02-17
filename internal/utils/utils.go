// internal/utils/utils.go
package utils

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
)

type contextKey string // Unexported type

const (
	RequestIDKey contextKey = "requestID"
)

// GenerateURLSafeToken generates a URL-safe, base64 encoded, cryptographically secure random string.
func GenerateURLSafeToken() (string, error) {
	b := make([]byte, 32)                 // 32 bytes will result in a 44-character string after base64 and padding.
	_, err := io.ReadFull(rand.Reader, b) // Read random bytes.
	if err != nil {
		// More specific custom error is not strictly needed here, existing error is clear
		return "", fmt.Errorf("failed to generate random bytes: %w", err) // Wrapped for context
	}
	return base64.URLEncoding.EncodeToString(b), nil // URL-safe base64 encoding.
}

// encryptedReader is a struct that wraps an io.Reader and performs AES-GCM streaming encryption.
type encryptedReader struct {
	plaintextReader io.Reader
	gcm             cipher.AEAD
	nonce           []byte
	done            bool // Flag to ensure nonce is prepended only once
}

// Read implements the io.Reader interface for encryptedReader.
func (er *encryptedReader) Read(p []byte) (n int, error error) {
	if !er.done {
		nonceLen := len(er.nonce)
		if len(p) < nonceLen {
			return 0, fmt.Errorf("buffer too small for nonce") // Handle small buffers
		}
		copy(p[:nonceLen], er.nonce)
		er.done = true
		return nonceLen, nil
	}
	buffer := make([]byte, len(p)) // Create a temporary buffer to read plaintext data into
	n, err := er.plaintextReader.Read(buffer)
	if err != nil {
		return n, err
	}
	// Seal and copy encrypted data to output buffer, handling potential small reads.
	// Encrypt only the portion of buffer that contains read data (up to n bytes).
	ciphertext := er.gcm.Seal(p[:0], er.nonce, buffer[:n], nil) // Use Seal directly on chunks
	copy(p[:len(ciphertext)], ciphertext)                       // Copy ciphertext to output buffer
	return len(ciphertext), nil                                 // Return length of ciphertext
}

// EncryptReader creates a reader that encrypts data read from the input reader using AES-GCM streaming encryption.
func EncryptReader(key []byte, plaintextReader io.Reader) (io.Reader, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("creating cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("creating GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("generating nonce: %w", err)
	}

	// Return the custom encryptedReader that handles nonce prepending and streaming encryption.
	return &encryptedReader{ // FIXED: Return the encryptedReader directly, not io.MultiReader
		plaintextReader: plaintextReader, // Renamed field for clarity // FIXED: Corrected field name to plaintextReader to match struct definition
		gcm:             gcm,
		nonce:           nonce,
		done:            false,
	}, nil
}

// decryptReaderWrapper implements io.ReadCloser interface for decryption. // FIXED: Renamed struct to decryptReaderWrapper
type decryptReaderWrapper struct {
	io.ReadCloser   // Embed io.ReadCloser to handle Close()
	gcm             cipher.AEAD
	nonce           []byte
	nonceSize       int
	offsetCounter   int
	ciphertextSize  int
	plaintextBuffer *bytes.Buffer // Buffer to store decrypted plaintext
}

// Read reads decrypted data and implements io.Reader for decryptReaderWrapper. // FIXED: Renamed receiver type to rcw decryptReaderWrapper
func (rcw *decryptReaderWrapper) Read(p []byte) (n int, err error) { // FIXED: Renamed receiver type to rcw decryptReaderWrapper
	if rcw.plaintextBuffer.Len() > 0 {
		return rcw.plaintextBuffer.Read(p)
	}

	chunkSize := 1024 // Adjust chunk size as needed for performance
	ciphertextChunk := make([]byte, chunkSize)
	n, err = rcw.ReadCloser.Read(ciphertextChunk) // Read ciphertext chunk

	if err != nil && err != io.EOF {
		return 0, fmt.Errorf("error reading ciphertext chunk: %w", err)
	}

	if n > 0 {
		// Decrypt chunk
		plaintextChunk, openErr := rcw.gcm.Open(nil, rcw.nonce, ciphertextChunk[:n], nil)
		if openErr != nil {
			return 0, fmt.Errorf("error decrypting chunk: %w", openErr)
		}
		_, err = rcw.plaintextBuffer.Write(plaintextChunk) // Write decrypted chunk to buffer
		if err != nil {
			return 0, fmt.Errorf("error buffering decrypted chunk: %w", err)
		}
	}

	bytesRead, err := rcw.plaintextBuffer.Read(p) // Read from buffer to output slice p
	if err != nil && err != io.EOF {
		return bytesRead, fmt.Errorf("error reading from buffer: %w", err)
	}

	return bytesRead, err // Return bytes read and any error (including io.EOF)

}

// Close implements io.Closer and closes the embedded ReadCloser for decryptReaderWrapper. // FIXED: Renamed receiver type to rcw decryptReaderWrapper
func (rcw *decryptReaderWrapper) Close() error { // FIXED: Renamed receiver type to rcw decryptReaderWrapper
	return rcw.ReadCloser.Close()
}

// DecryptReader creates a reader that decrypts data read from the input reader using AES-GCM streaming decryption.
func DecryptReader(key []byte, ciphertextReadCloser io.ReadCloser) (io.ReadCloser, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("creating cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("creating GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	nonce := make([]byte, nonceSize)

	ciphertext := new(bytes.Buffer)
	_, err = io.CopyN(ciphertext, ciphertextReadCloser, int64(nonceSize))
	if err != nil {
		return nil, fmt.Errorf("reading nonce: %w", err)
	}
	copy(nonce, ciphertext.Bytes()) //copy nonce to byte array

	return &decryptReaderWrapper{ //CHANGED: use decryptReaderWrapper // FIXED: Corrected to use decryptReaderWrapper (renamed struct)
		ReadCloser:      ciphertextReadCloser,
		gcm:             gcm,
		nonce:           nonce,
		nonceSize:       nonceSize,
		offsetCounter:   0,
		ciphertextSize:  0,
		plaintextBuffer: bytes.NewBuffer(nil), // Initialize plaintextBuffer
	}, nil
}
