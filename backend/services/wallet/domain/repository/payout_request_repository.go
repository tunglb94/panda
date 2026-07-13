package repository

import (
	"context"

	"github.com/fairride/wallet/domain/entity"
)

// PayoutRequestRepository persists PayoutRequest aggregates (Phần 5/8).
type PayoutRequestRepository interface {
	Save(ctx context.Context, p *entity.PayoutRequest) error
	FindByID(ctx context.Context, id string) (*entity.PayoutRequest, error)
	// FindInFlightByDriverID returns the driver's current Pending or
	// Approved request, or CodeNotFound if none — Phần 5's "Không có
	// payout đang Pending" check.
	FindInFlightByDriverID(ctx context.Context, driverID string) (*entity.PayoutRequest, error)
	// ListByDriverID returns a driver's own payout requests, newest first.
	ListByDriverID(ctx context.Context, driverID string, limit int) ([]*entity.PayoutRequest, error)
	// ListByFilter powers the admin list/search (Phần 10) — status/driverID
	// are optional filters (empty means "any").
	ListByFilter(ctx context.Context, status entity.PayoutStatus, driverID string, limit int) ([]*entity.PayoutRequest, error)
}
