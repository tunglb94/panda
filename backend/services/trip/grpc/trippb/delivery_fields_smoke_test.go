package trippb_test

// Smoke test proving the hand-mutated FileDescriptorProto (Delivery V1
// Phase 2, docs/business/DELIVERY_V1_DESIGN.md) round-trips correctly over
// the real reflection-based proto.Marshal/Unmarshal path. See
// backend/services/booking/grpc/bookingpb's equivalent test for the same
// rationale.

import (
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/fairride/trip/grpc/trippb"
)

func TestCreateTripRequest_DeliveryFieldsRoundTripOverWire(t *testing.T) {
	req := &trippb.CreateTripRequest{
		RiderId:            "r1",
		PickupAddress:      "123 Main St",
		DropoffAddress:     "456 Elm Ave",
		TripType:           "delivery",
		PickupContactName:  "Nguyen Van A",
		PickupContactPhone: "0912345678",
		ReceiverName:       "Tran Thi B",
		ReceiverPhone:      "0987654321",
		PackageNote:        "handle with care",
		PackageValue:       500000,
		PackageWeight:      2.5,
	}
	wire, err := proto.Marshal(req)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	got := &trippb.CreateTripRequest{}
	if err := proto.Unmarshal(wire, got); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if got.GetTripType() != "delivery" {
		t.Errorf("TripType = %q, want delivery", got.GetTripType())
	}
	if got.GetPickupContactName() != "Nguyen Van A" {
		t.Errorf("PickupContactName = %q", got.GetPickupContactName())
	}
	if got.GetReceiverPhone() != "0987654321" {
		t.Errorf("ReceiverPhone = %q", got.GetReceiverPhone())
	}
	if got.GetPackageValue() != 500000 {
		t.Errorf("PackageValue = %d, want 500000", got.GetPackageValue())
	}
	if got.GetPackageWeight() != 2.5 {
		t.Errorf("PackageWeight = %v, want 2.5", got.GetPackageWeight())
	}
	if got.GetRiderId() != "r1" || got.GetPickupAddress() != "123 Main St" {
		t.Errorf("pre-existing Ride fields did not round-trip: %+v", got)
	}
}

func TestCreateTripRequest_RideOnly_OmittedDeliveryFieldsStayEmpty(t *testing.T) {
	req := &trippb.CreateTripRequest{RiderId: "r1", PickupAddress: "p", DropoffAddress: "d"}
	wire, err := proto.Marshal(req)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	got := &trippb.CreateTripRequest{}
	if err := proto.Unmarshal(wire, got); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if got.GetTripType() != "" || got.GetReceiverName() != "" {
		t.Errorf("expected empty delivery fields for a plain Ride request, got %+v", got)
	}
}

func TestTripProto_DeliveryFieldsRoundTrip(t *testing.T) {
	tp := &trippb.TripProto{TripId: "t1", Status: "pending", TripType: "delivery", DeliveryId: "d1"}
	wire, err := proto.Marshal(tp)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	got := &trippb.TripProto{}
	if err := proto.Unmarshal(wire, got); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if got.GetTripType() != "delivery" || got.GetDeliveryId() != "d1" {
		t.Errorf("TripType/DeliveryId did not round-trip: %+v", got)
	}
}

// TestTripProto_DeliveryStatusRoundTrip proves the Delivery V1 Phase 4
// delivery_status field addition (docs/business/DELIVERY_V1_DESIGN.md)
// round-trips over the real wire, same as every prior additive field this
// project has hand-added to a generated .pb.go message.
func TestTripProto_DeliveryStatusRoundTrip(t *testing.T) {
	tp := &trippb.TripProto{TripId: "t1", Status: "in_progress", TripType: "delivery", DeliveryId: "d1", DeliveryStatus: "PARCEL_PICKED_UP"}
	wire, err := proto.Marshal(tp)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	got := &trippb.TripProto{}
	if err := proto.Unmarshal(wire, got); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if got.GetDeliveryStatus() != "PARCEL_PICKED_UP" {
		t.Errorf("DeliveryStatus = %q, want PARCEL_PICKED_UP", got.GetDeliveryStatus())
	}
}

func TestTripProto_RideOnly_DeliveryStatusStaysEmpty(t *testing.T) {
	tp := &trippb.TripProto{TripId: "t1", Status: "pending"}
	wire, err := proto.Marshal(tp)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	got := &trippb.TripProto{}
	if err := proto.Unmarshal(wire, got); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if got.GetDeliveryStatus() != "" {
		t.Errorf("DeliveryStatus = %q, want empty for a Ride trip", got.GetDeliveryStatus())
	}
}

func TestTripProto_RideOnly_DeliveryIdStaysEmpty(t *testing.T) {
	tp := &trippb.TripProto{TripId: "t1", Status: "pending"}
	wire, err := proto.Marshal(tp)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	got := &trippb.TripProto{}
	if err := proto.Unmarshal(wire, got); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if got.GetTripType() != "" || got.GetDeliveryId() != "" {
		t.Errorf("expected empty trip_type/delivery_id for a Ride trip, got %+v", got)
	}
}
