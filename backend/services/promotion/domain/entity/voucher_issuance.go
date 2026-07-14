package entity

import "time"

// VoucherIssuanceStatus tracks one rider's personal copy of a voucher —
// distinct from VoucherStatus (the campaign's own lifecycle). A rider only
// ever sees a voucher in their wallet once it's been issued to them
// specifically (see PromotionRepository.IssueToRider) — this is the
// "issuance riêng cho từng rider" architecture the Voucher Wallet phase
// asked to prepare, without building a full segment-targeting/Referral
// system on top of it yet.
type VoucherIssuanceStatus string

const (
	IssuanceStatusIssued VoucherIssuanceStatus = "issued"
	IssuanceStatusUsed   VoucherIssuanceStatus = "used"
)

// VoucherIssuance is one (voucher, rider) grant. Voucher is populated (a
// JOIN, not a separate lookup) by ListIssuancesForRider so callers can read
// discount/expiry/EffectiveState without a second round-trip.
type VoucherIssuance struct {
	VoucherID string
	RiderID   string
	Status    VoucherIssuanceStatus
	IssuedAt  time.Time
	UsedAt    *time.Time
	Voucher   *Voucher
}
