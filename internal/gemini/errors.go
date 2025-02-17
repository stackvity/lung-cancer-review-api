// internal/gemini/errors.go
package gemini

import (
	"fmt"

	"go.uber.org/zap"
)

// Define error codes and severity levels as constants for programmatic error checking and categorization - Recommendation: Error Codes or Enums & Error Classification
const (
	ErrCodeGeminiAPIRequestFailed      = "GEMINI_API_REQUEST_FAILED"          // Grouped with "GEMINI_API" prefix - Recommendation: Error Grouping/Categorization
	ErrCodeGeminiResponseParsingFailed = "GEMINI_API_RESPONSE_PARSING_FAILED" // Grouped with "GEMINI_API" prefix

	ErrCodeGeminiNoduleDetectionFailed       = "GEMINI_NODULE_DETECTION_FAILED"   // Grouped with "GEMINI_NODULE_DETECTION" prefix - Recommendation: More Granular Error Types & Grouping
	ErrCodeGeminiPathologyAnalysisFailed     = "GEMINI_PATHOLOGY_ANALYSIS_FAILED" // Grouped with "GEMINI_PATHOLOGY_ANALYSIS" prefix - Recommendation: More Granular Error Types & Grouping
	ErrCodeGeminiInformationExtractionFailed = "GEMINI_INFO_EXTRACTION_FAILED"    // Grouped with "GEMINI_INFO_EXTRACTION" prefix - Recommendation: More Granular Error Types & Grouping
	ErrCodeGeminiDiagnosisFailed             = "GEMINI_DIAGNOSIS_FAILED"          // Grouped with "GEMINI_DIAGNOSIS" prefix - Recommendation: More Granular Error Types & Grouping
	ErrCodeGeminiStagingFailed               = "GEMINI_STAGING_FAILED"            // Grouped with "GEMINI_STAGING" prefix - Recommendation: More Granular Error Types & Grouping
	ErrCodeGeminiTreatmentSuggestFailed      = "GEMINI_TREATMENT_SUGGEST_FAILED"  // Grouped with "GEMINI_TREATMENT_SUGGEST" prefix - Recommendation: More Granular Error Types & Grouping

	SeverityCritical = "Critical"      // ADDED: Severity Levels - Recommendation: Error Classification and Severity Levels
	SeverityError    = "Error"         // ADDED: Severity Levels - Recommendation: Error Classification and Severity Levels
	SeverityWarning  = "Warning"       // ADDED: Severity Levels - Recommendation: Error Classification and Severity Levels
	SeverityInfo     = "Informational" // ADDED: Severity Levels - Recommendation: Error Classification and Severity Levels
)

// ErrGeminiAPIError is a custom error type for general Gemini API related errors.
// It encapsulates a message, an error code (ErrCodeGeminiAPIRequestFailed by default), a severity level, and an underlying error for detailed error reporting.
//
// Error Code: ErrCodeGeminiAPIRequestFailed
// Severity Level: Error - Recommendation: Error Classification and Severity Levels
// Possible Causes: Network connectivity issues, Gemini API service unavailable, incorrect API endpoint, general unclassified API errors.
// Actionable Recommendations: Check network connectivity, verify API endpoint configuration, review underlying error for details, implement retry mechanisms.
type ErrGeminiAPIError struct {
	Code     string // ADDED: Error Code - Recommendation: Error Codes or Enums
	Severity string // ADDED: Severity Level - Recommendation: Error Classification and Severity Levels
	Message  string
	Err      error // Optional underlying error
	logger   *zap.Logger
}

// Error method to implement the error interface. Includes severity level and error code in the error message - Recommendation: Error Classification and Severity Levels
func (e *ErrGeminiAPIError) Error() string {
	return fmt.Sprintf("[%s] Gemini API error [%s]: %s - %v", e.Severity, e.Code, e.Message, e.Err) // ADDED: Include Severity Level and Error Code in Error Message
}

