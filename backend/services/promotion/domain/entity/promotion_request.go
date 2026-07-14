package entity

import "time"

// PromotionRequest is the input PromotionService.Evaluate needs to determine
// how much discount (if any) a trip is eligible for. Pricing Service is
// expected to build one of these per fare quote/confirmation and call the
// Promotion Engine — this is the "Pricing Service chỉ hỏi Promotion Engine
// được giảm bao nhiêu" contract from the sprint brief.
type PromotionRequest struct {
	RiderID     string
	VehicleType string // matches production entity.VehicleType values: car/motorcycle/van
	City        string
	OrderAmount int64 // VND, pre-discount trip fare (metered + surcharges + booking fee per BRB §3.2.1)

	// ServiceType is the Rider app's product tier (motorcycle/bike_plus/car/car_xl)
	// — see Voucher.ServiceTypes' doc comment for why this is a separate
	// dimension from VehicleType. Empty matches any voucher (including ones
	// scoped to specific service types, since an unset request-side filter
	// means "caller didn't restrict" — the voucher's own ServiceTypes list is
	// what actually gates eligibility).
	ServiceType string

	// TripType is "ride" or "delivery". Empty behaves like ServiceType above.
	TripType string

	RequestTime time.Time // trip request time, used for time-window rules (Weekend, Golden Hour, Birthday +/-1 day)

	// VoucherCode: set when the rider explicitly entered a code. If empty, the
	// engine only considers auto-apply campaigns (Voucher.Code == "").
	VoucherCode string

	// Rider profile fields needed by specific PromotionRule implementations.
	// All optional; a rule that needs a field it wasn't given simply treats the
	// rider as ineligible for that rule (fails closed, never fabricates
	// eligibility).
	AccountCreatedAt         *time.Time // BRB §3.2.1 First Ride: "account created within the last 30 days"
	CompletedTripsTotal      int64      // BRB §3.2.1 First Ride: "zero completed trips"
	CompletedTripsLast90Days int64      // BRB §3.2.2 Birthday: "at least 3 trips in the past 90 days"
	BirthdayDate             *time.Time // BRB §3.2.2 (month+day compared against RequestTime)
	RiderActiveSinceDays     int64      // BRB §3.2.5 Rain: "active on the platform for at least 7 days"
	MembershipTier           string     // ECONOMY_ENGINE membership tiers (eligibility gate only, never a discount input)
	ReferralCode             string     // BRB §3.2.7 (code of the referring rider, if any)
	IsReferredFirstTrip      bool       // BRB §3.4 #2: Referral only applies "the rider's first trip"
	LastTripAt               *time.Time // needed by the (currently TODO) Comeback rule
	IsRainSurchargeActive    bool       // BRB §3.2.5 Rain Campaign trigger condition
	IsGoldenHourWindowActive bool       // BRB §3.2.3: "Active window: Defined per campaign" — caller resolves whether now falls in the configured window
}
