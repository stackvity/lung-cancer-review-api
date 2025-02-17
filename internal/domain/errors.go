// internal/domain/errors.go
package domain

import (
	"fmt"

	"go.uber.org/zap"
)

// NotFoundError is returned when a requested resource is not found.
type NotFoundError struct {
	Resource string
	ID       string
	logger   *zap.Logger // Add a logger field, but don't directly inject.
}

func (e *NotFoundError) Error() string {
	// Log the error, but keep it minimal in the domain.
	if e.logger != nil { // Only log if logger is set.
		e.logger.Debug("not found error", zap.String("resource", e.Resource), zap.String("id", e.ID)) //DEBUG LEVEL
	}
	return fmt.Sprintf("%s with ID %s not found", e.Resource, e.ID)
}

// NewNotFoundError creates a new NotFoundError.  It *doesn't* accept a logger.
func NewNotFoundError(resource, id string) *NotFoundError {
	return &NotFoundError{Resource: resource, ID: id, logger: nil}
}

// IsNotFoundError checks if an error is a NotFoundError.
func IsNotFoundError(err error) bool {
	_, ok := err.(*NotFoundError)
	return ok
}

// SetLogger sets the logger for the NotFoundError.  This is called from the *infrastructure* layer.
func (e *NotFoundError) SetLogger(logger *zap.Logger) {
	e.logger = logger
}

// ValidationError represents an error related to invalid input data.
type ValidationError struct {
	Message string
	logger  *zap.Logger
}

// Error implements the error interface.
func (e *ValidationError) Error() string {
	if e.logger != nil { // Only log if logger is set.
		e.logger.Debug("validation error", zap.String("message", e.Message))
	}
	return e.Message
}

// NewValidationError creates a new ValidationError.
func NewValidationError(message string) *ValidationError {
	return &ValidationError{Message: message, logger: nil}
}
func (e *ValidationError) SetLogger(logger *zap.Logger) {
	e.logger = logger
}

// ConflictError represents a conflict error (e.g., trying to create a resource that already exists).
type ConflictError struct {
	Resource string
	ID       string //  The ID of the conflicting resource
	logger   *zap.Logger
}

func (e *ConflictError) Error() string {
	if e.logger != nil {
		e.logger.Debug("conflict error", zap.String("resource", e.Resource), zap.String("id", e.ID))
	}
	return fmt.Sprintf("Conflict error : %s with ID %s already exists", e.Resource, e.ID)
}

// NewConflictError creates a new ConflictError.
func NewConflictError(resource string, id string) *ConflictError {
	return &ConflictError{Resource: resource, ID: id, logger: nil}
}
func (e *ConflictError) SetLogger(logger *zap.Logger) {
	e.logger = logger
}

// UnauthorizedError is used for unauthorized access attempts.
type UnauthorizedError struct {
	Message string
	logger  *zap.Logger
}

func (e *UnauthorizedError) Error() string {
	if e.logger != nil {
		e.logger.Debug("unauthorized error", zap.String("message", e.Message))
	}
	return e.Message
}

func NewUnauthorizedError(message string) *UnauthorizedError {
	return &UnauthorizedError{Message: message, logger: nil} // Initialize logger to nil
}

func (e *UnauthorizedError) SetLogger(logger *zap.Logger) {
	e.logger = logger
}

// ForbiddenError is used when a user/link is not permitted to access a resource.
type ForbiddenError struct {
	Message string
	logger  *zap.Logger
}

func (e *ForbiddenError) Error() string {
	if e.logger != nil {
		e.logger.Debug("forbidden error", zap.String("message", e.Message))
	}
	return e.Message
}

func NewForbiddenError(message string) *ForbiddenError {
	return &ForbiddenError{Message: message, logger: nil} // Initialize logger to nil
}
func (e *ForbiddenError) SetLogger(logger *zap.Logger) {
	e.logger = logger
}

// DataAccessError is a general error for data access issues (database errors, etc.).
type DataAccessError struct {
	Message string
	Err     error // The underlying error.
	logger  *zap.Logger
}

func (e *DataAccessError) Error() string {
	if e.logger != nil {
		e.logger.Debug("data access error", zap.String("message", e.Message), zap.Error(e.Err))
	}
	if e.Err != nil {
		return fmt.Sprintf("data access error: %s - %v", e.Message, e.Err)
	}
	return fmt.Sprintf("data access error: %s", e.Message)
}

