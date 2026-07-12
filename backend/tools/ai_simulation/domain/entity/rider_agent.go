package entity

// Membership mirrors the passenger membership tiers referenced in
// ECONOMY_ENGINE.md §8 (Free/Silver/Gold/Diamond) — a rider-facing
// classification, distinct from Voucher.Membership eligibility gating in
// the real Promotion Engine.
type Membership string

const (
	MembershipFree    Membership = "free"
	MembershipSilver  Membership = "silver"
	MembershipGold    Membership = "gold"
	MembershipDiamond Membership = "diamond"
)

// CommuteHabit is a coarse behavioral archetype driving when a rider tends
// to request trips (see simulation/scenario_scheduler.go). Simulation-only
// classification — not sourced from any BRB rule.
type CommuteHabit string

const (
	HabitCommuter    CommuteHabit = "commuter"    // regular go-to-work/school trips
	HabitOccasional  CommuteHabit = "occasional"  // sporadic, weekend-leaning
	HabitNightOwl    CommuteHabit = "night_owl"   // evening/night entertainment trips
	HabitBusinessTraveler CommuteHabit = "business_traveler" // airport-heavy
)

// RiderAgent is one simulated rider — exactly the fields the sprint brief
// requires, plus tick-bookkeeping fields.
type RiderAgent struct {
	ID             string
	PriceSensitivity float64 // 0-1, 1 = extremely price-sensitive (BRB §1.2-adjacent behavior signal)
	Income           int64   // VND/month, informs price sensitivity and voucher affinity
	Habit            CommuteHabit
	WorkStartHour    int // "giờ đi làm"
	WorkEndHour      int // "giờ tan ca"
	Patience         float64 // 0-1, tolerance for ETA/surge before abandoning or switching apps
	Membership       Membership

	// Mutable simulation state
	HasActiveVoucher bool
	VoucherDiscountPercent int
	TripCount        int
	Zone             ZoneType
	CurrentTripID    string // empty when not mid-request/trip
}
