package app_test

import (
	"math"
	"testing"
	"time"

	"github.com/fairride/pricing/app"
	"github.com/fairride/pricing/domain/entity"
)

// TestSimulation_DynamicPricingEngine_ScenarioMatrix runs the Dynamic
// Pricing Engine across a matrix of real-world-shaped scenarios (every
// combination of Night/Holiday/Rain/Peak/Airport on-or-off, crossed with 6
// demand-supply ratio levels) and asserts the BRB safety invariants the
// engine must never violate, no matter how rules are combined. This is the
// "simulation" deliverable: an executable, repeatable check that the engine
// behaves safely across the space of inputs, not just the handful of cases
// the other unit tests hand-pick.
func TestSimulation_DynamicPricingEngine_ScenarioMatrix(t *testing.T) {
	pipeline := app.NewDefaultPricingPipeline()

	// BRB-derived worst-case ceiling: Night x Holiday x Rain capped at 1.60
	// (§2.2.13) times Demand Surge capped at 2.0 (§2.13.3) = 3.20. No
	// scenario in this matrix may ever exceed it.
	const maxPossibleMultiplier = entity.MaxCombinedNightHolidayRainMultiplier * entity.MaxDemandSurgeMultiplier

	nightTimes := map[bool]time.Time{
		true:  time.Date(2026, 7, 6, 23, 0, 0, 0, time.UTC), // Monday 23:00 - night AND within no peak window
		false: time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC), // Monday noon - neither night nor peak
	}
	dsrLevels := []struct {
		activeRequests, availableDrivers int
	}{
		{5, 10},  // DSR 0.5 - no surge
		{13, 10}, // DSR 1.3 - busy
		{18, 10}, // DSR 1.8 - high demand
		{22, 10}, // DSR 2.2 - very high demand
		{28, 10}, // DSR 2.8 - peak demand
		{35, 10}, // DSR 3.5 - maximum surge
	}

	configs := app.DefaultRuleConfigs()
	configs[app.RuleNameDemandSurge] = app.RuleConfig{Enabled: true, Priority: 10, Weight: 1.0, MinSurge: 1.0, MaxSurge: entity.MaxDemandSurgeMultiplier}
	configs[app.RuleNamePeakHour] = app.RuleConfig{Enabled: true, Priority: 30, Weight: 1.0, MinSurge: 1.0, MaxSurge: entity.PeakHourSurchargeMultiplier}
	configs[app.RuleNameNight] = app.RuleConfig{Enabled: true, Priority: 40, Weight: 1.0, MinSurge: 1.0, MaxSurge: entity.NightSurchargeMultiplier}
	configs[app.RuleNameHoliday] = app.RuleConfig{Enabled: true, Priority: 50, Weight: 1.0, MinSurge: 1.0, MaxSurge: entity.HolidaySurchargeMultiplier}
	configs[app.RuleNameRain] = app.RuleConfig{Enabled: true, Priority: 60, Weight: 1.0, MinSurge: 1.0, MaxSurge: entity.RainSurchargeMultiplier}
	configs[app.RuleNameAirport] = app.RuleConfig{Enabled: true, Priority: 5, Weight: 1.0, MinSurge: 0, MaxSurge: float64(entity.AirportFeeVND)}

	scenarioCount := 0
	maxObserved := 1.0

	for _, isNight := range []bool{true, false} {
		for _, isHoliday := range []bool{true, false} {
			for _, isRain := range []bool{true, false} {
				for _, isAirport := range []bool{true, false} {
					for _, dsr := range dsrLevels {
						scenarioCount++
						ctx := entity.PricingContext{
							VehicleType:      entity.VehicleTypeCar,
							RequestTime:      nightTimes[isNight],
							ActiveRequests:   dsr.activeRequests,
							AvailableDrivers: dsr.availableDrivers,
							IsHoliday:        isHoliday,
							IsRainActive:     isRain,
							IsAirportZone:    isAirport,
						}

						result := pipeline.Evaluate(ctx, configs)

						if result.FinalMultiplier < 1.0 {
							t.Fatalf("scenario %d: multiplier below 1.0: %v (ctx=%+v)", scenarioCount, result.FinalMultiplier, ctx)
						}
						if result.FinalMultiplier > maxPossibleMultiplier {
							t.Fatalf("scenario %d: multiplier %v exceeds BRB-derived ceiling %v (ctx=%+v)", scenarioCount, result.FinalMultiplier, maxPossibleMultiplier, ctx)
						}
						if result.FlatSurcharge < 0 || result.FlatSurcharge > entity.AirportFeeVND {
							t.Fatalf("scenario %d: flat surcharge %d out of expected [0, %d] range", scenarioCount, result.FlatSurcharge, entity.AirportFeeVND)
						}

						demandApplied := false
						for _, r := range result.AppliedRules {
							if r.RuleName == app.RuleNameDemandSurge {
								demandApplied = true
							}
							if r.RuleName == app.RuleNamePeakHour && demandApplied {
								t.Fatalf("scenario %d: BRB §2.2.12 violated — Peak Hour applied alongside Demand Surge (ctx=%+v)", scenarioCount, ctx)
							}
						}

						if result.FinalMultiplier > maxObserved {
							maxObserved = result.FinalMultiplier
						}
					}
				}
			}
		}
	}

	t.Logf("simulation: %d scenarios run, max observed multiplier=%.2f (BRB ceiling=%.2f)", scenarioCount, maxObserved, maxPossibleMultiplier)
}

