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
	// 5 km, 15 min (rates: base=15000, 6500/km, 400/min — calibrated to
	// undercut Be/Xanh SM by ~10-20%, see DefaultFareConfig doc comment)
	// base=15000, distance=5*6500=32500, time=15*400=6000 → ride=53500 (>30000 min) → total=55500
	calc := defaultCalc()
	fb, err := calc.Estimate(entity.VehicleTypeCar, 5.0, 15.0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fb.BaseFare != 15000 {
		t.Errorf("BaseFare: got %d, want 15000", fb.BaseFare)
	}
	if fb.DistanceFare != 32500 {
		t.Errorf("DistanceFare: got %d, want 32500", fb.DistanceFare)
	}
	if fb.TimeFare != 6000 {
		t.Errorf("TimeFare: got %d, want 6000", fb.TimeFare)
	}
	if fb.RideFare != 53500 {
		t.Errorf("RideFare: got %d, want 53500", fb.RideFare)
	}
	if fb.BookingFee != 2000 {
		t.Errorf("BookingFee: got %d, want 2000", fb.BookingFee)
	}
	if fb.Total != 55500 {
		t.Errorf("Total: got %d, want 55500", fb.Total)
	}
	if fb.IsFinal {
		t.Error("IsFinal should be false for Estimate")
	}
}

func TestEstimate_Motorcycle_BasicTrip(t *testing.T) {
	// 3 km, 10 min (rates: base=8000, 2800/km, 200/min, min=15000)
	// base=8000, distance=3*2800=8400, time=10*200=2000 → ride=18400 (> 15000 min, no enforcement) → total=20400
	calc := defaultCalc()
	fb, err := calc.Estimate(entity.VehicleTypeMotorcycle, 3.0, 10.0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fb.RideFare != 18400 {
		t.Errorf("RideFare: got %d, want 18400", fb.RideFare)
	}
	if fb.Total != 20400 {
		t.Errorf("Total: got %d, want 20400", fb.Total)
	}
}

func TestEstimate_Van_BasicTrip(t *testing.T) {
	// 10 km, 20 min (BRB XL rates: base=18000, 5000/km, 500/min)
	// base=18000, distance=10*5000=50000, time=20*500=10000 → ride=78000 (>40000 min) → total=80000
	calc := defaultCalc()
	fb, err := calc.Estimate(entity.VehicleTypeVan, 10.0, 20.0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fb.RideFare != 78000 {
		t.Errorf("RideFare: got %d, want 78000", fb.RideFare)
	}
	if fb.Total != 80000 {
		t.Errorf("Total: got %d, want 80000", fb.Total)
	}
}

// ─── Minimum fare enforcement ─────────────────────────────────────────────────

func TestEstimate_Car_MinimumFareEnforced(t *testing.T) {
	// 0 km, 0 min → base=15000 → ride=15000 (< 30000 min) → ride=30000
	calc := defaultCalc()
	fb, err := calc.Estimate(entity.VehicleTypeCar, 0, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fb.RideFare != 30000 {
		t.Errorf("RideFare: got %d, want 30000 (minimum fare)", fb.RideFare)
	}
	if fb.Total != 32000 {
		t.Errorf("Total: got %d, want 32000 (minimum + booking fee)", fb.Total)
	}
}

func TestEstimate_Motorcycle_MinimumFareNotOverApplied(t *testing.T) {
	// 10 km, 30 min → base=8000, distance=28000, time=6000 → ride=42000 (> 15000 min) → no minimum
	calc := defaultCalc()
	fb, err := calc.Estimate(entity.VehicleTypeMotorcycle, 10.0, 30.0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fb.RideFare != 42000 {
		t.Errorf("RideFare: got %d, want 42000", fb.RideFare)
	}
}

// ─── Distance fare component ──────────────────────────────────────────────────

func TestEstimate_DistanceFare_Rounding(t *testing.T) {
	// Car: 1.333... km * 6500/km = 8666.67... → rounds to 8667
	calc := defaultCalc()
	fb, err := calc.Estimate(entity.VehicleTypeCar, 1.0/3.0*4, 0) // 1.333... km
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fb.DistanceFare != 8667 {
		t.Errorf("DistanceFare: got %d, want 8667", fb.DistanceFare)
	}
}

func TestEstimate_DistanceFareComponent_Isolated(t *testing.T) {
	// 2 km, 0 min
	// Car: base=15000, distance=2*6500=13000, time=0 → ride=28000 (< 30000) → ride=30000
	calc := defaultCalc()
	fb, err := calc.Estimate(entity.VehicleTypeCar, 2.0, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fb.DistanceFare != 13000 {
		t.Errorf("DistanceFare: got %d, want 13000", fb.DistanceFare)
	}
}

// ─── Time fare component ──────────────────────────────────────────────────────

func TestEstimate_TimeFareComponent_Isolated(t *testing.T) {
	// 0 km, 10 min
	// Car: base=10000, distance=0, time=10*400=4000 → ride=14000 (< 25000) → ride=25000
	calc := defaultCalc()
	fb, err := calc.Estimate(entity.VehicleTypeCar, 0, 10.0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fb.TimeFare != 4000 {
		t.Errorf("TimeFare: got %d, want 4000", fb.TimeFare)
	}
}

func TestEstimate_TimeFare_Rounding(t *testing.T) {
	// Car: 1.667 min → 1.667*400 = 666.67 → rounds to 667
	calc := defaultCalc()
	fb, err := calc.Estimate(entity.VehicleTypeCar, 0, 5.0/3.0) // 1.6667 min
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fb.TimeFare != 667 {
		t.Errorf("TimeFare: got %d, want 667", fb.TimeFare)
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
	if fb.CurrencyCode != "VND" {
		t.Errorf("CurrencyCode: got %q, want VND", fb.CurrencyCode)
	}
}
