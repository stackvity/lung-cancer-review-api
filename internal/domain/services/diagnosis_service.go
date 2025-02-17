// internal/domain/services/diagnosis_service.go
package services

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stackvity/lung-server/internal/data/models"
	"github.com/stackvity/lung-server/internal/data/repositories/interfaces"
	"github.com/stackvity/lung-server/internal/domain"
	"github.com/stackvity/lung-server/internal/gemini"
	geminiModels "github.com/stackvity/lung-server/internal/gemini/models"
	"github.com/stackvity/lung-server/internal/knowledge"
	"github.com/stackvity/lung-server/internal/utils"
	"go.uber.org/zap"
)

// DiagnosisService encapsulates the business logic for generating preliminary diagnoses,
// staging information, and treatment recommendations.
type DiagnosisService struct {
	reportRepository interfaces.ReportRepository
	geminiClient     gemini.GeminiClient
	knowledgeBase    knowledge.KnowledgeBase
	logger           *zap.Logger
}

// NewDiagnosisService creates a new DiagnosisService instance.
func NewDiagnosisService(
	reportRepository interfaces.ReportRepository,
	geminiClient gemini.GeminiClient,
	knowledgeBase knowledge.KnowledgeBase,
	logger *zap.Logger,
) *DiagnosisService {
	return &DiagnosisService{
		reportRepository: reportRepository,
		geminiClient:     geminiClient,
		knowledgeBase:    knowledgeBase,
		logger:           logger.Named("DiagnosisService"),
	}
}

// GeneratePreliminaryDiagnosis orchestrates the generation of a preliminary diagnosis using the Gemini API.
func (s *DiagnosisService) GeneratePreliminaryDiagnosis(ctx context.Context, patientID uuid.UUID) (*models.Diagnosis, error) {
	const operation = "DiagnosisService.GeneratePreliminaryDiagnosis" // Corrected operation name for clarity
	requestID := utils.GetRequestID(ctx.(*gin.Context))

	s.logger.Info("Starting preliminary diagnosis generation", zap.String("operation", operation), zap.String("patient_id", patientID.String()), zap.String("request_id", requestID))

	// 1. Prepare Input for Gemini API - BE-039, BE-048a
	geminiInput := &geminiModels.DiagnosisInput{
		PatientID: patientID,
		Prompt:    "Generate a preliminary diagnosis for lung cancer based on the available medical information.", // Define prompt - consider refining prompt for better results
		// In real implementation, populate input with relevant patient data if needed from database
		// Example:
		//   reportData, err := s.reportRepository.GetReportByPatientID(ctx, patientID)
		//   if err != nil { return nil, fmt.Errorf("failed to retrieve report data: %w", err) }
		//   geminiInput.ReportText = reportData.ReportText // Assuming ReportText field in DiagnosisInput
		//   geminiInput.FindingsSummary = ... // Populate with findings summary if available
		//   geminiInput.PatientHistory = ... // Populate with patient history if available
	}

	// 2. Call Gemini API Client - BE-039, BE-048a
	geminiOutput, err := s.geminiClient.GeneratePreliminaryDiagnosis(ctx, geminiInput)
	if err != nil {
		geminiErr := domain.NewErrGeminiDiagnosisFailed("Gemini Preliminary Diagnosis API call failed", err) // BE-039 - Custom error type
		geminiErr.SetLogger(s.logger)                                                                        // Set logger for custom error for better context
		s.logger.Error("Gemini API call failed", zap.String("operation", operation), zap.String("request_id", requestID), zap.Error(geminiErr))
		return nil, geminiErr
	}

	// 3. Process Gemini API Output and Create Diagnosis Model - BE-039, BE-044, BE-048a
	diagnosis := &models.Diagnosis{ // BE-044 - Create Diagnosis model
		// ID:        uuid.New(), // Removed: ID is not set here, database will generate it. - Fixed issue: #2
		DiagnosisText: geminiOutput.Diagnosis,     // Extract diagnosis text from Gemini output
		Confidence:    geminiOutput.Confidence,    // Extract confidence level
		Justification: geminiOutput.Justification, // Extract justification
		SessionID:     patientID,                  // Assuming SessionID is the same as PatientID for this context - Corrected: SessionID is now correctly set. - Fixed issue: #1
	}

	// 4. (Optional) Integrate with Knowledge Base/Rules - BE-048a - Placeholder
	// In real implementation, knowledge base integration logic would be here to refine or validate the diagnosis
	// Example:
	//   refinedDiagnosis, err := s.knowledgeBase.RefineDiagnosis(ctx, diagnosis, patientData)
	//   if err != nil {
	//       s.logger.Warn("Knowledge base refinement failed, proceeding with preliminary diagnosis", zap.String("operation", operation), zap.Error(err))
	//       // Decide whether to return error or proceed with unrefined diagnosis
	//   } else {
	//       diagnosis = refinedDiagnosis // Use refined diagnosis from knowledge base
	//   }

	s.logger.Info("Successfully generated preliminary diagnosis", zap.String("operation", operation), zap.String("patient_id", patientID.String()), zap.String("request_id", requestID))
	return diagnosis, nil
}