func (e *DataAccessError) Unwrap() error { // Make it easy to get the underlying error
	return e.Err
}

func NewDataAccessError(message string, err error) *DataAccessError {
	return &DataAccessError{Message: message, Err: err, logger: nil}
}

func (e *DataAccessError) SetLogger(logger *zap.Logger) {
	e.logger = logger
}

// ErrDICOMParsingFailed represents an error during DICOM file parsing.  // ADDED: DICOM Parsing Error
type ErrDICOMParsingFailed struct {
	Filename string
	Err      error // Underlying error, if any
	logger   *zap.Logger
}

func (e *ErrDICOMParsingFailed) Error() string {
	return fmt.Sprintf("DICOM parsing failed for file '%s': %v", e.Filename, e.Err)
}

func (e *ErrDICOMParsingFailed) Unwrap() error {
	return e.Err
}

// NewErrDICOMParsingFailed creates a new ErrDICOMParsingFailed. // ADDED: Constructor
func NewErrDICOMParsingFailed(filename string, err error) *ErrDICOMParsingFailed {
	return &ErrDICOMParsingFailed{Filename: filename, Err: err, logger: nil}
}

// SetLogger implements the SetLogger method for ErrDICOMParsingFailed.  // ADDED SetLogger
func (e *ErrDICOMParsingFailed) SetLogger(logger *zap.Logger) {
	e.logger = logger
}

// ErrOCRExtractionFailed represents an error during OCR text extraction. // ADDED: OCR Extraction Error
type ErrOCRExtractionFailed struct {
	Filename string
	Err      error
	logger   *zap.Logger
}

func (e *ErrOCRExtractionFailed) Error() string {
	return fmt.Sprintf("OCR text extraction failed for file '%s': %v", e.Filename, e.Err)
}

func (e *ErrOCRExtractionFailed) Unwrap() error {
	return e.Err
}

// NewErrOCRExtractionFailed creates a new ErrOCRExtractionFailed. // ADDED: Constructor
func NewErrOCRExtractionFailed(filename string, err error) *ErrOCRExtractionFailed {
	return &ErrOCRExtractionFailed{Filename: filename, Err: err, logger: nil}
}
func (e *ErrOCRExtractionFailed) SetLogger(logger *zap.Logger) { // ADDED SetLogger
	e.logger = logger
}

// ErrGeminiNoduleDetectionFailed represents an error during Gemini Nodule Detection API call. // ADDED: Gemini Nodule Detection Error
type ErrGeminiNoduleDetectionFailed struct {
	Message string
	Err     error
	logger  *zap.Logger
}

func (e *ErrGeminiNoduleDetectionFailed) Error() string {
	return fmt.Sprintf("Gemini Nodule Detection API call failed: %s - %v", e.Message, e.Err)
}

func (e *ErrGeminiNoduleDetectionFailed) Unwrap() error {
	return e.Err
}

// NewErrGeminiNoduleDetectionFailed creates a new ErrGeminiNoduleDetectionFailed. // ADDED: Constructor
// Corrected constructor to return *ErrGeminiNoduleDetectionFailed
func NewErrGeminiNoduleDetectionFailed(message string, err error) *ErrGeminiNoduleDetectionFailed {
	return &ErrGeminiNoduleDetectionFailed{Message: message, Err: err, logger: nil}
}
func (e *ErrGeminiNoduleDetectionFailed) SetLogger(logger *zap.Logger) { // ADDED SetLogger
	e.logger = logger
}

// ErrInvalidLink represents an error for invalid access links.  // ADDED: Custom Error Type - Recommendation 4
type ErrInvalidLink struct {
	Message string
	logger  *zap.Logger
}

func (e *ErrInvalidLink) Error() string {
	if e.logger != nil {
		e.logger.Debug("invalid access link error", zap.String("message", e.Message)) // Debug level logging for invalid link errors
	}
	return e.Message
}

func NewErrInvalidLink(message string) *ErrInvalidLink { // Constructor for ErrInvalidLink
	return &ErrInvalidLink{Message: message, logger: nil}
}

func (e *ErrInvalidLink) SetLogger(logger *zap.Logger) { // SetLogger method for ErrInvalidLink
	e.logger = logger
}

