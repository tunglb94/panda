package entity_test

import (
	"testing"

	domainerrors "github.com/fairride/shared/errors"
	"github.com/fairride/trip/domain/entity"
)

const (
	validSenderPhone   = "0912345678"
	validReceiverPhone = "+84987654321"
)

func validDeliveryArgs() (deliveryID, senderName, senderPhone, receiverName, receiverPhone, pickupNote, deliveryNote string, packageType entity.PackageType, weightKg float64, fragile bool, declaredValue int64) {
	return "d1", "Nguyen Van A", validSenderPhone, "Tran Thi B", validReceiverPhone, "gate code 1234", "leave at reception", entity.PackageTypeSmall, 1.5, false, 500000
}

// ─── NewDelivery — happy path ───────────────────────────────────────────────

func TestNewDelivery_Valid(t *testing.T) {
	id, sName, sPhone, rName, rPhone, pNote, dNote, pkgType, weight, fragile, value := validDeliveryArgs()
	d, err := entity.NewDelivery(id, sName, sPhone, rName, rPhone, pNote, dNote, pkgType, weight, fragile, value, testNow)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d.DeliveryID != "d1" {
		t.Errorf("DeliveryID = %q, want %q", d.DeliveryID, "d1")
	}
	if d.Status != entity.DeliveryStatusCreated {
		t.Errorf("Status = %q, want CREATED", d.Status)
	}
	if d.CashOnDelivery {
		t.Error("CashOnDelivery must always be false in Phase 1/2")
	}
	if !d.CreatedAt.Equal(testNow) || !d.UpdatedAt.Equal(testNow) {
		t.Errorf("CreatedAt/UpdatedAt not set to now")
	}
	if d.PackageType != entity.PackageTypeSmall {
		t.Errorf("PackageType = %q, want SMALL", d.PackageType)
	}
	if d.EstimatedWeightKg != 1.5 {
		t.Errorf("EstimatedWeightKg = %v, want 1.5", d.EstimatedWeightKg)
	}
	if d.DeclaredValue != 500000 {
		t.Errorf("DeclaredValue = %d, want 500000", d.DeclaredValue)
	}
}

func TestNewDelivery_AllPackageTypesValid(t *testing.T) {
	for _, pt := range []entity.PackageType{
		entity.PackageTypeDocument,
		entity.PackageTypeSmall,
		entity.PackageTypeMedium,
		entity.PackageTypeLarge,
	} {
		id, sName, sPhone, rName, rPhone, pNote, dNote, _, weight, fragile, value := validDeliveryArgs()
		if _, err := entity.NewDelivery(id, sName, sPhone, rName, rPhone, pNote, dNote, pt, weight, fragile, value, testNow); err != nil {
			t.Errorf("package type %q: unexpected error: %v", pt, err)
		}
	}
}

func TestNewDelivery_ZeroDeclaredValueIsValid(t *testing.T) {
	id, sName, sPhone, rName, rPhone, pNote, dNote, pkgType, weight, fragile, _ := validDeliveryArgs()
	d, err := entity.NewDelivery(id, sName, sPhone, rName, rPhone, pNote, dNote, pkgType, weight, fragile, 0, testNow)
	if err != nil {
		t.Fatalf("unexpected error with declared value 0: %v", err)
	}
	if d.DeclaredValue != 0 {
		t.Errorf("DeclaredValue = %d, want 0", d.DeclaredValue)
	}
}

// ─── NewDelivery — validation ───────────────────────────────────────────────

func TestNewDelivery_EmptyDeliveryID(t *testing.T) {
	_, sName, sPhone, rName, rPhone, pNote, dNote, pkgType, weight, fragile, value := validDeliveryArgs()
	_, err := entity.NewDelivery("", sName, sPhone, rName, rPhone, pNote, dNote, pkgType, weight, fragile, value, testNow)
	if !domainerrors.IsCode(err, domainerrors.CodeInvalidArgument) {
		t.Errorf("expected InvalidArgument, got %v", err)
	}
}

