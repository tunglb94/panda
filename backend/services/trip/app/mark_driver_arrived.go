package app

import (
	"context"
	"time"

	"github.com/fairride/trip/domain/entity"
	"github.com/fairride/trip/domain/repository"
)

// MarkDriverArrivedUseCase transitions a trip from DriverAssigned to DriverArrived.
type MarkDriverArrivedUseCase struct {
	repo repository.TripRepository
}

func NewMarkDriverArrivedUseCase(repo repository.TripRepository) *MarkDriverArrivedUseCase {
	return &MarkDriverArrivedUseCase{repo: repo}
}

func (uc *MarkDriverArrivedUseCase) Execute(ctx context.Context, tripID string) (*entity.Trip, error) {
	trip, err := uc.repo.FindByID(ctx, tripID)
	if err != nil {
		return nil, err
	}
	if err := trip.MarkDriverArrived(time.Now().UTC()); err != nil {
		return nil, err
	}
	if err := uc.repo.Save(ctx, trip); err != nil {
		return nil, err
	}
	return trip, nil
}
