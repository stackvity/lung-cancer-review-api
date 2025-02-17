// internal/data/repositories/interfaces/image_repository.go
package interfaces

import (
	"context"

	"github.com/google/uuid"
	"github.com/stackvity/lung-server/internal/data/models"
)

// ImageRepository defines the interface for interacting with image metadata (DICOM Instances).
type ImageRepository interface {
	Repository // Embed the common repository interface

	// CreateImage creates a new image record.
	CreateImage(ctx context.Context, image *models.Image) error

	// GetImageByID retrieves an image by its unique ID.
	GetImageByID(ctx context.Context, imageID uuid.UUID) (*models.Image, error)

	// GetImageByStudyID retrieves all images associated with a given study ID. // ADD THIS
	GetImageByStudyID(ctx context.Context, studyID uuid.UUID) ([]*models.Image, error) // ADD THIS

	// GetImageByPatientID retrieves all images associated with a patient (session).
	// This is essential for data deletion.
	GetImageByPatientID(ctx context.Context, patientID uuid.UUID) ([]*models.Image, error)
	// UpdateImage is intentionally omitted.  Image metadata is unlikely to change.
	// DeleteImage deletes a single image record by ID.
	DeleteImage(ctx context.Context, imageID uuid.UUID) error

	// DeleteAllImagesByPatientID deletes all image records associated with a given patient ID.
	DeleteAllImagesByPatientID(ctx context.Context, patientID uuid.UUID) error

	// CreateNodule creates a new nodule record associated with an image. (BE-031)
	CreateNodule(ctx context.Context, nodule *models.Nodule) error
}
