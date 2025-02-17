// internal/domain/services/processing_service.go
package services

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/stackvity/lung-server/internal/data/models"
	"github.com/stackvity/lung-server/internal/data/repositories/interfaces"
	"github.com/stackvity/lung-server/internal/domain" // Corrected import: Use "internal/domain" not "internal/domain/entities"
	"github.com/stackvity/lung-server/internal/gemini"
	geminiModels "github.com/stackvity/lung-server/internal/gemini/models"
	"github.com/stackvity/lung-server/internal/knowledge"
	"github.com/stackvity/lung-server/internal/ocr"
	"github.com/stackvity/lung-server/internal/security"
	"github.com/stackvity/lung-server/internal/storage"
	"github.com/stackvity/lung-server/internal/utils"
	"github.com/stackvity/lung-server/pkg/dicom"
	"go.uber.org/zap"
)

type ProcessingService struct {
	fileStorage       storage.FileStorage
	ocrService        ocr.OCRService
	geminiClient      gemini.GeminiClient
	validator         *validator.Validate
	patientRepository interfaces.PatientRepository
	studyRepository   interfaces.StudyRepository
	imageRepository   interfaces.ImageRepository
	reportRepository  interfaces.ReportRepository
	knowledgeBase     knowledge.KnowledgeBase
	logger            *zap.Logger
}

// NewProcessingService creates a new ProcessingService with dependencies injected.
func NewProcessingService(
	fileStorage storage.FileStorage,
	ocrService ocr.OCRService,
	geminiClient gemini.GeminiClient,
	validator *validator.Validate,
	patientRepository interfaces.PatientRepository,
	studyRepository interfaces.StudyRepository,
	imageRepository interfaces.ImageRepository,
	reportRepository interfaces.ReportRepository,
	knowledgeBase knowledge.KnowledgeBase,
	logger *zap.Logger,
) *ProcessingService {
	return &ProcessingService{
		fileStorage:       fileStorage,
		ocrService:        ocrService,
		geminiClient:      geminiClient,
		validator:         validator,
		patientRepository: patientRepository,
		studyRepository:   studyRepository,
		imageRepository:   imageRepository,
		reportRepository:  reportRepository,
		knowledgeBase:     knowledgeBase,
		logger:            logger.Named("processing"),
	}
}

// ProcessDocument is the main entry point for document processing.
func (s *ProcessingService) ProcessDocument(ctx context.Context, patientID uuid.UUID, filename string, contentType string, file io.Reader) error {
	const operation = "ProcessDocument"
	requestID := utils.GetRequestID(ctx)

	s.logger.Info("Processing document",
		zap.String("operation", operation),
		zap.String("request_id", requestID),
		zap.String("patient_id", patientID.String()),
		zap.String("filename", filename),
		zap.String("content_type", contentType),
	)

	if err := utils.ValidateUUID(patientID.String()); err != nil {
		s.logger.Warn("Invalid patient ID", zap.String("operation", operation), zap.String("request_id", requestID), zap.String("patient_id", patientID.String()), zap.Error(err))
		return fmt.Errorf("invalid patient ID: %w", err)
	}

	// 1. Input Validation (File Type, Size, etc.) - BE-010, BE-011, BE-012, BE-055 (Enhanced Validation)
	if err := s.validateFile(filename, contentType); err != nil { // Enhanced validation in validateFile
		s.logger.Warn("File validation failed", zap.String("operation", operation), zap.String("request_id", requestID), zap.String("filename", filename), zap.String("content_type", contentType), zap.Error(err))
		return fmt.Errorf("validating file: %w", err)
	}

	// 2. Save File (Temporarily) - BE-014
	filePath, err := s.fileStorage.Save(ctx, filename, contentType, file)
	if err != nil {
		s.logger.Error("Failed to save file", zap.String("operation", operation), zap.String("request_id", requestID), zap.String("filename", filename), zap.Error(err))
		return fmt.Errorf("saving file: %w", err)
	}
	s.logger.Info("File saved temporarily", zap.String("operation", operation), zap.String("request_id", requestID), zap.String("file_path", filePath))

	defer func() { // BE-024, BE-025 (Secure Deletion)
		if err := s.fileStorage.Delete(ctx, filePath); err != nil {
			s.logger.Error("Failed to delete temporary file", zap.String("operation", operation), zap.String("request_id", requestID), zap.String("file_path", filePath), zap.Error(err))
		} else {
			s.logger.Info("Temporary file deleted", zap.String("operation", operation), zap.String("request_id", requestID), zap.String("file_path", filePath))
		}
	}()

	// 3.  File Type Handling
	switch contentType {
	case "application/dicom":
		return s.processDICOMFile(ctx, patientID, filename, filePath) // BE-029, BE-030
	case "application/pdf", "image/jpeg", "image/png":
		return s.processPDFImageFile(ctx, patientID, filename, filePath, contentType)
	default:
		s.logger.Error("Unsupported content type", zap.String("operation", operation), zap.String("request_id", requestID), zap.String("filename", filename), zap.String("content_type", contentType))
		return fmt.Errorf("unsupported content type: %s", contentType)
	}
}

