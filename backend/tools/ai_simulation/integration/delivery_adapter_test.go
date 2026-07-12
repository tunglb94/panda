package integration

import (
	"context"
	"math/rand"
	"testing"

	tripentity "github.com/fairride/trip/domain/entity"
)

// TestDeliveryAdapter_FullLifecycle drives a delivery through the REAL
// backend/services/trip state machine end-to-end — this is the regression
// test for the whole Delivery integration: if any of the real use cases'
// preconditions ever change (e.g. AcceptDeliveryUseCase's expected source
// status), this test fails loudly instead of silently degrading every
// delivery in a simulation run to Outcome=Cancelled.
func TestDeliveryAdapter_FullLifecycle(t *testing.T) {
	ctx := context.Background()
	a := NewDeliveryAdapter()

	trip, err := a.CreateDelivery(ctx, CreateDeliveryInput{
		RiderID: "rider-1", PickupAddress: "cbd", DropoffAddress: "residential",
		SenderName: "Sender A", SenderPhone: FakeVNPhone(rand.New(rand.NewSource(1))),
		ReceiverName: "Receiver B", ReceiverPhone: FakeVNPhone(rand.New(rand.NewSource(2))),
		PackageNote: "test parcel", PackageValue: 100_000, WeightKg: 2.5,
	})
	if err != nil {
		t.Fatalf("CreateDelivery failed: %v", err)
	}
	if trip.TripID == "" {
		t.Fatalf("expected a non-empty service-generated TripID")
	}
	if trip.TripType != tripentity.TripTypeDelivery {
		t.Errorf("expected TripType=delivery, got %q", trip.TripType)
	}
	if trip.Status != tripentity.StatusPending {
		t.Errorf("expected a freshly created trip to be Pending, got %q", trip.Status)
	}

	if err := a.AssignDriver(trip.TripID, "driver-1"); err != nil {
		t.Fatalf("AssignDriver failed: %v", err)
	}
	if err := a.MarkDriverArrived(ctx, trip.TripID); err != nil {
		t.Fatalf("MarkDriverArrived failed: %v", err)
	}

	delivery, err := a.AcceptDelivery(ctx, trip.TripID)
	if err != nil {
		t.Fatalf("AcceptDelivery failed: %v", err)
	}
	if delivery.Status != tripentity.DeliveryStatusAccepted {
		t.Errorf("expected DeliveryStatusAccepted, got %q", delivery.Status)
	}

	delivery, err = a.PickupParcel(ctx, trip.TripID)
	if err != nil {
		t.Fatalf("PickupParcel failed: %v", err)
	}
	if delivery.Status != tripentity.DeliveryStatusParcelPickedUp {
		t.Errorf("expected DeliveryStatusParcelPickedUp, got %q", delivery.Status)
	}

	delivery, err = a.StartDelivery(ctx, trip.TripID)
	if err != nil {
		t.Fatalf("StartDelivery failed: %v", err)
	}
	if delivery.Status != tripentity.DeliveryStatusInDelivery {
		t.Errorf("expected DeliveryStatusInDelivery, got %q", delivery.Status)
	}

	delivery, err = a.CompleteDelivery(ctx, trip.TripID)
	if err != nil {
		t.Fatalf("CompleteDelivery failed: %v", err)
	}
	if delivery.Status != tripentity.DeliveryStatusCompleted {
		t.Errorf("expected DeliveryStatusCompleted, got %q", delivery.Status)
	}
}

// TestDeliveryAdapter_PickupBeforeAcceptFails confirms the real state
// machine's preconditions are actually enforced (not silently bypassed) —
// calling PickupParcel before AcceptDelivery must fail, since
// DeliveryStatusCreated -> ParcelPickedUp is not a valid transition.
func TestDeliveryAdapter_PickupBeforeAcceptFails(t *testing.T) {
	ctx := context.Background()
	a := NewDeliveryAdapter()

	trip, err := a.CreateDelivery(ctx, CreateDeliveryInput{
		RiderID: "rider-1", PickupAddress: "cbd", DropoffAddress: "residential",
		SenderName: "Sender A", SenderPhone: FakeVNPhone(rand.New(rand.NewSource(1))),
		ReceiverName: "Receiver B", ReceiverPhone: FakeVNPhone(rand.New(rand.NewSource(2))),
		PackageNote: "test parcel", PackageValue: 100_000, WeightKg: 2.5,
	})
	if err != nil {
		t.Fatalf("CreateDelivery failed: %v", err)
	}
	if err := a.AssignDriver(trip.TripID, "driver-1"); err != nil {
		t.Fatalf("AssignDriver failed: %v", err)
	}
	if err := a.MarkDriverArrived(ctx, trip.TripID); err != nil {
		t.Fatalf("MarkDriverArrived failed: %v", err)
	}

	if _, err := a.PickupParcel(ctx, trip.TripID); err == nil {
		t.Fatalf("expected PickupParcel to fail before AcceptDelivery, got nil error")
	}
}

func TestFakeVNPhone_MatchesValidationPattern(t *testing.T) {
	rnd := rand.New(rand.NewSource(42))
	for i := 0; i < 50; i++ {
		phone := FakeVNPhone(rnd)
		if len(phone) != 10 {
			t.Errorf("expected a 10-digit VN phone number, got %q (len %d)", phone, len(phone))
		}
		if phone[0] != '0' {
			t.Errorf("expected phone to start with 0, got %q", phone)
		}
	}
}
