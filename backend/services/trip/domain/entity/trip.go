package entity

import (
	"strings"
	"time"

	"github.com/fairride/shared/errors"
)

// TripStatus represents the lifecycle stage of a trip.
type TripStatus string

const (
	StatusPending        TripStatus = "pending"
	StatusSearching      TripStatus = "searching"
	StatusDriverAssigned TripStatus = "driver_assigned"
	StatusDriverArrived  TripStatus = "driver_arrived"
	StatusInProgress     TripStatus = "in_progress"
	StatusCompleted      TripStatus = "completed"
	StatusCancelled      TripStatus = "cancelled"
	StatusPaymentPending TripStatus = "payment_pending"
	StatusPaymentSuccess TripStatus = "payment_success"
	StatusSettled        TripStatus = "settled"
)

// cancellableStatuses are the statuses from which a trip may be cancelled.
var cancellableStatuses = map[TripStatus]bool{
	StatusPending:        true,
	StatusSearching:      true,
	StatusDriverAssigned: true,
	StatusDriverArrived:  true,
}

// TripType discriminates what a Trip represents. See
// docs/business/DELIVERY_V1_DESIGN.md Phần 1/2 — this is the single
// discriminator the rest of the system needs to distinguish Ride from
// Delivery; every other Delivery-specific field lives on the separate
// Delivery aggregate (delivery.go), not on Trip itself.
type TripType string

const (
	TripTypeRide     TripType = "ride"
	TripTypeDelivery TripType = "delivery"
)