// Unwrap method to allow error unwrapping.
func (e *ErrGeminiAPIError) Unwrap() error {
	return e.Err
}

// NewErrGeminiAPIError creates a new ErrGeminiAPIError. Defaults to Error severity. - Recommendation: Error Classification and Severity Levels
func NewErrGeminiAPIError(message string, err error) *ErrGeminiAPIError {
	return &ErrGeminiAPIError{
		Code:     ErrCodeGeminiAPIRequestFailed, // Default Error Code
		Message:  message,
		Err:      err,
		Severity: SeverityError, // Default Severity - Recommendation: Error Classification and Severity Levels
		logger:   nil,
	}
}

// SetLogger sets the logger for the ErrGeminiAPIError.
func (e *ErrGeminiAPIError) SetLogger(logger *zap.Logger) {
	e.logger = logger
}

// ErrGeminiResponseParsingFailed is a custom error type for Gemini API response parsing failures.
// It includes a message, an error code (ErrCodeGeminiResponseParsingFailed), a severity level and the raw response string for debugging purposes.
//
// Error Code: ErrCodeGeminiResponseParsingFailed
// Severity Level: Warning - Recommendation: Error Classification and Severity Levels
// Possible Causes: Changes in Gemini API response format, incorrect parsing logic in the application, network issues leading to incomplete responses.
// Actionable Recommendations: Review Gemini API documentation for response format changes, update parsing logic, log raw API response for debugging, ensure network stability.
type ErrGeminiResponseParsingFailed struct {
	Code        string // ADDED: Error Code - Recommendation: Error Codes or Enums
	Severity    string // ADDED: Severity Level - Recommendation: Error Classification and Severity Levels
	Message     string
	RawResponse string // Raw response for debugging
	Err         error  // Underlying error, if any
	logger      *zap.Logger
}

// Error method for ErrGeminiResponseParsingFailed. Includes severity and error code in message.  - Recommendation: Error Classification and Severity Levels
func (e *ErrGeminiResponseParsingFailed) Error() string {
	return fmt.Sprintf("[%s] Gemini API response parsing failed [%s]: %s - Raw Response: '%s' - %v", e.Severity, e.Code, e.Message, e.RawResponse, e.Err) // ADDED: Include Severity Level and Error Code
}

// Unwrap method for ErrGeminiResponseParsingFailed.
func (e *ErrGeminiResponseParsingFailed) Unwrap() error {
	return e.Err
}

// NewErrGeminiResponseParsingFailed creates a new ErrGeminiResponseParsingFailed. Defaults to Warning severity. - Recommendation: Error Classification and Severity Levels
func NewErrGeminiResponseParsingFailed(message string, rawResponse string, err error) *ErrGeminiResponseParsingFailed {
	return &ErrGeminiResponseParsingFailed{
		Code:        ErrCodeGeminiResponseParsingFailed, // ADDED: Error Code - Recommendation: Error Codes or Enums
		Message:     message,
		RawResponse: rawResponse,
		Err:         err,
		Severity:    SeverityWarning, // Default Severity - Recommendation: Error Classification and Severity Levels
		logger:      nil,
	}
}

// SetLogger sets the logger for the ErrGeminiResponseParsingFailed.
func (e *ErrGeminiResponseParsingFailed) SetLogger(logger *zap.Logger) {
	e.logger = logger
}

// ErrGeminiRequestFailed represents an error for Gemini API request failures.
// It includes details about the request, an error code (ErrCodeGeminiAPIRequestFailed), severity and the underlying error.
//
// Error Code: ErrCodeGeminiAPIRequestFailed
// Severity Level: Error - Recommendation: Error Classification and Severity Levels
// Possible Causes: Network errors, timeouts, invalid request format, authentication failures, rate limiting.
// Actionable Recommendations: Check network connectivity, verify API key and authentication, review request parameters, implement retry mechanisms with exponential backoff, monitor API request rates.
type ErrGeminiRequestFailed struct {
	Code        string // ADDED: Error Code - Recommendation: Error Codes or Enums
	Severity    string // ADDED: Severity Level - Recommendation: Error Classification and Severity Levels
	Message     string
	RequestURL  string // URL of the failed request
	RequestBody string // Request body (consider scrubbing sensitive data)
	Err         error  // Underlying error
	logger      *zap.Logger
}

