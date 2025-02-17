// internal/gemini/prompts.go
package gemini

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync" // Import sync for using sync.Map

	"github.com/pelletier/go-toml"                     // Import for TOML support
	"github.com/stackvity/lung-server/internal/domain" // Import domain for custom errors
	"go.uber.org/zap"
	"gopkg.in/yaml.v3" // Import for YAML support
)

// PromptManager defines the interface for managing Gemini API prompts.
// This interface abstracts prompt retrieval for various storage mechanisms.
type PromptManager interface {
	GetPrompt(ctx context.Context, promptID string) (string, error)
}

// FilePromptManager implements PromptManager, loading prompts from JSON, YAML, and TOML files.
// It utilizes an in-memory cache (sync.Map) for efficient prompt retrieval.
type FilePromptManager struct {
	promptTemplates map[string]string // In-memory map to store prompt templates, key is prompt ID, value is the prompt template string.
	templatePath    string            // Path to the directory containing prompt template files.
	logger          *zap.Logger       // Logger for structured logging.
	promptCache     sync.Map          // ADDED: In-memory cache for prompts, using sync.Map for concurrent access safety.
}

// NewFilePromptManager creates a new FilePromptManager, loads prompts from files, and initializes the cache.
// It takes a path to the directory containing prompt files as input.
// Returns a FilePromptManager instance and an error if prompt loading fails.
func NewFilePromptManager(templatePath string, logger *zap.Logger) (*FilePromptManager, error) {
	const operation = "NewFilePromptManager"

	fpm := &FilePromptManager{
		promptTemplates: make(map[string]string),
		templatePath:    templatePath,
		logger:          logger.Named("FilePromptManager"),
		promptCache:     sync.Map{}, // Initialize the cache - Recommendation 1
	}

	err := fpm.loadPromptsFromFiles()
	if err != nil {
		fpm.logger.Error("Failed to load prompts from files", zap.String("operation", operation), zap.Error(err))
		return nil, fmt.Errorf("failed to initialize FilePromptManager: %w", err)
	}

	fpm.logger.Info("FilePromptManager initialized successfully", zap.String("operation", operation), zap.String("template_path", templatePath), zap.Int("prompt_count", len(fpm.promptTemplates)))
	return fpm, nil
}

// loadPromptsFromFiles reads and parses prompt files (JSON, YAML, TOML) from the template path.
func (fpm *FilePromptManager) loadPromptsFromFiles() error {
	const operation = "FilePromptManager.loadPromptsFromFiles"

	fpm.logger.Info("Loading prompts from files", zap.String("operation", operation), zap.String("template_path", fpm.templatePath))

	return filepath.Walk(fpm.templatePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fpm.logger.Error("Error accessing path", zap.String("operation", operation), zap.String("path", path), zap.Error(err))
			return err
		}
		if info.IsDir() {
			return nil
		}

		var promptMap map[string]string
		file, err := os.ReadFile(path)
		if err != nil {
			fpm.logger.Error("Error reading prompt file", zap.String("operation", operation), zap.String("path", path), zap.Error(err))
			return fmt.Errorf("error reading prompt file %s: %w", path, err)
		}

		// Determine file type and unmarshal accordingly - Recommendation 2
		switch ext := filepath.Ext(path); ext {
		case ".json":
			if err := json.Unmarshal(file, &promptMap); err != nil {
				fpm.logger.Error("Error parsing JSON", zap.String("operation", operation), zap.String("path", path), zap.Error(err))
				return fmt.Errorf("error parsing JSON file %s: %w", path, err)
			}
		case ".yaml", ".yml": // Support YAML format - Recommendation 2
			if err := yaml.Unmarshal(file, &promptMap); err != nil {
				fpm.logger.Error("Error parsing YAML", zap.String("operation", operation), zap.String("path", path), zap.Error(err))
				return fmt.Errorf("error parsing YAML file %s: %w", path, err)
			}
		case ".toml": // Support TOML format - Recommendation 2
			if err := toml.Unmarshal(file, &promptMap); err != nil {
				fpm.logger.Error("Error parsing TOML", zap.String("operation", operation), zap.String("path", path), zap.Error(err))
				return fmt.Errorf("error parsing TOML file %s: %w", path, err)
			}
		default:
			fpm.logger.Debug("Skipping unsupported file type", zap.String("operation", operation), zap.String("path", path), zap.String("extension", ext))
			return nil // Skip unsupported file types
		}

		promptID := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
		for _, prompt := range promptMap {
			fpm.promptTemplates[promptID] = prompt
			fpm.promptCache.Store(promptID, prompt) // Store prompt in cache - Recommendation 1
		}
		fpm.logger.Debug("Loaded prompt from file", zap.String("operation", operation), zap.String("path", path), zap.String("prompt_id", promptID))
		return nil
	})
}

// GetPrompt retrieves a prompt template by its ID, first checking the cache and then loading from files if cache miss.
func (fpm *FilePromptManager) GetPrompt(ctx context.Context, promptID string) (string, error) {
	const operation = "FilePromptManager.GetPrompt"

	fpm.logger.Debug("Retrieving prompt", zap.String("operation", operation), zap.String("prompt_id", promptID))

	// 1. Cache Lookup - Recommendation 1
	if cachedPrompt, ok := fpm.promptCache.Load(promptID); ok {
		prompt, ok := cachedPrompt.(string) // Type assertion
		if !ok {
			return "", fmt.Errorf("%s: invalid data type in cache for prompt ID: %s", operation, promptID) // Error for unexpected cache value type
		}
		fpm.logger.Debug("Prompt retrieved from cache", zap.String("operation", operation), zap.String("prompt_id", promptID)) // Debug log for cache hit
		return prompt, nil                                                                                                     // Return cached prompt
	}

	// 2. Cache Miss: Load from Templates Map (which loads from files on initialization)
	prompt, ok := fpm.promptTemplates[promptID]
	if !ok {
		notFoundErr := domain.NewNotFoundError("Gemini prompt template", promptID)
		notFoundErr.SetLogger(fpm.logger)
		fpm.logger.Warn("Prompt template not found", zap.String("operation", operation), zap.String("prompt_id", promptID), zap.Error(notFoundErr)) // Warn log for not found prompt
		return "", notFoundErr
	}

	// 3. Store in Cache after successful retrieval from files - Recommendation 1 (Cache population on cache miss)
	fpm.promptCache.Store(promptID, prompt)
	fpm.logger.Debug("Prompt retrieved from files and stored in cache", zap.String("operation", operation), zap.String("prompt_id", promptID)) // Debug log for cache population

	return prompt, nil
}
