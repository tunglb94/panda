package app

import (
	"context"

	domainerrors "github.com/fairride/shared/errors"
)

// AcceptDispatchOfferUseCase delegates to the Dispatch service to accept an offer.
type AcceptDispatchOfferUseCase struct {
	dispatch DispatchClient
	idem     IdempotencyStore // nil = no idempotency checking
}

func NewAcceptDispatchOfferUseCase(dispatch DispatchClient) *AcceptDispatchOfferUseCase {
	return &AcceptDispatchOfferUseCase{dispatch: dispatch}
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
