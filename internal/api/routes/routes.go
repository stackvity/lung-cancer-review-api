// internal/api/routes/routes.go
package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/stackvity/lung-server/internal/api/handlers" // Import handlers
)

// SetupRouter configures the application's routes using the Gin router.
// It defines the API endpoints and associates them with their corresponding handlers.
// This function sets up the entire routing structure for the backend API,
// including versioning and handler registration.
//
// It organizes routes into logical groups (e.g., API version 1, admin routes)
// and applies middleware where necessary to enforce security, data validation, and logging policies.
//
// Parameters:
//   - r *gin.Engine: The Gin router engine to which the routes will be attached.  This is the central routing component of the Gin web framework.
//   - h *handlers.FileHandler:  Handler for file-related endpoints.
//   - reportHandler *handlers.ReportHandler: Handler for report-related endpoints.
//   - healthHandler *handlers.HealthHandler: Handler for health check endpoints.
//   - diagnosisHandler *handlers.DiagnosisHandler: Handler for diagnosis-related endpoints.
func SetupRouter(
	r *gin.Engine,
	fileHandler *handlers.FileHandler, // Corrected: Use specific handler types instead of handlers.Handler
	reportHandler *handlers.ReportHandler, // Corrected: Use specific handler types instead of handlers.Handler
	healthHandler *handlers.HealthHandler, // Corrected: Use specific handler types instead of handlers.Handler
	diagnosisHandler *handlers.DiagnosisHandler, // Corrected: Use specific handler types instead of handlers.Handler
) {
	// --- API Version 1 Routes ---
	// Group for API version 1, under the path "/api/v1".
	// This group contains all patient-facing API endpoints for the Lung Cancer Review System.
	v1 := r.Group("/api/v1")
	{
		// --- Health Check Endpoints - Publicly accessible, no authentication or specific middleware needed ---
		// GET /api/v1/health: Endpoint for system health checks, used by load balancers and monitoring systems. - BE-006
		v1.GET("/health", healthHandler.HealthCheck) // Corrected: Use healthHandler parameter

		// --- File Upload Endpoints - Secure endpoints requiring access link validation and data quality checks ---
		// Apply LinkValidationMiddleware and DataQualityCheckMiddleware to the /upload group to ensure secure access and data integrity.
		fileUpload := v1.Group("/upload" /*, middleware.LinkValidationMiddleware(), middleware.DataQualityCheckMiddleware() */) // Example of commented-out middleware application for future implementation
		{
			// POST /api/v1/upload: Endpoint for patients to upload medical documents (DICOM, PDF, JPEG, PNG, CSV). - BE-010, US-003
			fileUpload.POST("", fileHandler.UploadFile) // Corrected: Use fileHandler parameter
			// GET /api/v1/status/:upload_id: Endpoint for retrieving the status of a file upload and processing job, using upload_id as a path parameter. - BE-017
			v1.GET("/status/:upload_id", fileHandler.GetUploadStatus) // Corrected: Use fileHandler parameter
		}

		// --- Report Endpoints - Secure endpoints requiring access link validation ---
		// Apply LinkValidationMiddleware to the /report group to secure access to patient reports.
		report := v1.Group("/report" /*, middleware.LinkValidationMiddleware() */) // Example of commented-out middleware application for future implementation
		{
			// GET /api/v1/report/:upload_id: Endpoint to generate and retrieve a patient-friendly PDF report, using upload_id as a path parameter. - BE-033, US-013
			report.GET("/:upload_id", reportHandler.GenerateReportHandler) // Corrected: Use reportHandler parameter
		}

		// --- Diagnosis Endpoints - Secure endpoints requiring access link validation (Future) ---
		diagnosis := v1.Group("/diagnosis" /*, middleware.LinkValidationMiddleware() */) // Example of grouped routes - to be implemented in later sprints
		{
			// GET /api/v1/diagnosis/preliminary/:upload_id: Placeholder for future preliminary diagnosis retrieval endpoint. - US-009, BE-039, BE-048a
			diagnosis.GET("/preliminary/:upload_id", diagnosisHandler.GeneratePreliminaryDiagnosisHandler) // Corrected: Use diagnosisHandler parameter
			// GET /api/v1/diagnosis/staging/:upload_id: Placeholder for future staging information retrieval endpoint. - US-010, BE-041, BE-048a
			diagnosis.GET("/staging/:upload_id", diagnosisHandler.GetStagingInformationHandler) // Corrected: Use diagnosisHandler parameter
			// GET /api/v1/diagnosis/treatment-options/:upload_id: Placeholder for future treatment options retrieval endpoint. - US-011, BE-043, BE-048a
			diagnosis.GET("/treatment-options/:upload_id", diagnosisHandler.SuggestTreatmentOptionsHandler) // Corrected: Use diagnosisHandler parameter
		}

		// --- Future Endpoints (Placeholders) - To be implemented in later sprints ---
		// v1.GET("/report/:report_id", h.ReportHandler.GetReport) // Placeholder for future GetReport functionality - Recommendation 2
		// v1.POST("/structured-data", h.DataHandler.ReceiveStructuredData) // Placeholder for structured data input - US-004
		// v1.POST("/feedback", h.FeedbackHandler.SubmitFeedback) // Placeholder for feedback submission - US-014
	}

	// --- Example Admin Routes Group (Not fully implemented in current scope) ---
	// Group for admin-specific functionalities, under the path "/api/v1/admin".
	// Requires separate admin authentication and authorization middleware (not implemented in this version).
	admin := v1.Group("/admin" /*, middleware.AdminAuthMiddleware(), middleware.AdminAccessControlMiddleware()*/) // Example of grouped admin routes with placeholder middleware
	{
		// --- Admin Monitoring Endpoints (Example) ---
		// GET /api/v1/admin/metrics: Placeholder for admin metrics endpoint (system performance monitoring). - US-017, BE-059
		admin.GET("/metrics", healthHandler.HealthCheck /* h.AdminHandler.GetMetrics*/) // Example placeholder for admin metrics endpoint - US-017, BE-059 // Corrected: Use healthHandler parameter
		// GET /api/v1/admin/audit-logs: Placeholder for admin audit logs endpoint (system activity tracking). - US-018, BE-060
		admin.GET("/audit-logs", healthHandler.HealthCheck /* h.AdminHandler.GetAuditLogs*/) // Example placeholder for admin audit logs endpoint - US-018, BE-060 // Corrected: Use healthHandler parameter
		// ... more admin routes ... (e.g., content management, prompt management, user management, etc.) - US-016, US-019, US-020, US-021, US-022, US-023
	}
}
