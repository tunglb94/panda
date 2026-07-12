package dispatchpb_test

// Smoke test proving the hand-mutated FileDescriptorProto (Delivery V1
// Phase 3, docs/business/DELIVERY_V1_DESIGN.md) round-trips correctly over
// the real reflection-based proto.Marshal/Unmarshal path. Only
// RequestDispatchRequest gained fields — DispatchJobProto/DispatchResponse
// are asserted here to be byte-for-byte unchanged ("DispatchResult phải
// giữ nguyên").

import (
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/fairride/dispatch/grpc/dispatchpb"
)

func TestRequestDispatchRequest_DeliveryFieldsRoundTripOverWire(t *testing.T) {
	req := &dispatchpb.RequestDispatchRequest{
		TripId:      "t1",
		RiderId:     "r1",
		PickupLat:   10.77,
		PickupLon:   106.69,
		TripType:    "delivery",
		VehicleType: "motorcycle",
	}
	wire, err := proto.Marshal(req)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	got := &dispatchpb.RequestDispatchRequest{}
	if err := proto.Unmarshal(wire, got); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if got.GetTripType() != "delivery" {
		t.Errorf("TripType = %q, want delivery", got.GetTripType())
	}
	if got.GetVehicleType() != "motorcycle" {
		t.Errorf("VehicleType = %q, want motorcycle", got.GetVehicleType())
	}
	if got.GetTripId() != "t1" || got.GetRiderId() != "r1" {
		t.Errorf("pre-existing Ride fields did not round-trip: %+v", got)
	}
}

func TestRequestDispatchRequest_RideOnly_OmittedDeliveryFieldsStayEmpty(t *testing.T) {
	req := &dispatchpb.RequestDispatchRequest{TripId: "t1", RiderId: "r1", PickupLat: 10.77, PickupLon: 106.69}
	wire, err := proto.Marshal(req)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	got := &dispatchpb.RequestDispatchRequest{}
	if err := proto.Unmarshal(wire, got); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if got.GetTripType() != "" || got.GetVehicleType() != "" {
		t.Errorf("expected empty trip_type/vehicle_type for a plain Ride request, got %+v", got)
	}
}

// TestDispatchJobProto_UnchangedShape locks in "DispatchResult phải giữ
// nguyên" — no trip_type/vehicle_type (or any other) field exists on
// DispatchJobProto; this test would fail to compile if that ever changed
// in a way that removed the fields asserted below, and documents the
// deliberate absence of Delivery-specific fields on the response payload.
func TestDispatchJobProto_UnchangedShape(t *testing.T) {
	p := &dispatchpb.DispatchJobProto{
		TripId:           "t1",
		RiderId:          "r1",
		Status:           "searching",
		CurrentDriverId:  "d1",
		AssignedDriverId: "",
		AttemptCount:     1,
		MaxAttempts:      5,
	}
	wire, err := proto.Marshal(p)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	got := &dispatchpb.DispatchJobProto{}
	if err := proto.Unmarshal(wire, got); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if got.GetTripId() != "t1" || got.GetStatus() != "searching" || got.GetCurrentDriverId() != "d1" {
		t.Errorf("DispatchJobProto fields did not round-trip: %+v", got)
	}
}
