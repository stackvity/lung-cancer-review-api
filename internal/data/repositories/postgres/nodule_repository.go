// internal/data/repositories/postgres/nodule_repository.go
package postgres

import (
	"context"
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stackvity/lung-server/internal/data/models"
	"github.com/stackvity/lung-server/internal/data/repositories/interfaces"
	postgres "github.com/stackvity/lung-server/internal/data/repositories/sqlc"
	"github.com/stackvity/lung-server/internal/domain"
	"github.com/stackvity/lung-server/internal/utils"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var _ interfaces.NoduleRepository = (*NoduleRepository)(nil)

// NoduleRepository implements the interfaces.NoduleRepository for PostgreSQL.
type NoduleRepository struct {
	db      *pgxpool.Pool
	queries *postgres.Queries // Use the generated Queries struct
	logger  *zap.Logger
}

// NewNoduleRepository creates a new NoduleRepository instance.
func NewNoduleRepository(db *pgxpool.Pool, logger *zap.Logger) *NoduleRepository {
	return &NoduleRepository{
		db:      db,
		queries: postgres.New(), // Initialize sqlc Queries
		logger:  logger,
	}
}

// CreateNodule implements interfaces.NoduleRepository.
func (r *NoduleRepository) CreateNodule(ctx context.Context, nodule *models.Nodule) error {
	const operation = "postgres.NoduleRepository.CreateNodule"
	requestID := utils.GetRequestID(ctx.(*gin.Context))

	r.logger.Debug("Starting DB operation", zap.String("operation", operation), zap.String("nodule_id", nodule.ID.String()), zap.String("request_id", requestID))

	params := &postgres.CreateFindingParams{ // Corrected to use CreateFindingParams
		FindingID:        pgtype.UUID{Bytes: uuid.UUID(nodule.ID), Valid: true},
		FileID:           pgtype.UUID{Bytes: uuid.UUID(nodule.ImageID), Valid: true}, // Corrected to use FileID
		FindingType:      "nodule",                                                   // Hardcoded to nodule type
		Description:      nodule.Location,                                            // Mapped to Location
		ImageCoordinates: []float64{nodule.Size},                                     // Mapped to Size (as first element)
		Source:           nodule.Shape,                                               // Mapped to Shape
	}

	if r.logger.Core().Enabled(zapcore.DebugLevel) {
		r.logger.Debug("DB parameters", zap.String("operation", operation), zap.String("request_id", requestID), zap.Any("params", params))
	}

	_, err := r.queries.CreateFinding(ctx, r.db, params) // Corrected to use CreateFinding
	if err != nil {
		r.logger.Error("DB error in CreateNodule", zap.String("operation", operation), zap.String("request_id", requestID), zap.Error(err))
		dbErr := utils.NewErrDBQuery("CreateNodule failed", operation, "CreateNodule", params, err) // Enhanced error
		if setLogger, ok := dbErr.(interface{ SetLogger(*zap.Logger) }); ok {                       // Type assertion to ErrDBQuery interface
			setLogger.SetLogger(r.logger)
		}
		return dbErr
	}

	r.logger.Debug("Successfully completed DB operation", zap.String("operation", operation), zap.String("nodule_id", nodule.ID.String()), zap.String("request_id", requestID))
	return nil
}

// GetNoduleByID implements interfaces.NoduleRepository.
func (r *NoduleRepository) GetNoduleByID(ctx context.Context, noduleID uuid.UUID) (*models.Nodule, error) {
	const operation = "postgres.NoduleRepository.GetNoduleByID"
	requestID := utils.GetRequestID(ctx.(*gin.Context))
	r.logger.Debug("Starting DB operation", zap.String("operation", operation), zap.String("nodule_id", noduleID.String()), zap.String("request_id", requestID))

	noduleRow, err := r.queries.GetNoduleByID(ctx, r.db, pgtype.UUID{Bytes: uuid.UUID(noduleID), Valid: true}) // Corrected to use GetNoduleByID and renamed variable
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			r.logger.Warn("Nodule not found", zap.String("operation", operation), zap.String("nodule_id", noduleID.String()), zap.String("request_id", requestID), zap.Error(err))
			notFoundErr := domain.NewNotFoundError("nodule", noduleID.String()) // Use domain-specific NotFoundError
			notFoundErr.SetLogger(r.logger)                                     // Set logger for domain error
			return nil, notFoundErr
		}
		r.logger.Error("DB error in GetNoduleByID", zap.String("operation", operation), zap.String("nodule_id", noduleID.String()), zap.String("request_id", requestID), zap.Error(err))
		dbQueryErr := utils.NewErrDBQuery("GetNoduleByID failed", operation, "GetNoduleByID", noduleID.String(), err) // Enhanced error
		if setLoggerErr, ok := dbQueryErr.(interface{ SetLogger(*zap.Logger) }); ok {                                 // Type assertion with interface
			setLoggerErr.SetLogger(r.logger)
		}
		return nil, dbQueryErr
	}

	modelNodule := &models.Nodule{
		ID:       uuid.UUID(noduleRow.FindingID.Bytes), // CORRECTED: Use FindingID from noduleRow
		ImageID:  uuid.UUID(noduleRow.FileID.Bytes),    // CORRECTED: Use FileID from noduleRow
		Location: noduleRow.Description,                // Corrected: Use noduleRow.Description
		Size:     noduleRow.ImageCoordinates[0],        // Corrected: Use noduleRow.ImageCoordinates
		Shape:    noduleRow.Source,                     // Corrected: Use noduleRow.Source
	}

	r.logger.Debug("Successfully completed DB operation", zap.String("operation", operation), zap.String("nodule_id", noduleID.String()), zap.String("request_id", requestID))
	return modelNodule, nil
}

// BeginTx implements interfaces.Repository.
func (r *NoduleRepository) BeginTx(ctx context.Context, opts ...pgx.TxOptions) (pgx.Tx, error) {
	r.logger.Debug("Starting transaction", zap.String("operation", "postgres.NoduleRepository.BeginTx"), zap.String("request_id", utils.GetRequestID(ctx.(*gin.Context))))
	if len(opts) > 0 {
		return r.db.BeginTx(ctx, opts[0])
	}
	return r.db.Begin(ctx)
}

// CommitTx implements interfaces.Repository.
func (r *NoduleRepository) CommitTx(ctx context.Context, tx pgx.Tx) error {
	r.logger.Debug("Committing transaction", zap.String("operation", "postgres.NoduleRepository.CommitTx"), zap.String("request_id", utils.GetRequestID(ctx.(*gin.Context))))
	return tx.Commit(ctx)
}

// RollbackTx implements interfaces.Repository.
func (r *NoduleRepository) RollbackTx(ctx context.Context, tx pgx.Tx) error {
	r.logger.Debug("Rolling back transaction", zap.String("operation", "postgres.NoduleRepository.RollbackTx"), zap.String("request_id", utils.GetRequestID(ctx.(*gin.Context))))
	return tx.Rollback(ctx)
}
