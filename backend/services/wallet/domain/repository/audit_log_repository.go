package repository

import (
	"context"

	"github.com/fairride/wallet/domain/entity"
)

// AuditLogRepository persists AuditLog entries (Phần 12). Append-only — no
// Update/Delete method exists on this interface by design.
type AuditLogRepository interface {
	Save(ctx context.Context, log *entity.AuditLog) error
	ListByDriverID(ctx context.Context, driverID string, limit int) ([]*entity.AuditLog, error)
}
