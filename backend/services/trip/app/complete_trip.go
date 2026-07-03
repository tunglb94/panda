package app

import (
	"context"
	"time"

	"github.com/fairride/trip/domain/entity"
	"github.com/fairride/trip/domain/repository"
)

// CompleteTripInput is the input to CompleteTripUseCase.
// FinalFareTotal and FareCurrency are computed by the pricing service before
// this use case is called.
type CompleteTripInput struct {
	TripID         string
	FinalFareTotal int64
	FareCurrency   string
}

// CompleteTripUseCase transitions a trip from InProgress to Completed and records the fare.
type CompleteTripUseCase struct {
	repo repository.TripRepository
}

func NewCompleteTripUseCase(repo repository.TripRepository) *CompleteTripUseCase {
	return &CompleteTripUseCase{repo: repo}
}

// Execute finds the trip, records the fare, transitions to Completed, and persists.
func (uc *CompleteTripUseCase) Execute(ctx context.Context, input CompleteTripInput) (*entity.Trip, error) {
	trip, err := uc.repo.FindByID(ctx, input.TripID)
	if err != nil {
		return nil, err
	}
	if err := trip.Complete(input.FinalFareTotal, input.FareCurrency, time.Now().UTC()); err != nil {
		return nil, err
	}
	if err := uc.repo.Save(ctx, trip); err != nil {
		return nil, err
	}
	return trip, nil
}
