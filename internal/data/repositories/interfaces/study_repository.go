// internal/data/repositories/interfaces/study_repository.go
package interfaces

import (
	"context"

	"github.com/google/uuid"
	"github.com/stackvity/lung-server/internal/data/models"
)

// StudyRepository defines the interface for interacting with study data (DICOM Studies).
type StudyRepository interface {
	Repository // Embed the common repository interface

	// CreateStudy creates a new study record.
	CreateStudy(ctx context.Context, study *models.Study) error

	// GetStudyByID retrieves a study by its unique ID.
	GetStudyByID(ctx context.Context, studyID uuid.UUID) (*models.Study, error)

	// GetStudiesByPatientID retrieves all studies associated with a patient (session).  CORRECTED NAME
	// This is essential for report generation and data deletion.
	GetStudiesByPatientID(ctx context.Context, patientID uuid.UUID) ([]*models.Study, error) // CORRECTED

	// UpdateStudy is intentionally omitted. Study metadata is extracted from DICOM files
	// and is unlikely to change after the initial upload.

	// DeleteStudy deletes a single study record by its ID.
	DeleteStudy(ctx context.Context, studyID uuid.UUID) error

	// DeleteAllStudiesByPatientID deletes all study records associated with a patient (session).
	// This is crucial for complete data deletion.
	DeleteAllStudiesByPatientID(ctx context.Context, patientID uuid.UUID) error
}
