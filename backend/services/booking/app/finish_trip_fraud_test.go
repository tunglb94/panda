package app_test

// Ride Lifecycle Fare Validation — no-movement fraud guard tests.
//
// The guard is opt-in via WithFraudGuard (see finish_trip.go), so every
// test here constructs its own use case with explicit thresholds rather
// than relying on NewFinishTripUseCase's zero-value (disabled) default —
// that default is what keeps every pre-existing FinishTrip test in this
// package passing unchanged.

import (
	"context"
	"testing"

	"github.com/fairride/booking/app"
	sharederrors "github.com/fairride/shared/errors"
)

const (
	fraudTestMinDistanceKm  = 0.3
	fraudTestMinDurationMin = 2.0
)

func newFraudGuardedUseCase(pricing *stubPricing, trip *stubTrip) *app.FinishTripUseCase {
	return app.NewFinishTripUseCase(pricing, trip).WithFraudGuard(fraudTestMinDistanceKm, fraudTestMinDurationMin)
}

// 1. Normal trip: well above both thresholds — completes normally, Pricing
// is called with the real reported distance/duration.
func TestFinishTrip_FraudGuard_NormalTrip_Succeeds(t *testing.T) {
	trip := newStubTrip()
	trip.trips["t1"] = &app.TripInfo{TripID: "t1", Status: "in_progress"}
	pricing := newStubPricing(45_000, "VND")
	uc := newFraudGuardedUseCase(pricing, trip)

	result, err := uc.Execute(context.Background(), app.FinishTripInput{
		TripID: "t1", VehicleType: "car", DistanceKM: 5.2, DurationMin: 18,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pricing.calls != 1 {
		t.Fatalf("pricing.calls = %d, want 1", pricing.calls)
	}
	if pricing.lastDistanceKM != 5.2 || pricing.lastDurationMin != 18 {
		t.Errorf("pricing got (%v, %v), want (5.2, 18)", pricing.lastDistanceKM, pricing.lastDurationMin)
	}
	if result.FinalFare != 45_000 {
		t.Errorf("FinalFare = %d, want 45000", result.FinalFare)
	}
}

// 2. Short legitimate trip: both values are small but ABOVE the configured
// minimums — must NOT be rejected, and must NOT be forced to some
// "base fare" substitute; Pricing is called with the real small values so
// its own Minimum Fare floor (untouched by this change) applies.
func TestFinishTrip_FraudGuard_ShortLegitimateTrip_Succeeds(t *testing.T) {
	trip := newStubTrip()
	trip.trips["t1"] = &app.TripInfo{TripID: "t1", Status: "in_progress"}
	pricing := newStubPricing(15_000, "VND") // e.g. Pricing's own Minimum Fare floor
	uc := newFraudGuardedUseCase(pricing, trip)

	_, err := uc.Execute(context.Background(), app.FinishTripInput{
		TripID: "t1", VehicleType: "car", DistanceKM: 0.4, DurationMin: 2.5,
	})
	if err != nil {
		t.Fatalf("legitimate short trip must not be rejected: %v", err)
	}
	if pricing.lastDistanceKM != 0.4 || pricing.lastDurationMin != 2.5 {
		t.Errorf("pricing got (%v, %v), want the real short-trip values (0.4, 2.5), never forced to 0", pricing.lastDistanceKM, pricing.lastDurationMin)
	}
}

// 3. Early passenger stop: the rider ends the trip well short of whatever
// was originally estimated. FinishTripInput carries only the actual
// distance/duration reported at finish time — there is no "estimated fare"
// field anywhere in this flow, so Pricing structurally cannot be charged
// the original estimate; it only ever sees what actually happened.
func TestFinishTrip_FraudGuard_EarlyPassengerStop_ChargesActualNotEstimate(t *testing.T) {
	trip := newStubTrip()
	trip.trips["t1"] = &app.TripInfo{TripID: "t1", Status: "in_progress"}
	pricing := newStubPricing(22_000, "VND")
	uc := newFraudGuardedUseCase(pricing, trip)

	// Actual trip was cut short: 0.8km / 3.5min, nowhere near a
	// hypothetical original full-route estimate.
	_, err := uc.Execute(context.Background(), app.FinishTripInput{
		TripID: "t1", VehicleType: "car", DistanceKM: 0.8, DurationMin: 3.5,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pricing.lastDistanceKM != 0.8 || pricing.lastDurationMin != 3.5 {
		t.Errorf("pricing got (%v, %v), want the actual travelled (0.8, 3.5) — never an estimate", pricing.lastDistanceKM, pricing.lastDurationMin)
	}
	if trip.lastCompleteFare.TravelledDistanceKm != 0.8 || trip.lastCompleteFare.TravelledDurationMin != 3.5 {
		t.Errorf("Trip Summary forwarded to Trip = (%v, %v), want (0.8, 3.5)",
			trip.lastCompleteFare.TravelledDistanceKm, trip.lastCompleteFare.TravelledDurationMin)
	}
}

// 4. No-movement fraud: both distance AND duration are below the configured
// minimums — FinishTrip must fail with a business error, and Pricing must
// never even be called (fail fast, no wasted fare calculation on an
// abnormal completion).
func TestFinishTrip_FraudGuard_NoMovement_Rejected(t *testing.T) {
	trip := newStubTrip()
	trip.trips["t1"] = &app.TripInfo{TripID: "t1", Status: "in_progress"}
	pricing := newStubPricing(45_000, "VND")
	uc := newFraudGuardedUseCase(pricing, trip)

	_, err := uc.Execute(context.Background(), app.FinishTripInput{
		TripID: "t1", VehicleType: "car", DistanceKM: 0.01, DurationMin: 0.1,
	})
	if !sharederrors.IsCode(err, sharederrors.CodePreconditionFailed) {
		t.Fatalf("want CodePreconditionFailed business error, got %v", err)
	}
	if pricing.calls != 0 {
		t.Errorf("pricing.calls = %d, want 0 (must fail fast before Pricing)", pricing.calls)
	}
	if trip.trips["t1"].Status != "in_progress" {
		t.Errorf("trip status = %q, want unchanged in_progress (never completes on rejection)", trip.trips["t1"].Status)
	}
}

// 5. GPS loss tolerance: distance is glitchy/under-reported (below the
// distance minimum, simulating a dropped GPS signal near the end of the
// trip) but duration — which runs off the device clock, not GPS — is well
// above its minimum. The AND-based guard must NOT reject: a real trip's
// elapsed time proves it happened even when the final distance reading is
// bad. Never falls back to any straight-line distance.
func TestFinishTrip_FraudGuard_GPSLossTolerated_Succeeds(t *testing.T) {
	trip := newStubTrip()
	trip.trips["t1"] = &app.TripInfo{TripID: "t1", Status: "in_progress"}
	pricing := newStubPricing(30_000, "VND")
	uc := newFraudGuardedUseCase(pricing, trip)

	_, err := uc.Execute(context.Background(), app.FinishTripInput{
		TripID: "t1", VehicleType: "car", DistanceKM: 0.05, DurationMin: 12,
	})
	if err != nil {
		t.Fatalf("GPS-loss trip (low distance, real duration) must be tolerated, got: %v", err)
	}
	if pricing.calls != 1 {
		t.Errorf("pricing.calls = %d, want 1", pricing.calls)
	}
	if pricing.lastDistanceKM != 0.05 {
		t.Errorf("pricing distance = %v, want the real reported 0.05 — never a straight-line substitute", pricing.lastDistanceKM)
	}
}
