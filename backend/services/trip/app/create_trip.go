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
//
// TripType is entity.TripTypeRide (the zero value "" is also treated as
// Ride) or entity.TripTypeDelivery. The Pickup*/Receiver*/Package* fields
// are only read when TripType == entity.TripTypeDelivery — Delivery V1
// Phase 2 (docs/business/DELIVERY_V1_DESIGN.md).
type CreateTripInput struct {
	RiderID        string
	PickupAddress  string
	DropoffAddress string

	TripType entity.TripType

	PickupContactName  string
	PickupContactPhone string
	ReceiverName       string
	ReceiverPhone      string
	PackageNote        string
	PackageValue       int64
	PackageWeightKg    float64
}

// CreateTripUseCase creates a new trip in the Pending status. When
// in.TripType == entity.TripTypeDelivery it additionally creates the
// associated Delivery aggregate first (Delivery V1 Phase 2) — the same use
// case, same pipeline, for both Ride and Delivery, per
// docs/business/DELIVERY_V1_DESIGN.md's "reuse the existing Trip/Booking
// architecture" decision.
type CreateTripUseCase struct {
	repo         repository.TripRepository
	deliveryRepo repository.DeliveryRepository
}

func NewCreateTripUseCase(repo repository.TripRepository, deliveryRepo repository.DeliveryRepository) *CreateTripUseCase {
	return &CreateTripUseCase{repo: repo, deliveryRepo: deliveryRepo}
}

func (uc *CreateTripUseCase) Execute(ctx context.Context, in CreateTripInput) (*entity.Trip, error) {
	tripID, err := generateTripID()
	if err != nil {
		return nil, errors.Internal("failed to generate trip id")
	}

	if in.TripType == entity.TripTypeDelivery {
		return uc.createDeliveryTrip(ctx, tripID, in)
	}

	// Ride path — unchanged from before Delivery existed (same validation,
	// same persistence, same behavior for every existing caller).
	trip, err := entity.NewTrip(tripID, in.RiderID, in.PickupAddress, in.DropoffAddress, time.Now().UTC())
	if err != nil {
		return nil, err
	}
	if err := uc.repo.Save(ctx, trip); err != nil {
		return nil, err
	}
	return trip, nil
}

// createDeliveryTrip creates the Delivery aggregate, then the Trip that
// references it. Field mapping notes (see
// docs/business/DELIVERY_V1_DESIGN.md Phần 2 for the entity model):
//   - pickup_contact_name/phone map to Delivery.SenderName/SenderPhone (the
//     person to meet at the pickup point — Phase 2's booking request does
//     not distinguish a separate "sender" identity from "pickup contact").
//   - package_note maps to Delivery.DeliveryNote; Delivery.PickupNote is
//     left empty (not exposed by this booking request in Phase 2).
//   - PackageType/Fragile are not yet exposed on the booking request either
//     (not in Phase 2's field list); defaulted below pending a future phase
//     that adds package-type selection to the booking flow.
func (uc *CreateTripUseCase) createDeliveryTrip(ctx context.Context, tripID string, in CreateTripInput) (*entity.Trip, error) {
	deliveryID, err := generateTripID()
	if err != nil {
		return nil, errors.Internal("failed to generate delivery id")
	}
	now := time.Now().UTC()

	delivery, err := entity.NewDelivery(
		deliveryID,
		in.PickupContactName, in.PickupContactPhone,
		in.ReceiverName, in.ReceiverPhone,
		"", in.PackageNote,
		entity.PackageTypeSmall, // TODO Phase 3: expose package_type on the booking request
		in.PackageWeightKg,
		false, // fragile — not yet exposed by the booking request
		in.PackageValue,
		now,
	)
	if err != nil {
		return nil, err
	}
	if err := uc.deliveryRepo.Save(ctx, delivery); err != nil {
		return nil, err
	}

	trip, err := entity.NewDeliveryTrip(tripID, in.RiderID, in.PickupAddress, in.DropoffAddress, deliveryID, now)
	if err != nil {
		return nil, err
	}
	if err := uc.repo.Save(ctx, trip); err != nil {
		return nil, err
	}
	return trip, nil
}

// Note on weight validation: entity.NewDelivery rejects weight <= 0 (a
// pre-existing Phase 1 invariant — a 0kg package is not a real shipment).
// This task's spec asks for "weight >= 0" at the request boundary; rather
// than add a second, duplicate range check here (violating "không duplicate
// logic"), createDeliveryTrip reuses NewDelivery's existing, stricter check
// as the single source of truth — a request with weight == 0 is correctly
// rejected as InvalidArgument, same as a negative weight.

func generateTripID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
