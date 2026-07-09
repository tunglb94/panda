package app

import (
	"context"
	"time"

	domainerrors "github.com/fairride/shared/errors"
	"github.com/fairride/trip/domain/entity"
	"github.com/fairride/trip/domain/repository"
)

// InitiatePaymentUseCase transitions a completed trip to payment_pending.
type InitiatePaymentUseCase struct {
	repo repository.TripRepository
}

func NewInitiatePaymentUseCase(repo repository.TripRepository) *InitiatePaymentUseCase {
	return &InitiatePaymentUseCase{repo: repo}
}

func (uc *InitiatePaymentUseCase) Execute(ctx context.Context, tripID string) (*entity.Trip, error) {
	if tripID == "" {
		return nil, domainerrors.InvalidArgument("trip_id must not be empty")
	}
	trip, err := uc.repo.FindByID(ctx, tripID)
	if err != nil {
		return nil, err
	}
	if err := trip.InitiatePayment(time.Now().UTC()); err != nil {
		return nil, err
	}
	if err := uc.repo.Save(ctx, trip); err != nil {
		return nil, err
	}
	return trip, nil
}
