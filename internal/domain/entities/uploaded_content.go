package entities

import "github.com/google/uuid"

type ContentType string

const (
	ContentTypeImage   ContentType = "image"
	ContentTypeReport  ContentType = "report"
	ContentTypeLabTest ContentType = "labtest"
)

type UploadedContent struct {
	BaseEntity
	SessionID   uuid.UUID   `json:"session_id,omitempty"`
	ContentType ContentType `json:"content_type,omitempty"`
	FilePath    string      `json:"file_path,omitempty"`
	StudyData   []byte      `json:"study_data,omitempty"`   // JSONB data
	ContentData []byte      `json:"content_data,omitempty"` // JSONB data, structure depends on content_type
	Findings    []byte      `json:"findings,omitempty"`     // JSONB data,
	Nodules     []byte      `json:"nodules,omitempty"`      // JSONB
}
