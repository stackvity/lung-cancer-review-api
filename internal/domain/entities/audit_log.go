package entities

import (
	"github.com/google/uuid"
)

// AuditLog represents an entry in the audit log, tracking system actions and events.
type AuditLog struct {
	BaseEntity
	SessionID uuid.UUID `json:"session_id,omitempty"`
	ContentID uuid.UUID `json:"content_id,omitempty"` //Link to Upload content
	ResultID  uuid.UUID `json:"result_id,omitempty"`  //Link to Anaysis Result
	Action    string    `json:"action,omitempty"`     // Action performed (e.g., "file upload", "report generation").
	Details   []byte    `json:"details,omitempty"`    // Additional details (JSON, structure defined in internal/data/models).
}
