package entity

import (
	"regexp"
	"strings"
	"time"

	"github.com/fairride/shared/errors"
)

// PackageType classifies what is being shipped. Closed set per
// docs/business/DELIVERY_V1_DESIGN.md Phần 3 (the task brief's simplified
// 4-value list) — narrower than the design doc's Phần 7 config-driven
// PackageCategory list, which is a Phase 2+ concern, not Phase 1.
type PackageType string

const (
	PackageTypeDocument PackageType = "DOCUMENT"
	PackageTypeSmall    PackageType = "SMALL"
	PackageTypeMedium   PackageType = "MEDIUM"
	PackageTypeLarge    PackageType = "LARGE"
)

// validPackageTypes is the closed set PackageType.IsValid checks against.
var validPackageTypes = map[PackageType]bool{
	PackageTypeDocument: true,
	PackageTypeSmall:    true,
	PackageTypeMedium:   true,
	PackageTypeLarge:    true,
}

// IsValid reports whether p is one of the four known package types.
func (p PackageType) IsValid() bool {
	return validPackageTypes[p]
}

// DeliveryStatus is the lifecycle stage of a Delivery aggregate. Delivery V1
// Phase 4 (docs/business/DELIVERY_V1_DESIGN.md Phần 3) completes the
// Phase 1 subset (Created/Accepted/PickedUp/Delivered/Cancelled) into the
// full happy-path lifecycle:
//
//	Created → Accepted → [ArrivedPickup, tracked on Trip.Status, see below]
//	  → ParcelPickedUp → InDelivery → Delivered → Completed
//	(Created|Accepted) → Cancelled
//
// "ArrivedPickup" is deliberately NOT a DeliveryStatus value: it is
// Trip.Status = driver_arrived, reached via Trip's existing, UNCHANGED
// MarkDriverArrived() — the exact same RPC/method Ride already uses. This
// is the "TRIP ENTITY: chỉ mở rộng additive" principle applied concretely:
// Delivery reuses Trip's arrival tracking rather than duplicating it as a
// second DeliveryStatus value. Richer states from the full design doc
// (RecipientUnavailable/Returning/Returned/DriverTimeout) remain deferred —
// out of scope per this phase's explicit exclusions (no OTP/COD/tracking).
type DeliveryStatus string

const (
	DeliveryStatusCreated        DeliveryStatus = "CREATED"
	DeliveryStatusAccepted       DeliveryStatus = "ACCEPTED"
	DeliveryStatusParcelPickedUp DeliveryStatus = "PARCEL_PICKED_UP"
	DeliveryStatusInDelivery     DeliveryStatus = "IN_DELIVERY"
	DeliveryStatusDelivered      DeliveryStatus = "DELIVERED"
	DeliveryStatusCompleted      DeliveryStatus = "COMPLETED"
	DeliveryStatusCancelled      DeliveryStatus = "CANCELLED"
)

// deliveryCancellableStatuses mirrors Trip's cancellableStatuses pattern
// (trip.go): a delivery may be cancelled any time before the driver has
// physically picked up the package, matching Ride's own rule that a trip
// cannot be cancelled once under way (BRB §10.1 analogy, see design doc
// Phần 4). Not cancellable once ParcelPickedUp/InDelivery/Delivered/
// Completed/Cancelled.
var deliveryCancellableStatuses = map[DeliveryStatus]bool{
	DeliveryStatusCreated:  true,
	DeliveryStatusAccepted: true,
}

// vnPhonePattern is a basic Vietnamese mobile phone number shape check
// (0xxxxxxxxx or +84xxxxxxxxx, mobile prefixes 3/5/7/8/9). It intentionally
// is not exhaustive telecom validation — just enough to reject empty/garbage
// input, per the task brief's "Validate: Sender phone; Receiver phone".
var vnPhonePattern = regexp.MustCompile(`^(\+84|0)[35789][0-9]{8}$`)

