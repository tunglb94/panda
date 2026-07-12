package aiengine

import (
	"crypto/sha256"
	"encoding/hex"
	"sync"
)

// ResponseCache is a thread-safe, content-addressed cache: identical
// prompts always return the same cached decision, satisfying the sprint
// brief's "Mọi prompt giống nhau phải cache. Không gọi AI lặp." Keyed by
// SHA-256 of the exact prompt text sent to Ollama, so a change to the
// prompt template naturally invalidates old cache entries without any
// explicit versioning logic.
type ResponseCache struct {
	mu      sync.RWMutex
	entries map[string]string // hash(prompt) -> raw response
	hits    int64
	misses  int64
}

func NewResponseCache() *ResponseCache {
	return &ResponseCache{entries: make(map[string]string)}
}

func hashPrompt(prompt string) string {
	sum := sha256.Sum256([]byte(prompt))
	return hex.EncodeToString(sum[:])
}

// Get returns the cached response for prompt, if any.
func (c *ResponseCache) Get(prompt string) (string, bool) {
	key := hashPrompt(prompt)
	c.mu.RLock()
	defer c.mu.RUnlock()
	v, ok := c.entries[key]
	return v, ok
}

// Put stores response under prompt's hash.
func (c *ResponseCache) Put(prompt, response string) {
	key := hashPrompt(prompt)
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries[key] = response
}

// RecordHit/RecordMiss track cache effectiveness for the benchmark report.
func (c *ResponseCache) RecordHit()  { c.mu.Lock(); c.hits++; c.mu.Unlock() }
func (c *ResponseCache) RecordMiss() { c.mu.Lock(); c.misses++; c.mu.Unlock() }

// Stats returns (hits, misses, hitRatePercent).
func (c *ResponseCache) Stats() (int64, int64, float64) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	total := c.hits + c.misses
	if total == 0 {
		return 0, 0, 0
	}
	return c.hits, c.misses, 100 * float64(c.hits) / float64(total)
}

// Size returns the number of distinct cached prompts.
func (c *ResponseCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.entries)
}
