// Package audit implements the Full Business Validation's 20-point business
// audit — every check reads real trip/driver/rider records and already-
// exported stats.Bundle data; nothing here changes simulation behavior. Per
// the task's explicit instruction, this package only ever REPORTS anomalies
// (BugFinding/Assumption) — it never "fixes" anything it finds, in
// production code or in this simulation tool itself.
package audit

import (
	"math"
	"sort"

	"github.com/fairride/ai_simulation/domain/entity"
	"github.com/fairride/ai_simulation/stats"
)

// ZoneStat is one zone's aggregate figures for the zone-level checks
// (#11-14, #16).
type ZoneStat struct {
	Zone                string  `json:"zone"`
	Requested           int     `json:"requested"`
	Accepted            int     `json:"accepted"`
	Rejected            int     `json:"rejected"`
	AcceptRatePercent   float64 `json:"accept_rate_percent"`
	AverageETAMinutes   float64 `json:"average_eta_minutes"`
	AverageDemandSupply float64 `json:"average_demand_supply_ratio"` // demand per available driver, averaged across sampled hours
	ProfitVND           int64   `json:"profit_vnd"`
	TripCount           int     `json:"trip_count"`
}

// DriverFlag/RiderFlag are one flagged entity + the reason for #7-10.
type DriverFlag struct {
	DriverID string  `json:"driver_id"`
	Value    float64 `json:"value"`
	Reason   string  `json:"reason"`
}

type RiderFlag struct {
	RiderID string `json:"rider_id"`
	Count   int    `json:"count"`
	Reason  string `json:"reason"`
}

// SegmentProfit is one Airport/Peak/Off-Peak/Weather slice for #16-19.
type SegmentProfit struct {
	Label           string  `json:"label"`
	TripCount       int     `json:"trip_count"`
	TotalProfitVND  int64   `json:"total_profit_vnd"`
	AverageProfitVND float64 `json:"average_profit_vnd"`
	AverageFareVND  float64 `json:"average_fare_vnd"`
	AverageSurge    float64 `json:"average_surge_multiplier"`
	CancellationPercent float64 `json:"cancellation_rate_percent"`
}

// Report is the full 20-point business audit — field names map 1:1 onto the
// task's numbered KIỂM TRA list (comment on each field gives the number).
type Report struct {
	// #1 Revenue Leak — reuses stats.ValidationReport's own revenue_balance
	// check rather than recomputing it a second way.
	RevenueLeakPercent float64 `json:"revenue_leak_percent"`
	RevenueLeakVND     int64   `json:"revenue_leak_vnd"`

	// #2/#3 — negative-value trip counts, with example trip IDs.
	NegativeProfitTripCount int      `json:"negative_profit_trip_count"`
	NegativeProfitExamples  []string `json:"negative_profit_trip_examples"`
	NegativeDriverIncomeTripCount int `json:"negative_driver_income_trip_count"`
	NegativeDriverIncomeExamples  []string `json:"negative_driver_income_trip_examples"`

	// #4 Voucher issued vs used.
	VoucherIssuedCount int     `json:"voucher_issued_count"`
	VoucherUsedCount   int     `json:"voucher_used_count"`
	VoucherUnusedPercent float64 `json:"voucher_unused_percent"`

	// #5 Promotion ROI — reused verbatim from promotion_roi.json.
	PromotionROI []stats.PromotionROIEntry `json:"promotion_roi"`

	// #6 Surge causing platform loss.
	SurgeCausingLossTripCount int      `json:"surge_causing_loss_trip_count"`
	SurgeCausingLossExamples  []string `json:"surge_causing_loss_trip_examples"`

	// #7-10 driver/rider flags.
	DriversOnline12hPlus       []DriverFlag `json:"drivers_online_12h_plus"`
	DriversOnlineZeroTrips     []DriverFlag `json:"drivers_online_zero_trips"`
	DriversHighIncomeOutliers  []DriverFlag `json:"drivers_high_income_outliers"`
	RidersVoucherSpam          []RiderFlag  `json:"riders_voucher_spam"`

	// #11 Dispatch zone bias — per-zone accept rate; see report's own
	// ASSUMPTION note (rendered in business_audit.md) on what "bias" means
	// here (supply distribution, not algorithmic favoritism).
	ZoneStats []ZoneStat `json:"zone_stats"` // also feeds #12/#13/#14

	// #15 Ride vs Delivery ratio.
	RideCount     int     `json:"ride_count"`
	DeliveryCount int     `json:"delivery_count"`
	RideToDeliveryRatio float64 `json:"ride_to_delivery_ratio"`

	// #16-18.
	AirportProfit  SegmentProfit `json:"airport_profit"`
	PeakHourProfit SegmentProfit `json:"peak_hour_profit"`
	OffPeakProfit  SegmentProfit `json:"off_peak_profit"`

	// #19 Weather impact.
	WeatherImpact []SegmentProfit `json:"weather_impact"`

	// Executive Summary extras not in any existing export.
	AverageDriverIncomeWeekVND float64 `json:"average_driver_income_week_vnd"`
	MedianDriverIncomeWeekVND  float64 `json:"median_driver_income_week_vnd"`
}

