package app

import (
	"context"

	domainerrors "github.com/fairride/shared/errors"
)

// ArriveAtPickupUseCase marks the driver as arrived at the pickup location.
// Transitions the trip from driver_assigned → driver_arrived.
type ArriveAtPickupUseCase struct {
	trip TripClient
}

func NewArriveAtPickupUseCase(trip TripClient) *ArriveAtPickupUseCase {
	return &ArriveAtPickupUseCase{trip: trip}
}

func (uc *ArriveAtPickupUseCase) Execute(ctx context.Context, tripID string) error {
	if tripID == "" {
		return domainerrors.InvalidArgument("trip_id must not be empty")
	}
	return uc.trip.MarkDriverArrived(ctx, tripID)
}
