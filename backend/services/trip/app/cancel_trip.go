package app

import (
	"context"
	"time"

	"github.com/fairride/trip/domain/entity"
	"github.com/fairride/trip/domain/repository"
)

// CancelTripInput carries the trip ID and optional cancellation reason.
type CancelTripInput struct {
	TripID string
	Reason string
}

// CancelTripUseCase cancels a trip that has not yet started or completed.
type CancelTripUseCase struct {
	repo repository.TripRepository
}

func NewCancelTripUseCase(repo repository.TripRepository) *CancelTripUseCase {
	return &CancelTripUseCase{repo: repo}
}

func (uc *CancelTripUseCase) Execute(ctx context.Context, in CancelTripInput) (*entity.Trip, error) {
	trip, err := uc.repo.FindByID(ctx, in.TripID)
	if err != nil {
		return nil, err
	}
	if err := trip.Cancel(in.Reason, time.Now().UTC()); err != nil {
		return nil, err
	}
	if err := uc.repo.Save(ctx, trip); err != nil {
		return nil, err
	}
	return trip, nil
}
