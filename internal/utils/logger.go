// internal/utils/logger.go
package utils

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/stackvity/lung-server/internal/config" // Import config
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Logger *zap.Logger // Global logger instance

func init() {
	// We now initialize the global Logger in the init() function.
	// This ensures that the logger is available as soon as the package is imported.
	cfg, err := config.LoadConfig(context.Background(), ".") // Load from current directory or env
	if err != nil {
		// Use standard log package to avoid dependency on the uninitialized logger.
		log.Printf("WARNING: Failed to load config, using default logger: %v", err)
		Logger, _ = zap.NewDevelopment() // Fallback logger
	} else {
		Logger, err = NewLogger(&cfg)
		if err != nil {
			log.Printf("WARNING: Failed to create logger, using default logger: %v", err)
			// Fallback to a basic logger if config loading fails
			Logger, _ = zap.NewDevelopment() // Discard the error, it's a fallback
		}
	}
}

// NewLogger creates a new Zap logger based on the provided configuration.
func NewLogger(cfg *config.Config) (*zap.Logger, error) {
	var loggerConfig zap.Config

	// Select configuration based on environment.
	if os.Getenv("ENVIRONMENT") == "production" {
		loggerConfig = zap.NewProductionConfig()
		loggerConfig.Sampling = nil // Disable sampling in production to capture ALL logs
	} else {
		loggerConfig = zap.NewDevelopmentConfig()
		loggerConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder // Add color
	}

	// Set the log level based on config.
	logLevel, err := zapcore.ParseLevel(cfg.LogLevel)
	if err != nil {
		return nil, fmt.Errorf("invalid log level: %w", err)
	}
	loggerConfig.Level = zap.NewAtomicLevelAt(logLevel)

	// Customize the encoder configuration for readability.
	loggerConfig.EncoderConfig.TimeKey = "timestamp" // Use "timestamp"
	loggerConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	// Configure JSON or console encoding.
	if cfg.LogFormat == "json" {
		loggerConfig.Encoding = "json"
	} else {
		loggerConfig.Encoding = "console"
	}

	logger, err := loggerConfig.Build() // Build at the end
	if err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}

	return logger, nil
}
