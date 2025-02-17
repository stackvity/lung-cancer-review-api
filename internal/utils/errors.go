// internal/utils/errors.go
package utils

import (
	"errors"
	"fmt"
	"net/http" // Import net/http for HTTP status codes

	"github.com/go-playground/validator/v10" // Import validator
	"github.com/google/uuid"                 // Import uuid
	"go.uber.org/zap"
)

// ValidationError represents a structured validation error.
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// Error implements the error interface for *ValidationError.
func (ve *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", ve.Field, ve.Message)
}

// ValidationErrors represents a collection of validation errors.
type ValidationErrors []ValidationError

// Error implements the error interface for ValidationErrors.
func (ve ValidationErrors) Error() string {
	// This provides a basic string representation of all validation errors.  A more
	// sophisticated implementation might format it more nicely (e.g., as JSON).
	errorString := "Validation errors: "
	for _, err := range ve {
		errorString += fmt.Sprintf("[%s: %s] ", err.Field, err.Message)
	}
	return errorString
}

// HandleValidationError converts validator.ValidationErrors into our custom
// ValidationErrors type, making them more suitable for API responses.
func HandleValidationError(verr validator.ValidationErrors) ValidationErrors {
	validationErrors := make(ValidationErrors, len(verr))
	for i, fieldError := range verr {
		validationErrors[i] = ValidationError{
			Field:   fieldError.Field(),
			Message: validationErrorMessage(fieldError),
		}
	}
	return validationErrors
}

