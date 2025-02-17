// internal/data/repositories/postgres/diagnosis_repository.go
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
	postgres "github.com/stackvity/lung-server/internal/data/repositories/sqlc" // Alias to avoid naming conflict
	"github.com/stackvity/lung-server/internal/domain"
	"github.com/stackvity/lung-server/internal/utils"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var _ interfaces.DiagnosisRepository = (*DiagnosisRepository)(nil)

// DiagnosisRepository implements the interfaces.DiagnosisRepository for PostgreSQL.
type DiagnosisRepository struct {
	db      *pgxpool.Pool
	queries *postgres.Queries // Use the generated Queries struct
	logger  *zap.Logger
}

// NewDiagnosisRepository creates a new DiagnosisRepository instance.
func NewDiagnosisRepository(db *pgxpool.Pool, logger *zap.Logger) *DiagnosisRepository {
	return &DiagnosisRepository{
		db:      db,
		queries: postgres.New(), // Initialize sqlc Queries
		logger:  logger,
	}
}

// CreateDiagnosis implements interfaces.DiagnosisRepository.
func (r *DiagnosisRepository) CreateDiagnosis(ctx context.Context, diagnosis *models.Diagnosis) error {
	const operation = "postgres.DiagnosisRepository.CreateDiagnosis"
	requestID := utils.GetRequestID(ctx.(*gin.Context))

	r.logger.Debug("Starting DB operation", zap.String("operation", operation), zap.String("diagnosis_id", diagnosis.ID.String()), zap.String("request_id", requestID))

	params := &postgres.CreateDiagnosisParams{
		ResultID:      pgtype.UUID{Bytes: uuid.UUID(diagnosis.ResultID), Valid: true},
		SessionID:     pgtype.UUID{Bytes: uuid.UUID(diagnosis.SessionID), Valid: true},
		DiagnosisText: pgtype.Text{String: diagnosis.DiagnosisText, Valid: true},
		Confidence:    pgtype.Text{String: diagnosis.Confidence, Valid: true},
		Justification: pgtype.Text{String: diagnosis.Justification, Valid: true},
	}

	if r.logger.Core().Enabled(zapcore.DebugLevel) {
		r.logger.Debug("DB parameters", zap.String("operation", operation), zap.String("request_id", requestID), zap.Any("params", params))
	}

	_, err := r.queries.CreateDiagnosis(ctx, r.db, params)
	if err != nil {
		r.logger.Error("DB error in CreateDiagnosis", zap.String("operation", operation), zap.String("request_id", requestID), zap.Error(err))
		return utils.NewErrDBQuery("CreateDiagnosis failed", operation, "CreateDiagnosis", params, err) // Enhanced error
	}

	r.logger.Debug("Successfully completed DB operation", zap.String("operation", operation), zap.String("diagnosis_id", diagnosis.ID.String()), zap.String("request_id", requestID))
	return nil
}

// GetDiagnosisByID implements interfaces.DiagnosisRepository.
func (r *DiagnosisRepository) GetDiagnosisByID(ctx context.Context, diagnosisID uuid.UUID) (*models.Diagnosis, error) {
	const operation = "postgres.DiagnosisRepository.GetDiagnosisByID"
	requestID := utils.GetRequestID(ctx.(*gin.Context))
	r.logger.Debug("Starting DB operation", zap.String("operation", operation), zap.String("diagnosis_id", diagnosisID.String()), zap.String("request_id", requestID))

	diagnosis, err := r.queries.GetDiagnosisByID(ctx, r.db, pgtype.UUID{Bytes: uuid.UUID(diagnosisID), Valid: true})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			r.logger.Warn("Diagnosis not found", zap.String("operation", operation), zap.String("diagnosis_id", diagnosisID.String()), zap.String("request_id", requestID), zap.Error(err))
			return nil, domain.NewNotFoundError("diagnosis", diagnosisID.String())
		}
		r.logger.Error("DB error in GetDiagnosisByID", zap.String("operation", operation), zap.String("diagnosis_id", diagnosisID.String()), zap.String("request_id", requestID), zap.Error(err))
		dbErr := utils.NewErrDBQuery("GetDiagnosisByID failed", operation, "GetDiagnosisByID", diagnosisID.String(), err) // Enhanced error
		if setLoggerErr, ok := dbErr.(interface{ SetLogger(*zap.Logger) }); ok {
			setLoggerErr.SetLogger(r.logger)
		}
		return nil, dbErr
	}

	modelDiagnosis := &models.Diagnosis{
		ID:            uuid.UUID(diagnosis.ID.Bytes),
		ResultID:      uuid.UUID(diagnosis.ResultID.Bytes),
		SessionID:     uuid.UUID(diagnosis.SessionID.Bytes),
		DiagnosisText: diagnosis.DiagnosisText.String,
		Confidence:    diagnosis.Confidence.String,
		Justification: diagnosis.Justification.String,
		CreatedAt:     diagnosis.CreatedAt.Time,
		UpdatedAt:     diagnosis.UpdatedAt.Time,
	}

	r.logger.Debug("Successfully completed DB operation", zap.String("operation", operation), zap.String("diagnosis_id", diagnosisID.String()), zap.String("request_id", requestID))
	return modelDiagnosis, nil
}

// BeginTx implements interfaces.Repository.
func (r *DiagnosisRepository) BeginTx(ctx context.Context, opts ...pgx.TxOptions) (pgx.Tx, error) {
	requestID := utils.GetRequestID(ctx.(*gin.Context)) // Get request ID
	r.logger.Debug("Starting transaction", zap.String("operation", "postgres.DiagnosisRepository.BeginTx"), zap.String("request_id", requestID))
	if len(opts) > 0 {
		return r.db.BeginTx(ctx, opts[0])
	}
	return r.db.Begin(ctx)
}

// CommitTx implements interfaces.Repository.
func (r *DiagnosisRepository) CommitTx(ctx context.Context, tx pgx.Tx) error {
	requestID := utils.GetRequestID(ctx.(*gin.Context)) // Get request ID
	r.logger.Debug("Commiting transaction", zap.String("operation", "postgres.DiagnosisRepository.CommitTx"), zap.String("request_id", requestID))
	return tx.Commit(ctx)
}

// RollbackTx implements interfaces.Repository.
func (r *DiagnosisRepository) RollbackTx(ctx context.Context, tx pgx.Tx) error {
	requestID := utils.GetRequestID(ctx.(*gin.Context)) // Get request ID
	r.logger.Debug("Rolling back transaction", zap.String("operation", "postgres.DiagnosisRepository.RollbackTx"), zap.String("request_id", requestID))
	return tx.Rollback(ctx)
}
