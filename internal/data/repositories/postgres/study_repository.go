// internal/data/repositories/postgres/study_repository.go
package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stackvity/lung-server/internal/data/models"
	"github.com/stackvity/lung-server/internal/data/repositories/interfaces"
	postgres "github.com/stackvity/lung-server/internal/data/repositories/sqlc" // Alias to avoid naming conflict
	"github.com/stackvity/lung-server/internal/domain"
	"go.uber.org/zap"
)

var _ interfaces.StudyRepository = (*StudyRepository)(nil)

// StudyRepository implements the interfaces.StudyRepository for PostgreSQL.
type StudyRepository struct {
	db      *pgxpool.Pool
	queries *postgres.Queries // Use the generated Queries struct
	logger  *zap.Logger
}

// NewStudyRepository creates a new StudyRepository instance.
// It requires a pgxpool.Pool for database access and a zap.Logger for logging.
func NewStudyRepository(db *pgxpool.Pool, logger *zap.Logger) *StudyRepository {
	return &StudyRepository{
		db:      db,
		queries: postgres.New(), // Initialize sqlc Queries
		logger:  logger,
	}
}

// CreateStudy implements interfaces.StudyRepository.
// CreateStudy creates a new study record in the database.
func (r *StudyRepository) CreateStudy(ctx context.Context, study *models.Study) error {
	const operation = "postgres.StudyRepository.CreateStudy" // Define operation for logging

	r.logger.Debug("Starting database operation", zap.String("operation", operation), zap.String("study_id", study.ID.String())) // Start log - DEBUG level

	params := &postgres.CreateStudyParams{
		ID:               pgtype.UUID{Bytes: uuid.UUID(study.ID), Valid: true},
		PatientID:        pgtype.UUID{Bytes: uuid.UUID(study.PatientID), Valid: true},
		StudyInstanceUid: study.StudyInstanceUID,
		StudyData:        study.StudyData,
	}
	_, err := r.queries.CreateStudy(ctx, r.db, params)
	if err != nil {
		r.logger.Error("Database error in CreateStudy", zap.String("operation", operation), zap.Error(err)) // Error log - ERROR level
		return fmt.Errorf("could not create study: %w", err)
	}

	r.logger.Debug("Successfully completed database operation", zap.String("operation", operation), zap.String("study_id", study.ID.String())) // End log - DEBUG level
	return nil
}

// GetStudyByID implements interfaces.StudyRepository.
// GetStudyByID retrieves a study record from the database by its unique ID.
// Returns a domain.NotFoundError if the study is not found.
func (r *StudyRepository) GetStudyByID(ctx context.Context, studyID uuid.UUID) (*models.Study, error) {
	const operation = "postgres.StudyRepository.GetStudyByID"
	r.logger.Debug("Starting database operation", zap.String("operation", operation), zap.String("study_id", studyID.String()))

	study, err := r.queries.GetStudyByID(ctx, r.db, pgtype.UUID{Bytes: uuid.UUID(studyID), Valid: true})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			r.logger.Warn("Study not found in database", zap.String("operation", operation), zap.String("study_id", studyID.String()), zap.Error(err)) // Warn level
			notFoundErr := domain.NewNotFoundError("study", studyID.String())
			notFoundErr.SetLogger(r.logger)
			return nil, notFoundErr
		}
		r.logger.Error("Database error in GetStudyByID", zap.String("operation", operation), zap.String("study_id", studyID.String()), zap.Error(err)) // Error level
		return nil, fmt.Errorf("could not get study by ID: %w", err)
	}

	modelStudy := &models.Study{
		ID:               uuid.UUID(study.ID.Bytes),
		PatientID:        uuid.UUID(study.PatientID.Bytes),
		StudyInstanceUID: study.StudyInstanceUid,
		StudyData:        study.StudyData,
		CreatedAt:        study.CreatedAt.Time,
		UpdatedAt:        study.UpdatedAt.Time,
	}

	r.logger.Debug("Successfully completed database operation", zap.String("operation", operation), zap.String("study_id", studyID.String())) // Debug level
	return modelStudy, nil
}

