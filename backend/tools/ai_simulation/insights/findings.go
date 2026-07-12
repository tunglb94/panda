// Package insights computes PHẦN 9/10's data-grounded findings and
// recommendations. Every Finding/Recommendation's numbers come from real
// aggregated simulation output (stats.Bundle/BusinessIntelligence) — this
// package never invents a number; the only thing an optional AI pass (see
// ai_writer.go) is allowed to change is the prose wrapped around numbers
// this package already computed.
package insights

import (
	"fmt"
	"sort"

	"github.com/fairride/ai_simulation/domain/entity"
	"github.com/fairride/ai_simulation/stats"
)

// Finding is one data-grounded observation — Signal is an internal ranking
// magnitude (not shown to the reader), used to pick the most noteworthy
// findings when more than 20 candidates fire.
type Finding struct {
	Text   string
	Signal float64
}

// ComputeFindings evaluates every candidate finding rule against real
// simulation output and returns up to 20, ranked by Signal descending. Rules
// that aren't meaningful for this run (e.g. no promotions redeemed at all)
// simply don't fire — this can return fewer than 20 on a very small/short
// run, which is reported honestly (see insights.RenderSummaryMarkdown).
func ComputeFindings(trips []*entity.SimTrip, bundle stats.Bundle, bi stats.BusinessIntelligence) []Finding {
	var f []Finding
	add := func(signal float64, format string, args ...any) {
		if signal <= 0 {
			return
		}
		f = append(f, Finding{Text: fmt.Sprintf(format, args...), Signal: signal})
	}

	zoneDemand, zoneSupplyGapRatio, peakSurgeHour, peakSurgeValue, worstCancelZone, worstCancelCount := analyzeHeatmap(bundle.Heatmap)

	if zoneDemand.zone != "" {
		add(float64(zoneDemand.count), "Khu vực %s có nhu cầu cao nhất trong toàn bộ mô phỏng, với %d yêu cầu.", zoneLabel(zoneDemand.zone), zoneDemand.count)
	}
	if zoneSupplyGapRatio.zone != "" && zoneSupplyGapRatio.ratio > 1.5 {
		add(zoneSupplyGapRatio.ratio*10, "Khu vực %s thường xuyên thiếu tài xế — tỉ lệ cầu/cung trung bình đạt %.1fx.", zoneLabel(zoneSupplyGapRatio.zone), zoneSupplyGapRatio.ratio)
	}
	if peakSurgeValue > 1.15 {
		add(peakSurgeValue*20, "Khung giờ %02d:00 có hệ số surge trung bình cao nhất trong ngày (%.2fx).", peakSurgeHour, peakSurgeValue)
	}
	if worstCancelZone != "" && worstCancelCount > 0 {
		add(float64(worstCancelCount)*2, "Khu vực %s ghi nhận số chuyến bị huỷ sau khi đã có tài xế nhận nhiều nhất (%d chuyến) — đáng điều tra nguyên nhân vận hành tại khu vực này.", zoneLabel(worstCancelZone), worstCancelCount)
	}

	bestROI, worstROI := analyzePromotionROI(bundle.PromotionROI)
	if bestROI.Type != "" {
		add(bestROI.ROI*50, "Chương trình khuyến mãi \"%s\" có ROI cao nhất (%.1fx chi phí bỏ ra), với %d lượt redeem.", promoLabel(bestROI.Type), bestROI.ROI, bestROI.RedeemedCount)
	}
	if worstROI.Type != "" && worstROI.Type != bestROI.Type {
		signal := 30.0
		if worstROI.ROI < 1 {
			signal = 80 // a promotion actually losing money vs its own cost is a much bigger flag
		}
		add(signal, "Chương trình khuyến mãi \"%s\" có ROI thấp nhất trong các chương trình đã redeem (%.1fx), CPA %s VND/lượt.", promoLabel(worstROI.Type), worstROI.ROI, formatInt(int64(worstROI.CPAVND)))
	}

	bestCat, worstCat := analyzePricingCategories(bundle.PricingAnalytics)
	if bestCat.Category != "" {
		add(bestCat.AverageProfitVND/100, "Hạng xe \"%s\" có lợi nhuận trung bình/chuyến cao nhất (%s VND), với %d chuyến hoàn tất.", bestCat.Category, formatInt(int64(bestCat.AverageProfitVND)), bestCat.TripCount)
	}
	if worstCat.Category != "" && worstCat.Category != bestCat.Category {
		add(40, "Hạng xe \"%s\" có lợi nhuận trung bình/chuyến thấp nhất trong các hạng đang vận hành (%s VND).", worstCat.Category, formatInt(int64(worstCat.AverageProfitVND)))
	}

	if bi.RidePercent > 0 || bi.DeliveryPercent > 0 {
		add(25, "Ride chiếm %.1f%% và Delivery chiếm %.1f%% tổng số chuyến hoàn tất — doanh thu tương ứng %s VND và %s VND.",
			bi.RidePercent, bi.DeliveryPercent, formatInt(bi.RideRevenueVND), formatInt(bi.DeliveryRevenueVND))
	}

	if bundle.DriverAnalytics.RetentionRatePercent > 0 {
		signal := 20.0
		if bundle.DriverAnalytics.RetentionRatePercent < 50 {
			signal = 90 // low driver retention is a critical business risk
		}
		add(signal, "Tỉ lệ giữ chân tài xế (retention) trung bình đạt %.1f%% số ngày mô phỏng.", bundle.DriverAnalytics.RetentionRatePercent)
	}
	if bundle.RiderAnalytics.RetentionRatePercent > 0 {
		signal := 20.0
		if bundle.RiderAnalytics.RetentionRatePercent < 30 {
			signal = 70
		}
		add(signal, "Tỉ lệ giữ chân khách hàng (retention) trung bình đạt %.1f%% số ngày mô phỏng.", bundle.RiderAnalytics.RetentionRatePercent)
	}

	if bi.AcceptanceRatePercent > 0 {
		signal := 15.0
		if bi.AcceptanceRatePercent < 70 {
			signal = 85
		}
		add(signal, "Tỉ lệ chấp nhận chuyến (dispatch acceptance rate) đạt %.1f%%.", bi.AcceptanceRatePercent)
	}
	if bi.CancellationRatePercent > 2 {
		add(bi.CancellationRatePercent*8, "Tỉ lệ huỷ chuyến tổng thể là %.1f%% trên tổng số yêu cầu.", bi.CancellationRatePercent)
	}

	if bundle.DeliveryStatistics.Requested > 0 {
		add(15, "Delivery: khoảng cách trung bình %.1fkm, trọng lượng trung bình %.1fkg, thời gian lấy hàng trung bình %.1f phút, thời gian giao trung bình %.1f phút.",
			bundle.DeliveryStatistics.AverageDistanceKM, bundle.DeliveryStatistics.AverageWeightKg,
			bundle.DeliveryStatistics.AveragePickupMinutes, bundle.DeliveryStatistics.AverageDeliveryMinutes)
	}

	if bronze, total := bundle.DriverStatistics.AccountTypeCounts["bronze"], bundle.DriverStatistics.TotalDrivers; total > 0 {
		pct := 100 * float64(bronze) / float64(total)
		if pct > 40 {
			add(pct, "%.1f%% tài xế vẫn ở hạng Bronze (hoa hồng cao nhất, thu nhập ròng thấp nhất theo BRB §7.1) — nhóm có rủi ro nghỉ việc cao nhất.", pct)
		}
	}

	if bundle.VoucherStatistics.UsedCount+bundle.VoucherStatistics.KeptCount > 0 {
		usedPct := 100 * float64(bundle.VoucherStatistics.UsedCount) / float64(bundle.VoucherStatistics.UsedCount+bundle.VoucherStatistics.KeptCount)
		add(20, "Khi được đề nghị dùng voucher, khách hàng chọn \"dùng ngay\" %.1f%% số lần, còn lại chọn giữ voucher cho chuyến sau.", usedPct)
	}

	if bi.PlatformMarginPercent > 0 {
		add(18, "Platform Margin (doanh thu nền tảng / GMV) đạt %.1f%%.", bi.PlatformMarginPercent)
	}
	if bi.AverageETAMinutes > 15 {
		add(bi.AverageETAMinutes, "Thời gian chờ trung bình (ETA) là %.1f phút — cao hơn ngưỡng 15 phút thường được xem là trải nghiệm tốt.", bi.AverageETAMinutes)
	}

	if freeShare, total := bundle.RiderAnalytics.MembershipCounts["free"], bundle.RiderAnalytics.MembershipCounts["free"]+bundle.RiderAnalytics.MembershipCounts["silver"]+bundle.RiderAnalytics.MembershipCounts["gold"]+bundle.RiderAnalytics.MembershipCounts["diamond"]; total > 0 {
		pct := 100 * float64(freeShare) / float64(total)
		if pct > 60 {
			add(pct/3, "%.1f%% khách hàng vẫn ở hạng thành viên Free — dư địa lớn để nâng cấp membership.", pct)
		}
	}

	if bundle.DriverAnalytics.AverageOnlineHours > 0 {
		add(12, "Trung bình mỗi tài xế online %.1f giờ trong toàn bộ thời gian mô phỏng.", bundle.DriverAnalytics.AverageOnlineHours)
	}

	sort.Slice(f, func(i, j int) bool { return f[i].Signal > f[j].Signal })
	if len(f) > 20 {
		f = f[:20]
	}
	return f
}

