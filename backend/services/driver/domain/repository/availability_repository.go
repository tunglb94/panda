package repository

import (
	"context"
	"time"

	"github.com/fairride/driver/domain/entity"
)

// AvailabilityRepository manages driver real-time presence state in Redis.
// All methods return *errors.DomainError on failure.
type AvailabilityRepository interface {
	// SetOnline marks the driver online and starts the heartbeat TTL clock.
	// Idempotent — calling when already online simply resets the TTL.
	SetOnline(ctx context.Context, driverID string, now time.Time) error

	// SetOffline removes the online key and records last_seen.
	// Idempotent — calling when already offline only updates last_seen.
	SetOffline(ctx context.Context, driverID string, now time.Time) error

	// RefreshHeartbeat extends the online key TTL.
	// Returns CodePreconditionFailed if the driver is not currently online
	// (key expired or never set).
	RefreshHeartbeat(ctx context.Context, driverID string, now time.Time) error

	// GetAvailability returns the driver's current online status and last-seen time.
	// Never returns CodeNotFound; an unseen driver has IsOnline=false, LastSeen=zero.
	GetAvailability(ctx context.Context, driverID string) (*entity.AvailabilityState, error)
}
