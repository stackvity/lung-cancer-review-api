// internal/data/repositories/postgres/image_repository.go
package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stackvity/lung-server/internal/data/models"
	"github.com/stackvity/lung-server/internal/data/repositories/interfaces"
	postgres "github.com/stackvity/lung-server/internal/data/repositories/sqlc" // Alias
	"github.com/stackvity/lung-server/internal/domain"
	"github.com/stackvity/lung-server/internal/utils" // Import utils for errors
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var _ interfaces.ImageRepository = (*ImageRepository)(nil)

// ImageRepository implements the interfaces.ImageRepository for PostgreSQL.
type ImageRepository struct {
	db      *pgxpool.Pool
	queries *postgres.Queries
	logger  *zap.Logger
}

// NewImageRepository creates a new ImageRepository instance.
func NewImageRepository(db *pgxpool.Pool, logger *zap.Logger) *ImageRepository {
	return &ImageRepository{
		db:      db,
		queries: postgres.New(),
		logger:  logger,
	}
}

// CreateImage implements interfaces.ImageRepository.
func (r *ImageRepository) CreateImage(ctx context.Context, image *models.Image) error {
	const operation = "postgres.ImageRepository.CreateImage"
	requestID := utils.GetRequestID(ctx.(*gin.Context))

	r.logger.Debug("Starting DB operation", zap.String("operation", operation), zap.String("image_id", image.ID.String()), zap.String("request_id", requestID))

	params := &postgres.CreateImageParams{
		ID:                pgtype.UUID{Bytes: uuid.UUID(image.ID), Valid: true},
		StudyID:           pgtype.UUID{Bytes: uuid.UUID(image.StudyID), Valid: true},
		FilePath:          image.FilePath,
		SeriesInstanceUid: image.SeriesInstanceUID,
		SopInstanceUid:    image.SOPInstanceUID,
		ImageType:         image.ImageType,
		ContentData:       image.ContentData,
	}

	if r.logger.Core().Enabled(zapcore.DebugLevel) {
		r.logger.Debug("DB parameters", zap.String("operation", operation), zap.String("request_id", requestID), zap.Any("params", params))
	}

	_, err := r.queries.CreateImage(ctx, r.db, params)
	if err != nil {
		r.logger.Error("DB error in CreateImage", zap.String("operation", operation), zap.String("request_id", requestID), zap.Error(err))
		return utils.NewErrDBQuery("CreateImage failed", operation, "CreateImage", params, err) // Enhanced error
	}
	r.logger.Debug("Successfully completed DB operation", zap.String("operation", operation), zap.String("image_id", image.ID.String()), zap.String("request_id", requestID))
	return nil
}

// GetImageByID implements interfaces.ImageRepository.
func (r *ImageRepository) GetImageByID(ctx context.Context, imageID uuid.UUID) (*models.Image, error) {
	const operation = "postgres.ImageRepository.GetImageByID"
	requestID := utils.GetRequestID(ctx.(*gin.Context))
	r.logger.Debug("Starting DB operation", zap.String("operation", operation), zap.String("image_id", imageID.String()), zap.String("request_id", requestID))

	image, err := r.queries.GetImageByID(ctx, r.db, pgtype.UUID{Bytes: uuid.UUID(imageID), Valid: true})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			r.logger.Warn("Image not found", zap.String("operation", operation), zap.String("image_id", imageID.String()), zap.String("request_id", requestID), zap.Error(err))
			return nil, domain.NewNotFoundError("image", imageID.String())
		}
		r.logger.Error("DB error in GetImageByID", zap.String("operation", operation), zap.String("image_id", imageID.String()), zap.String("request_id", requestID), zap.Error(err))
		return nil, utils.NewErrDBQuery("GetImageByID failed", operation, "GetImageByID", imageID, err) // Enhanced error
	}

	modelImage := &models.Image{
		ID:                uuid.UUID(image.ID.Bytes),
		StudyID:           uuid.UUID(image.StudyID.Bytes),
		FilePath:          image.FilePath,
		SeriesInstanceUID: image.SeriesInstanceUid,
		SOPInstanceUID:    image.SopInstanceUid,
		ImageType:         image.ImageType,
		ContentData:       image.ContentData,
		CreatedAt:         image.CreatedAt.Time,
		UpdatedAt:         image.UpdatedAt.Time,
	}

	r.logger.Debug("Successfully completed DB operation", zap.String("operation", operation), zap.String("image_id", imageID.String()), zap.String("request_id", requestID))
	return modelImage, nil
}

