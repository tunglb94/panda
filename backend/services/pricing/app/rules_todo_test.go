package app_test

import (
	"strings"
	"testing"

	"github.com/fairride/pricing/app"
	"github.com/fairride/pricing/domain/entity"
)

func TestTODORules_NeverContribute(t *testing.T) {
	rules := []app.PricingRule{
		app.NewSupplySurgeRule(),
		app.NewTrafficSurgeRule(),
		app.NewSpecialEventRule(),
	}

	// A deliberately "surge-looking" context — even so, none of these rules
	// may contribute, since BRB defines no formula for them.
	ctx := entity.PricingContext{
		ActiveRequests:   100,
		AvailableDrivers: 1,
		IsRainActive:     true,
		IsHoliday:        true,
		IsAirportZone:    true,
		IsSpecialEvent:   true,
		TrafficLevel:     entity.TrafficLevelHeavy,
	}

	for _, rule := range rules {
		out := rule.Evaluate(ctx)
		if out.Applied {
			t.Fatalf("%s: TODO rule must never apply (no BRB-approved formula)", rule.Name())
		}
		if out.Multiplier != 0 || out.FlatAmount != 0 {
			t.Fatalf("%s: TODO rule must never produce a non-zero contribution, got multiplier=%v flat=%v", rule.Name(), out.Multiplier, out.FlatAmount)
		}
		if !strings.Contains(out.Reason, "TODO") {
			t.Fatalf("%s: expected reason to disclose TODO status, got %q", rule.Name(), out.Reason)
		}
		if out.Category != entity.CategoryNotDefined {
			t.Fatalf("%s: expected CategoryNotDefined, got %v", rule.Name(), out.Category)
		}
	}
}