// validationErrorMessage generates user-friendly error messages based on
// validator tags.  This is a helper function to centralize error message logic.
func validationErrorMessage(fieldErr validator.FieldError) string { // Corrected variable name
	switch fieldErr.Tag() {
	case "required":
		return "This field is required and cannot be empty." // More explicit message
	case "email":
		return "Please enter a valid email address." // User-friendly email message
	case "min":
		return fmt.Sprintf("Must be at least %s characters long.", fieldErr.Param()) // More user-friendly formatting
	case "max":
		return fmt.Sprintf("Cannot exceed %s characters.", fieldErr.Param()) // More user-friendly formatting
	case "gt": // Greater than (for numbers)
		return fmt.Sprintf("Must be greater than %s.", fieldErr.Param()) // More user-friendly formatting
	case "gte": // Greater than or equal to (for numbers)
		return fmt.Sprintf("Must be greater than or equal to %s.", fieldErr.Param()) // More user-friendly formatting
	case "lt": // Less than (for numbers)
		return fmt.Sprintf("Must be less than %s.", fieldErr.Param()) // More user-friendly formatting
	case "lte": // Less than or equal to (for numbers)
		return fmt.Sprintf("Must be less than or equal to %s.", fieldErr.Param()) // More user-friendly formatting
	case "uuid": // Added for UUID validation
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

// Wrap adds context to an existing error. It's used for error propagation.
func Wrap(err error, message string) error {
	return fmt.Errorf("%s: %w", message, err)
}

// Wrapf adds context to an existing error, using printf-style formatting.
func Wrapf(err error, format string, args ...interface{}) error {
	return fmt.Errorf("%s: %w", fmt.Sprintf(format, args...), err)
}

// As allows checking for a specific error type in an error chain (using errors.As).
func As(err error, target interface{}) bool {
	return errors.As(err, target)
}

// Is allows checking if an error matches a specific error value (using errors.Is).
func Is(err, target error) bool {
	return errors.Is(err, target)
}

// NewError creates a new error with a formatted message.
func NewError(format string, args ...interface{}) error {
	return fmt.Errorf(format, args...)
}

// HTTPError represents an error with an associated HTTP status code.
type HTTPError struct {
	Code    int    `json:"code"`    // HTTP status code
	Message string `json:"message"` // Error message
	Err     error  `json:"-"`       // Underlying error (not serialized to JSON)
}

// Error implements the error interface.
func (e *HTTPError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("HTTP Error %d: %s - %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("HTTP Error %d: %s", e.Code, e.Message)
}

// Unwrap returns the underlying error.  This is important for errors.Is and errors.As.
func (e *HTTPError) Unwrap() error {
	return e.Err
}

// NewHTTPError creates a new HTTPError.  The 'err' argument is optional.
func NewHTTPError(code int, message string, err ...error) *HTTPError {
	httpError := &HTTPError{
		Code:    code,
		Message: message,
	}
	if len(err) > 0 {
		httpError.Err = err[0]
	}
	return httpError
}

// Error responses for common HTTP errors (used throughout the application).
var (
	ErrBadRequest            = NewHTTPError(http.StatusBadRequest, "Bad Request")
	ErrUnauthorized          = NewHTTPError(http.StatusUnauthorized, "Unauthorized")
	ErrForbidden             = NewHTTPError(http.StatusForbidden, "Forbidden")
	ErrNotFound              = NewHTTPError(http.StatusNotFound, "Not Found")
	ErrMethodNotAllowed      = NewHTTPError(http.StatusMethodNotAllowed, "Method Not Allowed")
	ErrConflict              = NewHTTPError(http.StatusConflict, "Conflict")
	ErrInternalServerError   = NewHTTPError(http.StatusInternalServerError, "Internal Server Error")
	ErrTooManyRequests       = NewHTTPError(http.StatusTooManyRequests, "Too Many Requests")
	ErrUnsupportedMediaType  = NewHTTPError(http.StatusUnsupportedMediaType, "Unsupported Media Type")
	ErrRequestEntityTooLarge = NewHTTPError(http.StatusRequestEntityTooLarge, "Request Entity Too Large")
	ErrServiceUnavailable    = NewHTTPError(http.StatusServiceUnavailable, "Service Unavailable")
)

// ValidateUUID checks if a given string is a valid UUID.
func ValidateUUID(id string) error {
	_, err := uuid.Parse(id)
	if err != nil {
		return NewErrInvalidUUID(id, err) // Use custom error
	}
	return nil
}

// ErrDBConnection represents an error during database connection.
type ErrDBConnection struct {
	Message string
	Err     error // Underlying error, if any
	logger  *zap.Logger
}

func (e *ErrDBConnection) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("database connection error: %s - %v", e.Message, e.Err)
	}
	return fmt.Sprintf("database connection error: %s", e.Message)
}

func (e *ErrDBConnection) Unwrap() error {
	return e.Err
}

func NewErrDBConnection(message string, err error) error {
	return &ErrDBConnection{Message: message, Err: err, logger: nil}
}

// SetLogger implements the SetLogger method for ErrDBConnection.
func (e *ErrDBConnection) SetLogger(logger *zap.Logger) {
	e.logger = logger
}

// ErrDBQuery represents an error during database query execution.
type ErrDBQuery struct {
	Message    string
	QueryName  string      // Name or description of the query that failed
	Query      string      // The actual SQL query that failed (sensitive - use with caution)
	Parameters interface{} // Parameters passed to the query (sensitive - use with caution)
	Err        error       // Underlying error, if any
	logger     *zap.Logger // ADDED: Logger for ErrDBQuery
}

func (e *ErrDBQuery) Error() string {
	if e.logger != nil {
		e.logger.Debug("database query error", zap.String("queryName", e.QueryName), zap.String("message", e.Message), zap.Error(e.Err))
	}
	if e.Err != nil {
		return fmt.Sprintf("database query error in '%s': %s - Query: '%s', Params: %+v, Underlying error: %v", e.QueryName, e.Message, e.Query, e.Parameters, e.Err)
	}
	return fmt.Sprintf("database query error in '%s': %s - Query: '%s', Params: %+v", e.QueryName, e.Message, e.Query, e.Parameters)
}
func (e *ErrDBQuery) Unwrap() error {
	return e.Err
}

func NewErrDBQuery(message string, queryName string, query string, params interface{}, err error) error {
	return &ErrDBQuery{Message: message, QueryName: queryName, Query: query, Parameters: params, Err: err, logger: nil} // Initialize logger to nil
}

// SetLogger implements the SetLogger method for ErrDBQuery.
func (e *ErrDBQuery) SetLogger(logger *zap.Logger) { // ADDED SetLogger method to ErrDBQuery
	e.logger = logger
}

// ErrDataIntegrity represents a data integrity violation error (e.g., unique constraint violation).
type ErrDataIntegrity struct {
	Message    string
	Constraint string // Optional: Constraint that was violated
	TableName  string // Optional: Table where the violation occurred
	Err        error  // Underlying error, if any
	logger     *zap.Logger
}

func (e *ErrDataIntegrity) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("data integrity error on table '%s' (constraint '%s'): %s - %v", e.TableName, e.Constraint, e.Message, e.Err)
	}
	return fmt.Sprintf("data integrity error on table '%s' (constraint '%s'): %s", e.TableName, e.Constraint, e.Message)
}

