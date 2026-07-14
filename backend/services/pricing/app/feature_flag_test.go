package app_test

import (
	"testing"

	"github.com/fairride/pricing/app"
	"github.com/fairride/pricing/config"
	"github.com/fairride/pricing/domain/entity"
)

// ─── PHẦN 12 / 17: Backward Compatibility + Feature Flag ────────────────────

func TestVersionedFareCalculator_DefaultsToV2(t *testing.T) {
	v2 := app.NewFareCalculator(entity.DefaultFareConfig())
	cfg := config.Default()
	v3 := app.NewFareCalculatorV3(cfg.Fare, cfg.Airport, cfg.Commission, cfg.VATRate, app.DefaultRuleConfigs())

	// Empty version string — the exact scenario an unconfigured
	// PRICING_VERSION env var produces — must fail closed to v2, per the
	// sprint's "Không bật Pricing V3 mặc định."
	versioned := app.NewVersionedFareCalculator("", v2, v3)
	if versioned.Version != app.PricingVersionV2 {
		t.Errorf("Version = %q, want v2 as the default", versioned.Version)
	}
}

func TestVersionedFareCalculator_UnrecognisedVersionFailsClosedToV2(t *testing.T) {
	v2 := app.NewFareCalculator(entity.DefaultFareConfig())
	versioned := app.NewVersionedFareCalculator(app.PricingVersion("v99"), v2, nil)
	if versioned.Version != app.PricingVersionV2 {
		t.Errorf("Version = %q, want v2 for an unrecognised value", versioned.Version)
	}
}

func TestVersionedFareCalculator_V2ModeMatchesDirectFareCalculator(t *testing.T) {
	v2 := app.NewFareCalculator(entity.DefaultFareConfig())
	versioned := app.NewVersionedFareCalculator(app.PricingVersionV2, v2, nil)

	direct, err := v2.Estimate(entity.VehicleTypeCar, 12, 26)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	viaVersioned, err := versioned.Estimate(entity.VehicleTypeCar, 12, 26)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if direct.Total != viaVersioned.Total {
		t.Errorf("v2-mode Total via VersionedFareCalculator = %d, want %d (must be byte-identical to calling FareCalculator directly)", viaVersioned.Total, direct.Total)
	}
	if direct.CurrencyCode != viaVersioned.CurrencyCode {
		t.Errorf("CurrencyCode mismatch: %q vs %q", viaVersioned.CurrencyCode, direct.CurrencyCode)
	}
}

func TestVersionedFareCalculator_V3ModeReturnsV2Shape(t *testing.T) {
	v2 := app.NewFareCalculator(entity.DefaultFareConfig())
	cfg := config.Default()
	v3 := app.NewFareCalculatorV3(cfg.Fare, cfg.Airport, cfg.Commission, cfg.VATRate, app.DefaultRuleConfigs())
	versioned := app.NewVersionedFareCalculator(app.PricingVersionV3, v2, v3)

	fb, err := versioned.Estimate(entity.VehicleTypeCar, 10, 22)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fb.Total == 0 {
		t.Error("expected a non-zero Total in v3 mode")
	}
	// v3-mode Total must differ from what v2 alone would compute (different
	// rates/formula) — confirms the flag is actually switching engines, not
	// silently still calling v2.
	v2Only, err := v2.Estimate(entity.VehicleTypeCar, 10, 22)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fb.Total == v2Only.Total {
		t.Error("v3-mode Total unexpectedly matches v2's Total exactly — flag may not be switching engines")
	}
}

func TestVersionedFareCalculator_EstimateV3Detailed_RequiresV3Mode(t *testing.T) {
	v2 := app.NewFareCalculator(entity.DefaultFareConfig())
	cfg := config.Default()
	v3 := app.NewFareCalculatorV3(cfg.Fare, cfg.Airport, cfg.Commission, cfg.VATRate, app.DefaultRuleConfigs())
	versioned := app.NewVersionedFareCalculator(app.PricingVersionV2, v2, v3) // v2 mode

	_, err := versioned.EstimateV3Detailed(entity.RideInputV3{VehicleType: entity.VehicleTypeCar, DistanceKM: 5, DurationMin: 11})
	if err == nil {
		t.Fatal("expected an error requesting a V3 detailed breakdown while running in v2 mode")
	}
}

func TestVersionedFareCalculator_EstimateV3Detailed_WorksInV3Mode(t *testing.T) {
	v2 := app.NewFareCalculator(entity.DefaultFareConfig())
	cfg := config.Default()
	v3 := app.NewFareCalculatorV3(cfg.Fare, cfg.Airport, cfg.Commission, cfg.VATRate, app.DefaultRuleConfigs())
	versioned := app.NewVersionedFareCalculator(app.PricingVersionV3, v2, v3)

	full, err := versioned.EstimateV3Detailed(entity.RideInputV3{VehicleType: entity.VehicleTypeCar, DistanceKM: 5, DurationMin: 11, CommissionTier: entity.CommissionTierGold})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if full.Commission == 0 {
		t.Error("expected a non-zero Commission in the detailed V3 breakdown")
	}
}

// TestFareCalculatorV2_UnaffectedByV3Additions is the sprint's core
// backward-compatibility assertion at the ROOT: the pre-existing V2
// FareCalculator (fare_calculator.go), completely unmodified by this
// sprint's changes, must still produce byte-identical output to before —
// re-asserts what fare_calculator_test.go already covers, from the V3 test
// suite's vantage point, so a V3 regression test run alone still catches a
// V2 regression.
func TestFareCalculatorV2_UnaffectedByV3Additions(t *testing.T) {
	calc := app.NewFareCalculator(entity.DefaultFareConfig())
	fb, err := calc.Estimate(entity.VehicleTypeCar, 5.0, 15.0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fb.Total != 73500 {
		t.Errorf("Total = %d, want 73500 (unchanged VND formula — see fare_calculator_test.go)", fb.Total)
	}
}
