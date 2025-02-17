// internal/domain/services/link_service.go
package services

import (
	"context"
	"fmt"
	"time"

	"github.com/gin-gonic/gin" // Import gin to access Gin context for client IP - Recommendation 1
	"github.com/google/uuid"
	"github.com/stackvity/lung-server/internal/config"
	"github.com/stackvity/lung-server/internal/data/models"
	"github.com/stackvity/lung-server/internal/data/repositories/interfaces"
	"github.com/stackvity/lung-server/internal/domain" // Import domain for custom errors - Recommendation 4
	"github.com/stackvity/lung-server/internal/utils"
	"go.uber.org/zap"
)

// LinkService encapsulates the business logic for managing access links.
// It provides functionalities for generating, validating, and invalidating patient access links,
// ensuring secure and controlled access to the system.
type LinkService struct {
	patientRepository interfaces.PatientRepository // PatientRepository interface for data access
	config            *config.Config               // Configuration for accessing configuration parameters
	logger            *zap.Logger                  // Logger for structured logging
}

// NewLinkService creates a new LinkService instance, injecting the required dependencies.
// It takes a PatientRepository, Config, and Logger as arguments, promoting dependency injection
// and making the service more modular, testable, and configurable.
func NewLinkService(patientRepository interfaces.PatientRepository, config *config.Config, logger *zap.Logger) *LinkService {
	return &LinkService{
		patientRepository: patientRepository,
		config:            config,
		logger:            logger.Named("LinkService"), // Creates a logger specific to LinkService for contextual logging
	}
}

// GenerateLink creates a unique, time-limited, and expiring access link for a patient session.
// This function is central to the secure access mechanism, ensuring that each patient access is controlled
// and temporary. It generates a cryptographically secure link and persists it in the database
// along with an expiration timestamp and other relevant session data.
// It takes a context for cancellation and timeout, and returns the generated access link string or an error if generation fails.
func (s *LinkService) GenerateLink(ctx context.Context) (string, error) {
	const operation = "GenerateLink" // Define operation name for structured logging

	s.logger.Info("Generating new access link", zap.String("operation", operation)) // Log operation start

	// 1. Generate Unique Access Link (BE-002, US-023)
	accessLink, err := utils.GenerateURLSafeToken() // Use utils package to generate cryptographically secure token
	if err != nil {
		s.logger.Error("Failed to generate URL-safe token", zap.String("operation", operation), zap.Error(err)) // Log error if token generation fails
		return "", fmt.Errorf("generating access link: %w", err)                                                // Return error with context
	}
	s.logger.Debug("Generated unique access link", zap.String("operation", operation), zap.String("access_link", accessLink)) // Debug log showing generated link

	// 2. Set Expiration Timestamp (BE-024, US-024)
	expirationTime := time.Now().UTC().Add(s.config.LinkExpiration) // Use configuration for link expiration duration - BE-024
	s.logger.Debug("Link expiration time set", zap.String("operation", operation), zap.Time("expiration_time", expirationTime))

	// 3. Create Patient Session Record (BE-002, US-023, BE-004)
	patient := &models.Patient{ // Corrected: Create models.Patient struct
		SessionID:           uuid.New(),
		AccessLink:          accessLink,
		ExpirationTimestamp: expirationTime,
		Used:                false, // Link is initially unused - US-001
	}

	// 4. Save to Database (BE-004)
	err = s.patientRepository.CreatePatient(ctx, patient) // Corrected: Pass models.Patient to CreatePatient
	if err != nil {
		s.logger.Error("Failed to save patient session to database", zap.String("operation", operation), zap.Error(err)) // Log error if database save fails
		return "", fmt.Errorf("saving access link to database: %w", err)                                                 // Return error with context
	}

	s.logger.Info("Access link generated and saved", zap.String("operation", operation), zap.String("access_link", accessLink), zap.Time("expiration_time", expirationTime)) // Log success and link details
	return accessLink, nil                                                                                                                                                   // Return the generated access link and nil error for success
}

