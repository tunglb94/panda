// Package redis implements the driver availability repository backed by Redis.
package redis

import (
	"context"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"github.com/fairride/driver/domain/entity"
	domainerrors "github.com/fairride/shared/errors"
)

const (
	// DefaultHeartbeatTimeout is the TTL of the online key.
	// If a driver sends no heartbeat within this window they are considered offline.
	DefaultHeartbeatTimeout = 5 * time.Minute

	// DefaultLastSeenTTL is the retention period for the last-seen timestamp.
	DefaultLastSeenTTL = 30 * 24 * time.Hour

	// lastSeenLayout is the timestamp format stored in Redis.
	lastSeenLayout = time.RFC3339Nano
)

// AvailabilityRepository persists driver real-time presence state in Redis.
type AvailabilityRepository struct {
	client           *goredis.Client
	heartbeatTimeout time.Duration
	lastSeenTTL      time.Duration
}

// NewAvailabilityRepository creates an AvailabilityRepository with default TTLs.
func NewAvailabilityRepository(client *goredis.Client) *AvailabilityRepository {
	return &AvailabilityRepository{
		client:           client,
		heartbeatTimeout: DefaultHeartbeatTimeout,
		lastSeenTTL:      DefaultLastSeenTTL,
	}
}

// NewAvailabilityRepositoryWithTTL creates an AvailabilityRepository with custom TTLs.
// Useful for tests where a very short TTL is needed to verify expiry behaviour.
func NewAvailabilityRepositoryWithTTL(client *goredis.Client, heartbeatTimeout, lastSeenTTL time.Duration) *AvailabilityRepository {
	return &AvailabilityRepository{
		client:           client,
		heartbeatTimeout: heartbeatTimeout,
		lastSeenTTL:      lastSeenTTL,
	}
}

// SetOnline marks the driver online and (re)starts the heartbeat TTL.
func (r *AvailabilityRepository) SetOnline(ctx context.Context, driverID string, now time.Time) error {
	pipe := r.client.Pipeline()
	pipe.Set(ctx, onlineKey(driverID), "1", r.heartbeatTimeout)
	pipe.Set(ctx, lastSeenKey(driverID), now.UTC().Format(lastSeenLayout), r.lastSeenTTL)
	if _, err := pipe.Exec(ctx); err != nil {
		return domainerrors.Internal("set driver online").WithMeta("error", err.Error())
	}
	return nil
}

// SetOffline removes the online key and records last_seen.
func (r *AvailabilityRepository) SetOffline(ctx context.Context, driverID string, now time.Time) error {
	pipe := r.client.Pipeline()
	pipe.Del(ctx, onlineKey(driverID))
	pipe.Set(ctx, lastSeenKey(driverID), now.UTC().Format(lastSeenLayout), r.lastSeenTTL)
	if _, err := pipe.Exec(ctx); err != nil {
		return domainerrors.Internal("set driver offline").WithMeta("error", err.Error())
	}
	return nil
}

// RefreshHeartbeat extends the online key TTL.
// Returns CodePreconditionFailed if the driver is not currently online.
func (r *AvailabilityRepository) RefreshHeartbeat(ctx context.Context, driverID string, now time.Time) error {
	// EXPIRE returns false when the key does not exist.
	refreshed, err := r.client.Expire(ctx, onlineKey(driverID), r.heartbeatTimeout).Result()
	if err != nil {
		return domainerrors.Internal("heartbeat expire").WithMeta("error", err.Error())
	}
	if !refreshed {
		return domainerrors.PreconditionFailed("driver is not online — call GoOnline first")
	}
	// Update last_seen independently; a non-fatal failure here is tolerable
	// but we propagate it so the caller can log it.
	if err := r.client.Set(ctx, lastSeenKey(driverID), now.UTC().Format(lastSeenLayout), r.lastSeenTTL).Err(); err != nil {
		return domainerrors.Internal("heartbeat last_seen update").WithMeta("error", err.Error())
	}
	return nil
}

// GetAvailability returns the driver's current online status and last-seen time.
// Never returns CodeNotFound; an unseen driver has IsOnline=false, LastSeen=zero.
func (r *AvailabilityRepository) GetAvailability(ctx context.Context, driverID string) (*entity.AvailabilityState, error) {
	pipe := r.client.Pipeline()
	existsCmd := pipe.Exists(ctx, onlineKey(driverID))
	lastSeenCmd := pipe.Get(ctx, lastSeenKey(driverID))
	if _, err := pipe.Exec(ctx); err != nil && err != goredis.Nil {
		return nil, domainerrors.Internal("get availability").WithMeta("error", err.Error())
	}

	isOnline := existsCmd.Val() > 0

	var lastSeen time.Time
	if raw, err := lastSeenCmd.Result(); err == nil {
		if t, err := time.Parse(lastSeenLayout, raw); err == nil {
			lastSeen = t.UTC()
		}
	}
	// redis.Nil on lastSeenCmd just means never seen — leave lastSeen as zero.

	return &entity.AvailabilityState{
		DriverID: driverID,
		IsOnline: isOnline,
		LastSeen: lastSeen,
	}, nil
}

// ─── private helpers ──────────────────────────────────────────────────────────

func onlineKey(driverID string) string   { return "fairride:drv:online:" + driverID }
func lastSeenKey(driverID string) string { return "fairride:drv:lastseen:" + driverID }
