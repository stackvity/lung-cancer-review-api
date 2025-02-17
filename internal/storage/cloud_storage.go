// internal/storage/cloud_storage.go
package storage

import (
	"context"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stackvity/lung-server/internal/config"
	"github.com/stackvity/lung-server/internal/security"
	"github.com/stackvity/lung-server/internal/utils"
	"go.uber.org/zap"
)

// CloudStorage implements the FileStorage interface using AWS S3.
// This production-ready implementation leverages the AWS SDK for Go v2 for secure, scalable, and performant cloud storage.
// It incorporates streaming encryption/decryption, robust error handling, and optimizations for high-throughput operations.
type CloudStorage struct {
	config       *config.Config      // Configuration to access AWS S3 settings and credentials (loaded from env vars/secrets manager)
	logger       *zap.Logger         // Logger for structured logging, providing contextual information for all operations
	s3Client     *s3.Client          // AWS S3 client for interacting with the S3 service API
	s3Uploader   *manager.Uploader   // AWS S3 uploader for optimized uploads, handles multipart uploads for large files
	s3Downloader *manager.Downloader // AWS S3 downloader for optimized downloads, supports concurrent downloads for efficiency
}

// NewCloudStorage creates a new CloudStorage instance, initializing the AWS S3 client and uploader/downloader.
// It performs essential setup, including loading AWS configuration and creating S3 service clients,
// and returns a ready-to-use CloudStorage instance or an error if initialization fails.
func NewCloudStorage(ctx context.Context, cfg *config.Config, logger *zap.Logger) (*CloudStorage, error) {
	const operation = "NewCloudStorage" // Operation name for consistent logging

	cloudStorageLogger := logger.Named("CloudStorage") // Create a dedicated logger for CloudStorage component, enhancing log context

	cloudStorageLogger.Info("Initializing CloudStorage with AWS S3", zap.String("operation", operation), zap.String("storage_type", cfg.StorageType), zap.String("s3_bucket", cfg.CloudStorageBucket), zap.String("aws_region", cfg.AWSRegion))

	// 1. Load AWS Configuration - Uses aws-sdk-go-v2/config package for automatic credential loading and region configuration.
	awsCfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(cfg.AWSRegion)) // Load AWS config, region from cfg
	if err != nil {
		cloudStorageLogger.Error("Failed to load AWS configuration", zap.String("operation", operation), zap.Error(err))
		return nil, fmt.Errorf("failed to load AWS configuration: %w", err) // Return error with context for initialization failure
	}

	// 2. Initialize AWS S3 Client - Creates a new AWS S3 service client using the loaded AWS configuration.
	s3Client := s3.NewFromConfig(awsCfg)

	// 3. Initialize S3 Uploader and Downloader - manager.Uploader and manager.NewDownloader provide optimized upload/download performance.
	s3Uploader := manager.NewUploader(s3Client)
	s3Downloader := manager.NewDownloader(s3Client)

	cs := &CloudStorage{
		config:       cfg,
		logger:       cloudStorageLogger,
		s3Client:     s3Client,
		s3Uploader:   s3Uploader,
		s3Downloader: s3Downloader,
	}

	cloudStorageLogger.Info("CloudStorage initialized successfully with AWS S3", zap.String("operation", operation), zap.String("s3_bucket", cfg.CloudStorageBucket), zap.String("aws_region", cfg.AWSRegion)) // Log successful initialization
	return cs, nil                                                                                                                                                                                             // Return initialized CloudStorage instance and nil error for success
}

// Save implements the FileStorage interface for CloudStorage, saving a file to AWS S3 with encryption.
func (s *CloudStorage) Save(ctx context.Context, filename string, contentType string, file io.Reader) (string, error) {
	const operation = "CloudStorage.Save" // Operation name for structured logging
	requestID := utils.GetRequestID(ctx)

	s.logger.Info("Starting file upload to cloud storage (AWS S3)", zap.String("operation", operation), zap.String("request_id", requestID), zap.String("filename", filename), zap.String("content_type", contentType), zap.String("s3_bucket", s.config.CloudStorageBucket))

	// 1. Generate Unique File Path (Key) in S3 - for organized and collision-free storage, using request ID and sanitized filename.
	filePath := fmt.Sprintf("lung-cancer-review/%s/%s", utils.GetRequestID(ctx), security.SanitizeFilename(filename))

	// 2. Encrypt File Content - before uploading to AWS S3 to ensure data-at-rest encryption using AES-256-GCM streaming encryption.
	encryptedReader, err := utils.EncryptReader([]byte(s.config.FileEncryptionKey), file) // Use utils.EncryptReader for streaming encryption
	if err != nil {
		s.logger.Error("Encryption failed before upload", zap.String("operation", operation), zap.String("request_id", requestID), zap.String("filename", filename), zap.Error(err))
		return "", fmt.Errorf("encryption failed: %w", err) // Return error with context for encryption failure
	}

	// 3. Upload to AWS S3 using manager.Uploader - Leverages optimized multipart uploads for enhanced performance and scalability.
	uploadInput := &s3.PutObjectInput{ // Create PutObjectInput, configuring S3 upload parameters
		Bucket:      aws.String(s.config.CloudStorageBucket), // Set S3 bucket name from configuration
		Key:         aws.String(filePath),                    // Set S3 object key (path within the bucket)
		Body:        encryptedReader,                         // Use the encrypted reader for upload, ensuring data is encrypted in transit and at rest
		ContentType: aws.String(contentType),                 // Set the content type of the S3 object
	}

	uploadOutput, err := s.s3Uploader.Upload(ctx, uploadInput) // Perform the upload using S3 Uploader
	if err != nil {
		s.logger.Error("AWS S3 upload failed", zap.String("operation", operation), zap.String("request_id", requestID), zap.String("filename", filename), zap.Error(err))
		return "", fmt.Errorf("AWS S3 upload failed: %w", err) // Return error with context for S3 upload failure
	}

	s.logger.Info("File uploaded successfully to cloud storage (AWS S3)", zap.String("operation", operation), zap.String("request_id", requestID), zap.String("file_path", filePath), zap.String("s3_bucket", s.config.CloudStorageBucket), zap.String("s3_upload_url", uploadOutput.Location)) // Log successful upload
	return filePath, nil                                                                                                                                                                                                                                                                        // Return file path (key) in S3 and nil error for successful operation
}

