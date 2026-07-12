package entity

import (
	"time"

	"github.com/fairride/shared/errors"
)

// Category classifies a Notification for client-side icon/routing.
type Category string

const (
	CategoryTrip     Category = "trip"
	CategoryDelivery Category = "delivery"
	CategoryChat     Category = "chat"
	CategoryCall     Category = "call"
)

// Notification is one entry in a user's in-app notification feed.
// TripID/ConversationID are optional context for deep-linking; both empty is
// valid for a purely informational notification.
type Notification struct {
	ID             string
	UserID         string
	Category       Category
	Title          string
	Body           string
	TripID         string
	ConversationID string
	CreatedAt      time.Time
	ReadAt         *time.Time
}

func NewNotification(id, userID string, category Category, title, body, tripID, conversationID string, now time.Time) (*Notification, error) {
	if id == "" || userID == "" {
		return nil, errors.InvalidArgument("notification requires id and user_id")
	}
	if title == "" {
		return nil, errors.InvalidArgument("notification title must not be empty")
	}
	return &Notification{
		ID:             id,
		UserID:         userID,
		Category:       category,
		Title:          title,
		Body:           body,
		TripID:         tripID,
		ConversationID: conversationID,
		CreatedAt:      now,
	}, nil
}

// ReconstituteNotification rebuilds a Notification from a persistence record. No validation.
func ReconstituteNotification(id, userID string, category Category, title, body, tripID, conversationID string, createdAt time.Time, readAt *time.Time) *Notification {
	return &Notification{
		ID:             id,
		UserID:         userID,
		Category:       category,
		Title:          title,
		Body:           body,
		TripID:         tripID,
		ConversationID: conversationID,
		CreatedAt:      createdAt,
		ReadAt:         readAt,
	}
}
