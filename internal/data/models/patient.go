// internal/data/models/patient.go
package models

import (
	"time"

	"github.com/google/uuid"
)

// Patient represents a patient *session* in the system.  It does *not* store
// any persistent patient data or personally identifiable information (PII).
// All data is associated with a temporary, anonymized session.
type Patient struct {
	SessionID           uuid.UUID `json:"session_id"` // Add SessionID
	AccessLink          string    `json:"access_link"`
	ExpirationTimestamp time.Time `json:"expiration_timestamp"`
	Used                bool      `json:"used"`
	PatientData         string    `json:"patient_data"` // Optional: Patient-provided data
}

// PatientSession represents a complete patient session with all fields.
// This is primarily used for database interactions (sqlc generated code).
type PatientSession struct {
	SessionID           uuid.UUID `json:"session_id"`
	AccessLink          string    `json:"access_link"`
	ExpirationTimestamp time.Time `json:"expiration_timestamp"`
	Used                bool      `json:"used"`
	PatientData         string    `json:"patient_data"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}
