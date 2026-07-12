package audit

import (
	"fmt"
	"strings"

	"github.com/fairride/ai_simulation/stats"
)

// RenderBusinessAuditMarkdown walks all 20 numbered checks the task
// requires, each with the real computed number(s) — this is the
// comprehensive audit trail; validation_report.html/CEO_report.html/top_20/
// top_30/top_50 files are focused views derived from the same Report.
func RenderBusinessAuditMarkdown(r Report, validation stats.ValidationReport, assumptions []Assumption, bugs []BugFinding) string {
	var b strings.Builder
	b.WriteString("# Panda — Full Business Validation: Business Audit\n\n")
	b.WriteString("Mô phỏng hành vi (không phải load test) — mọi số liệu dưới đây tính trực tiếp từ dữ liệu mô phỏng thật, không suy đoán.\n\n")

	fmt.Fprintf(&b, "## 1. Revenue Leak\n\nGMV lệch **%.3f%%** (%s VND) so với tổng đã phân bổ cho driver + platform + voucher + promotion.%s\n\n",
		r.RevenueLeakPercent, formatVND(r.RevenueLeakVND), leakVerdict(r.RevenueLeakPercent))

	fmt.Fprintf(&b, "## 2. Chuyến có Platform Profit < 0\n\n**%d** chuyến hoàn tất có lợi nhuận ước tính âm sau VAT + chi phí hạ tầng giả định (xem `unit_economics.json` cho giả định chi phí). Ví dụ: %s\n\n",
		r.NegativeProfitTripCount, joinOrNone(r.NegativeProfitExamples))

	fmt.Fprintf(&b, "## 3. Chuyến có Driver Income < 0\n\n**%d** chuyến có driver net âm. Ví dụ: %s\n\n",
		r.NegativeDriverIncomeTripCount, joinOrNone(r.NegativeDriverIncomeExamples))

	fmt.Fprintf(&b, "## 4. Voucher phát nhiều hơn dùng bao nhiêu %%\n\nPhát hành: **%d**, đã dùng: **%d** → **%.1f%%** voucher chưa từng được redeem trong lần chạy này.\n\n",
		r.VoucherIssuedCount, r.VoucherUsedCount, r.VoucherUnusedPercent)

	b.WriteString("## 5. Promotion ROI\n\n")
	if len(r.PromotionROI) == 0 {
		b.WriteString("Không có chương trình khuyến mãi nào được redeem trong lần chạy này.\n\n")
	} else {
		for _, p := range r.PromotionROI {
			fmt.Fprintf(&b, "- **%s**: ROI %.2fx, CPA %s VND/lượt, %d lượt redeem, repeat rate %.1f%%\n", p.Type, p.ROI, formatVND(int64(p.CPAVND)), p.RedeemedCount, p.RepeatRatePercent)
		}
		b.WriteString("\n")
	}

	fmt.Fprintf(&b, "## 6. Surge làm Platform lỗ\n\n**%d** chuyến có surge (>1.0x) nhưng vẫn lỗ ước tính. Ví dụ: %s\n\n",
		r.SurgeCausingLossTripCount, joinOrNone(r.SurgeCausingLossExamples))

	fmt.Fprintf(&b, "## 7. Driver online 12h+\n\n**%d** tài xế từng online liên tục >=12h trong một ca (ngưỡng cứng của FatigueDecision).\n\n", len(r.DriversOnline12hPlus))

	fmt.Fprintf(&b, "## 8. Driver online nhưng 0 chuyến\n\n**%d** tài xế có online trong lần chạy này nhưng hoàn tất 0 chuyến.\n\n", len(r.DriversOnlineZeroTrips))

	fmt.Fprintf(&b, "## 9. Driver thu nhập cao bất thường\n\n**%d** tài xế có thu nhập/tuần vượt ngưỡng thống kê (mean + 3×độ lệch chuẩn).\n\n", len(r.DriversHighIncomeOutliers))

	fmt.Fprintf(&b, "## 10. Passenger spam voucher\n\n**%d** khách hàng redeem voucher >5 lần trong lần chạy này (ngưỡng thiết kế, xem ASSUMPTION).\n\n", len(r.RidersVoucherSpam))

	b.WriteString("## 11. Dispatch có bị thiên vị khu vực không\n\n")
	b.WriteString("**ASSUMPTION quan trọng**: Dispatch thật (`RequestDispatchUseCase`/`offerNextDriver`) chỉ ghép theo khoảng cách gần nhất, không có khái niệm \"ưu tiên khu vực\" trong thuật toán. Chênh lệch accept-rate giữa các khu vực dưới đây phản ánh **phân bố cung tài xế thực tế theo khu vực**, không phải thiên vị thuật toán.\n\n")
	b.WriteString(renderZoneAcceptTable(r.ZoneStats))
	b.WriteString("\n")

	if best, ok := r.HighestETAZone(); ok {
		fmt.Fprintf(&b, "## 12. ETA cao nhất\n\nKhu vực **%s** có ETA trung bình cao nhất: **%.1f phút**.\n\n", zoneLabel(best.Zone), best.AverageETAMinutes)
	}

	b.WriteString("## 13. Khu vực thiếu driver\n\n")
	for _, z := range r.UndersuppliedZones(3) {
		fmt.Fprintf(&b, "- %s: cầu/cung trung bình **%.1fx**\n", zoneLabel(z.Zone), z.AverageDemandSupply)
	}
	b.WriteString("\n## 14. Khu vực thừa driver\n\n")
	for _, z := range r.OversuppliedZones(3) {
		fmt.Fprintf(&b, "- %s: cầu/cung trung bình **%.2fx**\n", zoneLabel(z.Zone), z.AverageDemandSupply)
	}

	fmt.Fprintf(&b, "\n## 15. Ride vs Delivery Ratio\n\nRide: **%d**, Delivery: **%d** → tỉ lệ **%.2f:1**.\n\n", r.RideCount, r.DeliveryCount, r.RideToDeliveryRatio)

	fmt.Fprintf(&b, "## 16. Airport Profit\n\n%d chuyến liên quan Sân bay, lợi nhuận trung bình/chuyến **%s VND**, tổng **%s VND**.\n\n",
		r.AirportProfit.TripCount, formatVND(int64(r.AirportProfit.AverageProfitVND)), formatVND(r.AirportProfit.TotalProfitVND))

	fmt.Fprintf(&b, "## 17. Peak Hour Profit\n\n%d chuyến giờ cao điểm (07:00-09:00, 17:00-20:00), lợi nhuận trung bình/chuyến **%s VND**.\n\n",
		r.PeakHourProfit.TripCount, formatVND(int64(r.PeakHourProfit.AverageProfitVND)))

	fmt.Fprintf(&b, "## 18. Off Peak Profit\n\n%d chuyến giờ thấp điểm, lợi nhuận trung bình/chuyến **%s VND**.\n\n",
		r.OffPeakProfit.TripCount, formatVND(int64(r.OffPeakProfit.AverageProfitVND)))

	b.WriteString("## 19. Weather ảnh hưởng thế nào\n\n")
	for _, w := range r.WeatherImpact {
		fmt.Fprintf(&b, "- **%s**: %d chuyến, fare TB %s VND, surge TB %.2fx, cancel rate %.1f%%\n", w.Label, w.TripCount, formatVND(int64(w.AverageFareVND)), w.AverageSurge, w.CancellationPercent)
	}

	b.WriteString("\n## 20. Top 50 Anomaly\n\nXem `top_50_anomalies.json` cho danh sách đầy đủ, xếp hạng theo mức độ nghiêm trọng.\n\n")

	b.WriteString("---\n\n## Validation (đối chiếu với validation_report.json)\n\n")
	fmt.Fprintf(&b, "**Passed: %v**\n\n", validation.Passed)
	for _, w := range validation.Warnings {
		fmt.Fprintf(&b, "- [%s] **%s**: %s\n", strings.ToUpper(w.Severity), w.Check, w.Message)
	}

	b.WriteString("\n## Bugs phát hiện (không tự sửa)\n\n")
	if len(bugs) == 0 {
		b.WriteString("Không phát hiện bug mới nào trong lần audit này ngoài các bug đã được sửa ở phase trước (xem CHANGELOG).\n\n")
	} else {
		for i, bug := range bugs {
			fmt.Fprintf(&b, "### Bug %d: %s\n\n- **Nguyên nhân:** %s\n- **Ảnh hưởng:** %s\n- **Cách tái hiện:** %s\n- **File liên quan:** `%s`\n\n", i+1, bug.Title, bug.Cause, bug.Impact, bug.Reproduction, bug.File)
		}
	}

	b.WriteString("\n## ASSUMPTION (logic không hợp lý/giả định, không tự sửa)\n\n")
	for i, a := range assumptions {
		fmt.Fprintf(&b, "%d. **%s** — %s\n", i+1, a.Title, a.Detail)
	}

	return b.String()
}

func leakVerdict(pct float64) string {
	switch {
	case pct > 1.0:
		return " **VƯỢT NGƯỠNG 1% — cần điều tra ngay.**"
	case pct > 0.1:
		return " Trong ngưỡng sai số làm tròn, không đáng lo ngại."
	default:
		return " Không phát hiện thất thoát đáng kể."
	}
}

func joinOrNone(ids []string) string {
	if len(ids) == 0 {
		return "(không có)"
	}
	return "`" + strings.Join(ids, "`, `") + "`"
}

func renderZoneAcceptTable(zones []ZoneStat) string {
	var b strings.Builder
	b.WriteString("| Khu vực | Requested | Accept Rate | ETA TB | Cầu/Cung TB |\n|---|---:|---:|---:|---:|\n")
	for _, z := range zones {
		fmt.Fprintf(&b, "| %s | %d | %.1f%% | %.1f phút | %.2fx |\n", zoneLabel(z.Zone), z.Requested, z.AcceptRatePercent, z.AverageETAMinutes, z.AverageDemandSupply)
	}
	return b.String()
}