func TestNewDelivery_EmptySenderName(t *testing.T) {
	id, _, sPhone, rName, rPhone, pNote, dNote, pkgType, weight, fragile, value := validDeliveryArgs()
	_, err := entity.NewDelivery(id, "  ", sPhone, rName, rPhone, pNote, dNote, pkgType, weight, fragile, value, testNow)
	if !domainerrors.IsCode(err, domainerrors.CodeInvalidArgument) {
		t.Errorf("expected InvalidArgument, got %v", err)
	}
}

func TestNewDelivery_InvalidSenderPhone(t *testing.T) {
	id, sName, _, rName, rPhone, pNote, dNote, pkgType, weight, fragile, value := validDeliveryArgs()
	for _, bad := range []string{"", "abc", "123", "0012345678"} {
		_, err := entity.NewDelivery(id, sName, bad, rName, rPhone, pNote, dNote, pkgType, weight, fragile, value, testNow)
		if !domainerrors.IsCode(err, domainerrors.CodeInvalidArgument) {
			t.Errorf("sender phone %q: expected InvalidArgument, got %v", bad, err)
		}
	}
}

func TestNewDelivery_EmptyReceiverName(t *testing.T) {
	id, sName, sPhone, _, rPhone, pNote, dNote, pkgType, weight, fragile, value := validDeliveryArgs()
	_, err := entity.NewDelivery(id, sName, sPhone, "", rPhone, pNote, dNote, pkgType, weight, fragile, value, testNow)
	if !domainerrors.IsCode(err, domainerrors.CodeInvalidArgument) {
		t.Errorf("expected InvalidArgument, got %v", err)
	}
}

func TestNewDelivery_InvalidReceiverPhone(t *testing.T) {
	id, sName, sPhone, rName, _, pNote, dNote, pkgType, weight, fragile, value := validDeliveryArgs()
	for _, bad := range []string{"", "not-a-phone", "84987654321"} {
		_, err := entity.NewDelivery(id, sName, sPhone, rName, bad, pNote, dNote, pkgType, weight, fragile, value, testNow)
		if !domainerrors.IsCode(err, domainerrors.CodeInvalidArgument) {
			t.Errorf("receiver phone %q: expected InvalidArgument, got %v", bad, err)
		}
	}
}

func TestNewDelivery_InvalidPackageType(t *testing.T) {
	id, sName, sPhone, rName, rPhone, pNote, dNote, _, weight, fragile, value := validDeliveryArgs()
	_, err := entity.NewDelivery(id, sName, sPhone, rName, rPhone, pNote, dNote, entity.PackageType("HEAVY_MACHINERY"), weight, fragile, value, testNow)
	if !domainerrors.IsCode(err, domainerrors.CodeInvalidArgument) {
		t.Errorf("expected InvalidArgument, got %v", err)
	}
}

func TestNewDelivery_WeightMustBePositive(t *testing.T) {
	id, sName, sPhone, rName, rPhone, pNote, dNote, pkgType, _, fragile, value := validDeliveryArgs()
	for _, w := range []float64{0, -1, -0.5} {
		_, err := entity.NewDelivery(id, sName, sPhone, rName, rPhone, pNote, dNote, pkgType, w, fragile, value, testNow)
		if !domainerrors.IsCode(err, domainerrors.CodeInvalidArgument) {
			t.Errorf("weight %v: expected InvalidArgument, got %v", w, err)
		}
	}
}

func TestNewDelivery_DeclaredValueMustNotBeNegative(t *testing.T) {
	id, sName, sPhone, rName, rPhone, pNote, dNote, pkgType, weight, fragile, _ := validDeliveryArgs()
	for _, v := range []int64{-1, -500000} {
		_, err := entity.NewDelivery(id, sName, sPhone, rName, rPhone, pNote, dNote, pkgType, weight, fragile, v, testNow)
		if !domainerrors.IsCode(err, domainerrors.CodeInvalidArgument) {
			t.Errorf("declared value %d: expected InvalidArgument, got %v", v, err)
		}
	}
}

