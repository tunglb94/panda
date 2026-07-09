package app

import "context"

// CancelRideUseCase cancels a trip on behalf of the rider.
// Delegates to TripClient.CancelTrip with reason "rider_cancelled".
type CancelRideUseCase struct {
	trip TripClient
}

func NewCancelRideUseCase(trip TripClient) *CancelRideUseCase {
	return &CancelRideUseCase{trip: trip}
}

func (uc *CancelRideUseCase) Execute(ctx context.Context, tripID string) error {
	return uc.trip.CancelTrip(ctx, tripID, "rider_cancelled")
}
