package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stackvity/lung-server/internal/config" // Import the config package
	"go.uber.org/zap"                                  // Import zap for structured logging
)

// NewPostgresDB creates a new PostgreSQL database connection pool.
func NewPostgresDB(ctx context.Context, cfg *config.Config, logger *zap.Logger) (*pgxpool.Pool, error) { // Added logger
	// Construct the connection string.  Using fmt.Sprintf for clarity.
	// IMPORTANT: In production, NEVER store DB_PASSWORD in environment variables directly.
	// Use a dedicated secrets management service (e.g., HashiCorp Vault, AWS Secrets Manager).
	connString := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName, cfg.DBSslMode)

	// Create a configuration for the connection pool.
	poolConfig, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse connection string: %w", err)
	}

	// Set connection pool settings from the configuration.
	poolConfig.MaxConns = int32(cfg.DBMaxOpenConns)
	poolConfig.MinConns = int32(cfg.DBMaxIdleConns) // Use MinConns for idle connections

	// Set connection pool settings from the configuration.
	poolConfig.MaxConns = int32(cfg.DBMaxOpenConns)
	poolConfig.MinConns = int32(cfg.DBMaxIdleConns) // Use MinConns for idle connections

	// Use context with config timeout
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	poolConfig.MaxConnLifetime = cfg.DBConnMaxLifetime
	poolConfig.MaxConnIdleTime = cfg.DBConnMaxIdleTime

	//  Establish the connection pool.
	dbpool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Ping the database to verify the connection (also using the context).
	if err := dbpool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info("Successfully connected to PostgreSQL database!",
		zap.String("host", cfg.DBHost),
		zap.Int("port", cfg.DBPort),
		zap.String("database", cfg.DBName),
	) // Structured log message
	return dbpool, nil
}
