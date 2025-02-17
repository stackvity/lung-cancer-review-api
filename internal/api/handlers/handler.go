// internal/api/handlers/handler.go
package handlers

// Handler is a struct that groups all handlers for the API.
// This is used for dependency injection in api.go and wire.go,
// making it easier to manage and inject all handlers as a single dependency.
type Handler struct {
	FileHandler      *FileHandler
	ReportHandler    *ReportHandler
	HealthHandler    *HealthHandler
	DiagnosisHandler *DiagnosisHandler
	// Add other handlers here as you create them (e.g., AdminHandler, etc.)
}

// NewHandler creates a new Handler instance, injecting all handler dependencies.
// It takes instances of each individual handler as arguments and groups them into the Handler struct.
// This constructor ensures that the Handler struct is properly initialized with all its required handler dependencies,
// facilitating dependency management and injection throughout the API layer.
func NewHandler(
	fileHandler *FileHandler,
	reportHandler *ReportHandler,
	healthHandler *HealthHandler,
	diagnosisHandler *DiagnosisHandler,
	// Inject other handlers here as arguments (e.g., adminHandler *AdminHandler)
) *Handler {
	return &Handler{
		FileHandler:      fileHandler,
		ReportHandler:    reportHandler,
		HealthHandler:    healthHandler,
		DiagnosisHandler: diagnosisHandler,
		// Initialize other handlers here (e.g., AdminHandler: adminHandler)
	}
}
