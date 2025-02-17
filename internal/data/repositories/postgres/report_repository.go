// internal/data/repositories/postgres/report_repository.go
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

var _ interfaces.ReportRepository = (*ReportRepository)(nil)

// ReportRepository implements the interfaces.ReportRepository for PostgreSQL.
type ReportRepository struct {
	db      *pgxpool.Pool
	queries *postgres.Queries // Use the generated Queries struct
	logger  *zap.Logger
}

// NewReportRepository creates a new ReportRepository instance.
func NewReportRepository(db *pgxpool.Pool, logger *zap.Logger) *ReportRepository {
	return &ReportRepository{
		db:      db,
		queries: postgres.New(), // Initialize sqlc Queries
		logger:  logger,
	}
}

// CreateReport implements interfaces.ReportRepository.
func (r *ReportRepository) CreateReport(ctx context.Context, report *models.Report) error {
	const operation = "postgres.ReportRepository.CreateReport"
	requestID := utils.GetRequestID(ctx.(*gin.Context))

	r.logger.Debug("Starting DB operation", zap.String("operation", operation), zap.String("report_id", report.ID.String()), zap.String("request_id", requestID))

	params := &postgres.CreateReportParams{
		ID:         pgtype.UUID{Bytes: uuid.UUID(report.ID), Valid: true},
		PatientID:  pgtype.UUID{Bytes: uuid.UUID(report.PatientID), Valid: true},
		Filename:   report.Filename,
		ReportType: postgres.ReportType(report.ReportType), // Assuming ReportType is an enum in sqlc/models
		ReportText: pgtype.Text{String: report.ReportText, Valid: true},
		Filepath:   report.Filepath,
	}

	if r.logger.Core().Enabled(zapcore.DebugLevel) {
		r.logger.Debug("DB parameters", zap.String("operation", operation), zap.String("request_id", requestID), zap.Any("params", params))
	}

	_, err := r.queries.CreateReport(ctx, r.db, params)
	if err != nil {
		r.logger.Error("DB error in CreateReport", zap.String("operation", operation), zap.String("request_id", requestID), zap.Error(err))
		return utils.NewErrDBQuery("CreateReport failed", operation, "CreateReport", params, err) // Enhanced error
	}

	r.logger.Debug("Successfully completed DB operation", zap.String("operation", operation), zap.String("report_id", report.ID.String()), zap.String("request_id", requestID))
	return nil
}

// GetReportByID implements interfaces.ReportRepository.
func (r *ReportRepository) GetReportByID(ctx context.Context, reportID uuid.UUID) (*models.Report, error) {
	const operation = "postgres.ReportRepository.GetReportByID"
	requestID := utils.GetRequestID(ctx.(*gin.Context))
	r.logger.Debug("Starting DB operation", zap.String("operation", operation), zap.String("report_id", reportID.String()), zap.String("request_id", requestID))

	report, err := r.queries.GetReportByID(ctx, r.db, pgtype.UUID{Bytes: uuid.UUID(reportID), Valid: true})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			r.logger.Warn("Report not found", zap.String("operation", operation), zap.String("report_id", reportID.String()), zap.String("request_id", requestID), zap.Error(err))
			return nil, domain.NewNotFoundError("report", reportID.String())
		}
		r.logger.Error("DB error in GetReportByID", zap.String("operation", operation), zap.String("report_id", reportID.String()), zap.String("request_id", requestID), zap.Error(err))
		return nil, utils.NewErrDBQuery("GetReportByID failed", operation, "GetReportByID", reportID.String(), err) // Enhanced error
	}

	modelReport := &models.Report{
		ID:         uuid.UUID(report.ID.Bytes),
		PatientID:  uuid.UUID(report.PatientID.Bytes),
		Filename:   report.Filename,
		ReportType: string(report.ReportType), // Enum conversion to string
		ReportText: report.ReportText.String,
		Filepath:   report.Filepath,
		CreatedAt:  report.CreatedAt.Time,
		UpdatedAt:  report.UpdatedAt.Time,
	}

	r.logger.Debug("Successfully completed DB operation", zap.String("operation", operation), zap.String("report_id", reportID.String()), zap.String("request_id", requestID))
	return modelReport, nil
}

// GetReportByPatientID implements interfaces.ReportRepository.
func (r *ReportRepository) GetReportByPatientID(ctx context.Context, patientID uuid.UUID) ([]*models.Report, error) {
	const operation = "postgres.ReportRepository.GetReportByPatientID"
	requestID := utils.GetRequestID(ctx.(*gin.Context))
	r.logger.Debug("Starting DB operation", zap.String("operation", operation), zap.String("patient_id", patientID.String()), zap.String("request_id", requestID))

	reports, err := r.queries.GetReportByPatientID(ctx, r.db, pgtype.UUID{Bytes: uuid.UUID(patientID), Valid: true})
	if err != nil {
		r.logger.Error("DB error in GetReportByPatientID", zap.String("operation", operation), zap.String("patient_id", patientID.String()), zap.String("request_id", requestID), zap.Error(err))
		return nil, utils.NewErrDBQuery("GetReportByPatientID failed", operation, "GetReportByPatientID", patientID, err) // Enhanced error
	}

	modelReports := make([]*models.Report, len(reports))
	for i, report := range reports {
		modelReports[i] = &models.Report{
			ID:         uuid.UUID(report.ID.Bytes),
			PatientID:  uuid.UUID(report.PatientID.Bytes),
			Filename:   report.Filename,
			ReportType: string(report.ReportType), // Enum conversion to string
			ReportText: report.ReportText.String,
			Filepath:   report.Filepath,
			CreatedAt:  report.CreatedAt.Time,
			UpdatedAt:  report.UpdatedAt.Time,
		}
	}

	r.logger.Debug("Successfully completed DB operation", zap.String("operation", operation), zap.String("patient_id", patientID.String()), zap.String("request_id", requestID))
	return modelReports, nil
}