// Error method for ErrGeminiRequestFailed. Includes severity and error code in message. - Recommendation: Error Classification and Severity Levels
func (e *ErrGeminiRequestFailed) Error() string {
	return fmt.Sprintf("[%s] Gemini API request failed [%s] to URL: %s - Message: %s - Request Body: '%s' - %v", e.Severity, e.Code, e.RequestURL, e.Message, e.RequestBody, e.Err) // ADDED: Include Severity Level and Error Code
}

// Unwrap method for ErrGeminiRequestFailed.
func (e *ErrGeminiRequestFailed) Unwrap() error {
	return e.Err
}

// NewErrGeminiRequestFailed creates a new ErrGeminiRequestFailed. Defaults to Error severity. - Recommendation: Error Classification and Severity Levels
func NewErrGeminiRequestFailed(message string, requestURL string, requestBody string, err error) *ErrGeminiRequestFailed {
	return &ErrGeminiRequestFailed{
		Code:        ErrCodeGeminiAPIRequestFailed, // Default Error Code
		Severity:    SeverityError,                 // Default Severity - Recommendation: Error Classification and Severity Levels
		Message:     message,
		RequestURL:  requestURL,
		RequestBody: requestBody,
		Err:         err,
		logger:      nil,
	}
}

// SetLogger sets the logger for the ErrGeminiRequestFailed.
func (e *ErrGeminiRequestFailed) SetLogger(logger *zap.Logger) {
	e.logger = logger
}

// ErrGeminiNoduleDetectionFailed represents an error during Gemini Nodule Detection API call. // ADDED: Gemini Nodule Detection Error - Recommendation 1
//
// Error Code: ErrCodeGeminiNoduleDetectionFailed - Recommendation: More Granular Error Types & Grouping
// Severity Level: Warning - Recommendation: Error Classification and Severity Levels
// Possible Causes: Issues specific to the Nodule Detection API endpoint, such as incorrect image data format, invalid prompts for nodule detection, model-specific errors.
// Actionable Recommendations: Verify image data format and compatibility with Gemini API, review and refine nodule detection prompts, check Gemini API documentation for nodule detection specific errors.
type ErrGeminiNoduleDetectionFailed struct {
	Code     string // ADDED: Error Code - Recommendation: Error Codes or Enums
	Severity string // ADDED: Severity Level - Recommendation: Error Classification and Severity Levels
	Message  string
	Err      error
	logger   *zap.Logger
}

func (e *ErrGeminiNoduleDetectionFailed) Error() string {
	return fmt.Sprintf("[%s] Gemini Nodule Detection API call failed [%s]: %s - %v", e.Severity, e.Code, e.Message, e.Err) // ADDED: Include Severity Level and Error Code
}

func (e *ErrGeminiNoduleDetectionFailed) Unwrap() error {
	return e.Err
}

// NewErrGeminiNoduleDetectionFailed creates a new ErrGeminiNoduleDetectionFailed. Defaults to Warning severity. - Recommendation: Error Classification and Severity Levels
func NewErrGeminiNoduleDetectionFailed(message string, err error) *ErrGeminiNoduleDetectionFailed {
	return &ErrGeminiNoduleDetectionFailed{
		Code:     ErrCodeGeminiNoduleDetectionFailed, // ADDED: Error Code - Recommendation: Error Codes or Enums
		Severity: SeverityWarning,                    // Default Severity - Recommendation: Error Classification and Severity Levels
		Message:  message,
		Err:      err,
		logger:   nil,
	}
}

