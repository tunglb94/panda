package integration

import (
	"context"

	dispatchapp "github.com/fairride/dispatch/app"
	dispatchentity "github.com/fairride/dispatch/domain/entity"

	"github.com/fairride/ai_simulation/domain/entity"
)

// DispatchAdapter wraps the real backend/services/dispatch use cases
// (RequestDispatchUseCase/AcceptTripUseCase/RejectTripUseCase) over the
// in-memory fakes in dispatch_fakes.go — the actual nearest-driver matching
// algorithm (offerNextDriver) is the production one, unmodified.
//
// Not used: DispatchEngine (the background auto-retry worker). It polls on
// a real wall-clock ticker (default 5s), which would desynchronize a
// deterministic tick simulation (1 tick = 1 simulated minute, run as fast
// as the host allows) from real time. Instead, the simulation calls
// RejectTripUseCase directly to model a driver failing to respond in time —
// functionally the same recovery path (job.Reject → offerNextDriver), just
// triggered by simulated-time logic in simulation/ride_flow.go instead of a
// real timer. This is a documented scope simplification, not a
// reimplementation of the matching algorithm itself.
type DispatchAdapter struct {
	locationRepo *fakeDriverLocationRepository
	tripUpdater  *fakeTripUpdater

	requestUC *dispatchapp.RequestDispatchUseCase
	acceptUC  *dispatchapp.AcceptTripUseCase
	rejectUC  *dispatchapp.RejectTripUseCase
}

func NewDispatchAdapter() *DispatchAdapter {
	jobRepo := newFakeDispatchJobRepository()
	locationRepo := newFakeDriverLocationRepository()
	tripUpdater := newFakeTripUpdater()
	transactor := &fakeTransactor{jobRepo: jobRepo, tripUpdater: tripUpdater}

	return &DispatchAdapter{
		locationRepo: locationRepo,
		tripUpdater:  tripUpdater,
		requestUC:    dispatchapp.NewRequestDispatchUseCase(jobRepo, locationRepo, transactor),
		acceptUC:     dispatchapp.NewAcceptTripUseCase(jobRepo, transactor),
		rejectUC:     dispatchapp.NewRejectTripUseCase(jobRepo, locationRepo, tripUpdater),
	}
}

// SetDriverPosition publishes a driver's current city-plane coordinates,
// service type, and trip-type capability — the simulation calls this
// whenever a driver's zone changes (goes online, completes a trip,
// relocates) so FindNearby has fresh data, mirroring how production
// drivers periodically push GPS updates. serviceType is optional (empty =
// not reported, matching the production contract).
func (a *DispatchAdapter) SetDriverPosition(ctx context.Context, driverID string, x, y float64, serviceType entity.ServiceType, rideEnabled, deliveryEnabled bool) error {
	return a.locationRepo.UpdateLocation(ctx, driverID, x, y, toDispatchServiceType(serviceType), rideEnabled, deliveryEnabled)
}

// RemoveDriver takes a driver out of matching range (going offline).
func (a *DispatchAdapter) RemoveDriver(ctx context.Context, driverID string) error {
	return a.locationRepo.RemoveLocation(ctx, driverID)
}

// RequestRide creates a dispatch job and offers it to the nearest available
// driver reporting a matching service type and RideEnabled capability —
// calls the real RequestDispatchUseCase. serviceType is checked against
// entity.ServiceType.IsSupported's allow-list before any job is created
// (Vehicle/Service Catalog refactor) and used by offerNextDriver to
// exclude non-matching candidates.
func (a *DispatchAdapter) RequestRide(ctx context.Context, tripID, riderID string, pickupX, pickupY float64, serviceType entity.ServiceType) (*dispatchentity.DispatchJob, error) {
	return a.requestUC.Execute(ctx, dispatchapp.RequestDispatchInput{
		TripID:      tripID,
		RiderID:     riderID,
		PickupLat:   pickupX,
		PickupLon:   pickupY,
		ServiceType: toDispatchServiceType(serviceType),
	})
}

// RequestDelivery is RequestRide's Delivery counterpart — same
// RequestDispatchUseCase, same offerNextDriver matching algorithm
// (unchanged either way, see that use case's own doc comment), just stamped
// with TripType=Delivery. There is no "delivery_bike"/"delivery_car"
// ServiceType — the same 4-value catalog applies to both TripTypes (see
// entity.ServiceType's doc comment); matching a candidate additionally
// requires DeliveryEnabled=true.
func (a *DispatchAdapter) RequestDelivery(ctx context.Context, tripID, riderID string, pickupX, pickupY float64, serviceType entity.ServiceType) (*dispatchentity.DispatchJob, error) {
	return a.requestUC.Execute(ctx, dispatchapp.RequestDispatchInput{
		TripID:      tripID,
		RiderID:     riderID,
		PickupLat:   pickupX,
		PickupLon:   pickupY,
		TripType:    dispatchentity.TripTypeDelivery,
		ServiceType: toDispatchServiceType(serviceType),
	})
}

// toDispatchServiceType maps the simulation's local ServiceType to
// dispatch's own (string-identical by convention, not a shared Go type —
// same pattern the production services themselves use). Every one of the 4
// catalog values has an explicit case — an empty/unrecognized input maps
// to "" (no service type set, i.e. unfiltered), not to a guessed type.
func toDispatchServiceType(s entity.ServiceType) dispatchentity.ServiceType {
	switch s {
	case entity.ServiceBike:
		return dispatchentity.ServiceTypeBike
	case entity.ServiceBikePlus:
		return dispatchentity.ServiceTypeBikePlus
	case entity.ServiceCar:
		return dispatchentity.ServiceTypeCar
	case entity.ServiceCarXL:
		return dispatchentity.ServiceTypeCarXL
	default:
		return ""
	}
}

// Accept records a driver's acceptance — calls the real AcceptTripUseCase.
func (a *DispatchAdapter) Accept(ctx context.Context, tripID, driverID string) (*dispatchentity.DispatchJob, error) {
	return a.acceptUC.Execute(ctx, tripID, driverID)
}

// Reject records a driver's rejection (or a simulated non-response timeout)
// and re-offers to the next nearest driver — calls the real
// RejectTripUseCase, which internally calls the same offerNextDriver
// matching logic RequestDispatchUseCase uses.
func (a *DispatchAdapter) Reject(ctx context.Context, tripID, driverID string) (*dispatchentity.DispatchJob, error) {
	return a.rejectUC.Execute(ctx, tripID, driverID)
}

// AssignedDriver returns the driver AssignDriver was called with for tripID,
// if any — the simulation reads this after RequestRide/Accept to learn the
// match outcome without re-deriving it from DispatchJob state.
func (a *DispatchAdapter) AssignedDriver(tripID string) (string, bool) {
	return a.tripUpdater.AssignedDriver(tripID)
}
