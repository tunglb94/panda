package app

import "context"

// BookingDetails combines trip and dispatch state for a given trip.
type BookingDetails struct {
	TripID         string
	TripStatus     string
	RiderID        string
	DriverID       string
	PickupAddress  string
	DropoffAddress string
	DispatchStatus string
	FinalFare      int64
	Currency       string
}

// GetBookingDetailsUseCase fetches trip state and dispatch status in parallel context.
// If dispatch info is unavailable (job not yet created), DispatchStatus is "unknown".
type GetBookingDetailsUseCase struct {
	trip     TripClient
	dispatch DispatchClient
}

func NewGetBookingDetailsUseCase(trip TripClient, dispatch DispatchClient) *GetBookingDetailsUseCase {
	return &GetBookingDetailsUseCase{trip: trip, dispatch: dispatch}
}

func (uc *GetBookingDetailsUseCase) Execute(ctx context.Context, tripID string) (*BookingDetails, error) {
	tripInfo, err := uc.trip.GetTrip(ctx, tripID)
	if err != nil {
		return nil, err
	}

	dispatchStatus := "unknown"
	if di, err := uc.dispatch.GetDispatchStatus(ctx, tripID); err == nil {
		dispatchStatus = di.Status
	}

	return &BookingDetails{
		TripID:         tripInfo.TripID,
		TripStatus:     tripInfo.Status,
		RiderID:        tripInfo.RiderID,
		DriverID:       tripInfo.DriverID,
		PickupAddress:  tripInfo.PickupAddress,
		DropoffAddress: tripInfo.DropoffAddress,
		DispatchStatus: dispatchStatus,
		FinalFare:      tripInfo.FinalFareTotal,
		Currency:       tripInfo.FareCurrency,
	}, nil
}