// SetLogger implements the SetLogger method for ErrGeminiNoduleDetectionFailed. // ADDED SetLogger
func (e *ErrGeminiNoduleDetectionFailed) SetLogger(logger *zap.Logger) {
	e.logger = logger
}

// ErrGeminiPathologyAnalysisFailed represents an error during Gemini Pathology Report Analysis API call. // ADDED: Pathology Report Analysis Error - Recommendation 1
//
// Error Code: ErrCodeGeminiPathologyAnalysisFailed - Recommendation: More Granular Error Types & Grouping
// Severity Level: Warning  - Recommendation: Error Classification and Severity Levels
// Possible Causes: Issues specific to the Pathology Report Analysis API endpoint, such as incorrect report text format, invalid prompts for pathology analysis, model-specific errors.
// Actionable Recommendations: Verify report text format and ensure accurate OCR extraction, review and refine pathology analysis prompts, check Gemini API documentation for pathology report analysis specific errors.
type ErrGeminiPathologyAnalysisFailed struct {
	Code     string // ADDED: Error Code - Recommendation: Error Codes or Enums
	Severity string // ADDED: Severity Level - Recommendation: Error Classification and Severity Levels
	Message  string
	Err      error
	logger   *zap.Logger
}

func (e *ErrGeminiPathologyAnalysisFailed) Error() string {
	return fmt.Sprintf("[%s] Gemini Pathology Report Analysis API call failed [%s]: %s - %v", e.Severity, e.Code, e.Message, e.Err) // ADDED: Include Severity Level and Error Code
}

func (e *ErrGeminiPathologyAnalysisFailed) Unwrap() error {
	return e.Err
}

// NewErrGeminiPathologyAnalysisFailed creates a new ErrGeminiPathologyAnalysisFailed. Defaults to Warning severity. - Recommendation: Error Classification and Severity Levels
func NewErrGeminiPathologyAnalysisFailed(message string, err error) *ErrGeminiPathologyAnalysisFailed {
	return &ErrGeminiPathologyAnalysisFailed{
		Code:     ErrCodeGeminiPathologyAnalysisFailed, // ADDED: Error Code - Recommendation: Error Codes or Enums
		Severity: SeverityWarning,                      // Default Severity - Recommendation: Error Classification and Severity Levels
		Message:  message,
		Err:      err,
		logger:   nil,
	}
}

// SetLogger implements the SetLogger method for ErrGeminiPathologyAnalysisFailed.
func (e *ErrGeminiPathologyAnalysisFailed) SetLogger(logger *zap.Logger) {
	e.logger = logger
}

// ErrGeminiInformationExtractionFailed represents an error during Gemini Information Extraction API call. // ADDED: Information Extraction Error - Recommendation 1
//
// Error Code: ErrCodeGeminiInformationExtractionFailed - Recommendation: More Granular Error Types & Grouping
// Severity Level: Warning - Recommendation: Error Classification and Severity Levels
// Possible Causes: Issues specific to the Information Extraction API endpoint, such as incorrect report text format, overly broad or narrow prompts for information extraction, model-specific errors.
// Actionable Recommendations: Verify report text format and ensure accurate OCR extraction, review and refine information extraction prompts to be more specific and targeted, check Gemini API documentation for information extraction specific errors.
type ErrGeminiInformationExtractionFailed struct {
	Code     string // ADDED: Error Code - Recommendation: Error Codes or Enums
	Severity string // ADDED: Severity Level - Recommendation: Error Classification and Severity Levels
	Message  string
	Err      error
	logger   *zap.Logger
}

func (e *ErrGeminiInformationExtractionFailed) Error() string {
	return fmt.Sprintf("[%s] Gemini Information Extraction API call failed [%s]: %s - %v", e.Severity, e.Code, e.Message, e.Err) // ADDED: Include Severity Level and Error Code
}

