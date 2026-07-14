package repository

import (
	"context"

	"github.com/fairride/identity/domain/entity"
)

// DeviceRepository defines persistence operations for UserDevice entities.
type DeviceRepository interface {
	// Upsert inserts or refreshes a (user_id, device_id) row — matched on
	// that composite key. CreatedAt is only written on first insert.
	Upsert(ctx context.Context, device *entity.UserDevice) error
}
