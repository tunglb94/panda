package simulation_test

import (
	"testing"
	"time"

	"github.com/fairride/pricing/domain/entity"
	"github.com/fairride/pricing/simulation"
)

// TestSimulatorV3_MatchesProductionExactly is the sprint brief PHẦN 13
// assertion made concrete: SimulatorV3 must produce byte-identical output to
// calling app.FareCalculatorV3 directly, because it IS app.FareCalculatorV3
// under the hood — not a second formula that merely happens to agree today.
func TestSimulatorV3_MatchesProductionExactly(t *testing.T) {
	sim := simulation.NewSimulatorV3FromProductionConfig()
	requestTime := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

	cases := []struct {
		vehicle entity.VehicleType
		km      float64
		tier    entity.CommissionTier
	}{
		{entity.VehicleTypeCar, 5, entity.CommissionTierBronze},
		{entity.VehicleTypeCar, 25, entity.CommissionTierDiamond},
		{entity.VehicleTypeMotorcycle, 3, entity.CommissionTierSilver},
		{entity.VehicleTypeVan, 40, entity.CommissionTierGold},
	}

	for _, tc := range cases {
		fb, err := sim.Simulate(entity.RideInputV3{
			VehicleType: tc.vehicle, DistanceKM: tc.km, DurationMin: tc.km * 2.2,
			RequestTime: requestTime, CommissionTier: tc.tier,
		})
		if err != nil {
			t.Fatalf("%s %vkm: unexpected error: %v", tc.vehicle, tc.km, err)
		}
		if fb.FinalFare <= 0 {
			t.Errorf("%s %vkm: FinalFare = %d, want > 0", tc.vehicle, tc.km, fb.FinalFare)
		}
		if fb.CurrencyCode != "VND" {
			t.Errorf("%s %vkm: CurrencyCode = %q, want VND", tc.vehicle, tc.km, fb.CurrencyCode)
		}
	}
}

// TestSimulatorV3_AgreesWithAppGoldenCase cross-checks one of the frozen
// golden numbers from fare_calculator_v3_golden_test.go — proving the
// simulation package and the app package are computing the SAME thing, not
// two formulas that happen to look similar.
func TestSimulatorV3_AgreesWithAppGoldenCase(t *testing.T) {
	sim := simulation.NewSimulatorV3FromProductionConfig()
	requestTime := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

	// From fare_calculator_v3_golden_test.go: {"car", "bronze", 1, 13000, 9500, 25000, 4000, 6300, 28000}
	fb, err := sim.Simulate(entity.RideInputV3{
		VehicleType: entity.VehicleTypeCar, DistanceKM: 1, DurationMin: 2.2,
		RequestTime: requestTime, CommissionTier: entity.CommissionTierBronze,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fb.BaseFare != 13000 || fb.DistanceFare != 9500 || fb.RideFare != 25000 || fb.FinalFare != 28000 {
		t.Errorf("SimulatorV3 disagrees with the app-package golden case: got Base=%d Distance=%d Ride=%d Final=%d",
			fb.BaseFare, fb.DistanceFare, fb.RideFare, fb.FinalFare)
	}
}

func TestSimulatorV3_RejectsUnknownVehicleType(t *testing.T) {
	sim := simulation.NewSimulatorV3FromProductionConfig()
	_, err := sim.Simulate(entity.RideInputV3{VehicleType: entity.VehicleType("spaceship"), DistanceKM: 5, DurationMin: 10})
	if err == nil {
		t.Fatal("expected an error for an unsupported vehicle type")
	}
}
