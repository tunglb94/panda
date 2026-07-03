package app

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/fairride/shared/errors"
	"github.com/fairride/trip/domain/entity"
	"github.com/fairride/trip/domain/repository"
)

// CreateTripInput carries the caller-supplied fields for a new trip.
type CreateTripInput struct {
	RiderID        string
	PickupAddress  string
	DropoffAddress string
}

// CreateTripUseCase creates a new trip in the Pending status.
type CreateTripUseCase struct {
	repo repository.TripRepository
}

func NewCreateTripUseCase(repo repository.TripRepository) *CreateTripUseCase {
	return &CreateTripUseCase{repo: repo}
}

func (uc *CreateTripUseCase) Execute(ctx context.Context, in CreateTripInput) (*entity.Trip, error) {
	tripID, err := generateTripID()
	if err != nil {
		return nil, errors.Internal("failed to generate trip id")
	}
	trip, err := entity.NewTrip(tripID, in.RiderID, in.PickupAddress, in.DropoffAddress, time.Now().UTC())
	if err != nil {
		return nil, err
	}
	if err := uc.repo.Save(ctx, trip); err != nil {
		return nil, err
	}
	return trip, nil
}

func generateTripID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
