package audit

import (
	"testing"

	"github.com/fairride/ai_simulation/domain/entity"
	"github.com/fairride/ai_simulation/stats"
)

func TestComputeReport_NegativeProfitDetection(t *testing.T) {
	trips := []*entity.SimTrip{
		// Tiny commission after VAT/infra costs goes negative.
		{TripID: "t1", Outcome: entity.OutcomeCompleted, CommissionVND: 100, DriverNetVND: 5000, FinalFareVND: 5100},
		// Healthy trip, no anomaly.
		{TripID: "t2", Outcome: entity.OutcomeCompleted, CommissionVND: 20000, DriverNetVND: 80000, FinalFareVND: 100000},
	}
	r := ComputeReport(trips, nil, nil, stats.Bundle{}, stats.ValidationReport{})

	if r.NegativeProfitTripCount != 1 {
		t.Fatalf("expected exactly 1 negative-profit trip, got %d", r.NegativeProfitTripCount)
	}
	if len(r.NegativeProfitExamples) != 1 || r.NegativeProfitExamples[0] != "t1" {
		t.Errorf("expected t1 in negative profit examples, got %+v", r.NegativeProfitExamples)
	}
}

func TestComputeReport_NegativeDriverIncomeDetection(t *testing.T) {
	trips := []*entity.SimTrip{
		{TripID: "bad-1", Outcome: entity.OutcomeCompleted, DriverNetVND: -500},
		{TripID: "good-1", Outcome: entity.OutcomeCompleted, DriverNetVND: 40000},
	}
	r := ComputeReport(trips, nil, nil, stats.Bundle{}, stats.ValidationReport{})
	if r.NegativeDriverIncomeTripCount != 1 {
		t.Fatalf("expected exactly 1 negative driver income trip, got %d", r.NegativeDriverIncomeTripCount)
	}
}

func TestComputeReport_VoucherUnusedPercent(t *testing.T) {
	bundle := stats.Bundle{PromotionROI: stats.PromotionROI{IssuedVoucherCount: 100, UsedVoucherCount: 40}}
	r := ComputeReport(nil, nil, nil, bundle, stats.ValidationReport{})
	if r.VoucherUnusedPercent != 60 {
		t.Errorf("expected 60%% unused, got %.1f%%", r.VoucherUnusedPercent)
	}
}

func TestComputeDriverFlags_12hPlusAndZeroTrips(t *testing.T) {
	drivers := map[string]*entity.DriverAgent{
		"tired":    {ID: "tired", MaxHoursOnlineContinuous: 12.5},
		"idle":     {ID: "idle", TotalHoursOnline: 20, TripsThisRun: 0},
		"normal":   {ID: "normal", MaxHoursOnlineContinuous: 6, TotalHoursOnline: 10, TripsThisRun: 5},
	}
	var r Report
	computeDriverFlags(&r, drivers)

	if len(r.DriversOnline12hPlus) != 1 || r.DriversOnline12hPlus[0].DriverID != "tired" {
		t.Errorf("expected exactly driver 'tired' flagged for 12h+, got %+v", r.DriversOnline12hPlus)
	}
	if len(r.DriversOnlineZeroTrips) != 1 || r.DriversOnlineZeroTrips[0].DriverID != "idle" {
		t.Errorf("expected exactly driver 'idle' flagged for zero trips, got %+v", r.DriversOnlineZeroTrips)
	}
}

func TestComputeDriverFlags_IncomeOutlier(t *testing.T) {
	drivers := map[string]*entity.DriverAgent{}
	// 20 drivers earning ~1,000,000, one earning 50,000,000 — a clear outlier.
	for i := 0; i < 20; i++ {
		id := "normal-" + string(rune('a'+i))
		drivers[id] = &entity.DriverAgent{ID: id, IncomeWeek: 1_000_000}
	}
	drivers["outlier"] = &entity.DriverAgent{ID: "outlier", IncomeWeek: 50_000_000}

	var r Report
	computeDriverFlags(&r, drivers)

	found := false
	for _, f := range r.DriversHighIncomeOutliers {
		if f.DriverID == "outlier" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected 'outlier' driver flagged as an income outlier, got %+v", r.DriversHighIncomeOutliers)
	}
	for _, f := range r.DriversHighIncomeOutliers {
		if f.DriverID != "outlier" {
			t.Errorf("did not expect a normal-income driver flagged, got %q", f.DriverID)
		}
	}
}

func TestComputeRiderFlags_VoucherSpam(t *testing.T) {
	var trips []*entity.SimTrip
	for i := 0; i < 8; i++ {
		trips = append(trips, &entity.SimTrip{RiderID: "spammer", Outcome: entity.OutcomeCompleted, PromotionType: "manual_coupon"})
	}
	trips = append(trips, &entity.SimTrip{RiderID: "normal", Outcome: entity.OutcomeCompleted, PromotionType: "manual_coupon"})

	var r Report
	computeRiderFlags(&r, trips)

	if len(r.RidersVoucherSpam) != 1 || r.RidersVoucherSpam[0].RiderID != "spammer" {
		t.Errorf("expected only 'spammer' (8 redemptions > threshold 5) flagged, got %+v", r.RidersVoucherSpam)
	}
}

