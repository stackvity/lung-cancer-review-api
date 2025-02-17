// internal/data/models/treatment_recommendation.go
package models

import (
	"time"

	"github.com/google/uuid"
)

// TreatmentRecommendation represents a *potential* treatment option.
type TreatmentRecommendation struct {
	ID              uuid.UUID `json:"id" db:"result_id"`
	ResultID        uuid.UUID `json:"result_id" db:"result_id"`   // Corrected: Added ResultID, removed PatientID
	SessionID       uuid.UUID `json:"session_id" db:"session_id"` // Corrected: Added SessionID, removed PatientID
	DiagnosisID     uuid.UUID `json:"diagnosis_id" db:"diagnosis_id"`
	TreatmentOption string    `json:"treatment_option" db:"treatment_recommendations"` //  treatment option.
	Rationale       string    `json:"rationale" db:"rationale"`                        // Rationale (plain language).
	Benefits        string    `json:"benefits" db:"benefits"`                          // Potential benefits (plain language).
	Risks           string    `json:"risks" db:"risks"`                                // Potential risks (plain language).
	SideEffects     string    `json:"side_effects" db:"side_effects"`                  // Potential side effects (plain language).
	Confidence      string    `json:"confidence" db:"confidence"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}
