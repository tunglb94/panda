package repository

import (
	"context"

	"github.com/fairride/user/domain/entity"
)

// RiderVerificationRepository defines persistence operations for
// RiderVerification entities. All methods return *errors.DomainError on failure.
type RiderVerificationRepository interface {
	// FindByUserID returns the rider's verification record.
	// Returns errors.CodeNotFound if the rider has never uploaded anything.
	FindByUserID(ctx context.Context, userID string) (*entity.RiderVerification, error)

	// Save upserts a record, matched by ID.
	Save(ctx context.Context, v *entity.RiderVerification) error

	// ListByStatus returns every record in the given status, oldest first —
	// backs the admin-only review API (no dashboard UI in this phase).
	ListByStatus(ctx context.Context, status entity.RiderKYCStatus) ([]*entity.RiderVerification, error)
}
