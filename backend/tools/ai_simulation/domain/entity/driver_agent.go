package entity

// AccountType mirrors the driver tiers documented in
// business-rule-bible-v1.0.md §7.1 (commission rate by tier — see
// integration/driver_economy.go, which is new code since no production Go
// implementation of tiered commission exists yet).
type AccountType string

const (
	AccountBronze   AccountType = "bronze"
	AccountSilver   AccountType = "silver"
	AccountGold     AccountType = "gold"
	AccountPlatinum AccountType = "platinum"
	AccountDiamond  AccountType = "diamond"
)

// DriverVehicleType matches the production entity.VehicleType values used by
// backend/services/driver and backend/services/pricing (car/motorcycle/van)
// — the driver's PHYSICAL vehicle, kept as a separate simulation-local type
// so this package has no compile dependency on either service's entity
// package for a single string enum.
type DriverVehicleType string

const (
	VehicleCar        DriverVehicleType = "car"
	VehicleMotorcycle DriverVehicleType = "motorcycle"
	VehicleVan        DriverVehicleType = "van"
)

// ServiceType matches production's driver/dispatch ServiceType — the
// product/service tier (Bike, Bike Plus, Car, Car XL), a dimension
// separate from DriverVehicleType (Vehicle/Service Catalog refactor: a
// Honda Wave and a Honda SH are both VehicleMotorcycle, but might operate
// as ServiceBike vs ServiceBikePlus respectively). Applies identically to
// Ride and Delivery — there is no "delivery_bike"/"delivery_car" ServiceType.
type ServiceType string

const (
	// ServiceBike and ServiceCar reuse VehicleMotorcycle's/VehicleCar's
	// exact wire values (product-facing alias, not a rename).
	ServiceBike     ServiceType = "motorcycle"
	ServiceBikePlus ServiceType = "bike_plus"
	ServiceCar      ServiceType = "car"
	ServiceCarXL    ServiceType = "car_xl"
)

// RequiredVehicleType mirrors driver.ServiceType.RequiredVehicleType —
// which physical vehicle a driver needs to legitimately offer this service
// tier.
func (s ServiceType) RequiredVehicleType() DriverVehicleType {
	switch s {
	case ServiceBike, ServiceBikePlus:
		return VehicleMotorcycle
	case ServiceCarXL:
		return VehicleVan
	default:
		return VehicleCar
	}
}

// DriverAgent is one simulated driver — exactly the fields the sprint brief
// requires, plus the bookkeeping a tick-based simulation needs (Zone/Online
// for the world to place and query agents).
type DriverAgent struct {
	ID          string
	Age         int
	Experience  int // years driving for any platform
	Rating      float64
	VehicleType DriverVehicleType
	ServiceType ServiceType
	// RideEnabled/DeliveryEnabled mirror driver.DriverProfile's capability
	// flags (migration 008) — a driver can be Ride-only, Delivery-only, or
	// both.
	RideEnabled     bool
	DeliveryEnabled bool
	AccountType     AccountType

	// Mutable simulation state
	IncomeToday  int64   // VND, resets at the start of each simulated day
	IncomeWeek   int64   // VND, resets every 7 simulated days
	Satisfaction float64 // 0-1, drifts based on income/fatigue/wait outcomes
	Fatigue      float64 // 0-1, rises with consecutive online hours, falls when offline
	// FatigueGainPerTick is this driver's own stamina — how much Fatigue
	// rises per 15-min tick while online (see simulation/engine.go's
	// evaluateDriverState). Seeded once per driver (simulation/seed.go) so
	// individual online-stretch length naturally varies across the
	// population instead of every driver sharing one flat rate — see
	// CHANGELOG's Driver Economy shift-classification fix.
	FatigueGainPerTick float64
	PhoneBattery       float64 // 0-1
	Fuel               float64 // 0-1 (or state-of-charge for EVs — simulation doesn't distinguish)
	Cash               int64   // VND on hand (wallet-adjacent, not the real Wallet service)
	Zone               ZoneType

	Online           bool
	HoursOnlineToday float64
	CurrentTripID    string // empty when idle
	TotalTrips       int

	// Lifetime bookkeeping for driver_analytics.json — never reset daily,
	// unlike IncomeToday/HoursOnlineToday.
	TotalHoursOnline float64 // cumulative online hours across the whole run
	DaysActive       int     // simulated days this driver was online at least once

	// Audit-only bookkeeping (backend/tools/ai_simulation/audit) — additive
	// instrumentation, not decision-affecting state.
	// TripsThisRun counts only trips completed during THIS run; TotalTrips
	// is seeded with a random pre-existing lifetime count (see
	// simulation/seed.go) specifically to give new DriverEconomy.TierForDriver
	// callers a realistic tier distribution, so it is NOT a reliable signal
	// for "did this driver actually work this run" — TripsThisRun is.
	TripsThisRun int
	// MaxHoursOnlineContinuous is the highest HoursOnlineToday ever observed
	// for this driver before a day boundary/offline reset — HoursOnlineToday
	// itself resets daily, so it alone cannot answer "did any driver ever
	// exceed the 12h fatigue floor".
	MaxHoursOnlineContinuous float64

	// BI-only bookkeeping (backend/tools/ai_simulation/bi) — per-driver
	// offer accept/reject counts, needed for driver_leaderboard.json and
	// driver_economy.json's per-shift-category Acceptance Rate. The World-
	// level driverOffersAccepted/Rejected counters already track this in
	// aggregate (see simulation/world.go) but not broken out per driver.
	OffersAccepted int
	OffersRejected int
}

// IsAvailable reports whether this driver can currently be offered a trip.
func (d *DriverAgent) IsAvailable() bool {
	return d.Online && d.CurrentTripID == ""
}