// GetImageByStudyID implements interfaces.ImageRepository.
func (r *ImageRepository) GetImageByStudyID(ctx context.Context, studyID uuid.UUID) ([]*models.Image, error) {
	const operation = "postgres.ImageRepository.GetImageByStudyID"
	requestID := utils.GetRequestID(ctx.(*gin.Context))
	r.logger.Debug("Starting DB operation", zap.String("operation", operation), zap.String("study_id", studyID.String()), zap.String("request_id", requestID))

	images, err := r.queries.GetImageByStudyID(ctx, r.db, pgtype.UUID{Bytes: uuid.UUID(studyID), Valid: true})
	if err != nil {
		r.logger.Error("DB error in GetImageByStudyID", zap.String("operation", operation), zap.String("study_id", studyID.String()), zap.String("request_id", requestID), zap.Error(err))
		return nil, utils.NewErrDBQuery("GetImageByStudyID failed", operation, "GetImageByStudyID", studyID, err) // Enhanced error
	}

	modelImages := make([]*models.Image, len(images))
	for i, image := range images {
		modelImages[i] = &models.Image{
			ID:                uuid.UUID(image.ID.Bytes),
			StudyID:           uuid.UUID(image.StudyID.Bytes),
			FilePath:          image.FilePath,
			SeriesInstanceUID: image.SeriesInstanceUid,
			SOPInstanceUID:    image.SopInstanceUid,
			ImageType:         image.ImageType,
			ContentData:       image.ContentData,
			CreatedAt:         image.CreatedAt.Time,
			UpdatedAt:         image.UpdatedAt.Time,
		}
	}

	r.logger.Debug("Successfully completed DB operation", zap.String("operation", operation), zap.String("study_id", studyID.String()), zap.String("request_id", requestID))
	return modelImages, nil
}

// GetImageByPatientID implements interfaces.ImageRepository.
func (r *ImageRepository) GetImageByPatientID(ctx context.Context, patientID uuid.UUID) ([]*models.Image, error) {
	const operation = "postgres.ImageRepository.GetImageByPatientID"
	requestID := utils.GetRequestID(ctx.(*gin.Context))
	r.logger.Debug("Starting DB operation", zap.String("operation", operation), zap.String("patient_id", patientID.String()), zap.String("request_id", requestID))

	images, err := r.queries.ListImagesByPatientID(ctx, r.db, pgtype.UUID{Bytes: uuid.UUID(patientID), Valid: true})
	if err != nil {
		r.logger.Error("DB error in GetImageByPatientID", zap.String("operation", operation), zap.String("patient_id", patientID.String()), zap.String("request_id", requestID), zap.Error(err))
		return nil, utils.NewErrDBQuery("ListImagesByPatientID failed", operation, "ListImagesByPatientID", patientID, err) // Enhanced error
	}

	modelImages := make([]*models.Image, len(images))
	for i, image := range images {
		modelImages[i] = &models.Image{
			ID:                uuid.UUID(image.ID.Bytes),
			StudyID:           uuid.UUID(image.StudyID.Bytes),
			FilePath:          image.FilePath,
			SeriesInstanceUID: image.SeriesInstanceUid,
			SOPInstanceUID:    image.SopInstanceUid,
			ImageType:         image.ImageType,
			ContentData:       image.ContentData,
			CreatedAt:         image.CreatedAt.Time,
			UpdatedAt:         image.UpdatedAt.Time,
		}
	}

	r.logger.Debug("Successfully completed DB operation", zap.String("operation", operation), zap.String("patient_id", patientID.String()), zap.String("request_id", requestID))
	return modelImages, nil
}