func TestNewDelivery_NotesAreOptional(t *testing.T) {
	id, sName, sPhone, rName, rPhone, _, _, pkgType, weight, fragile, value := validDeliveryArgs()
	d, err := entity.NewDelivery(id, sName, sPhone, rName, rPhone, "", "", pkgType, weight, fragile, value, testNow)
	if err != nil {
		t.Fatalf("unexpected error with empty notes: %v", err)
	}
	if d.PickupNote != "" || d.DeliveryNote != "" {
		t.Errorf("expected empty notes to round-trip empty")
	}
}

// ─── Status transitions (Delivery V1 Phase 4, docs/business/DELIVERY_V1_DESIGN.md) ──

func deliveryAt(status entity.DeliveryStatus) *entity.Delivery {
	return entity.ReconstituteDelivery(
		"d1", "Nguyen Van A", validSenderPhone, "Tran Thi B", validReceiverPhone,
		"", "", entity.PackageTypeSmall, 1.5, false, false, 500000,
		status, testNow, testNow,
	)
}

// allDeliveryStatuses is used by the exhaustive "no skip state" matrix below.
var allDeliveryStatuses = []entity.DeliveryStatus{
	entity.DeliveryStatusCreated,
	entity.DeliveryStatusAccepted,
	entity.DeliveryStatusParcelPickedUp,
	entity.DeliveryStatusInDelivery,
	entity.DeliveryStatusDelivered,
	entity.DeliveryStatusCompleted,
	entity.DeliveryStatusCancelled,
}

func TestAcceptByDriver_FromCreated(t *testing.T) {
	d := deliveryAt(entity.DeliveryStatusCreated)
	if err := d.AcceptByDriver(testNow); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d.Status != entity.DeliveryStatusAccepted {
		t.Errorf("Status = %q, want ACCEPTED", d.Status)
	}
}

func TestAcceptByDriver_FromNonCreatedFails(t *testing.T) {
	for _, st := range []entity.DeliveryStatus{
		entity.DeliveryStatusAccepted,
		entity.DeliveryStatusParcelPickedUp,
		entity.DeliveryStatusInDelivery,
		entity.DeliveryStatusDelivered,
		entity.DeliveryStatusCompleted,
		entity.DeliveryStatusCancelled,
	} {
		d := deliveryAt(st)
		err := d.AcceptByDriver(testNow)
		if !domainerrors.IsCode(err, domainerrors.CodePreconditionFailed) {
			t.Errorf("from %q: expected PreconditionFailed, got %v", st, err)
		}
	}
}

func TestMarkParcelPickedUp_FromAccepted(t *testing.T) {
	d := deliveryAt(entity.DeliveryStatusAccepted)
	if err := d.MarkParcelPickedUp(testNow); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d.Status != entity.DeliveryStatusParcelPickedUp {
		t.Errorf("Status = %q, want PARCEL_PICKED_UP", d.Status)
	}
}

func TestMarkParcelPickedUp_FromCreatedFails(t *testing.T) {
	// "Created -> ParcelPickedUp" would skip Accepted — must be rejected.
	d := deliveryAt(entity.DeliveryStatusCreated)
	err := d.MarkParcelPickedUp(testNow)
	if !domainerrors.IsCode(err, domainerrors.CodePreconditionFailed) {
		t.Errorf("expected PreconditionFailed, got %v", err)
	}
}

func TestStartDelivery_FromParcelPickedUp(t *testing.T) {
	d := deliveryAt(entity.DeliveryStatusParcelPickedUp)
	if err := d.StartDelivery(testNow); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d.Status != entity.DeliveryStatusInDelivery {
		t.Errorf("Status = %q, want IN_DELIVERY", d.Status)
	}
}

