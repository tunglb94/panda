package app

import "context"

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
	Status      string // "completed"
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
}

func NewFinishTripUseCase(pricing PricingClient, trip TripClient) *FinishTripUseCase {
	return &FinishTripUseCase{pricing: pricing, trip: trip}
}

func (uc *FinishTripUseCase) Execute(ctx context.Context, in FinishTripInput) (*FinishedTripResult, error) {
	fare, err := uc.pricing.CalculateFinalFare(ctx, in.VehicleType, in.DistanceKM, in.DurationMin)
	if err != nil {
		return nil, err
	}
	tripInfo, err := uc.trip.CompleteTrip(ctx, in.TripID, fare.Total, fare.CurrencyCode)
	if err != nil {
		return nil, err
	}
	return &FinishedTripResult{
		TripID:      tripInfo.TripID,
		Status:      tripInfo.Status,
		FinalFare:   fare.Total,
		Currency:    fare.CurrencyCode,
		VehicleType: in.VehicleType,
		DistanceKM:  in.DistanceKM,
		DurationMin: in.DurationMin,
	}, nil
}
