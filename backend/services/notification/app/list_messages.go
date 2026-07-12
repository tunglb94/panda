package app

import (
	"context"

	"github.com/fairride/notification/domain/entity"
	"github.com/fairride/notification/domain/repository"
	domainerrors "github.com/fairride/shared/errors"
)

const defaultMessageListLimit = 200

// ListMessagesUseCase returns messages newer than a cursor, for a verified participant only.
type ListMessagesUseCase struct {
	conversations repository.ConversationRepository
	messages      repository.MessageRepository
}

func NewListMessagesUseCase(conversations repository.ConversationRepository, messages repository.MessageRepository) *ListMessagesUseCase {
	return &ListMessagesUseCase{conversations: conversations, messages: messages}
}

func (uc *ListMessagesUseCase) Execute(ctx context.Context, conversationID, requesterID string, sinceSeq int64) ([]*entity.Message, error) {
	if err := authorizeParticipant(ctx, uc.conversations, conversationID, requesterID); err != nil {
		return nil, err
	}
	return uc.messages.ListSince(ctx, conversationID, sinceSeq, defaultMessageListLimit)
}

// authorizeParticipant loads the conversation and returns PermissionDenied
// unless requesterID is one of its two participants. Shared by every
// use case in this package that reads/writes an existing conversation.
func authorizeParticipant(ctx context.Context, conversations repository.ConversationRepository, conversationID, requesterID string) error {
	if conversationID == "" || requesterID == "" {
		return domainerrors.InvalidArgument("conversation_id and requester_id are required")
	}
	conv, err := conversations.FindByID(ctx, conversationID)
	if err != nil {
		return err
	}
	if conv.RoleOf(requesterID) == "" {
		return domainerrors.PermissionDenied("requester is not a participant of this conversation")
	}
	return nil
}
