// Package redis_test contains integration tests for the driver Redis repositories.
// Tests require a running Redis instance.
// When REDIS_ADDR is not set the entire package is skipped (exit 0).
//
// Start Redis with: make infra-up
// Then run: REDIS_ADDR="localhost:6379" \
//           go test github.com/fairride/driver/infrastructure/redis/... -v
package redis_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

var testClient *goredis.Client

var testNow = time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)

func TestMain(m *testing.M) {
	os.Exit(run(m))
}

func run(m *testing.M) int {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		fmt.Fprintln(os.Stderr, "redis tests: REDIS_ADDR not set — skipping")
		return 0
	}

	client := goredis.NewClient(&goredis.Options{Addr: addr})
	defer client.Close()

	if err := client.Ping(context.Background()).Err(); err != nil {
		fmt.Fprintf(os.Stderr, "redis tests: ping %s: %v\n", addr, err)
		return 1
	}
	testClient = client
	return m.Run()
}

// cleanKeys deletes all test keys for the given driver IDs.
func cleanKeys(t *testing.T, driverIDs ...string) {
	t.Helper()
	ctx := context.Background()
	for _, id := range driverIDs {
		testClient.Del(ctx,
			"fairride:drv:online:"+id,
			"fairride:drv:lastseen:"+id,
		)
	}
}
