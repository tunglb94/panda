package bookingpb_test

// Smoke test proving the hand-mutated FileDescriptorProto (Delivery V1 Phase 2,
// docs/business/DELIVERY_V1_DESIGN.md) actually round-trips correctly over
// the real reflection-based proto.Marshal/Unmarshal path — not just that the
// Go struct compiles. If the rawDesc bytes were inconsistent with the struct
// tags, this would panic at package init (protoimpl.TypeBuilder.Build) or
// silently drop the new fields during Marshal/Unmarshal.

import (
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/fairride/booking/grpc/bookingpb"
)

func TestBookRideRequest_DeliveryFieldsRoundTripOverWire(t *testing.T) {
	req := &bookingpb.BookRideRequest{
		RiderId:            "r1",
		PickupAddress:      "123 Main St",
		DropoffAddress:     "456 Elm Ave",
		PickupLat:          10.77,
		PickupLon:          106.69,
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

	got := &bookingpb.BookRideRequest{}
	if err := proto.Unmarshal(wire, got); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if got.GetTripType() != "delivery" {
		t.Errorf("TripType = %q, want %q", got.GetTripType(), "delivery")
	}
	if got.GetPickupContactName() != "Nguyen Van A" {
		t.Errorf("PickupContactName = %q", got.GetPickupContactName())
	}
	if got.GetPickupContactPhone() != "0912345678" {
		t.Errorf("PickupContactPhone = %q", got.GetPickupContactPhone())
	}
	if got.GetReceiverName() != "Tran Thi B" {
		t.Errorf("ReceiverName = %q", got.GetReceiverName())
	}
	if got.GetReceiverPhone() != "0987654321" {
		t.Errorf("ReceiverPhone = %q", got.GetReceiverPhone())
	}
	if got.GetPackageNote() != "handle with care" {
		t.Errorf("PackageNote = %q", got.GetPackageNote())
	}
	if got.GetPackageValue() != 500000 {
		t.Errorf("PackageValue = %d, want 500000", got.GetPackageValue())
	}
	if got.GetPackageWeight() != 2.5 {
		t.Errorf("PackageWeight = %v, want 2.5", got.GetPackageWeight())
	}
	// Pre-existing Ride fields must still round-trip unchanged.
	if got.GetRiderId() != "r1" || got.GetPickupAddress() != "123 Main St" || got.GetDropoffAddress() != "456 Elm Ave" {
		t.Errorf("pre-existing Ride fields did not round-trip correctly: %+v", got)
	}
}

func TestBookRideRequest_RideOnly_OmittedDeliveryFieldsStayEmpty(t *testing.T) {
	// A Ride booking that never sets trip_type or any delivery field must
	// round-trip with those fields at their proto3 zero value — proving
	// backward compatibility for every existing Ride caller.
	req := &bookingpb.BookRideRequest{
		RiderId:        "r1",
		PickupAddress:  "123 Main St",
		DropoffAddress: "456 Elm Ave",
		PickupLat:      10.77,
		PickupLon:      106.69,
	}
	wire, err := proto.Marshal(req)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	got := &bookingpb.BookRideRequest{}
	if err := proto.Unmarshal(wire, got); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if got.GetTripType() != "" {
		t.Errorf("TripType = %q, want empty for a plain Ride request", got.GetTripType())
	}
	if got.GetReceiverName() != "" || got.GetReceiverPhone() != "" {
		t.Errorf("delivery fields should be empty for a Ride request, got %+v", got)
	}
}

func TestBookRideResponse_DeliveryIdRoundTrips(t *testing.T) {
	resp := &bookingpb.BookRideResponse{
		TripId:     "t1",
		Status:     "searching",
		DeliveryId: "d1",
	}
	wire, err := proto.Marshal(resp)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	got := &bookingpb.BookRideResponse{}
	if err := proto.Unmarshal(wire, got); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if got.GetDeliveryId() != "d1" {
		t.Errorf("DeliveryId = %q, want %q", got.GetDeliveryId(), "d1")
	}
}

func TestBookRideResponse_RideOnly_DeliveryIdStaysEmpty(t *testing.T) {
	resp := &bookingpb.BookRideResponse{TripId: "t1", Status: "searching"}
	wire, err := proto.Marshal(resp)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	got := &bookingpb.BookRideResponse{}
	if err := proto.Unmarshal(wire, got); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if got.GetDeliveryId() != "" {
		t.Errorf("DeliveryId = %q, want empty for a Ride response", got.GetDeliveryId())
	}
}
