// internal/api/handlers/health_handler.go
package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool" // Import pgxpool for database interaction
	"go.uber.org/zap"                 // Import zap for structured logging
)

// HealthHandler handles health check requests for the application.
// It is responsible for providing endpoints that can be used by monitoring systems and load balancers
// to verify the health and operational status of the application and its dependencies.
type HealthHandler struct {
	dbPool *pgxpool.Pool // dbPool: Database connection pool dependency for database health checks. Injected during handler creation for modularity and testability.
	logger *zap.Logger   // logger: Logger dependency for structured logging within the handler.  Injected for consistent logging practices.
}

// NewHealthHandler creates a new HealthHandler instance.
// It takes a *pgxpool.Pool and a *zap.Logger as dependencies, enabling database health checks and structured logging.
// This constructor ensures that the handler is properly initialized with its required dependencies.
func NewHealthHandler(dbPool *pgxpool.Pool, logger *zap.Logger) *HealthHandler { // MODIFIED: Accept dbPool and logger
	return &HealthHandler{
		dbPool: dbPool,
		logger: logger.Named("HealthHandler"), // logger: Creates a named logger for HealthHandler to provide contextual information in logs.
	}
}

// HealthCheck performs a comprehensive health check of the application.
// It verifies the overall application status, including its database connectivity,
// and returns a detailed JSON response indicating the health of each component.
// This endpoint is designed to be used by monitoring systems, load balancers, and orchestration platforms
// to automatically assess the health of the application and route traffic accordingly.
//
// Returns:
//   - 200 OK: If the application and all critical dependencies (currently just the database) are healthy. The response body contains a JSON payload with detailed health information.
//   - 503 Service Unavailable: If any critical dependency is unhealthy (e.g., database connection fails). The response body contains a JSON payload detailing the degraded health status.
func (h *HealthHandler) HealthCheck(c *gin.Context) {
	const operation = "HealthHandler.HealthCheck" // operation: Operation name for structured logging, aids in log filtering and analysis.

	healthStatus := gin.H{ // healthStatus: Use gin.H for constructing the JSON response payload. gin.H is a convenience type for creating maps for JSON responses in Gin.
		"status":    "OK",             // status: Overall application status. Initialized to "OK" and may be updated to "Degraded" if any health checks fail.
		"database":  "OK",             // database: Database component status. Initialized to "OK" and updated to "Degraded" if the database ping fails.
		"message":   "System healthy", // message: Human-readable message providing a summary of the health status. Defaults to "System healthy" and updated if degradation is detected.
		"timestamp": timeNow(),        // timestamp: Current timestamp in RFC3339 format, indicating when the health check was performed. Useful for time-series monitoring and correlation with other logs.
	}

	// Database Health Check - Perform a database ping to verify connectivity to the PostgreSQL database.
	if err := h.dbPool.Ping(c); err != nil { // dbPool.Ping: Execute a simple ping query against the database to check connection status.
		h.logger.Warn("Database health check failed", zap.String("operation", operation), zap.Error(err)) // Log database health check failure at Warn level for immediate attention.

		healthStatus["database"] = "Degraded"                                   // Update database status in healthStatus map to "Degraded" to reflect database connectivity issue.
		healthStatus["status"] = "Degraded"                                     // Update overall status in healthStatus map to "Degraded" as database is a critical dependency.
		healthStatus["message"] = "System degraded: database connection failed" // Update overall message in healthStatus map to provide user-friendly status information.
		c.JSON(http.StatusServiceUnavailable, healthStatus)                     // Respond with 503 Service Unavailable status code, indicating service degradation due to database issue.
		return                                                                  // Return immediately after sending 503 response to indicate unhealthy status and prevent further request processing.
	}

	// Future Enhancement: Add checks for other dependencies (e.g., Gemini API connectivity, caching service, etc.) - Recommendation: Extend Health Checks
	// Example (Conceptual - not implemented in this iteration):
	// if !h.geminiClient.IsHealthy(ctx) {
	// 	healthStatus["geminiAPI"] = "Unavailable"
	// 	healthStatus["status"] = "Degraded" // or "Warning" depending on criticality
	// 	healthStatus["message"] = "System degraded: Gemini API unavailable"
	// }
	// Future Enhancement: Metrics Exposure - Recommendation: Metrics Integration
	// In production, integrate with Prometheus or a similar metrics system to expose health check metrics.
	// Example (Conceptual - not implemented in this iteration):
	// prometheus.MustRegister(healthCheckSuccess) // Example: Register Prometheus metric
	// healthCheckSuccess.Inc() // Increment metric on successful health check

	// Respond with 200 OK status code and the structured JSON payload if all health checks pass.
	c.JSON(http.StatusOK, healthStatus)                                                                               // Respond with 200 OK status code to indicate healthy service.
	h.logger.Debug("Health check passed", zap.String("operation", operation), zap.Any("health_status", healthStatus)) // Debug log: Log successful health check and detailed status information at Debug level for detailed monitoring.

	// Future Enhancement: Alerting - Recommendation: Alerting Integration
	// In production, configure alerting based on health check status.
	// Example (Conceptual - not implemented in this iteration):
	// if healthStatus["status"] != "OK" {
	//   // Trigger alert via Sentry, Prometheus Alertmanager, or a dedicated alerting system.
	//   sentry.CaptureMessage("Health check degraded: " + healthStatus["message"].(string))
	// }
}

// timeNow is a helper function to get the current time in UTC and format it as RFC3339 string.
// This utility function ensures consistent timestamp formatting throughout the health check responses and logs,
// promoting uniformity and simplifying time-based analysis and correlation of events across different system components.
func timeNow() string {
	return time.Now().UTC().Format(time.RFC3339) // time.Now().UTC(): Get current time in UTC to ensure time zone consistency across systems. Format time in RFC3339 for standardized, human-readable timestamps in JSON responses and logs.
}
