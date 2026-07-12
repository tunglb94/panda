package app

import (
	"context"
	"time"

	"github.com/fairride/trip/domain/entity"
	"github.com/fairride/trip/domain/repository"
)

// StartDeliveryInput is the input to StartDeliveryUseCase.
type StartDeliveryInput struct {
	TripID string
}

// StartDeliveryUseCase transitions a Delivery from ParcelPickedUp to
// InDelivery — the driver is now en route to the drop-off address. Purely a
// Delivery-internal sub-status; Trip.Status is deliberately left untouched
// (already in_progress from the preceding PickupParcel step), matching
// docs/business/DELIVERY_V1_DESIGN.md Phần 3's "Delivering — Trip.Status
// không đổi" mapping.
type StartDeliveryUseCase struct {
	tripRepo     repository.TripRepository
	deliveryRepo repository.DeliveryRepository
}

func NewStartDeliveryUseCase(tripRepo repository.TripRepository, deliveryRepo repository.DeliveryRepository) *StartDeliveryUseCase {
	return &StartDeliveryUseCase{tripRepo: tripRepo, deliveryRepo: deliveryRepo}
}

func (uc *StartDeliveryUseCase) Execute(ctx context.Context, in StartDeliveryInput) (*entity.Delivery, error) {
	_, delivery, err := loadDeliveryTrip(ctx, uc.tripRepo, uc.deliveryRepo, in.TripID)
	if err != nil {
		return nil, err
	}

	if err := delivery.StartDelivery(time.Now().UTC()); err != nil {
		return nil, err
	}
	if err := uc.deliveryRepo.Save(ctx, delivery); err != nil {
		return nil, err
	}
	return delivery, nil
}
