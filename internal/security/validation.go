// internal/security/validation.go
package security

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/stackvity/lung-server/internal/config"
	"github.com/stackvity/lung-server/internal/utils" // Corrected import: utils package is now correctly imported
	"go.uber.org/zap"                                 // Import zap for logging
)

var validate *validator.Validate // Global validator instance for struct validation
var cfg *config.Config           // Global configuration instance, loaded on init

var logger *zap.Logger // Global logger for validation operations

func init() {
	logger = zap.L().Named("validation") // Initialize logger with "validation" namespace for context

	validate = validator.New() // Create a new validator instance

	// Register custom validation functions with the validator.
	// These custom validators extend the validation capabilities beyond the built-in validators.
	if err := validate.RegisterValidation("dicomuid", ValidateDICOMUID); err != nil {
		logger.Error("Failed to register dicomuid validator", zap.Error(err))
	}
	if err := validate.RegisterValidation("bloodpressure", ValidateBloodPressure); err != nil {
		logger.Error("Failed to register bloodpressure validator", zap.Error(err))
	}

	// Load configuration to access validation parameters like MaxFileSize.
	tempCfg, err := config.LoadConfig(context.Background(), ".")
	cfg = &tempCfg // Assign loaded config to the package-level cfg variable
	if err != nil {
		logger.Warn("Failed to load config for validation, using defaults", zap.Error(err))
		// Note: Validation functions that rely on config (like ValidateFileSize) will use default values if config loading fails.
	}
}

// NewValidator creates a new validator instance.  (Corrected: Capital 'N' - Exported Function)
func NewValidator() *validator.Validate { // Corrected: Capital 'N' - Exported Function
	return validator.New()
}

// ValidateStruct validates a struct using the go-playground/validator library.
func ValidateStruct(s interface{}) error {
	const operation = "security.ValidateStruct" // Define operation name for logging
	logger.Debug("Validating struct", zap.String("operation", operation))

	err := validate.Struct(s)
	if err != nil {
		// Type assertion to check if the error is a validator.ValidationErrors type.
		validationErrors, ok := err.(validator.ValidationErrors)
		if ok {
			// Handle validation errors and convert them to a more API-friendly format (ValidationErrors).
			processedErrors := utils.HandleValidationError(validationErrors)                                        // Call HandleValidationError from utils package
			logger.Warn("Struct validation failed", zap.String("operation", operation), zap.Error(processedErrors)) // Log validation errors at Warn level
			return processedErrors
		}
		// For non-validation errors, wrap and return with context.
		validationErr := fmt.Errorf("struct validation failed: %w", err)
		logger.Error("Unexpected validation error", zap.String("operation", operation), zap.Error(validationErr)) // Log unexpected errors at Error level
		return validationErr
	}

	logger.Debug("Struct validation successful", zap.String("operation", operation)) // Log successful validation at Debug level
	return nil
}

// ValidateUUID validates if the given string is a valid UUID format.
func ValidateUUID(id string) error {
	const operation = "security.ValidateUUID"
	logger.Debug("Validating UUID", zap.String("operation", operation), zap.String("uuid_value", id))

	_, err := uuid.Parse(id)
	if err != nil {
		validationErr := utils.NewErrInvalidUUID(id, err) // Use utils.NewErrInvalidUUID (Corrected)
		if errWithLogger, ok := validationErr.(interface{ SetLogger(*zap.Logger) }); ok {
			errWithLogger.SetLogger(logger) // Set logger for enhanced error reporting in domain error
		}
		logger.Warn("Invalid UUID format", zap.String("operation", operation), zap.String("uuid_value", id), zap.Error(validationErr))
		return validationErr
	}

	logger.Debug("UUID validation successful", zap.String("operation", operation), zap.String("uuid_value", id))
	return nil
}

// isValidContentType checks if the given content type string is within the list of allowed content types.
func isValidContentType(contentType string) bool {
	allowedTypes := map[string]bool{
		"application/dicom":       true,
		"application/pdf":         true,
		"image/jpeg":              true,
		"image/png":               true,
		"text/csv; charset=utf-8": true, // Explicitly handle CSV with charset for broader compatibility
	}
	return allowedTypes[contentType]
}

