package repository

import (
	"context"

	"github.com/fairride/wallet/domain/entity"
)

// SettlementRepository persists immutable Settlement records (Phần 2/13).
// No Update/Delete method exists on this interface by design.
type SettlementRepository interface {
	// Save inserts a new Settlement. Returns CodeAlreadyExists if a
	// Settlement for this TripID already exists (Settlement Engine
	// idempotency — see UNIQUE(trip_id) in the migration).
	Save(ctx context.Context, s *entity.Settlement) error
	// MarkPosted is the one narrow, explicit exception to "no Update method"
	// — it transitions Status Pending->Posted once the Settlement Engine has
	// finished writing every ledger entry (crash-safety, critique #7). It
	// never touches any financial field.
	MarkPosted(ctx context.Context, settlementID string) error
	FindByTripID(ctx context.Context, tripID string) (*entity.Settlement, error)
	FindByID(ctx context.Context, settlementID string) (*entity.Settlement, error)
	// ListByDriverID returns a driver's settlements in [from, to)
	// (Phần 9 — Driver Statement), newest first, capped at limit.
	ListByDriverID(ctx context.Context, driverID string, from, to int64, limit int) ([]*entity.Settlement, error)
	// ListAll powers the admin Settlement List (Phần 10) — driverID/from/to
	// are optional filters (empty/0 means "any").
	ListAll(ctx context.Context, driverID string, from, to int64, limit int) ([]*entity.Settlement, error)
}