// ErrLinkExpired represents an error for expired or used access links. // ADDED: Custom Error Type - Recommendation 4
type ErrLinkExpired struct {
	Message string
	logger  *zap.Logger
}

func (e *ErrLinkExpired) Error() string {
	if e.logger != nil {
		e.logger.Debug("expired access link error", zap.String("message", e.Message)) // Debug level logging for expired link errors
	}
	return e.Message
}

func NewErrLinkExpired(message string) *ErrLinkExpired { // Constructor for ErrLinkExpired
	return &ErrLinkExpired{Message: message, logger: nil}
}

func (e *ErrLinkExpired) SetLogger(logger *zap.Logger) { // SetLogger method for ErrLinkExpired
	e.logger = logger
}

// ErrGeminiDiagnosisFailed represents an error during Gemini Preliminary Diagnosis API call. // ADDED: Gemini Preliminary Diagnosis Error - Recommendation 1
type ErrGeminiDiagnosisFailed struct {
	Message string
	Err     error
	logger  *zap.Logger
}

func (e *ErrGeminiDiagnosisFailed) Error() string {
	return fmt.Sprintf("Gemini Preliminary Diagnosis API call failed: %s - %v", e.Message, e.Err)
}

func (e *ErrGeminiDiagnosisFailed) Unwrap() error {
	return e.Err
}

// NewErrGeminiDiagnosisFailed creates a new ErrGeminiDiagnosisFailed. // ADDED: Constructor - Recommendation 1
func NewErrGeminiDiagnosisFailed(message string, err error) *ErrGeminiDiagnosisFailed {
	return &ErrGeminiDiagnosisFailed{Message: message, Err: err, logger: nil}
}

// SetLogger implements the SetLogger method for ErrGeminiDiagnosisFailed. // ADDED SetLogger - Recommendation 1
func (e *ErrGeminiDiagnosisFailed) SetLogger(logger *zap.Logger) {
	e.logger = logger
}

// ErrDBDiagnosisSaveFailed represents an error during saving diagnosis data to the database. // ADDED: DB Diagnosis Save Error - Recommendation 1
type ErrDBDiagnosisSaveFailed struct {
	Message string
	Err     error
	logger  *zap.Logger
}

// Error implements the error interface for ErrDBDiagnosisSaveFailed.
func (e *ErrDBDiagnosisSaveFailed) Error() string {
	return fmt.Sprintf("Database error saving diagnosis: %s - %v", e.Message, e.Err)
}

// Unwrap returns the underlying error for ErrDBDiagnosisSaveFailed.
func (e *ErrDBDiagnosisSaveFailed) Unwrap() error {
	return e.Err
}

// NewErrDBDiagnosisSaveFailed creates a new ErrDBDiagnosisSaveFailed. // ADDED: Constructor - Recommendation 1
func NewErrDBDiagnosisSaveFailed(message string, err error) *ErrDBDiagnosisSaveFailed {
	return &ErrDBDiagnosisSaveFailed{Message: message, Err: err, logger: nil}
}

// SetLogger implements the SetLogger method for ErrDBDiagnosisSaveFailed.
func (e *ErrDBDiagnosisSaveFailed) SetLogger(logger *zap.Logger) {
	e.logger = logger
}

// ErrGeminiStagingFailed represents an error during Gemini Staging API call. // ADDED: Gemini Staging Error - Recommendation 1
type ErrGeminiStagingFailed struct {
	Message string
	Err     error
	logger  *zap.Logger
}

func (e *ErrGeminiStagingFailed) Error() string {
	return fmt.Sprintf("Gemini Staging API call failed: %s - %v", e.Message, e.Err)
}

func (e *ErrGeminiStagingFailed) Unwrap() error {
	return e.Err
}

// NewErrGeminiStagingFailed creates a new ErrGeminiStagingFailed. // ADDED: Constructor - Recommendation 1
func NewErrGeminiStagingFailed(message string, err error) *ErrGeminiStagingFailed {
	return &ErrGeminiStagingFailed{Message: message, Err: err, logger: nil}
}

// SetLogger implements the SetLogger method for ErrGeminiStagingFailed.
func (e *ErrGeminiStagingFailed) SetLogger(logger *zap.Logger) {
	e.logger = logger
}

