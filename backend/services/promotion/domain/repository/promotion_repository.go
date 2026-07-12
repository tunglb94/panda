package repository

import (
	"context"

	"github.com/fairride/promotion/domain/entity"
)

// PromotionRepository is the persistence port for Voucher campaigns and
// their per-rider redemption counters. Implementations: infrastructure/postgres
// (production) and infrastructure/fake (tests / local dev without a DB).
type PromotionRepository interface {
	// FindByID returns a single voucher campaign, or a NotFound *errors.DomainError.
	FindByID(ctx context.Context, id string) (*entity.Voucher, error)

	// FindByCode returns the voucher campaign matching an exact, rider-entered
	// code, or a NotFound *errors.DomainError. Codes are unique (BRB §4.15 Rule 4).
	FindByCode(ctx context.Context, code string) (*entity.Voucher, error)

	// FindAutoApplyCandidates returns active campaigns with no redemption code
	// (Voucher.Code == "") whose Type is in types, scoped to city/vehicleType
	// when set. Used by PromotionService to gather auto-apply candidates
	// (First Ride, Birthday, Weekend, ...) without the rider entering anything.
	FindAutoApplyCandidates(ctx context.Context, city, vehicleType string, types []entity.PromotionType) ([]*entity.Voucher, error)

	// Save inserts or updates a voucher campaign (ON CONFLICT DO UPDATE on ID).
	Save(ctx context.Context, v *entity.Voucher) error

	// UsageCountForRider returns how many times riderID has redeemed voucherID,
	// for the BRB §4.6 per-rider usage limit check ("hết lượt").
	UsageCountForRider(ctx context.Context, voucherID, riderID string) (int64, error)

	// RecordRedemption persists a redemption event (rider, voucher, trip,
	// discount amount, timestamp) and atomically decrements the voucher's
	// remaining budget / increments usage — implementations must do the
	// counter update and the event insert in one transaction.
	RecordRedemption(ctx context.Context, voucherID, riderID, tripID string, discountAmount int64) error

	// ReleaseRedemption reverses RecordRedemption (BRB §4.13 Refund Behaviour /
	// §4.14 Cancellation Behaviour: voucher reinstated when not the rider's fault).
	// discountAmount must match the amount originally passed to RecordRedemption
	// so the voucher's budget is reinstated exactly (BRB §4.13/§4.14 do not
	// permit partial or approximate reinstatement).
	ReleaseRedemption(ctx context.Context, voucherID, riderID, tripID string, discountAmount int64) error
}
