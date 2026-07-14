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

	// Commission detail computed by Pricing V3 at Complete() time — Settlement
	// must read these instead of inventing its own commission rate.
	// HasCommissionDetail is false (and the *Cents/*Rate fields are 0) when
	// Pricing was running V2, which does not compute this detail at all.
	HasCommissionDetail  bool
	CommissionCents      int64
	DriverIncomeCents    int64
	VoucherDiscountCents int64
	CommissionRate       float64

	// VoucherID/VoucherCode identify the voucher redeemed against this trip
	// (Promotion Engine is the source of truth — see gateway's
	// BookingHandler.FinishTrip, which resolves the reservation and passes
	// it through here). Empty means no voucher was applied. Independent of
	// HasCommissionDetail — a trip can have a voucher with or without
	// Pricing V3 commission detail also being present.
	VoucherID   string
	VoucherCode string

	// ArrivedAt/StartedAt are server-stamped (MarkDriverArrived/Start) —
	// nil until reached. The only GPS-independent, server-authoritative
	// timestamps in the trip lifecycle; WaitingDurationMin is derived from
	// them, never from client input.
	ArrivedAt *time.Time
	StartedAt *time.Time

	// Trip Summary — the business-data record of what actually happened,
	// built by Complete(). Ride Lifecycle Fare Validation hardening: Pricing
	// reads only these fields for the final fare, never raw GPS pings or a
	// straight-line pickup/dropoff distance. TravelledDistanceKm/
	// TravelledDurationMin are sourced from the driver app's on-device GPS
	// tracking (industry-standard — Grab/Be/Uber all compute actual fare
	// from device-reported driven distance, not server-side re-derivation);
	// WaitingDurationMin is computed purely from ArrivedAt/StartedAt.
	// TollFeeCents/ExtraFeeCents are reserved for a future fee-entry source
	// (Known Gap — always 0 today).
	TravelledDistanceKm  float64
	TravelledDurationMin float64
	WaitingDurationMin   float64
	TollFeeCents         int64
	ExtraFeeCents        int64
}

// TripSummary is the actual-trip business record passed to Complete().
// WaitingDurationMin is intentionally absent here — Trip derives it itself
// from ArrivedAt/StartedAt (see Complete), the only server-authoritative
// timestamps available, rather than trusting a client-supplied value.
type TripSummary struct {
	TravelledDistanceKm  float64
	TravelledDurationMin float64
	TollFeeCents         int64
	ExtraFeeCents        int64
}

// CompleteFinancials carries the commission detail Pricing V3 computed for
// this trip's final fare (see pricing/app/feature_flag.go's
// CalculateFinalDetailed), plus the Promotion Engine's voucher redemption
// detail (independent of Pricing V3 — see VoucherID's doc comment on Trip).
// Zero value means "no detail available" (Pricing running V2 / no voucher)
// — HasCommissionDetail distinguishes that from a real 0 for the commission
// fields specifically; VoucherID == "" is its own independent "no voucher" signal.
type CompleteFinancials struct {
	HasCommissionDetail  bool
	CommissionCents      int64
	DriverIncomeCents    int64
	VoucherDiscountCents int64
	CommissionRate       float64
	VoucherID            string
	VoucherCode          string
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
// arrivedAt/startedAt may be nil (not yet reached); summary is the
// persisted Trip Summary (zero value before Complete()).
func ReconstituteTrip(
	tripID, riderID, driverID string,
	status TripStatus,
	pickupAddress, dropoffAddress, cancellationReason string,
	finalFareTotal int64, fareCurrency string,
	paymentMethod string,
	createdAt, updatedAt time.Time,
	fin CompleteFinancials,
	arrivedAt, startedAt *time.Time,
	waitingDurationMin float64,
	summary TripSummary,
) *Trip {
	return &Trip{
		TripID:               tripID,
		RiderID:              riderID,
		DriverID:             driverID,
		Status:               status,
		PickupAddress:        pickupAddress,
		DropoffAddress:       dropoffAddress,
		CancellationReason:   cancellationReason,
		FinalFareTotal:       finalFareTotal,
		FareCurrency:         fareCurrency,
		PaymentMethod:        paymentMethod,
		CreatedAt:            createdAt,
		UpdatedAt:            updatedAt,
		HasCommissionDetail:  fin.HasCommissionDetail,
		CommissionCents:      fin.CommissionCents,
		DriverIncomeCents:    fin.DriverIncomeCents,
		VoucherDiscountCents: fin.VoucherDiscountCents,
		CommissionRate:       fin.CommissionRate,
		VoucherID:            fin.VoucherID,
		VoucherCode:          fin.VoucherCode,
		ArrivedAt:            arrivedAt,
		StartedAt:            startedAt,
		TravelledDistanceKm:  summary.TravelledDistanceKm,
		TravelledDurationMin: summary.TravelledDurationMin,
		WaitingDurationMin:   waitingDurationMin,
		TollFeeCents:         summary.TollFeeCents,
		ExtraFeeCents:        summary.ExtraFeeCents,
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
	arrivedAt := now
	t.ArrivedAt = &arrivedAt
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
	startedAt := now
	t.StartedAt = &startedAt
	t.UpdatedAt = now
	return nil
}

// Complete transitions the trip from InProgress to Completed, records the
// fare, and persists the Trip Summary. Returns CodePreconditionFailed if the
// trip is not InProgress.
//
// Ride Lifecycle Fare Validation: abnormal-completion (no-movement fraud)
// validation happens upstream, in booking's FinishTripUseCase, before
// Pricing is even called — Complete trusts the summary it's given here and
// never re-derives distance from coordinates (never straight-line pickup/
// dropoff). WaitingDurationMin is always computed from ArrivedAt/StartedAt
// — the only GPS-independent, server-authoritative timestamps — never taken
// from the caller-supplied summary.
func (t *Trip) Complete(finalFareTotal int64, fareCurrency string, fin CompleteFinancials, summary TripSummary, now time.Time) error {
	if t.Status != StatusInProgress {
		return errors.PreconditionFailed("trip cannot be completed from status: " + string(t.Status))
	}
	t.Status = StatusCompleted
	t.FinalFareTotal = finalFareTotal
	t.FareCurrency = fareCurrency
	t.HasCommissionDetail = fin.HasCommissionDetail
	t.CommissionCents = fin.CommissionCents
	t.DriverIncomeCents = fin.DriverIncomeCents
	t.VoucherDiscountCents = fin.VoucherDiscountCents
	t.CommissionRate = fin.CommissionRate
	t.VoucherID = fin.VoucherID
	t.VoucherCode = fin.VoucherCode
	t.TravelledDistanceKm = summary.TravelledDistanceKm
	t.TravelledDurationMin = summary.TravelledDurationMin
	t.TollFeeCents = summary.TollFeeCents
	t.ExtraFeeCents = summary.ExtraFeeCents
	t.WaitingDurationMin = t.computeWaitingDurationMin()
	t.UpdatedAt = now
	return nil
}

// computeWaitingDurationMin derives wait-at-pickup time from the trip's own
// server-stamped timestamps — 0 if either is missing (e.g. a trip Started
// directly from DriverAssigned without a recorded arrival) or non-positive.
func (t *Trip) computeWaitingDurationMin() float64 {
	if t.ArrivedAt == nil || t.StartedAt == nil {
		return 0
	}
	d := t.StartedAt.Sub(*t.ArrivedAt).Minutes()
	if d < 0 {
		return 0
	}
	return d
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
