package postgres

import (
	"context"

	"github.com/fairride/identity/domain/entity"
	domainerrors "github.com/fairride/shared/errors"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DeviceRepository is the PostgreSQL implementation of repository.DeviceRepository.
type DeviceRepository struct {
	pool *pgxpool.Pool
}

func NewDeviceRepository(pool *pgxpool.Pool) *DeviceRepository {
	return &DeviceRepository{pool: pool}
}

// Upsert matches on (user_id, device_id). created_at is only set by the
// initial INSERT — ON CONFLICT deliberately does not overwrite it.
func (r *DeviceRepository) Upsert(ctx context.Context, d *entity.UserDevice) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO identity_user_devices
			(user_id, device_id, platform, model, app_version, fcm_token, last_seen, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (user_id, device_id) DO UPDATE SET
			platform    = EXCLUDED.platform,
			model       = EXCLUDED.model,
			app_version = EXCLUDED.app_version,
			fcm_token   = EXCLUDED.fcm_token,
			last_seen   = EXCLUDED.last_seen
	`, d.UserID, d.DeviceID, d.Platform, d.Model, d.AppVersion, d.FCMToken, d.LastSeen, d.CreatedAt)
	if err != nil {
		return domainerrors.Internal("upsert user device").WithMeta("error", err.Error())
	}
	return nil
}
