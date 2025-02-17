// internal/api/api.go
package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stackvity/lung-server/internal/api/handlers"
	"github.com/stackvity/lung-server/internal/api/routes"
	"github.com/stackvity/lung-server/internal/config"
	"go.uber.org/zap"
)

// API struct encapsulates the Gin engine and handler dependencies for the API server.
// It is designed to manage the lifecycle and dependencies of the entire API application.
type API struct {
	Engine  *gin.Engine
	Handler *handlers.Handler
	Config  *config.Config
	Logger  *zap.Logger
}

// NewAPI creates and configures a new API instance.
// It initializes the Gin engine, sets up middleware, registers routes using routes.SetupRouter,
// and returns a fully configured API instance.
//
// Dependencies:
//   - handler *handlers.Handler:  Struct containing all API handlers, injected for dependency inversion and modularity.
//   - cfg *config.Config: Application configuration, providing settings for server, database, and other components.
//   - logger *zap.Logger: Structured logger for consistent and detailed logging across the API.
//
// Returns:
//   - *API: A pointer to the newly created and configured API instance.
//   - error: An error if API initialization fails at any step.
func NewAPI(handler *handlers.Handler, cfg *config.Config, logger *zap.Logger) (*API, error) {
	const operation = "api.NewAPI"

	logger.Info("Initializing API", zap.String("operation", operation))

	// 1. Gin Engine Setup
	if cfg.Environment == config.DevelopmentEnvironment {
		gin.SetMode(gin.DebugMode)
		logger.Debug("Gin mode set to Debug Mode", zap.String("operation", operation), zap.String("environment", cfg.Environment))
	} else {
		gin.SetMode(gin.ReleaseMode)
		logger.Info("Gin mode set to Release Mode", zap.String("operation", operation), zap.String("environment", cfg.Environment))
	}
	engine := gin.New()

	// 2. Middleware Setup
	engine.Use(handlers.MiddlewareSetup(handlers.MiddlewareConfig{
		PatientRepo: handler.FileHandler.ProcessingService.GetPatientRepository(), // CORRECT - Access via getter method
		Logger:      logger,
		Config:      cfg,
	}))

	// 3. Route Setup
	routes.SetupRouter(engine, handler.FileHandler, handler.ReportHandler, handler.HealthHandler, handler.DiagnosisHandler)

	api := &API{
		Engine:  engine,
		Handler: handler,
		Config:  cfg,
		Logger:  logger,
	}

	logger.Info("API initialized successfully", zap.String("operation", operation))
	return api, nil
}

// StartServer starts the Gin HTTP server.
func (api *API) StartServer() error {
	const operation = "api.StartServer"

	api.Logger.Info("Starting HTTP server", zap.String("operation", operation), zap.String("address", api.Config.HTTPServerAddress), zap.String("environment", api.Config.Environment))

	server := &http.Server{
		Addr:    api.Config.HTTPServerAddress,
		Handler: api.Engine,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			api.Logger.Fatal("HTTP server failed to start", zap.String("operation", operation), zap.Error(err))
		}
	}()

	<-quit
	api.Logger.Info("Shutting down server...", zap.String("operation", operation))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		api.Logger.Fatal("Server forced to shutdown", zap.String("operation", operation), zap.Error(err))
		return fmt.Errorf("server shutdown forced: %w", err)
	}

	api.Logger.Info("Server exited properly", zap.String("operation", operation))
	return nil
}