type zoneCount struct {
	zone  string
	count int
}

type zoneRatio struct {
	zone  string
	ratio float64
}

// analyzeHeatmap scans every heatmap cell once for the 4 zone/hour-level
// findings that all share the same underlying data.
func analyzeHeatmap(h stats.Heatmap) (topDemand zoneCount, worstSupplyGap zoneRatio, peakSurgeHour int, peakSurgeValue float64, worstCancelZone string, worstCancelCount int) {
	demandByZone := map[string]int{}
	ratioSumByZone := map[string]float64{}
	ratioNByZone := map[string]int{}
	surgeByHour := map[int]float64{}
	surgeNByHour := map[int]int{}
	cancelByZone := map[string]int{}

	for _, c := range h.Cells {
		demandByZone[c.Zone] += c.Count
		cancelByZone[c.Zone] += c.CancelledCount
		if c.DemandCount > 0 {
			supply := c.AverageSupply
			if supply < 0.1 {
				supply = 0.1 // avoid a divide-by-near-zero blowing up the ratio
			}
			ratioSumByZone[c.Zone] += float64(c.DemandCount) / supply
			ratioNByZone[c.Zone]++
		}
		if c.AverageSurgeMultiplier > 0 {
			surgeByHour[c.Hour] += c.AverageSurgeMultiplier
			surgeNByHour[c.Hour]++
		}
	}

	for zone, count := range demandByZone {
		if count > topDemand.count {
			topDemand = zoneCount{zone: zone, count: count}
		}
		if cancelByZone[zone] > worstCancelCount {
			worstCancelCount, worstCancelZone = cancelByZone[zone], zone
		}
	}
	for zone, sum := range ratioSumByZone {
		avgRatio := sum / float64(ratioNByZone[zone])
		if avgRatio > worstSupplyGap.ratio {
			worstSupplyGap = zoneRatio{zone: zone, ratio: avgRatio}
		}
	}
	for hour, sum := range surgeByHour {
		avgSurge := sum / float64(surgeNByHour[hour])
		if avgSurge > peakSurgeValue {
			peakSurgeValue, peakSurgeHour = avgSurge, hour
		}
	}
	return
}