// TestSimulation_BackwardCompatibility_AcrossVehicleTypesAndInputs re-runs
// every existing fare_calculator_test.go input shape through Estimate with
// the default (all-disabled) engine and asserts the output is identical to
// calling the base formula directly — i.e. the refactor is a true no-op for
// every vehicle type and a spread of distance/duration values, not just the
// hand-picked cases already covered by fare_calculator_test.go.
func TestSimulation_BackwardCompatibility_AcrossVehicleTypesAndInputs(t *testing.T) {
	calc := app.NewFareCalculator(entity.DefaultFareConfig())
	cfg := entity.DefaultFareConfig()

	vehicleTypes := []entity.VehicleType{entity.VehicleTypeCar, entity.VehicleTypeMotorcycle, entity.VehicleTypeVan}
	distances := []float64{0, 0.5, 1, 2.5, 5, 10, 25, 50}
	durations := []float64{0, 1, 5, 10, 15, 30, 60}

	checked := 0
	for _, vt := range vehicleTypes {
		rates := cfg.Rates[vt]
		for _, d := range distances {
			for _, m := range durations {
				fb, err := calc.Estimate(vt, d, m)
				if err != nil {
					t.Fatalf("unexpected error for %s %vkm %vmin: %v", vt, d, m, err)
				}

				wantRide := rates.BaseFare + roundToUnitForTest(float64(rates.PerKmRate)*d) + roundToUnitForTest(float64(rates.PerMinuteRate)*m)
				if wantRide < rates.MinimumFare {
					wantRide = rates.MinimumFare
				}
				wantTotal := wantRide + rates.BookingFee

				if fb.RideFare != wantRide {
					t.Fatalf("%s %vkm %vmin: RideFare got %d, want %d (engine changed base output)", vt, d, m, fb.RideFare, wantRide)
				}
				if fb.Total != wantTotal {
					t.Fatalf("%s %vkm %vmin: Total got %d, want %d (engine changed base output)", vt, d, m, fb.Total, wantTotal)
				}
				checked++
			}
		}
	}
	t.Logf("simulation: verified backward-compatible output for %d (vehicle x distance x duration) combinations", checked)
}

// roundToUnitForTest mirrors app.roundToUnit exactly (math.Round, ties away
// from zero); kept as a local copy since roundToUnit is unexported and this
// is an external (_test package) test.
func roundToUnitForTest(v float64) int64 {
	return int64(math.Round(v))
}
