package entities

import (
	"time"

	"github.com/google/uuid"
)

// Link represents a unique, time-limited, single-use, and expiring access link
// for a patient to upload documents and view *preliminary* results.
type Link struct {
	BaseEntity
	SessionID           uuid.UUID `json:"session_id,omitempty"`
	AccessLink          string    `json:"access_link,omitempty"`
	ExpirationTimestamp time.Time `json:"expiration_timestamp,omitempty"`
	Used                bool      `json:"used,omitempty"`
}
