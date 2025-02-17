// internal/data/models/finding.go
package models

import "github.com/google/uuid"

// Finding represents a generic finding extracted from a report or image.
type Finding struct {
	FindingID        uuid.UUID `json:"finding_id" db:"finding_id"`
	FileID           uuid.UUID `json:"file_id" db:"-"`                           // Links to Report or Image
	ContentID        uuid.UUID `json:"content_id" db:"-"`                        //Link to Uploaded Content
	FindingType      string    `json:"finding_type" db:"finding_type"`           // e.g., "potential nodule", "text finding".
	Location         string    `json:"location" db:"location"`                   // Location within the lung (if applicable).
	Description      string    `json:"description" db:"description"`             // Patient-friendly description.
	ImageCoordinates []float64 `json:"image_coordinates" db:"image_coordinates"` // Optional: Image coordinates (if applicable).
	Source           string    `json:"source" db:"source"`                       // e.g., "radiology report", "pathology report", "patient input"
}
