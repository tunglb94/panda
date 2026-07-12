package repository

import (
	"context"
	"time"

	"github.com/fairride/notification/domain/entity"
)

// ConversationRepository persists Conversation aggregates.
type ConversationRepository interface {
	Save(ctx context.Context, c *entity.Conversation) error
	FindByTripID(ctx context.Context, tripID string) (*entity.Conversation, error)
	FindByID(ctx context.Context, id string) (*entity.Conversation, error)
	Update(ctx context.Context, c *entity.Conversation) error
}

// MessageRepository persists Message entities scoped to a Conversation.
type MessageRepository interface {
	Save(ctx context.Context, m *entity.Message) error
	// ListSince returns messages with Seq > sinceSeq for a conversation,
	// oldest first, capped at limit.
	ListSince(ctx context.Context, conversationID string, sinceSeq int64, limit int) ([]*entity.Message, error)
	// MarkReadByRecipient stamps ReadAt=now on every message in the
	// conversation that recipientID did not send and that isn't already read.
	MarkReadByRecipient(ctx context.Context, conversationID, recipientID string, now time.Time) error
	// CountUnread counts messages in the conversation that recipientID did
	// not send and that have no ReadAt yet.
	CountUnread(ctx context.Context, conversationID, recipientID string) (int, error)
}

// NotificationRepository persists Notification entities.
type NotificationRepository interface {
	Save(ctx context.Context, n *entity.Notification) error
	ListByUser(ctx context.Context, userID string, limit int) ([]*entity.Notification, error)
	MarkRead(ctx context.Context, id, userID string, now time.Time) error
	CountUnread(ctx context.Context, userID string) (int, error)
}

// CallSessionRepository persists CallSession audit records.
type CallSessionRepository interface {
	Save(ctx context.Context, cs *entity.CallSession) error
}
