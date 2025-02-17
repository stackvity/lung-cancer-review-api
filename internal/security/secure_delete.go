// internal/security/secure_delete.go
package security

import (
	"crypto/rand"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"go.uber.org/zap"
)

// var logger *zap.Logger // Logger instance

func init() {
	logger = zap.L().Named("security")
}

// SecureDeleteFile securely deletes a file by overwriting it with random data
// multiple times before removing it.  This is a best-effort approach.  True
// secure deletion depends on the underlying storage medium and is often not
// fully guaranteed (especially on SSDs with wear leveling).
//
// OS-Specific Tools: For higher assurance, consider using OS-specific tools
// like `shred` (Linux) or `sdelete` (Windows), but these are OUTSIDE THE SCOPE
// of this basic implementation and would require platform-specific code.
func SecureDeleteFile(filePath string) error {
	const operation = "security.SecureDeleteFile"

	// Input Validation: Basic path traversal check.
	if filepath.IsAbs(filePath) || strings.Contains(filePath, "..") {
		logger.Error("Invalid file path (potential path traversal)", zap.String("operation", operation), zap.String("file_path", filePath))
		return fmt.Errorf("invalid file path: %s", filePath) // Don't reveal too much in the error message
	}
	// SUGGESTION: Use a temp file directory to ensure the file path is in the right directory.

	fileInfo, err := os.Stat(filePath) // Use os.Stat, not os.Open
	if err != nil {
		logger.Error("Failed to retrieve file information", zap.String("operation", operation), zap.String("file_path", filePath), zap.Error(err))
		return fmt.Errorf("failed to stat file: %w", err)
	}

	if fileInfo.IsDir() {
		logger.Error("Path points to directory, not a file", zap.String("operation", operation), zap.String("file_path", filePath))
		return fmt.Errorf("path is a directory, not a file")
	}

	file, err := os.OpenFile(filePath, os.O_WRONLY, 0) // Open for writing only
	if err != nil {
		logger.Error("Failed to open file for overwriting", zap.String("operation", operation), zap.String("file_path", filePath), zap.Error(err))
		return fmt.Errorf("failed to open file for writing: %w", err)
	}
	defer file.Close() // Ensure file is closed

	fileSize := fileInfo.Size()

	// Overwrite the file multiple times with random data.  The optimal number of
	// passes depends on the sensitivity of the data and storage medium.  3
	// passes is often considered a reasonable balance between security and
	// performance.  More passes increase security but take longer.
	for i := 0; i < 3; i++ {
		_, err = io.CopyN(file, rand.Reader, fileSize) // Overwrite entire file
		if err != nil {
			logger.Error("Error overwriting file with random data", zap.String("operation", operation), zap.String("file_path", filePath), zap.Error(err), zap.Int("pass", i+1))
			return fmt.Errorf("error overwriting file (pass %d): %w", i+1, err)
		}
		if err := file.Sync(); err != nil { // Ensure data written to disk.
			logger.Error("Failed to sync file to disk after overwriting", zap.String("operation", operation), zap.String("file_path", filePath), zap.Error(err), zap.Int("pass", i+1))
			return fmt.Errorf("error syncing file after overwrite (pass %d): %w", i+1, err)
		}
		// Reset read/write pointer to the start of file
		if _, err := file.Seek(0, 0); err != nil {
			logger.Error("Failed to seek file to the beginning", zap.String("operation", operation), zap.String("file_path", filePath), zap.Error(err), zap.Int("pass", i+1))
			return fmt.Errorf("failed to seek to beginning of file (pass %d): %w", i+1, err)
		}
	}

	// Truncate the file to zero length (optional, some sources recommend this)
	if err := file.Truncate(0); err != nil {
		logger.Error("Failed to truncate the file", zap.String("operation", operation), zap.String("file_path", filePath), zap.Error(err))
		return fmt.Errorf("failed to truncate file: %w", err)
	}
	// Close before deleting, ensure file is closed
	if err = file.Close(); err != nil {
		logger.Error("Failed close file before removing", zap.String("operation", operation), zap.String("file_path", filePath), zap.Error(err))
		return fmt.Errorf("failed close file before removing : %w", err)
	}

	// Remove the file.
	if err := os.Remove(filePath); err != nil {
		logger.Error("Failed to remove file", zap.String("operation", operation), zap.String("file_path", filePath), zap.Error(err))
		return fmt.Errorf("failed to remove file: %w", err)
	}
	logger.Debug("File deleted successfully", zap.String("operation", operation), zap.String("file_path", filePath))

	return nil
}
