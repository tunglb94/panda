package integration

import (
	"context"
	"sync"

	tripentity "github.com/fairride/trip/domain/entity"
	triprepo "github.com/fairride/trip/domain/repository"
	domainerrors "github.com/fairride/shared/errors"
)

// fakeTripRepository implements repository.TripRepository entirely in
// memory — the simulation's equivalent of the Trip service's own
// (not-yet-written) Postgres-backed implementation. Only FindByID/Save are
// ever exercised by the delivery use cases this tool calls
// (CreateTrip/AcceptDelivery/PickupParcel/StartDelivery/CompleteDelivery);
// FindByRiderID/FindByDriverID are implemented for interface completeness
// but unused.
type fakeTripRepository struct {
	mu    sync.Mutex
	byID  map[string]*tripentity.Trip
}

var _ triprepo.TripRepository = (*fakeTripRepository)(nil)

func newFakeTripRepository() *fakeTripRepository {
	return &fakeTripRepository{byID: make(map[string]*tripentity.Trip)}
}

func (r *fakeTripRepository) Save(_ context.Context, trip *tripentity.Trip) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	clone := *trip
	r.byID[trip.TripID] = &clone
	return nil
}

func (r *fakeTripRepository) FindByID(_ context.Context, tripID string) (*tripentity.Trip, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	t, ok := r.byID[tripID]
	if !ok {
		return nil, domainerrors.NotFound("trip not found: " + tripID)
	}
	clone := *t
	return &clone, nil
}

func (r *fakeTripRepository) FindByRiderID(_ context.Context, riderID string) ([]*tripentity.Trip, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var out []*tripentity.Trip
	for _, t := range r.byID {
		if t.RiderID == riderID {
			clone := *t
			out = append(out, &clone)
		}
	}
	return out, nil
}

func (r *fakeTripRepository) FindByDriverID(_ context.Context, driverID string) ([]*tripentity.Trip, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var out []*tripentity.Trip
	for _, t := range r.byID {
		if t.DriverID == driverID {
			clone := *t
			out = append(out, &clone)
		}
	}
	return out, nil
}

// assignDriver mirrors backend/services/dispatch/infrastructure/postgres's
// TripUpdater.AssignDriver — in production, Dispatch and Trip share one
// Postgres database and Dispatch updates the trips row directly with a raw
// SQL UPDATE (status='driver_assigned'), bypassing Trip's own package-private
// invariants because it is a cross-service write. This is the same pattern
// reproduced here: no Trip entity method exists for this transition (by
// design — only Dispatch triggers it), so the fake repository performs the
// equivalent direct field mutation.
func (r *fakeTripRepository) assignDriver(tripID, driverID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	t, ok := r.byID[tripID]
	if !ok {
		return domainerrors.NotFound("trip not found: " + tripID)
	}
	t.Status = tripentity.StatusDriverAssigned
	t.DriverID = driverID
	return nil
}
