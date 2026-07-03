package entity

import (
	"strings"
	"time"

	"github.com/fairride/shared/errors"
)

// JobStatus represents the lifecycle of a dispatch job.
type JobStatus string

const (
	JobStatusPending   JobStatus = "pending"   // created, no offer sent yet
	JobStatusSearching JobStatus = "searching" // offer sent, awaiting driver response
	JobStatusAssigned  JobStatus = "assigned"  // driver accepted
	JobStatusFailed    JobStatus = "failed"    // all attempts exhausted
	JobStatusCancelled JobStatus = "cancelled" // trip was cancelled before assignment
)

// DefaultOfferTimeoutSec is the per-offer TTL used when the caller supplies 0.
const DefaultOfferTimeoutSec = 30

// DefaultMaxAttempts is the maximum number of drivers to try when the caller supplies 0.
const DefaultMaxAttempts = 5

// NearbyDriver is the value returned by DriverLocationRepository.FindNearby.
type NearbyDriver struct {
	DriverID string
}

// DispatchJob is the aggregate root tracking a single dispatch cycle for a trip.
// It records every driver offered and enforces valid state transitions.
type DispatchJob struct {
	JobID            string
	TripID           string
	RiderID          string
	PickupLat        float64
	PickupLon        float64
	Status           JobStatus
	CurrentDriverID  string    // driver currently holding the offer; empty when none
	AssignedDriverID string    // driver who accepted; empty until assigned
	OfferedDriverIDs []string  // all drivers offered in this cycle (skip on retry)
	OfferExpiresAt   time.Time // zero when no active offer
	OfferTimeoutSec  int
	MaxAttempts      int
	AttemptCount     int
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// NewDispatchJob creates a validated DispatchJob in the Pending status.
func NewDispatchJob(
	jobID, tripID, riderID string,
	pickupLat, pickupLon float64,
	offerTimeoutSec, maxAttempts int,
	now time.Time,
) (*DispatchJob, error) {
	if jobID == "" {
		return nil, errors.InvalidArgument("job id must not be empty")
	}
	if tripID == "" {
		return nil, errors.InvalidArgument("trip id must not be empty")
	}
	if riderID == "" {
		return nil, errors.InvalidArgument("rider id must not be empty")
	}
	if offerTimeoutSec <= 0 {
		offerTimeoutSec = DefaultOfferTimeoutSec
	}
	if maxAttempts <= 0 {
		maxAttempts = DefaultMaxAttempts
	}
	return &DispatchJob{
		JobID:           jobID,
		TripID:          tripID,
		RiderID:         riderID,
		PickupLat:       pickupLat,
		PickupLon:       pickupLon,
		Status:          JobStatusPending,
		OfferTimeoutSec: offerTimeoutSec,
		MaxAttempts:     maxAttempts,
		CreatedAt:       now,
		UpdatedAt:       now,
	}, nil
}

// ReconstituteDispatchJob rebuilds a DispatchJob from persistence. No validation.
func ReconstituteDispatchJob(
	jobID, tripID, riderID string,
	pickupLat, pickupLon float64,
	status JobStatus,
	currentDriverID, assignedDriverID string,
	offeredDriverIDs []string,
	offerExpiresAt time.Time,
	offerTimeoutSec, maxAttempts, attemptCount int,
	createdAt, updatedAt time.Time,
) *DispatchJob {
	if offeredDriverIDs == nil {
		offeredDriverIDs = []string{}
	}
	return &DispatchJob{
		JobID:            jobID,
		TripID:           tripID,
		RiderID:          riderID,
		PickupLat:        pickupLat,
		PickupLon:        pickupLon,
		Status:           status,
		CurrentDriverID:  currentDriverID,
		AssignedDriverID: assignedDriverID,
		OfferedDriverIDs: offeredDriverIDs,
		OfferExpiresAt:   offerExpiresAt,
		OfferTimeoutSec:  offerTimeoutSec,
		MaxAttempts:      maxAttempts,
		AttemptCount:     attemptCount,
		CreatedAt:        createdAt,
		UpdatedAt:        updatedAt,
	}
}

// OfferToDriver sends the trip offer to a specific driver.
// Transitions: Pending → Searching, or Searching (no active offer) → Searching.
// Returns CodePreconditionFailed if the job already has an active unresolved offer,
// or if the job is in a terminal/assigned state.
func (j *DispatchJob) OfferToDriver(driverID string, now time.Time) error {
	if j.Status != JobStatusPending && j.Status != JobStatusSearching {
		return errors.PreconditionFailed("cannot offer from status: " + string(j.Status))
	}
	if j.CurrentDriverID != "" {
		return errors.PreconditionFailed("there is already an active offer for driver: " + j.CurrentDriverID)
	}
	j.Status = JobStatusSearching
	j.CurrentDriverID = driverID
	j.OfferExpiresAt = now.Add(time.Duration(j.OfferTimeoutSec) * time.Second)
	j.OfferedDriverIDs = append(j.OfferedDriverIDs, driverID)
	j.AttemptCount++
	j.UpdatedAt = now
	return nil
}

// Accept records that the current driver accepted the trip.
// Returns CodePreconditionFailed if the job is not in searching status,
// the driverID doesn't match the current offer, or the offer has expired.
func (j *DispatchJob) Accept(driverID string, now time.Time) error {
	if j.Status != JobStatusSearching {
		return errors.PreconditionFailed("job is not in searching status")
	}
	if j.CurrentDriverID != driverID {
		return errors.PreconditionFailed("driver is not the current offer holder")
	}
	if j.IsOfferExpired(now) {
		return errors.PreconditionFailed("offer has expired")
	}
	j.Status = JobStatusAssigned
	j.AssignedDriverID = driverID
	j.CurrentDriverID = ""
	j.OfferExpiresAt = time.Time{}
	j.UpdatedAt = now
	return nil
}

// Reject records that the current driver declined the trip.
// Clears the current offer so a new one can be issued.
// Returns CodePreconditionFailed if the job is not in searching status
// or the driverID doesn't match the current offer holder.
func (j *DispatchJob) Reject(driverID string, now time.Time) error {
	if j.Status != JobStatusSearching {
		return errors.PreconditionFailed("job is not in searching status")
	}
	if j.CurrentDriverID != driverID {
		return errors.PreconditionFailed("driver is not the current offer holder")
	}
	j.clearOffer(now)
	return nil
}

// TimeoutOffer clears an expired offer so a retry can be attempted.
// Returns CodePreconditionFailed if the job is not in searching status
// or the offer has not yet expired.
func (j *DispatchJob) TimeoutOffer(now time.Time) error {
	if j.Status != JobStatusSearching {
		return errors.PreconditionFailed("job is not in searching status")
	}
	if !j.IsOfferExpired(now) {
		return errors.PreconditionFailed("offer has not yet expired")
	}
	j.clearOffer(now)
	return nil
}

// MarkFailed transitions the job to Failed status.
// Called when all candidates are exhausted or max attempts reached.
func (j *DispatchJob) MarkFailed(now time.Time) error {
	if j.Status == JobStatusAssigned || j.Status == JobStatusCancelled {
		return errors.PreconditionFailed("cannot fail a job in status: " + string(j.Status))
	}
	j.Status = JobStatusFailed
	j.CurrentDriverID = ""
	j.OfferExpiresAt = time.Time{}
	j.UpdatedAt = now
	return nil
}

// Cancel transitions the job to Cancelled status.
// Returns CodePreconditionFailed if already assigned or cancelled.
func (j *DispatchJob) Cancel(now time.Time) error {
	if j.Status == JobStatusAssigned || j.Status == JobStatusCancelled || j.Status == JobStatusFailed {
		return errors.PreconditionFailed("cannot cancel a job in status: " + string(j.Status))
	}
	j.Status = JobStatusCancelled
	j.CurrentDriverID = ""
	j.OfferExpiresAt = time.Time{}
	j.UpdatedAt = now
	return nil
}

// HasBeenOffered reports whether the driver has already received an offer in this cycle.
func (j *DispatchJob) HasBeenOffered(driverID string) bool {
	for _, id := range j.OfferedDriverIDs {
		if id == driverID {
			return true
		}
	}
	return false
}

// IsOfferExpired reports whether the current offer's deadline has passed.
// Returns false if there is no active offer.
func (j *DispatchJob) IsOfferExpired(now time.Time) bool {
	if j.OfferExpiresAt.IsZero() || j.CurrentDriverID == "" {
		return false
	}
	return now.After(j.OfferExpiresAt)
}

// OfferedDriverIDsCSV serialises the offered driver list for persistence.
func (j *DispatchJob) OfferedDriverIDsCSV() string {
	return strings.Join(j.OfferedDriverIDs, ",")
}

// ─── private helpers ──────────────────────────────────────────────────────────

func (j *DispatchJob) clearOffer(now time.Time) {
	j.CurrentDriverID = ""
	j.OfferExpiresAt = time.Time{}
	j.UpdatedAt = now
}
