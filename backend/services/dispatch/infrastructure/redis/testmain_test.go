// Package redis_test contains integration tests for the dispatch Redis repositories.
// Tests require a running Redis instance pointed to by REDIS_ADDR.
// When REDIS_ADDR is not set the entire package is skipped (exit 0).
//
// Start Redis with: make infra-up
// Then run: REDIS_ADDR=localhost:6379 go test github.com/fairride/dispatch/infrastructure/redis/... -v
package redis_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/redis/go-redis/v9"
)

var testClient *redis.Client

func TestMain(m *testing.M) {
	os.Exit(run(m))
}

func run(m *testing.M) int {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		fmt.Fprintln(os.Stderr, "dispatch redis tests: REDIS_ADDR not set — skipping")
		return 0
	}

	client := redis.NewClient(&redis.Options{Addr: addr})
	if err := client.Ping(context.Background()).Err(); err != nil {
		fmt.Fprintf(os.Stderr, "dispatch redis tests: ping %s: %v\n", addr, err)
		return 1
	}
	testClient = client
	defer client.Close()

	return m.Run()
}

func flushKeys(t *testing.T, pattern string) {
	t.Helper()
	keys, err := testClient.Keys(context.Background(), pattern).Result()
	if err != nil || len(keys) == 0 {
		return
	}
	if err := testClient.Del(context.Background(), keys...).Err(); err != nil {
		t.Fatalf("flushKeys %q: %v", pattern, err)
	}
}
