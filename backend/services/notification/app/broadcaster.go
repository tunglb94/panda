package app

import "sync"

// Broadcaster fans out a lightweight wake-up signal per conversation so
// PollMessagesUseCase's long-poll can return as soon as a new message
// arrives, instead of waiting out the full timeout. In-memory only —
// correct for a single Gateway process. See the module report's Known Gap
// for what a multi-replica deployment would need instead (e.g. Redis pubsub).
type Broadcaster struct {
	mu   sync.Mutex
	subs map[string][]chan struct{}
}

func NewBroadcaster() *Broadcaster {
	return &Broadcaster{subs: make(map[string][]chan struct{})}
}

// Subscribe returns a channel that receives one value the next time Publish
// is called for conversationID. The caller must invoke cancel exactly once
// when done (whether or not the channel fired) to avoid leaking the
// subscription.
func (b *Broadcaster) Subscribe(conversationID string) (ch <-chan struct{}, cancel func()) {
	c := make(chan struct{}, 1)
	b.mu.Lock()
	b.subs[conversationID] = append(b.subs[conversationID], c)
	b.mu.Unlock()

	cancel = func() {
		b.mu.Lock()
		defer b.mu.Unlock()
		list := b.subs[conversationID]
		for i, existing := range list {
			if existing == c {
				b.subs[conversationID] = append(list[:i], list[i+1:]...)
				break
			}
		}
		if len(b.subs[conversationID]) == 0 {
			delete(b.subs, conversationID)
		}
	}
	return c, cancel
}

// Publish wakes every current subscriber of conversationID. Non-blocking: a
// subscriber channel is buffered size 1, so a wake-up is never lost even if
// the subscriber hasn't started selecting on it yet.
func (b *Broadcaster) Publish(conversationID string) {
	b.mu.Lock()
	subs := append([]chan struct{}(nil), b.subs[conversationID]...)
	b.mu.Unlock()
	for _, c := range subs {
		select {
		case c <- struct{}{}:
		default:
		}
	}
}