// GetStagingInformation retrieves preliminary staging information using the Gemini API.
func (s *DiagnosisService) GetStagingInformation(ctx context.Context, patientID uuid.UUID) (*models.Stage, error) {
	const operation = "DiagnosisService.GetStagingInformation" // Corrected operation name
	requestID := utils.GetRequestID(ctx.(*gin.Context))

	s.logger.Info("Starting staging information retrieval", zap.String("operation", operation), zap.String("patient_id", patientID.String()), zap.String("request_id", requestID))

	// 1. Prepare Input for Gemini API - BE-041, BE-048a
	geminiInput := &geminiModels.StagingInput{
		PatientID: patientID,
		Prompt:    "Get TNM staging information for lung cancer based on the available medical information.", // Define prompt - consider refining prompt
		// In real implementation, populate input with relevant patient data if needed
		// Example:
		//   reportData, err := s.reportRepository.GetReportByPatientID(ctx, patientID)
		//   if err != nil {  return nil, fmt.Errorf("failed to get report data: %w", err) }
		//   geminiInput.ReportText = reportData.ReportText // Assuming ReportText field in StagingInput
		//   geminiInput.FindingsSummary = ... // Populate with findings summary if available
	}

	// 2. Call Gemini API Client - BE-041, BE-048a
	geminiOutput, err := s.geminiClient.GetStagingInformation(ctx, geminiInput)
	if err != nil {
		geminiErr := domain.NewErrGeminiStagingFailed("Gemini Staging API call failed", err) // BE-041 - Custom error type
		geminiErr.SetLogger(s.logger)
		s.logger.Error("Gemini API call failed", zap.String("operation", operation), zap.String("request_id", requestID), zap.Error(geminiErr))
		return nil, geminiErr
	}

	// 3. Process Gemini API Output and Create Stage Model - BE-041, BE-044, BE-048a
	stage := &models.Stage{ // BE-044 - Create Stage model
		// ID:        uuid.New(),  // Removed: ID is not set here, database will generate it. - Fixed issue: #2
		T:          geminiOutput.T,          // Extract T staging from Gemini Output
		N:          geminiOutput.N,          // Extract N staging
		M:          geminiOutput.M,          // Extract M staging
		Confidence: geminiOutput.Confidence, // Extract confidence
		SessionID:  patientID,               // Corrected: SessionID is now correctly set. - Fixed issue: #1
	}

	// 4. (Optional) Integrate with Knowledge Base/Rules - BE-048a - Placeholder
	// In real implementation, knowledge base integration logic would be here to refine or validate staging
	// Example:
	//   refinedStage, err := s.knowledgeBase.RefineStaging(ctx, stage, patientData)
	//   if err != nil {
	//       s.logger.Warn("Knowledge base staging refinement failed, proceeding with preliminary staging", zap.String("operation", operation), zap.Error(err))
	//       // Decide whether to return error or proceed with unrefined staging
	//   } else {
	//       stage = refinedStage // Use refined staging from knowledge base
	//   }

	s.logger.Info("Successfully retrieved staging information", zap.String("operation", operation), zap.String("patient_id", patientID.String()), zap.String("request_id", requestID))
	return stage, nil
}

