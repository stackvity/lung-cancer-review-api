// internal/data/repositories/postgres/stage_repository.go
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

var _ interfaces.StageRepository = (*StageRepository)(nil)

// StageRepository implements the interfaces.StageRepository for PostgreSQL.
// It handles database interactions for the Stage entity, providing methods
// to create, retrieve, and delete stage information.
type StageRepository struct {
	db      *pgxpool.Pool
	queries *postgres.Queries // Use the generated Queries struct from sqlc
	logger  *zap.Logger
}

// NewStageRepository creates a new StageRepository instance.
// It takes a pgxpool.Pool for database access and a zap.Logger for logging,
// and initializes the sqlc Queries for database interactions.
func NewStageRepository(db *pgxpool.Pool, logger *zap.Logger) *StageRepository {
	return &StageRepository{
		db:      db,
		queries: postgres.New(), // Initialize sqlc Queries
		logger:  logger,
	}
}

// CreateStaging implements interfaces.StageRepository.
// CreateStaging creates a new staging record in the database.
// It takes a context and a models.Stage object as input and returns an error if creation fails.
func (r *StageRepository) CreateStaging(ctx context.Context, stage *models.Stage) error {
	const operation = "postgres.StageRepository.CreateStaging"
	requestID := utils.GetRequestID(ctx.(*gin.Context))

	r.logger.Debug("Starting DB operation", zap.String("operation", operation), zap.String("stage_id", stage.ID.String()), zap.String("request_id", requestID))

	params := &postgres.CreateStagingParams{ // corrected struct name - using alias 'postgres'
		ResultID:   pgtype.UUID{Bytes: uuid.UUID(stage.ResultID), Valid: true},
		SessionID:  pgtype.UUID{Bytes: uuid.UUID(stage.SessionID), Valid: true},
		T:          pgtype.Text{String: stage.T, Valid: true},
		N:          pgtype.Text{String: stage.N, Valid: true},
		M:          pgtype.Text{String: stage.M, Valid: true},
		Confidence: pgtype.Text{String: stage.Confidence, Valid: true},
	}

	if r.logger.Core().Enabled(zapcore.DebugLevel) {
		r.logger.Debug("DB parameters", zap.String("operation", operation), zap.String("request_id", requestID), zap.Any("params", params))
	}

	_, err := r.queries.CreateStaging(ctx, r.db, params)
	if err != nil {
		r.logger.Error("DB error in CreateStaging", zap.String("operation", operation), zap.String("request_id", requestID), zap.Error(err))
		return utils.NewErrDBQuery("CreateStaging failed", operation, "CreateStaging", params, err) // Enhanced error wrapping
	}

	r.logger.Debug("Successfully completed DB operation", zap.String("operation", operation), zap.String("stage_id", stage.ID.String()), zap.String("request_id", requestID))
	return nil
}

// GetStageByID implements interfaces.StageRepository.
// GetStageByID retrieves a stage record from the database by its unique ID.
// It returns a populated models.Stage object if found, or a domain.NotFoundError if not.
func (r *StageRepository) GetStageByID(ctx context.Context, stageID uuid.UUID) (*models.Stage, error) {
	const operation = "postgres.StageRepository.GetStageByID"
	requestID := utils.GetRequestID(ctx.(*gin.Context))
	r.logger.Debug("Starting DB operation", zap.String("operation", operation), zap.String("stage_id", stageID.String()), zap.String("request_id", requestID))

	stage, err := r.queries.GetStageByID(ctx, r.db, pgtype.UUID{Bytes: uuid.UUID(stageID), Valid: true})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			r.logger.Warn("Stage not found in database", zap.String("operation", operation), zap.String("stage_id", stageID.String()), zap.String("request_id", requestID), zap.Error(err))
			return nil, domain.NewNotFoundError("stage", stageID.String()) // Return domain-specific NotFoundError
		}
		r.logger.Error("DB error in GetStageByID", zap.String("operation", operation), zap.String("stage_id", stageID.String()), zap.String("request_id", requestID), zap.Error(err))
		dbErr := utils.NewErrDBQuery("GetStageByID failed", operation, "GetStageByID", stageID.String(), err) // Enhanced error wrapping
		if setLoggerErr, ok := dbErr.(interface{ SetLogger(*zap.Logger) }); ok {
			setLoggerErr.SetLogger(r.logger)
		}
		return nil, dbErr
	}

	modelStage := &models.Stage{
		ID:         uuid.UUID(stage.ID.Bytes),
		ResultID:   uuid.UUID(stage.ResultID.Bytes),
		SessionID:  uuid.UUID(stage.SessionID.Bytes),
		T:          stage.T.String,
		N:          stage.N.String,
		M:          stage.M.String,
		Confidence: stage.Confidence.String,
		CreatedAt:  stage.CreatedAt.Time,
		UpdatedAt:  stage.UpdatedAt.Time,
	}

	r.logger.Debug("Successfully completed DB operation", zap.String("operation", operation), zap.String("stage_id", stageID.String()), zap.String("request_id", requestID))
	return modelStage, nil
}