// ValidateFileType checks if the uploaded file has an allowed content type.
func ValidateFileType(file *multipart.FileHeader) error {
	const operation = "security.ValidateFileType"
	logger.Debug("Validating file type", zap.String("operation", operation), zap.String("filename", file.Filename))

	if file == nil {
		fileTypeErr := errors.New("file is nil") // Use standard error for nil file
		logger.Warn("File header is nil", zap.String("operation", operation), zap.Error(fileTypeErr))
		return fileTypeErr
	}

	contentType := file.Header.Get("Content-Type")
	if !isValidContentType(contentType) {
		invalidFileErr := utils.NewErrInvalidFileType(file.Filename, contentType) // Use utils.NewErrInvalidFileType (Corrected)
		if errWithLogger, ok := invalidFileErr.(interface{ SetLogger(*zap.Logger) }); ok {
			errWithLogger.SetLogger(logger) // Set logger for enhanced error reporting in domain error if supported
		}
		logger.Warn("Invalid file type", zap.String("operation", operation), zap.String("filename", file.Filename), zap.String("content_type", contentType), zap.Error(invalidFileErr))
		return invalidFileErr
	}

	logger.Debug("File type validation successful", zap.String("operation", operation), zap.String("filename", file.Filename), zap.String("content_type", contentType))
	return nil
}

// ValidateFileSize checks if the uploaded file size exceeds the configured maximum file size limit.
func ValidateFileSize(file *multipart.FileHeader) error {
	const operation = "security.ValidateFileSize"
	logger.Debug("Validating file size", zap.String("operation", operation), zap.String("filename", file.Filename), zap.Int64("file_size_bytes", file.Size))

	if cfg == nil { // Check if config loaded correctly; this is a safety check.
		configLoadErr := fmt.Errorf("configuration not loaded")
		logger.Error("Configuration error", zap.String("operation", operation), zap.Error(configLoadErr)) // Log as error as config is essential
		return configLoadErr                                                                              // Or a more specific custom error like domain.ErrSystemConfigurationError.
	}

	// Use the MaxFileSize from the loaded configuration to determine the file size limit.
	if file.Size > cfg.MaxFileSize {
		fileSizeErr := utils.NewErrFileSizeExceeded(file.Filename, file.Size, cfg.MaxFileSize) // Use utils.NewErrFileSizeExceeded (Corrected)
		if errWithLogger, ok := fileSizeErr.(interface{ SetLogger(*zap.Logger) }); ok {
			errWithLogger.SetLogger(logger) // Set logger for enhanced error reporting if supported
		}
		logger.Warn("File size exceeded", zap.String("operation", operation), zap.String("filename", file.Filename), zap.Int64("file_size_bytes", file.Size), zap.Int64("max_file_size_limit", cfg.MaxFileSize), zap.Error(fileSizeErr))
		return fileSizeErr
	}

	logger.Debug("File size validation successful", zap.String("operation", operation), zap.String("filename", file.Filename), zap.Int64("file_size_bytes", file.Size))
	return nil
}

// SanitizeFilename removes potentially dangerous characters from a filename to prevent security vulnerabilities like directory traversal.
func SanitizeFilename(filename string) string {
	const operation = "security.SanitizeFilename"
	logger.Debug("Sanitizing filename", zap.String("operation", operation), zap.String("filename", filename))

	re := regexp.MustCompile(`[<>:"/\\|?*\x00-\x1F]`) // Regex to remove control chars and Windows reserved chars
	sanitized := re.ReplaceAllString(filename, "_")   // Replace invalid characters with underscores for safety

	// Prevent directory traversal vulnerabilities by cleaning the filepath.
	sanitized = filepath.Clean(sanitized)
	if strings.Contains(sanitized, "..") {
		// Further sanitize against path traversal by replacing any remaining ".." sequences.
		sanitized = strings.ReplaceAll(sanitized, "..", "__") // Replace ".." with "__" for extra security
	}

	// Limit filename length to prevent database or filesystem issues with excessively long filenames.
	if len(sanitized) > 255 {
		sanitized = sanitized[:255] // Truncate filename to the maximum allowed length
	}

	logger.Debug("Filename sanitized", zap.String("operation", operation), zap.String("original_filename", filename), zap.String("sanitized_filename", sanitized))
	return sanitized
}

