package app

import (
	"context"

	domainerrors "github.com/fairride/shared/errors"
)

// GetDriverCurrentOfferUseCase retrieves the active trip offer for a driver,
// orchestrating dispatch (get offer) → trip (get addresses).
type GetDriverCurrentOfferUseCase struct {
	dispatch DispatchClient
	trip     TripClient
}

func NewGetDriverCurrentOfferUseCase(dispatch DispatchClient, trip TripClient) *GetDriverCurrentOfferUseCase {
	return &GetDriverCurrentOfferUseCase{dispatch: dispatch, trip: trip}
}

// Execute returns nil DriverOfferInfo (no error) when the driver has no active offer.
func (uc *GetDriverCurrentOfferUseCase) Execute(ctx context.Context, driverID string) (*DriverOfferInfo, error) {
	if driverID == "" {
		return nil, domainerrors.InvalidArgument("driver_id is required")
	}

	offer, err := uc.dispatch.GetDriverOffer(ctx, driverID)
	if err != nil {
		return nil, err
	}
	if offer == nil {
		return nil, nil
	}

	tripInfo, err := uc.trip.GetTrip(ctx, offer.TripID)
	if err != nil {
		return nil, err
	}

	return &DriverOfferInfo{
		TripID:         offer.TripID,
		PickupAddress:  tripInfo.PickupAddress,
		DropoffAddress: tripInfo.DropoffAddress,
		OfferExpiresAt: offer.OfferExpiresAt,
	}, nil
}
