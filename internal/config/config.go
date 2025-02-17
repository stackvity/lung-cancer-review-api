// internal/config/config.go
package config

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/spf13/viper"
)

// Config stores all the configuration settings for the application.
// It uses `mapstructure` tags for automatic unmarshaling from Viper configurations.
// This struct is designed to hold environment-specific and application-wide settings,
// loaded from environment variables and/or a .env file.
type Config struct {
	Environment       string   `mapstructure:"ENVIRONMENT"`         // "development", "staging", "production"
	HTTPServerAddress string   `mapstructure:"HTTP_SERVER_ADDRESS"` // Address (host:port) for the HTTP server to listen on. Example: ":8080"
	LogLevel          string   `mapstructure:"LOG_LEVEL"`           // Logging level for Zap logger (debug, info, warn, error, fatal). Default: "info"
	LogFormat         string   `mapstructure:"LOG_FORMAT"`          // Logging format ("text" or "json"). Default: "text"
	SentryDSN         string   `mapstructure:"SENTRY_DSN"`          // Sentry Data Source Name for error tracking (optional, but recommended for production)
	SentryEnvironment string   `mapstructure:"SENTRY_ENVIRONMENT"`  // Environment name for Sentry, e.g., "development", "staging", "production"
	AllowedOrigins    []string `mapstructure:"ALLOWED_ORIGINS"`     // CORS allowed origins (for web UI access). Comma-separated list or YAML/TOML array.

	DBDriver          string        `mapstructure:"DB_DRIVER"`             // Database driver name (e.g., "postgres")
	DBHost            string        `mapstructure:"DB_HOST"`               // Database host address (e.g., "localhost" or IP address)
	DBPort            int           `mapstructure:"DB_PORT"`               // Database port number (e.g., 5432)
	DBUser            string        `mapstructure:"DB_USER"`               // Database username
	DBPassword        string        `mapstructure:"DB_PASSWORD"`           // Database password (sensitive, use secrets management in production)
	DBName            string        `mapstructure:"DB_NAME"`               // Database name
	DBSslMode         string        `mapstructure:"DB_SSL_MODE"`           // Database SSL mode (e.g., "disable", "require", "verify-full")
	DBMaxOpenConns    int           `mapstructure:"DB_MAX_OPEN_CONNS"`     // Maximum number of open database connections in the connection pool
	DBMaxIdleConns    int           `mapstructure:"DB_MAX_IDLE_CONNS"`     // Maximum number of idle database connections in the connection pool
	DBConnMaxLifetime time.Duration `mapstructure:"DB_CONN_MAX_LIFETIME"`  // Maximum lifetime of a database connection before it's closed and reopened
	DBConnMaxIdleTime time.Duration `mapstructure:"DB_CONN_MAX_IDLE_TIME"` // Maximum time a connection may be idle before being closed

	GeminiAPIKey     string        `mapstructure:"GEMINI_API_KEY"`     // Google AI Gemini API key (sensitive, use secrets management in production)
	GeminiAPITimeout time.Duration `mapstructure:"GEMINI_API_TIMEOUT"` // Timeout for Gemini API calls, e.g., "30s", "1m"
	GcloudProject    string        `mapstructure:"GCLOUD_PROJECT"`     // Google Cloud Project ID (required if using Google Cloud services)

	ReportTemplatePath string `mapstructure:"REPORT_TEMPLATE_PATH"` // Path to the directory containing report templates (e.g., "./internal/pdf/templates")

	StorageType        string `mapstructure:"STORAGE_TYPE"`         // Storage type: "cloud" (AWS S3, GCP Storage, Azure Blob Storage) or "local" (local filesystem)
	CloudStorageBucket string `mapstructure:"CLOUD_STORAGE_BUCKET"` // Name of the cloud storage bucket (required if STORAGE_TYPE=cloud, sensitive!)
	AWSRegion          string `mapstructure:"AWS_REGION"`           // AWS region for cloud storage (e.g., "us-east-1") - Required for AWS S3
	FileEncryptionKey  string `mapstructure:"FILE_ENCRYPTION_KEY"`  // Encryption key for file storage (sensitive, use secrets management in production!)
	MaxFileSize        int64  `mapstructure:"MAX_FILE_SIZE"`        // Maximum allowed file upload size in bytes (e.g., 50MB = 50 * 1024 * 1024)

	LinkExpiration time.Duration `mapstructure:"LINK_EXPIRATION"` // Duration for which access links are valid (e.g., "24h", "48h")
	DataRetention  time.Duration `mapstructure:"DATA_RETENTION"`  // Duration for which patient data is retained before secure deletion (e.g., "90d", "180d")
}

const DevelopmentEnvironment = "development" // Constant defining the "development" environment string