// ValidateDICOMUID validates a DICOM Unique Identifier (UID) format.
func ValidateDICOMUID(fl validator.FieldLevel) bool {
	const operation = "security.ValidateDICOMUID"
	logger.Debug("Validating DICOM UID", zap.String("operation", operation), zap.String("dicom_uid_value", fl.Field().String()))

	dicomUIDRegex := regexp.MustCompile(`^[0-2](\.(0|[1-9][0-9]*))*$`) // DICOM UID regex pattern
	isValid := dicomUIDRegex.MatchString(fl.Field().String())

	if !isValid {
		logger.Debug("DICOM UID validation failed", zap.String("operation", operation), zap.String("dicom_uid_value", fl.Field().String()))
	} else {
		logger.Debug("DICOM UID validation successful", zap.String("operation", operation), zap.String("dicom_uid_value", fl.Field().String()))
	}
	return isValid
}

// ValidateBloodPressure validates a blood pressure string format (e.g., "120/80").
func ValidateBloodPressure(fl validator.FieldLevel) bool {
	const operation = "security.ValidateBloodPressure"
	logger.Debug("Validating blood pressure format", zap.String("operation", operation), zap.String("blood_pressure_value", fl.Field().String()))

	bpRegex := regexp.MustCompile(`^\d{2,3}/\d{2,3}$`) // Regex for blood pressure format (e.g., "120/80")
	isValid := bpRegex.MatchString(fl.Field().String())
	if !isValid {
		logger.Debug("Blood pressure validation failed", zap.String("operation", operation), zap.String("blood_pressure_value", fl.Field().String()))
	} else {
		logger.Debug("Blood pressure validation successful", zap.String("operation", operation), zap.String("blood_pressure_value", fl.Field().String()))
	}
	return isValid
}

// HandleValidationError converts validator.ValidationErrors (from go-playground/validator)
// into a more user-friendly ValidationErrors slice, containing custom ValidationError structs.
func HandleValidationError(verr validator.ValidationErrors) utils.ValidationErrors { // Corrected return type to utils.ValidationErrors
	const operation = "security.HandleValidationError"
	logger.Debug("Handling validation errors", zap.String("operation", operation), zap.Int("error_count", len(verr)))

	validationErrors := make(utils.ValidationErrors, len(verr)) // Use utils.ValidationErrors
	for i, fieldError := range verr {
		validationErrors[i] = utils.ValidationError{ // Use utils.ValidationError
			Field:   fieldError.Field(),
			Message: validationErrorMessage(fieldError),
		}
		logger.Debug("Validation error details", zap.String("operation", operation), zap.Int("error_index", i), zap.String("field", fieldError.Field()), zap.String("tag", fieldError.Tag()), zap.String("error_message", validationErrorMessage(fieldError)))
	}

	logger.Debug("Validation error handling complete", zap.String("operation", operation), zap.Int("error_count", len(verr)))
	return validationErrors
}

// validationErrorMessage generates user-friendly error messages based on validator tags.
func validationErrorMessage(fieldErr validator.FieldError) string {
	// Improved and expanded error messages for various validation tags.
	switch fieldErr.Tag() {
	case "required":
		return "This field is required and cannot be empty." // More explicit message
	case "email":
		return "Please enter a valid email address." // User-friendly email message
	case "min":
		return fmt.Sprintf("Must be at least %s characters long.", fieldErr.Param()) // More user-friendly formatting
	case "max":
		return fmt.Sprintf("Cannot exceed %s characters.", fieldErr.Param()) // More user-friendly formatting
	case "gt":
		return fmt.Sprintf("Must be greater than %s.", fieldErr.Param()) // More user-friendly formatting
	case "gte":
		return fmt.Sprintf("Must be greater than or equal to %s.", fieldErr.Param()) // More user-friendly formatting
	case "lt":
		return fmt.Sprintf("Must be less than %s.", fieldErr.Param()) // More user-friendly formatting
	case "lte":
		return fmt.Sprintf("Must be less than or equal to %s.", fieldErr.Param()) // More user-friendly formatting
	case "uuid":
		return "Must be a valid UUID (Universally Unique Identifier)." // More descriptive message for UUIDs
	case "dicomuid": // Added specific message for dicomuid
		return "Must be a valid DICOM UID (Unique Identifier)." // More descriptive message
	case "bloodpressure": // Added for blood pressure
		return "Must be a valid Blood Pressure format (e.g., 120/80)." // More descriptive message
		// Add more cases for other validation tags as needed...
	default:
		return fmt.Sprintf("Value is invalid for field '%s'.", fieldErr.Field()) // More informative generic message.  Include the field name.
	}
}
