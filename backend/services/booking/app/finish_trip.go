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
// Steps: pricing.CalculateFinalFare → trip.CompleteTrip
type FinishTripUseCase struct {
	pricing PricingClient
	trip    TripClient
	idem    IdempotencyStore // nil = no idempotency checking
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

	fare, err := uc.pricing.CalculateFinalFare(ctx, in.VehicleType, in.DistanceKM, in.DurationMin)
	if err != nil {
		return nil, err
	}
	_, err = uc.trip.CompleteTrip(ctx, in.TripID, fare.Total, fare.CurrencyCode)
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
