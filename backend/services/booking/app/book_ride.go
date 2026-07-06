package app

import (
	"context"

	domainerrors "github.com/fairride/shared/errors"
)

// BookRideInput is the input to BookRideUseCase.
type BookRideInput struct {
	RiderID        string
	PickupAddress  string
	DropoffAddress string
	PickupLat      float64
	PickupLon      float64
	// IdempotencyKey deduplicates retried requests when non-empty.
	// Callers that do not need idempotency may leave this blank.
	IdempotencyKey string
}

// BookRideResult holds the outcome of a BookRide call.
type BookRideResult struct {
	TripID string
	Status string // "searching"
}

// BookRideUseCase creates a trip and immediately requests dispatch.
// Steps: trip.CreateTrip → dispatch.RequestDispatch
// Compensation: if dispatch fails, trip.CancelTrip is called to avoid orphaned trips.
type BookRideUseCase struct {
	trip     TripClient
	dispatch DispatchClient
	idem     IdempotencyStore // nil = no idempotency checking
}

func NewBookRideUseCase(trip TripClient, dispatch DispatchClient) *BookRideUseCase {
	return &BookRideUseCase{trip: trip, dispatch: dispatch}
}

// WithIdempotency attaches an idempotency store. Keys are checked/recorded when
// BookRideInput.IdempotencyKey is non-empty.
func (uc *BookRideUseCase) WithIdempotency(store IdempotencyStore) *BookRideUseCase {
	uc.idem = store
	return uc
}

func (uc *BookRideUseCase) Execute(ctx context.Context, in BookRideInput) (*BookRideResult, error) {
	if uc.idem != nil && in.IdempotencyKey != "" {
		exists, err := uc.idem.Exists(ctx, in.IdempotencyKey)
		if err != nil {
			return nil, domainerrors.Internal("idempotency check failed")
		}
		if exists {
			return nil, domainerrors.AlreadyExists("duplicate book_ride request")
		}
	}

	tripID, err := uc.trip.CreateTrip(ctx, in.RiderID, in.PickupAddress, in.DropoffAddress)
	if err != nil {
		return nil, err
	}

	if err := uc.dispatch.RequestDispatch(ctx, tripID, in.RiderID, in.PickupLat, in.PickupLon); err != nil {
		// Compensate: cancel the trip so it doesn't stay orphaned in 'pending' state.
		// Best-effort — if cancellation also fails the trip will be GC'd by a background
		// reconciler; we still surface the original dispatch error to the caller.
		_ = uc.trip.CancelTrip(ctx, tripID, "dispatch_request_failed")
		return nil, err
	}

	if uc.idem != nil && in.IdempotencyKey != "" {
		_ = uc.idem.Record(ctx, in.IdempotencyKey) // best-effort; duplicate inserts are no-ops
	}

	return &BookRideResult{TripID: tripID, Status: "searching"}, nil
}