// ValidateLink checks if a given access link is valid.
// Validation involves verifying the link's existence in the database, ensuring it has not expired,
// and confirming it has not been used previously, thus enforcing the time-limited and single-use nature of access links.
// It takes a context and the access link string as input and returns a boolean indicating validity and an error if validation fails.
func (s *LinkService) ValidateLink(ctx context.Context, accessLink string, ginContext *gin.Context) (bool, error) { // Modified to accept ginContext for client IP - Recommendation 1
	const operation = "ValidateLink" // Operation name for structured logging
	requestID := utils.GetRequestID(ctx)

	s.logger.Debug("Validating access link", zap.String("operation", operation), zap.String("access_link", accessLink), zap.String("request_id", requestID)) // Debug log for link validation start

	// 1. Retrieve Patient Session by Link (BE-003, US-001)
	patientSession, err := s.patientRepository.GetPatientSessionByLink(ctx, accessLink) // Use PatientRepository to fetch session by link
	if err != nil {
		logger := s.logger.With(zap.String("operation", operation), zap.String("access_link", accessLink), zap.Error(err), zap.String("request_id", requestID)) // Create logger with context
		if ginContext != nil {
			logger = logger.With(zap.String("client_ip", ginContext.ClientIP())) // Conditionally add client IP to log context - Recommendation 1
		}
		logger.Warn("Access link validation failed: link not found", zap.String("request_id", requestID), zap.String("client_ip", ginContext.ClientIP())) // Enhanced logging - Recommendation 1
		return false, domain.NewErrInvalidLink("Access link not found or invalid")                                                                        // Return custom error - Recommendation 4
	}

	// 2. Check if Link Expired (BE-024, US-024)
	if time.Now().UTC().After(patientSession.ExpirationTimestamp) { // Check if current time is after the link's expiration time
		logger := s.logger.With(zap.String("operation", operation), zap.String("access_link", accessLink), zap.Time("expiration_timestamp", patientSession.ExpirationTimestamp), zap.String("request_id", requestID), zap.String("client_ip", ginContext.ClientIP())) // Logger with context including client IP and request ID - Recommendation 1
		if ginContext != nil {
			logger = logger.With(zap.String("client_ip", ginContext.ClientIP())) // Conditionally add client IP to log context - Recommendation 1
		}
		logger.Warn("Expired access link", zap.String("request_id", requestID), zap.String("client_ip", ginContext.ClientIP())) // Enhanced logging - Recommendation 1
		return false, domain.NewErrLinkExpired("Access link has expired")                                                       // Return custom error - Recommendation 4
	}

	// 3. Check if Link Already Used (BE-003, US-001)
	if patientSession.Used { // Check the 'used' flag in the patient session record
		logger := s.logger.With(zap.String("operation", operation), zap.String("access_link", accessLink), zap.String("request_id", requestID), zap.String("client_ip", ginContext.ClientIP())) // Logger with context including client IP and request ID - Recommendation 1
		if ginContext != nil {
			logger = logger.With(zap.String("client_ip", ginContext.ClientIP())) // Conditionally add client IP to log context - Recommendation 1
		}
		logger.Warn("Used access link", zap.String("request_id", requestID), zap.String("client_ip", ginContext.ClientIP())) // Enhanced logging - Recommendation 1
		return false, domain.NewErrLinkExpired("Access link already used")                                                   // Return custom error - Recommendation 4
	}

	s.logger.Debug("Access link validated successfully", zap.String("operation", operation), zap.String("access_link", accessLink), zap.String("request_id", requestID)) // Debug log for successful validation
	return true, nil                                                                                                                                                     // Return true (valid link) and nil error for successful validation
}

// InvalidateLink marks a given access link as used, preventing further access using the same link.
// This function is typically called after a patient has successfully accessed the system and uploaded their documents,
// enforcing the single-use nature of the access links as a security measure.
// It takes a context and the access link string as input and returns an error if the invalidation process fails.
func (s *LinkService) InvalidateLink(ctx context.Context, accessLink string) error {
	const operation = "InvalidateLink" // Operation name for structured logging

	s.logger.Info("Invalidating access link", zap.String("operation", operation), zap.String("access_link", accessLink)) // Info log for link invalidation

	// 1. Invalidate Access Link in Database (BE-024, US-024)
	err := s.patientRepository.InvalidateLink(ctx, accessLink) // Use PatientRepository to invalidate the link in the database
	if err != nil {
		s.logger.Error("Failed to invalidate access link in database", zap.String("operation", operation), zap.String("access_link", accessLink), zap.Error(err)) // Log error if database update fails
		return fmt.Errorf("invalidating access link: %w", err)                                                                                                    // Return error with context
	}

	s.logger.Info("Access link invalidated successfully", zap.String("operation", operation), zap.String("access_link", accessLink)) // Info log for successful invalidation
	return nil                                                                                                                       // Return nil error for successful invalidation
}
