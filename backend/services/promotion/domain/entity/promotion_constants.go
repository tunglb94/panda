package entity

// Constants sourced verbatim from Business Rule Bible v1.0 (BRB) Part 3.
// These are the DEFAULT values used when seeding a BRB-defined campaign; the
// Voucher itself still carries its own DiscountValue/MaxDiscount/etc so
// Operations can run a different, explicitly-approved campaign instance
// without a code change. No constant exists here for a promotion type BRB
// does not define — see PromotionType.IsDefinedInBRB.
const (
	// BRB §3.2.1 First Ride Promotion
	FirstRideDiscountPercent int64 = 50
	FirstRideMaxDiscountVND  int64 = 30_000
	FirstRideAccountAgeDays  int64 = 30

	// BRB §3.2.2 Birthday Promotion
	BirthdayDiscountPercent  int64 = 40
	BirthdayMaxDiscountVND   int64 = 20_000
	BirthdayMinTrips90Days   int64 = 3
	BirthdayWindowDays       int64 = 1 // +/- 1 day around birthday (3 days total)

	// BRB §3.2.3 Golden Hour Promotion
	GoldenHourDiscountPercent int64 = 20
	GoldenHourMaxDiscountVND  int64 = 15_000

	// BRB §3.2.4 Weekend Promotion
	WeekendDiscountPercent int64 = 15
	WeekendMaxDiscountVND  int64 = 20_000
	WeekendMaxRidesPerUser int64 = 2 // per weekend

	// BRB §3.2.5 Rain Campaign
	RainDiscountPercent   int64 = 10
	RainMaxDiscountVND    int64 = 10_000
	RainMinActiveDays     int64 = 7

	// BRB §3.2.7 Referral Programme
	ReferralReferrerRewardVND int64 = 20_000 // Rider A wallet credit
	ReferralRefereeDiscountVND int64 = 30_000 // Rider B first-ride discount
	ReferralHoldDays          int64 = 7
	ReferralMaxPerYear        int64 = 50

	// BRB §3.2.8 Cashback Promotion (post-trip; out of scope for this engine's
	// pre-trip Evaluate(), documented here for completeness since it appears in
	// BRB §3.4 Campaign Priority Rules).
	CashbackMaxVND int64 = 15_000

	// BRB §4.5 Voucher Expiration
	VoucherDefaultExpiryDays      int64 = 30
	VoucherCompensationExpiryDays int64 = 90

	// BRB §4.15 Fraud Protection
	FraudMaxVouchersPer7Days int64 = 3
	FraudShortTripKM         float64 = 2.0
	VoucherCodeMinLength     int    = 12
)