// DeleteReport implements interfaces.ReportRepository.
func (r *ReportRepository) DeleteReport(ctx context.Context, reportID uuid.UUID) error {
	const operation = "postgres.ReportRepository.DeleteReport"
	requestID := utils.GetRequestID(ctx.(*gin.Context))
	r.logger.Debug("Starting DB operation", zap.String("operation", operation), zap.String("report_id", reportID.String()), zap.String("request_id", requestID))

	err := r.queries.DeleteReport(ctx, r.db, pgtype.UUID{Bytes: uuid.UUID(reportID), Valid: true})
	if err != nil {
		r.logger.Error("DB error in DeleteReport", zap.String("operation", operation), zap.String("report_id", reportID.String()), zap.String("request_id", requestID), zap.Error(err))
		return utils.NewErrDBQuery("DeleteReport failed", operation, "DeleteReport", reportID.String(), err) // Enhanced error
	}

	r.logger.Debug("Successfully completed DB operation", zap.String("operation", operation), zap.String("report_id", reportID.String()), zap.String("request_id", requestID))
	return err
}

// DeleteAllReportsByPatientID implements interfaces.ReportRepository.
func (r *ReportRepository) DeleteAllReportsByPatientID(ctx context.Context, patientID uuid.UUID) error {
	const operation = "postgres.ReportRepository.DeleteAllReportsByPatientID"
	requestID := utils.GetRequestID(ctx.(*gin.Context))
	r.logger.Debug("Starting DB operation", zap.String("operation", operation), zap.String("patient_id", patientID.String()), zap.String("request_id", requestID))

	err := r.queries.DeleteAllReportsByPatientID(ctx, r.db, pgtype.UUID{Bytes: uuid.UUID(patientID), Valid: true})
	if err != nil {
		r.logger.Error("DB error in DeleteAllReportsByPatientID", zap.String("operation", operation), zap.String("patient_id", patientID.String()), zap.String("request_id", requestID), zap.Error(err))
		return utils.NewErrDBQuery("DeleteAllReportsByPatientID failed", operation, "DeleteAllReportsByPatientID", patientID.String(), err) // Enhanced error
	}

	r.logger.Debug("Successfully completed DB operation", zap.String("operation", operation), zap.String("patient_id", patientID.String()), zap.String("request_id", requestID))
	return nil
}

// CreateFinding implements interfaces.ReportRepository.
func (r *ReportRepository) CreateFinding(ctx context.Context, finding *models.Finding) error {
	const operation = "postgres.ReportRepository.CreateFinding"
	requestID := utils.GetRequestID(ctx.(*gin.Context))

	r.logger.Debug("Starting DB operation", zap.String("operation", operation), zap.String("finding_id", finding.FindingID.String()), zap.String("request_id", requestID))

	params := &postgres.CreateFindingParams{
		FindingID:        pgtype.UUID{Bytes: uuid.UUID(finding.FindingID), Valid: true},
		FileID:           pgtype.UUID{Bytes: uuid.UUID(finding.FileID), Valid: true},
		FindingType:      finding.FindingType,
		Description:      finding.Description,
		ImageCoordinates: finding.ImageCoordinates,
		Source:           finding.Source,
	}

	if r.logger.Core().Enabled(zapcore.DebugLevel) {
		r.logger.Debug("DB parameters", zap.String("operation", operation), zap.String("request_id", requestID), zap.Any("params", params))
	}

	_, err := r.queries.CreateFinding(ctx, r.db, params)
	if err != nil {
		r.logger.Error("DB error in CreateFinding", zap.String("operation", operation), zap.String("request_id", requestID), zap.Error(err))
		return utils.NewErrDBQuery("CreateFinding failed", operation, "CreateFinding", params, err) // Enhanced error
	}

	r.logger.Debug("Successfully completed DB operation", zap.String("operation", operation), zap.String("finding_id", finding.FindingID.String()), zap.String("request_id", requestID))
	return nil
}

