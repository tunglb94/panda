package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fairride/pricing/domain/entity"
)

func TestDefault_LoadsWithoutError(t *testing.T) {
	cfg := Default()
	if cfg.Fare.CurrencyCode != "VND" {
		t.Errorf("currency = %q, want VND", cfg.Fare.CurrencyCode)
	}
	if len(cfg.Fare.Rates) != 3 {
		t.Errorf("expected 3 vehicle classes, got %d", len(cfg.Fare.Rates))
	}
	for _, vt := range []entity.VehicleType{entity.VehicleTypeCar, entity.VehicleTypeMotorcycle, entity.VehicleTypeVan} {
		if _, ok := cfg.Fare.Rates[vt]; !ok {
			t.Errorf("missing rates for vehicle type %s", vt)
		}
	}
}

func TestDefault_CarMinimumFareRevertedToBRB(t *testing.T) {
	// docs/business/PRICING_V3_REVIEW.md Phần 13 mục 1 — must be 25,000
	// (BRB §2.2.4), not the V3 Design's original 30,000.
	cfg := Default()
	car := cfg.Fare.Rates[entity.VehicleTypeCar]
	if car.MinimumFare != 25000 {
		t.Errorf("car minimum fare = %d, want 25000 (BRB §2.2.4, reverted per PRICING_V3_REVIEW.md)", car.MinimumFare)
	}
}

func TestDefault_CommissionOnlyBronzeChanged_P0_3(t *testing.T) {
	// P0-3: Bronze 20% -> 16%, "Launch only" — Silver/Gold/Platinum/Diamond
	// stay at their original BRB §7.1 values. Deliberately NOT
	// PRICING_V3_REVIEW.md Phần 6.2's full-ladder recommendation (16/15/14/13/12).
	cfg := Default()
	want := map[entity.CommissionTier]float64{
		entity.CommissionTierBronze:   0.16, // changed (BRB: 0.20)
		entity.CommissionTierSilver:   0.18, // unchanged (BRB §7.1)
		entity.CommissionTierGold:     0.16, // unchanged (BRB §7.1)
		entity.CommissionTierPlatinum: 0.14, // unchanged (BRB §7.1)
		entity.CommissionTierDiamond:  0.12, // unchanged (BRB §7.1)
	}
	for tier, rate := range want {
		if got := cfg.Commission.Rate(tier); got != rate {
			t.Errorf("%s rate = %v, want %v", tier, got, rate)
		}
	}
}

func TestDefault_AirportFeesExcludeMotorcycle(t *testing.T) {
	cfg := Default()
	if got := cfg.Airport.FeeFor(entity.VehicleTypeMotorcycle, entity.AirportLegPickup); got != 0 {
		t.Errorf("motorcycle airport pickup fee = %d, want 0", got)
	}
	if got := cfg.Airport.FeeFor(entity.VehicleTypeCar, entity.AirportLegPickup); got == 0 {
		t.Error("car airport pickup fee should be non-zero")
	}
}

func TestDefault_LastTierRatioGuardrailPasses(t *testing.T) {
	// The shipped default config must itself satisfy its own guardrail.
	cfg := Default()
	for vt, rates := range cfg.Fare.Rates {
		if len(rates.DistanceTiers) == 0 {
			t.Fatalf("%s: no distance tiers", vt)
		}
		first := rates.DistanceTiers[0].RatePerKM
		last := rates.DistanceTiers[len(rates.DistanceTiers)-1].RatePerKM
		if float64(last) < float64(first)*0.35 {
			t.Errorf("%s: last tier rate %d is below 35%% of first tier rate %d", vt, last, first)
		}
	}
}

func writeTempConfig(t *testing.T, yamlBody string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "pricing.yaml")
	if err := os.WriteFile(path, []byte(yamlBody), 0o644); err != nil {
		t.Fatalf("write temp config: %v", err)
	}
	return path
}

const validMinimalYAML = `
currency_code: VND
vat_rate: 0.1
last_tier_min_ratio: 0.35
commission:
  bronze: 0.16
vehicles:
  car:
    base_fare: 10000
    minimum_fare: 25000
    booking_fee: 2000
    traffic_time_per_minute: 400
    waiting_fee_per_minute: 500
    waiting_grace_minutes: 3
    distance_tiers:
      - {from_km: 0, to_km: 10, rate_per_km: 5000}
      - {from_km: 10, to_km: 0, rate_per_km: 4000}
airport:
  pickup_fee:
    car: 15000
`

