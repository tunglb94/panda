package app

import (
	"context"

	domainerrors "github.com/fairride/shared/errors"
)

// ListRiderTripsUseCase returns all trips for a rider, newest first.
type ListRiderTripsUseCase struct {
	trip TripClient
}

func NewListRiderTripsUseCase(trip TripClient) *ListRiderTripsUseCase {
	return &ListRiderTripsUseCase{trip: trip}
}

func (uc *ListRiderTripsUseCase) Execute(ctx context.Context, riderID string) ([]TripSummary, error) {
	if riderID == "" {
		return nil, domainerrors.InvalidArgument("rider_id must not be empty")
	}
	return uc.trip.ListByRider(ctx, riderID)
}

// ListDriverTripsUseCase returns all trips for a driver, newest first.
type ListDriverTripsUseCase struct {
	trip TripClient
}

func NewListDriverTripsUseCase(trip TripClient) *ListDriverTripsUseCase {
	return &ListDriverTripsUseCase{trip: trip}
}

func (uc *ListDriverTripsUseCase) Execute(ctx context.Context, driverID string) ([]TripSummary, error) {
	if driverID == "" {
		return nil, domainerrors.InvalidArgument("driver_id must not be empty")
	}
	return uc.trip.ListByDriver(ctx, driverID)
}
