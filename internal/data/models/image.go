// internal/data/models/image.go
package models

import (
	"time"

	"github.com/google/uuid"
)

// Image represents metadata about an uploaded image (e.g., DICOM image).
type Image struct {
	ID                uuid.UUID `json:"id" db:"content_id"`                           // Links to UploadedContent
	StudyID           uuid.UUID `json:"study_id" db:"-"`                              //FK
	FilePath          string    `json:"file_path" db:"file_path"`                     // Path to the (temporary) encrypted file.
	SeriesInstanceUID string    `json:"series_instance_uid" db:"series_instance_uid"` // Anonymized
	SOPInstanceUID    string    `json:"sop_instance_uid" db:"sop_instance_uid"`       // Anonymized
	ImageType         string    `json:"image_type" db:"image_type"`                   // e.g., "CT", "CXR".  Could become an enum if we have a fixed set.
	ContentData       []byte    `json:"content_data" db:"content_data"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`
}
