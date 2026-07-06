package app

import (
	"context"
	"sync"
)

// IdempotencyStore records and checks processed request keys.
// A nil store disables idempotency checking — safe for tests and in-process use.
type IdempotencyStore interface {
	Exists(ctx context.Context, key string) (bool, error)
	Record(ctx context.Context, key string) error
}

// MemoryIdempotencyStore is a thread-safe in-memory store for tests and development.
type MemoryIdempotencyStore struct {
	mu   sync.Mutex
	keys map[string]struct{}
}

func NewMemoryIdempotencyStore() *MemoryIdempotencyStore {
	return &MemoryIdempotencyStore{keys: make(map[string]struct{})}
}

func (s *MemoryIdempotencyStore) Exists(_ context.Context, key string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, exists := s.keys[key]
	return exists, nil
}

func (s *MemoryIdempotencyStore) Record(_ context.Context, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.keys[key] = struct{}{}
	return nil
}