func TestStartDelivery_FromAcceptedFails(t *testing.T) {
	// "Accepted -> InDelivery" would skip ParcelPickedUp — must be rejected.
	d := deliveryAt(entity.DeliveryStatusAccepted)
	err := d.StartDelivery(testNow)
	if !domainerrors.IsCode(err, domainerrors.CodePreconditionFailed) {
		t.Errorf("expected PreconditionFailed, got %v", err)
	}
}

func TestMarkDelivered_FromInDelivery(t *testing.T) {
	d := deliveryAt(entity.DeliveryStatusInDelivery)
	if err := d.MarkDelivered(testNow); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d.Status != entity.DeliveryStatusDelivered {
		t.Errorf("Status = %q, want DELIVERED", d.Status)
	}
}

func TestMarkDelivered_FromCreatedFails(t *testing.T) {
	// Explicit task case: "Created -> Delivered" must be rejected.
	d := deliveryAt(entity.DeliveryStatusCreated)
	err := d.MarkDelivered(testNow)
	if !domainerrors.IsCode(err, domainerrors.CodePreconditionFailed) {
		t.Errorf("expected PreconditionFailed, got %v", err)
	}
}

func TestMarkDelivered_FromParcelPickedUpFails(t *testing.T) {
	// "ParcelPickedUp -> Delivered" would skip InDelivery — must be rejected.
	d := deliveryAt(entity.DeliveryStatusParcelPickedUp)
	err := d.MarkDelivered(testNow)
	if !domainerrors.IsCode(err, domainerrors.CodePreconditionFailed) {
		t.Errorf("expected PreconditionFailed, got %v", err)
	}
}

func TestCompleteDelivery_FromDelivered(t *testing.T) {
	d := deliveryAt(entity.DeliveryStatusDelivered)
	if err := d.CompleteDelivery(testNow); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d.Status != entity.DeliveryStatusCompleted {
		t.Errorf("Status = %q, want COMPLETED", d.Status)
	}
}

func TestCompleteDelivery_FromParcelPickedUpFails(t *testing.T) {
	// Explicit task case: "Pickup -> Completed" must be rejected.
	d := deliveryAt(entity.DeliveryStatusParcelPickedUp)
	err := d.CompleteDelivery(testNow)
	if !domainerrors.IsCode(err, domainerrors.CodePreconditionFailed) {
		t.Errorf("expected PreconditionFailed, got %v", err)
	}
}

func TestCompleteDelivery_FromCompletedFails(t *testing.T) {
	// Explicit task case: "Completed -> Pickup" is covered by
	// TestMarkParcelPickedUp above (Completed is not Accepted); this
	// covers the symmetric "already completed, cannot complete again".
	d := deliveryAt(entity.DeliveryStatusCompleted)
	err := d.CompleteDelivery(testNow)
	if !domainerrors.IsCode(err, domainerrors.CodePreconditionFailed) {
		t.Errorf("expected PreconditionFailed, got %v", err)
	}
}

func TestMarkParcelPickedUp_FromCompletedFails(t *testing.T) {
	// Explicit task case: "Completed -> Pickup" must be rejected.
	d := deliveryAt(entity.DeliveryStatusCompleted)
	err := d.MarkParcelPickedUp(testNow)
	if !domainerrors.IsCode(err, domainerrors.CodePreconditionFailed) {
		t.Errorf("expected PreconditionFailed, got %v", err)
	}
}

func TestAcceptByDriver_FromDeliveredFails(t *testing.T) {
	// Explicit task case: "Delivered -> Accepted" must be rejected.
	d := deliveryAt(entity.DeliveryStatusDelivered)
	err := d.AcceptByDriver(testNow)
	if !domainerrors.IsCode(err, domainerrors.CodePreconditionFailed) {
		t.Errorf("expected PreconditionFailed, got %v", err)
	}
}