// DeleteAllStagesByPatientID implements interfaces.StageRepository.
// Currently, explicit deletion of stages is not implemented here.
// It relies on the database's cascade delete functionality, configured on the
// foreign key relationship from 'analysis_result' to 'patientsession'.
// This function logs a warning to explicitly document this behavior.
func (r *StageRepository) DeleteAllStagesByPatientID(ctx context.Context, patientID uuid.UUID) error {
	const operation = "postgres.StageRepository.DeleteAllStagesByPatientID"
	requestID := utils.GetRequestID(ctx.(*gin.Context))
	r.logger.Debug("Starting DB operation", zap.String("operation", operation), zap.String("patient_id", patientID.String()), zap.String("request_id", requestID))

	// In the current database schema, stages are linked to analysis_result,
	// which in turn has a cascade delete to patient_session. Therefore,
	// deleting the patient session will automatically delete associated stages
	// due to the ON DELETE CASCADE constraint.
	// Explicit deletion of stages here is thus redundant and potentially less efficient.
	r.logger.Warn("DeleteAllStagesByPatientID: Explicit deletion of stages is not implemented in code. Deletion is implicitly handled by database cascade delete through analysis_result to patient_session.", zap.String("operation", operation), zap.String("patient_id", patientID.String()), zap.String("request_id", requestID))
	return nil // No explicit deletion action taken, relying on cascade delete.
}

// BeginTx implements interfaces.Repository.
// BeginTx starts a new database transaction. It can accept transaction options.
func (r *StageRepository) BeginTx(ctx context.Context, opts ...pgx.TxOptions) (pgx.Tx, error) {
	requestID := utils.GetRequestID(ctx.(*gin.Context)) // Get request ID
	r.logger.Debug("Starting transaction", zap.String("operation", "postgres.StageRepository.BeginTx"), zap.String("request_id", requestID))
	if len(opts) > 0 {
		return r.db.BeginTx(ctx, opts[0]) // Begin transaction with provided options
	}
	return r.db.Begin(ctx) // Begin transaction with default options
}

// CommitTx implements interfaces.Repository.
// CommitTx commits the database transaction. Returns an error if commit fails.
func (r *StageRepository) CommitTx(ctx context.Context, tx pgx.Tx) error {
	requestID := utils.GetRequestID(ctx.(*gin.Context)) // Get request ID
	r.logger.Debug("Commiting transaction", zap.String("operation", "postgres.StageRepository.CommitTx"), zap.String("request_id", requestID))
	return tx.Commit(ctx)
}

// RollbackTx implements interfaces.Repository.
// RollbackTx rolls back the database transaction. Returns an error if rollback fails.
func (r *StageRepository) RollbackTx(ctx context.Context, tx pgx.Tx) error {
	requestID := utils.GetRequestID(ctx.(*gin.Context)) // Get request ID
	r.logger.Debug("Rolling back transaction", zap.String("operation", "postgres.StageRepository.RollbackTx"), zap.String("request_id", requestID))
	return tx.Rollback(ctx)
}
