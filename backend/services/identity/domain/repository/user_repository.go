package repository

import (
	"context"

	"github.com/fairride/identity/domain/entity"
)

// UserRepository defines persistence operations for User entities.
// All methods return *errors.DomainError on failure.
// FindByID and FindByPhone return CodeNotFound when no record exists.
type UserRepository interface {
	FindByID(ctx context.Context, id string) (*entity.User, error)
	FindByPhone(ctx context.Context, phoneNumber string) (*entity.User, error)
	FindAll(ctx context.Context) ([]*entity.User, error)
	Save(ctx context.Context, user *entity.User) error
	Delete(ctx context.Context, id string) error
}
