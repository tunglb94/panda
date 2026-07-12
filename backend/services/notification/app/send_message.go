package app

import (
	"context"
	"strings"
	"time"

	"github.com/fairride/notification/domain/entity"
	"github.com/fairride/notification/domain/repository"
	domainerrors "github.com/fairride/shared/errors"
	"github.com/google/uuid"
)

// SendMessageUseCase appends a message to a conversation, on behalf of a
// verified participant only, and notifies the other party.
type SendMessageUseCase struct {
	conversations repository.ConversationRepository
	messages      repository.MessageRepository
	notify        *CreateNotificationUseCase
	broadcaster   *Broadcaster
}

func NewSendMessageUseCase(conversations repository.ConversationRepository, messages repository.MessageRepository, notify *CreateNotificationUseCase, broadcaster *Broadcaster) *SendMessageUseCase {
	return &SendMessageUseCase{conversations: conversations, messages: messages, notify: notify, broadcaster: broadcaster}
}

type SendMessageInput struct {
	ConversationID string
	SenderID       string
	Text           string
	QuickReplyKey  string
}

// Execute validates the sender is a participant of an open conversation,
// resolves QuickReplyKey to canonical text server-side when present
// (Part 2 — a client never has to send freeform text for a canned reply),
// saves the message, wakes any long-poll waiters, and best-effort notifies
// the recipient (Part 3).
func (uc *SendMessageUseCase) Execute(ctx context.Context, in SendMessageInput) (*entity.Message, error) {
	conv, err := uc.conversations.FindByID(ctx, in.ConversationID)
	if err != nil {
		return nil, err
	}
	role := conv.RoleOf(in.SenderID)
	if role == "" {
		return nil, domainerrors.PermissionDenied("sender is not a participant of this conversation")
	}
	if !conv.IsOpen() {
		return nil, domainerrors.PreconditionFailed("conversation is closed")
	}

	body := strings.TrimSpace(in.Text)
	quickReplyKey := in.QuickReplyKey
	if quickReplyKey != "" {
		if !entity.IsQuickReplyValidForTripType(quickReplyKey, conv.TripType) {
			return nil, domainerrors.InvalidArgument("unknown or unsupported quick_reply_key: " + quickReplyKey)
		}
		body = entity.QuickReplyText[quickReplyKey]
	}
	if body == "" {
		return nil, domainerrors.InvalidArgument("message text must not be empty")
	}

	senderRole := entity.SenderRider
	if role == "driver" {
		senderRole = entity.SenderDriver
	}
	now := time.Now().UTC()
	msg, err := entity.NewMessage(uuid.NewString(), conv.ID, in.SenderID, senderRole, body, quickReplyKey, now)
	if err != nil {
		return nil, err
	}
	if err := uc.messages.Save(ctx, msg); err != nil {
		return nil, err
	}

	if uc.broadcaster != nil {
		uc.broadcaster.Publish(conv.ID)
	}
	if uc.notify != nil {
		recipient := conv.OtherParty(in.SenderID)
		title := "Tin nhắn mới"
		preview := body
		if len(preview) > 80 {
			preview = preview[:80] + "…"
		}
		_, _ = uc.notify.Execute(ctx, CreateNotificationInput{
			UserID:         recipient,
			Category:       entity.CategoryChat,
			Title:          title,
			Body:           preview,
			TripID:         conv.TripID,
			ConversationID: conv.ID,
		})
	}
	return msg, nil
}