// CreateDiagnosis implements interfaces.ReportRepository.
func (r *ReportRepository) CreateDiagnosis(ctx context.Context, diagnosis *models.Diagnosis) error {
	const operation = "postgres.ReportRepository.CreateDiagnosis"
	requestID := utils.GetRequestID(ctx.(*gin.Context))

	r.logger.Debug("Starting DB operation", zap.String("operation", operation), zap.String("diagnosis_id", diagnosis.ID.String()), zap.String("request_id", requestID))

	params := &postgres.CreateDiagnosisParams{ // Corrected struct name - was the cause of the error.
		// ID:            pgtype.UUID{Bytes: uuid.UUID(diagnosis.ID), Valid: true},
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

// CreateStaging implements interfaces.ReportRepository.
func (r *ReportRepository) CreateStaging(ctx context.Context, stage *models.Stage) error {
	const operation = "postgres.ReportRepository.CreateStaging"
	requestID := utils.GetRequestID(ctx.(*gin.Context))

	r.logger.Debug("Starting DB operation", zap.String("operation", operation), zap.String("stage_id", stage.ID.String()), zap.String("request_id", requestID))

	params := &postgres.CreateStagingParams{ // corrected struct name
		// ID:         pgtype.UUID{Bytes: uuid.UUID(stage.ID), Valid: true},
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
		return utils.NewErrDBQuery("CreateStaging failed", operation, "CreateStaging", params, err) // Enhanced error
	}

	r.logger.Debug("Successfully completed DB operation", zap.String("operation", operation), zap.String("stage_id", stage.ID.String()), zap.String("request_id", requestID))
	return nil
}

// CreateTreatmentRecommendation implements interfaces.ReportRepository.
func (r *ReportRepository) CreateTreatmentRecommendation(ctx context.Context, treatmentRecommendation *models.TreatmentRecommendation) error {
	const operation = "postgres.ReportRepository.CreateTreatmentRecommendation"
	requestID := utils.GetRequestID(ctx.(*gin.Context))

	r.logger.Debug("Starting DB operation", zap.String("operation", operation), zap.String("treatment_recommendation_id", treatmentRecommendation.ID.String()), zap.String("request_id", requestID))

	params := &postgres.CreateTreatmentRecommendationParams{ // corrected struct name
		// ID:              pgtype.UUID{Bytes: uuid.UUID(treatmentRecommendation.ID), Valid: true},
		ResultID:        pgtype.UUID{Bytes: uuid.UUID(treatmentRecommendation.ResultID), Valid: true},
		SessionID:       pgtype.UUID{Bytes: uuid.UUID(treatmentRecommendation.SessionID), Valid: true},
		DiagnosisID:     pgtype.UUID{Bytes: uuid.UUID(treatmentRecommendation.DiagnosisID), Valid: true},
		TreatmentOption: pgtype.Text{String: treatmentRecommendation.TreatmentOption, Valid: true},
		Rationale:       pgtype.Text{String: treatmentRecommendation.Rationale, Valid: true},
		Benefits:        pgtype.Text{String: treatmentRecommendation.Benefits, Valid: true},
		Risks:           pgtype.Text{String: treatmentRecommendation.Risks, Valid: true},
		SideEffects:     pgtype.Text{String: treatmentRecommendation.SideEffects, Valid: true},
		Confidence:      pgtype.Text{String: treatmentRecommendation.Confidence, Valid: true},
	}

	if r.logger.Core().Enabled(zapcore.DebugLevel) {
		r.logger.Debug("DB parameters", zap.String("operation", operation), zap.String("request_id", requestID), zap.Any("params", params))
	}

	_, err := r.queries.CreateTreatmentRecommendation(ctx, r.db, params)
	if err != nil {
		r.logger.Error("DB error in CreateTreatmentRecommendation", zap.String("operation", operation), zap.String("request_id", requestID), zap.Error(err))
		return utils.NewErrDBQuery("CreateTreatmentRecommendation failed", operation, "CreateTreatmentRecommendation", params, err) // Enhanced error
	}

	r.logger.Debug("Successfully completed DB operation", zap.String("operation", operation), zap.String("treatment_recommendation_id", treatmentRecommendation.ID.String()), zap.String("request_id", requestID))
	return nil
}

// BeginTx implements interfaces.Repository.
func (r *ReportRepository) BeginTx(ctx context.Context, opts ...pgx.TxOptions) (pgx.Tx, error) {
	r.logger.Debug("Starting transaction", zap.String("operation", "postgres.ReportRepository.BeginTx"), zap.String("request_id", utils.GetRequestID(ctx.(*gin.Context))))
	if len(opts) > 0 {
		return r.db.BeginTx(ctx, opts[0])
	}
	return r.db.Begin(ctx)
}

// CommitTx implements interfaces.Repository.
func (r *ReportRepository) CommitTx(ctx context.Context, tx pgx.Tx) error {
	r.logger.Debug("Committing transaction", zap.String("operation", "postgres.ReportRepository.CommitTx"), zap.String("request_id", utils.GetRequestID(ctx.(*gin.Context))))
	return tx.Commit(ctx)
}

// RollbackTx implements interfaces.Repository.
func (r *ReportRepository) RollbackTx(ctx context.Context, tx pgx.Tx) error {
	r.logger.Debug("Rolling back transaction", zap.String("operation", "postgres.ReportRepository.RollbackTx"), zap.String("request_id", utils.GetRequestID(ctx.(*gin.Context))))
	return tx.Rollback(ctx)
}
