package app

import (
	"context"

	"github.com/fairride/driver/domain/entity"
	"github.com/fairride/driver/domain/repository"
	domainerrors "github.com/fairride/shared/errors"
)

// ListAuditLogsUseCase surfaces a driver's full KYC audit trail (submit/
// modify/approve/reject/expire, across both DriverVerification and
// VehicleVerification) for the Admin Review Dashboard's timeline (Phần 9).
// The repository/entity side of this has existed since the KYC Hardening
// phase (Phần 7 — "Không được mất lịch sử") but had no reader until now.
type ListAuditLogsUseCase struct {
	repo repository.AuditLogRepository
}

func NewListAuditLogsUseCase(repo repository.AuditLogRepository) *ListAuditLogsUseCase {
	return &ListAuditLogsUseCase{repo: repo}
}

func (uc *ListAuditLogsUseCase) Execute(ctx context.Context, driverID string, limit int) ([]*entity.AuditLog, error) {
	if driverID == "" {
		return nil, domainerrors.InvalidArgument("driver_id is required")
	}
	return uc.repo.ListByDriverID(ctx, driverID, limit)
}
