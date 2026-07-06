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
	transactor   repository.Transactor
	radiusKM     float64
	searchLimit  int
}

func NewRequestDispatchUseCase(
	jobRepo repository.DispatchJobRepository,
	locationRepo repository.DriverLocationRepository,
	transactor repository.Transactor,
) *RequestDispatchUseCase {
	return &RequestDispatchUseCase{
		jobRepo:      jobRepo,
		locationRepo: locationRepo,
		transactor:   transactor,
		radiusKM:     DefaultSearchRadiusKM,
		searchLimit:  DefaultSearchLimit,
	}
}

func (uc *RequestDispatchUseCase) Execute(ctx context.Context, in RequestDispatchInput) (*entity.DispatchJob, error) {
	if in.TripID == "" {
		return nil, domainerrors.InvalidArgument("trip_id is required")
	}

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

	// Atomic: mark trip as searching and persist the initial job record together.
	// If either write fails neither is committed, preventing a trip stuck in
	// 'searching' with no corresponding dispatch job.
	if err := uc.transactor.WithinTx(ctx, func(jobs repository.DispatchJobRepository, trips repository.TripUpdater) error {
		if err := trips.SetSearching(ctx, in.TripID, now); err != nil {
			return err
		}
		return jobs.Save(ctx, job)
	}); err != nil {
		return nil, err
	}

	// Find the first available driver and issue the offer.
	// This runs outside the transaction: offerNextDriver only writes to
	// dispatch_jobs (no trip update), so a failure here leaves the trip in
	// 'searching' with a 'pending' job — a recoverable state the engine retries.
	if err := offerNextDriver(ctx, job, uc.locationRepo, nil, uc.jobRepo, uc.radiusKM, uc.searchLimit); err != nil {
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