// LoadConfig reads configuration from environment variables and/or a .env file using Viper.
// It populates the Config struct with values from environment variables, falling back to defaults or values from a .env file if set.
// Returns a Config struct containing the loaded configuration and an error if configuration loading fails.
func LoadConfig(ctx context.Context, path string) (config Config, err error) {
	viper.AddConfigPath(path)   // Add the config path to Viper's lookup paths
	viper.SetConfigName(".env") // Set the base name of the config file (without extension) to ".env"
	viper.SetConfigType("env")  // Set the config file type to "env" for .env file format

	viper.AutomaticEnv()      // Enable automatic reading of environment variables
	viper.AllowEmptyEnv(true) // Allow empty environment variables to be read without error

	if err = viper.ReadInConfig(); err != nil { // Attempt to read config from the configured paths and file name
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// .env file not found; not a fatal error, proceed with environment variables or defaults
			log.Println("No .env file found, relying on environment variables.") // Log informational message, not error
		} else {
			// Config file was found, but another error occurred during reading or parsing
			return Config{}, fmt.Errorf("failed to read config file: %w", err) // Return error with context for config file read failure
		}
	}

	if err = viper.Unmarshal(&config); err != nil { // Unmarshal the configuration into the Config struct
		return Config{}, fmt.Errorf("failed to unmarshal config: %w", err) // Return error with context for unmarshaling failure
	}

	// --- Configuration Validation (with specific error messages for required variables) ---
	// Ensure essential configuration parameters are set. Return specific errors if they are missing.
	if config.DBDriver == "" {
		return Config{}, fmt.Errorf("environment variable DB_DRIVER is required") // Return error if DB_DRIVER is not set
	}
	if config.DBHost == "" {
		return Config{}, fmt.Errorf("environment variable DB_HOST is required") // Return error if DB_HOST is not set
	}
	if config.DBPort == 0 { // 0 is generally an invalid port and likely indicates unset DB_PORT
		return Config{}, fmt.Errorf("environment variable DB_PORT is required") // Return error if DB_PORT is not set or invalid
	}
	if config.DBUser == "" {
		return Config{}, fmt.Errorf("environment variable DB_USER is required") // Return error if DB_USER is not set
	}
	if config.DBPassword == "" {
		return Config{}, fmt.Errorf("environment variable DB_PASSWORD is required") // Return error if DB_PASSWORD is not set
	}
	if config.DBName == "" {
		return Config{}, fmt.Errorf("environment variable DB_NAME is required") // Return error if DB_NAME is not set
	}
	if config.DBSslMode == "" {
		return Config{}, fmt.Errorf("environment variable DB_SSL_MODE is required") // Return error if DB_SSL_MODE is not set
	}
	if config.HTTPServerAddress == "" {
		return Config{}, fmt.Errorf("environment variable HTTP_SERVER_ADDRESS is required") // Return error if HTTP_SERVER_ADDRESS is not set
	}

	// Security-Critical Check: Ensure FILE_ENCRYPTION_KEY is set in non-development environments.
	// This is MANDATORY for data protection in staging and production.
	if config.FileEncryptionKey == "" && os.Getenv("ENVIRONMENT") != DevelopmentEnvironment {
		return Config{}, fmt.Errorf("environment variable FILE_ENCRYPTION_KEY is required in non-development environments") // Return error if FILE_ENCRYPTION_KEY is missing in production
	}

	// Storage Type Validation and Defaults
	if config.StorageType == "" {
		config.StorageType = "cloud"                               // Default to cloud storage if STORAGE_TYPE is not explicitly set
		log.Println("STORAGE_TYPE not set, defaulting to 'cloud'") // Log default value assignment
	}
	if config.CloudStorageBucket == "" && config.StorageType == "cloud" {
		return Config{}, fmt.Errorf("environment variable CLOUD_STORAGE_BUCKET is required when STORAGE_TYPE is 'cloud'") // Return error if CLOUD_STORAGE_BUCKET is missing for cloud storage
	}
	if config.AWSRegion == "" && config.StorageType == "cloud" { // Ensure AWS_REGION is set for cloud storage
		return Config{}, fmt.Errorf("environment variable AWS_REGION is required when STORAGE_TYPE is 'cloud'")
	}

	// Duration Configuration Defaults and Logging - Apply default values and log if durations are not explicitly set in config
	if config.LinkExpiration == 0 {
		config.LinkExpiration = 24 * time.Hour                          // Default link expiration to 24 hours if not set
		log.Println("LINK_EXPIRATION not set, using default: 24 hours") // Log default value assignment
	}

	if config.DataRetention == 0 {
		config.DataRetention = 90 * 24 * time.Hour                    // Default data retention to 90 days if not set
		log.Println("DATA_RETENTION not set, using default: 90 days") // Log default value assignment
	}

	if config.LogLevel == "" {
		config.LogLevel = "info"                               // Default log level to "info" if not set
		log.Println("LOG_LEVEL not set, defaulting to 'info'") // Log default value assignment
	}
	if config.LogFormat == "" {
		config.LogFormat = "text"                               // Default log format to "text" if not set
		log.Println("LOG_FORMAT not set, defaulting to 'text'") // Log default value assignment
	}
	if config.GeminiAPITimeout == 0 {
		config.GeminiAPITimeout = 30 * time.Second                          // Default Gemini API timeout to 30 seconds
		log.Println("GEMINI_API_TIMEOUT not set, defaulting to 30 seconds") // Log default value assignment
	}
	if config.MaxFileSize == 0 {
		config.MaxFileSize = 50 * 1024 * 1024                    // Default max file size to 50MB if not set
		log.Println("MAX_FILE_SIZE not set, defaulting to 50MB") // Log default value assignment
	}

	// Basic Context Handling Example (for future expansion - not strictly necessary for config loading itself, but good practice)
	select {
	case <-ctx.Done():
		return Config{}, ctx.Err() // Return context error if context is cancelled during config loading
	default:
		// Proceed with normal config loading if context is not cancelled
	}

	return // Return the populated Config struct and nil error for successful config loading
}
