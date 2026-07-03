package app

import "context"

// StartTripUseCase calls the Trip service to begin a ride.
type StartTripUseCase struct {
	trip TripClient
}

func NewStartTripUseCase(trip TripClient) *StartTripUseCase {
	return &StartTripUseCase{trip: trip}
}

func (uc *StartTripUseCase) Execute(ctx context.Context, tripID string) error {
	return uc.trip.StartTrip(ctx, tripID)
}
