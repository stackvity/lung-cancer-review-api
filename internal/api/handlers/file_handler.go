// internal/api/handlers/file_handler.go
package handlers

import (
	"net/http"
	"strings" // ADDED: Import strings package for Content-Type check

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stackvity/lung-server/internal/domain/services" // Import services
	"github.com/stackvity/lung-server/internal/utils"           // Import utils
	"go.uber.org/zap"
)

// FileHandler handles HTTP requests related to file uploads and file operations.
// It depends on the ProcessingService to perform the actual file processing and storage logic.
type FileHandler struct {
	ProcessingService *services.ProcessingService // Exported ProcessingService field (capital 'P')
	logger            *zap.Logger
}

// NewFileHandler creates a new FileHandler instance, injecting the required ProcessingService and Logger.
func NewFileHandler(processingService *services.ProcessingService, logger *zap.Logger) *FileHandler {
	return &FileHandler{
		ProcessingService: processingService, // Use exported ProcessingService field
		logger:            logger.Named("FileHandler"),
	}
}

// UploadFile handles the file upload HTTP request.
// It receives a file from the client, validates the Content-Type header, and file retrieval,
// and then uses the ProcessingService to process the uploaded file.
// It is responsible for handling HTTP-specific tasks such as request parsing, response writing, and error handling at the API level.
func (h *FileHandler) UploadFile(c *gin.Context) {
	const operation = "FileHandler.UploadFile" // Define operation name for structured logging
	requestID := utils.GetRequestID(c.Request.Context())

	h.logger.Info("Start handling file upload request", zap.String("operation", operation), zap.String("request_id", requestID))

	// 1. Extract Patient ID from Context (Middleware responsibility) - BE-003, US-001
	patientIDRaw, exists := c.Get("patientID") // Retrieve patientID from Gin context, set by LinkValidationMiddleware
	if !exists {
		h.logger.Error("Patient ID not found in context", zap.String("operation", operation), zap.String("request_id", requestID)) // Log error if patientID is missing in context
		utils.RespondWithError(c, http.StatusBadRequest, "Patient ID missing from request context")                                // Respond with 400 Bad Request if patientID is not found
		return                                                                                                                     // Abort handler execution
	}

	patientID, ok := patientIDRaw.(uuid.UUID) // Type assert patientID from interface{} to uuid.UUID
	if !ok {
		h.logger.Error("Invalid patient ID format in context", zap.String("operation", operation), zap.String("request_id", requestID), zap.Any("patient_id_raw", patientIDRaw)) // Log error if patientID is not UUID
		utils.RespondWithError(c, http.StatusInternalServerError, "Invalid patient ID format")                                                                                   // Respond with 500 Internal Server Error for type mismatch
		return                                                                                                                                                                   // Abort handler execution
	}

	// Enhanced Input Validation - Recommendation 1 (Enhanced Validation in UploadFile Handler)
	// 2. Validate Content-Type Header - Ensure request Content-Type is multipart/form-data
	if !strings.HasPrefix(c.Request.Header.Get("Content-Type"), "multipart/form-data") { // Check if Content-Type header starts with multipart/form-data
		h.logger.Warn("Invalid Content-Type header", zap.String("operation", operation), zap.String("request_id", requestID), zap.String("content_type", c.Request.Header.Get("Content-Type"))) // Log warning for invalid content type
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid Content-Type header, expected multipart/form-data")                                                                           // Respond with 400 Bad Request for invalid header
		return                                                                                                                                                                                  // Abort handler execution for invalid Content-Type
	}

	// 3. Get File from Request - BE-010, US-003
	fileHeader, err := c.FormFile("file") // Extract the uploaded file from the request form data, key "file"
	if err != nil {
		h.logger.Warn("File upload failed to get file from request", zap.String("operation", operation), zap.String("request_id", requestID), zap.Error(err)) // Log warning if file extraction fails
		utils.RespondWithError(c, http.StatusBadRequest, "Error retrieving file from request")                                                                // Respond with 400 Bad Request for file retrieval error
		return                                                                                                                                                // Abort handler execution
	}
	filename := fileHeader.Filename                      // Extract filename from the file header
	contentType := fileHeader.Header.Get("Content-Type") // Extract content type from file header

	// 4. Open the uploaded file for reading - needed for subsequent processing
	file, err := fileHeader.Open()
	if err != nil {
		h.logger.Error("File upload failed to open file", zap.String("operation", operation), zap.String("request_id", requestID), zap.String("filename", filename), zap.Error(err)) // Log error if file opening fails
		utils.RespondWithError(c, http.StatusInternalServerError, "Error opening uploaded file")                                                                                     // Respond with 500 Internal Server Error if file cannot be opened
		return                                                                                                                                                                       // Abort handler execution
	}
	defer file.Close() // Ensure file is closed after handler execution

	// 5. Pass to Processing Service - Delegate file processing to the ProcessingService
	if err := h.ProcessingService.ProcessDocument(c.Request.Context(), patientID, filename, contentType, file); err != nil { // Call ProcessDocument on ProcessingService to handle business logic
		h.logger.Error("File processing failed", zap.String("operation", operation), zap.String("request_id", requestID), zap.String("filename", filename), zap.Error(err)) // Log error if processing fails
		utils.RespondWithError(c, http.StatusInternalServerError, "File processing failed")                                                                                 // Respond with 500 Internal Server Error if processing fails
		return                                                                                                                                                              // Abort handler execution
	}

	// 6. Respond with Success - If file processing is successful, respond to the client with 202 Accepted status
	c.JSON(http.StatusAccepted, gin.H{ // Respond with 202 Accepted to indicate successful upload and processing initiation
		"message":   "File uploaded and is being processed", // Success message for client
		"upload_id": patientID,                              // Include patientID (acting as upload_id for this session) in response
	})
	h.logger.Info("File upload request handled successfully", zap.String("operation", operation), zap.String("request_id", requestID), zap.String("filename", filename), zap.String("patient_id", patientID.String())) // Log successful handling of upload request

	// Future Enhancement: Idempotency Key Handling - Recommendation 3 (Idempotency)
	// TODO: Implement idempotency key extraction from headers (e.g., "Idempotency-Key").
	// TODO: Check for existing requests with the same idempotency key in a data store (e.g., Redis, database).
	// TODO: If a duplicate request is detected, return the previous response instead of reprocessing.
	// TODO: Store the idempotency key and response status for each request to prevent reprocessing.

}

