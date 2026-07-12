// Package entity defines the domain model for the Communication Module:
// Conversation/Message (in-app chat), Notification (in-app notification
// feed), and CallSession (phone-call audit log). Shared by Ride and Delivery
// alike — every aggregate here keys off TripID, and TripType is carried only
// for display, never for branching behavior (Part 8 of the module spec).
package entity

import "time"

// ConversationStatus tracks whether a conversation still accepts new
// messages. A conversation is lazily closed the next time it's touched
// after its trip reaches a terminal status — see IsTripStatusClosed in the
// app package.
type ConversationStatus string

const (
	ConversationOpen   ConversationStatus = "open"
	ConversationClosed ConversationStatus = "closed"
)

// Conversation is the aggregate root scoping chat to exactly the rider and
// driver of one trip. One conversation per trip (enforced by the trip_id
// UNIQUE constraint in migration 009).
type Conversation struct {
	ID        string
	TripID    string
	RiderID   string
	DriverID  string
	TripType  string
	Status    ConversationStatus
	CreatedAt time.Time
	ClosedAt  *time.Time
}

// NewConversation creates a validated, open Conversation for a trip that
// already has both a rider and a driver assigned.
func NewConversation(id, tripID, riderID, driverID, tripType string, now time.Time) *Conversation {
	return &Conversation{
		ID:        id,
		TripID:    tripID,
		RiderID:   riderID,
		DriverID:  driverID,
		TripType:  tripType,
		Status:    ConversationOpen,
		CreatedAt: now,
	}
}

// ReconstituteConversation rebuilds a Conversation from a persistence record. No validation.
func ReconstituteConversation(id, tripID, riderID, driverID, tripType string, status ConversationStatus, createdAt time.Time, closedAt *time.Time) *Conversation {
	return &Conversation{
		ID:        id,
		TripID:    tripID,
		RiderID:   riderID,
		DriverID:  driverID,
		TripType:  tripType,
		Status:    status,
		CreatedAt: createdAt,
		ClosedAt:  closedAt,
	}
}

// RoleOf returns "rider" or "driver" for a participant, or "" if userID is
// not a participant of this conversation (Part 7 — security boundary).
func (c *Conversation) RoleOf(userID string) string {
	switch userID {
	case c.RiderID:
		return "rider"
	case c.DriverID:
		return "driver"
	default:
		return ""
	}
}

// OtherParty returns the participant ID opposite userID. Only meaningful when
// userID is already known to be a participant (check RoleOf first).
func (c *Conversation) OtherParty(userID string) string {
	if userID == c.RiderID {
		return c.DriverID
	}
	return c.RiderID
}

func (c *Conversation) IsOpen() bool { return c.Status == ConversationOpen }

// Close transitions the conversation to Closed. Idempotent.
func (c *Conversation) Close(now time.Time) {
	if c.Status == ConversationClosed {
		return
	}
	c.Status = ConversationClosed
	c.ClosedAt = &now
}
