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
//
// TripType is entity.TripTypeRide (the zero value "" is also treated as
// Ride) or entity.TripTypeDelivery — Delivery V1 Phase 3
// (docs/business/DELIVERY_V1_DESIGN.md Phần 9). ServiceType is optional;
// when set, it is checked against entity.ServiceType.IsSupported (the same
// allow-list for either TripType — see that method's doc comment) before
// any dispatch job is created. When set, offerNextDriver also uses it to
// exclude candidates whose own reported service type doesn't match exactly,
// AND whose driver capability doesn't cover this TripType (Vehicle/Service
// Catalog refactor — see offer_next_driver.go's doc comment). Requests
// that omit ServiceType entirely are unaffected by either check —
// unchanged from every caller written before this catalog existed.
type RequestDispatchInput struct {
	TripID          string
	RiderID         string
	PickupLat       float64
	PickupLon       float64
	OfferTimeoutSec int // 0 = use entity default
	MaxAttempts     int // 0 = use entity default

	TripType    entity.TripType
	ServiceType entity.ServiceType
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

	tripType := entity.TripTypeRide
	if in.TripType == entity.TripTypeDelivery {
		tripType = entity.TripTypeDelivery
	}

	// Reject an unsupported service type before any dispatch job is created
	// or the trip is marked searching — no partial state, no offer attempt
	// ever made. One allow-list for both TripTypes now (ServiceType no
	// longer encodes delivery-vs-ride — see entity.ServiceType's doc
	// comment). Requests that omit service_type entirely are unaffected
	// (see entity.ServiceType.IsSupported's doc comment on why an empty
	// ServiceType is not itself rejected here) — preserves every caller
	// written before this catalog existed.
	if in.ServiceType != "" && !in.ServiceType.IsSupported() {
		return nil, domainerrors.InvalidArgument("service type is not supported: " + string(in.ServiceType))
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
	job.TripType = tripType
	job.ServiceType = in.ServiceType

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
