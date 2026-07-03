package app

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/fairride/dispatch/domain/entity"
	"github.com/fairride/dispatch/domain/repository"
	domainerrors "github.com/fairride/shared/errors"
)

// RequestDispatchInput carries the caller-supplied fields for a new dispatch job.
type RequestDispatchInput struct {
	TripID          string
	RiderID         string
	PickupLat       float64
	PickupLon       float64
	OfferTimeoutSec int // 0 = use entity default
	MaxAttempts     int // 0 = use entity default
}

// RequestDispatchUseCase creates a dispatch job and immediately offers the trip
// to the nearest available driver.
type RequestDispatchUseCase struct {
	jobRepo      repository.DispatchJobRepository
	locationRepo repository.DriverLocationRepository
	tripUpdater  repository.TripUpdater
	radiusKM     float64
	searchLimit  int
}

func NewRequestDispatchUseCase(
	jobRepo repository.DispatchJobRepository,
	locationRepo repository.DriverLocationRepository,
	tripUpdater repository.TripUpdater,
) *RequestDispatchUseCase {
	return &RequestDispatchUseCase{
		jobRepo:      jobRepo,
		locationRepo: locationRepo,
		tripUpdater:  tripUpdater,
		radiusKM:     DefaultSearchRadiusKM,
		searchLimit:  DefaultSearchLimit,
	}
}

func (uc *RequestDispatchUseCase) Execute(ctx context.Context, in RequestDispatchInput) (*entity.DispatchJob, error) {
	jobID, err := generateJobID()
	if err != nil {
		return nil, domainerrors.Internal("failed to generate job id")
	}

	now := time.Now().UTC()
	job, err := entity.NewDispatchJob(
		jobID, in.TripID, in.RiderID,
		in.PickupLat, in.PickupLon,
		in.OfferTimeoutSec, in.MaxAttempts,
		now,
	)
	if err != nil {
		return nil, err
	}

	// Mark the trip as searching.
	if err := uc.tripUpdater.SetSearching(ctx, in.TripID, now); err != nil {
		return nil, err
	}

	// Persist the job before attempting to find a driver so we have a record even
	// if the search fails immediately.
	if err := uc.jobRepo.Save(ctx, job); err != nil {
		return nil, err
	}

	// Find the first available driver and issue the offer.
	if err := offerNextDriver(ctx, job, uc.locationRepo, uc.tripUpdater, uc.jobRepo, uc.radiusKM, uc.searchLimit); err != nil {
		return nil, err
	}

	return job, nil
}

func generateJobID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
