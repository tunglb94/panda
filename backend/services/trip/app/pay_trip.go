package app

import (
	"context"
	"time"

	domainerrors "github.com/fairride/shared/errors"
	"github.com/fairride/trip/domain/entity"
	"github.com/fairride/trip/domain/repository"
)

// PayTripInput is the input to PayTripUseCase.
type PayTripInput struct {
	TripID        string
	PaymentMethod string // e.g. "cash", "wallet"
}

// PayTripUseCase processes mock payment: payment_pending → payment_success → settled.
type PayTripUseCase struct {
	repo repository.TripRepository
}

func NewPayTripUseCase(repo repository.TripRepository) *PayTripUseCase {
	return &PayTripUseCase{repo: repo}
}

func (uc *PayTripUseCase) Execute(ctx context.Context, input PayTripInput) (*entity.Trip, error) {
	if input.TripID == "" {
		return nil, domainerrors.InvalidArgument("trip_id must not be empty")
	}
	method := input.PaymentMethod
	if method == "" {
		method = "cash"
	}
	trip, err := uc.repo.FindByID(ctx, input.TripID)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	if err := trip.MarkPaid(method, now); err != nil {
		return nil, err
	}
	if err := uc.repo.Save(ctx, trip); err != nil {
		return nil, err
	}
	if err := trip.Settle(now); err != nil {
		return nil, err
	}
	if err := uc.repo.Save(ctx, trip); err != nil {
		return nil, err
	}
	return trip, nil
}
