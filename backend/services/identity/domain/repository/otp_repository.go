package repository

import (
	"context"

	"github.com/fairride/identity/domain/entity"
)

// OTPRepository defines persistence operations for OTPChallenge entities.
// All methods return *errors.DomainError on failure.
type OTPRepository interface {
	// Save upserts a challenge (matched by ID).
	Save(ctx context.Context, challenge *entity.OTPChallenge) error

	// FindLatestByPhone returns the most recently created challenge for
	// phoneNumber, regardless of its consumed/expired state — callers decide
	// what to do with it (cooldown check, verification, etc).
	// Returns errors.CodeNotFound if no challenge was ever created for this phone.
	FindLatestByPhone(ctx context.Context, phoneNumber string) (*entity.OTPChallenge, error)
}
