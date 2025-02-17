// internal/data/repositories/interfaces/treatment_recommendation_repository.go
package interfaces

import (
	"context"

	"github.com/google/uuid"
	"github.com/stackvity/lung-server/internal/data/models"
)

// TreatmentRecommendationRepository defines the interface for *potential* treatment recommendations.
type TreatmentRecommendationRepository interface {
	Repository // Embed the common repository interface

	// CreateTreatmentRecommendation creates a new potential treatment recommendation record.
	CreateTreatmentRecommendation(ctx context.Context, recommendation *models.TreatmentRecommendation) error

	// GetTreatmentRecommendationByID retrieves a potential treatment recommendation by its unique ID.
	GetTreatmentRecommendationByID(ctx context.Context, recommendationID uuid.UUID) (*models.TreatmentRecommendation, error) // Optional

	// UpdateTreatmentRecommendation and DeleteTreatmentRecommendation are intentionally
	// omitted. These recommendations are AI-generated and should not be directly modified.
	// Deletion is handled as part of overall patient session data deletion.
	DeleteAllTreatmentRecommendationsByPatientID(ctx context.Context, patientID uuid.UUID) error // Added method
}
