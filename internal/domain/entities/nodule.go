package entities

import (
	"github.com/google/uuid"
)

// Nodule represents a potential lung nodule detected in an image, as identified by the AI.
// Note that this is *not* a definitive diagnosis, but a *potential* finding.
type Nodule struct {
	BaseEntity
	SessionID uuid.UUID `json:"session_id,omitempty"`
	ImageID   uuid.UUID `json:"image_id,omitempty"`   // Link to the Image entity.
	ContentID uuid.UUID `json:"content_id,omitempty"` // Link to the UploadedContent entity.
	Location  string    `json:"location,omitempty"`   // Text description of location (e.g., "upper lobe of the right lung").
	Size      float64   `json:"size,omitempty"`       // Size in mm (or appropriate unit).
	Shape     string    `json:"shape,omitempty"`      // e.g., "round", "irregular".
	// Other relevant characteristics as per Gemini 2.0 output could be added here.
}
