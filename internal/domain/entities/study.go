package entities

import (
	"github.com/google/uuid"
)

// Study represents a medical study (e.g., CT scan, X-ray) associated with a patient's session.
type Study struct {
	BaseEntity
	SessionID        uuid.UUID `json:"session_id,omitempty"`         // Links to the PatientSession and UploadedContent
	StudyInstanceUID string    `json:"study_instance_uid,omitempty"` // Anonymized StudyInstanceUID.
	StudyData        []byte    `json:"study_data,omitempty"`         // JSONB data (structure defined in internal/data/models).
}
