package insights

import (
	"fmt"
	"sort"

	"github.com/fairride/ai_simulation/domain/entity"
	"github.com/fairride/ai_simulation/stats"
)

// Recommendation is one PHẦN 10 business suggestion — the numbers quoted in
// Text are always real (from the same Compute* helpers findings.go uses);
// ExpectedImpact/Risk/Priority are the qualitative business judgment a
// recommendation inherently requires (unlike a Finding, which is pure data),
// each fixed to the rule that generated it, not invented per-run.
type Recommendation struct {
	Text           string
	ExpectedImpact string
	Risk           string
	Priority       string // "High" | "Medium" | "Low"
	Signal         float64
}

// ComputeRecommendations mirrors ComputeFindings' structure: ~28 candidate
// rules, each gated by a real condition on simulation output, ranked by
// Signal, top 30 kept.
func ComputeRecommendations(trips []*entity.SimTrip, bundle stats.Bundle, bi stats.BusinessIntelligence) []Recommendation {
	var r []Recommendation
	add := func(signal float64, priority, impact, risk, format string, args ...any) {
		if signal <= 0 {
			return
		}
		r = append(r, Recommendation{
			Text: fmt.Sprintf(format, args...), ExpectedImpact: impact, Risk: risk, Priority: priority, Signal: signal,
		})
	}

	zoneDemand, zoneSupplyGap, peakSurgeHour, peakSurgeValue, worstCancelZone, worstCancelCount := analyzeHeatmap(bundle.Heatmap)
	bestROI, worstROI := analyzePromotionROI(bundle.PromotionROI)
	bestCat, worstCat := analyzePricingCategories(bundle.PricingAnalytics)

	if zoneSupplyGap.zone != "" && zoneSupplyGap.ratio > 1.5 {
		add(zoneSupplyGap.ratio*10, "High", "Giảm ETA và tăng acceptance rate tại khu vực thiếu cung", "Chi phí incentive tăng ngắn hạn",
			"Tăng incentive/thưởng theo giờ cho tài xế hoạt động tại %s — tỉ lệ cầu/cung trung bình đang ở mức %.1fx.", zoneLabel(zoneSupplyGap.zone), zoneSupplyGap.ratio)
	}
	if zoneDemand.zone != "" {
		add(float64(zoneDemand.count)/2, "Medium", "Tối đa hoá GMV tại khu vực có nhu cầu cao nhất", "Có thể gây mất cân bằng phân bổ tài xế ở khu vực khác",
			"Ưu tiên phân bổ tài xế và khuyến mãi mục tiêu tại %s — khu vực có nhu cầu cao nhất (%d yêu cầu).", zoneLabel(zoneDemand.zone), zoneDemand.count)
	}
	if worstCancelZone != "" && worstCancelCount > 0 {
		add(float64(worstCancelCount)*3, "High", "Giảm cancellation rate", "Cần điều tra vận hành cụ thể trước khi hành động",
			"Điều tra nguyên nhân vận hành tại %s — khu vực có số chuyến huỷ sau khi đã ghép tài xế cao nhất (%d chuyến).", zoneLabel(worstCancelZone), worstCancelCount)
	}
	if peakSurgeValue > 1.3 {
		add(peakSurgeValue*15, "Medium", "Giảm surge trung bình cho khách, tăng thu nhập tài xế trước giờ cao điểm", "Cần ngân sách incentive bổ sung",
			"Khuyến khích tài xế online sớm trước khung giờ %02d:00 (surge trung bình %.2fx) bằng thưởng cố định thay vì để khách chịu surge toàn bộ.", peakSurgeHour, peakSurgeValue)
	}

	// Only recommend cutting the worst-ROI promotion when it's genuinely
	// distinct from the best-performing one (a single redeemed promotion
	// type is simultaneously "best" and "worst" by definition — recommending
	// both expanding and cutting the same program would be self-contradictory).
	if worstROI.Type != "" && (worstROI.Type != bestROI.Type || worstROI.ROI < 1) {
		signal := 40.0
		priority := "Medium"
		if worstROI.ROI < 1 {
			signal, priority = 90, "High"
		}
		add(signal, priority, "Giảm chi phí voucher/promotion, tăng lợi nhuận ròng", "Có thể làm giảm số chuyến nếu cắt giảm quá mạnh",
			"Xem xét giảm ngân sách hoặc thắt chặt điều kiện áp dụng cho \"%s\" — ROI thấp nhất trong các chương trình đã redeem (%.1fx).", promoLabel(worstROI.Type), worstROI.ROI)
	}
	if bestROI.Type != "" && bestROI.ROI > 3 {
		add(bestROI.ROI*8, "Medium", "Tăng GMV và số chuyến với chi phí khuyến mãi hiệu quả", "Ngân sách marketing tăng",
			"Mở rộng ngân sách cho \"%s\" — ROI cao nhất (%.1fx), %d lượt redeem.", promoLabel(bestROI.Type), bestROI.ROI, bestROI.RedeemedCount)
	}

	if bundle.DriverAnalytics.RetentionRatePercent > 0 && bundle.DriverAnalytics.RetentionRatePercent < 70 {
		add(100-bundle.DriverAnalytics.RetentionRatePercent, "High", "Tăng driver retention, giảm chi phí tuyển tài xế mới", "Cần ngân sách chương trình giữ chân",
			"Xây dựng chương trình giữ chân tài xế — retention hiện tại chỉ đạt %.1f%% số ngày mô phỏng.", bundle.DriverAnalytics.RetentionRatePercent)
	}
	if bundle.RiderAnalytics.RetentionRatePercent > 0 && bundle.RiderAnalytics.RetentionRatePercent < 40 {
		add(60-bundle.RiderAnalytics.RetentionRatePercent, "High", "Tăng passenger retention, tăng LTV mỗi khách hàng", "Cần ngân sách CRM/loyalty",
			"Xây dựng chương trình giữ chân khách hàng — retention hiện tại chỉ đạt %.1f%% số ngày mô phỏng.", bundle.RiderAnalytics.RetentionRatePercent)
	}

	if bronze, total := bundle.DriverStatistics.AccountTypeCounts["bronze"], bundle.DriverStatistics.TotalDrivers; total > 0 {
		pct := 100 * float64(bronze) / float64(total)
		if pct > 40 {
			add(pct, "Medium", "Tăng driver income trung bình, giảm rủi ro nghỉ việc ở nhóm mới", "Giảm doanh thu hoa hồng ngắn hạn nếu hạ ngưỡng lên hạng",
				"Xem xét lộ trình lên hạng (tier) rõ ràng hơn cho tài xế mới — %.1f%% tài xế vẫn ở hạng Bronze (hoa hồng 20%%, cao nhất theo BRB §7.1).", pct)
		}
	}

	if bi.AcceptanceRatePercent > 0 && bi.AcceptanceRatePercent < 80 {
		add(90-bi.AcceptanceRatePercent, "High", "Tăng acceptance rate, giảm thời gian tìm tài xế cho khách", "Cần điều tra fare/khoảng cách cụ thể",
			"Điều tra nguyên nhân tài xế từ chối chuyến (fare thấp, khoảng cách xa, khu vực kém an toàn) — acceptance rate hiện tại %.1f%%.", bi.AcceptanceRatePercent)
	}
	if bi.CancellationRatePercent > 3 {
		add(bi.CancellationRatePercent*10, "High", "Giảm cancellation rate, tăng trải nghiệm khách hàng", "Cần xác định nguyên nhân gốc trước khi hành động",
			"Giảm cancellation rate (hiện tại %.1f%% tổng yêu cầu) bằng cách cải thiện độ chính xác ETA hiển thị cho khách trước khi đặt.", bi.CancellationRatePercent)
	}
	if bi.AverageETAMinutes > 15 {
		add(bi.AverageETAMinutes*2, "Medium", "Giảm ETA trung bình, tăng trải nghiệm khách hàng", "Cần tăng mật độ tài xế, có thể tăng chi phí incentive",
			"Tăng mật độ tài xế tại các khu vực/khung giờ có ETA cao — ETA trung bình hiện tại %.1f phút.", bi.AverageETAMinutes)
	}
	if bi.PlatformMarginPercent > 0 && bi.PlatformMarginPercent < 15 {
		add(20-bi.PlatformMarginPercent, "Medium", "Tăng platform margin", "Rủi ro giảm sức cạnh tranh giá nếu tăng hoa hồng/booking fee",
			"Rà soát cơ cấu hoa hồng/booking fee — platform margin hiện tại chỉ %.1f%% GMV.", bi.PlatformMarginPercent)
	}

	if bi.DeliveryPercent > 0 && bi.DeliveryPercent < 20 {
		add(20-bi.DeliveryPercent, "Low", "Đa dạng hoá nguồn doanh thu ngoài Ride", "Chi phí marketing bổ sung cho sản phẩm mới",
			"Đầu tư marketing cho dịch vụ Delivery — hiện chỉ chiếm %.1f%% tổng số chuyến hoàn tất dù đã có đầy đủ hạ tầng vận hành.", bi.DeliveryPercent)
	}

	if freeShare, total := bundle.RiderAnalytics.MembershipCounts["free"], bundle.RiderAnalytics.MembershipCounts["free"]+bundle.RiderAnalytics.MembershipCounts["silver"]+bundle.RiderAnalytics.MembershipCounts["gold"]+bundle.RiderAnalytics.MembershipCounts["diamond"]; total > 0 {
		pct := 100 * float64(freeShare) / float64(total)
		if pct > 60 {
			add(pct/2, "Medium", "Tăng retention và giảm price sensitivity trung bình", "Cần thiết kế ưu đãi hấp dẫn nhưng không ăn mòn margin",
				"Thiết kế ưu đãi nâng hạng thành viên — %.1f%% khách hàng vẫn ở hạng Free.", pct)
		}
	}

	if worstCat.Category != "" && worstCat.AverageProfitVND < 0 {
		add(80, "High", "Ngừng lỗ cho hạng xe đang âm lợi nhuận", "Có thể giảm lựa chọn cho khách hàng ở phân khúc đó",
			"Xem xét điều chỉnh giá hoặc tạm dừng hạng \"%s\" — lợi nhuận trung bình/chuyến đang âm (%s VND).", worstCat.Category, formatInt(int64(worstCat.AverageProfitVND)))
	} else if worstCat.Category != "" && bestCat.Category != "" && worstCat.Category != bestCat.Category {
		add(15, "Low", "Cải thiện lợi nhuận ở hạng xe yếu nhất", "Cần phân tích thêm nguyên nhân chênh lệch trước khi đổi giá",
			"Rà soát cấu trúc giá cho hạng \"%s\" — lợi nhuận trung bình/chuyến thấp hơn đáng kể so với \"%s\".", worstCat.Category, bestCat.Category)
	}

	if bundle.VoucherStatistics.UsedCount+bundle.VoucherStatistics.KeptCount > 5 {
		usedPct := 100 * float64(bundle.VoucherStatistics.UsedCount) / float64(bundle.VoucherStatistics.UsedCount+bundle.VoucherStatistics.KeptCount)
		if usedPct > 80 {
			add(usedPct-70, "Medium", "Kiểm soát chi phí voucher tốt hơn", "Có thể làm giảm hài lòng khách nếu giới hạn quá chặt",
				"Xem xét giới hạn tần suất sử dụng voucher/tuần — tỉ lệ khách chọn dùng ngay lên tới %.1f%%.", usedPct)
		} else if usedPct < 30 {
			add(40-usedPct, "Low", "Tăng tỉ lệ redeem voucher đã phát hành", "Chi phí voucher có thể tăng nếu quá dễ áp dụng",
				"Đơn giản hoá điều kiện áp dụng voucher — chỉ %.1f%% lượt được đề nghị chọn dùng ngay.", usedPct)
		}
	}

	if bundle.DeliveryStatistics.Requested > 0 && bundle.DeliveryStatistics.AveragePickupMinutes > 20 {
		add(bundle.DeliveryStatistics.AveragePickupMinutes, "Medium", "Giảm thời gian lấy hàng cho Delivery", "Cần thêm tài xế chuyên trách Delivery ở khu vực nguồn hàng",
			"Tối ưu vùng phủ tài xế cho Delivery — thời gian lấy hàng trung bình hiện tại %.1f phút.", bundle.DeliveryStatistics.AveragePickupMinutes)
	}
	if bundle.DeliveryStatistics.Requested > 0 && bundle.DeliveryStatistics.AverageWeightKg > 10 {
		add(30, "Low", "Tăng doanh thu Delivery cho đơn hàng nặng/cồng kềnh", "Có thể làm giảm nhu cầu ở phân khúc hàng nặng nếu phụ phí quá cao",
			"Xem xét bổ sung phụ phí quá khổ/quá tải cho Delivery — trọng lượng trung bình mỗi đơn hiện tại đã đạt %.1fkg.", bundle.DeliveryStatistics.AverageWeightKg)
	}

	if bestROI.Type != "" && bestROI.RepeatRatePercent < 50 && bestROI.RedeemedCount > 3 {
		add(50-bestROI.RepeatRatePercent, "Low", "Tăng tỉ lệ khách quay lại sau khuyến mãi", "Cần thêm ngân sách cho ưu đãi lần 2",
			"Thiết kế thêm ưu đãi lần-2 cho khách vừa dùng \"%s\" — tỉ lệ quay lại hiện chỉ %.1f%%.", promoLabel(bestROI.Type), bestROI.RepeatRatePercent)
	}

	if bundle.DriverAnalytics.AverageOnlineHours > 0 && bundle.DriverAnalytics.AverageOnlineHours < 4 {
		add(20, "Medium", "Tăng thu nhập trung bình mỗi tài xế", "Có thể không hiệu quả nếu nhu cầu thực tế thấp ở khung giờ đó",
			"Khuyến khích tài xế online nhiều hơn qua thưởng theo khung giờ cao điểm — trung bình hiện tại chỉ %.1f giờ/tài xế trong suốt thời gian mô phỏng.", bundle.DriverAnalytics.AverageOnlineHours)
	}
	if highFatiguePct := highFatigueSharePercent(bundle.DriverAnalytics); highFatiguePct > 15 {
		add(highFatiguePct*2, "High", "Giảm rủi ro an toàn và burn-out tài xế", "Có thể làm giảm giờ online, giảm cung ngắn hạn",
			"Theo dõi sát chỉ số fatigue — %.1f%% tài xế đang có mức fatigue từ 0.7 trở lên tại thời điểm cuối mô phỏng.", highFatiguePct)
	}

	if bi.DeliveryRevenueVND > 0 && bi.RideRevenueVND > 0 {
		add(10, "Low", "Theo dõi tính mùa vụ để tối ưu phân bổ ngân sách marketing", "Không có",
			"Theo dõi biến động GMV theo Ride (%s VND) và Delivery (%s VND) theo thời gian thực để phát hiện tính mùa vụ sớm.",
			formatInt(bi.RideRevenueVND), formatInt(bi.DeliveryRevenueVND))
	}

	sort.Slice(r, func(i, j int) bool { return r[i].Signal > r[j].Signal })
	if len(r) > 30 {
		r = r[:30]
	}
	return r
}
