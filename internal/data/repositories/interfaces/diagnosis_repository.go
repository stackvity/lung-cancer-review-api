// internal/data/repositories/interfaces/diagnosis_repository.go
package interfaces

import (
	"context"

	"github.com/google/uuid"
	"github.com/stackvity/lung-server/internal/data/models"
)

// DiagnosisRepository defines the interface for interacting with *preliminary* diagnosis data.
type DiagnosisRepository interface {
	Repository // Embed the common repository interface

	// CreateDiagnosis creates a new preliminary diagnosis record.
	CreateDiagnosis(ctx context.Context, diagnosis *models.Diagnosis) error

	// GetDiagnosisByID retrieves a preliminary diagnosis by its unique ID.
	GetDiagnosisByID(ctx context.Context, diagnosisID uuid.UUID) (*models.Diagnosis, error) // Optional

	// UpdateDiagnosis and DeleteDiagnosis are intentionally omitted.  The
	// preliminary diagnosis is AI-generated and should not be modified directly.
	// Deletion is handled as part of the overall patient session data deletion.

	// You might have GetDiagnosisByPatientID or GetDiagnosisBySessionID
	// depending on your access patterns, but these are likely redundant given
	// the 1:1 relationship between a patient session and an analysis result.
}
