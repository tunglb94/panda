package app

import (
	"context"
	"time"

	"github.com/fairride/notification/domain/repository"
)

// MarkReadUseCase marks every message sent by the *other* participant as
// read, on behalf of a verified participant. A sender's own messages are
// never touched (there is nothing to "read" about a message you sent).
type MarkReadUseCase struct {
	conversations repository.ConversationRepository
	messages      repository.MessageRepository
}

func NewMarkReadUseCase(conversations repository.ConversationRepository, messages repository.MessageRepository) *MarkReadUseCase {
	return &MarkReadUseCase{conversations: conversations, messages: messages}
}

func (uc *MarkReadUseCase) Execute(ctx context.Context, conversationID, requesterID string) error {
	if err := authorizeParticipant(ctx, uc.conversations, conversationID, requesterID); err != nil {
		return err
	}
	return uc.messages.MarkReadByRecipient(ctx, conversationID, requesterID, time.Now().UTC())
}
