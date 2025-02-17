// internal/storage/storage.go
package storage

import (
	"context"
	"io"
)

// FileStorage defines the interface for file storage operations.
// This interface provides an abstraction layer for different storage implementations
// (e.g., local file system, cloud storage like AWS S3 or Google Cloud Storage).
// Implementations MUST handle:
//   - Concurrent access: Ensure thread-safety and handle concurrent read/write operations correctly. *CRITICAL for application stability and data integrity.*
//   - Error handling: Implement robust error handling for file I/O operations, including disk space errors, permission errors, and network errors (for cloud storage). *MANDATORY for production readiness.*
//   - Security: Ensure secure handling of image data and API keys (if using cloud services). *ESSENTIAL for data privacy and compliance.*
//   - Performance: Optimize for latency and throughput, especially for large files and high request volumes. *IMPORTANT for user experience and scalability.*
//   - Context cancellation: Implementations MUST respect context cancellation and timeouts to prevent resource leaks and ensure responsiveness under load. *Context handling is REQUIRED for production readiness.*
type FileStorage interface {
	// Save stores a file from an io.Reader to the storage backend.
	// It takes a context for cancellation, the filename, content type, and an io.Reader for the file content.
	// Implementations MUST:
	//   - Generate a unique file path for storing the file. Consider using UUIDs or timestamps to ensure uniqueness and prevent filename collisions.
	//   - Handle file saving efficiently, especially for large files. Consider using buffered I/O and streaming techniques to minimize memory footprint and improve performance.
	//   - ENCRYPT the file content *before* writing to storage. Use strong encryption algorithms (e.g., AES-256-GCM) and secure key management practices. *Encryption is MANDATORY for protecting patient data at rest.*
	//   - Return the full file path (or URI/URL for cloud storage) where the file is stored. This path will be used to retrieve the file later.
	//   - Implement robust error handling, returning specific error types (defined in `internal/domain/errors.go`) for different failure scenarios (e.g., disk space errors, permission errors, network errors, encryption failures). *ROBUST ERROR HANDLING IS MANDATORY.*
	//   - Log file saving operations with sufficient detail for auditing and debugging, including filename, content type, and any errors encountered.
	//
	// Parameters:
	//   - ctx context.Context: Context for cancellation and timeout. Implementations MUST respect context deadlines and cancellations to prevent resource leaks and ensure responsiveness under load.
	//   - filename string:  The desired filename for the file. Implementations SHOULD sanitize this filename to prevent path traversal vulnerabilities and ensure compatibility with the underlying storage system.
	//   - contentType string: MIME type or content type of the file. This information can be used for storing metadata or setting appropriate headers when serving files.
	//   - file io.Reader:  An io.Reader from which the file content can be read. This allows the Save method to handle files of any size without loading the entire file into memory at once (important for memory efficiency and handling large files).
	//
	// Returns:
	//   - string: The full file path (or URI/URL) where the file is stored. This path is used to retrieve the file later using the Get method.
	//   - error: An error if file saving fails. Implementations MUST return `nil` on successful save, and a *specific, custom error type* from `internal/domain/errors.go` on failure, such as `domain.ErrFileSaveFailed` for disk space or permission errors, or `domain.ErrEncryptionFailed` for encryption-related issues.
	//
	// Example usage:
	// func someService(fs FileStorage, file io.Reader, filename, contentType string) error {
	//     filePath, err := fs.Save(ctx context.Background(), filename, contentType, file)
	//     if err != nil {
	//         return fmt.Errorf("failed to save file: %w", err)
	//     }
	//     log.Printf("File saved to: %s", filePath)
	//     return nil
	// }
	Save(ctx context.Context, filename string, contentType string, file io.Reader) (string, error) // Returns file path

	// Get retrieves a file as an io.ReadCloser from the storage backend, given its file path.
	// It takes a context for cancellation and the file path as input.
	// Implementations MUST:
	//   - Retrieve the file content from the storage backend based on the provided file path.
	//   - DECRYPT the file content *after* retrieving it from storage, before returning the io.ReadCloser. Use the corresponding decryption method for the encryption algorithm used in the Save method. *Decryption is MANDATORY for accessing encrypted patient data.*
	//   - Return an io.ReadCloser to allow streaming reading of the file content, especially for large files. This avoids loading the entire file into memory.
	//   - Implement robust error handling, returning specific error types (defined in `internal/domain/errors.go`) for different failure scenarios (e.g., file not found, permission errors, network errors, decryption failures). *ROBUST ERROR HANDLING IS MANDATORY.*
	//   - Log file retrieval operations with sufficient detail for auditing and debugging, including file path and any errors encountered.
	//
	// Parameters:
	//   - ctx context.Context: Context for cancellation and timeout. Implementations MUST respect context deadlines and cancellations to prevent resource leaks and ensure responsiveness under load.
	//   - filepath string: The full file path (or URI/URL) of the file to retrieve. This path SHOULD be the same path returned by the Save method. Implementations MUST validate the filepath to prevent unauthorized access or path traversal vulnerabilities.
	//
	// Returns:
	//   - io.ReadCloser: An io.ReadCloser that provides access to the decrypted file content. The caller is responsible for closing the ReadCloser after reading the file.
	//   - error: An error if file retrieval fails. Implementations MUST return `nil` on successful retrieval and a *specific, custom error type* from `internal/domain/errors.go` on failure, such as `domain.ErrFileNotFound` if the file does not exist, or `domain.ErrDecryptionFailed` for decryption errors.
	//
	// Example usage:
	// func someHandler(fs FileStorage, filePath string) error {
	//    readCloser, err := fs.Get(ctx context.Background(), filePath)
	//    if err != nil {
	//        return fmt.Errorf("failed to get file: %w", err)
	//    }
	//    defer readCloser.Close() // Important: close the ReadCloser after use
	//    // ... process the file content from readCloser ...
	//}
	Get(ctx context.Context, filepath string) (io.ReadCloser, error)

	// Delete securely deletes a file from the storage backend, given its file path.
	// It takes a context for cancellation and the file path as input.
	// Implementations MUST:
	//   - Securely delete the file from the storage backend. For local file storage, this might involve overwriting the file data multiple times before unlinking (using `internal/security.SecureDeleteFile`). For cloud storage, use the service's secure deletion mechanisms (e.g., S3 DeleteObject, Google Cloud Storage Delete). *SECURE DELETION IS MANDATORY for compliance and data privacy.*
	//   - Implement robust error handling for file deletion operations, including file not found errors, permission errors, and network errors (for cloud storage). *ROBUST ERROR HANDLING IS MANDATORY.*
	//   - Log file deletion operations with sufficient detail for auditing and debugging, including file path and any errors encountered.
	//
	// Parameters:
	//   - ctx context.Context: Context for cancellation and timeout. Implementations MUST respect context deadlines and cancellations to prevent resource leaks and ensure responsiveness under load.
	//   - filepath string: The full file path (or URI/URL) of the file to delete. This path SHOULD be the same path returned by the Save method. Implementations MUST validate the filepath to prevent accidental deletion of unintended files and path traversal vulnerabilities.
	//
	// Returns:
	//   - error: An error if file deletion fails. Implementations MUST return `nil` on successful deletion and a *specific, custom error type* from `internal/domain/errors.go` on failure, such as `domain.ErrFileDeleteFailed` for permission or file system errors.
	//
	// Example usage:
	// func cleanupService(fs FileStorage, filePath string) error {
	//     if err := fs.Delete(ctx context.Background(), filePath); err != nil {
	//         return fmt.Errorf("failed to delete file: %w", err)
	//     }
	//     log.Printf("File deleted successfully: %s", filePath)
	//     return nil
	// }
	Delete(ctx context.Context, filepath string) error
}
