package app

import (
	"context"
	"time"

	"github.com/fairride/shared/errors"
	"github.com/fairride/trip/domain/entity"
	"github.com/fairride/trip/domain/repository"
)

// AcceptDeliveryInput is the input to AcceptDeliveryUseCase.
type AcceptDeliveryInput struct {
	TripID string
}

// AcceptDeliveryUseCase transitions a Delivery from Created to Accepted —
// production hardening P0-1. Booking's AcceptDispatchOfferUseCase records
// driver acceptance by writing directly to the (Postgres-backed) trips
// table within Dispatch's transaction; it never reaches the Delivery
// aggregate, which only ever lives in the Trip service process's own
// repository (in-memory today — see cmd/server/main.go). Without this
// call, Delivery.Status is stuck at Created forever, and PickupParcel
// (which requires Accepted) always fails — this is the exact bug
// production hardening P0-1 fixes. Booking calls this RPC unconditionally
// right after a successful accept, for every trip, Ride or Delivery, so
// both "not a delivery" and "already accepted" (a duplicate/retried
// accept request) must be safe no-ops, not errors — idempotent by design,
// not by accident.
type AcceptDeliveryUseCase struct {
	tripRepo     repository.TripRepository
	deliveryRepo repository.DeliveryRepository
}

func NewAcceptDeliveryUseCase(tripRepo repository.TripRepository, deliveryRepo repository.DeliveryRepository) *AcceptDeliveryUseCase {
	return &AcceptDeliveryUseCase{tripRepo: tripRepo, deliveryRepo: deliveryRepo}
}

// Execute returns the Delivery it transitioned, or nil if there was
// nothing to do (not a delivery trip, or the delivery was already past
// Created). A nil result is not an error.
func (uc *AcceptDeliveryUseCase) Execute(ctx context.Context, in AcceptDeliveryInput) (*entity.Delivery, error) {
	if in.TripID == "" {
		return nil, errors.InvalidArgument("trip_id is required")
	}
	trip, err := uc.tripRepo.FindByID(ctx, in.TripID)
	if err != nil {
		return nil, err
	}
	if trip.TripType != entity.TripTypeDelivery {
		return nil, nil // Ride trip — nothing to accept, not an error.
	}
	if trip.DeliveryID == "" {
		return nil, errors.Internal("delivery trip has no linked delivery id: " + in.TripID)
	}
	delivery, err := uc.deliveryRepo.FindByID(ctx, trip.DeliveryID)
	if err != nil {
		return nil, err
	}
	if delivery.Status != entity.DeliveryStatusCreated {
		// Already accepted (or further along) — a retried/duplicate accept
		// request. Idempotent no-op: return the current state, don't error.
		return delivery, nil
	}

	now := time.Now().UTC()
	if err := delivery.AcceptByDriver(now); err != nil {
		return nil, err
	}
	if err := uc.deliveryRepo.Save(ctx, delivery); err != nil {
		return nil, err
	}
	return delivery, nil
}
