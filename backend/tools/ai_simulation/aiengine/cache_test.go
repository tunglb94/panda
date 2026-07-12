package aiengine

import "testing"

func TestResponseCache_GetMissThenPutThenHit(t *testing.T) {
	c := NewResponseCache()
	if _, ok := c.Get("prompt-a"); ok {
		t.Fatalf("expected a miss on an empty cache")
	}
	c.Put("prompt-a", "CONTINUE")
	got, ok := c.Get("prompt-a")
	if !ok || got != "CONTINUE" {
		t.Errorf("expected cached hit %q, got %q ok=%v", "CONTINUE", got, ok)
	}
}

func TestResponseCache_DistinctPromptsDoNotCollide(t *testing.T) {
	c := NewResponseCache()
	c.Put("prompt-a", "CONTINUE")
	c.Put("prompt-b", "STOP")
	if v, _ := c.Get("prompt-a"); v != "CONTINUE" {
		t.Errorf("prompt-a leaked prompt-b's value: %q", v)
	}
	if v, _ := c.Get("prompt-b"); v != "STOP" {
		t.Errorf("prompt-b leaked prompt-a's value: %q", v)
	}
	if c.Size() != 2 {
		t.Errorf("expected 2 distinct entries, got %d", c.Size())
	}
}

func TestResponseCache_IdenticalPromptOverwrites(t *testing.T) {
	c := NewResponseCache()
	c.Put("prompt-a", "CONTINUE")
	c.Put("prompt-a", "STOP")
	if v, _ := c.Get("prompt-a"); v != "STOP" {
		t.Errorf("expected the later Put to win, got %q", v)
	}
	if c.Size() != 1 {
		t.Errorf("re-Put of the same prompt must not grow the cache, got size %d", c.Size())
	}
}

func TestResponseCache_Stats(t *testing.T) {
	c := NewResponseCache()
	c.RecordMiss()
	c.RecordMiss()
	c.RecordHit()
	hits, misses, hitRate := c.Stats()
	if hits != 1 || misses != 2 {
		t.Errorf("expected hits=1 misses=2, got hits=%d misses=%d", hits, misses)
	}
	wantRate := 100.0 / 3.0
	if hitRate < wantRate-0.01 || hitRate > wantRate+0.01 {
		t.Errorf("expected hit rate ~%.2f, got %.2f", wantRate, hitRate)
	}
}

func TestResponseCache_StatsWithNoActivity(t *testing.T) {
	c := NewResponseCache()
	hits, misses, hitRate := c.Stats()
	if hits != 0 || misses != 0 || hitRate != 0 {
		t.Errorf("expected all-zero stats on an untouched cache, got hits=%d misses=%d rate=%v", hits, misses, hitRate)
	}
}
