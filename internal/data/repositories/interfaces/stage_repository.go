// internal/data/repositories/interfaces/stage_repository.go
package interfaces

import (
	"context"

	"github.com/google/uuid"
	"github.com/stackvity/lung-server/internal/data/models"
)

// StageRepository defines the interface for interacting with *preliminary* staging data.
type StageRepository interface {
	Repository // Embed the common repository interface

	// CreateStaging creates a new preliminary staging record.
	CreateStaging(ctx context.Context, stage *models.Stage) error

	// GetStageByID retrieves preliminary staging information by its unique ID.
	GetStageByID(ctx context.Context, stageID uuid.UUID) (*models.Stage, error) // Optional

	// UpdateStage and DeleteStage are intentionally omitted.  Staging
	// information is AI-generated and should not be modified directly.
	// Deletion is handled as part of the overall patient session data deletion.
	DeleteAllStagesByPatientID(ctx context.Context, patientID uuid.UUID) error
}
