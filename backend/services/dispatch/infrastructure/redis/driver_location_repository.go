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

// UpdateLocation stores or refreshes the driver's coordinates and resets the
// active TTL. serviceType is optional (empty string = not recorded).
// rideEnabled/deliveryEnabled are stored as given — the caller (see
// gateway/http/handlers/location_handler.go and
// dispatch/grpc/handler.go's UpdateDriverLocation) is responsible for
// defaulting an unreported capability to rideEnabled=true/
// deliveryEnabled=false before calling this, matching migration 008's DB
// column defaults.
func (r *DriverLocationRepository) UpdateLocation(ctx context.Context, driverID string, lat, lon float64, serviceType entity.ServiceType, rideEnabled, deliveryEnabled bool) error {
	pipe := r.client.Pipeline()
	pipe.GeoAdd(ctx, geoKey, &goredis.GeoLocation{
		Name:      driverID,
		Latitude:  lat,
		Longitude: lon,
	})
	pipe.Set(ctx, activeKey(driverID), "1", r.locationTTL)
	if serviceType != "" {
		pipe.Set(ctx, serviceTypeKey(driverID), string(serviceType), r.locationTTL)
	}
	pipe.Set(ctx, rideEnabledKey(driverID), boolFlag(rideEnabled), r.locationTTL)
	pipe.Set(ctx, deliveryEnabledKey(driverID), boolFlag(deliveryEnabled), r.locationTTL)
	if _, err := pipe.Exec(ctx); err != nil {
		return domainerrors.Internal("dispatch: update driver location").WithMeta("error", err.Error())
	}
	return nil
}

// FindNearby returns drivers within radiusKM of (lat, lon), sorted nearest
// first. At most limit results are returned. Each result's ServiceType is
// populated from the last UpdateLocation call that included one — empty if
// the driver has never reported one (older client, or the key expired).
// RideEnabled/DeliveryEnabled default to true/false (migration 008's DB
// column defaults) when the driver has never reported them or the key
// expired — backward compatible with clients that predate this field.
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
	if len(members) == 0 {
		return result, nil
	}

	// One pipelined round-trip for every candidate's service type +
	// capability flags, instead of N individual GETs.
	pipe := r.client.Pipeline()
	serviceCmds := make([]*goredis.StringCmd, len(members))
	rideCmds := make([]*goredis.StringCmd, len(members))
	deliveryCmds := make([]*goredis.StringCmd, len(members))
	for i, m := range members {
		serviceCmds[i] = pipe.Get(ctx, serviceTypeKey(m))
		rideCmds[i] = pipe.Get(ctx, rideEnabledKey(m))
		deliveryCmds[i] = pipe.Get(ctx, deliveryEnabledKey(m))
	}
	if _, err := pipe.Exec(ctx); err != nil && err != goredis.Nil {
		return nil, domainerrors.Internal("dispatch: fetch driver capability failed").WithMeta("error", err.Error())
	}

	for i, m := range members {
		st, _ := serviceCmds[i].Result() // Nil (driver never reported one) -> "", not an error.
		result = append(result, &entity.NearbyDriver{
			DriverID:        m,
			ServiceType:     entity.ServiceType(st),
			RideEnabled:     parseBoolFlag(rideCmds[i], true),
			DeliveryEnabled: parseBoolFlag(deliveryCmds[i], false),
		})
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

// GetLocation returns the driver's last known coordinates from the Redis GEO set.
// Returns CodeNotFound if the driver has never reported a location.
func (r *DriverLocationRepository) GetLocation(ctx context.Context, driverID string) (lat, lon float64, err error) {
	positions, geoErr := r.client.GeoPos(ctx, geoKey, driverID).Result()
	if geoErr != nil {
		return 0, 0, domainerrors.Internal("dispatch: get driver location").WithMeta("error", geoErr.Error())
	}
	if len(positions) == 0 || positions[0] == nil {
		return 0, 0, domainerrors.NotFound("driver location not found: " + driverID)
	}
	return positions[0].Latitude, positions[0].Longitude, nil
}

// RemoveLocation removes the driver from the geo set and deletes the active,
// service-type, and capability keys.
func (r *DriverLocationRepository) RemoveLocation(ctx context.Context, driverID string) error {
	pipe := r.client.Pipeline()
	pipe.ZRem(ctx, geoKey, driverID) // GEO is backed by sorted set; ZRem removes the member
	pipe.Del(ctx, activeKey(driverID))
	pipe.Del(ctx, serviceTypeKey(driverID))
	pipe.Del(ctx, rideEnabledKey(driverID))
	pipe.Del(ctx, deliveryEnabledKey(driverID))
	if _, err := pipe.Exec(ctx); err != nil {
		return domainerrors.Internal("dispatch: remove driver location").WithMeta("error", err.Error())
	}
	return nil
}

// ─── private helpers ──────────────────────────────────────────────────────────

func activeKey(driverID string) string { return "fairride:dispatch:drv:active:" + driverID }

func serviceTypeKey(driverID string) string { return "fairride:dispatch:drv:st:" + driverID }

func rideEnabledKey(driverID string) string { return "fairride:dispatch:drv:ride:" + driverID }

func deliveryEnabledKey(driverID string) string { return "fairride:dispatch:drv:delivery:" + driverID }

func boolFlag(b bool) string {
	if b {
		return "1"
	}
	return "0"
}

// parseBoolFlag reads a Redis GET result written by boolFlag, falling back
// to defaultVal when the key was never set (Nil) or expired.
func parseBoolFlag(cmd *goredis.StringCmd, defaultVal bool) bool {
	v, err := cmd.Result()
	if err != nil || v == "" {
		return defaultVal
	}
	return v == "1"
}