// Trip is the aggregate root for a FAIRRIDE ride or delivery request.
// DriverID is empty until a driver is assigned.
// CancellationReason is empty unless the trip is Cancelled.
// FinalFareTotal and FareCurrency are set when the trip is Completed.
// PaymentMethod is set when the rider pays (e.g. "cash", "wallet").
// TripType is TripTypeRide for every trip created via NewTrip; DeliveryID is
// empty ("nil") for Ride trips and only set for trips created via
// NewDeliveryTrip (docs/business/DELIVERY_V1_DESIGN.md Phần 5 "Mapping Trip").
type Trip struct {
	TripID             string
	RiderID            string
	DriverID           string
	Status             TripStatus
	PickupAddress      string
	DropoffAddress     string
	CancellationReason string
	FinalFareTotal     int64  // 0 until Completed; smallest currency unit
	FareCurrency       string // e.g. "USD"; empty until Completed
	PaymentMethod      string // empty until paid
	TripType           TripType
	DeliveryID         string // empty for Ride trips ("nil")
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

// NewTrip creates a validated Ride Trip in StatusPending. Unchanged from
// before Delivery was introduced: same signature, same validation, same
// behavior for every existing caller — it now simply also stamps
// TripType=TripTypeRide (DeliveryID stays "", i.e. nil) on the returned Trip.
// riderID, pickupAddress, and dropoffAddress are required.
func NewTrip(tripID, riderID, pickupAddress, dropoffAddress string, now time.Time) (*Trip, error) {
	if tripID == "" {
		return nil, errors.InvalidArgument("trip id must not be empty")
	}
	if riderID == "" {
		return nil, errors.InvalidArgument("rider id must not be empty")
	}
	if strings.TrimSpace(pickupAddress) == "" {
		return nil, errors.InvalidArgument("pickup address must not be empty")
	}
	if strings.TrimSpace(dropoffAddress) == "" {
		return nil, errors.InvalidArgument("dropoff address must not be empty")
	}
	return &Trip{
		TripID:         tripID,
		RiderID:        riderID,
		Status:         StatusPending,
		PickupAddress:  pickupAddress,
		DropoffAddress: dropoffAddress,
		TripType:       TripTypeRide,
		CreatedAt:      now,
		UpdatedAt:      now,
	}, nil
}

// NewDeliveryTrip creates a validated Delivery Trip in StatusPending. It is
// additive alongside NewTrip (Phase 1 of docs/business/DELIVERY_V1_DESIGN.md)
// — no existing caller of NewTrip is affected. deliveryID must reference an
// already-created Delivery aggregate (see delivery.go); riderID here plays
// the role of "Sender" for a delivery (Phần 2 of the design doc).
func NewDeliveryTrip(tripID, riderID, pickupAddress, dropoffAddress, deliveryID string, now time.Time) (*Trip, error) {
	if deliveryID == "" {
		return nil, errors.InvalidArgument("delivery id must not be empty")
	}
	trip, err := NewTrip(tripID, riderID, pickupAddress, dropoffAddress, now)
	if err != nil {
		return nil, err
	}
	trip.TripType = TripTypeDelivery
	trip.DeliveryID = deliveryID
	return trip, nil
}

// ReconstituteTrip rebuilds a Trip from a persistence record. No validation.
func ReconstituteTrip(
	tripID, riderID, driverID string,
	status TripStatus,
	pickupAddress, dropoffAddress, cancellationReason string,
	finalFareTotal int64, fareCurrency string,
	paymentMethod string,
	createdAt, updatedAt time.Time,
) *Trip {
	return &Trip{
		TripID:             tripID,
		RiderID:            riderID,
		DriverID:           driverID,
		Status:             status,
		PickupAddress:      pickupAddress,
		DropoffAddress:     dropoffAddress,
		CancellationReason: cancellationReason,
		FinalFareTotal:     finalFareTotal,
		FareCurrency:       fareCurrency,
		PaymentMethod:      paymentMethod,
		CreatedAt:          createdAt,
		UpdatedAt:          updatedAt,
	}
}

// Cancel transitions the trip to StatusCancelled.
// Returns CodePreconditionFailed if the trip is InProgress, Completed, or
// already Cancelled.
func (t *Trip) Cancel(reason string, now time.Time) error {
	if !cancellableStatuses[t.Status] {
		return errors.PreconditionFailed("trip cannot be cancelled from status: " + string(t.Status))
	}
	t.Status = StatusCancelled
	t.CancellationReason = reason
	t.UpdatedAt = now
	return nil
}

// IsCancellable reports whether the trip can currently be cancelled.
func (t *Trip) IsCancellable() bool {
	return cancellableStatuses[t.Status]
}

// MarkDriverArrived transitions the trip from DriverAssigned to DriverArrived.
// Returns CodePreconditionFailed if the trip is not in DriverAssigned status.
func (t *Trip) MarkDriverArrived(now time.Time) error {
	if t.Status != StatusDriverAssigned {
		return errors.PreconditionFailed("driver arrived cannot be set from status: " + string(t.Status))
	}
	t.Status = StatusDriverArrived
	t.UpdatedAt = now
	return nil
}

// Start transitions the trip from DriverAssigned or DriverArrived to InProgress.
// Returns CodePreconditionFailed if the current status is not a valid start point.
func (t *Trip) Start(now time.Time) error {
	if t.Status != StatusDriverAssigned && t.Status != StatusDriverArrived {
		return errors.PreconditionFailed("trip cannot be started from status: " + string(t.Status))
	}
	t.Status = StatusInProgress
	t.UpdatedAt = now
	return nil
}

// Complete transitions the trip from InProgress to Completed and records the fare.
// Returns CodePreconditionFailed if the trip is not InProgress.
func (t *Trip) Complete(finalFareTotal int64, fareCurrency string, now time.Time) error {
	if t.Status != StatusInProgress {
		return errors.PreconditionFailed("trip cannot be completed from status: " + string(t.Status))
	}
	t.Status = StatusCompleted
	t.FinalFareTotal = finalFareTotal
	t.FareCurrency = fareCurrency
	t.UpdatedAt = now
	return nil
}

// InitiatePayment transitions the trip from Completed to PaymentPending.
// Called immediately after Complete so the rider is prompted to pay.
func (t *Trip) InitiatePayment(now time.Time) error {
	if t.Status != StatusCompleted {
		return errors.PreconditionFailed("payment cannot be initiated from status: " + string(t.Status))
	}
	t.Status = StatusPaymentPending
	t.UpdatedAt = now
	return nil
}

// MarkPaid transitions the trip from PaymentPending to PaymentSuccess.
func (t *Trip) MarkPaid(method string, now time.Time) error {
	if t.Status != StatusPaymentPending {
		return errors.PreconditionFailed("trip cannot be marked paid from status: " + string(t.Status))
	}
	t.Status = StatusPaymentSuccess
	t.PaymentMethod = method
	t.UpdatedAt = now
	return nil
}

// Settle transitions the trip from PaymentSuccess to Settled.
func (t *Trip) Settle(now time.Time) error {
	if t.Status != StatusPaymentSuccess {
		return errors.PreconditionFailed("trip cannot be settled from status: " + string(t.Status))
	}
	t.Status = StatusSettled
	t.UpdatedAt = now
	return nil
}
