package app

import (
	"context"

	domainerrors "github.com/fairride/shared/errors"
)

// PayRideInput is the input to PayRideUseCase.
type PayRideInput struct {
	TripID        string
	PaymentMethod string // "cash" or "wallet"; defaults to "cash"
}

// PayRideResult holds the settled trip state.
type PayRideResult struct {
	TripID    string
	Status    string // "settled"
	FinalFare int64
	Currency  string
}

// PayRideUseCase processes mock payment for a trip in payment_pending status.
// Transitions: payment_pending → payment_success → settled.
type PayRideUseCase struct {
	trip TripClient
}

func NewPayRideUseCase(trip TripClient) *PayRideUseCase {
	return &PayRideUseCase{trip: trip}
}

func (uc *PayRideUseCase) Execute(ctx context.Context, in PayRideInput) (*PayRideResult, error) {
	if in.TripID == "" {
		return nil, domainerrors.InvalidArgument("trip_id is required")
	}
	method := in.PaymentMethod
	if method == "" {
		method = "cash"
	}
	info, err := uc.trip.PayTrip(ctx, in.TripID, method)
	if err != nil {
		return nil, err
	}
	return &PayRideResult{
		TripID:    info.TripID,
		Status:    info.Status,
		FinalFare: info.FinalFareTotal,
		Currency:  info.FareCurrency,
	}, nil
}
