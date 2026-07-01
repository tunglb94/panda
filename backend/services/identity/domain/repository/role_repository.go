package repository

import (
	"context"

	"github.com/fairride/identity/domain/entity"
)

// RoleRepository is the persistence contract for Role entities.
//
// All methods return a *errors.DomainError from the shared errors package when the
// operation fails. Callers use errors.IsCode to distinguish error kinds.
type RoleRepository interface {
	// FindByID returns the role with the given ID, including its permissions.
	// Returns errors.CodeNotFound if no such role exists.
	FindByID(ctx context.Context, id string) (*entity.Role, error)

	// FindByName returns the role with the given name (e.g. "rider").
	// Returns errors.CodeNotFound if no such role exists.
	FindByName(ctx context.Context, name string) (*entity.Role, error)

	// FindAll returns every role in the store with their permissions.
	FindAll(ctx context.Context) ([]*entity.Role, error)

	// Save persists a role and its current permission set.
	// Performs an upsert: insert on new ID, update on existing ID.
	// The full permission set is replaced on every save.
	Save(ctx context.Context, role *entity.Role) error

	// Delete removes the role with the given ID.
	// Callers MUST call role.CanDelete() before invoking this method.
	// Returns errors.CodeNotFound if no such role exists.
	Delete(ctx context.Context, id string) error
}
