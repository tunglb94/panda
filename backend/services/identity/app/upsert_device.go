package app

import (
	"context"
	"time"

	"github.com/fairride/identity/domain/entity"
	"github.com/fairride/identity/domain/repository"
	domainerrors "github.com/fairride/shared/errors"
)

// UpsertDeviceInput carries the device metadata a client sends alongside a
// successful login. Every field but UserID/DeviceID is optional/best-effort.
type UpsertDeviceInput struct {
	UserID     string
	DeviceID   string
	Platform   string
	Model      string
	AppVersion string
	FCMToken   string
}

// UpsertDeviceUseCase records/refreshes the calling device's registration.
type UpsertDeviceUseCase struct {
	repo repository.DeviceRepository
}

func NewUpsertDeviceUseCase(repo repository.DeviceRepository) *UpsertDeviceUseCase {
	return &UpsertDeviceUseCase{repo: repo}
}

// Execute is a no-op (not an error) when DeviceID is empty — device
// reporting is optional; older clients or a login response that never
// collected a device ID must not fail the login over this.
func (uc *UpsertDeviceUseCase) Execute(ctx context.Context, in UpsertDeviceInput) error {
	if in.DeviceID == "" {
		return nil
	}
	if in.UserID == "" {
		return domainerrors.InvalidArgument("user_id must not be empty")
	}
	d, err := entity.NewUserDevice(in.UserID, in.DeviceID, in.Platform, in.Model, in.AppVersion, in.FCMToken, time.Now())
	if err != nil {
		return err
	}
	return uc.repo.Upsert(ctx, d)
}
