// Package redis implements the DriverLocationRepository backed by Redis GEO commands.
package redis

import (
	"context"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"github.com/fairride/dispatch/domain/entity"
	"github.com/fairride/dispatch/domain/repository"
	domainerrors "github.com/fairride/shared/errors"
)

const (
	// DefaultLocationTTL is how long a driver remains "active" after a location update.
	// Drivers that miss two updates at a 15-second interval are considered gone.
	DefaultLocationTTL = 35 * time.Second

	geoKey = "fairride:dispatch:drv:loc"
)

// DriverLocationRepository stores driver coordinates in a Redis GEO sorted set
// and uses a companion TTL key to track whether a driver is still actively sending
// location updates.
type DriverLocationRepository struct {
	client      *goredis.Client
	locationTTL time.Duration
}

var _ repository.DriverLocationRepository = (*DriverLocationRepository)(nil)

// NewDriverLocationRepository creates a repository with the default location TTL.
func NewDriverLocationRepository(client *goredis.Client) *DriverLocationRepository {
	return &DriverLocationRepository{client: client, locationTTL: DefaultLocationTTL}
}

// NewDriverLocationRepositoryWithTTL creates a repository with a custom TTL.
// Useful for tests that need to verify expiry behaviour quickly.
func NewDriverLocationRepositoryWithTTL(client *goredis.Client, ttl time.Duration) *DriverLocationRepository {
	return &DriverLocationRepository{client: client, locationTTL: ttl}
}

// UpdateLocation stores or refreshes the driver's coordinates and resets the active TTL.
func (r *DriverLocationRepository) UpdateLocation(ctx context.Context, driverID string, lat, lon float64) error {
	pipe := r.client.Pipeline()
	pipe.GeoAdd(ctx, geoKey, &goredis.GeoLocation{
		Name:      driverID,
		Latitude:  lat,
		Longitude: lon,
	})
	pipe.Set(ctx, activeKey(driverID), "1", r.locationTTL)
	if _, err := pipe.Exec(ctx); err != nil {
		return domainerrors.Internal("dispatch: update driver location").WithMeta("error", err.Error())
	}
	return nil
}

// FindNearby returns drivers within radiusKM of (lat, lon), sorted nearest first.
// At most limit results are returned.
func (r *DriverLocationRepository) FindNearby(ctx context.Context, lat, lon, radiusKM float64, limit int) ([]*entity.NearbyDriver, error) {
	members, err := r.client.GeoSearch(ctx, geoKey, &goredis.GeoSearchQuery{
		Latitude:   lat,
		Longitude:  lon,
		Radius:     radiusKM,
		RadiusUnit: "km",
		Sort:       "ASC",
		Count:      limit,
	}).Result()
	if err != nil && err != goredis.Nil {
		return nil, domainerrors.Internal("dispatch: geo search failed").WithMeta("error", err.Error())
	}
	result := make([]*entity.NearbyDriver, 0, len(members))
	for _, m := range members {
		result = append(result, &entity.NearbyDriver{DriverID: m})
	}
	return result, nil
}

// IsActive returns true if the driver has sent a location update within the TTL window.
func (r *DriverLocationRepository) IsActive(ctx context.Context, driverID string) (bool, error) {
	n, err := r.client.Exists(ctx, activeKey(driverID)).Result()
	if err != nil {
		return false, domainerrors.Internal("dispatch: check driver active").WithMeta("error", err.Error())
	}
	return n > 0, nil
}

// RemoveLocation removes the driver from the geo set and deletes the active key.
func (r *DriverLocationRepository) RemoveLocation(ctx context.Context, driverID string) error {
	pipe := r.client.Pipeline()
	pipe.ZRem(ctx, geoKey, driverID) // GEO is backed by sorted set; ZRem removes the member
	pipe.Del(ctx, activeKey(driverID))
	if _, err := pipe.Exec(ctx); err != nil {
		return domainerrors.Internal("dispatch: remove driver location").WithMeta("error", err.Error())
	}
	return nil
}

// ─── private helpers ──────────────────────────────────────────────────────────

func activeKey(driverID string) string { return "fairride:dispatch:drv:active:" + driverID }
