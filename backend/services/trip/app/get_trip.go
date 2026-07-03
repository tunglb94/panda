package app

import (
	"context"

	"github.com/fairride/trip/domain/entity"
	"github.com/fairride/trip/domain/repository"
)

// GetTripUseCase retrieves a single trip by ID.
type GetTripUseCase struct {
	repo repository.TripRepository
}

func NewGetTripUseCase(repo repository.TripRepository) *GetTripUseCase {
	return &GetTripUseCase{repo: repo}
}

func (uc *GetTripUseCase) Execute(ctx context.Context, tripID string) (*entity.Trip, error) {
	return uc.repo.FindByID(ctx, tripID)
}
