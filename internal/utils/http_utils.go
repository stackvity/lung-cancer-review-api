// internal/utils/http_utils.go
package utils

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

// Logger is a global logger instance (initialized in init()).
// var Logger *zap.Logger

// RequestIDKey is the key used to store the request ID in the context.
// const RequestIDKey = "requestID" // Use a constant for the context key

// RespondWithError sends a JSON error response.  It takes a Gin context,
// an HTTP status code, and an error message (which can be any type).
func RespondWithError(c *gin.Context, code int, message interface{}) {
	c.AbortWithStatusJSON(code, gin.H{"error": message})
}

// RespondWithJSON sends a JSON response with the provided status code and data.
func RespondWithJSON(c *gin.Context, code int, payload interface{}) {
	c.JSON(code, payload)
}

// GetRequestID retrieves the request ID from the context.  It now correctly
// accepts a standard context.Context, *not* a *gin.Context.
func GetRequestID(ctx context.Context) string {
	requestID, ok := ctx.Value(RequestIDKey).(string)
	if !ok {
		// Log a warning; this should not happen in normal operation
		// because RequestIDMiddleware should always set the requestID.
		if Logger != nil { // Check if logger is initialized
			Logger.Warn("requestID not found in context")
		}
		return "" // Return empty string if not found
	}
	return requestID
}

// DetectContentTypeFromFile attempts to determine the content type of a file
// by reading the first 512 bytes of the file and using http.DetectContentType.
func DetectContentTypeFromFile(filename string) (string, error) {
	// Open the file
	file, err := os.Open(filename)
	if err != nil {
		return "", fmt.Errorf("opening file: %w", err)
	}
	defer file.Close() // Ensure the file is closed, even if errors occur

	// Read the first 512 bytes (the maximum size needed by DetectContentType)
	buffer := make([]byte, 512)
	_, err = file.Read(buffer) //Read the file
	// io.EOF is okay here; it means the file is smaller than 512 bytes,
	// and DetectContentType will still work.  We only error for *other* read errors.
	if err != nil && err != io.EOF {
		return "", fmt.Errorf("reading file header: %w", err)
	}

	// Detect the content type.  We use the buffer[:n] slice to pass only the
	// bytes that were actually read from the file.
	contentType := http.DetectContentType(buffer)

	return contentType, nil
}