// TestDeliveryLifecycle_NoStateMaySkipAhead is the exhaustive "mọi
// transition sai" matrix: for every (status, method) pair, the method must
// succeed if and only if status is exactly that method's single legal
// precondition — every other starting status must be rejected. This proves
// no method can be used to skip a state in the Created → Accepted →
// ParcelPickedUp → InDelivery → Delivered → Completed chain.
func TestDeliveryLifecycle_NoStateMaySkipAhead(t *testing.T) {
	type transition struct {
		name  string
		from  entity.DeliveryStatus // the one legal precondition
		apply func(d *entity.Delivery) error
	}
	transitions := []transition{
		{"AcceptByDriver", entity.DeliveryStatusCreated, func(d *entity.Delivery) error { return d.AcceptByDriver(testNow) }},
		{"MarkParcelPickedUp", entity.DeliveryStatusAccepted, func(d *entity.Delivery) error { return d.MarkParcelPickedUp(testNow) }},
		{"StartDelivery", entity.DeliveryStatusParcelPickedUp, func(d *entity.Delivery) error { return d.StartDelivery(testNow) }},
		{"MarkDelivered", entity.DeliveryStatusInDelivery, func(d *entity.Delivery) error { return d.MarkDelivered(testNow) }},
		{"CompleteDelivery", entity.DeliveryStatusDelivered, func(d *entity.Delivery) error { return d.CompleteDelivery(testNow) }},
	}

	for _, tr := range transitions {
		for _, from := range allDeliveryStatuses {
			d := deliveryAt(from)
			err := tr.apply(d)
			if from == tr.from {
				if err != nil {
					t.Errorf("%s from %q: expected success, got %v", tr.name, from, err)
				}
			} else {
				if !domainerrors.IsCode(err, domainerrors.CodePreconditionFailed) {
					t.Errorf("%s from %q: expected PreconditionFailed (would skip/reorder state), got %v", tr.name, from, err)
				}
			}
		}
	}
}

func TestCancel_FromCreated_Delivery(t *testing.T) {
	d := deliveryAt(entity.DeliveryStatusCreated)
	if err := d.Cancel(testNow); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d.Status != entity.DeliveryStatusCancelled {
		t.Errorf("Status = %q, want CANCELLED", d.Status)
	}
}

func TestCancel_FromAccepted_Delivery(t *testing.T) {
	d := deliveryAt(entity.DeliveryStatusAccepted)
	if err := d.Cancel(testNow); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d.Status != entity.DeliveryStatusCancelled {
		t.Errorf("Status = %q, want CANCELLED", d.Status)
	}
}

func TestCancel_FromParcelPickedUpFails_Delivery(t *testing.T) {
	d := deliveryAt(entity.DeliveryStatusParcelPickedUp)
	err := d.Cancel(testNow)
	if !domainerrors.IsCode(err, domainerrors.CodePreconditionFailed) {
		t.Errorf("expected PreconditionFailed, got %v", err)
	}
}

func TestCancel_FromInDeliveryFails(t *testing.T) {
	d := deliveryAt(entity.DeliveryStatusInDelivery)
	err := d.Cancel(testNow)
	if !domainerrors.IsCode(err, domainerrors.CodePreconditionFailed) {
		t.Errorf("expected PreconditionFailed, got %v", err)
	}
}

func TestCancel_FromDeliveredFails(t *testing.T) {
	d := deliveryAt(entity.DeliveryStatusDelivered)
	err := d.Cancel(testNow)
	if !domainerrors.IsCode(err, domainerrors.CodePreconditionFailed) {
		t.Errorf("expected PreconditionFailed, got %v", err)
	}
}

func TestCancel_FromCompletedFails(t *testing.T) {
	d := deliveryAt(entity.DeliveryStatusCompleted)
	err := d.Cancel(testNow)
	if !domainerrors.IsCode(err, domainerrors.CodePreconditionFailed) {
		t.Errorf("expected PreconditionFailed, got %v", err)
	}
}

func TestCancel_AlreadyCancelledFails(t *testing.T) {
	d := deliveryAt(entity.DeliveryStatusCancelled)
	err := d.Cancel(testNow)
	if !domainerrors.IsCode(err, domainerrors.CodePreconditionFailed) {
		t.Errorf("expected PreconditionFailed, got %v", err)
	}
}

