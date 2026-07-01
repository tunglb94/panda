// Package repository defines the persistence contracts for the Identity domain.
// Implementations live in the infrastructure layer; only interfaces are declared here.
package repository

import (
	"context"

	"github.com/fairride/identity/domain/entity"
)

// PermissionRepository is the persistence contract for Permission entities.
//
// All methods return a *errors.DomainError from the shared errors package when the
// operation fails. Callers use errors.IsCode to distinguish error kinds.
type PermissionRepository interface {
	// FindByID returns the permission with the given ID.
	// Returns errors.CodeNotFound if no such permission exists.
	FindByID(ctx context.Context, id string) (*entity.Permission, error)

	// FindByName returns the permission with the given canonical name (e.g. "trips:read").
	// Returns errors.CodeNotFound if no such permission exists.
	FindByName(ctx context.Context, name string) (*entity.Permission, error)

	// FindByResource returns all permissions whose Resource field matches.
	// Returns an empty slice (not an error) when none are found.
	FindByResource(ctx context.Context, resource string) ([]*entity.Permission, error)

	// FindAll returns every permission in the store.
	FindAll(ctx context.Context) ([]*entity.Permission, error)

	// Save persists a permission. Performs an upsert: insert on new ID, update on existing ID.
	Save(ctx context.Context, permission *entity.Permission) error

	// Delete removes the permission with the given ID.
	// Returns errors.CodeNotFound if no such permission exists.
	Delete(ctx context.Context, id string) error
}