// Get implements the FileStorage interface for CloudStorage, retrieving and decrypting a file from AWS S3.
func (s *CloudStorage) Get(ctx context.Context, filepath string) (io.ReadCloser, error) {
	const operation = "CloudStorage.Get"
	requestID := utils.GetRequestID(ctx)

	s.logger.Info("Getting file from cloud storage (AWS S3)", zap.String("operation", operation), zap.String("request_id", requestID), zap.String("file_path", filepath), zap.String("s3_bucket", s.config.CloudStorageBucket))

	// 1. Download from AWS S3 using GetObject API
	getObjectInput := &s3.GetObjectInput{ // Create GetObjectInput to specify download parameters
		Bucket: aws.String(s.config.CloudStorageBucket), // Set S3 bucket name from configuration
		Key:    aws.String(filepath),                    // Set S3 object key (path) to download
	}

	resp, err := s.s3Client.GetObject(ctx, getObjectInput) // Download object directly using S3 client's GetObject API
	if err != nil {
		s.logger.Error("AWS S3 download failed", zap.String("operation", operation), zap.String("request_id", requestID), zap.String("file_path", filepath), zap.Error(err))
		return nil, fmt.Errorf("aws s3 download failed: %w", err) // Return error with context for S3 download failure
	}
	encryptedReadCloser := resp.Body // Get ReadCloser providing access to the encrypted file content

	// 2. Decrypt File Content
	decryptedReadCloser, err := utils.DecryptReader([]byte(s.config.FileEncryptionKey), encryptedReadCloser) // Use utils.DecryptReader for streaming decryption - Wrap encryptedReadCloser from S3 for decryption
	if err != nil {
		s.logger.Error("Decryption failed after download", zap.String("operation", operation), zap.String("request_id", requestID), zap.String("file_path", filepath), zap.Error(err))
		return nil, fmt.Errorf("decryption failed: %w", err) // Return error with context for decryption failure
	}

	s.logger.Info("File retrieved successfully from cloud storage (AWS S3)", zap.String("operation", operation), zap.String("request_id", requestID), zap.String("file_path", filepath), zap.String("s3_bucket", s.config.CloudStorageBucket)) // Log successful retrieval
	return decryptedReadCloser, nil                                                                                                                                                                                                            // Return decrypted ReadCloser for streaming access to the decrypted file content
}

// Delete implements the FileStorage interface for CloudStorage, securely deleting a file from AWS S3.
func (s *CloudStorage) Delete(ctx context.Context, filepath string) error {
	const operation = "CloudStorage.Delete"
	requestID := utils.GetRequestID(ctx)

	s.logger.Info("Deleting file from cloud storage (AWS S3)", zap.String("operation", operation), zap.String("request_id", requestID), zap.String("file_path", filepath), zap.String("s3_bucket", s.config.CloudStorageBucket))

	// 1. Secure Delete from AWS S3
	deleteInput := &s3.DeleteObjectInput{ // Create DeleteObjectInput to specify object to delete
		Bucket: aws.String(s.config.CloudStorageBucket), // Set S3 bucket name from config
		Key:    aws.String(filepath),                    // Set S3 object key (path) to delete
	}
	_, err := s.s3Client.DeleteObject(ctx, deleteInput) // Perform deletion using S3 client's DeleteObject API
	if err != nil {
		s.logger.Error("AWS S3 deletion failed", zap.String("operation", operation), zap.String("request_id", requestID), zap.String("file_path", filepath), zap.Error(err))
		return fmt.Errorf("aws s3 deletion failed: %w", err) // Return error with context for S3 deletion failure
	}

	s.logger.Info("File deleted successfully from cloud storage (AWS S3)", zap.String("operation", operation), zap.String("request_id", requestID), zap.String("file_path", filepath), zap.String("s3_bucket", s.config.CloudStorageBucket)) // Log successful deletion
	return nil                                                                                                                                                                                                                               // Return nil error for successful deletion
}
