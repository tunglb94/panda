package entity

import (
	"time"

	"github.com/fairride/shared/errors"
)

// SenderRole identifies which side of the conversation sent a message.
type SenderRole string

const (
	SenderRider  SenderRole = "rider"
	SenderDriver SenderRole = "driver"
)

// Message is one chat bubble. Seq is a monotonic, database-assigned ordering
// cursor (see migration 009) — external callers use it as the "since_id"
// polling cursor; ID is the stable external identifier.
type Message struct {
	ID             string
	Seq            int64
	ConversationID string
	SenderID       string
	SenderRole     SenderRole
	Body           string
	QuickReplyKey  string
	CreatedAt      time.Time
	DeliveredAt    *time.Time
	ReadAt         *time.Time
}

// NewMessage creates a validated Message. Delivery is immediate in this
// design (no offline-recipient queueing on the server side — see the
// module's Offline Strategy, which is a client-side send queue instead), so
// DeliveredAt is stamped at creation time.
func NewMessage(id, conversationID, senderID string, senderRole SenderRole, body, quickReplyKey string, now time.Time) (*Message, error) {
	if id == "" || conversationID == "" || senderID == "" {
		return nil, errors.InvalidArgument("message requires id, conversation_id and sender_id")
	}
	if body == "" {
		return nil, errors.InvalidArgument("message body must not be empty")
	}
	if senderRole != SenderRider && senderRole != SenderDriver {
		return nil, errors.InvalidArgument("sender_role must be 'rider' or 'driver'")
	}
	delivered := now
	return &Message{
		ID:             id,
		ConversationID: conversationID,
		SenderID:       senderID,
		SenderRole:     senderRole,
		Body:           body,
		QuickReplyKey:  quickReplyKey,
		CreatedAt:      now,
		DeliveredAt:    &delivered,
	}, nil
}

// ReconstituteMessage rebuilds a Message from a persistence record. No validation.
func ReconstituteMessage(id string, seq int64, conversationID, senderID string, senderRole SenderRole, body, quickReplyKey string, createdAt time.Time, deliveredAt, readAt *time.Time) *Message {
	return &Message{
		ID:             id,
		Seq:            seq,
		ConversationID: conversationID,
		SenderID:       senderID,
		SenderRole:     senderRole,
		Body:           body,
		QuickReplyKey:  quickReplyKey,
		CreatedAt:      createdAt,
		DeliveredAt:    deliveredAt,
		ReadAt:         readAt,
	}
}
