package entities

import (
	"github.com/google/uuid"
)

// Patient represents a patient's session in the system.  It does *not* store
// any persistent patient data or personally identifiable information (PII).
// All data is associated with a temporary, anonymized session.
type Patient struct {
	BaseEntity
	SessionID   uuid.UUID `json:"session_id,omitempty"`   // Links to the PatientSession.
	PatientData []byte    `json:"patient_data,omitempty"` // Optional: Patient-provided data (JSONB).  Structure defined in internal/data/models.
}
