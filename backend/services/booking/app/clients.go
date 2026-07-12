// Package app contains the booking orchestration use cases.
// It depends only on the interfaces defined here — never on concrete gRPC clients —
// so every use case is testable without a running service.
package app

import (
	"context"
	"time"
)

// TripInfo is a lightweight view of a trip returned by TripClient.
type TripInfo struct {
	TripID             string
	RiderID            string
	DriverID           string
	Status             string
	PickupAddress      string
	DropoffAddress     string
	FinalFareTotal     int64
	FareCurrency       string
	CancellationReason string
}

// DispatchInfo is a lightweight view of a dispatch job returned by DispatchClient.
type DispatchInfo struct {
	TripID           string
	Status           string
	AssignedDriverID string
}

// FareInfo is the fare result returned by PricingClient.
type FareInfo struct {
	Total        int64
	CurrencyCode string
}

// TripSummary is a lightweight trip view used in list operations.
type TripSummary struct {
	TripID         string
	Status         string
	PickupAddress  string
	DropoffAddress string
	FinalFare      int64
	Currency       string
	CreatedAt      time.Time
}

// CreateTripParams is the input to TripClient.CreateTrip. TripType is
// "ride" (default; empty also means ride) or "delivery" — Delivery V1
// Phase 2 (docs/business/DELIVERY_V1_DESIGN.md). The PickupContact*/
// Receiver*/Package* fields are only read when TripType == "delivery".
type CreateTripParams struct {
	RiderID        string
	PickupAddress  string
	DropoffAddress string

	TripType string

	PickupContactName  string
	PickupContactPhone string
	ReceiverName       string
	ReceiverPhone      string
	PackageNote        string
	PackageValue       int64
	PackageWeightKg    float64
}

// CreateTripResult is the output of TripClient.CreateTrip. DeliveryID is
// empty unless TripType == "delivery".
type CreateTripResult struct {
	TripID     string
	DeliveryID string
}

// TripClient abstracts calls to the Trip service.
type TripClient interface {
	CreateTrip(ctx context.Context, in CreateTripParams) (*CreateTripResult, error)
	// MarkDriverArrived transitions a trip from driver_assigned to driver_arrived.
	MarkDriverArrived(ctx context.Context, tripID string) error
	StartTrip(ctx context.Context, tripID string) error
	CompleteTrip(ctx context.Context, tripID string, finalFareTotal int64, fareCurrency string) (*TripInfo, error)
	GetTrip(ctx context.Context, tripID string) (*TripInfo, error)
	// CancelTrip cancels a trip, used for saga compensation when downstream steps fail.
	CancelTrip(ctx context.Context, tripID, reason string) error
	// InitiatePayment transitions a completed trip to payment_pending.
	InitiatePayment(ctx context.Context, tripID string) error
	// PayTrip processes mock payment and transitions payment_pending → settled.
	PayTrip(ctx context.Context, tripID, paymentMethod string) (*TripInfo, error)
	// ListByRider returns all trips for a rider, newest first.
	ListByRider(ctx context.Context, riderID string) ([]TripSummary, error)
	// ListByDriver returns all trips for a driver, newest first.
	ListByDriver(ctx context.Context, driverID string) ([]TripSummary, error)
	// AcceptDelivery transitions the trip's linked Delivery (if any) from
	// Created to Accepted — production hardening P0-1. A no-op, not an
	// error, for a Ride trip or an already-accepted Delivery (idempotent).
	AcceptDelivery(ctx context.Context, tripID string) error
}

// DriverOfferInfo is the active pending offer directed at a driver.
type DriverOfferInfo struct {
	TripID         string
	PickupAddress  string
	DropoffAddress string
	OfferExpiresAt time.Time
}

// DispatchClient abstracts calls to the Dispatch service.
type DispatchClient interface {
	// RequestDispatch initiates a dispatch job. tripType is "ride" (default;
	// empty also means ride) or "delivery" — Delivery V1 Phase 3
	// (docs/business/DELIVERY_V1_DESIGN.md Phần 9), forwarded unchanged from
	// BookRideInput.TripType so Dispatch treats Delivery as a Trip type it
	// is aware of, per "Dispatch phải coi Delivery là một loại Trip".
	// serviceType is one of the 4 Vehicle/Service Catalog tiers
	// (bike/bike_plus/car/car_xl) or empty (no service-type filter).
	RequestDispatch(ctx context.Context, tripID, riderID, tripType, serviceType string, pickupLat, pickupLon float64) error
	AcceptTrip(ctx context.Context, tripID, driverID string) error
	RejectTrip(ctx context.Context, tripID, driverID string) error
	GetDispatchStatus(ctx context.Context, tripID string) (*DispatchInfo, error)
	GetDriverOffer(ctx context.Context, driverID string) (*DriverOfferInfo, error)
}

// PricingClient abstracts calls to the Pricing service.
type PricingClient interface {
	CalculateFinalFare(ctx context.Context, vehicleType string, distanceKM, durationMin float64) (*FareInfo, error)
}