func (e *ErrGeminiInformationExtractionFailed) Unwrap() error {
	return e.Err
}

// NewErrGeminiInformationExtractionFailed creates a new ErrGeminiInformationExtractionFailed. Defaults to Warning severity. - Recommendation: Error Classification and Severity Levels
func NewErrGeminiInformationExtractionFailed(message string, err error) *ErrGeminiInformationExtractionFailed {
	return &ErrGeminiInformationExtractionFailed{
		Code:     ErrCodeGeminiInformationExtractionFailed, // ADDED: Error Code - Recommendation: Error Codes or Enums
		Severity: SeverityWarning,                          // Default Severity - Recommendation: Error Classification and Severity Levels
		Message:  message,
		Err:      err,
		logger:   nil,
	}
}

// SetLogger implements the SetLogger method for ErrGeminiInformationExtractionFailed.
func (e *ErrGeminiInformationExtractionFailed) SetLogger(logger *zap.Logger) {
	e.logger = logger
}

// ErrGeminiDiagnosisFailed represents an error during Gemini Preliminary Diagnosis API call. // ADDED: Gemini Preliminary Diagnosis Error - Recommendation 1
// ... (ErrGeminiDiagnosisFailed struct and methods - No changes needed, already defined in previous response) ...

// ErrGeminiStagingFailed represents an error during Gemini Staging API call. // ADDED: Gemini Staging Error - Recommendation 1
// ... (ErrGeminiStagingFailed struct and methods - No changes needed, already defined in previous response) ...

// ErrGeminiTreatmentSuggestFailed represents an error during Gemini Treatment Suggestion API call. // ADDED: Gemini Treatment Suggestion Error - Recommendation 1
// ... (ErrGeminiTreatmentSuggestFailed struct and methods - No changes needed, already defined in previous response) ...

// Is function for Custom Errors
func (e *ErrGeminiAPIError) Is(target error) bool {
	_, ok := target.(*ErrGeminiAPIError)
	return ok
}

// Is function for Custom Errors
func (e *ErrGeminiResponseParsingFailed) Is(target error) bool {
	_, ok := target.(*ErrGeminiResponseParsingFailed)
	return ok
}

// Is function for Custom Errors
func (e *ErrGeminiRequestFailed) Is(target error) bool {
	_, ok := target.(*ErrGeminiRequestFailed)
	return ok
}

// Is function for Custom Errors - Recommendation 1 - ADDED Is for ErrGeminiNoduleDetectionFailed
func (e *ErrGeminiNoduleDetectionFailed) Is(target error) bool {
	_, ok := target.(*ErrGeminiNoduleDetectionFailed)
	return ok
}

// Is function for Custom Errors - Recommendation 1 - ADDED Is for ErrGeminiPathologyAnalysisFailed
func (e *ErrGeminiPathologyAnalysisFailed) Is(target error) bool {
	_, ok := target.(*ErrGeminiPathologyAnalysisFailed)
	return ok
}

// Is function for Custom Errors - Recommendation 1 - ADDED Is for ErrGeminiInformationExtractionFailed
func (e *ErrGeminiInformationExtractionFailed) Is(target error) bool {
	_, ok := target.(*ErrGeminiInformationExtractionFailed)
	return ok
}

// Is function for Custom Errors - Recommendation 1 - ADDED Is for ErrGeminiDiagnosisFailed
// ... (Is function for ErrGeminiDiagnosisFailed - No changes needed, already defined in previous response) ...

// Is function for Custom Errors - Recommendation 1 - ADDED Is for ErrGeminiStagingFailed
// ... (Is function for ErrGeminiStagingFailed - No changes needed, already defined in previous response) ...

// Is function for Custom Errors - Recommendation 1 - ADDED Is for ErrGeminiTreatmentSuggestFailed
// ... (Is function for ErrGeminiTreatmentSuggestFailed - No changes needed, already defined in previous response) ...

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
