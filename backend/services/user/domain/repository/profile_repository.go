// Package repository defines the persistence interfaces for the User service.
package repository

import (
	"context"

	"github.com/fairride/user/domain/entity"
)

// ProfileRepository defines persistence operations for UserProfile entities.
// All methods return *errors.DomainError on failure.
// FindByID returns CodeNotFound when no profile exists for the given ID.
type ProfileRepository interface {
	FindByID(ctx context.Context, id string) (*entity.UserProfile, error)
	Save(ctx context.Context, profile *entity.UserProfile) error
}
