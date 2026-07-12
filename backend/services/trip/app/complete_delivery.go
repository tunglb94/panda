package app

import (
	"context"
	"time"

	"github.com/fairride/trip/domain/entity"
	"github.com/fairride/trip/domain/repository"
)

// CompleteDeliveryInput is the input to CompleteDeliveryUseCase.
type CompleteDeliveryInput struct {
	TripID string
}

// CompleteDeliveryUseCase marks a Delivery as delivered and immediately
// completed: InDelivery → Delivered → Completed, chaining MarkDelivered()
// then CompleteDelivery() in one call — the same "two domain transitions,
// one caller-facing action" pattern PayTripUseCase already uses for
// MarkPaid→Settle. The driver taps "Delivered" once; the backend records
// both the delivery event and the terminal completion.
//
// Trip.Status is deliberately NOT advanced to completed here. Ride's
// Trip.Complete(finalFareTotal, fareCurrency, now) requires a real computed
// fare (via Pricing.CalculateFinalFare, see Booking's FinishTripUseCase);
// Pricing is out of scope for Delivery V1 Phase 4 ("Không sửa: Pricing"),
// so there is no fare value this use case could honestly supply. Wiring
// Trip.Complete() for Delivery is deferred to the phase that adds
// Delivery↔Pricing integration.
type CompleteDeliveryUseCase struct {
	tripRepo     repository.TripRepository
	deliveryRepo repository.DeliveryRepository
}

func NewCompleteDeliveryUseCase(tripRepo repository.TripRepository, deliveryRepo repository.DeliveryRepository) *CompleteDeliveryUseCase {
	return &CompleteDeliveryUseCase{tripRepo: tripRepo, deliveryRepo: deliveryRepo}
}

func (uc *CompleteDeliveryUseCase) Execute(ctx context.Context, in CompleteDeliveryInput) (*entity.Delivery, error) {
	_, delivery, err := loadDeliveryTrip(ctx, uc.tripRepo, uc.deliveryRepo, in.TripID)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	if err := delivery.MarkDelivered(now); err != nil {
		return nil, err
	}
	if err := uc.deliveryRepo.Save(ctx, delivery); err != nil {
		return nil, err
	}
	if err := delivery.CompleteDelivery(now); err != nil {
		return nil, err
	}
	if err := uc.deliveryRepo.Save(ctx, delivery); err != nil {
		return nil, err
	}
	return delivery, nil
}
