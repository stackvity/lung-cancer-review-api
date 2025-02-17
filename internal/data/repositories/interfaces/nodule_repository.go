// internal/data/repositories/interfaces/nodule_repository.go
package interfaces

import (
	"context"

	"github.com/google/uuid"
	"github.com/stackvity/lung-server/internal/data/models"
)

type NoduleRepository interface {
	Repository // Embed the common repository interface

	// CreateNodule creates a new nodule record associated with an image.
	CreateNodule(ctx context.Context, nodule *models.Nodule) error

	// GetNoduleByID retrieves a nodule by its unique ID.  This is optional;
	// it depends on whether you'll need to access individual nodules directly
	// outside the context of a patient session or image.
	GetNoduleByID(ctx context.Context, noduleID uuid.UUID) (*models.Nodule, error) // Optional

	// UpdateNodule is intentionally omitted. Nodule data is derived from the
	// AI analysis and is unlikely to be modified directly.

	// DeleteNodule is *not* included here.  Nodules are deleted as part of the
	// image deletion process (via DeleteAllImagesByPatientID in ImageRepository).
	// This avoids inconsistencies and ensures that all related data is removed.

	// You might have GetNodulesByImageID or GetNodulesByStudyID
	// depending on your access patterns.  Since the primary flow is
	// processing a session and generating a report, these aren't
	// strictly necessary in the initial scope, but might be useful later.
}