// ErrDBStagingSaveFailed represents an error during saving staging data to the database. // ADDED: DB Staging Save Error - Recommendation 1
type ErrDBStagingSaveFailed struct {
	Message string
	Err     error
	logger  *zap.Logger
}

// Error implements the error interface for ErrDBStagingSaveFailed.
func (e *ErrDBStagingSaveFailed) Error() string {
	return fmt.Sprintf("Database error saving staging information: %s - %v", e.Message, e.Err)
}
func (e *ErrDBStagingSaveFailed) Unwrap() error {
	return e.Err
}

// NewErrDBStagingSaveFailed creates a new ErrDBStagingSaveFailed. // ADDED: Constructor - Recommendation 1
func NewErrDBStagingSaveFailed(message string, err error) *ErrDBStagingSaveFailed {
	return &ErrDBStagingSaveFailed{Message: message, Err: err, logger: nil}
}

// SetLogger implements the SetLogger method for ErrDBStagingSaveFailed.
func (e *ErrDBStagingSaveFailed) SetLogger(logger *zap.Logger) {
	e.logger = logger
}

// ErrGeminiTreatmentSuggestFailed represents an error during Gemini Treatment Suggestion API call. // ADDED: Gemini Treatment Suggestion Error - Recommendation 1
type ErrGeminiTreatmentSuggestFailed struct {
	Message string
	Err     error
	logger  *zap.Logger
}

func (e *ErrGeminiTreatmentSuggestFailed) Error() string {
	return fmt.Sprintf("Gemini Treatment Suggestion API call failed: %s - %v", e.Message, e.Err)
}

func (e *ErrGeminiTreatmentSuggestFailed) Unwrap() error {
	return e.Err
}

// NewErrGeminiTreatmentSuggestFailed creates a new ErrGeminiTreatmentSuggestFailed.  // ADDED: Constructor - Recommendation 1
func NewErrGeminiTreatmentSuggestFailed(message string, err error) *ErrGeminiTreatmentSuggestFailed {
	return &ErrGeminiTreatmentSuggestFailed{Message: message, Err: err, logger: nil}
}
func (e *ErrGeminiTreatmentSuggestFailed) SetLogger(logger *zap.Logger) {
	e.logger = logger
}

// ErrDBTreatmentSaveFailed represents an error during saving treatment data to the database. // ADDED: DB Treatment Save Error - Recommendation 1
type ErrDBTreatmentSaveFailed struct {
	Message string
	Err     error
	logger  *zap.Logger
}

// Error implements the error interface for ErrDBTreatmentSaveFailed.
func (e *ErrDBTreatmentSaveFailed) Error() string {
	return fmt.Sprintf("Database error saving treatment recommendation: %s - %v", e.Message, e.Err)
}
func (e *ErrDBTreatmentSaveFailed) Unwrap() error {
	return e.Err
}

// NewErrDBTreatmentSaveFailed creates a new ErrDBTreatmentSaveFailed. // ADDED: Constructor - Recommendation 1
func NewErrDBTreatmentSaveFailed(message string, err error) *ErrDBTreatmentSaveFailed {
	return &ErrDBTreatmentSaveFailed{Message: message, Err: err, logger: nil}
}

// SetLogger implements the SetLogger method for ErrDBTreatmentSaveFailed.
func (e *ErrDBTreatmentSaveFailed) SetLogger(logger *zap.Logger) {
	e.logger = logger
}

// IsNotFoundError checks if an error is a NotFoundError.
// ... (IsNotFoundError function - no changes needed) ...

// Implement the `Is` method for the new custom error types to support error wrapping and checking.
// Is function for Custom Errors
func (e *ErrDICOMParsingFailed) Is(target error) bool { // ADDED: Is for ErrDICOMParsingFailed
	_, ok := target.(*ErrDICOMParsingFailed)
	return ok
}

// Is function for Custom Errors
func (e *ErrOCRExtractionFailed) Is(target error) bool { // ADDED: Is for ErrOCRExtractionFailed
	_, ok := target.(*ErrOCRExtractionFailed)
	return ok
}

// Is function for Custom Errors
func (e *ErrGeminiNoduleDetectionFailed) Is(target error) bool { // ADDED: Is for ErrGeminiNoduleDetectionFailed
	_, ok := target.(*ErrGeminiNoduleDetectionFailed)
	return ok
}