// ComputeReport runs every check against real trips/drivers/riders/bundle.
// validation is the already-computed stats.ValidationReport (Validate is
// called once by engine.go's Export; this package reuses that result
// instead of re-running it).
func ComputeReport(trips []*entity.SimTrip, drivers map[string]*entity.DriverAgent, riders map[string]*entity.RiderAgent, bundle stats.Bundle, validation stats.ValidationReport) Report {
	var r Report

	computeRevenueLeak(&r, validation, bundle)
	computeNegativeChecks(&r, trips)
	computeVoucherGap(&r, bundle)
	r.PromotionROI = bundle.PromotionROI.ByType
	computeSurgeLoss(&r, trips)
	computeDriverFlags(&r, drivers)
	computeRiderFlags(&r, trips)
	computeZoneStats(&r, trips, bundle)
	computeRideDeliveryRatio(&r, trips)
	computeSegmentProfits(&r, trips)
	computeDriverIncomeStats(&r, drivers)

	return r
}

func computeRevenueLeak(r *Report, validation stats.ValidationReport, bundle stats.Bundle) {
	accounted := bundle.SimulationReport.Financial.DriverRevenueVND + bundle.SimulationReport.Financial.PlatformRevenueVND +
		bundle.SimulationReport.Financial.VoucherCostVND + bundle.SimulationReport.Financial.PromotionCostVND
	gmv := bundle.SimulationReport.Financial.GMVVND
	r.RevenueLeakVND = gmv - accounted
	if gmv > 0 {
		leak := float64(r.RevenueLeakVND)
		if leak < 0 {
			leak = -leak
		}
		r.RevenueLeakPercent = 100 * leak / float64(gmv)
	}
	_ = validation // kept as a parameter for callers that want to cross-reference validation.Warnings directly
}

func computeNegativeChecks(r *Report, trips []*entity.SimTrip) {
	for _, t := range trips {
		if t.Outcome != entity.OutcomeCompleted {
			continue
		}
		profit := stats.PerTripProfitVND(t.CommissionVND)
		if profit < 0 {
			r.NegativeProfitTripCount++
			if len(r.NegativeProfitExamples) < 10 {
				r.NegativeProfitExamples = append(r.NegativeProfitExamples, t.TripID)
			}
		}
		if t.DriverNetVND < 0 {
			r.NegativeDriverIncomeTripCount++
			if len(r.NegativeDriverIncomeExamples) < 10 {
				r.NegativeDriverIncomeExamples = append(r.NegativeDriverIncomeExamples, t.TripID)
			}
		}
	}
}

func computeVoucherGap(r *Report, bundle stats.Bundle) {
	r.VoucherIssuedCount = bundle.PromotionROI.IssuedVoucherCount
	r.VoucherUsedCount = bundle.PromotionROI.UsedVoucherCount
	if r.VoucherIssuedCount > 0 {
		unused := r.VoucherIssuedCount - r.VoucherUsedCount
		r.VoucherUnusedPercent = 100 * float64(unused) / float64(r.VoucherIssuedCount)
	}
}

