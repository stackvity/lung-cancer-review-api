// internal/data/repositories/postgres/patient_repository.go
package postgres

import (
	"context"
	"errors" // Import the standard errors package
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool" // Import pgxpool
	"github.com/stackvity/lung-server/internal/data/models"
	"github.com/stackvity/lung-server/internal/data/repositories/interfaces"
	postgres "github.com/stackvity/lung-server/internal/data/repositories/sqlc" // Import the generated sqlc package AS postgres  <--- ALIAS IS HERE
	"github.com/stackvity/lung-server/internal/domain"                          // For custom errors
)

// Ensure PatientRepository implements the interface.
var _ interfaces.PatientRepository = (*PatientRepository)(nil)

// PatientRepository provides PostgreSQL-specific access to patient data.
type PatientRepository struct {
	db      *pgxpool.Pool     // Corrected type: Use *pgxpool.Pool directly
	Queries *postgres.Queries // Use the generated Queries struct.  Now using the alias 'postgres'
}

// NewPatientRepository creates a new PatientRepository instance.
func NewPatientRepository(db *pgxpool.Pool) *PatientRepository { // Corrected type: Use *pgxpool.Pool directly
	return &PatientRepository{
		Queries: postgres.New(), // Corrected: Call postgres.New() with NO arguments
		db:      db,
	}
}

// CreatePatient creates a new patient session record.
func (r *PatientRepository) CreatePatient(ctx context.Context, patient *models.Patient) error {
	arg := &postgres.CreatePatientSessionParams{ // CORRECTED: Use the alias 'postgres'
		AccessLink:          patient.AccessLink,
		ExpirationTimestamp: pgtype.Timestamptz{Time: patient.ExpirationTimestamp, Valid: true},
		Used:                patient.Used,
		PatientData:         []byte(patient.PatientData),
	}

	_, err := r.Queries.CreatePatientSession(ctx, r.db, arg) // Use generated function - Queries is already using alias
	if err != nil {
		return fmt.Errorf("CreatePatientSession failed: %w", err) // Wrap for context
	}
	return nil
}

// GetPatient retrieves a patient session by its ID (UUID).
func (r *PatientRepository) GetPatient(ctx context.Context, patientID uuid.UUID) (*models.Patient, error) {
	// sqlc doesn't have GetPatientByID, and we can't add it based on current constraints.
	// We're limited to using GetPatientSessionByLink and would need to first look up
	// the link associated with the patient ID (which is the session ID in this case).
	// This is inefficient, but it's a constraint imposed by the task's limitations.

	// In a real implementation, you'd have a query for GetPatientByID in queries.sql.

	return nil, fmt.Errorf("GetPatientByID is not directly supported; use link-based access") // Return an error.
}

// DeletePatient deletes a patient session and all associated data.
func (r *PatientRepository) DeletePatient(ctx context.Context, patientID uuid.UUID) error {
	// Correctly form the pgtype.UUID value
	var bytes [16]byte
	copy(bytes[:], patientID[:])
	err := r.Queries.DeletePatientSession(ctx, r.db, pgtype.UUID{Bytes: bytes, Valid: true}) // Corrected: Use Bytes field
	if err != nil {
		return fmt.Errorf("DeletePatientSession failed: %w", err)
	}
	return nil
}

// InvalidateLink marks a patient session's access link as used.
func (r *PatientRepository) InvalidateLink(ctx context.Context, accessLink string) error {
	err := r.Queries.InvalidateLink(ctx, r.db, accessLink) // Use r.Queries alias
	return err
}

// BeginTx implements the interface method.
func (r *PatientRepository) BeginTx(ctx context.Context, opts ...pgx.TxOptions) (pgx.Tx, error) {
	if len(opts) > 0 {
		return r.db.BeginTx(ctx, opts[0]) // Use provided options if available
	}
	return r.db.Begin(ctx) // Use default transaction options
}

// CommitTx implements the interface method.
func (r *PatientRepository) CommitTx(ctx context.Context, tx pgx.Tx) error {
	return tx.Commit(ctx)
}

// RollbackTx implements the interface method.
func (r *PatientRepository) RollbackTx(ctx context.Context, tx pgx.Tx) error {
	return tx.Rollback(ctx)
}

// GetPatientSessionByLink implements GetPatientSessionByLink
func (r *PatientRepository) GetPatientSessionByLink(ctx context.Context, accessLink string) (*models.PatientSession, error) {
	patientSession, err := r.Queries.GetPatientSessionByLink(ctx, r.db, accessLink) // Use r.Queries alias
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.NewNotFoundError("patient session", accessLink)
		}
		return nil, fmt.Errorf("GetPatientSessionByLink failed: %w", err)
	}
	return &models.PatientSession{
		SessionID:           uuid.UUID(patientSession.SessionID.Bytes),
		AccessLink:          patientSession.AccessLink,
		ExpirationTimestamp: patientSession.ExpirationTimestamp.Time,
		Used:                patientSession.Used,
		PatientData:         string(patientSession.PatientData),
		CreatedAt:           patientSession.CreatedAt.Time,
		UpdatedAt:           patientSession.UpdatedAt.Time,
	}, nil
}
