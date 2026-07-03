package app

import (
	"context"

	"github.com/fairride/driver/domain/entity"
	"github.com/fairride/driver/domain/repository"
)

// GetDriverProfileUseCase retrieves a driver profile by its driver_id.
type GetDriverProfileUseCase struct {
	repo repository.DriverRepository
}

func NewGetDriverProfileUseCase(repo repository.DriverRepository) *GetDriverProfileUseCase {
	return &GetDriverProfileUseCase{repo: repo}
}

func (uc *GetDriverProfileUseCase) Execute(ctx context.Context, driverID string) (*entity.DriverProfile, error) {
	return uc.repo.FindByID(ctx, driverID)
}

// GetDriverProfileByUserIDUseCase retrieves a driver profile by its linked user_id.
type GetDriverProfileByUserIDUseCase struct {
	repo repository.DriverRepository
}

func NewGetDriverProfileByUserIDUseCase(repo repository.DriverRepository) *GetDriverProfileByUserIDUseCase {
	return &GetDriverProfileByUserIDUseCase{repo: repo}
}

func (uc *GetDriverProfileByUserIDUseCase) Execute(ctx context.Context, userID string) (*entity.DriverProfile, error) {
	return uc.repo.FindByUserID(ctx, userID)
}
