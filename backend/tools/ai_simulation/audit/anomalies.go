package audit

import (
	"sort"

	"github.com/fairride/ai_simulation/domain/entity"
	"github.com/fairride/ai_simulation/stats"
)

// Anomaly is one entry in top_50_anomalies.json — a single anomalous
// trip/driver/rider record with enough identifying detail to look it up
// directly (TripID/DriverID/RiderID), ranked by Severity across every
// anomaly type this audit checks for (not just one category), so "top 50"
// genuinely means the 50 most severe findings across the whole run.
type Anomaly struct {
	Type     string  `json:"type"` // "negative_profit" | "negative_driver_income" | "surge_loss" | "driver_12h_plus" | "driver_zero_trips" | "driver_income_outlier" | "rider_voucher_spam" | "zone_undersupplied"
	EntityID string  `json:"entity_id"`
	Severity float64 `json:"severity"` // internal ranking magnitude, larger = more severe
	Detail   string  `json:"detail"`
}

// ComputeTop50Anomalies re-scans the same trip/driver/rider data
// ComputeReport already checked, but keeps every individual flagged record
// (not just counts/examples) so the anomaly list is independently complete
// even if a caller only has this file, not the full Report.
func ComputeTop50Anomalies(trips []*entity.SimTrip, drivers map[string]*entity.DriverAgent, r Report) []Anomaly {
	var out []Anomaly

	for _, t := range trips {
		if t.Outcome != entity.OutcomeCompleted {
			continue
		}
		profit := stats.PerTripProfitVND(t.CommissionVND)
		if profit < 0 {
			sev := float64(-profit)
			if t.SurgeMultiplier > 1.0 {
				out = append(out, Anomaly{Type: "surge_loss", EntityID: t.TripID, Severity: sev + 1000, Detail: describeSurgeLoss(t, profit)})
			} else {
				out = append(out, Anomaly{Type: "negative_profit", EntityID: t.TripID, Severity: sev, Detail: describeNegativeProfit(t, profit)})
			}
		}
		if t.DriverNetVND < 0 {
			out = append(out, Anomaly{Type: "negative_driver_income", EntityID: t.TripID, Severity: float64(-t.DriverNetVND) + 2000, Detail: describeNegativeDriverIncome(t)})
		}
	}

	for _, f := range r.DriversOnline12hPlus {
		out = append(out, Anomaly{Type: "driver_12h_plus", EntityID: f.DriverID, Severity: f.Value * 10, Detail: f.Reason})
	}
	for _, f := range r.DriversOnlineZeroTrips {
		out = append(out, Anomaly{Type: "driver_zero_trips", EntityID: f.DriverID, Severity: f.Value * 5, Detail: f.Reason})
	}
	for _, f := range r.DriversHighIncomeOutliers {
		out = append(out, Anomaly{Type: "driver_income_outlier", EntityID: f.DriverID, Severity: f.Value / 1000, Detail: f.Reason})
	}
	for _, f := range r.RidersVoucherSpam {
		out = append(out, Anomaly{Type: "rider_voucher_spam", EntityID: f.RiderID, Severity: float64(f.Count) * 20, Detail: f.Reason})
	}
	for _, z := range r.UndersuppliedZones(3) {
		if z.AverageDemandSupply > 2 {
			out = append(out, Anomaly{Type: "zone_undersupplied", EntityID: z.Zone, Severity: z.AverageDemandSupply * 15, Detail: describeUndersupply(z)})
		}
	}

	sort.Slice(out, func(i, j int) bool { return out[i].Severity > out[j].Severity })
	if len(out) > 50 {
		out = out[:50]
	}
	return out
}

func describeNegativeProfit(t *entity.SimTrip, profit int64) string {
	return "Chuyến " + t.TripID + " (" + string(t.Kind) + ") có lợi nhuận ước tính âm (" + formatVND(profit) + " VND) sau khi trừ VAT/chi phí hạ tầng ước tính — fare quá nhỏ so với chi phí cố định giả định."
}

func describeNegativeDriverIncome(t *entity.SimTrip) string {
	return "Chuyến " + t.TripID + " có driver net âm (" + formatVND(t.DriverNetVND) + " VND) — vi phạm bất biến commission+driver_net=fare, cần kiểm tra lại logic chia commission cho chuyến này."
}

func describeSurgeLoss(t *entity.SimTrip, profit int64) string {
	return "Chuyến " + t.TripID + " có surge x" + formatFloat(t.SurgeMultiplier) + " nhưng vẫn lỗ ước tính (" + formatVND(profit) + " VND) — surge không đủ bù chi phí cố định giả định."
}

func describeUndersupply(z ZoneStat) string {
	return "Khu vực " + z.Zone + " có tỉ lệ cầu/cung trung bình " + formatFloat(z.AverageDemandSupply) + "x — thiếu tài xế nghiêm trọng."
}
