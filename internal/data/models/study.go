// internal/data/models/study.go
package models

import (
	"time"

	"github.com/google/uuid"
)

// Study represents a DICOM study in the database.
type Study struct {
	ID               uuid.UUID `json:"id"`
	PatientID        uuid.UUID `json:"patient_id"`         // Foreign key to PatientSession
	StudyInstanceUID string    `json:"study_instance_uid"` // Anonymized StudyInstanceUID
	StudyData        []byte    `json:"study_data"`         // JSONB data (structure defined elsewhere)
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}
