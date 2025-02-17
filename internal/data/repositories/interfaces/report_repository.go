// internal/data/repositories/interfaces/report_repository.go
package interfaces

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5" // Import the pgx transaction interface
	"github.com/stackvity/lung-server/internal/data/models"
)

// ReportRepository defines the interface for interacting with report data (text-based reports).
type ReportRepository interface {
	Repository // Embed the common repository interface

	// CreateReport creates a new report record.
	CreateReport(ctx context.Context, report *models.Report) error

	// GetReportByID retrieves a report by its unique ID.
	GetReportByID(ctx context.Context, reportID uuid.UUID) (*models.Report, error)

	// UpdateReport is intentionally omitted.  Report data is derived from uploaded files
	// and the AI analysis and is unlikely to be modified directly.

	// DeleteReport deletes a single report record by its ID.
	DeleteReport(ctx context.Context, reportID uuid.UUID) error

	// GetReportByPatientID retrieves all reports associated with a patient (session).
	GetReportByPatientID(ctx context.Context, patientID uuid.UUID) ([]*models.Report, error) // Added method

	// DeleteAllReportsByPatientID deletes all report records associated with a patient,
	// as part of the complete data deletion process.
	DeleteAllReportsByPatientID(ctx context.Context, patientID uuid.UUID) error

	//CreateFinding creates new finding data to report
	CreateFinding(ctx context.Context, finding *models.Finding) error

	// BeginTx, CommitTx, and RollbackTx are explicitly defined here *and* in the
	// embedded `Repository` interface.  This redundancy is intentional.  It
	// clarifies that the ReportRepository *must* support transactions, as it will
	// likely be involved in operations that modify multiple tables.
	BeginTx(ctx context.Context, opts ...pgx.TxOptions) (pgx.Tx, error)                                               // For transactions
	CommitTx(ctx context.Context, tx pgx.Tx) error                                                                    // For transactions
	RollbackTx(ctx context.Context, tx pgx.Tx) error                                                                  // For transactions
	CreateDiagnosis(ctx context.Context, diagnosis *models.Diagnosis) error                                           // Added in BE-044
	CreateStaging(ctx context.Context, stage *models.Stage) error                                                     // Added in BE-044
	CreateTreatmentRecommendation(ctx context.Context, treatmentRecommendation *models.TreatmentRecommendation) error // Added in BE-044
}
