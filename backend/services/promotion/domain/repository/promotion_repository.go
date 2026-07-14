package repository

import (
	"context"
	"time"

	"github.com/fairride/promotion/domain/entity"
)

// PromotionRepository is the persistence port for Voucher campaigns, their
// per-rider redemption lifecycle, and per-rider issuance. Implementations:
// infrastructure/postgres (production) and infrastructure/fake (tests /
// local dev without a DB).
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

	// FindAll returns every voucher campaign — the Admin app's CRUD list.
	// No filtering/pagination: expected campaign count is small (tens, not
	// thousands); callers filter/search client-side.
	FindAll(ctx context.Context) ([]*entity.Voucher, error)

	// UsageCountForRider returns how many times riderID has redeemed voucherID
	// (status IN ('reserved','redeemed') — an in-flight reservation already
	// counts against the per-user limit so the same rider can't double-book
	// with the same voucher), for the BRB §4.6 per-rider usage limit check.
	UsageCountForRider(ctx context.Context, voucherID, riderID string) (int64, error)

	// ─── Redemption lifecycle: Reserve (booking) -> ConfirmRedeem (trip
	// completed) -> Release (cancelled). All three are idempotent. ───────────

	// Reserve tentatively holds discountAmount against the voucher's budget
	// for (voucherID, riderID, tripID) — status starts 'reserved'. Calling
	// Reserve again for the same (voucher, rider, trip) is a no-op (UNIQUE
	// constraint backs this), never a double deduction.
	Reserve(ctx context.Context, voucherID, riderID, tripID string, discountAmount int64) error

	// ConfirmRedeem transitions a 'reserved' row to 'redeemed' — called once
	// the trip actually completes. Does not touch budget (already deducted
	// at Reserve time). Idempotent: already-'redeemed' is a no-op success;
	// no matching 'reserved' row is a CodeNotFound error the caller can
	// choose to ignore (e.g. no voucher was ever reserved for this trip).
	ConfirmRedeem(ctx context.Context, voucherID, riderID, tripID string) error

	// Release transitions a 'reserved' row to 'released' and reinstates
	// discountAmount to the voucher's budget — called when a trip is
	// cancelled before completion (BRB §4.13/§4.14: rider keeps the
	// voucher). Idempotent: already-'released' is a no-op success. Only
	// valid from 'reserved' — a 'redeemed' (trip already completed)
	// reservation cannot be released.
	Release(ctx context.Context, voucherID, riderID, tripID string, discountAmount int64) error

	// FindReservationByTrip returns the redemption row for tripID regardless
	// of status (one voucher per trip, BRB §4.7), or CodeNotFound if no
	// voucher was ever reserved against it. Used by ConfirmRedeem/Release
	// callers that only have a trip_id (Gateway's FinishTrip/CancelRide
	// hooks) to resolve which voucher/discount to act on.
	FindReservationByTrip(ctx context.Context, tripID string) (*entity.RedemptionRecord, error)

	// ListRedemptionsByRider returns riderID's full redemption history
	// (reserved/redeemed/released rows), newest first.
	ListRedemptionsByRider(ctx context.Context, riderID string) ([]*entity.RedemptionRecord, error)

	// ─── Per-rider issuance — see entity.VoucherIssuance's doc comment. ─────

	// IssueToRider grants voucherID to riderID. Idempotent: issuing the same
	// pair twice is a no-op (does not reset an already-'used' issuance back
	// to 'issued').
	IssueToRider(ctx context.Context, voucherID, riderID string, now time.Time) error

	// ListIssuancesForRider returns every voucher issued to riderID, newest
	// first, each with its Voucher populated (JOIN) — the Rider app's
	// voucher wallet (Available/Used/Expired, bucketed client-side via
	// Voucher.EffectiveState).
	ListIssuancesForRider(ctx context.Context, riderID string) ([]*entity.VoucherIssuance, error)

	// MarkIssuanceUsed transitions an issuance to 'used' — best-effort,
	// called alongside ConfirmRedeem. No-op if no issuance row exists (a
	// rider who entered a code that was never explicitly issued to them
	// still redeems successfully; issuance is a wallet/reporting concern,
	// never an eligibility gate).
	MarkIssuanceUsed(ctx context.Context, voucherID, riderID string, now time.Time) error

	// CountIssued/CountRedeemed/CountExpired back the Admin app's per-voucher
	// stats (Issued/Redeemed/Remaining/Expired) — Remaining is derived by
	// the caller from Voucher.RemainingBudget/MaxUsage, not stored separately.
	CountIssued(ctx context.Context, voucherID string) (int64, error)
	CountRedeemed(ctx context.Context, voucherID string) (int64, error)
	CountExpiredIssuances(ctx context.Context, voucherID string, now time.Time) (int64, error)
}
