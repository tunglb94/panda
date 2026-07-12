package app

import (
	"context"

	domainerrors "github.com/fairride/shared/errors"
)

// AcceptDispatchOfferUseCase delegates to the Dispatch service to accept an
// offer, then — production hardening P0-1 — tells the Trip service to
// accept the linked Delivery, if any. Dispatch's AcceptTrip only updates
// the (Postgres-backed) trips table and the dispatch job; it never reaches
// the Delivery aggregate, which lives only in the Trip service process's
// own repository. Without the second call, Delivery.Status stays stuck at
// Created forever and PickupParcel always fails — the exact P0-1 bug.
// trip is optional (nil-safe): if unset, the Delivery-acceptance step is
// skipped and only Dispatch is notified, same as before this fix — kept
// nil-safe so every existing test/call site that doesn't care about
// Delivery keeps working unchanged.
type AcceptDispatchOfferUseCase struct {
	dispatch DispatchClient
	trip     TripClient
	idem     IdempotencyStore // nil = no idempotency checking
}

func NewAcceptDispatchOfferUseCase(dispatch DispatchClient, trip TripClient) *AcceptDispatchOfferUseCase {
	return &AcceptDispatchOfferUseCase{dispatch: dispatch, trip: trip}
}

// WithIdempotency attaches an idempotency store. The natural key is "accept:" + tripID,
// preventing a duplicate accept from reaching the dispatch service.
func (uc *AcceptDispatchOfferUseCase) WithIdempotency(store IdempotencyStore) *AcceptDispatchOfferUseCase {
	uc.idem = store
	return uc
}

func (uc *AcceptDispatchOfferUseCase) Execute(ctx context.Context, tripID, driverID string) error {
	if uc.idem != nil {
		key := "accept:" + tripID
		exists, err := uc.idem.Exists(ctx, key)
		if err != nil {
			return domainerrors.Internal("idempotency check failed")
		}
		if exists {
			return domainerrors.AlreadyExists("duplicate accept_dispatch_offer request")
		}
	}

	if err := uc.dispatch.AcceptTrip(ctx, tripID, driverID); err != nil {
		return err
	}

	// P0-1: accept the linked Delivery, if any — see the type doc comment.
	// Propagates a real failure (the driver's accept must surface it, not
	// silently leave the delivery stuck at Created); a Ride trip is a
	// harmless no-op on the Trip-service side, not an error here.
	if uc.trip != nil {
		if err := uc.trip.AcceptDelivery(ctx, tripID); err != nil {
			return err
		}
	}

	if uc.idem != nil {
		_ = uc.idem.Record(ctx, "accept:"+tripID) // best-effort
	}
	return nil
}

// RejectDispatchOfferUseCase delegates to the Dispatch service to reject an offer.
type RejectDispatchOfferUseCase struct {
	dispatch DispatchClient
}

func NewRejectDispatchOfferUseCase(dispatch DispatchClient) *RejectDispatchOfferUseCase {
	return &RejectDispatchOfferUseCase{dispatch: dispatch}
}

func (uc *RejectDispatchOfferUseCase) Execute(ctx context.Context, tripID, driverID string) error {
	return uc.dispatch.RejectTrip(ctx, tripID, driverID)
}