// Delivery is the aggregate root for a FAIRRIDE parcel delivery request. It
// is a standalone aggregate referenced by Trip.DeliveryID (see trip.go's
// NewDeliveryTrip) rather than embedded on Trip, per
// docs/business/DELIVERY_V1_DESIGN.md Phần 1 decision #2.
//
// CashOnDelivery is always false in Phase 1 — COD is explicitly out of scope
// (design doc Phần 4/8/18) and there is no constructor parameter to set it;
// the field exists only so the concept has a named, discoverable place to
// land in a future phase without a breaking struct change.
type Delivery struct {
	DeliveryID        string
	SenderName        string
	SenderPhone       string
	ReceiverName      string
	ReceiverPhone     string
	PickupNote        string
	DeliveryNote      string
	PackageType       PackageType
	EstimatedWeightKg float64
	Fragile           bool
	// DeclaredValue is the sender-declared value of the package (smallest
	// currency unit, matching Trip.FinalFareTotal's convention). Added in
	// Phase 2 (docs/business/DELIVERY_V1_DESIGN.md's Part 4 flagged package
	// liability as an unresolved gap — a declared value is a step toward
	// that, not a full insurance/liability resolution). Zero is valid
	// (no declared value / not insured).
	DeclaredValue  int64
	CashOnDelivery bool // always false in Phase 1/2 — see doc comment above
	Status         DeliveryStatus
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// NewDelivery creates a validated Delivery in DeliveryStatusCreated.
// pickupNote and deliveryNote are optional (may be empty); every other
// field is required and validated per
// docs/business/DELIVERY_V1_DESIGN.md Phần 6 ("Validation").
func NewDelivery(
	deliveryID string,
	senderName, senderPhone string,
	receiverName, receiverPhone string,
	pickupNote, deliveryNote string,
	packageType PackageType,
	estimatedWeightKg float64,
	fragile bool,
	declaredValue int64,
	now time.Time,
) (*Delivery, error) {
	if deliveryID == "" {
		return nil, errors.InvalidArgument("delivery id must not be empty")
	}
	if strings.TrimSpace(senderName) == "" {
		return nil, errors.InvalidArgument("sender name must not be empty")
	}
	if !vnPhonePattern.MatchString(senderPhone) {
		return nil, errors.InvalidArgument("sender phone is not a valid phone number")
	}
	if strings.TrimSpace(receiverName) == "" {
		return nil, errors.InvalidArgument("receiver name must not be empty")
	}
	if !vnPhonePattern.MatchString(receiverPhone) {
		return nil, errors.InvalidArgument("receiver phone is not a valid phone number")
	}
	if !packageType.IsValid() {
		return nil, errors.InvalidArgument("package type is not valid: " + string(packageType))
	}
	if estimatedWeightKg <= 0 {
		return nil, errors.InvalidArgument("estimated weight must be greater than 0")
	}
	if declaredValue < 0 {
		return nil, errors.InvalidArgument("declared value must not be negative")
	}
	return &Delivery{
		DeliveryID:        deliveryID,
		SenderName:        senderName,
		SenderPhone:       senderPhone,
		ReceiverName:      receiverName,
		ReceiverPhone:     receiverPhone,
		PickupNote:        pickupNote,
		DeliveryNote:      deliveryNote,
		PackageType:       packageType,
		EstimatedWeightKg: estimatedWeightKg,
		Fragile:           fragile,
		DeclaredValue:     declaredValue,
		CashOnDelivery:    false,
		Status:            DeliveryStatusCreated,
		CreatedAt:         now,
		UpdatedAt:         now,
	}, nil
}

// ReconstituteDelivery rebuilds a Delivery from a persistence record. No
// validation — mirrors trip.go's ReconstituteTrip convention.
func ReconstituteDelivery(
	deliveryID string,
	senderName, senderPhone string,
	receiverName, receiverPhone string,
	pickupNote, deliveryNote string,
	packageType PackageType,
	estimatedWeightKg float64,
	fragile, cashOnDelivery bool,
	declaredValue int64,
	status DeliveryStatus,
	createdAt, updatedAt time.Time,
) *Delivery {
	return &Delivery{
		DeliveryID:        deliveryID,
		SenderName:        senderName,
		SenderPhone:       senderPhone,
		ReceiverName:      receiverName,
		ReceiverPhone:     receiverPhone,
		PickupNote:        pickupNote,
		DeliveryNote:      deliveryNote,
		PackageType:       packageType,
		EstimatedWeightKg: estimatedWeightKg,
		Fragile:           fragile,
		DeclaredValue:     declaredValue,
		CashOnDelivery:    cashOnDelivery,
		Status:            status,
		CreatedAt:         createdAt,
		UpdatedAt:         updatedAt,
	}
}

// AcceptByDriver transitions the delivery from CREATED to ACCEPTED.
func (d *Delivery) AcceptByDriver(now time.Time) error {
	if d.Status != DeliveryStatusCreated {
		return errors.PreconditionFailed("delivery cannot be driver-accepted from status: " + string(d.Status))
	}
	d.Status = DeliveryStatusAccepted
	d.UpdatedAt = now
	return nil
}

// MarkParcelPickedUp transitions the delivery from ACCEPTED to
// PARCEL_PICKED_UP. Called once the driver has physically collected the
// package — by the time this is called, Trip.Status is expected to already
// be driver_arrived (via Trip's existing MarkDriverArrived, reused
// unchanged for Delivery), but this method only enforces the Delivery-side
// precondition; it does not itself read Trip state (see
// app.PickupParcelUseCase for the combined Trip+Delivery orchestration).
func (d *Delivery) MarkParcelPickedUp(now time.Time) error {
	if d.Status != DeliveryStatusAccepted {
		return errors.PreconditionFailed("delivery cannot be marked picked up from status: " + string(d.Status))
	}
	d.Status = DeliveryStatusParcelPickedUp
	d.UpdatedAt = now
	return nil
}

// StartDelivery transitions the delivery from PARCEL_PICKED_UP to
// IN_DELIVERY — the driver is now en route to the drop-off address. Purely
// a Delivery-internal sub-status; Trip.Status stays in_progress (already
// set by the preceding PickupParcel step), matching
// docs/business/DELIVERY_V1_DESIGN.md Phần 3's "Delivering — Trip.Status
// không đổi" mapping.
func (d *Delivery) StartDelivery(now time.Time) error {
	if d.Status != DeliveryStatusParcelPickedUp {
		return errors.PreconditionFailed("delivery cannot start delivering from status: " + string(d.Status))
	}
	d.Status = DeliveryStatusInDelivery
	d.UpdatedAt = now
	return nil
}

// MarkDelivered transitions the delivery from IN_DELIVERY to DELIVERED.
func (d *Delivery) MarkDelivered(now time.Time) error {
	if d.Status != DeliveryStatusInDelivery {
		return errors.PreconditionFailed("delivery cannot be marked delivered from status: " + string(d.Status))
	}
	d.Status = DeliveryStatusDelivered
	d.UpdatedAt = now
	return nil
}

// CompleteDelivery transitions the delivery from DELIVERED to COMPLETED —
// the terminal happy-path state, mirroring Trip's completed→settled final
// step conceptually (though Delivery has no payment/settlement state of
// its own in this phase; see app.CompleteDeliveryUseCase's doc comment for
// why Trip.Complete() is deliberately NOT called here).
func (d *Delivery) CompleteDelivery(now time.Time) error {
	if d.Status != DeliveryStatusDelivered {
		return errors.PreconditionFailed("delivery cannot be completed from status: " + string(d.Status))
	}
	d.Status = DeliveryStatusCompleted
	d.UpdatedAt = now
	return nil
}

// Cancel transitions the delivery to CANCELLED. Returns PreconditionFailed
// if the package has already been picked up, delivered, completed, or
// already cancelled — mirrors Trip.Cancel's cancellable-status-set pattern.
func (d *Delivery) Cancel(now time.Time) error {
	if !deliveryCancellableStatuses[d.Status] {
		return errors.PreconditionFailed("delivery cannot be cancelled from status: " + string(d.Status))
	}
	d.Status = DeliveryStatusCancelled
	d.UpdatedAt = now
	return nil
}

// IsCancellable reports whether the delivery can currently be cancelled.
func (d *Delivery) IsCancellable() bool {
	return deliveryCancellableStatuses[d.Status]
}
