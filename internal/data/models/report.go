// internal/data/models/report.go
package models

import (
	"time"

	"github.com/google/uuid"
)

// Report represents a medical report (radiology, pathology) in the system.
type Report struct {
	ID         uuid.UUID `json:"id" db:"content_id"`           // Corresponds to UploadedContent.content_id
	PatientID  uuid.UUID `json:"patient_id" db:"-"`            //FK
	Filename   string    `json:"filename" db:"filename"`       // Original filename
	ReportType string    `json:"report_type" db:"report_type"` // e.g., "radiology", "pathology".
	ReportText string    `json:"report_text" db:"report_text"` // The extracted, *anonymized* text.
	Filepath   string    `json:"filepath" db:"file_path"`      //Store file path to connect to the file.
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}