// DeleteImage implements interfaces.ImageRepository.
func (r *ImageRepository) DeleteImage(ctx context.Context, imageID uuid.UUID) error {
	const operation = "postgres.ImageRepository.DeleteImage"
	requestID := utils.GetRequestID(ctx.(*gin.Context))
	r.logger.Debug("Starting DB operation", zap.String("operation", operation), zap.String("image_id", imageID.String()), zap.String("request_id", requestID))

	err := r.queries.DeleteImage(ctx, r.db, pgtype.UUID{Bytes: uuid.UUID(imageID), Valid: true})
	if err != nil {
		r.logger.Error("DB error in DeleteImage", zap.String("operation", operation), zap.String("image_id", imageID.String()), zap.String("request_id", requestID), zap.Error(err))
		return utils.NewErrDBQuery("DeleteImage failed", operation, "DeleteImage", imageID.String(), err) // Enhanced error
	}

	r.logger.Debug("Successfully completed DB operation", zap.String("operation", operation), zap.String("image_id", imageID.String()), zap.String("request_id", requestID))
	return nil
}

// DeleteAllImagesByPatientID implements interfaces.ImageRepository.
func (r *ImageRepository) DeleteAllImagesByPatientID(ctx context.Context, patientID uuid.UUID) error {
	const operation = "postgres.ImageRepository.DeleteAllImagesByPatientID"
	requestID := utils.GetRequestID(ctx.(*gin.Context))
	r.logger.Debug("Starting DB operation", zap.String("operation", operation), zap.String("patient_id", patientID.String()), zap.String("request_id", requestID))

	err := r.queries.DeleteAllImagesByPatientID(ctx, r.db, pgtype.UUID{Bytes: uuid.UUID(patientID), Valid: true})
	if err != nil {
		r.logger.Error("DB error in DeleteAllImagesByPatientID", zap.String("operation", operation), zap.String("patient_id", patientID.String()), zap.String("request_id", requestID), zap.Error(err))
		return utils.NewErrDBQuery("DeleteAllImagesByPatientID failed", operation, "DeleteAllImagesByPatientID", patientID.String(), err) // Enhanced error
	}

	r.logger.Debug("Successfully completed DB operation", zap.String("operation", operation), zap.String("patient_id", patientID.String()), zap.String("request_id", requestID))
	return nil
}

// CreateNodule implements interfaces.ImageRepository.
func (r *ImageRepository) CreateNodule(ctx context.Context, nodule *models.Nodule) error {
	// Implementation for CreateNodule (if needed, based on your design)
	// This might involve direct SQL queries or using sqlc generated code if you extend queries.sql
	const operation = "postgres.ImageRepository.CreateNodule"
	requestID := utils.GetRequestID(ctx.(*gin.Context))
	r.logger.Warn("Operation not implemented", zap.String("operation", operation), zap.String("request_id", requestID)) // Use Warn level as it's not critical yet
	return fmt.Errorf("%s: not implemented", operation)                                                                 // Return a "not implemented" error
}

// BeginTx implements interfaces.Repository.
func (r *ImageRepository) BeginTx(ctx context.Context, opts ...pgx.TxOptions) (pgx.Tx, error) {
	r.logger.Debug("Starting transaction", zap.String("operation", "postgres.ImageRepository.BeginTx"), zap.String("request_id", utils.GetRequestID(ctx.(*gin.Context))))
	if len(opts) > 0 {
		return r.db.BeginTx(ctx, opts[0])
	}
	return r.db.Begin(ctx)
}

// CommitTx implements interfaces.Repository.
func (r *ImageRepository) CommitTx(ctx context.Context, tx pgx.Tx) error {
	r.logger.Debug("Committing transaction", zap.String("operation", "postgres.ImageRepository.CommitTx"), zap.String("request_id", utils.GetRequestID(ctx.(*gin.Context))))
	return tx.Commit(ctx)
}

// RollbackTx implements interfaces.Repository.
func (r *ImageRepository) RollbackTx(ctx context.Context, tx pgx.Tx) error {
	r.logger.Debug("Rolling back transaction", zap.String("operation", "postgres.ImageRepository.RollbackTx"), zap.String("request_id", utils.GetRequestID(ctx.(*gin.Context))))
	return tx.Rollback(ctx)
}