func TestComputeRideDeliveryRatio(t *testing.T) {
	trips := []*entity.SimTrip{
		{Outcome: entity.OutcomeCompleted, Kind: entity.KindRide},
		{Outcome: entity.OutcomeCompleted, Kind: entity.KindRide},
		{Outcome: entity.OutcomeCompleted, Kind: entity.KindDelivery},
	}
	var r Report
	computeRideDeliveryRatio(&r, trips)
	if r.RideCount != 2 || r.DeliveryCount != 1 {
		t.Fatalf("expected ride=2 delivery=1, got ride=%d delivery=%d", r.RideCount, r.DeliveryCount)
	}
	if r.RideToDeliveryRatio != 2.0 {
		t.Errorf("expected ratio 2.0, got %.2f", r.RideToDeliveryRatio)
	}
}

func TestComputeTop50Anomalies_CapsAt50AndRanksBySeverity(t *testing.T) {
	var trips []*entity.SimTrip
	for i := 0; i < 60; i++ {
		trips = append(trips, &entity.SimTrip{
			TripID: "neg-" + string(rune('a'+i%26)) + string(rune('0'+i/26)),
			Outcome: entity.OutcomeCompleted, CommissionVND: 100, DriverNetVND: 100, FinalFareVND: 200,
		})
	}
	anomalies := ComputeTop50Anomalies(trips, nil, Report{})
	if len(anomalies) > 50 {
		t.Errorf("expected at most 50 anomalies, got %d", len(anomalies))
	}
	for i := 1; i < len(anomalies); i++ {
		if anomalies[i].Severity > anomalies[i-1].Severity {
			t.Errorf("expected anomalies ranked by descending severity, violated at index %d", i)
		}
	}
}

func TestComputeTop20Risks_CapsAt20(t *testing.T) {
	r := Report{
		RevenueLeakPercent: 5, NegativeProfitTripCount: 100, NegativeDriverIncomeTripCount: 50,
		VoucherUnusedPercent: 80, VoucherIssuedCount: 100, SurgeCausingLossTripCount: 20,
		DriversOnline12hPlus:      make([]DriverFlag, 10),
		DriversOnlineZeroTrips:    make([]DriverFlag, 10),
		DriversHighIncomeOutliers: make([]DriverFlag, 10),
		RidersVoucherSpam:         make([]RiderFlag, 10),
		ZoneStats: []ZoneStat{
			{Zone: "airport", AverageDemandSupply: 10, AverageETAMinutes: 40},
		},
		PeakHourProfit: SegmentProfit{TripCount: 10, AverageProfitVND: 20000},
		OffPeakProfit:  SegmentProfit{TripCount: 10, AverageProfitVND: 1000},
	}
	risks := ComputeTop20Risks(r)
	if len(risks) > 20 {
		t.Errorf("expected at most 20 risks, got %d", len(risks))
	}
	if len(risks) == 0 {
		t.Fatalf("expected at least one risk given a report full of red flags")
	}
}

func TestDetectBugs_OnlyCriticalPlusStructural(t *testing.T) {
	validation := stats.ValidationReport{
		Passed: false,
		Warnings: []stats.ValidationWarning{
			{Check: "commission", Severity: "warning", Message: "minor drift"},
			{Check: "revenue_balance", Severity: "critical", Message: "major imbalance"},
		},
	}
	bugs := DetectBugs(validation)
	// StructuralBugs() (currently 1: seed non-determinism) is always
	// included regardless of this run's validation result, plus exactly
	// the one critical warning above (the "warning"-severity one must not
	// be surfaced as a bug).
	wantCount := len(StructuralBugs()) + 1
	if len(bugs) != wantCount {
		t.Fatalf("expected exactly %d bugs (structural + 1 critical warning), got %d: %+v", wantCount, len(bugs), bugs)
	}
	foundCritical := false
	for _, b := range bugs {
		if b.Cause == "major imbalance" {
			foundCritical = true
		}
	}
	if !foundCritical {
		t.Errorf("expected the critical warning's message to appear as a bug's Cause, got %+v", bugs)
	}
}

func TestDetectBugs_NoCriticalStillReturnsStructuralBugs(t *testing.T) {
	validation := stats.ValidationReport{Passed: true, Warnings: []stats.ValidationWarning{
		{Check: "revenue_balance", Severity: "info", Message: "tiny rounding drift"},
	}}
	bugs := DetectBugs(validation)
	// No critical validation warning this run, but StructuralBugs() (a
	// property of the code, not of one run) must still be reported.
	if len(bugs) != len(StructuralBugs()) {
		t.Errorf("expected exactly the structural bugs (no critical run-specific ones), got %d: %+v", len(bugs), bugs)
	}
}
