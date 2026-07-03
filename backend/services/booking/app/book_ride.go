package app

import "context"

// BookRideInput is the input to BookRideUseCase.
type BookRideInput struct {
	RiderID        string
	PickupAddress  string
	DropoffAddress string
	PickupLat      float64
	PickupLon      float64
}

// BookRideResult holds the outcome of a BookRide call.
type BookRideResult struct {
	TripID string
	Status string // "searching"
}

// BookRideUseCase creates a trip and immediately requests dispatch.
// Steps: trip.CreateTrip → dispatch.RequestDispatch
type BookRideUseCase struct {
	trip     TripClient
	dispatch DispatchClient
}

func NewBookRideUseCase(trip TripClient, dispatch DispatchClient) *BookRideUseCase {
	return &BookRideUseCase{trip: trip, dispatch: dispatch}
}

func (uc *BookRideUseCase) Execute(ctx context.Context, in BookRideInput) (*BookRideResult, error) {
	tripID, err := uc.trip.CreateTrip(ctx, in.RiderID, in.PickupAddress, in.DropoffAddress)
	if err != nil {
		return nil, err
	}
	if err := uc.dispatch.RequestDispatch(ctx, tripID, in.RiderID, in.PickupLat, in.PickupLon); err != nil {
		return nil, err
	}
	return &BookRideResult{TripID: tripID, Status: "searching"}, nil
}
