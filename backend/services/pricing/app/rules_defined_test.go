package app_test

import (
	"testing"
	"time"

	"github.com/fairride/pricing/app"
	"github.com/fairride/pricing/domain/entity"
)

func TestDemandSurgeRule_TierLookup(t *testing.T) {
	rule := app.NewDemandSurgeRule(entity.DefaultDSRTiers())

	cases := []struct {
		name             string
		activeRequests   int
		availableDrivers int
		wantApplied      bool
		wantMultiplier   float64
	}{
		{"no demand", 0, 10, false, 1.0},
		{"DSR 1.0 -> normal", 10, 10, false, 1.0},
		{"DSR 1.3 -> busy", 13, 10, true, 1.2},
		{"DSR 1.8 -> high demand", 18, 10, true, 1.4},
		{"DSR 2.2 -> very high demand", 22, 10, true, 1.6},
		{"DSR 2.8 -> peak demand", 28, 10, true, 1.8},
		{"DSR 3.5 -> maximum surge", 35, 10, true, 2.0},
		{"DSR way above max still capped at 2.0", 1000, 1, true, 2.0},
		{"no drivers at all, demand present -> maximum surge", 5, 0, true, 2.0},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := entity.PricingContext{ActiveRequests: tc.activeRequests, AvailableDrivers: tc.availableDrivers}
			out := rule.Evaluate(ctx)
			if out.Applied != tc.wantApplied {
				t.Fatalf("Applied: got %v, want %v (reason=%q)", out.Applied, tc.wantApplied, out.Reason)
			}
			if tc.wantApplied && out.Multiplier != tc.wantMultiplier {
				t.Fatalf("Multiplier: got %v, want %v", out.Multiplier, tc.wantMultiplier)
			}
			if out.Multiplier > entity.MaxDemandSurgeMultiplier {
				t.Fatalf("BRB §2.13.3 violated: multiplier %v exceeds hard cap %v", out.Multiplier, entity.MaxDemandSurgeMultiplier)
			}
		})
	}
}

func TestPeakHourRule(t *testing.T) {
	rule := app.NewPeakHourRule(entity.DefaultPeakHourWindows())

	monday830 := time.Date(2026, 7, 6, 8, 30, 0, 0, time.UTC) // Monday
	monday1200 := time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)
	saturday830 := time.Date(2026, 7, 11, 8, 30, 0, 0, time.UTC) // Saturday, not a weekday peak

	if out := rule.Evaluate(entity.PricingContext{RequestTime: monday830}); !out.Applied || out.Multiplier != entity.PeakHourSurchargeMultiplier {
		t.Fatalf("expected morning peak on Monday 08:30, got %+v", out)
	}
	if out := rule.Evaluate(entity.PricingContext{RequestTime: monday1200}); out.Applied {
		t.Fatalf("expected no peak at Monday noon, got %+v", out)
	}
	if out := rule.Evaluate(entity.PricingContext{RequestTime: saturday830}); out.Applied {
		t.Fatalf("expected no peak on Saturday (BRB default windows are Mon-Fri), got %+v", out)
	}
}

func TestNightSurchargeRule(t *testing.T) {
	rule := app.NewNightSurchargeRule(entity.NightWindowStartHour, entity.NightWindowEndHour)

	night2230 := time.Date(2026, 7, 6, 22, 30, 0, 0, time.UTC)
	night0300 := time.Date(2026, 7, 6, 3, 0, 0, 0, time.UTC)
	day1400 := time.Date(2026, 7, 6, 14, 0, 0, 0, time.UTC)
	boundary0500 := time.Date(2026, 7, 6, 5, 0, 0, 0, time.UTC)

	if out := rule.Evaluate(entity.PricingContext{RequestTime: night2230}); !out.Applied || out.Multiplier != entity.NightSurchargeMultiplier {
		t.Fatalf("expected night surcharge at 22:30, got %+v", out)
	}
	if out := rule.Evaluate(entity.PricingContext{RequestTime: night0300}); !out.Applied {
		t.Fatalf("expected night surcharge at 03:00 (wraps past midnight), got %+v", out)
	}
	if out := rule.Evaluate(entity.PricingContext{RequestTime: day1400}); out.Applied {
		t.Fatalf("expected no night surcharge at 14:00, got %+v", out)
	}
	if out := rule.Evaluate(entity.PricingContext{RequestTime: boundary0500}); out.Applied {
		t.Fatalf("expected no night surcharge exactly at 05:00 (end hour exclusive), got %+v", out)
	}
}

func TestHolidaySurchargeRule(t *testing.T) {
	rule := app.NewHolidaySurchargeRule()

	if out := rule.Evaluate(entity.PricingContext{IsHoliday: true}); !out.Applied || out.Multiplier != entity.HolidaySurchargeMultiplier {
		t.Fatalf("expected holiday surcharge, got %+v", out)
	}
	if out := rule.Evaluate(entity.PricingContext{IsHoliday: false}); out.Applied {
		t.Fatalf("expected no holiday surcharge, got %+v", out)
	}
}

func TestRainSurchargeRule(t *testing.T) {
	rule := app.NewRainSurchargeRule()

	if out := rule.Evaluate(entity.PricingContext{IsRainActive: true}); !out.Applied || out.Multiplier != entity.RainSurchargeMultiplier {
		t.Fatalf("expected rain surcharge, got %+v", out)
	}
	if out := rule.Evaluate(entity.PricingContext{IsRainActive: false}); out.Applied {
		t.Fatalf("expected no rain surcharge, got %+v", out)
	}
}

func TestAirportFeeRule(t *testing.T) {
	rule := app.NewAirportFeeRule()

	if out := rule.Evaluate(entity.PricingContext{IsAirportZone: true}); !out.Applied || out.FlatAmount != entity.AirportFeeVND {
		t.Fatalf("expected airport fee %d, got %+v", entity.AirportFeeVND, out)
	}
	if out := rule.Evaluate(entity.PricingContext{IsAirportZone: false}); out.Applied {
		t.Fatalf("expected no airport fee, got %+v", out)
	}
}
