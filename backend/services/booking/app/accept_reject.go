package app

import "context"

// AcceptDispatchOfferUseCase delegates to the Dispatch service to accept an offer.
type AcceptDispatchOfferUseCase struct {
	dispatch DispatchClient
}

func NewAcceptDispatchOfferUseCase(dispatch DispatchClient) *AcceptDispatchOfferUseCase {
	return &AcceptDispatchOfferUseCase{dispatch: dispatch}
}

func (uc *AcceptDispatchOfferUseCase) Execute(ctx context.Context, tripID, driverID string) error {
	return uc.dispatch.AcceptTrip(ctx, tripID, driverID)
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
