package stats

import (
	"testing"

	"github.com/fairride/ai_simulation/domain/entity"
)

func TestValidate_CleanDataPasses(t *testing.T) {
	c := NewCollector()
	trips := []*entity.SimTrip{
		{Outcome: entity.OutcomeCompleted, BaseFareVND: 50_000, FinalFareVND: 52_000, CommissionVND: 10_000, DriverNetVND: 42_000},
	}
	drivers := map[string]*entity.DriverAgent{"d1": {IncomeToday: 42_000, IncomeWeek: 42_000}}
	bi := c.BuildBusinessIntelligence(trips, 80, 80)

	report := c.Validate(trips, drivers, bi)
	for _, w := range report.Warnings {
		if w.Severity == "critical" {
			t.Errorf("expected no critical warnings on clean data, got %+v", w)
		}
	}
	if !report.Passed {
		t.Errorf("expected Passed=true on clean data, got warnings %+v", report.Warnings)
	}
}

func TestValidate_NegativeDriverIncomeFlagsCritical(t *testing.T) {
	c := NewCollector()
	drivers := map[string]*entity.DriverAgent{"d1": {IncomeToday: -5000}}
	report := c.Validate(nil, drivers, BusinessIntelligence{})

	if report.Passed {
		t.Fatalf("expected Passed=false when a driver has negative income")
	}
	found := false
	for _, w := range report.Warnings {
		if w.Check == "driver_income" && w.Severity == "critical" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected a critical driver_income warning, got %+v", report.Warnings)
	}
}

func TestValidate_NegativeCommissionFlagsCritical(t *testing.T) {
	c := NewCollector()
	trips := []*entity.SimTrip{
		{Outcome: entity.OutcomeCompleted, CommissionVND: -1000, DriverNetVND: 40_000, FinalFareVND: 39_000},
	}
	report := c.Validate(trips, nil, BusinessIntelligence{})

	if report.Passed {
		t.Fatalf("expected Passed=false when commission is negative")
	}
	found := false
	for _, w := range report.Warnings {
		if w.Check == "commission" && w.Severity == "critical" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected a critical commission warning, got %+v", report.Warnings)
	}
}

func TestValidate_CommissionDriverNetConservation(t *testing.T) {
	c := NewCollector()
	// commission+driverNet (10,000+42,000=52,000) matches FinalFareVND
	// (52,000) exactly — the real invariant Split() guarantees.
	trips := []*entity.SimTrip{
		{Outcome: entity.OutcomeCompleted, CommissionVND: 10_000, DriverNetVND: 42_000, FinalFareVND: 52_000},
	}
	report := c.Validate(trips, nil, BusinessIntelligence{})
	for _, w := range report.Warnings {
		if w.Check == "commission" {
			t.Errorf("expected no commission warning when commission+driverNet == FinalFareVND, got %+v", w)
		}
	}
}

func TestValidate_CommissionDriverNetMismatchFlagsWarning(t *testing.T) {
	c := NewCollector()
	// commission+driverNet (10,000+42,000=52,000) is 10,000 VND more than
	// FinalFareVND (42,000) — a real accounting mismatch.
	trips := []*entity.SimTrip{
		{Outcome: entity.OutcomeCompleted, CommissionVND: 10_000, DriverNetVND: 42_000, FinalFareVND: 42_000},
	}
	report := c.Validate(trips, nil, BusinessIntelligence{})
	found := false
	for _, w := range report.Warnings {
		if w.Check == "commission" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected a commission conservation warning on a real mismatch, got %+v", report.Warnings)
	}
}

func TestValidate_RevenueBalanceDriftFlagsCritical(t *testing.T) {
	c := NewCollector()
	bi := BusinessIntelligence{
		GMVVND: 1_000_000, DriverRevenueVND: 100_000, PlatformRevenueVND: 50_000,
		VoucherCostVND: 0, PromotionCostVND: 0, // accounted total (150,000) is 85% short of GMV
	}
	report := c.Validate(nil, nil, bi)
	if report.Passed {
		t.Fatalf("expected Passed=false on an 85%% revenue balance drift")
	}
}
