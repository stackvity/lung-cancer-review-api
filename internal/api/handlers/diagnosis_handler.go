// internal/api/handlers/diagnosis_handler.go
package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stackvity/lung-server/internal/domain"          // Import domain for custom errors
	"github.com/stackvity/lung-server/internal/domain/services" // Import services
	"github.com/stackvity/lung-server/internal/utils"           // Import utils
	"go.uber.org/zap"
)

// DiagnosisHandler handles HTTP requests related to diagnosis, staging, and treatment recommendations.
// It depends on the DiagnosisService to orchestrate the business logic for these operations.
type DiagnosisHandler struct {
	diagnosisService *services.DiagnosisService
	logger           *zap.Logger
}

// NewDiagnosisHandler creates a new DiagnosisHandler instance, injecting the required DiagnosisService and Logger.
func NewDiagnosisHandler(diagnosisService *services.DiagnosisService, logger *zap.Logger) *DiagnosisHandler {
	return &DiagnosisHandler{
		diagnosisService: diagnosisService,
		logger:           logger.Named("DiagnosisHandler"),
	}
}

// GeneratePreliminaryDiagnosisHandler handles the HTTP request to generate a preliminary diagnosis.
// It extracts the patient ID from the request context (set by middleware),
// calls the DiagnosisService to generate the preliminary diagnosis, and then sends the diagnosis in the JSON response.
// This handler is responsible for API-specific tasks like request parsing, response formatting, and robust error handling at the API level for diagnosis operations.
func (h *DiagnosisHandler) GeneratePreliminaryDiagnosisHandler(c *gin.Context) {
	const operation = "DiagnosisHandler.GeneratePreliminaryDiagnosisHandler" // Operation name for structured logging
	requestID := utils.GetRequestID(c.Request.Context())                     // Retrieve request ID from context

	h.logger.Info("Starting preliminary diagnosis request", zap.String("operation", operation), zap.String("request_id", requestID)) // Info log indicating start of handler execution.

	// 1. Extract Patient ID from Context (Middleware responsibility) - The Patient ID should have been placed in the Gin context by the LinkValidationMiddleware.
	patientIDRaw, exists := c.Get("patientID") // Retrieve patientID from Gin context using the key "patientID".
	if !exists {
		h.logger.Error("Patient ID not found in context", zap.String("operation", operation), zap.String("request_id", requestID)) // Error log if patientID is missing, indicating a middleware or context setup issue.
		utils.RespondWithError(c, http.StatusBadRequest, "Patient ID missing from request context")                                // Respond with 400 Bad Request, indicating client-side error due to missing patient ID.
		return                                                                                                                     // Abort handler execution as Patient ID is essential for processing.
	}

	patientID, ok := patientIDRaw.(uuid.UUID) // Type assert patientID from interface{} to uuid.UUID to ensure correct data type.
	if !ok {
		h.logger.Error("Invalid patient ID format in context", zap.String("operation", operation), zap.String("request_id", requestID), zap.Any("patient_id_raw", patientIDRaw)) // Error log for invalid patientID format, indicating a middleware or context issue.
		utils.RespondWithError(c, http.StatusInternalServerError, "Invalid patient ID format")                                                                                   // Respond with 500 Internal Server Error, indicating server-side error due to unexpected data type.
		return                                                                                                                                                                   // Abort handler execution due to invalid Patient ID.
	}

	// Enhanced Input Validation - Recommendation 4 (Input Validation in Handler) - Adding handler-level validation for robustness.
	// 2. Validate Patient ID Format (even though middleware validates, adding handler-level validation for robustness) - Re-validate patientID format within the handler for defense-in-depth.
	if err := utils.ValidateUUID(patientID.String()); err != nil {
		h.logger.Warn("Invalid patient ID format", zap.String("operation", operation), zap.String("request_id", requestID), zap.String("patient_id", patientID.String()), zap.Error(err)) // Warn log for invalid patient ID format at handler level.
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid patient ID format")                                                                                                     // Respond with 400 Bad Request, indicating client-side error due to invalid Patient ID format.
		return                                                                                                                                                                            // Abort handler execution due to invalid Patient ID.
	}

	// 3. Call Diagnosis Service to Generate Preliminary Diagnosis - Delegate the core business logic of diagnosis generation to the DiagnosisService.
	diagnosis, err := h.diagnosisService.GeneratePreliminaryDiagnosis(c.Request.Context(), patientID) // Call the diagnosis service, passing the context and patientID.
	if err != nil {
		h.logger.Error("Preliminary diagnosis generation failed", zap.String("operation", operation), zap.String("request_id", requestID), zap.String("patient_id", patientID.String()), zap.Error(err)) // Error log for diagnosis generation failure in the service layer.

		// Enhanced Error Handling: Check for specific error types from service and respond accordingly - Recommendation 2 (Enhanced Error Handling) - Differentiated error responses based on error type.
		if errors.Is(err, &domain.ErrGeminiDiagnosisFailed{}) { // Check if the error is specifically a Gemini API related error using errors.Is and type assertion.
			utils.RespondWithError(c, http.StatusServiceUnavailable, "Diagnosis service unavailable: "+err.Error()) // Respond with 503 Service Unavailable if Gemini API is the root cause, indicating external service dependency issue.
		} else {
			// General Internal Server Error for other unclassified service errors - Fallback for any other errors from the service layer that are not Gemini API related.
			utils.RespondWithError(c, http.StatusInternalServerError, "Failed to generate preliminary diagnosis: "+err.Error()) // Respond with 500 Internal Server Error for generic service errors.
		}
		return // Abort handler execution after error response.
	}

	// 4. Respond with JSON - Construct and send a successful JSON response to the client.
	c.JSON(http.StatusOK, gin.H{ // Respond with 200 OK status code, indicating successful processing of the request.
		"message":   "Preliminary diagnosis generated successfully", // Success message to inform the client about the operation outcome.
		"diagnosis": diagnosis,                                      // Include the generated diagnosis data in the response payload.
	})

	h.logger.Info("Preliminary diagnosis request handled successfully", zap.String("operation", operation), zap.String("request_id", requestID), zap.String("patient_id", patientID.String())) // Info log for successful handler execution.
}

