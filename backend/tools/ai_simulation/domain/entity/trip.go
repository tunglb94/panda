package entity

import "time"

// TripOutcome is the terminal state of one simulated ride request.
type TripOutcome string

const (
	OutcomeCompleted TripOutcome = "completed"
	OutcomeRejected  TripOutcome = "rejected"  // every candidate driver rejected / none found
	OutcomeCancelled TripOutcome = "cancelled" // rider cancelled (e.g. abandoned due to ETA/price)
	OutcomePending   TripOutcome = "pending"   // still in flight at end of simulation
)

// TripKind distinguishes a Ride request from a Delivery (parcel) request —
// the simulation's own local mirror of the production TripType discriminator
// (github.com/fairride/trip/domain/entity.TripType,
// github.com/fairride/dispatch/domain/entity.TripType), kept as a separate
// type so domain/entity has no compile dependency on either service's
// package for a single string enum (same pattern as DriverVehicleType).
type TripKind string

const (
	KindRide     TripKind = "ride"
	KindDelivery TripKind = "delivery"
)

// SimTrip is one ride or delivery request's full record, retained for
// statistics and the dashboard (heatmap/timeline). Money fields are VND,
// matching BRB convention.
type SimTrip struct {
	TripID   string
	RiderID  string
	DriverID string // empty until assigned

	Kind TripKind // "ride" (default/zero value) or "delivery"

	PickupZone      ZoneType
	DestinationZone ZoneType
	DistanceKM      float64

	RequestedAtTick int64
	AssignedAtTick  int64 // 0 if never assigned
	CompletedAtTick int64 // 0 if never completed

	Outcome TripOutcome

	// ServiceType is the product/service tier this trip was requested at
	// (Bike/Bike Plus/Car/Car XL) — applies identically whether Kind is
	// Ride or Delivery (Vehicle/Service Catalog refactor).
	ServiceType ServiceType

	BaseFareVND        int64
	SurgeMultiplier    float64
	VoucherDiscountVND int64
	PromotionType      string // empty if none applied
	FinalFareVND       int64
	CommissionVND      int64
	DriverNetVND       int64

	Weather Weather
	Traffic Traffic

	ETAMinutes    float64
	PickupMinutes float64

	// Delivery-only fields (Kind == KindDelivery); zero-valued for Ride.
	// Like Ride's own PickupMinutes/ETAMinutes, these are computed durations
	// resolved within the single tick a request is processed on — the
	// simulation does not model a delivery's transit as elapsing real ticks.
	PackageWeightKg        float64
	DeliveryTransitMinutes float64 // pickup -> delivered travel time (0 if never picked up)

	CreatedAt time.Time
}
