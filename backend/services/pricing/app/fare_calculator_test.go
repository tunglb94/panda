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
	// 5 km, 15 min (rates: base=11250, 10700/km, 450/min — calibrated to
	// ±5% of the real Grab/Be/Xanh SM average, see DefaultFareConfig doc comment)
	// base=11250, distance=5*10700=53500, time=15*450=6750 → ride=71500 (>25000 min) → total=73500
	calc := defaultCalc()
	fb, err := calc.Estimate(entity.VehicleTypeCar, 5.0, 15.0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fb.BaseFare != 11250 {
		t.Errorf("BaseFare: got %d, want 11250", fb.BaseFare)
	}
	if fb.DistanceFare != 53500 {
		t.Errorf("DistanceFare: got %d, want 53500", fb.DistanceFare)
	}
	if fb.TimeFare != 6750 {
		t.Errorf("TimeFare: got %d, want 6750", fb.TimeFare)
	}
	if fb.RideFare != 71500 {
		t.Errorf("RideFare: got %d, want 71500", fb.RideFare)
	}
	if fb.BookingFee != 2000 {
		t.Errorf("BookingFee: got %d, want 2000", fb.BookingFee)
	}
	if fb.Total != 73500 {
		t.Errorf("Total: got %d, want 73500", fb.Total)
	}
	if fb.IsFinal {
		t.Error("IsFinal should be false for Estimate")
	}
}

func TestEstimate_Motorcycle_BasicTrip(t *testing.T) {
	// 3 km, 10 min (rates: base=2700, 4140/km, 200/min, min=9000)
	// base=2700, distance=3*4140=12420, time=10*200=2000 → ride=17120 (> 9000 min, no enforcement) → total=19120
	calc := defaultCalc()
	fb, err := calc.Estimate(entity.VehicleTypeMotorcycle, 3.0, 10.0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fb.RideFare != 17120 {
		t.Errorf("RideFare: got %d, want 17120", fb.RideFare)
	}
	if fb.Total != 19120 {
		t.Errorf("Total: got %d, want 19120", fb.Total)
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
	// 0 km, 0 min → base=11250 → ride=11250 (< 25000 min) → ride=25000
	calc := defaultCalc()
	fb, err := calc.Estimate(entity.VehicleTypeCar, 0, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fb.RideFare != 25000 {
		t.Errorf("RideFare: got %d, want 25000 (minimum fare)", fb.RideFare)
	}
	if fb.Total != 27000 {
		t.Errorf("Total: got %d, want 27000 (minimum + booking fee)", fb.Total)
	}
}

func TestEstimate_Motorcycle_MinimumFareNotOverApplied(t *testing.T) {
	// 10 km, 30 min → base=2700, distance=41400, time=6000 → ride=50100 (> 9000 min) → no minimum
	calc := defaultCalc()
	fb, err := calc.Estimate(entity.VehicleTypeMotorcycle, 10.0, 30.0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fb.RideFare != 50100 {
		t.Errorf("RideFare: got %d, want 50100", fb.RideFare)
	}
}

// ─── Distance fare component ──────────────────────────────────────────────────

func TestEstimate_DistanceFare_Rounding(t *testing.T) {
	// Car: 1.333... km * 10700/km = 14266.67... → rounds to 14267
	calc := defaultCalc()
	fb, err := calc.Estimate(entity.VehicleTypeCar, 1.0/3.0*4, 0) // 1.333... km
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fb.DistanceFare != 14267 {
		t.Errorf("DistanceFare: got %d, want 14267", fb.DistanceFare)
	}
}

func TestEstimate_DistanceFareComponent_Isolated(t *testing.T) {
	// 2 km, 0 min
	// Car: base=11250, distance=2*10700=21400, time=0 → ride=32650 (> 25000, no floor)
	calc := defaultCalc()
	fb, err := calc.Estimate(entity.VehicleTypeCar, 2.0, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fb.DistanceFare != 21400 {
		t.Errorf("DistanceFare: got %d, want 21400", fb.DistanceFare)
	}
}

// ─── Time fare component ──────────────────────────────────────────────────────

func TestEstimate_TimeFareComponent_Isolated(t *testing.T) {
	// 0 km, 10 min
	// Car: base=11250, distance=0, time=10*450=4500 → ride=15750 (< 25000) → ride=25000
	calc := defaultCalc()
	fb, err := calc.Estimate(entity.VehicleTypeCar, 0, 10.0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fb.TimeFare != 4500 {
		t.Errorf("TimeFare: got %d, want 4500", fb.TimeFare)
	}
}

func TestEstimate_TimeFare_Rounding(t *testing.T) {
	// Car: 1.6667 min * 450/min = 750.0 exactly
	calc := defaultCalc()
	fb, err := calc.Estimate(entity.VehicleTypeCar, 0, 5.0/3.0) // 1.6667 min
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fb.TimeFare != 750 {
		t.Errorf("TimeFare: got %d, want 750", fb.TimeFare)
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
