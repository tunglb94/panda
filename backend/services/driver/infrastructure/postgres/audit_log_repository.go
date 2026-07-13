package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/fairride/driver/domain/entity"
	"github.com/fairride/driver/domain/repository"
	domainerrors "github.com/fairride/shared/errors"
)

const auditLogFields = `id, entity_type, entity_id, driver_id, action, actor_id, old_value, new_value, reason, created_at`

// AuditLogRepository is the PostgreSQL implementation of
// repository.AuditLogRepository (Phần 7) — append-only, no Update/Delete
// method exists here.
type AuditLogRepository struct {
	pool *pgxpool.Pool
}

var _ repository.AuditLogRepository = (*AuditLogRepository)(nil)

func NewAuditLogRepository(pool *pgxpool.Pool) *AuditLogRepository {
	return &AuditLogRepository{pool: pool}
}

func (r *AuditLogRepository) Save(ctx context.Context, log *entity.AuditLog) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO kyc_audit_logs (`+auditLogFields+`)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		log.ID, string(log.EntityType), log.EntityID, log.DriverID, string(log.Action), log.ActorID,
		log.OldValue, log.NewValue, log.Reason, log.CreatedAt.UTC(),
	)
	if err != nil {
		return domainerrors.Internal("audit_log: save failed").WithMeta("error", err.Error())
	}
	return nil
}

func (r *AuditLogRepository) ListByDriverID(ctx context.Context, driverID string, limit int) ([]*entity.AuditLog, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	rows, err := r.pool.Query(ctx, `
		SELECT `+auditLogFields+` FROM kyc_audit_logs
		WHERE driver_id = $1
		ORDER BY created_at DESC
		LIMIT $2`, driverID, limit)
	if err != nil {
		return nil, domainerrors.Internal("audit_log: list failed").WithMeta("error", err.Error())
	}
	defer rows.Close()

	var out []*entity.AuditLog
	for rows.Next() {
		var (
			id, entityType, entityID, dID, action, actorID string
			oldValue, newValue, reason                     string
			createdAt                                      time.Time
		)
		if err := rows.Scan(&id, &entityType, &entityID, &dID, &action, &actorID, &oldValue, &newValue, &reason, &createdAt); err != nil {
			return nil, domainerrors.Internal("audit_log: scan failed").WithMeta("error", err.Error())
		}
		out = append(out, &entity.AuditLog{
			ID: id, EntityType: entity.AuditEntityType(entityType), EntityID: entityID, DriverID: dID,
			Action: entity.AuditAction(action), ActorID: actorID,
			OldValue: oldValue, NewValue: newValue, Reason: reason, CreatedAt: createdAt.UTC(),
		})
	}
	if err := rows.Err(); err != nil {
		return nil, domainerrors.Internal("audit_log: rows error").WithMeta("error", err.Error())
	}
	return out, nil
}
