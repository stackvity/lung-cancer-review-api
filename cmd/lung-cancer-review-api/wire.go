package main

import (
	"github.com/google/wire"
	"github.com/stackvity/lung-server/internal/api"
	"github.com/stackvity/lung-server/internal/api/handlers"
	"github.com/stackvity/lung-server/internal/config"
	"github.com/stackvity/lung-server/internal/data"
	"github.com/stackvity/lung-server/internal/data/repositories/interfaces"
	postgresRepo "github.com/stackvity/lung-server/internal/data/repositories/postgres" // Alias for clarity
	"github.com/stackvity/lung-server/internal/domain/services"                         // Corrected import: Explicitly import services package
	"github.com/stackvity/lung-server/internal/gemini"
	"github.com/stackvity/lung-server/internal/knowledge"
	"github.com/stackvity/lung-server/internal/ocr"
	"github.com/stackvity/lung-server/internal/security" // Import security package
	"github.com/stackvity/lung-server/internal/storage"
	"github.com/stackvity/lung-server/internal/utils"
)

// --- Wire Sets ---

// serviceSet: Wire set for service layer dependencies.
// Defines providers for all services, including ProcessingService, ReportService, and LinkService.
// Binds concrete service implementations to their interface types for dependency injection.
var serviceSet = wire.NewSet(
	services.NewProcessingService, // Provider for Processing Service
	services.NewReportService,     // Provider for Report Service
	services.NewLinkService,       // Provider for Link Service
	services.NewDiagnosisService,  // Provider for Diagnosis Service
)

// repositorySet: Wire set for repository layer dependencies.
// Includes providers for all repository implementations (PostgreSQL), and interface bindings.
// This set ensures that the service layer and other components depend on repository interfaces, not concrete implementations, promoting loose coupling.
var repositorySet = wire.NewSet(
	postgresRepo.NewPatientRepository,                                                                                  // Provider for PatientRepository (PostgreSQL implementation)
	postgresRepo.NewStudyRepository,                                                                                    // Provider for StudyRepository (PostgreSQL implementation)
	postgresRepo.NewImageRepository,                                                                                    // Provider for ImageRepository (PostgreSQL implementation)
	postgresRepo.NewReportRepository,                                                                                   // Provider for ReportRepository (PostgreSQL implementation)
	postgresRepo.NewNoduleRepository,                                                                                   // Provider for NoduleRepository (PostgreSQL implementation)
	postgresRepo.NewDiagnosisRepository,                                                                                // Provider for DiagnosisRepository (PostgreSQL implementation)
	postgresRepo.NewStageRepository,                                                                                    // Provider for StageRepository (PostgreSQL implementation)
	postgresRepo.NewTreatmentRecommendationRepository,                                                                  // Provider for TreatmentRecommendationRepository (PostgreSQL implementation)
	wire.Bind(new(interfaces.PatientRepository), new(*postgresRepo.PatientRepository)),                                 // Binds PatientRepository interface to its PostgreSQL implementation
	wire.Bind(new(interfaces.StudyRepository), new(*postgresRepo.StudyRepository)),                                     // Binds StudyRepository interface to its PostgreSQL implementation
	wire.Bind(new(interfaces.ImageRepository), new(*postgresRepo.ImageRepository)),                                     // Binds ImageRepository interface to its PostgreSQL implementation
	wire.Bind(new(interfaces.ReportRepository), new(*postgresRepo.ReportRepository)),                                   // Binds ReportRepository interface to its PostgreSQL implementation
	wire.Bind(new(interfaces.NoduleRepository), new(*postgresRepo.NoduleRepository)),                                   // Binds NoduleRepository interface to its PostgreSQL implementation
	wire.Bind(new(interfaces.DiagnosisRepository), new(*postgresRepo.DiagnosisRepository)),                             // Binds DiagnosisRepository interface to its PostgreSQL implementation
	wire.Bind(new(interfaces.StageRepository), new(*postgresRepo.StageRepository)),                                     // Binds StageRepository interface to its PostgreSQL implementation
	wire.Bind(new(interfaces.TreatmentRecommendationRepository), new(*postgresRepo.TreatmentRecommendationRepository)), // Binds TreatmentRecommendationRepository interface to its PostgreSQL implementation
)

