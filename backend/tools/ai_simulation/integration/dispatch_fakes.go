package integration

import (
	"context"
	"math"
	"math/rand"
	"sync"
	"time"

	dispatchentity "github.com/fairride/dispatch/domain/entity"
	dispatchrepo "github.com/fairride/dispatch/domain/repository"
	domainerrors "github.com/fairride/shared/errors"
)

// This file implements the 4 repository interfaces
// backend/services/dispatch/domain/repository declares, entirely in
// memory. These are the simulation's equivalent of that service's own
// infrastructure/postgres and infrastructure/redis implementations — same
// interfaces, no real database. The dispatch app-layer use cases
// (RequestDispatchUseCase/AcceptTripUseCase/RejectTripUseCase) are used
// completely unmodified against these fakes.

// fakeDispatchJobRepository implements repository.DispatchJobRepository.
type fakeDispatchJobRepository struct {
	mu       sync.Mutex
	byID     map[string]*dispatchentity.DispatchJob
	byTripID map[string]string // tripID -> jobID
}

func newFakeDispatchJobRepository() *fakeDispatchJobRepository {
	return &fakeDispatchJobRepository{
		byID:     make(map[string]*dispatchentity.DispatchJob),
		byTripID: make(map[string]string),
	}
}

func (r *fakeDispatchJobRepository) Save(_ context.Context, job *dispatchentity.DispatchJob) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.byID[job.JobID] = job
	r.byTripID[job.TripID] = job.JobID
	return nil
}

func (r *fakeDispatchJobRepository) FindByID(_ context.Context, jobID string) (*dispatchentity.DispatchJob, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	j, ok := r.byID[jobID]
	if !ok {
		return nil, domainerrors.NotFound("dispatch job not found")
	}
	return j, nil
}

func (r *fakeDispatchJobRepository) FindByTripID(_ context.Context, tripID string) (*dispatchentity.DispatchJob, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	jobID, ok := r.byTripID[tripID]
	if !ok {
		return nil, domainerrors.NotFound("dispatch job not found for trip")
	}
	return r.byID[jobID], nil
}

// FindExpiredOffers is unused by the simulation (see dispatch_adapter.go doc
// comment: timeouts are modeled as explicit Reject calls, not the
// background DispatchEngine's real-time polling, which would tie a
// deterministic tick simulation to wall-clock time). Implemented correctly
// regardless, for interface completeness and in case a future caller needs it.
func (r *fakeDispatchJobRepository) FindExpiredOffers(_ context.Context, now time.Time) ([]*dispatchentity.DispatchJob, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var out []*dispatchentity.DispatchJob
	for _, j := range r.byID {
		if j.Status == dispatchentity.JobStatusSearching && j.IsOfferExpired(now) {
			out = append(out, j)
		}
	}
	return out, nil
}

func (r *fakeDispatchJobRepository) FindCurrentOfferForDriver(_ context.Context, driverID string) (*dispatchentity.DispatchJob, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, j := range r.byID {
		if j.Status == dispatchentity.JobStatusSearching && j.CurrentDriverID == driverID {
			return j, nil
		}
	}
	return nil, domainerrors.NotFound("no current offer for driver")
}

// fakeDriverLocationRepository implements repository.DriverLocationRepository
// over the simulation's own city-plane coordinates (see domain/entity/city.go)
// instead of real GPS lat/lon — the interface only cares about two floats and
// a distance metric, so this is a faithful stand-in for Redis GEO, not a
// simplified reimplementation of the matching algorithm itself.
type fakeDriverLocationRepository struct {
	mu              sync.Mutex
	positions       map[string][2]float64 // driverID -> (x, y)
	active          map[string]bool
	serviceTypes    map[string]dispatchentity.ServiceType
	rideEnabled     map[string]bool
	deliveryEnabled map[string]bool
	// rng breaks exact-distance ties in FindNearby (see shuffleTiedGroups) —
	// a separate source from the simulation's main Rand so this doesn't
	// change the random-draw sequence every other subsystem depends on.
	rng *rand.Rand
}

func newFakeDriverLocationRepository(rng *rand.Rand) *fakeDriverLocationRepository {
	return &fakeDriverLocationRepository{
		positions:       make(map[string][2]float64),
		active:          make(map[string]bool),
		serviceTypes:    make(map[string]dispatchentity.ServiceType),
		rideEnabled:     make(map[string]bool),
		deliveryEnabled: make(map[string]bool),
		rng:             rng,
	}
}

// serviceType is optional (empty = not reported), matching the real Redis
// implementation's contract (backend/services/dispatch/infrastructure/redis).
func (r *fakeDriverLocationRepository) UpdateLocation(_ context.Context, driverID string, lat, lon float64, serviceType dispatchentity.ServiceType, rideEnabled, deliveryEnabled bool) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.positions[driverID] = [2]float64{lat, lon}
	r.active[driverID] = true
	if serviceType != "" {
		r.serviceTypes[driverID] = serviceType
	}
	r.rideEnabled[driverID] = rideEnabled
	r.deliveryEnabled[driverID] = deliveryEnabled
	return nil
}