// GetPatientRepository provides controlled access to the PatientRepository dependency.  // ADDED Getter Method
func (s *ProcessingService) GetPatientRepository() interfaces.PatientRepository { // ADDED Getter Method
	return s.patientRepository
}

func (s *ProcessingService) validateFile(filename string, contentType string) error {
	// 1. Basic Content Type Validation (using net/http):
	detectedContentType, err := utils.DetectContentTypeFromFile(filename)
	if err != nil {
		return utils.Wrapf(err, "detecting content type for file: %s", filename) // Enhanced error wrapping with Wrapf
	}
	if !strings.HasPrefix(detectedContentType, contentType) {
		return fmt.Errorf("mismatched content type: expected %s, detected %s for file: %s", contentType, detectedContentType, filename)
	}

	// 2.  DICOM-Specific Validation (Magic Number Check) - BE-011 (Magic Number Implementation)
	if contentType == "application/dicom" {
		if err := s.validateDICOMMagicNumber(filename); err != nil { // Added DICOM magic number validation
			return utils.Wrapf(err, "DICOM magic number validation failed for file: %s", filename) // Enhanced error wrapping
		}

		// Basic DICOM Header Validation (as before)
		_, err := dicom.ParseFile(filename, 1024*1024*2)
		if err != nil {
			return utils.Wrapf(err, "basic DICOM header validation failed for file: %s", filename) // Enhanced error wrapping
		}
	}

	// 3.  Further checks can be added here (e.g., using go-playground/validator for file metadata).

	return nil
}

func (s *ProcessingService) validateDICOMMagicNumber(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("opening file: %w", err)
	}
	defer file.Close()

	header := make([]byte, 132) // DICOM preamble + magic number is 132 bytes
	_, err = file.Read(header)
	if err != nil {
		return fmt.Errorf("reading file header: %w", err)
	}

	// DICOM magic number is bytes 128-131 (0-indexed) and should be "DICM"
	magicNumber := string(header[128:132])
	if magicNumber != "DICM" {
		return errors.New("DICOM magic number not found")
	}
	return nil
}

func (s *ProcessingService) determineReportType(filename string, text string) string {
	lowerFilename := strings.ToLower(filename)
	lowerText := strings.ToLower(text)

	// Improved keyword-based classification (BE-021 - Expanded Keywords)
	if strings.Contains(lowerFilename, "pathology") || strings.Contains(lowerText, "pathology") || strings.Contains(lowerText, "biopsy") || strings.Contains(lowerText, "surgical specimen") || strings.Contains(lowerText, "histology") { // Expanded pathology keywords
		return "pathology"
	}
	if strings.Contains(lowerFilename, "radiology") || strings.Contains(lowerText, "radiology") ||
		strings.Contains(lowerFilename, "ct") || strings.Contains(lowerFilename, "ct scan") || strings.Contains(lowerText, "x-ray") || strings.Contains(lowerText, "xray") || // Radiology keywords
		strings.Contains(lowerText, "radiograph") || strings.Contains(lowerText, "imaging") {
		return "radiology"
	}
	if strings.Contains(lowerFilename, "lab") || strings.Contains(lowerFilename, "laboratory") ||
		strings.Contains(lowerFilename, "test") || strings.Contains(lowerText, "lab ") || strings.Contains(lowerText, "test ") || // Lab test keywords (with spaces to avoid partial matches)
		strings.Contains(lowerText, "blood") || strings.Contains(lowerText, "cbc") || strings.Contains(lowerText, "chemistry panel") || strings.Contains(lowerText, "urinalysis") { // More lab-related terms
		return "labtest"
	}

	s.logger.Warn("Unable to determine report type, defaulting to radiology", zap.String("filename", filename))
	return "radiology" // Default - should log a warning/error in real implementation
}