func TestIsCancellable_Delivery(t *testing.T) {
	cases := map[entity.DeliveryStatus]bool{
		entity.DeliveryStatusCreated:        true,
		entity.DeliveryStatusAccepted:       true,
		entity.DeliveryStatusParcelPickedUp: false,
		entity.DeliveryStatusInDelivery:     false,
		entity.DeliveryStatusDelivered:      false,
		entity.DeliveryStatusCompleted:      false,
		entity.DeliveryStatusCancelled:      false,
	}
	for status, want := range cases {
		d := deliveryAt(status)
		if got := d.IsCancellable(); got != want {
			t.Errorf("IsCancellable at %q = %v, want %v", status, got, want)
		}
	}
}

// ─── Reconstitute ────────────────────────────────────────────────────────────

func TestReconstituteDelivery_NoValidation(t *testing.T) {
	d := entity.ReconstituteDelivery(
		"", "", "", "", "", "", "",
		entity.PackageType("garbage"), -5, true, true, -1,
		entity.DeliveryStatusDelivered, testNow, testNow,
	)
	if d == nil {
		t.Fatal("expected non-nil delivery")
	}
	if d.Status != entity.DeliveryStatusDelivered {
		t.Errorf("Status = %q, want DELIVERED", d.Status)
	}
}

// ─── PackageType.IsValid ─────────────────────────────────────────────────────

func TestPackageType_IsValid(t *testing.T) {
	valid := []entity.PackageType{entity.PackageTypeDocument, entity.PackageTypeSmall, entity.PackageTypeMedium, entity.PackageTypeLarge}
	for _, pt := range valid {
		if !pt.IsValid() {
			t.Errorf("%q should be valid", pt)
		}
	}
	invalid := []entity.PackageType{"", "document", "XL", "HEAVY"}
	for _, pt := range invalid {
		if pt.IsValid() {
			t.Errorf("%q should not be valid", pt)
		}
	}
}

// ─── Trip mapping (docs/business/DELIVERY_V1_DESIGN.md Phần 5) ─────────────

func TestNewTrip_RideDefaultsTripTypeAndNilDeliveryID(t *testing.T) {
	trip, err := entity.NewTrip("t1", "r1", "pickup", "dropoff", testNow)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if trip.TripType != entity.TripTypeRide {
		t.Errorf("TripType = %q, want %q", trip.TripType, entity.TripTypeRide)
	}
	if trip.DeliveryID != "" {
		t.Errorf("DeliveryID = %q, want empty (nil) for a Ride trip", trip.DeliveryID)
	}
}

func TestNewDeliveryTrip_Valid(t *testing.T) {
	trip, err := entity.NewDeliveryTrip("t2", "r1", "pickup", "dropoff", "d1", testNow)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if trip.TripType != entity.TripTypeDelivery {
		t.Errorf("TripType = %q, want %q", trip.TripType, entity.TripTypeDelivery)
	}
	if trip.DeliveryID != "d1" {
		t.Errorf("DeliveryID = %q, want %q", trip.DeliveryID, "d1")
	}
}

func TestNewDeliveryTrip_EmptyDeliveryIDFails(t *testing.T) {
	_, err := entity.NewDeliveryTrip("t2", "r1", "pickup", "dropoff", "", testNow)
	if !domainerrors.IsCode(err, domainerrors.CodeInvalidArgument) {
		t.Errorf("expected InvalidArgument, got %v", err)
	}
}

func TestNewDeliveryTrip_InheritsRideValidation(t *testing.T) {
	// empty riderID should still fail via the shared NewTrip validation path
	_, err := entity.NewDeliveryTrip("t2", "", "pickup", "dropoff", "d1", testNow)
	if !domainerrors.IsCode(err, domainerrors.CodeInvalidArgument) {
		t.Errorf("expected InvalidArgument, got %v", err)
	}
}
