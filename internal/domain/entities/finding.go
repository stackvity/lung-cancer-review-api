package entities

import "github.com/google/uuid"

// Finding represents a generic finding (from any source - image, report, lab test).
type Finding struct {
	BaseEntity
	SessionID        uuid.UUID `json:"session_id,omitempty"`
	FileID           uuid.UUID `json:"file_id,omitempty"`           // Links to image, report, or lab test.
	ContentID        uuid.UUID `json:"content_id,omitempty"`        // Links to UploadedContent
	FindingType      string    `json:"finding_type,omitempty"`      // e.g., "potential nodule", "text finding".
	Location         string    `json:"location,omitempty"`          // Location of the finding.
	Description      string    `json:"description,omitempty"`       // Patient-friendly description.
	ImageCoordinates []float64 `json:"image_coordinates,omitempty"` // Optional: Image coordinates (if applicable).
	Source           string    `json:"source,omitempty"`            // e.g., "radiology report", "pathology report", "patient input".
}
