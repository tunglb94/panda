package app

import (
	"context"
	"time"

	"github.com/fairride/trip/domain/entity"
	"github.com/fairride/trip/domain/repository"
)

// StartTripInput is the input to StartTripUseCase.
type StartTripInput struct {
	TripID string
}

// StartTripUseCase transitions a trip from DriverAssigned/DriverArrived to InProgress.
type StartTripUseCase struct {
	repo repository.TripRepository
}

func NewStartTripUseCase(repo repository.TripRepository) *StartTripUseCase {
	return &StartTripUseCase{repo: repo}
}

// Execute finds the trip, calls Start, and persists the result.
func (uc *StartTripUseCase) Execute(ctx context.Context, input StartTripInput) (*entity.Trip, error) {
	trip, err := uc.repo.FindByID(ctx, input.TripID)
	if err != nil {
		return nil, err
	}
	if err := trip.Start(time.Now().UTC()); err != nil {
		return nil, err
	}
	if err := uc.repo.Save(ctx, trip); err != nil {
		return nil, err
	}
	return trip, nil
}