// GetStudiesByPatientID implements interfaces.StudyRepository.
// GetStudiesByPatientID retrieves all studies associated with a given patient ID (session ID).
func (r *StudyRepository) GetStudiesByPatientID(ctx context.Context, patientID uuid.UUID) ([]*models.Study, error) {
	const operation = "postgres.StudyRepository.GetStudiesByPatientID"
	r.logger.Debug("Starting database operation", zap.String("operation", operation), zap.String("patient_id", patientID.String()))

	studies, err := r.queries.ListStudiesByPatientID(ctx, r.db, pgtype.UUID{Bytes: uuid.UUID(patientID), Valid: true})
	if err != nil {
		r.logger.Error("Database error in GetStudiesByPatientID", zap.String("operation", operation), zap.String("patient_id", patientID.String()), zap.Error(err)) // Error level
		return nil, fmt.Errorf("could not list studies by patient ID: %w", err)
	}

	modelStudies := make([]*models.Study, len(studies))
	for i, study := range studies {
		modelStudies[i] = &models.Study{
			ID:               uuid.UUID(study.ID.Bytes),
			PatientID:        uuid.UUID(study.PatientID.Bytes),
			StudyInstanceUID: study.StudyInstanceUid,
			StudyData:        study.StudyData,
			CreatedAt:        study.CreatedAt.Time,
			UpdatedAt:        study.UpdatedAt.Time,
		}
	}

	r.logger.Debug("Successfully completed database operation", zap.String("operation", operation), zap.String("patient_id", patientID.String())) // Debug level
	return modelStudies, nil
}

// DeleteStudy implements interfaces.StudyRepository.
// DeleteStudy deletes a study record from the database by its unique ID.
func (r *StudyRepository) DeleteStudy(ctx context.Context, studyID uuid.UUID) error {
	const operation = "postgres.StudyRepository.DeleteStudy"
	r.logger.Debug("Starting database operation", zap.String("operation", operation), zap.String("study_id", studyID.String()))

	err := r.queries.DeleteStudy(ctx, r.db, pgtype.UUID{Bytes: uuid.UUID(studyID), Valid: true})
	if err != nil {
		r.logger.Error("Database error in DeleteStudy", zap.String("operation", operation), zap.String("study_id", studyID.String()), zap.Error(err)) // Error level
		return fmt.Errorf("could not delete study: %w", err)
	}

	r.logger.Debug("Successfully completed database operation", zap.String("operation", operation), zap.String("study_id", studyID.String())) // Debug level
	return nil
}

// DeleteAllStudiesByPatientID implements interfaces.StudyRepository.
// DeleteAllStudiesByPatientID deletes all study records associated with a given patient ID.
// This is used for cleanup operations when deleting patient data.
func (r *StudyRepository) DeleteAllStudiesByPatientID(ctx context.Context, patientID uuid.UUID) error {
	const operation = "postgres.StudyRepository.DeleteAllStudiesByPatientID"
	r.logger.Debug("Starting database operation", zap.String("operation", operation), zap.String("patient_id", patientID.String()))

	err := r.queries.DeleteAllStudiesByPatientID(ctx, r.db, pgtype.UUID{Bytes: uuid.UUID(patientID), Valid: true})
	if err != nil {
		r.logger.Error("Database error in DeleteAllStudiesByPatientID", zap.String("operation", operation), zap.String("patient_id", patientID.String()), zap.Error(err)) // Error level
		return fmt.Errorf("could not delete all studies by patient ID: %w", err)
	}

	r.logger.Debug("Successfully completed database operation", zap.String("operation", operation), zap.String("patient_id", patientID.String())) // Debug level
	return nil
}

// BeginTx implements interfaces.Repository.
// BeginTx starts a new database transaction.
func (r *StudyRepository) BeginTx(ctx context.Context, opts ...pgx.TxOptions) (pgx.Tx, error) {
	r.logger.Debug("Starting transaction", zap.String("operation", "postgres.StudyRepository.BeginTx")) // Debug log
	if len(opts) > 0 {
		return r.db.BeginTx(ctx, opts[0])
	}
	return r.db.Begin(ctx)
}

// CommitTx implements interfaces.Repository.
// CommitTx commits the database transaction.
func (r *StudyRepository) CommitTx(ctx context.Context, tx pgx.Tx) error {
	r.logger.Debug("Committing transaction", zap.String("operation", "postgres.StudyRepository.CommitTx")) // Debug log
	return tx.Commit(ctx)
}

// RollbackTx implements interfaces.Repository.
// RollbackTx rolls back the database transaction.
func (r *StudyRepository) RollbackTx(ctx context.Context, tx pgx.Tx) error {
	r.logger.Debug("Rolling back transaction", zap.String("operation", "postgres.StudyRepository.RollbackTx")) // Debug log
	return tx.Rollback(ctx)
}
