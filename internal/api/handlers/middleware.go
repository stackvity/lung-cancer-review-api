// internal/api/handlers/middleware.go
package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stackvity/lung-server/internal/config"
	"github.com/stackvity/lung-server/internal/data/repositories/interfaces"
	"github.com/stackvity/lung-server/internal/domain"
	"github.com/stackvity/lung-server/internal/utils"
	"go.uber.org/zap"
)

// MiddlewareConfig holds the necessary dependencies for the middleware.
// It encapsulates dependencies like PatientRepo, Logger, Config, and Validator,
// making it easier to inject these dependencies into the middleware functions and for testing.
type MiddlewareConfig struct {
	PatientRepo interfaces.PatientRepository
	Logger      *zap.Logger
	Config      *config.Config // Include Config for potential future use in middleware, e.g., for content type validation, rate limiting configs
}

// MiddlewareSetup initializes and returns the complete middleware chain for the application.
// It takes a MiddlewareConfig struct to inject dependencies and configuration.
// The middleware chain is executed in the order they are registered here, which is crucial for request processing flow.
// 1. Request Logging (executed first for logging as early as possible).
// 2. Link Validation (executed second for security access control before further processing).
func MiddlewareSetup(cfg MiddlewareConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Request Logging Middleware (Executed first): Logs incoming requests for monitoring, debugging, and audit trails.
		RequestLoggerMiddleware(cfg.Logger)(c)

		// 2. Link Validation Middleware (Executed second, after logging): Validates the access link from the request header for secure access control.
		LinkValidationMiddleware(cfg.PatientRepo, cfg.Logger)(c)

		c.Next() // Process the request - continue to the next middleware or handler in the chain
	}
}

// RequestLoggerMiddleware logs incoming HTTP requests with request IDs and tracing context.
// It generates a unique request ID for each request, adds it to the context for tracing, and logs comprehensive request details
// including timestamp, operation, request ID, method, path, IP address, HTTP status code, latency, and user agent.
func RequestLoggerMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		const operation = "RequestLoggerMiddleware"
		requestID := uuid.New().String() // Generate unique request ID at the beginning of each request
		ctx := context.WithValue(c.Request.Context(), utils.RequestIDKey, requestID)
		c.Request = c.Request.WithContext(ctx)

		startTime := time.Now()

		c.Next() // Call next handler in the chain - important to move to the next middleware/handler

		latency := time.Since(startTime)

		// Logging with structured fields, including the request ID for traceability and potential trace IDs for distributed tracing correlation.
		logger.Info("Incoming request processed", // More descriptive and action-oriented log message
			zap.Time("timestamp", startTime), // Explicitly log timestamp for precise time-based analysis
			zap.String("operation", operation),
			zap.String("request_id", requestID), // Include request ID in log context for traceability across logs
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("ip", c.ClientIP()),
			zap.Int("status", c.Writer.Status()), // Log the HTTP status code of the response for monitoring error rates
			zap.Duration("latency", latency),     // Log the request processing latency for performance analysis
			zap.String("user-agent", c.Request.UserAgent()),
		)
	}
}

// LinkValidationMiddleware validates the access link from the request header.
func LinkValidationMiddleware(repo interfaces.PatientRepository, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		const operation = "LinkValidationMiddleware"
		requestID := utils.GetRequestID(c.Request.Context()) // Retrieve request ID from context for logging context

		accessLink := c.GetHeader("X-Access-Link") // Extract the access link from the "X-Access-Link" header
		if accessLink == "" {
			logger.Warn("Access link missing from header", zap.String("operation", operation), zap.String("request_id", requestID), zap.String("header", "X-Access-Link"))
			utils.RespondWithError(c, http.StatusBadRequest, "Access link is required") // Respond with 400 Bad Request for missing link - client-side error
			c.Abort()                                                                   // Abort the request processing chain - no further processing if link is missing
			return
		}

		// Basic UUID format validation - BE-003, BE-055 - Validate that the access link is a valid UUID format
		if err := utils.ValidateUUID(accessLink); err != nil {
			logger.Warn("Invalid access link format", zap.String("operation", operation), zap.String("request_id", requestID), zap.String("access_link", accessLink), zap.Error(err))
			utils.RespondWithError(c, http.StatusBadRequest, "Invalid access link format") // Respond with 400 for invalid format - client-side error
			c.Abort()                                                                      // Abort request
			return
		}

		// Fetch patient session by link - BE-003 - Retrieve patient session from the repository using the access link
		patientSession, err := repo.GetPatientSessionByLink(c.Request.Context(), accessLink)
		if err != nil {
			if _, ok := err.(*domain.NotFoundError); ok { // Check if the error is a domain-specific NotFoundError
				logger.Warn("Access link not found", zap.String("operation", operation), zap.String("request_id", requestID), zap.String("access_link", accessLink), zap.Error(err))
				utils.RespondWithError(c, http.StatusNotFound, "Access link not found or invalid") // Respond with 404 Not Found to avoid exposing link validity - client-side error (but potentially also server-side data issue)
			} else {
				logger.Error("Error validating access link", zap.String("operation", operation), zap.String("request_id", requestID), zap.String("access_link", accessLink), zap.Error(err))
				utils.RespondWithError(c, http.StatusInternalServerError, "Error validating access link") // Respond with 500 Internal Server Error for backend issues - server-side error
			}
			c.Abort() // Abort request processing
			return
		}

		// Check if link is expired - BE-024, US-001 - Verify if the access link has expired based on the expiration timestamp
		if time.Now().UTC().After(patientSession.ExpirationTimestamp) {
			logger.Warn("Expired access link", zap.String("operation", operation), zap.String("request_id", requestID), zap.String("access_link", accessLink), zap.Time("expiration_timestamp", patientSession.ExpirationTimestamp))
			utils.RespondWithError(c, http.StatusGone, "Access link has expired") // Respond with 410 Gone for expired links - client-side error (link was valid but is no longer)
			c.Abort()                                                             // Abort request
			return
		}

		// Check if link already used - BE-003, US-001 - Ensure the access link has not been used before to maintain single-use policy
		if patientSession.Used {
			logger.Warn("Used access link", zap.String("operation", operation), zap.String("request_id", requestID), zap.String("access_link", accessLink))
			utils.RespondWithError(c, http.StatusGone, "Access link already used") // Respond with 410 Gone for used links - client-side error (link was valid but is no longer valid for repeated use)
			c.Abort()                                                              // Abort request
			return
		}

		// Link is valid: Add patient session ID (pseudonym) to context - BE-003 - Store patient session ID in Gin context for use in subsequent handlers
		c.Set("patientID", patientSession.SessionID)

		c.Next() // Go to the next middleware/handler - Proceed to the next stage in request processing if link is valid
	}
}