func (s *ProcessingService) preprocessImage(_ *dicom.DataSet) ([]byte, error) {
	// Placeholder implementation for image preprocessing (BE-029 - Placeholder with Comments)
	// In a real implementation, this function would perform steps like:

	// 1. Image Resizing:
	//    - Resize the image to a standard input size (e.g., 256x256, 512x512)
	//    - Use Lanczos resampling for high-quality resizing
	//    - Example using gocv (or similar library):
	//      resizedImage := gocv.NewMat()
	//      gocv.Resize(inputImage, resizedImage, image.Size{X: 256, Y: 256}, 0, 0, gocv.InterpolationLanczos4)
	//      defer resizedImage.Close()

	// 2. Intensity Normalization:
	//    - Normalize pixel intensities to a standard range (e.g., 0-1 or -1 to 1)
	//    - Common methods: Min-Max scaling, Z-score normalization
	//    - Example (Min-Max scaling - conceptual):
	//      minVal, maxVal := gocv.MinMaxLoc(resizedImage)
	//      normalizedImage := gocv.NewMat()
	//      gocv.SubtractScalar(resizedImage, scalar.NewScalar(minVal, minVal, minVal, 0), normalizedImage)
	//      gocv.DivideScalar(normalizedImage, scalar.NewScalar(maxVal-minVal, maxVal-minVal, maxVal-minVal, 0), normalizedImage)
	//      defer normalizedImage.Close()

	// 3. (Optional) Lung Segmentation:
	//    - Use a pre-trained lung segmentation model (if available and performant)
	//    - Apply segmentation mask to focus analysis on lung regions
	//    - Example (conceptual):
	//      mask := loadLungSegmentationMask(resizedImage) // Load or generate mask
	//      maskedImage := gocv.NewMat()
	//      gocv.BitwiseAnd(resizedImage, resizedImage, maskedImage, mask, mask) // Apply mask
	//      defer maskedImage.Close()

	// 4. Encoding and Output:
	//    - Encode the preprocessed image to JPEG or PNG format for Gemini API.
	//    - Return the encoded image data as bytes.
	//    - Example (JPEG encoding using gocv - conceptual):
	//      encodedImageData, _ := gocv.IMEncode(".jpg", normalizedImage)
	//      return encodedImageData, nil

	s.logger.Warn("preprocessImage is a placeholder - basic image preprocessing steps (resizing, normalization) not yet implemented") // Updated log message
	return []byte{}, nil                                                                                                              // Placeholder!
}

