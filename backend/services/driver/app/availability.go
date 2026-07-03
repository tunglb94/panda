package app

import (
	"context"
	"time"

	"github.com/fairride/driver/domain/entity"
	"github.com/fairride/driver/domain/repository"
	"github.com/fairride/shared/errors"
)

// ─── GoOnlineUseCase ─────────────────────────────────────────────────────────

// GoOnlineUseCase marks a driver as online in Redis.
type GoOnlineUseCase struct {
	repo repository.AvailabilityRepository
}

func NewGoOnlineUseCase(repo repository.AvailabilityRepository) *GoOnlineUseCase {
	return &GoOnlineUseCase{repo: repo}
}

func (uc *GoOnlineUseCase) Execute(ctx context.Context, driverID string) (*entity.AvailabilityState, error) {
	if driverID == "" {
		return nil, errors.InvalidArgument("driver id must not be empty")
	}
	if err := uc.repo.SetOnline(ctx, driverID, time.Now()); err != nil {
		return nil, err
	}
	return uc.repo.GetAvailability(ctx, driverID)
}

// ─── GoOfflineUseCase ────────────────────────────────────────────────────────

// GoOfflineUseCase marks a driver as offline in Redis.
type GoOfflineUseCase struct {
	repo repository.AvailabilityRepository
}

func NewGoOfflineUseCase(repo repository.AvailabilityRepository) *GoOfflineUseCase {
	return &GoOfflineUseCase{repo: repo}
}

func (uc *GoOfflineUseCase) Execute(ctx context.Context, driverID string) (*entity.AvailabilityState, error) {
	if driverID == "" {
		return nil, errors.InvalidArgument("driver id must not be empty")
	}
	if err := uc.repo.SetOffline(ctx, driverID, time.Now()); err != nil {
		return nil, err
	}
	return uc.repo.GetAvailability(ctx, driverID)
}

// ─── HeartbeatUseCase ────────────────────────────────────────────────────────

// HeartbeatUseCase refreshes the online TTL for an active driver.
// Returns CodePreconditionFailed if the driver is not currently online.
type HeartbeatUseCase struct {
	repo repository.AvailabilityRepository
}

func NewHeartbeatUseCase(repo repository.AvailabilityRepository) *HeartbeatUseCase {
	return &HeartbeatUseCase{repo: repo}
}

func (uc *HeartbeatUseCase) Execute(ctx context.Context, driverID string) (*entity.AvailabilityState, error) {
	if driverID == "" {
		return nil, errors.InvalidArgument("driver id must not be empty")
	}
	if err := uc.repo.RefreshHeartbeat(ctx, driverID, time.Now()); err != nil {
		return nil, err
	}
	return uc.repo.GetAvailability(ctx, driverID)
}

// ─── GetAvailabilityUseCase ───────────────────────────────────────────────────

// GetAvailabilityUseCase returns a driver's current online status and last-seen time.
type GetAvailabilityUseCase struct {
	repo repository.AvailabilityRepository
}

func NewGetAvailabilityUseCase(repo repository.AvailabilityRepository) *GetAvailabilityUseCase {
	return &GetAvailabilityUseCase{repo: repo}
}

func (uc *GetAvailabilityUseCase) Execute(ctx context.Context, driverID string) (*entity.AvailabilityState, error) {
	if driverID == "" {
		return nil, errors.InvalidArgument("driver id must not be empty")
	}
	return uc.repo.GetAvailability(ctx, driverID)
}
