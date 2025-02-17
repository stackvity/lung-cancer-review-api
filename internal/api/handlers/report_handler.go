// internal/api/handlers/report_handler.go
package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stackvity/lung-server/internal/domain/services"
	"github.com/stackvity/lung-server/internal/utils"
	"go.uber.org/zap"
)

// ReportHandler handles HTTP requests related to report generation and retrieval.
// It depends on the ReportService to perform the actual report generation logic.
type ReportHandler struct {
	reportService *services.ReportService
	logger        *zap.Logger
}

// NewReportHandler creates a new ReportHandler instance, injecting the required ReportService and Logger.
// This constructor ensures that the ReportHandler has access to the necessary business logic and logging capabilities.
func NewReportHandler(reportService *services.ReportService, logger *zap.Logger) *ReportHandler {
	return &ReportHandler{
		reportService: reportService,
		logger:        logger.Named("ReportHandler"),
	}
}

// GenerateReportHandler handles the HTTP request to generate a patient report.
// It extracts the patient ID from the request context (set by middleware),
// calls the ReportService to generate the report, and then sends the report file path in the response.
// It is responsible for handling HTTP-specific tasks such as request parsing, response writing, and error handling at the API level.
func (h *ReportHandler) GenerateReportHandler(c *gin.Context) {
	const operation = "ReportHandler.GenerateReportHandler"
	requestID := utils.GetRequestID(c.Request.Context())

	h.logger.Info("Starting report generation request", zap.String("operation", operation), zap.String("request_id", requestID))

	patientIDRaw, exists := c.Get("patientID")
	if !exists {
		h.logger.Error("Patient ID not found in context", zap.String("operation", operation), zap.String("request_id", requestID))
		utils.RespondWithError(c, http.StatusBadRequest, "Patient ID missing from request context")
		return
	}

	patientID, ok := patientIDRaw.(uuid.UUID)
	if !ok {
		h.logger.Error("Invalid patient ID format in context", zap.String("operation", operation), zap.String("request_id", requestID), zap.Any("patient_id_raw", patientIDRaw))
		utils.RespondWithError(c, http.StatusInternalServerError, "Invalid patient ID format")
		return
	}

	filePath, err := h.reportService.GenerateReport(c.Request.Context(), patientID)
	if err != nil {
		h.logger.Error("Report generation failed", zap.String("operation", operation), zap.String("request_id", requestID), zap.String("patient_id", patientID.String()), zap.Error(err))
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to generate report")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Report generated successfully",
		"file_path": filePath,
		// Future Enhancement (Serve File Directly):
		// In future sprints, consider serving the PDF file directly instead of returning the file path.
		// This can be done using c.FileAttachment(filePath, "report.pdf") or c.File(filePath) for inline display.
		// When serving directly, ensure to set the correct Content-Type header (application/pdf)
		// and Content-Disposition header if you want to force a download dialog.
	})

	h.logger.Info("Report generation request handled successfully", zap.String("operation", operation), zap.String("request_id", requestID), zap.String("patient_id", patientID.String()), zap.String("file_path", filePath))
}

// GetReport is a placeholder for future implementation. - Recommendation 2 (Placeholder for GetReport)
func (h *ReportHandler) GetReport(c *gin.Context) {
	// Placeholder for GetReport handler function (BE-033) - Recommendation 2
	// TODO: Implement logic to: - Recommendation 2
	// 1. Extract upload_id/report_id from request parameters (path or query). - Recommendation 2
	// 2. Retrieve the generated report (PDF file) from storage using the report_id. - Recommendation 2
	// 3. Serve the PDF file to the client using c.FileAttachment or c.File, setting appropriate Content-Type headers. - Recommendation 2
	//    - Use c.FileAttachment(filePath, "report.pdf") to force a download dialog, prompting the user to save the file.
	//      This is generally preferred for reports to ensure the user can easily save the document.
	//      Example:  c.FileAttachment(filePath, "report.pdf")
	//
	//    - Use c.File(filePath) to serve the file inline in the browser if appropriate for certain use cases.
	//      For inline display, the browser will attempt to render the PDF within the browser window, if supported.
	//      Example: c.File(filePath)
	//
	//    - When serving the file (using either c.FileAttachment or c.File), ensure to set the Content-Type header to "application/pdf" for PDF files:
	//      c.Header("Content-Type", "application/pdf") // Example of setting Content-Type header
	//
	// 4. Handle cases where the report is not found or an error occurs during retrieval. - Recommendation 2
	utils.RespondWithError(c, http.StatusNotImplemented, "Not implemented - Functionality to retrieve and serve the generated report to be implemented in future sprints") // Respond with 501 Not Implemented as it's a placeholder, updated message
}
