package entities

import (
	"time"

	"github.com/google/uuid"
)

// Entity is a common interface for domain entities.
// It provides basic methods for getting common entity properties.
type Entity interface {
	GetID() uuid.UUID
	GetCreatedAt() time.Time
	GetUpdatedAt() time.Time
}

// BaseEntity is a common struct that can be embedded in other entities
// to provide common fields like ID, CreatedAt, and UpdatedAt.
type BaseEntity struct {
	ID        uuid.UUID `json:"id,omitempty"`
	CreatedAt time.Time `json:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}

// GetID returns the ID of the entity.
func (be *BaseEntity) GetID() uuid.UUID {
	return be.ID
}

// GetCreatedAt returns the creation timestamp of the entity.
func (be *BaseEntity) GetCreatedAt() time.Time {
	return be.CreatedAt
}

// GetUpdatedAt returns the last update timestamp of the entity.
func (be *BaseEntity) GetUpdatedAt() time.Time {
	return be.UpdatedAt
}
