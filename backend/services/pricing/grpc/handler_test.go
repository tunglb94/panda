package grpc_test

import (
	"context"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/fairride/pricing/app"
	"github.com/fairride/pricing/domain/entity"
	pricinggrpc "github.com/fairride/pricing/grpc"
	"github.com/fairride/pricing/grpc/pricingpb"
)

func newHandler() *pricinggrpc.Handler {
	return pricinggrpc.NewHandler(app.NewFareCalculator(entity.DefaultFareConfig()))
}

// ─── EstimateFare ─────────────────────────────────────────────────────────────

func TestEstimateFare_ValidCar(t *testing.T) {
	h := newHandler()
	resp, err := h.EstimateFare(context.Background(), &pricingpb.EstimateFareRequest{
		VehicleType:     "car",
		DistanceKm:      5.0,
		DurationMinutes: 15.0,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.GetFare() == nil {
		t.Fatal("expected non-nil fare")
	}
	if resp.GetFare().GetTotal() != 55500 {
		t.Errorf("Total: got %d, want 55500", resp.GetFare().GetTotal())
	}
	if resp.GetFare().GetIsFinal() {
		t.Error("IsFinal should be false for EstimateFare")
	}
}

func TestEstimateFare_MissingVehicleType(t *testing.T) {
	h := newHandler()
	_, err := h.EstimateFare(context.Background(), &pricingpb.EstimateFareRequest{
		DistanceKm:      5.0,
		DurationMinutes: 15.0,
	})
	if err == nil {
		t.Fatal("expected error for missing vehicle_type")
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected gRPC status error, got: %v", err)
	}
	if st.Code() != codes.InvalidArgument {
		t.Errorf("code: got %v, want InvalidArgument", st.Code())
	}
}

func TestEstimateFare_UnknownVehicleType(t *testing.T) {
	h := newHandler()
	_, err := h.EstimateFare(context.Background(), &pricingpb.EstimateFareRequest{
		VehicleType:     "spaceship",
		DistanceKm:      5.0,
		DurationMinutes: 15.0,
	})
	if err == nil {
		t.Fatal("expected error for unknown vehicle type")
	}
	st, _ := status.FromError(err)
	if st.Code() != codes.InvalidArgument {
		t.Errorf("code: got %v, want InvalidArgument", st.Code())
	}
}

func TestEstimateFare_NegativeDistance(t *testing.T) {
	h := newHandler()
	_, err := h.EstimateFare(context.Background(), &pricingpb.EstimateFareRequest{
		VehicleType:     "car",
		DistanceKm:      -1.0,
		DurationMinutes: 10.0,
	})
	if err == nil {
		t.Fatal("expected error for negative distance")
	}
	st, _ := status.FromError(err)
	if st.Code() != codes.InvalidArgument {
		t.Errorf("code: got %v, want InvalidArgument", st.Code())
	}
}

func TestEstimateFare_MinimumFareReturned(t *testing.T) {
	h := newHandler()
	resp, err := h.EstimateFare(context.Background(), &pricingpb.EstimateFareRequest{
		VehicleType:     "car",
		DistanceKm:      0,
		DurationMinutes: 0,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// ride_fare should be minimum (30000) + booking_fee (2000) = 32000
	if resp.GetFare().GetTotal() != 32000 {
		t.Errorf("Total: got %d, want 32000", resp.GetFare().GetTotal())
	}
}

// ─── CalculateFinalFare ───────────────────────────────────────────────────────

func TestCalculateFinalFare_ValidCar(t *testing.T) {
	h := newHandler()
	resp, err := h.CalculateFinalFare(context.Background(), &pricingpb.CalculateFinalFareRequest{
		VehicleType:           "car",
		ActualDistanceKm:      5.0,
		ActualDurationMinutes: 15.0,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.GetFare().GetIsFinal() {
		t.Error("IsFinal should be true for CalculateFinalFare")
	}
	if resp.GetFare().GetTotal() != 55500 {
		t.Errorf("Total: got %d, want 55500", resp.GetFare().GetTotal())
	}
}

func TestCalculateFinalFare_MissingVehicleType(t *testing.T) {
	h := newHandler()
	_, err := h.CalculateFinalFare(context.Background(), &pricingpb.CalculateFinalFareRequest{
		ActualDistanceKm:      5.0,
		ActualDurationMinutes: 15.0,
	})
	if err == nil {
		t.Fatal("expected error for missing vehicle_type")
	}
	st, _ := status.FromError(err)
	if st.Code() != codes.InvalidArgument {
		t.Errorf("code: got %v, want InvalidArgument", st.Code())
	}
}

func TestCalculateFinalFare_AllVehicleTypes(t *testing.T) {
	h := newHandler()
	for _, vt := range []string{"car", "motorcycle", "van"} {
		resp, err := h.CalculateFinalFare(context.Background(), &pricingpb.CalculateFinalFareRequest{
			VehicleType:           vt,
			ActualDistanceKm:      10.0,
			ActualDurationMinutes: 20.0,
		})
		if err != nil {
			t.Errorf("%s: unexpected error: %v", vt, err)
			continue
		}
		if resp.GetFare().GetTotal() <= 0 {
			t.Errorf("%s: expected positive total, got %d", vt, resp.GetFare().GetTotal())
		}
		if resp.GetFare().GetVehicleType() != vt {
			t.Errorf("%s: VehicleType mismatch: got %q", vt, resp.GetFare().GetVehicleType())
		}
	}
}

func TestCalculateFinalFare_ProtoBreakdownComplete(t *testing.T) {
	h := newHandler()
	resp, err := h.CalculateFinalFare(context.Background(), &pricingpb.CalculateFinalFareRequest{
		VehicleType:           "car",
		ActualDistanceKm:      5.0,
		ActualDurationMinutes: 15.0,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	fb := resp.GetFare()
	if fb.GetCurrencyCode() != "VND" {
		t.Errorf("CurrencyCode: got %q, want VND", fb.GetCurrencyCode())
	}
	if fb.GetBaseFare() == 0 {
		t.Error("BaseFare should be non-zero")
	}
	if fb.GetBookingFee() == 0 {
		t.Error("BookingFee should be non-zero")
	}
	if fb.GetRideFare()+fb.GetBookingFee() != fb.GetTotal() {
		t.Errorf("Total invariant broken: RideFare(%d)+BookingFee(%d)=%d, Total=%d",
			fb.GetRideFare(), fb.GetBookingFee(),
			fb.GetRideFare()+fb.GetBookingFee(), fb.GetTotal())
	}
}
