package app

import "context"

// TripSnapshot is the minimal trip state the notification app layer needs
// to authorize participants and detect when a conversation should lazily
// close. Decoupled from any specific transport (gRPC proto vs direct
// struct) so it's trivial to fake in tests.
type TripSnapshot struct {
	TripID   string
	RiderID  string
	DriverID string
	Status   string
	TripType string
}

// TripReader is the read-only trip lookup this package depends on.
// Implemented in Gateway by wrapping the existing trippb.TripServiceClient —
// see gateway/http/handlers/chat_handler.go.
type TripReader interface {
	GetTrip(ctx context.Context, tripID string) (TripSnapshot, error)
}

// closedTripStatuses mirrors trip.entity's terminal statuses (completed,
// cancelled, settled — both Ride and Delivery reuse the same Trip.Status
// field, see docs/business/DELIVERY_V1_DESIGN.md). payment_pending/
// payment_success are deliberately excluded: the ride/delivery itself is
// over but rider and driver may still legitimately need to message about
// payment.
var closedTripStatuses = map[string]bool{
	"completed": true,
	"cancelled": true,
	"settled":   true,
}

// IsTripStatusClosed reports whether a trip status means its conversation
// should be lazily closed on next access (Part 7 — security).
func IsTripStatusClosed(status string) bool {
	return closedTripStatuses[status]
}
