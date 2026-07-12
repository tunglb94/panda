package bi

import "fmt"

// Alert is one auto-generated business alert — PHẦN 16.
type Alert struct {
	Text     string  `json:"text"` // includes the ⚠ prefix, matching the brief's own examples verbatim
	Severity float64 `json:"severity"`
}

// livingWagePerDayVND is a simulation-design ASSUMPTION reference point for
// "driver income below a livable level" — approximated from Vietnam's 2026
// regional minimum wage (Region I ~4.96 triệu/tháng ÷ 26 công ≈ 190,000đ/
// ngày), NOT an official BRB or government figure re-derived precisely;
// used only to flag drivers whose net income is far below any reasonable
// subsistence bar, not as a compliance/legal determination.
const livingWagePerDayVND = 190_000

// ComputeBusinessAlerts is PHẦN 16 — every alert is gated on a real
// threshold computed from this run's own data (reuses driver_economy.go's
// segments, surge_analysis.go's tier counts, and stats.Bundle directly)
// rather than a separate parallel computation.
func ComputeBusinessAlerts(in Input, driverEconomy DriverEconomyReport) []Alert {
	var alerts []Alert
	add := func(sev float64, format string, args ...any) {
		if sev <= 0 {
			return
		}
		alerts = append(alerts, Alert{Text: fmt.Sprintf(format, args...), Severity: sev})
	}

	// Airport thiếu tài xế
	if airportGap := airportDemandSupplyGap(in); airportGap > 5 {
		add(airportGap*10, "⚠ Airport thiếu tài xế — tỉ lệ cầu/cung trung bình %.1fx.", airportGap)
	}

	// Surge quá nhiều
	if in.Bundle.PricingStatistics.SurgedTripPercent > 50 {
		add(in.Bundle.PricingStatistics.SurgedTripPercent, "⚠ Surge quá nhiều — %.1f%% tổng số chuyến bị surge.", in.Bundle.PricingStatistics.SurgedTripPercent)
	}

	// Voucher ROI thấp
	for _, p := range in.Bundle.PromotionROI.ByType {
		if p.TotalCostVND > 0 && p.ROI < 2 {
			add((2-p.ROI)*30, "⚠ Voucher ROI thấp — chương trình \"%s\" chỉ đạt %.1fx.", p.Type, p.ROI)
		}
	}

	// Driver fatigue cao
	for _, seg := range driverEconomy.Segments {
		if seg.DriverCount > 0 && seg.AverageFatigue > 0.6 {
			add(seg.AverageFatigue*50, "⚠ Driver fatigue cao ở nhóm %s — trung bình %.2f.", seg.Category, seg.AverageFatigue)
		}
	}

	// ETA vượt 12 phút
	if in.BI.AverageETAMinutes > 12 {
		add(in.BI.AverageETAMinutes, "⚠ ETA vượt ngưỡng 12 phút — trung bình hiện tại %.1f phút.", in.BI.AverageETAMinutes)
	}

	// Cancel tăng — no historical baseline exists in a single run, so this
	// compares against a fixed 3% reference rate (a reasonable healthy-
	// platform bar, not a measured Panda historical average).
	if in.BI.CancellationRatePercent > 3 {
		add(in.BI.CancellationRatePercent*10, "⚠ Cancel tăng — tỉ lệ huỷ hiện tại %.1f%%, vượt ngưỡng tham chiếu 3%%.", in.BI.CancellationRatePercent)
	}

	// Driver income dưới mức sống
	for _, seg := range driverEconomy.Segments {
		if seg.DriverCount > 0 && seg.NetIncomePerDayVND > 0 && seg.NetIncomePerDayVND < livingWagePerDayVND {
			add((livingWagePerDayVND-seg.NetIncomePerDayVND)/1000, "⚠ Driver income dưới mức sống — nhóm %s chỉ đạt %s VND/ngày (tham chiếu %s VND/ngày).", seg.Category, formatVNDBi(int64(seg.NetIncomePerDayVND)), formatVNDBi(livingWagePerDayVND))
		}
	}

	// Driver/Passenger retention giảm — compared against an 80% healthy bar.
	if in.Bundle.DriverAnalytics.RetentionRatePercent > 0 && in.Bundle.DriverAnalytics.RetentionRatePercent < 80 {
		add(80-in.Bundle.DriverAnalytics.RetentionRatePercent, "⚠ Driver retention giảm — hiện tại %.1f%%, dưới ngưỡng tham chiếu 80%%.", in.Bundle.DriverAnalytics.RetentionRatePercent)
	}
	if in.Bundle.RiderAnalytics.RetentionRatePercent > 0 && in.Bundle.RiderAnalytics.RetentionRatePercent < 50 {
		add(50-in.Bundle.RiderAnalytics.RetentionRatePercent, "⚠ Passenger retention giảm — hiện tại %.1f%%, dưới ngưỡng tham chiếu 50%%.", in.Bundle.RiderAnalytics.RetentionRatePercent)
	}

	return alerts
}

func airportDemandSupplyGap(in Input) float64 {
	var sum float64
	var n int
	for _, c := range in.Bundle.Heatmap.Cells {
		if c.Zone != "airport" || c.DemandCount == 0 {
			continue
		}
		supply := c.AverageSupply
		if supply < 0.1 {
			supply = 0.1
		}
		sum += float64(c.DemandCount) / supply
		n++
	}
	if n == 0 {
		return 0
	}
	return sum / float64(n)
}

func formatVNDBi(v int64) string {
	s := fmt.Sprintf("%d", v)
	neg := len(s) > 0 && s[0] == '-'
	if neg {
		s = s[1:]
	}
	var out []byte
	for i, c := range []byte(s) {
		if i > 0 && (len(s)-i)%3 == 0 {
			out = append(out, ',')
		}
		out = append(out, c)
	}
	if neg {
		return "-" + string(out)
	}
	return string(out)
}
