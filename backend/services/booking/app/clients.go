// Package app contains the booking orchestration use cases.
// It depends only on the interfaces defined here — never on concrete gRPC clients —
// so every use case is testable without a running service.
package app

import "context"

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

// TripClient abstracts calls to the Trip service.
type TripClient interface {
	CreateTrip(ctx context.Context, riderID, pickupAddress, dropoffAddress string) (tripID string, err error)
	StartTrip(ctx context.Context, tripID string) error
	CompleteTrip(ctx context.Context, tripID string, finalFareTotal int64, fareCurrency string) (*TripInfo, error)
	GetTrip(ctx context.Context, tripID string) (*TripInfo, error)
	// CancelTrip cancels a trip, used for saga compensation when downstream steps fail.
	CancelTrip(ctx context.Context, tripID, reason string) error
}

// DispatchClient abstracts calls to the Dispatch service.
type DispatchClient interface {
	RequestDispatch(ctx context.Context, tripID, riderID string, pickupLat, pickupLon float64) error
	AcceptTrip(ctx context.Context, tripID, driverID string) error
	RejectTrip(ctx context.Context, tripID, driverID string) error
	GetDispatchStatus(ctx context.Context, tripID string) (*DispatchInfo, error)
}

// PricingClient abstracts calls to the Pricing service.
type PricingClient interface {
	CalculateFinalFare(ctx context.Context, vehicleType string, distanceKM, durationMin float64) (*FareInfo, error)
}