func (s *ProcessingService) processDICOMFile(ctx context.Context, patientID uuid.UUID, filename, filePath string) error {
	const operation = "processDICOMFile"

	// 1. Parse DICOM file - BE-029
	dicomData, err := dicom.ParseFile(filePath, 1024*1024*2)
	if err != nil {
		utils.Logger.Warn("Failed to parse DICOM", zap.String("filename", filename), zap.Error(err), zap.String("operation", operation))
		parseErr := domain.NewErrDICOMParsingFailed(filename, err) // BE-021 (OCR Error Handling) - Custom Error
		parseErr.SetLogger(s.logger)                               //Set logger for domain error
		return parseErr
	}
	utils.Logger.Info("DICOM Parsed")

	// 2. Extract relevant metadata (Study, Series, Instance) - BE-029
	studyInstanceUID := dicomData.StudyInstanceUID
	seriesInstanceUID := dicomData.SeriesInstanceUID
	sopInstanceUID := dicomData.SOPInstanceUID

	// 3. Data Anonymization/De-identification (before storing or further processing) - BE-055
	anonymizedDicomData, err := security.AnonymizeDICOMData(dicomData) // Implement this function
	if err != nil {
		return fmt.Errorf("anonymizing DICOM data: %w", err) // BE-055 - Data Anonymization
	}

	// 4. Create or Update Database Records (Patient, Study, Image) - BE-024
	// (Use repositories to interact with the database)

	// Check if patient already exists, create if it does not
	_, err = s.patientRepository.GetPatient(ctx, patientID) // Assumes pseudonym is used as the ID
	if err != nil {
		if _, ok := err.(*domain.NotFoundError); ok { // Use domain.NotFoundError
			newPatient := &models.Patient{
				SessionID: patientID, // Using the pseudonym
			}
			if err := s.patientRepository.CreatePatient(ctx, newPatient); err != nil {
				return fmt.Errorf("create patient in db: %w", err) // BE-024 - DB Interaction
			}
		} else {
			return fmt.Errorf("getting patient: %w", err) // BE-024 - DB Interaction
		}
	}

	study := &models.Study{ // BE-024
		ID:               uuid.New(),       // Or generate based on StudyInstanceUID
		PatientID:        patientID,        // Foreign key
		StudyInstanceUID: studyInstanceUID, // Store anonymized UID
		// ... other relevant fields ...
	}
	if err := s.studyRepository.CreateStudy(ctx, study); err != nil {
		return fmt.Errorf("creating study in db: %w", err) // BE-024 - DB Interaction
	}

	image := &models.Image{ // BE-024
		ID:                uuid.New(), // Or generate based on SOPInstanceUID
		StudyID:           study.ID,   // Foreign Key
		FilePath:          filePath,
		SeriesInstanceUID: seriesInstanceUID, // Store anonymized UID
		SOPInstanceUID:    sopInstanceUID,    // Store anonymized UID
		ImageType:         "dicom",           // Correctly set image type
	}
	if err := s.imageRepository.CreateImage(ctx, image); err != nil {
		return fmt.Errorf("creating image in db: %w", err) // BE-024 - DB Interaction
	}
	// 5. Image Preprocessing (Resizing, Normalization, etc.) - if needed - BE-029
	preprocessedImageData, err := s.preprocessImage(anonymizedDicomData)
	if err != nil {
		return fmt.Errorf("preprocessing image: %w", err) // BE-029 - Image Preprocessing
	}

	// 6.  Call Gemini API for Nodule Detection - BE-030
	geminiInput := &geminiModels.NoduleDetectionInput{
		ImageData: preprocessedImageData,
		ImageType: "dicom",                                                                                             // Or determine from DICOM metadata
		Prompt:    "Identify potential lung nodules in this DICOM image.  Report location, size, and characteristics.", // Use a managed prompt - BE-028
	}
	geminiOutput, err := s.geminiClient.DetectNodules(ctx, geminiInput) // BE-030 - Gemini API Call
	if err != nil {
		geminiErr := domain.NewErrGeminiNoduleDetectionFailed("Gemini Nodule Detection API call failed", err) // BE-030 - Gemini API Error Handling - Custom Error
		geminiErr.SetLogger(s.logger)                                                                         // Set logger for domain error
		return geminiErr

	}

	// 7. Process Gemini Output (Store Nodule Information)
	for _, noduleInfo := range geminiOutput.Nodules {
		nodule := &models.Nodule{ // BE-030 - Process Gemini Output
			ID:       uuid.New(),
			ImageID:  image.ID, // Link to the Image
			Location: noduleInfo.Location,
			Size:     noduleInfo.Size,
			Shape:    noduleInfo.Shape,
			// ... other characteristics ...
		}
		if err := s.imageRepository.CreateNodule(ctx, nodule); err != nil {
			return fmt.Errorf("saving nodule: %w", err) // BE-048a - Store Nodule Information
		}
	}

	// 8.  (Optional) If other analyses are needed (e.g., staging from image), call other Gemini methods.

	return nil
}