// GetStagingInformationHandler handles the HTTP request to retrieve staging information.
// This handler is a placeholder and is not yet implemented in this version.
// It currently returns a "Not Implemented" response, as per the project's incremental development plan.
// TODO: Implement GetStagingInformationHandler - BE-041, BE-048a - Implementation task ID for future sprints.
func (h *DiagnosisHandler) GetStagingInformationHandler(c *gin.Context) {
	const operation = "DiagnosisHandler.GetStagingInformationHandler"
	requestID := utils.GetRequestID(c.Request.Context())

	h.logger.Warn("Get Staging Information Handler - Not Implemented", zap.String("operation", operation), zap.String("request_id", requestID), zap.String("handler_status", "placeholder")) // Logging placeholder handler - Recommendation 1 and 2

	// Placeholder response with "Not Implemented" status code and user-friendly message - Recommendation 1 and 2
	utils.RespondWithError(c, http.StatusNotImplemented, "Staging information functionality is not yet implemented in this version") // Updated message - Recommendation 1 and 2
}

// SuggestTreatmentOptionsHandler handles the HTTP request to get treatment recommendations.
// This handler is a placeholder and is not yet implemented in this version.
// It currently returns a "Not Implemented" response, as per the project's incremental development plan.
// TODO: Implement SuggestTreatmentOptionsHandler - BE-043, BE-048a - Implementation task ID for future sprints.
func (h *DiagnosisHandler) SuggestTreatmentOptionsHandler(c *gin.Context) {
	const operation = "DiagnosisHandler.SuggestTreatmentOptionsHandler"
	requestID := utils.GetRequestID(c.Request.Context())

	h.logger.Warn("Suggest Treatment Options Handler - Not Implemented", zap.String("operation", operation), zap.String("request_id", requestID), zap.String("handler_status", "placeholder")) // Logging placeholder handler - Recommendation 1 and 2

	// Placeholder response with "Not Implemented" status code and user-friendly message - Recommendation 1 and 2
	utils.RespondWithError(c, http.StatusNotImplemented, "Treatment recommendation functionality is not yet implemented in this version") // Updated message - Recommendation 1 and 2
}
