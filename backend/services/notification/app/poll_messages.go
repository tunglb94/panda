package app

import (
	"context"
	"time"

	"github.com/fairride/notification/domain/entity"
	"github.com/fairride/notification/domain/repository"
)

// DefaultPollTimeout is how long PollMessagesUseCase holds a request open
// waiting for a new message before returning an empty result. Chosen well
// under typical load-balancer/HTTP client timeouts (~30s) while still being
// long enough to avoid tight client-side re-poll loops.
const DefaultPollTimeout = 25 * time.Second

// PollMessagesUseCase implements a genuine long-poll: if no message is
// already available past sinceSeq, it subscribes to the conversation's
// Broadcaster and blocks until either a new message is published, timeout
// elapses, or the request context is cancelled — then re-queries once more
// before returning. This is real long-polling (a held-open request that
// wakes on an event), not short-interval client polling and not WebSocket
// (neither exists in this environment — see module report's Kien truc).
type PollMessagesUseCase struct {
	conversations repository.ConversationRepository
	messages      repository.MessageRepository
	broadcaster   *Broadcaster
}

func NewPollMessagesUseCase(conversations repository.ConversationRepository, messages repository.MessageRepository, broadcaster *Broadcaster) *PollMessagesUseCase {
	return &PollMessagesUseCase{conversations: conversations, messages: messages, broadcaster: broadcaster}
}

func (uc *PollMessagesUseCase) Execute(ctx context.Context, conversationID, requesterID string, sinceSeq int64, timeout time.Duration) ([]*entity.Message, error) {
	if err := authorizeParticipant(ctx, uc.conversations, conversationID, requesterID); err != nil {
		return nil, err
	}
	if timeout <= 0 {
		timeout = DefaultPollTimeout
	}

	msgs, err := uc.messages.ListSince(ctx, conversationID, sinceSeq, defaultMessageListLimit)
	if err != nil {
		return nil, err
	}
	if len(msgs) > 0 || uc.broadcaster == nil {
		return msgs, nil
	}

	wake, cancel := uc.broadcaster.Subscribe(conversationID)
	defer cancel()

	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case <-wake:
	case <-timer.C:
	case <-ctx.Done():
		return msgs, nil
	}
	return uc.messages.ListSince(ctx, conversationID, sinceSeq, defaultMessageListLimit)
}
