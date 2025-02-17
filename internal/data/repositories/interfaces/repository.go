// internal/data/repositories/interfaces/repository.go
package interfaces

import (
	"context"

	"github.com/jackc/pgx/v5"
)

// Repository is a common interface for basic database operations, primarily
// for transaction management.  It's optional but promotes consistency.
type Repository interface {
	// BeginTx starts a new database transaction.  It accepts a context and
	// optional transaction options (e.g., isolation level).
	BeginTx(ctx context.Context, opts ...pgx.TxOptions) (pgx.Tx, error)

	// CommitTx commits an existing transaction.
	CommitTx(ctx context.Context, tx pgx.Tx) error

	// RollbackTx rolls back an existing transaction.
	RollbackTx(ctx context.Context, tx pgx.Tx) error
}
