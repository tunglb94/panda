package app_test

import (
	"testing"

	"github.com/fairride/pricing/app"
	"github.com/fairride/pricing/domain/entity"
	domainerrors "github.com/fairride/shared/errors"
)

func defaultCalc() *app.FareCalculator {
	return app.NewFareCalculator(entity.DefaultFareConfig())
}

// ─── Estimate — valid inputs ──────────────────────────────────────────────────

func TestEstimate_Car_BasicTrip(t *testing.T) {
	// 5 km, 15 min
	// base=50, distance=5*30=150, time=15*5=75 → ride=275 (>200 min) → total=325
	calc := defaultCalc()
	fb, err := calc.Estimate(entity.VehicleTypeCar, 5.0, 15.0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fb.BaseFare != 50 {
		t.Errorf("BaseFare: got %d, want 50", fb.BaseFare)
	}
	if fb.DistanceFare != 150 {
		t.Errorf("DistanceFare: got %d, want 150", fb.DistanceFare)
	}
	if fb.TimeFare != 75 {
		t.Errorf("TimeFare: got %d, want 75", fb.TimeFare)
	}
	if fb.RideFare != 275 {
		t.Errorf("RideFare: got %d, want 275", fb.RideFare)
	}
	if fb.BookingFee != 50 {
		t.Errorf("BookingFee: got %d, want 50", fb.BookingFee)
	}
	if fb.Total != 325 {
		t.Errorf("Total: got %d, want 325", fb.Total)
	}
	if fb.IsFinal {
		t.Error("IsFinal should be false for Estimate")
	}
}

func TestEstimate_Motorcycle_BasicTrip(t *testing.T) {
	// 3 km, 10 min
	// base=30, distance=3*20=60, time=10*3=30 → ride=120 (< 150 min) → ride=150 → total=180
	calc := defaultCalc()
	fb, err := calc.Estimate(entity.VehicleTypeMotorcycle, 3.0, 10.0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fb.RideFare != 150 {
		t.Errorf("RideFare: got %d, want 150 (minimum fare enforced)", fb.RideFare)
	}
	if fb.Total != 180 {
		t.Errorf("Total: got %d, want 180", fb.Total)
	}
}

func TestEstimate_Van_BasicTrip(t *testing.T) {
	// 10 km, 20 min
	// base=100, distance=10*50=500, time=20*8=160 → ride=760 (>300 min) → total=835
	calc := defaultCalc()
	fb, err := calc.Estimate(entity.VehicleTypeVan, 10.0, 20.0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fb.RideFare != 760 {
		t.Errorf("RideFare: got %d, want 760", fb.RideFare)
	}
	if fb.Total != 835 {
		t.Errorf("Total: got %d, want 835", fb.Total)
	}
}

// ─── Minimum fare enforcement ─────────────────────────────────────────────────

func TestEstimate_Car_MinimumFareEnforced(t *testing.T) {
	// 0 km, 0 min → base=50 → ride=50 (< 200 min) → ride=200
	calc := defaultCalc()
	fb, err := calc.Estimate(entity.VehicleTypeCar, 0, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fb.RideFare != 200 {
		t.Errorf("RideFare: got %d, want 200 (minimum fare)", fb.RideFare)
	}
	if fb.Total != 250 {
		t.Errorf("Total: got %d, want 250 (minimum + booking fee)", fb.Total)
	}
}

func TestEstimate_Motorcycle_MinimumFareNotOverApplied(t *testing.T) {
	// 10 km, 30 min → base=30, distance=200, time=90 → ride=320 (> 150 min) → no minimum
	calc := defaultCalc()
	fb, err := calc.Estimate(entity.VehicleTypeMotorcycle, 10.0, 30.0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fb.RideFare != 320 {
		t.Errorf("RideFare: got %d, want 320", fb.RideFare)
	}
}

// ─── Distance fare component ──────────────────────────────────────────────────

func TestEstimate_DistanceFare_Rounding(t *testing.T) {
	// Car: 1.5 km → 1.5*30 = 45 (exact, no rounding needed)
	// 1.3 km → 1.3*30 = 39 (exact)
	// 1.333... km → 1.333*30 = 39.99 → rounds to 40
	calc := defaultCalc()
	fb, err := calc.Estimate(entity.VehicleTypeCar, 1.0/3.0*4, 0) // 1.333... km
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 1.333...*30 = 40.0 → 40
	if fb.DistanceFare != 40 {
		t.Errorf("DistanceFare: got %d, want 40", fb.DistanceFare)
	}
}

func TestEstimate_DistanceFareComponent_Isolated(t *testing.T) {
	// 2 km, 0 min
	// Car: base=50, distance=60, time=0 → ride=110 (< 200) → ride=200
	calc := defaultCalc()
	fb, err := calc.Estimate(entity.VehicleTypeCar, 2.0, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fb.DistanceFare != 60 {
		t.Errorf("DistanceFare: got %d, want 60", fb.DistanceFare)
	}
}

// ─── Time fare component ──────────────────────────────────────────────────────

func TestEstimate_TimeFareComponent_Isolated(t *testing.T) {
	// 0 km, 10 min
	// Car: base=50, distance=0, time=50 → ride=100 (< 200) → ride=200
	calc := defaultCalc()
	fb, err := calc.Estimate(entity.VehicleTypeCar, 0, 10.0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fb.TimeFare != 50 {
		t.Errorf("TimeFare: got %d, want 50", fb.TimeFare)
	}
}

func TestEstimate_TimeFare_Rounding(t *testing.T) {
	// Car: 1.667 min → 1.667*5 = 8.333 → rounds to 8
	calc := defaultCalc()
	fb, err := calc.Estimate(entity.VehicleTypeCar, 0, 5.0/3.0) // 1.6667 min
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fb.TimeFare != 8 {
		t.Errorf("TimeFare: got %d, want 8", fb.TimeFare)
	}
}

// ─── Booking fee ──────────────────────────────────────────────────────────────

func TestEstimate_BookingFeeAlwaysIncluded(t *testing.T) {
	calc := defaultCalc()
	for _, vt := range []entity.VehicleType{entity.VehicleTypeCar, entity.VehicleTypeMotorcycle, entity.VehicleTypeVan} {
		fb, err := calc.Estimate(vt, 0, 0)
		if err != nil {
			t.Fatalf("%s: unexpected error: %v", vt, err)
		}
		cfg := entity.DefaultFareConfig()
		expectedFee := cfg.Rates[vt].BookingFee
		if fb.BookingFee != expectedFee {
			t.Errorf("%s: BookingFee got %d, want %d", vt, fb.BookingFee, expectedFee)
		}
		if fb.Total != fb.RideFare+fb.BookingFee {
			t.Errorf("%s: Total (%d) != RideFare (%d) + BookingFee (%d)", vt, fb.Total, fb.RideFare, fb.BookingFee)
		}
	}
}

// ─── IsFinal flag ─────────────────────────────────────────────────────────────

func TestCalculateFinal_IsFinalTrue(t *testing.T) {
	calc := defaultCalc()
	fb, err := calc.CalculateFinal(entity.VehicleTypeCar, 5.0, 15.0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !fb.IsFinal {
		t.Error("IsFinal should be true for CalculateFinal")
	}
}

func TestEstimate_IsFinalFalse(t *testing.T) {
	calc := defaultCalc()
	fb, err := calc.Estimate(entity.VehicleTypeCar, 5.0, 15.0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fb.IsFinal {
		t.Error("IsFinal should be false for Estimate")
	}
}

func TestEstimateAndFinal_SameFormula(t *testing.T) {
	// Upfront pricing guarantee: same formula, only IsFinal differs.
	calc := defaultCalc()
	est, err := calc.Estimate(entity.VehicleTypeCar, 7.5, 22.0)
	if err != nil {
		t.Fatalf("Estimate error: %v", err)
	}
	fin, err := calc.CalculateFinal(entity.VehicleTypeCar, 7.5, 22.0)
	if err != nil {
		t.Fatalf("CalculateFinal error: %v", err)
	}
	if est.Total != fin.Total {
		t.Errorf("Total mismatch: Estimate=%d, Final=%d", est.Total, fin.Total)
	}
	if est.RideFare != fin.RideFare {
		t.Errorf("RideFare mismatch: Estimate=%d, Final=%d", est.RideFare, fin.RideFare)
	}
}

// ─── Error cases ──────────────────────────────────────────────────────────────

func TestEstimate_UnknownVehicleType(t *testing.T) {
	calc := defaultCalc()
	_, err := calc.Estimate("bicycle", 5.0, 15.0)
	if err == nil {
		t.Fatal("expected error for unknown vehicle type")
	}
	if !domainerrors.IsCode(err, domainerrors.CodeInvalidArgument) {
		t.Errorf("expected CodeInvalidArgument, got: %v", err)
	}
}

func TestEstimate_NegativeDistance(t *testing.T) {
	calc := defaultCalc()
	_, err := calc.Estimate(entity.VehicleTypeCar, -1.0, 10.0)
	if err == nil {
		t.Fatal("expected error for negative distance")
	}
	if !domainerrors.IsCode(err, domainerrors.CodeInvalidArgument) {
		t.Errorf("expected CodeInvalidArgument, got: %v", err)
	}
}

func TestEstimate_NegativeDuration(t *testing.T) {
	calc := defaultCalc()
	_, err := calc.Estimate(entity.VehicleTypeCar, 5.0, -1.0)
	if err == nil {
		t.Fatal("expected error for negative duration")
	}
	if !domainerrors.IsCode(err, domainerrors.CodeInvalidArgument) {
		t.Errorf("expected CodeInvalidArgument, got: %v", err)
	}
}

func TestCalculateFinal_NegativeDistance(t *testing.T) {
	calc := defaultCalc()
	_, err := calc.CalculateFinal(entity.VehicleTypeCar, -0.1, 5.0)
	if err == nil {
		t.Fatal("expected error for negative distance")
	}
	if !domainerrors.IsCode(err, domainerrors.CodeInvalidArgument) {
		t.Errorf("expected CodeInvalidArgument, got: %v", err)
	}
}

func TestCalculateFinal_UnknownVehicleType(t *testing.T) {
	calc := defaultCalc()
	_, err := calc.CalculateFinal("tuk-tuk", 5.0, 15.0)
	if err == nil {
		t.Fatal("expected error for unknown vehicle type")
	}
	if !domainerrors.IsCode(err, domainerrors.CodeInvalidArgument) {
		t.Errorf("expected CodeInvalidArgument, got: %v", err)
	}
}

// ─── Zero distance + zero duration ───────────────────────────────────────────

func TestEstimate_ZeroZero_MinimumFare(t *testing.T) {
	calc := defaultCalc()
	cfg := entity.DefaultFareConfig()
	for _, vt := range []entity.VehicleType{entity.VehicleTypeCar, entity.VehicleTypeMotorcycle, entity.VehicleTypeVan} {
		fb, err := calc.Estimate(vt, 0, 0)
		if err != nil {
			t.Fatalf("%s: unexpected error: %v", vt, err)
		}
		minFare := cfg.Rates[vt].MinimumFare
		if fb.RideFare != minFare {
			t.Errorf("%s: 0km/0min RideFare=%d, want minimum %d", vt, fb.RideFare, minFare)
		}
	}
}

// ─── FareBreakdown fields populated correctly ─────────────────────────────────

func TestEstimate_BreakdownFields(t *testing.T) {
	calc := defaultCalc()
	fb, err := calc.Estimate(entity.VehicleTypeCar, 5.0, 15.0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fb.VehicleType != entity.VehicleTypeCar {
		t.Errorf("VehicleType: got %q, want %q", fb.VehicleType, entity.VehicleTypeCar)
	}
	if fb.DistanceKM != 5.0 {
		t.Errorf("DistanceKM: got %f, want 5.0", fb.DistanceKM)
	}
	if fb.DurationMin != 15.0 {
		t.Errorf("DurationMin: got %f, want 15.0", fb.DurationMin)
	}
	if fb.CurrencyCode != "USD" {
		t.Errorf("CurrencyCode: got %q, want USD", fb.CurrencyCode)
	}
}
