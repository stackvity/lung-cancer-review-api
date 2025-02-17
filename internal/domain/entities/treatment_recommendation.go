package entities

import "github.com/google/uuid"

// TreatmentRecommendation represents a *potential* treatment option identified by the AI system.
// These are *suggestions only* for discussion with a physician and *do not* constitute medical advice.
type TreatmentRecommendation struct {
	BaseEntity
	SessionID       uuid.UUID `json:"session_id,omitempty"`
	DiagnosisID     uuid.UUID `json:"diagnosis_id,omitempty"` // Link to the preliminary diagnosis.
	TreatmentOption string    `json:"treatment_option,omitempty"`
	Rationale       string    `json:"rationale,omitempty"`    // Rationale for this option (plain language).
	Benefits        string    `json:"benefits,omitempty"`     // Potential benefits (plain language).
	Risks           string    `json:"risks,omitempty"`        // Potential risks (plain language).
	SideEffects     string    `json:"side_effects,omitempty"` // Potential side effects (plain language).
	Confidence      string    `json:"confidence,omitempty"`
}
