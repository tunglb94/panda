package postgres

import (
	"context"

	"github.com/fairride/identity/domain/entity"
	domainerrors "github.com/fairride/shared/errors"
	"github.com/jackc/pgx/v5/pgxpool"
)

// LoginHistoryRepository is the PostgreSQL implementation of repository.LoginHistoryRepository.
type LoginHistoryRepository struct {
	pool *pgxpool.Pool
}

func NewLoginHistoryRepository(pool *pgxpool.Pool) *LoginHistoryRepository {
	return &LoginHistoryRepository{pool: pool}
}

// Append inserts a new login-history row. Append-only — never updates or deletes.
func (r *LoginHistoryRepository) Append(ctx context.Context, rec *entity.LoginRecord) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO identity_login_history
			(id, user_id, login_time, ip, device_id, platform, login_method, success)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, rec.ID, rec.UserID, rec.LoginTime, rec.IP, rec.DeviceID, rec.Platform, string(rec.LoginMethod), rec.Success)
	if err != nil {
		return domainerrors.Internal("append login history").WithMeta("error", err.Error())
	}
	return nil
}