// handlerSet: Wire set for API handler dependencies.
// Defines providers for all API handlers and the grouped Handler struct.
// This set ensures that the API layer has access to all necessary handlers for request processing.
var handlerSet = wire.NewSet(
	handlers.NewFileHandler,      // Provider for FileHandler
	handlers.NewReportHandler,    // Provider for Report Handler
	handlers.NewHealthHandler,    // Provider for Health Handler
	handlers.NewDiagnosisHandler, // Provider for Diagnosis Handler
	handlers.NewHandler,          // Provider for the grouped Handler struct
)

// geminiSet: Wire set for Gemini API client dependency.
// Includes the provider for the GeminiProClient and binds it to the GeminiClient interface.
// This set facilitates the integration with the Google AI Gemini Pro API for AI functionalities.
var geminiSet = wire.NewSet(
	gemini.NewGeminiProClient, // Provider for GeminiProClient (concrete implementation for Gemini API)
	wire.Bind(new(gemini.GeminiClient), new(*gemini.GeminiProClient)), // Binds GeminiClient interface to its concrete implementation (GeminiProClient)
)

// ocrSet: Wire set for OCR service dependency.
// Defines the provider for the GoogleVisionService and binds it to the OCRService interface.
// This set enables the use of Google Cloud Vision API for Optical Character Recognition tasks within the application.
var ocrSet = wire.NewSet(
	ocr.NewGoogleVisionService,                                    // Provider for GoogleVisionService (concrete implementation using Google Cloud Vision API)
	wire.Bind(new(ocr.OCRService), new(*ocr.GoogleVisionService)), // Binds OCRService interface to its concrete implementation (GoogleVisionService)
)

// storageSet: Wire set for storage service dependency.
// Includes the provider for CloudStorage and binds it to the FileStorage interface.
// This set allows the application to utilize cloud-based file storage (e.g., AWS S3) for temporary file persistence and management.
var storageSet = wire.NewSet(
	storage.NewCloudStorage, // Provider for CloudStorage (concrete implementation using AWS S3 or similar)
	wire.Bind(new(storage.FileStorage), new(*storage.CloudStorage)), // Binds FileStorage interface to its concrete implementation (CloudStorage)
)

// knowledgeSet: Wire set for knowledge base dependency.
// Defines the provider for the MockKnowledgeBase and binds it to the KnowledgeBase interface.
// In the current implementation, a mock knowledge base is used. This set can be replaced with a provider for a real knowledge base in future sprints (BE-005, BE-048a, BE-006, BE-064, KB-001, KB-002, KB-003, KB-004, BE-070, BE-081, US-005, US-006, US-007, US-008, US-009, US-010, US-011, US-012, US-015, US-020).
var knowledgeSet = wire.NewSet(
	knowledge.NewMockKnowledgeBase,                                             // Provider for MockKnowledgeBase (mock implementation for testing and development)
	wire.Bind(new(knowledge.KnowledgeBase), new(*knowledge.MockKnowledgeBase)), // Binds KnowledgeBase interface to its mock implementation (MockKnowledgeBase)
)

// utilsSet: Wire set for utility dependencies.
var utilsSet = wire.NewSet(
	security.NewValidator, // Provider for Validator
	utils.NewLogger,       // Provider for Logger
)

// configSet: Wire set for configuration.
var configSet = wire.NewSet(
	config.LoadConfig, // Provider for configuration loading
)

// apiSet: Wire set for API.
var apiSet = wire.NewSet(
	api.NewAPI, // Provider for API
)

// dataSet: Wire set for data layer.
var dataSet = wire.NewSet(
	data.NewData, // Provider for Data layer
)

// InitializeAPI: Assembles API dependencies using Wire.
func InitializeAPI() (*api.API, func(), error) {
	panic(wire.Build(
		configSet,     // Configuration Set
		utilsSet,      // Utility Set
		dataSet,       // Data Layer Set
		repositorySet, // Repository Set
		serviceSet,    // Service Set
		handlerSet,    // Handler Set
		geminiSet,     // Gemini Set
		ocrSet,        // OCR Set
		storageSet,    // Storage Set
		knowledgeSet,  // Knowledge Set
		apiSet,        // API Set
	))
}
