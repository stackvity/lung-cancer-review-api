package entities

import (
	"github.com/google/uuid"
)

// Report represents a medical report (radiology, pathology) associated with a patient's session.
type Report struct {
	BaseEntity
	SessionID  uuid.UUID `json:"session_id,omitempty"`  // Links to the PatientSession.
	ContentID  uuid.UUID `json:"content_id,omitempty"`  // Links to UploadedContent
	Filename   string    `json:"filename,omitempty"`    // Original filename.
	ReportType string    `json:"report_type,omitempty"` // e.g., "radiology", "pathology".
	ReportText string    `json:"report_text,omitempty"` // The extracted, anonymized text of the report.
	Filepath   string    `json:"filepath,omitempty"`    //Store file path to connect to the file.
}