// SuggestTreatmentOptions retrieves potential treatment options using the Gemini API.
func (s *DiagnosisService) SuggestTreatmentOptions(ctx context.Context, patientID uuid.UUID) ([]*models.TreatmentRecommendation, error) {
	const operation = "DiagnosisService.SuggestTreatmentOptions" // Corrected operation name
	requestID := utils.GetRequestID(ctx.(*gin.Context))

	s.logger.Info("Starting treatment options suggestion", zap.String("operation", operation), zap.String("patient_id", patientID.String()), zap.String("request_id", requestID))

	// 1. Prepare Input for Gemini API - BE-043, BE-048a
	geminiInput := &geminiModels.TreatmentRecommendationInput{
		PatientID: patientID,
		Prompt:    "Suggest potential treatment options for lung cancer based on the preliminary diagnosis and staging.", // Define prompt - consider refine prompt
		// In real implementation, populate input with preliminary diagnosis and staging info
		// Example:
		//   diagnosis, err := s.GeneratePreliminaryDiagnosis(ctx, patientID)
		//   if err != nil {  return nil, fmt.Errorf("failed to get preliminary diagnosis: %w", err) }
		//   stagingInfo, err := s.GetStagingInformation(ctx, patientID)
		//   if err != nil {  return nil, fmt.Errorf("failed to get staging information: %w", err) }
		//   geminiInput.Diagnosis = diagnosis.DiagnosisText // Assuming DiagnosisText field in TreatmentRecommendationInput
		//   geminiInput.Stage = stagingInfo.TNMStage // Assuming TNMStage field in TreatmentRecommendationInput
		//   geminiInput.PatientPreferences = ... // Add patient preferences if available
	}

	// 2. Call Gemini API Client - BE-043, BE-048a
	geminiOutput, err := s.geminiClient.SuggestTreatmentOptions(ctx, geminiInput)
	if err != nil {
		geminiErr := domain.NewErrGeminiTreatmentSuggestFailed("Gemini Treatment Suggestion API call failed", err) // BE-043 - Custom error type
		geminiErr.SetLogger(s.logger)
		s.logger.Error("Gemini API call failed", zap.String("operation", operation), zap.String("request_id", requestID), zap.Error(geminiErr))
		return nil, geminiErr
	}

	// 3. Process Gemini API Output and Create TreatmentRecommendation Models - BE-043, BE-044, BE-048a
	treatmentRecommendations := make([]*models.TreatmentRecommendation, len(geminiOutput.Recommendations))
	for i, rec := range geminiOutput.Recommendations {
		treatmentRecommendations[i] = &models.TreatmentRecommendation{ // BE-044 - Create TreatmentRecommendation model
			// ID:              uuid.New(), // Removed: ID is not set here, database will generate it. - Fixed issue: #2
			TreatmentOption: rec.TreatmentOption, // Extract treatment option from Gemini output
			Rationale:       rec.Rationale,       // Extract rationale
			Benefits:        rec.Benefits,        // Extract benefits
			Risks:           rec.Risks,           // Extract risk
			SideEffects:     rec.SideEffects,     // Extract side effects
			Confidence:      rec.Confidence,      // Extract confidence
			SessionID:       patientID,           // Corrected: SessionID is now correctly set. - Fixed issue: #1
		}
	}

	// 4. (Optional) Integrate with Knowledge Base/Rules - BE-048a - Placeholder
	// In real implementation, knowledge base integration logic would be here to refine or validate treatment recommendations
	// Example:
	//   refinedRecommendations, err := s.knowledgeBase.RefineTreatmentOptions(ctx, treatmentRecommendations, patientData)
	//   if err != nil {
	//       s.logger.Warn("Knowledge base treatment refinement failed, proceeding with preliminary suggestions", zap.String("operation", operation), zap.Error(err))
	//       // Decide whether to return error or proceed with unrefined recommendations
	//   } else {
	//       treatmentRecommendations = refinedRecommendations // Use refined recommendations from knowledge base
	//   }

	s.logger.Info("Successfully retrieved treatment options suggestions", zap.String("operation", operation), zap.String("patient_id", patientID.String()), zap.String("request_id", requestID))
	return treatmentRecommendations, nil
}
