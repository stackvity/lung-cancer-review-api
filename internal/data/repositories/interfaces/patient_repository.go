// internal/data/repositories/interfaces/patient_repository.go
package interfaces

import (
	"context"

	"github.com/google/uuid"
	"github.com/stackvity/lung-server/internal/data/models"
	// Import for custom error
)

// PatientRepository defines the interface for interacting with patient data.
// Note: There are NO user accounts, so this manages the temporary "patient"
// session associated with a unique access link.  All operations are tied to
// a session, and data is deleted after a defined retention period.
type PatientRepository interface {
	Repository // Embed the common repository interface

	// CreatePatient creates a new patient session record.  The ID will be
	// a UUID associated with the temporary session, *not* a persistent patient ID.
	CreatePatient(ctx context.Context, patient *models.Patient) error

	// GetPatient retrieves a patient session by its ID (which is a UUID, *not* a
	// traditional patient identifier).  Returns a NotFoundError if not found.
	GetPatient(ctx context.Context, patientID uuid.UUID) (*models.Patient, error)

	GetPatientSessionByLink(ctx context.Context, accessLink string) (*models.PatientSession, error) // Add

	// InvalidateLink marks a patient session's access link as used.  // <-- ADD THIS METHOD TO THE INTERFACE
	InvalidateLink(ctx context.Context, accessLink string) error // <-- ADD THIS METHOD TO THE INTERFACE

	// DeletePatient deletes a patient session and all associated data. This is
	// crucial for data privacy and compliance.
	DeletePatient(ctx context.Context, patientID uuid.UUID) error

	// UpdatePatient is intentionally omitted.  Patient data is tied to a session
	// and should not be directly updated.  Any modifications would involve
	// creating new records, and the old session would eventually be deleted.

	// ListPatients is intentionally omitted.  Listing all patients is a major privacy risk.
}