// Is function for Data Access Errors
func (e *DataAccessError) Is(target error) bool {
	_, ok := target.(*DataAccessError)
	return ok
}

// Is function for Validation Errors
func (e *ValidationError) Is(target error) bool {
	_, ok := target.(*ValidationError)
	return ok
}

// Is function for Conflict Errors
func (e *ConflictError) Is(target error) bool {
	_, ok := target.(*ConflictError)
	return ok
}

// Is function for Unauthorized Errors
func (e *UnauthorizedError) Is(target error) bool {
	_, ok := target.(*UnauthorizedError)
	return ok
}

// Is function for Forbidden Errors
func (e *ForbiddenError) Is(target error) bool {
	_, ok := target.(*ForbiddenError)
	return ok
}

// Is function for Custom Errors - Recommendation 4 - ADDED Is for ErrInvalidLink
func (e *ErrInvalidLink) Is(target error) bool {
	_, ok := target.(*ErrInvalidLink)
	return ok
}

// Is function for Custom Errors - Recommendation 4 - ADDED Is for ErrLinkExpired
func (e *ErrLinkExpired) Is(target error) bool {
	_, ok := target.(*ErrLinkExpired)
	return ok
}

// Is function for Custom Errors - Recommendation 1 - ADDED Is for ErrGeminiDiagnosisFailed
func (e *ErrGeminiDiagnosisFailed) Is(target error) bool {
	_, ok := target.(*ErrGeminiDiagnosisFailed)
	return ok
}

// Is function for Custom Errors - Recommendation 1 - ADDED Is for ErrDBDiagnosisSaveFailed
func (e *ErrDBDiagnosisSaveFailed) Is(target error) bool {
	_, ok := target.(*ErrDBDiagnosisSaveFailed)
	return ok
}

// Is function for Custom Errors - Recommendation 1 - ADDED Is for ErrGeminiStagingFailed
func (e *ErrGeminiStagingFailed) Is(target error) bool {
	_, ok := target.(*ErrGeminiStagingFailed)
	return ok
}

// Is function for Custom Errors - Recommendation 1 - ADDED Is for ErrDBStagingSaveFailed
func (e *ErrDBStagingSaveFailed) Is(target error) bool {
	_, ok := target.(*ErrDBStagingSaveFailed)
	return ok
}

// Is function for Custom Errors - Recommendation 1 - ADDED Is for ErrGeminiTreatmentSuggestFailed
func (e *ErrGeminiTreatmentSuggestFailed) Is(target error) bool {
	_, ok := target.(*ErrGeminiTreatmentSuggestFailed)
	return ok
}

// Is function for Custom Errors - Recommendation 1 - ADDED Is for ErrDBTreatmentSaveFailed
func (e *ErrDBTreatmentSaveFailed) Is(target error) bool {
	_, ok := target.(*ErrDBTreatmentSaveFailed)
	return ok
}

// --- Metrics Integration (Comment - Recommendation 3 - Metrics Integration): ---
// In a production environment, consider adding metrics instrumentation to these error types.
// Example using Prometheus (conceptual):
//
// var (
//    geminiAPIErrorsTotal = prometheus.NewCounterVec(
//        prometheus.CounterOpts{
//            Name: "gemini_api_errors_total",
//            Help: "Total number of Gemini API errors",
//        },
//        []string{"error_code", "operation", "severity"}, // ADDED: Severity as label - Recommendation: Error Classification and Severity Levels
//    )
// )
//
// func init() {
//    prometheus.MustRegister(geminiAPIErrorsTotal)
// }
//
// Then, in each NewErr... constructor:
//
//  NewErrGeminiAPIError(message string, err error) *ErrGeminiAPIError {
// 	   geminiAPIErrorsTotal.With(prometheus.Labels{
//         "error_code": ErrCodeGeminiAPIRequestFailed,
//         "operation": "GeneralAPI",
//         "severity": SeverityError, // ADDED: Severity Label - Recommendation: Error Classification and Severity Levels
//     }).Inc() // Increment counter
//     return &ErrGeminiAPIError{ ... }
//  }
//
// This would allow you to track the frequency of different Gemini API errors over time in Prometheus/Grafana,
// broken down by error code, operation, and now also by severity level, providing even richer insights
// into error patterns, enabling prioritization of error resolution based on severity, and facilitating more granular alerting strategies.
