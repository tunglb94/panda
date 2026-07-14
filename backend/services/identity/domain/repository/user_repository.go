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
	// FindByEmail returns the user with the given email. Only meaningful for
	// non-empty emails — callers must not call this with "" (every user
	// without an email shares the empty string, so a lookup would be
	// ambiguous; the email column's uniqueness is only enforced for non-empty
	// values, see migration 018).
	FindByEmail(ctx context.Context, email string) (*entity.User, error)
	// FindByGoogleSub returns the user linked to the given Google subject ID.
	// Same non-empty-only contract as FindByEmail (see migration 020).
	FindByGoogleSub(ctx context.Context, googleSub string) (*entity.User, error)
	FindAll(ctx context.Context) ([]*entity.User, error)
	Save(ctx context.Context, user *entity.User) error
	Delete(ctx context.Context, id string) error
}
