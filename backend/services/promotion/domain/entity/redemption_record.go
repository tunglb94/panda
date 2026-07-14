package entity

import "time"

// RedemptionRecord is one row of a rider's voucher redemption history —
// joins a voucher_redemptions row with its voucher's display info, for the
// Rider app's voucher wallet ("Used"/"Expired" tabs — see
// PromotionRepository.ListRedemptionsByRider).
type RedemptionRecord struct {
	VoucherID      string
	VoucherCode    string
	VoucherName    string
	RiderID        string
	TripID         string
	DiscountAmount int64
	Status         string // "reserved", "redeemed", or "released" (see infrastructure/postgres's Reserve/ConfirmRedeem/Release)
	RedeemedAt     time.Time
}