// GetUploadStatus is a placeholder for future implementation. - Recommendation 2 (Placeholder for GetUploadStatus)
func (h *FileHandler) GetUploadStatus(c *gin.Context) {
	// Placeholder for GetUploadStatus handler function (BE-017) - Recommendation 2
	// TODO: Implement logic to: - Recommendation 2
	// 1. Extract upload_id from request parameters (path or query). - Recommendation 2
	// 2. Query a data store (e.g., database, in-memory cache) to retrieve the processing status for the given upload_id. - Recommendation 2
	// 3. Return the processing status in JSON format to the client. - Recommendation 2
	utils.RespondWithError(c, http.StatusNotImplemented, "Not implemented - Functionality to be implemented in future sprints") // Respond with 501 Not Implemented as it's a placeholder, updated message
}

// Future Enhancement: Request Size Limiting - Recommendation 4 (Request Size Middleware)
// TODO: Implement Request Size Limiting Middleware (using gin.LimitBodySize or custom middleware)
// and register it in routes/routes.go to protect against large request attacks.
// This can be done as a separate middleware function and applied to the /upload route.

// Future Enhancement: Content-Type Validation for other handlers- Recommendation 5 (Proactive Validation)
// TODO: Apply Content-Type validation to other handlers (e.g., structured data input handlers)
// to ensure they only process requests with the expected Content-Type headers.