func TestLoad_ValidMinimalConfig(t *testing.T) {
	path := writeTempConfig(t, validMinimalYAML)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Fare.Rates[entity.VehicleTypeCar].BaseFare != 10000 {
		t.Errorf("base fare not loaded correctly")
	}
}

func TestLoad_MissingFile(t *testing.T) {
	_, err := Load(filepath.Join(t.TempDir(), "does-not-exist.yaml"))
	if err == nil {
		t.Fatal("expected an error for a missing file")
	}
}

func TestLoad_RejectsGapInTiers(t *testing.T) {
	yamlBody := strings.Replace(validMinimalYAML, "{from_km: 10, to_km: 0, rate_per_km: 4000}", "{from_km: 12, to_km: 0, rate_per_km: 4000}", 1)
	path := writeTempConfig(t, yamlBody)
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected an error for a gap between distance tiers")
	}
}

func TestLoad_RejectsNonLastOpenEndedTier(t *testing.T) {
	yamlBody := `
currency_code: VND
vat_rate: 0.1
commission:
  bronze: 0.16
vehicles:
  car:
    base_fare: 10000
    minimum_fare: 25000
    booking_fee: 2000
    traffic_time_per_minute: 400
    waiting_fee_per_minute: 500
    waiting_grace_minutes: 3
    distance_tiers:
      - {from_km: 0, to_km: 0, rate_per_km: 5000}
      - {from_km: 10, to_km: 0, rate_per_km: 4000}
`
	path := writeTempConfig(t, yamlBody)
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected an error when an open-ended tier is not last")
	}
}

func TestLoad_RejectsLastTierBelowGuardrailRatio(t *testing.T) {
	yamlBody := `
currency_code: VND
vat_rate: 0.1
last_tier_min_ratio: 0.35
commission:
  bronze: 0.16
vehicles:
  car:
    base_fare: 10000
    minimum_fare: 25000
    booking_fee: 2000
    traffic_time_per_minute: 400
    waiting_fee_per_minute: 500
    waiting_grace_minutes: 3
    distance_tiers:
      - {from_km: 0, to_km: 10, rate_per_km: 10000}
      - {from_km: 10, to_km: 0, rate_per_km: 1000}
`
	path := writeTempConfig(t, yamlBody)
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected an error: last tier (1000) is far below 35% of first tier (10000)")
	}
	if !strings.Contains(err.Error(), "guardrail") {
		t.Errorf("expected error to mention the guardrail, got: %v", err)
	}
}

func TestLoad_RejectsMissingBronzeCommission(t *testing.T) {
	yamlBody := strings.Replace(validMinimalYAML, "  bronze: 0.16", "  gold: 0.14", 1)
	path := writeTempConfig(t, yamlBody)
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected an error: commission.bronze is required as the fallback rate")
	}
}

func TestLoad_RejectsNegativeRate(t *testing.T) {
	yamlBody := strings.Replace(validMinimalYAML, "base_fare: 10000", "base_fare: -1", 1)
	path := writeTempConfig(t, yamlBody)
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected an error for a negative base_fare")
	}
}

func TestLoad_RejectsCommissionRateOver100Percent(t *testing.T) {
	yamlBody := strings.Replace(validMinimalYAML, "  bronze: 0.16", "  bronze: 1.5", 1)
	path := writeTempConfig(t, yamlBody)
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected an error for a commission rate > 100%")
	}
}

func TestLoad_RejectsMissingLastTierMinRatio(t *testing.T) {
	yamlBody := strings.Replace(validMinimalYAML, "last_tier_min_ratio: 0.35\n", "", 1)
	path := writeTempConfig(t, yamlBody)
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected an error: last_tier_min_ratio has no implicit default")
	}
}

func TestLoad_RejectsUnknownVehicleKey(t *testing.T) {
	yamlBody := strings.Replace(validMinimalYAML, "  car:", "  spaceship:", 1)
	path := writeTempConfig(t, yamlBody)
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected an error for an unrecognised vehicle key")
	}
}
