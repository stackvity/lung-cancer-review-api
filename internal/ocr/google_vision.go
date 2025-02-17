// internal/ocr/google_vision.go
package ocr

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	vision "cloud.google.com/go/vision/apiv1"
	"github.com/stackvity/lung-server/internal/config"
	"github.com/stackvity/lung-server/internal/domain"
	"github.com/stackvity/lung-server/internal/utils"
	"go.uber.org/zap"
	"google.golang.org/api/option"
	visionpb "google.golang.org/genproto/googleapis/cloud/vision/v1"
)

// GoogleVisionService implements the OCRService interface using the Google Cloud Vision API.
type GoogleVisionService struct {
	config       *config.Config
	logger       *zap.Logger
	visionClient *vision.ImageAnnotatorClient
}

// NewGoogleVisionService creates a new GoogleVisionService instance.
func NewGoogleVisionService(ctx context.Context, cfg *config.Config, logger *zap.Logger) (*GoogleVisionService, error) {
	const operation = "NewGoogleVisionService"
	gvsLogger := logger.Named("GoogleVisionService")

	gvsLogger.Info("Initializing Google Vision Service", zap.String("operation", operation))

	visionClient, err := vision.NewImageAnnotatorClient(ctx, option.WithAPIKey(cfg.GeminiAPIKey))
	if err != nil {
		gvsLogger.Error("Failed to create Google Vision client", zap.Error(err), zap.String("operation", operation))
		return nil, fmt.Errorf("failed to create Google Vision client: %w", err)
	}

	gvs := &GoogleVisionService{
		config:       cfg,
		logger:       gvsLogger,
		visionClient: visionClient,
	}

	gvsLogger.Info("Google Vision Service initialized successfully", zap.String("operation", operation))
	return gvs, nil
}

// ExtractText implements the OCRService interface.
func (s *GoogleVisionService) ExtractText(ctx context.Context, imageData []byte, imageType string) (string, float64, error) {
	const operation = "GoogleVisionService.ExtractText"
	requestID := utils.GetRequestID(ctx)

	s.logger.Info("Starting OCR extraction with Google Vision API", zap.String("operation", operation), zap.String("request_id", requestID), zap.String("image_type", imageType))

	client := s.visionClient

	image := &visionpb.Image{ // Corrected: Use visionpb.Image from the visionpb package
		Content: imageData, // Corrected: Assign imageData to Content field
	}
	request := &visionpb.AnnotateImageRequest{ // Corrected: Use visionpb.AnnotateImageRequest
		Image: image,
		Features: []*visionpb.Feature{{ // Corrected: Use visionpb.Feature
			Type:       visionpb.Feature_DOCUMENT_TEXT_DETECTION, // Corrected: Use visionpb.FeatureTypeDocumentTextDetection
			MaxResults: 1,
		}},
	}
	batchRequest := &visionpb.BatchAnnotateImagesRequest{ // Corrected: Use visionpb.BatchAnnotateImagesRequest
		Requests: []*visionpb.AnnotateImageRequest{request},
	}

	payload, err := json.Marshal(batchRequest)
	if err != nil {
		s.logger.Error("Failed to marshal request payload for logging", zap.String("operation", operation), zap.String("request_id", requestID), zap.Error(err))
	}

	apiStartTime := time.Now()
	resp, err := client.BatchAnnotateImages(ctx, batchRequest)
	apiLatency := time.Since(apiStartTime)
	responseBytes, respErr := json.Marshal(resp)
	if respErr != nil {
		s.logger.Warn("Failed to marshal response for logging", zap.String("operation", operation), zap.String("request_id", requestID), zap.Error(respErr))
		responseBytes = []byte("response marshaling failed")
	}

	s.logger.Debug("Google Vision API Request", zap.String("operation", operation), zap.String("request_id", requestID), zap.String("image_type", imageType), zap.ByteString("request_payload", payload))
	s.logger.Debug("Google Vision API Response", zap.String("operation", operation), zap.String("request_id", requestID), zap.Int("status_code", int(resp.Responses[0].Error.GetCode())), zap.ByteString("response_payload", responseBytes), zap.Duration("api_latency", apiLatency))

	if err != nil {
		apiErr := fmt.Errorf("Google Vision API call failed: %w", err)
		s.logger.Error("Google Vision API error", zap.String("operation", operation), zap.String("request_id", requestID), zap.Error(apiErr))
		return "", 0.0, fmt.Errorf("%s: %w", operation, domain.NewErrOCRExtractionFailed("Google Vision API call failed", apiErr))
	}
	if apiError := resp.Responses[0].Error; apiError != nil {
		apiErr := fmt.Errorf("Google Vision API returned error: %s (code: %d)", apiError.GetMessage(), apiError.GetCode())
		s.logger.Warn("Google Vision API returned error", zap.String("operation", operation), zap.String("request_id", requestID), zap.Int("status_code", int(apiError.GetCode())), zap.String("error_message", apiError.GetMessage()), zap.Error(apiErr))
		return "", 0.0, fmt.Errorf("%s: %w", operation, domain.NewErrOCRExtractionFailed("Google Vision API returned an error", apiErr))
	}

	var extractedText string
	var confidence float64 = 0.0

	if resp != nil && len(resp.Responses) > 0 && resp.Responses[0] != nil && resp.Responses[0].FullTextAnnotation != nil {
		extractedText = resp.Responses[0].FullTextAnnotation.GetText()
		totalConfidence := 0.0
		wordCount := 0
		for _, page := range resp.Responses[0].FullTextAnnotation.Pages {
			for _, block := range page.Blocks {
				for _, paragraph := range block.Paragraphs {
					for _, word := range paragraph.Words {
						for _, symbol := range word.Symbols {
							totalConfidence += float64(symbol.GetConfidence())
							wordCount++
						}
					}
				}
			}
		}
		if wordCount > 0 {
			confidence = totalConfidence / float64(wordCount)
		}
	}

	s.logger.Debug("Google Vision API Response Parsed", zap.String("operation", operation), zap.String("request_id", requestID), zap.Int("response_pages", len(resp.Responses[0].FullTextAnnotation.Pages)), zap.Float64("confidence_score", confidence), zap.Int("text_length", len(extractedText)))

	s.logger.Info("Successfully extracted text with Google Vision API", zap.String("operation", operation), zap.String("request_id", requestID), zap.String("image_type", imageType), zap.Float64("confidence_score", confidence), zap.Int("text_length", len(extractedText)))

	return extractedText, confidence, nil
}
