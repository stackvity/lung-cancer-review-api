package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"runtime/debug" // Correct import path for debug.Stack()

	"github.com/stackvity/lung-server/internal/config" // Import the config package
	"github.com/stackvity/lung-server/internal/utils"  // Import the utils package
	"go.uber.org/zap"
)

func main() {
	// 1. Load Configuration
	cfg, err := config.LoadConfig(context.Background(), ".")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
		os.Exit(1)
	}

	// 2. Initialize Logger
	logger, err := utils.NewLogger(&cfg)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
		os.Exit(1)
	}
	defer func() {
		if syncErr := logger.Sync(); syncErr != nil {
			log.Printf("Failed to sync logger during shutdown: %v", syncErr)
		}
	}()

	logger.Info("Starting Lung Cancer Review API service...", zap.String("version", "1.0.0"))

	// 3. Initialize API via Wire
	app, cleanup, err := InitializeAPI()
	if err != nil {
		logger.Error("Failed to initialize API", zap.Error(err))
		os.Exit(1)
	}
	defer func() {
		cleanup()
	}()

	// 4. Start HTTP Server
	if err := app.StartServer(); err != nil {
		logger.Error("API server failed to start", zap.Error(err))
		os.Exit(1)
	}

	// 5. Handle Panic Recovery
	defer func() {
		if r := recover(); r != nil {
			err := fmt.Errorf("application panicked: %v\nStack Trace: %s", r, debug.Stack())
			logger.Error("Panic recovered in main", zap.Error(err), zap.String("stack_trace", string(debug.Stack())))
			os.Exit(1)
		}
	}()

	logger.Info("Service stopped gracefully.")
}