func (r *fakeDriverLocationRepository) FindNearby(_ context.Context, lat, lon, radiusKM float64, limit int) ([]*dispatchentity.NearbyDriver, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	var candidates []locationCandidate
	for id, pos := range r.positions {
		if !r.active[id] {
			continue
		}
		dx, dy := pos[0]-lat, pos[1]-lon
		dist := math.Sqrt(dx*dx + dy*dy)
		if dist <= radiusKM {
			candidates = append(candidates, locationCandidate{id: id, dist: dist})
		}
	}
	sortCandidatesByDistance(candidates)
	shuffleTiedGroups(candidates, r.rng)
	if len(candidates) > limit {
		candidates = candidates[:limit]
	}
	out := make([]*dispatchentity.NearbyDriver, len(candidates))
	for i, c := range candidates {
		out[i] = &dispatchentity.NearbyDriver{
			DriverID:        c.id,
			ServiceType:     r.serviceTypes[c.id],
			RideEnabled:     r.rideEnabled[c.id],
			DeliveryEnabled: r.deliveryEnabled[c.id],
		}
	}
	return out, nil
}

type locationCandidate struct {
	id   string
	dist float64
}

func sortCandidatesByDistance(c []locationCandidate) {
	// Simple insertion sort — candidate lists are small (search radius +
	// driver-count bounded), no need for sort.Slice's reflection overhead
	// on a hot simulation path.
	//
	// Tie-break by id when dist is equal (drivers in the same zone share
	// the exact same city-plane coordinates, so exact ties are common) —
	// c is built by ranging over a map (see FindNearby), whose iteration
	// order Go randomizes per-process, so without this secondary key the
	// same --seed picks a different "nearest" driver among tied
	// candidates on every run (the --seed determinism bug). This id order
	// is only a deterministic STARTING point — shuffleTiedGroups (called
	// right after, in FindNearby) randomizes which driver actually wins
	// each tied group, so the same low-numbered id doesn't win every tie.
	for i := 1; i < len(c); i++ {
		for j := i; j > 0 && less(c[j], c[j-1]); j-- {
			c[j], c[j-1] = c[j-1], c[j]
		}
	}
}

func less(a, b locationCandidate) bool {
	if a.dist != b.dist {
		return a.dist < b.dist
	}
	return a.id < b.id
}

// shuffleTiedGroups randomizes candidate order within each run of exactly
// equal distance, using rng. Fixes a real dispatch-fairness bug: because
// this simulation's city-plane model gives every driver in the same zone
// the identical (x, y), FindNearby's candidates are tied on distance far
// more often than not — sortCandidatesByDistance's id tie-break (needed for
// --seed determinism, see its doc comment) meant the same lowest-id driver
// in a zone won the "nearest driver" slot on essentially every request,
// which is how a handful of low-numbered drivers (driver-0002, -0004,
// -0007...) ended up receiving 10-40x the trips/income of their peers (see
// CHANGELOG). Shuffling within each tied group keeps genuine distance
// ordering intact (a real nearest driver still wins over a farther one)
// while giving every driver in a tied group an equal chance, and remains
// fully deterministic for a given --seed since rng is seeded from it.
func shuffleTiedGroups(c []locationCandidate, rng *rand.Rand) {
	start := 0
	for i := 1; i <= len(c); i++ {
		if i == len(c) || c[i].dist != c[start].dist {
			if i-start > 1 {
				group := c[start:i]
				rng.Shuffle(len(group), func(a, b int) { group[a], group[b] = group[b], group[a] })
			}
			start = i
		}
	}
}

func (r *fakeDriverLocationRepository) IsActive(_ context.Context, driverID string) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.active[driverID], nil
}

func (r *fakeDriverLocationRepository) GetLocation(_ context.Context, driverID string) (float64, float64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	pos, ok := r.positions[driverID]
	if !ok {
		return 0, 0, domainerrors.NotFound("driver location not found")
	}
	return pos[0], pos[1], nil
}

func (r *fakeDriverLocationRepository) RemoveLocation(_ context.Context, driverID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.positions, driverID)
	delete(r.active, driverID)
	delete(r.serviceTypes, driverID)
	delete(r.rideEnabled, driverID)
	delete(r.deliveryEnabled, driverID)
	return nil
}

// fakeTripUpdater implements repository.TripUpdater by recording the two
// trip-status transitions dispatch needs to trigger, queryable by the
// simulation's own ride_flow.go to learn the outcome.
type fakeTripUpdater struct {
	mu       sync.Mutex
	assigned map[string]string // tripID -> driverID
}

func newFakeTripUpdater() *fakeTripUpdater {
	return &fakeTripUpdater{assigned: make(map[string]string)}
}

func (t *fakeTripUpdater) SetSearching(_ context.Context, _ string, _ time.Time) error {
	return nil // no separate state to track — the DispatchJob's own status is authoritative
}

func (t *fakeTripUpdater) AssignDriver(_ context.Context, tripID, driverID string, _ time.Time) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.assigned[tripID] = driverID
	return nil
}

func (t *fakeTripUpdater) AssignedDriver(tripID string) (string, bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	id, ok := t.assigned[tripID]
	return id, ok
}

// fakeTransactor implements repository.Transactor with no real transaction
// semantics — the simulation processes one ride-request at a time on a
// single goroutine per tick, so there is no concurrent-write hazard for
// this in-memory implementation to guard against.
type fakeTransactor struct {
	jobRepo     *fakeDispatchJobRepository
	tripUpdater *fakeTripUpdater
}

func (t *fakeTransactor) WithinTx(ctx context.Context, fn func(jobs dispatchrepo.DispatchJobRepository, trips dispatchrepo.TripUpdater) error) error {
	return fn(t.jobRepo, t.tripUpdater)
}