func (s *ProcessingService) processPDFImageFile(ctx context.Context, patientID uuid.UUID, filename, filePath string, contentType string) error {
	const operation = "processPDFImageFile"

	// 1. Read File Data - BE-010
	fileData, err := os.ReadFile(filePath) // Read the file from temporary storage
	if err != nil {
		return fmt.Errorf("reading file data: %w", err) // BE-010 - File Handling
	}
	// 2. OCR (if it's a scanned document or PDF) - BE-021
	extractedText, _, err := s.ocrService.ExtractText(ctx, fileData, contentType) // contentType helps determine image type - BE-021 // Corrected line
	if err != nil {
		ocrErr := domain.NewErrOCRExtractionFailed(filename, err) // BE-021 - OCR Error Handling - Custom Error
		ocrErr.SetLogger(s.logger)                                // Set logger for domain error
		return ocrErr

	}

	// 3. Data Anonymization (of extracted text) - BE-055
	anonymizedText := security.AnonymizeText(extractedText) // Implement this function (mask PII) - BE-055

	// 4. Create or Update Database Records - BE-024
	_, err = s.patientRepository.GetPatient(ctx, patientID)
	if err != nil {
		if _, ok := err.(*domain.NotFoundError); ok {
			newPatient := &models.Patient{
				SessionID: patientID, // Using the pseudonym - BE-024
			}
			if err := s.patientRepository.CreatePatient(ctx, newPatient); err != nil {
				return fmt.Errorf("create patient when processing pdf: %w", err) // BE-024 - DB Interaction
			}
		} else {
			return fmt.Errorf("getting patient: %w", err) // BE-024 - DB Interaction
		}
	}

	report := &models.Report{ // BE-024
		ID:         uuid.New(),
		PatientID:  patientID,
		Filename:   filename,
		ReportType: s.determineReportType(filename, extractedText), // Implement a function to guess report type - BE-021
		ReportText: anonymizedText,                                 // Store anonymized text *temporarily* - BE-055
		Filepath:   filePath,                                       //Store file path to connect to the file. - BE-014
	}
	if err := s.reportRepository.CreateReport(ctx, report); err != nil {
		return fmt.Errorf("creating report in db: %w", err) // BE-024 - DB Interaction
	}

	// 5. Call Gemini API for Analysis (Pathology, Information Extraction, etc.) - BE-030, BE-048a
	if report.ReportType == "pathology" {
		geminiInput := &geminiModels.PathologyReportAnalysisInput{
			ReportText: anonymizedText,
			Prompt:     "Extract key findings from this pathology report...", // Use a managed prompt - BE-028
		}
		geminiOutput, err := s.geminiClient.AnalyzePathologyReport(ctx, geminiInput) // BE-030 - Gemini API Call
		if err != nil {
			return fmt.Errorf("calling Gemini API for pathology report analysis: %w", err) // BE-030 - Gemini API Error Handling
		}
		// Process Gemini Output (store key findings) - BE-048a
		for _, finding := range geminiOutput.Findings {
			dbFinding := &models.Finding{
				FindingID:   uuid.New(),
				FileID:      report.ID, // Link to the Report
				FindingType: "pathology",
				Description: finding.Description,
				// ... other details ...
			}
			if err := s.reportRepository.CreateFinding(ctx, dbFinding); err != nil {
				return fmt.Errorf("creating pathology finding: %w", err) // BE-048a - Store Findings
			}
		}

	} else { // Assume radiology report (or other) - BE-030, BE-048a
		geminiInput := &geminiModels.InformationExtractionInput{ // You'll need to define this model
			ReportText: anonymizedText,
			Prompt:     "Extract symptoms, signs, and investigations mentioned in this report...", // Use managed prompt - BE-028
		}
		geminiOutput, err := s.geminiClient.ExtractInformation(ctx, geminiInput) // and this method - BE-030 - Gemini API Call
		if err != nil {
			return fmt.Errorf("calling Gemini API for information extraction: %w", err) // BE-030 - Gemini API Error Handling
		}

		// Process Gemini Output (store findings) - BE-048a
		for _, finding := range geminiOutput.Findings { // Assuming a Findings field in the output
			dbFinding := &models.Finding{
				FindingID:   uuid.New(),
				FileID:      report.ID, // Link to Report
				FindingType: finding.Type,
				Description: finding.Description,
				//... other details...
			}
			if err := s.reportRepository.CreateFinding(ctx, dbFinding); err != nil {
				return fmt.Errorf("creating extracted finding %w", err) // BE-048a - Store Extracted Info
			}
		}
	}

	return nil
}

func (s *ProcessingService) DeleteAllPatientData(ctx context.Context, patientID uuid.UUID) error {
	//  1. Get all file paths associated with the patient (from the database).
	//  2. Delete files from storage (using s.fileStorage.Delete).
	//  3. Delete database records (Patient, Study, Image, Nodule, Report, AuditLog entries).
	//    * Use transactions to ensure all deletions are successful or rolled back.
	//   * Use the repository interfaces for all database interactions.
	tx, err := s.reportRepository.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)
	images, err := s.imageRepository.GetImageByPatientID(ctx, patientID)
	if err != nil {
		return fmt.Errorf("getting image filepaths: %w", err)
	}

	for _, image := range images {
		if err := s.fileStorage.Delete(ctx, image.FilePath); err != nil {
			return fmt.Errorf("deleting image files: %w", err)
		}
	}

	reports, err := s.reportRepository.GetReportByPatientID(ctx, patientID)
	if err != nil {
		return fmt.Errorf("getting report filepaths: %w", err)
	}

	for _, report := range reports {
		if err := s.fileStorage.Delete(ctx, report.Filepath); err != nil {
			return fmt.Errorf("delete report file: %w", err)
		}
	}

	// Delete all data from the database
	if err := s.reportRepository.DeleteAllReportsByPatientID(ctx, patientID); err != nil {
		return fmt.Errorf("deleting reports from db: %w", err)
	}
	if err := s.imageRepository.DeleteAllImagesByPatientID(ctx, patientID); err != nil {
		return fmt.Errorf("deleting images from db: %w", err)
	}

	if err := s.studyRepository.DeleteAllStudiesByPatientID(ctx, patientID); err != nil {
		return fmt.Errorf("deleting studies from db: %w", err)
	}

	if err := s.patientRepository.DeletePatient(ctx, patientID); err != nil {
		return fmt.Errorf("deleting patient from db: %w", err)
	}
	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("committing transaction: %w", err)
	}
	return nil
}
