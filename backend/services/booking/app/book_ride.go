package app

import (
	"context"

	domainerrors "github.com/fairride/shared/errors"
)

// BookRideInput is the input to BookRideUseCase.
//
// TripType is "ride" (default; empty also means ride) or "delivery" —
// Delivery V1 Phase 2 (docs/business/DELIVERY_V1_DESIGN.md). Ride and
// Delivery bookings share this single input struct and the single
// Execute pipeline below; the Pickup*Contact*/Receiver*/Package* fields
// are only read when TripType == "delivery".
type BookRideInput struct {
	RiderID        string
	PickupAddress  string
	DropoffAddress string
	PickupLat      float64
	PickupLon      float64
	// IdempotencyKey deduplicates retried requests when non-empty.
	// Callers that do not need idempotency may leave this blank.
	IdempotencyKey string

	TripType string

	// ServiceType is one of the Vehicle/Service Catalog's 4 tiers
	// (bike/bike_plus/car/car_xl) or empty (no service-type filter) —
	// forwarded to Dispatch so RequestDispatch can match the rider's
	// chosen tier against a driver's reported ServiceType.
	ServiceType string

	PickupContactName  string
	PickupContactPhone string
	ReceiverName       string
	ReceiverPhone      string
	PackageNote        string
	PackageValue       int64
	PackageWeightKg    float64
}

// BookRideResult holds the outcome of a BookRide call.
type BookRideResult struct {
	TripID string
	Status string // "searching"
	// DeliveryID is set only when the booking was a Delivery
	// (TripType == "delivery"); empty for Ride bookings.
	DeliveryID string
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

	created, err := uc.trip.CreateTrip(ctx, CreateTripParams{
		RiderID:            in.RiderID,
		PickupAddress:      in.PickupAddress,
		DropoffAddress:     in.DropoffAddress,
		TripType:           in.TripType,
		PickupContactName:  in.PickupContactName,
		PickupContactPhone: in.PickupContactPhone,
		ReceiverName:       in.ReceiverName,
		ReceiverPhone:      in.ReceiverPhone,
		PackageNote:        in.PackageNote,
		PackageValue:       in.PackageValue,
		PackageWeightKg:    in.PackageWeightKg,
	})
	if err != nil {
		return nil, err
	}
	tripID := created.TripID

	if err := uc.dispatch.RequestDispatch(ctx, tripID, in.RiderID, in.TripType, in.ServiceType, in.PickupLat, in.PickupLon); err != nil {
		// Compensate: cancel the trip so it doesn't stay orphaned in 'pending' state.
		// Best-effort — if cancellation also fails the trip will be GC'd by a background
		// reconciler; we still surface the original dispatch error to the caller.
		_ = uc.trip.CancelTrip(ctx, tripID, "dispatch_request_failed")
		return nil, err
	}

	if uc.idem != nil && in.IdempotencyKey != "" {
		_ = uc.idem.Record(ctx, in.IdempotencyKey) // best-effort; duplicate inserts are no-ops
	}

	return &BookRideResult{TripID: tripID, Status: "searching", DeliveryID: created.DeliveryID}, nil
}
