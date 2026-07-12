package integration

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	tripapp "github.com/fairride/trip/app"
	tripentity "github.com/fairride/trip/domain/entity"
	tripmemory "github.com/fairride/trip/infrastructure/memory"
)

// DeliveryAdapter wraps the real backend/services/trip Delivery use cases
// (CreateTripUseCase/AcceptDeliveryUseCase/PickupParcelUseCase/
// StartDeliveryUseCase/CompleteDeliveryUseCase) — the actual production
// Delivery state machine (Created -> Accepted -> ParcelPickedUp ->
// InDelivery -> Delivered -> Completed on the Delivery aggregate, with
// Trip.Status tracking DriverAssigned/DriverArrived/InProgress alongside
// it), not a simulation-local reimplementation.
//
// DeliveryRepository reuses the Trip service's own in-memory implementation
// (tripmemory.NewDeliveryRepository()) unmodified — it is already a
// self-contained fake with no Postgres dependency (no Postgres-backed
// DeliveryRepository exists in production yet either, see that package's
// doc comment). TripRepository has no equivalent reusable fake (production
// has no Trip repository at all beyond Postgres), so delivery_fakes.go
// provides one.
type DeliveryAdapter struct {
	tripRepo     *fakeTripRepository
	deliveryRepo *tripmemory.DeliveryRepository

	createUC   *tripapp.CreateTripUseCase
	acceptUC   *tripapp.AcceptDeliveryUseCase
	pickupUC   *tripapp.PickupParcelUseCase
	startUC    *tripapp.StartDeliveryUseCase
	completeUC *tripapp.CompleteDeliveryUseCase
}

func NewDeliveryAdapter() *DeliveryAdapter {
	tripRepo := newFakeTripRepository()
	deliveryRepo := tripmemory.NewDeliveryRepository()
	return &DeliveryAdapter{
		tripRepo:     tripRepo,
		deliveryRepo: deliveryRepo,
		createUC:     tripapp.NewCreateTripUseCase(tripRepo, deliveryRepo),
		acceptUC:     tripapp.NewAcceptDeliveryUseCase(tripRepo, deliveryRepo),
		pickupUC:     tripapp.NewPickupParcelUseCase(tripRepo, deliveryRepo),
		startUC:      tripapp.NewStartDeliveryUseCase(tripRepo, deliveryRepo),
		completeUC:   tripapp.NewCompleteDeliveryUseCase(tripRepo, deliveryRepo),
	}
}

// CreateDeliveryInput carries what CreateTrip needs for a delivery booking —
// a direct translation into tripapp.CreateTripInput.
type CreateDeliveryInput struct {
	RiderID        string
	PickupAddress  string
	DropoffAddress string

	SenderName    string
	SenderPhone   string
	ReceiverName  string
	ReceiverPhone string
	PackageNote   string
	PackageValue  int64
	WeightKg      float64
}

// CreateDelivery calls the real CreateTripUseCase with TripType=Delivery —
// this creates BOTH the Delivery aggregate (Status=Created) and its linked
// Trip (Status=Pending), exactly as production's BookDelivery flow would.
// The returned Trip.TripID (service-generated, not caller-supplied — see
// CreateTripUseCase.Execute) is the single ID the caller must use for every
// subsequent call (dispatch, Accept/Pickup/Start/Complete) — mirrors the
// real system, where Booking learns the TripID from this same call before
// requesting dispatch.
func (a *DeliveryAdapter) CreateDelivery(ctx context.Context, in CreateDeliveryInput) (*tripentity.Trip, error) {
	return a.createUC.Execute(ctx, tripapp.CreateTripInput{
		RiderID: in.RiderID, PickupAddress: in.PickupAddress, DropoffAddress: in.DropoffAddress,
		TripType:           tripentity.TripTypeDelivery,
		PickupContactName:  in.SenderName,
		PickupContactPhone: in.SenderPhone,
		ReceiverName:       in.ReceiverName,
		ReceiverPhone:      in.ReceiverPhone,
		PackageNote:        in.PackageNote,
		PackageValue:       in.PackageValue,
		PackageWeightKg:    in.WeightKg,
	})
}

// AssignDriver mirrors production's cross-service Dispatch->Trip write (see
// delivery_fakes.go's assignDriver doc comment): Pending -> DriverAssigned.
func (a *DeliveryAdapter) AssignDriver(tripID, driverID string) error {
	return a.tripRepo.assignDriver(tripID, driverID)
}

// MarkDriverArrived calls the real Trip.MarkDriverArrived (the same,
// unchanged method Ride already uses): DriverAssigned -> DriverArrived.
func (a *DeliveryAdapter) MarkDriverArrived(ctx context.Context, tripID string) error {
	trip, err := a.tripRepo.FindByID(ctx, tripID)
	if err != nil {
		return err
	}
	if err := trip.MarkDriverArrived(time.Now().UTC()); err != nil {
		return err
	}
	return a.tripRepo.Save(ctx, trip)
}

// AcceptDelivery calls the real AcceptDeliveryUseCase: Delivery Created -> Accepted.
func (a *DeliveryAdapter) AcceptDelivery(ctx context.Context, tripID string) (*tripentity.Delivery, error) {
	return a.acceptUC.Execute(ctx, tripapp.AcceptDeliveryInput{TripID: tripID})
}

// PickupParcel calls the real PickupParcelUseCase: Delivery Accepted ->
// ParcelPickedUp, AND Trip DriverArrived -> InProgress in the same call.
func (a *DeliveryAdapter) PickupParcel(ctx context.Context, tripID string) (*tripentity.Delivery, error) {
	return a.pickupUC.Execute(ctx, tripapp.PickupParcelInput{TripID: tripID})
}

// StartDelivery calls the real StartDeliveryUseCase: Delivery ParcelPickedUp -> InDelivery.
func (a *DeliveryAdapter) StartDelivery(ctx context.Context, tripID string) (*tripentity.Delivery, error) {
	return a.startUC.Execute(ctx, tripapp.StartDeliveryInput{TripID: tripID})
}

// CompleteDelivery calls the real CompleteDeliveryUseCase: Delivery
// InDelivery -> Delivered -> Completed. Trip.Status deliberately stays
// InProgress — production itself never advances it further for Delivery
// today (no Pricing<->Delivery integration exists yet, see that use case's
// own doc comment), a real, disclosed backend gap this simulation
// faithfully reproduces rather than papering over.
func (a *DeliveryAdapter) CompleteDelivery(ctx context.Context, tripID string) (*tripentity.Delivery, error) {
	return a.completeUC.Execute(ctx, tripapp.CompleteDeliveryInput{TripID: tripID})
}

// vnPhonePrefixes are the mobile carrier prefixes Delivery's own validation
// (backend/services/trip/domain/entity/delivery.go's vnPhonePattern) accepts
// as the digit right after the leading 0 — 3/5/7/8/9.
var vnPhonePrefixes = []byte{'3', '5', '7', '8', '9'}

// FakeVNPhone generates a syntactically valid Vietnamese mobile number
// (matches vnPhonePattern) for a simulated sender/receiver — no real PII
// involved, just satisfying Delivery's own format validation.
func FakeVNPhone(rnd *rand.Rand) string {
	prefix := vnPhonePrefixes[rnd.Intn(len(vnPhonePrefixes))]
	rest := make([]byte, 8)
	for i := range rest {
		rest[i] = byte('0' + rnd.Intn(10))
	}
	return fmt.Sprintf("0%c%s", prefix, rest)
}
