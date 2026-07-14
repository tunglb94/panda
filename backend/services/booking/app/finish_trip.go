package app

import (
	"context"

	domainerrors "github.com/fairride/shared/errors"
)

// FinishTripInput is the input to FinishTripUseCase.
type FinishTripInput struct {
	TripID      string
	VehicleType string
	DistanceKM  float64
	DurationMin float64

	// VoucherID/VoucherCode/VoucherDiscountCents are resolved by the gateway
	// (Promotion Engine's ConfirmRedeem — see gateway's BookingHandler.FinishTrip)
	// and passed through here so they land on Trip alongside the fare, without
	// this use case (or Pricing) ever computing a discount itself. Empty
	// VoucherID means no voucher was applied to this trip.
	VoucherID            string
	VoucherCode          string
	VoucherDiscountCents int64
}

// FinishedTripResult holds the fare and final trip state.
type FinishedTripResult struct {
	TripID      string
	Status      string // "payment_pending" after B1
	FinalFare   int64
	Currency    string
	VehicleType string
	DistanceKM  float64
	DurationMin float64
}

// FinishTripUseCase orchestrates fare calculation and trip completion.
// Steps: fraud guard (Trip Summary) → pricing.CalculateFinalFare → trip.CompleteTrip
type FinishTripUseCase struct {
	pricing PricingClient
	trip    TripClient
	idem    IdempotencyStore // nil = no idempotency checking

	// minDistanceKm/minDurationMin gate the no-movement fraud guard — zero
	// value (the default, unless WithFraudGuard is called) disables it, so
	// existing callers/tests are unaffected unless they opt in. See
	// WithFraudGuard and Execute's doc comment for the AND-based rule.
	minDistanceKm  float64
	minDurationMin float64
}

func NewFinishTripUseCase(pricing PricingClient, trip TripClient) *FinishTripUseCase {
	return &FinishTripUseCase{pricing: pricing, trip: trip}
}

// WithIdempotency attaches an idempotency store. The natural key is "finish:" + tripID,
// preventing a duplicate finish from triggering a second fare charge.
func (uc *FinishTripUseCase) WithIdempotency(store IdempotencyStore) *FinishTripUseCase {
	uc.idem = store
	return uc
}

// WithFraudGuard sets the configurable minimum-distance/minimum-duration
// thresholds for the no-movement abnormal-completion check (Ride Lifecycle
// Fare Validation). Thresholds come from config (see booking's cmd/server/main.go),
// never hardcoded in this use case.
func (uc *FinishTripUseCase) WithFraudGuard(minDistanceKm, minDurationMin float64) *FinishTripUseCase {
	uc.minDistanceKm = minDistanceKm
	uc.minDurationMin = minDurationMin
	return uc
}

// Execute validates the driver-reported Trip Summary (distance/duration —
// never raw GPS points or a straight-line pickup/dropoff distance) before
// calling Pricing.
//
// Abnormal-completion guard (BRB-aligned with Grab/Be/Uber's own fraud
// checks): rejected ONLY when travelled_distance is below minDistanceKm
// AND travelled_duration is below minDurationMin — the AND (not OR) is what
// gives GPS-loss tolerance: a real trip's duration keeps advancing on the
// device clock regardless of GPS signal, so a momentary GPS glitch that
// under-reports distance near the end of a real trip will still have a
// legitimate duration and pass. A legitimate short trip (small distance,
// but the minutes to match) also passes, and is charged Pricing's own
// Minimum Fare floor — this use case never substitutes Base Fare or the
// original booking estimate.
func (uc *FinishTripUseCase) Execute(ctx context.Context, in FinishTripInput) (*FinishedTripResult, error) {
	if uc.idem != nil {
		key := "finish:" + in.TripID
		exists, err := uc.idem.Exists(ctx, key)
		if err != nil {
			return nil, domainerrors.Internal("idempotency check failed")
		}
		if exists {
			return nil, domainerrors.AlreadyExists("duplicate finish_trip request")
		}
	}

	if in.DistanceKM < uc.minDistanceKm && in.DurationMin < uc.minDurationMin {
		return nil, domainerrors.PreconditionFailed(
			"trip shows no meaningful movement or duration — cancel the trip instead of finishing it",
		).WithMeta("travelled_distance_km", in.DistanceKM).
			WithMeta("travelled_duration_min", in.DurationMin).
			WithMeta("min_distance_km", uc.minDistanceKm).
			WithMeta("min_duration_min", uc.minDurationMin)
	}

	fare, err := uc.pricing.CalculateFinalFare(ctx, in.VehicleType, in.DistanceKM, in.DurationMin)
	if err != nil {
		return nil, err
	}
	// Voucher detail comes from the gateway (Promotion Engine), never from
	// Pricing — merged onto the already-computed fare, not folded into any
	// fare math (fare.Total is untouched; see FinishTripInput's doc comment).
	if in.VoucherID != "" {
		fare.VoucherID = in.VoucherID
		fare.VoucherCode = in.VoucherCode
		fare.VoucherDiscountCents = in.VoucherDiscountCents
	}
	// Trip Summary rides along to Trip verbatim — the actual reported
	// distance/duration for this finish, never the original booking
	// estimate (early-passenger-stop trips are charged what actually
	// happened, per Execute's doc comment).
	fare.TravelledDistanceKm = in.DistanceKM
	fare.TravelledDurationMin = in.DurationMin
	_, err = uc.trip.CompleteTrip(ctx, in.TripID, *fare)
	if err != nil {
		return nil, err
	}
	if err := uc.trip.InitiatePayment(ctx, in.TripID); err != nil {
		return nil, err
	}

	if uc.idem != nil {
		_ = uc.idem.Record(ctx, "finish:"+in.TripID) // best-effort
	}

	return &FinishedTripResult{
		TripID:      in.TripID,
		Status:      "payment_pending",
		FinalFare:   fare.Total,
		Currency:    fare.CurrencyCode,
		VehicleType: in.VehicleType,
		DistanceKM:  in.DistanceKM,
		DurationMin: in.DurationMin,
	}, nil
}