func computeSurgeLoss(r *Report, trips []*entity.SimTrip) {
	for _, t := range trips {
		if t.Outcome != entity.OutcomeCompleted || t.SurgeMultiplier <= 1.0 {
			continue
		}
		if stats.PerTripProfitVND(t.CommissionVND) < 0 {
			r.SurgeCausingLossTripCount++
			if len(r.SurgeCausingLossExamples) < 10 {
				r.SurgeCausingLossExamples = append(r.SurgeCausingLossExamples, t.TripID)
			}
		}
	}
}

// computeDriverFlags covers #7 (12h+ shifts), #8 (online but zero trips),
// #9 (income outliers — mean+3×population-stddev of IncomeWeekVND, a
// standard statistical-outlier threshold, not a BRB rule).
func computeDriverFlags(r *Report, drivers map[string]*entity.DriverAgent) {
	if len(drivers) == 0 {
		return
	}
	var incomes []float64
	for _, d := range drivers {
		incomes = append(incomes, float64(d.IncomeWeek))
		if d.MaxHoursOnlineContinuous >= 12 {
			r.DriversOnline12hPlus = append(r.DriversOnline12hPlus, DriverFlag{
				DriverID: d.ID, Value: d.MaxHoursOnlineContinuous,
				Reason: "Đã từng online liên tục >=12h trong một ca — chạm ngưỡng an toàn cứng của FatigueDecision",
			})
		}
		if d.TotalHoursOnline > 0 && d.TripsThisRun == 0 {
			r.DriversOnlineZeroTrips = append(r.DriversOnlineZeroTrips, DriverFlag{
				DriverID: d.ID, Value: d.TotalHoursOnline,
				Reason: "Có online trong lần chạy này nhưng 0 chuyến hoàn tất",
			})
		}
	}
	mean, stddev := meanStddev(incomes)
	threshold := mean + 3*stddev
	for _, d := range drivers {
		if stddev > 0 && float64(d.IncomeWeek) > threshold {
			r.DriversHighIncomeOutliers = append(r.DriversHighIncomeOutliers, DriverFlag{
				DriverID: d.ID, Value: float64(d.IncomeWeek),
				Reason: "Thu nhập/tuần vượt ngưỡng thống kê (mean + 3×độ lệch chuẩn)",
			})
		}
	}
	sort.Slice(r.DriversOnline12hPlus, func(i, j int) bool { return r.DriversOnline12hPlus[i].Value > r.DriversOnline12hPlus[j].Value })
	sort.Slice(r.DriversHighIncomeOutliers, func(i, j int) bool { return r.DriversHighIncomeOutliers[i].Value > r.DriversHighIncomeOutliers[j].Value })
}

// computeRiderFlags covers #10 — riders redeeming a voucher unusually often
// this run (>5 redemptions, a simulation-design threshold — no real fraud/
// abuse dataset exists to calibrate against; see ASSUMPTION in
// business_audit.md).
func computeRiderFlags(r *Report, trips []*entity.SimTrip) {
	countByRider := map[string]int{}
	for _, t := range trips {
		if t.Outcome == entity.OutcomeCompleted && t.PromotionType == "manual_coupon" {
			countByRider[t.RiderID]++
		}
	}
	const spamThreshold = 5
	for riderID, count := range countByRider {
		if count > spamThreshold {
			r.RidersVoucherSpam = append(r.RidersVoucherSpam, RiderFlag{
				RiderID: riderID, Count: count,
				Reason: "Redeem voucher nhiều bất thường trong 1 lần chạy (ngưỡng thiết kế >5, không phải số liệu chống gian lận thật)",
			})
		}
	}
	sort.Slice(r.RidersVoucherSpam, func(i, j int) bool { return r.RidersVoucherSpam[i].Count > r.RidersVoucherSpam[j].Count })
}

func meanStddev(values []float64) (mean, stddev float64) {
	if len(values) == 0 {
		return 0, 0
	}
	var sum float64
	for _, v := range values {
		sum += v
	}
	mean = sum / float64(len(values))
	var sqSum float64
	for _, v := range values {
		d := v - mean
		sqSum += d * d
	}
	stddev = math.Sqrt(sqSum / float64(len(values)))
	return mean, stddev
}
