package app

import (
	"context"
	"time"

	"github.com/fairride/driver/domain/entity"
	"github.com/fairride/driver/domain/repository"
	"github.com/fairride/shared/errors"
)

// ─── UpdateDriverProfile ─────────────────────────────────────────────────────

// UpdateDriverProfileInput carries the fields the caller wants to change.
type UpdateDriverProfileInput struct {
	DriverID      string
	LicenseNumber string
	VehicleType   entity.VehicleType
	VehicleBrand  string
	VehicleModel  string
	VehicleColor  string
	PlateNumber   string
}

// UpdateDriverProfileUseCase replaces a driver's vehicle and license info.
type UpdateDriverProfileUseCase struct {
	repo repository.DriverRepository
}

func NewUpdateDriverProfileUseCase(repo repository.DriverRepository) *UpdateDriverProfileUseCase {
	return &UpdateDriverProfileUseCase{repo: repo}
}

func (uc *UpdateDriverProfileUseCase) Execute(ctx context.Context, in UpdateDriverProfileInput) (*entity.DriverProfile, error) {
	if in.DriverID == "" {
		return nil, errors.InvalidArgument("driver id must not be empty")
	}
	d, err := uc.repo.FindByID(ctx, in.DriverID)
	if err != nil {
		return nil, err
	}
	if err := d.Update(in.LicenseNumber, in.VehicleType, in.VehicleBrand, in.VehicleModel, in.VehicleColor, in.PlateNumber, time.Now()); err != nil {
		return nil, err
	}
	if err := uc.repo.Save(ctx, d); err != nil {
		return nil, err
	}
	return d, nil
}

// ─── UpdateOnlineStatus ───────────────────────────────────────────────────────

// UpdateOnlineStatusUseCase transitions a driver between online and offline.
type UpdateOnlineStatusUseCase struct {
	repo repository.DriverRepository
}

func NewUpdateOnlineStatusUseCase(repo repository.DriverRepository) *UpdateOnlineStatusUseCase {
	return &UpdateOnlineStatusUseCase{repo: repo}
}

func (uc *UpdateOnlineStatusUseCase) Execute(ctx context.Context, driverID string, status entity.OnlineStatus) (*entity.DriverProfile, error) {
	if driverID == "" {
		return nil, errors.InvalidArgument("driver id must not be empty")
	}
	d, err := uc.repo.FindByID(ctx, driverID)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	switch status {
	case entity.OnlineStatusOnline:
		err = d.GoOnline(now)
	case entity.OnlineStatusOffline:
		err = d.GoOffline(now)
	default:
		return nil, errors.InvalidArgument("unknown online status: " + string(status))
	}
	if err != nil {
		return nil, err
	}
	if err := uc.repo.Save(ctx, d); err != nil {
		return nil, err
	}
	return d, nil
}

// ─── UpdateVerificationStatus ─────────────────────────────────────────────────

// VerificationAction names admin actions that drive the verification state machine.
type VerificationAction string

const (
	VerificationActionVerify    VerificationAction = "verified"
	VerificationActionReject    VerificationAction = "rejected"
	VerificationActionSuspend   VerificationAction = "suspended"
	VerificationActionReinstate VerificationAction = "reinstated"
)

// UpdateVerificationStatusUseCase applies admin verification actions to a driver.
type UpdateVerificationStatusUseCase struct {
	repo repository.DriverRepository
}

func NewUpdateVerificationStatusUseCase(repo repository.DriverRepository) *UpdateVerificationStatusUseCase {
	return &UpdateVerificationStatusUseCase{repo: repo}
}

func (uc *UpdateVerificationStatusUseCase) Execute(ctx context.Context, driverID string, action VerificationAction) (*entity.DriverProfile, error) {
	if driverID == "" {
		return nil, errors.InvalidArgument("driver id must not be empty")
	}
	d, err := uc.repo.FindByID(ctx, driverID)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	switch action {
	case VerificationActionVerify:
		err = d.Verify(now)
	case VerificationActionReject:
		err = d.Reject(now)
	case VerificationActionSuspend:
		err = d.Suspend(now)
	case VerificationActionReinstate:
		err = d.Reinstate(now)
	default:
		return nil, errors.InvalidArgument("unknown verification action: " + string(action))
	}
	if err != nil {
		return nil, err
	}
	if err := uc.repo.Save(ctx, d); err != nil {
		return nil, err
	}
	return d, nil
}
