// internal/data/data.go
package data

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stackvity/lung-server/internal/config"
	postgres "github.com/stackvity/lung-server/internal/data/repositories/sqlc" // Import with alias
	database "github.com/stackvity/lung-server/pkg/database/postgres"           // CORRECTED: Import the postgres package
	"go.uber.org/zap"
)

// Data provides access to all data layer components.
type Data struct {
	DB      *pgxpool.Pool    // Exported for potential direct access (e.g., migrations)
	Queries postgres.Querier // Use the generated Querier interface
	// Add repositories here (when needed for dependency injection,
	// but currently, they are initialized within services, not here).
}

// NewData creates a new Data instance, establishing the database connection.
func NewData(ctx context.Context, cfg *config.Config, logger *zap.Logger) (*Data, error) {
	dbpool, err := database.NewPostgresDB(ctx, cfg, logger) // CORRECTED: Use the database package
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Create an instance of the generated Queries struct, using the dbpool
	queries := postgres.New()
	var querier postgres.Querier = queries

	return &Data{
		DB:      dbpool,
		Queries: querier, // Use the interface
	}, nil
}

// Close closes the database connection pool.  It should be called during
// application shutdown.
func (d *Data) Close() {
	if d.DB != nil {
		d.DB.Close()
	}
}
