// internal/data/models/nodule.go
package models

import (
	"github.com/google/uuid"
)

// Nodule represents a potential lung nodule detected in an image.
type Nodule struct {
	ID       uuid.UUID `json:"id" db:"finding_id"`          // Corrected db tag to "finding_id" to match queries.sql.
	ImageID  uuid.UUID `json:"image_id" db:"file_id"`       // Foreign key to Image, corrected db tag to "file_id"
	Location string    `json:"location" db:"description"`   // Corrected db tag to "description" - maps to finding description.
	Size     float64   `json:"size" db:"image_coordinates"` // Corrected db tag to "image_coordinates", assuming size is derived from coordinates.
	Shape    string    `json:"shape" db:"source"`           // Corrected db tag to "source" -  maps to finding source.
}