func analyzePromotionROI(p stats.PromotionROI) (best, worst stats.PromotionROIEntry) {
	for _, e := range p.ByType {
		if e.TotalCostVND <= 0 {
			continue
		}
		if best.Type == "" || e.ROI > best.ROI {
			best = e
		}
		if worst.Type == "" || e.ROI < worst.ROI {
			worst = e
		}
	}
	return
}

func analyzePricingCategories(p stats.PricingAnalytics) (best, worst stats.PricingCategoryStat) {
	for _, c := range p.Categories {
		if c.TripCount == 0 {
			continue
		}
		if best.Category == "" || c.AverageProfitVND > best.AverageProfitVND {
			best = c
		}
		if worst.Category == "" || c.AverageProfitVND < worst.AverageProfitVND {
			worst = c
		}
	}
	return
}

func zoneLabel(z string) string {
	labels := map[string]string{
		"cbd": "Trung tâm (CBD)", "residential": "Khu dân cư", "industrial": "Khu công nghiệp (KCN)",
		"airport": "Sân bay", "bus_station": "Bến xe", "hospital": "Bệnh viện",
		"school": "Trường học", "entertainment": "Khu vui chơi",
	}
	if l, ok := labels[z]; ok {
		return l
	}
	return z
}

func promoLabel(t string) string {
	labels := map[string]string{
		"first_ride": "First Ride", "birthday": "Birthday", "weekend": "Weekend", "manual_coupon": "Manual Coupon",
	}
	if l, ok := labels[t]; ok {
		return l
	}
	return t
}

// highFatigueSharePercent sums the FatigueDistribution histogram's buckets
// at 0.7 and above (buckets are fixed 0.1-wide starting at 0, so index 7 is
// "0.7-0.8" — see stats.buildLinearHistogram) — reuses the histogram
// BuildDriverAnalytics already computed rather than re-deriving fatigue
// from DriverAgent state (insights has no access to that; it only sees
// aggregated stats.Bundle).
func highFatigueSharePercent(da stats.DriverAnalytics) float64 {
	if len(da.FatigueDistribution) < 8 {
		return 0
	}
	var high, total int
	for i, b := range da.FatigueDistribution {
		total += b.Count
		if i >= 7 {
			high += b.Count
		}
	}
	if total == 0 {
		return 0
	}
	return 100 * float64(high) / float64(total)
}

func formatInt(v int64) string {
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