func (e *ErrDataIntegrity) Unwrap() error {
	return e.Err
}
func NewErrDataIntegrity(message string, constraint string, tableName string, err error) error {
	return &ErrDataIntegrity{Message: message, Constraint: constraint, TableName: tableName, Err: err, logger: nil}
}

// SetLogger implements the SetLogger method for DataIntegrityError.
func (e *ErrDataIntegrity) SetLogger(logger *zap.Logger) {
	e.logger = logger
}

// ErrInvalidUUID represents an error for invalid UUID format.
type ErrInvalidUUID struct {
	Value  string
	Err    error
	logger *zap.Logger
}

func (e *ErrInvalidUUID) Error() string {
	return fmt.Sprintf("invalid UUID format: '%s' - %v", e.Value, e.Err)
}
func (e *ErrInvalidUUID) Unwrap() error {
	return e.Err
}

// NewErrInvalidUUID creates a new ErrInvalidUUID.
func NewErrInvalidUUID(value string, err error) error {
	return &ErrInvalidUUID{Value: value, Err: err, logger: nil}
}

// SetLogger implements the SetLogger method for ErrInvalidUUID.
func (e *ErrInvalidUUID) SetLogger(logger *zap.Logger) {
	e.logger = logger
}

// ErrInvalidFileType represents an error for invalid file type.
type ErrInvalidFileType struct {
	Filename string
	Type     string
	logger   *zap.Logger
}

func (e *ErrInvalidFileType) Error() string {
	return fmt.Sprintf("invalid file type for '%s': '%s' is not allowed", e.Filename, e.Type)
}
func (e *ErrInvalidFileType) Unwrap() error {
	return nil
}
func NewErrInvalidFileType(filename string, fileType string) error {
	return &ErrInvalidFileType{Filename: filename, Type: fileType, logger: nil}
}
func (e *ErrInvalidFileType) SetLogger(logger *zap.Logger) {
	e.logger = logger
}

// ErrFileSizeExceeded represents an error for exceeding maximum file size.
type ErrFileSizeExceeded struct {
	Filename string
	Size     int64
	Limit    int64
	logger   *zap.Logger
}

func (e *ErrFileSizeExceeded) Error() string {
	return fmt.Sprintf("file '%s' exceeds size limit: %d bytes, limit is %d bytes", e.Filename, e.Size, e.Limit)
}
func (e *ErrFileSizeExceeded) Unwrap() error {
	return nil
}
func NewErrFileSizeExceeded(filename string, size int64, limit int64) error {
	return &ErrFileSizeExceeded{Filename: filename, Size: size, Limit: limit, logger: nil}
}
func (e *ErrFileSizeExceeded) SetLogger(logger *zap.Logger) {
	e.logger = logger
}

// Is function for Data Access Errors
func (e *ErrDBConnection) Is(target error) bool {
	_, ok := target.(*ErrDBConnection)
	return ok
}

// Is function for Data Access Errors
func (e *ErrDBQuery) Is(target error) bool {
	_, ok := target.(*ErrDBQuery)
	return ok
}

// Is function for Data Access Errors
func (e *ErrDataIntegrity) Is(target error) bool {
	_, ok := target.(*ErrDataIntegrity)
	return ok
}

// Is function for Validation Errors
func (e *ValidationError) Is(target error) bool {
	_, ok := target.(*ValidationError)
	return ok
}

// Is function for HTTP Errors
func (e *HTTPError) Is(target error) bool {
	_, ok := target.(*HTTPError)
	return ok
}

// Is function for Custom Errors
func (e *ErrInvalidUUID) Is(target error) bool {
	_, ok := target.(*ErrInvalidUUID)
	return ok
}

// Is function for Custom Errors
func (e *ErrInvalidFileType) Is(target error) bool {
	_, ok := target.(*ErrInvalidFileType)
	return ok
}

// Is function for Custom Errors
func (e *ErrFileSizeExceeded) Is(target error) bool {
	_, ok := target.(*ErrFileSizeExceeded)
	return ok
}
