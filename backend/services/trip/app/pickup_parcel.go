package app

import (
	"context"
	"time"

	"github.com/fairride/shared/errors"
	"github.com/fairride/trip/domain/entity"
	"github.com/fairride/trip/domain/repository"
)

// PickupParcelInput is the input to PickupParcelUseCase.
type PickupParcelInput struct {
	TripID string
}

// PickupParcelUseCase transitions a Delivery from Accepted to
// ParcelPickedUp, and — reusing Trip's existing, unchanged Start() method —
// also transitions the linked Trip from DriverAssigned/DriverArrived to
// InProgress. This is the Delivery V1 Phase 4 equivalent of Ride's
// StartTripUseCase (docs/business/DELIVERY_V1_DESIGN.md Phần 3/6): the
// driver has now physically collected the package and is proceeding, the
// same real-world moment Ride's StartTrip represents for a passenger.
//
// By the time this is called, Trip.Status is expected to already be
// driver_arrived (via the existing, unchanged ArriveAtPickup/
// MarkDriverArrived flow reused as-is for Delivery — no new "ArrivedPickup"
// method or DeliveryStatus value was added; see delivery.go's doc comment).
type PickupParcelUseCase struct {
	tripRepo     repository.TripRepository
	deliveryRepo repository.DeliveryRepository
}

func NewPickupParcelUseCase(tripRepo repository.TripRepository, deliveryRepo repository.DeliveryRepository) *PickupParcelUseCase {
	return &PickupParcelUseCase{tripRepo: tripRepo, deliveryRepo: deliveryRepo}
}

func (uc *PickupParcelUseCase) Execute(ctx context.Context, in PickupParcelInput) (*entity.Delivery, error) {
	trip, delivery, err := loadDeliveryTrip(ctx, uc.tripRepo, uc.deliveryRepo, in.TripID)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	if err := delivery.MarkParcelPickedUp(now); err != nil {
		return nil, err
	}
	if err := trip.Start(now); err != nil {
		return nil, err
	}

	if err := uc.deliveryRepo.Save(ctx, delivery); err != nil {
		return nil, err
	}
	if err := uc.tripRepo.Save(ctx, trip); err != nil {
		return nil, err
	}
	return delivery, nil
}

// loadDeliveryTrip loads the Trip and its linked Delivery aggregate for
// tripID, rejecting Ride trips (a delivery-lifecycle action was called on a
// non-delivery trip) or a data-integrity gap (a delivery Trip with no
// DeliveryID). Shared by PickupParcelUseCase/StartDeliveryUseCase/
// CompleteDeliveryUseCase so the lookup logic exists in exactly one place.
func loadDeliveryTrip(
	ctx context.Context,
	tripRepo repository.TripRepository,
	deliveryRepo repository.DeliveryRepository,
	tripID string,
) (*entity.Trip, *entity.Delivery, error) {
	if tripID == "" {
		return nil, nil, errors.InvalidArgument("trip_id is required")
	}
	trip, err := tripRepo.FindByID(ctx, tripID)
	if err != nil {
		return nil, nil, err
	}
	if trip.TripType != entity.TripTypeDelivery {
		return nil, nil, errors.InvalidArgument("trip is not a delivery: " + tripID)
	}
	if trip.DeliveryID == "" {
		return nil, nil, errors.Internal("delivery trip has no linked delivery id: " + tripID)
	}
	delivery, err := deliveryRepo.FindByID(ctx, trip.DeliveryID)
	if err != nil {
		return nil, nil, err
	}
	return trip, delivery, nil
}
