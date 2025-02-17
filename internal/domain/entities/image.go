package entities

import (
	"github.com/google/uuid"
)

// Image represents metadata about an uploaded image (e.g., DICOM image) associated with a patient's session.
type Image struct {
	BaseEntity
	SessionID         uuid.UUID `json:"session_id,omitempty"`          // Links to the PatientSession.
	ContentID         uuid.UUID `json:"content_id,omitempty"`          // Links to UploadedContent.
	FilePath          string    `json:"file_path,omitempty"`           // Path to the (temporary) encrypted file.
	SeriesInstanceUID string    `json:"series_instance_uid,omitempty"` // Anonymized SeriesInstanceUID.
	SOPInstanceUID    string    `json:"sop_instance_uid,omitempty"`    // Anonymized SOPInstanceUID.
	ImageType         string    `json:"image_type,omitempty"`          // e.g., "CT", "CXR".
	ContentData       []byte    `json:"content_data,omitempty"`        // Any other necessary metadata (JSONB, structure defined in internal/data/models).
}
