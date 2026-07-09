package app

import (
	"context"

	"github.com/fairride/trip/domain/entity"
	"github.com/fairride/trip/domain/repository"
	domainerrors "github.com/fairride/shared/errors"
)

// ListTripsByRiderUseCase returns all trips for a rider, newest first.
type ListTripsByRiderUseCase struct {
	repo repository.TripRepository
}

func NewListTripsByRiderUseCase(repo repository.TripRepository) *ListTripsByRiderUseCase {
	return &ListTripsByRiderUseCase{repo: repo}
}

func (uc *ListTripsByRiderUseCase) Execute(ctx context.Context, riderID string) ([]*entity.Trip, error) {
	if riderID == "" {
		return nil, domainerrors.InvalidArgument("rider_id must not be empty")
	}
	return uc.repo.FindByRiderID(ctx, riderID)
}

// ListTripsByDriverUseCase returns all trips for a driver, newest first.
type ListTripsByDriverUseCase struct {
	repo repository.TripRepository
}

func NewListTripsByDriverUseCase(repo repository.TripRepository) *ListTripsByDriverUseCase {
	return &ListTripsByDriverUseCase{repo: repo}
}

func (uc *ListTripsByDriverUseCase) Execute(ctx context.Context, driverID string) ([]*entity.Trip, error) {
	if driverID == "" {
		return nil, domainerrors.InvalidArgument("driver_id must not be empty")
	}
	return uc.repo.FindByDriverID(ctx, driverID)
}
