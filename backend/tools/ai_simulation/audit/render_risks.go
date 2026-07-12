package audit

import (
	"fmt"
	"sort"
	"strings"
)

// Risk is one entry in top_20_business_risks.md — distinct from
// insights.Finding (a neutral observation) and from BugFinding (a code
// defect): a Risk is "this could hurt the business if left unaddressed",
// always traceable to a concrete number this audit computed.
type Risk struct {
	Text     string
	Severity float64
}

// ComputeTop20Risks turns the Report's flagged checks into a ranked risk
// list — every entry cites a real number already in Report.
func ComputeTop20Risks(r Report) []Risk {
	var risks []Risk
	add := func(sev float64, format string, args ...any) {
		if sev <= 0 {
			return
		}
		risks = append(risks, Risk{Text: fmt.Sprintf(format, args...), Severity: sev})
	}

	if r.RevenueLeakPercent > 0.1 {
		add(r.RevenueLeakPercent*20, "Revenue Leak: GMV lệch %.3f%% (%s VND) so với tổng đã phân bổ cho driver/platform/voucher/promotion.", r.RevenueLeakPercent, formatVND(r.RevenueLeakVND))
	}
	if r.NegativeProfitTripCount > 0 {
		add(float64(r.NegativeProfitTripCount)*3, "%d chuyến có lợi nhuận ước tính âm sau chi phí hạ tầng/VAT giả định — xem top_50_anomalies.json để có danh sách đầy đủ.", r.NegativeProfitTripCount)
	}
	if r.NegativeDriverIncomeTripCount > 0 {
		add(float64(r.NegativeDriverIncomeTripCount)*50, "%d chuyến có driver net âm — vi phạm bất biến tài chính cơ bản, cần điều tra ngay.", r.NegativeDriverIncomeTripCount)
	}
	if r.VoucherUnusedPercent > 50 {
		add(r.VoucherUnusedPercent, "%.1f%% voucher đã phát hành chưa từng được sử dụng (%d/%d) — ngân sách khuyến mãi có thể đang bị khoá không hiệu quả.", r.VoucherUnusedPercent, r.VoucherIssuedCount-r.VoucherUsedCount, r.VoucherIssuedCount)
	}
	if r.SurgeCausingLossTripCount > 0 {
		add(float64(r.SurgeCausingLossTripCount)*10, "%d chuyến có surge nhưng vẫn lỗ ước tính — surge hiện tại không đủ bù chi phí vận hành giả định cho các chuyến giá trị thấp.", r.SurgeCausingLossTripCount)
	}
	if n := len(r.DriversOnline12hPlus); n > 0 {
		add(float64(n)*8, "%d tài xế từng online liên tục >=12h trong một ca — rủi ro an toàn/pháp lý nếu xảy ra tai nạn.", n)
	}
	if n := len(r.DriversOnlineZeroTrips); n > 0 {
		add(float64(n)*4, "%d tài xế online nhưng không có chuyến nào hoàn tất trong lần chạy này — có thể phản ánh mất cân bằng cung/cầu ở khu vực họ hoạt động.", n)
	}
	if n := len(r.DriversHighIncomeOutliers); n > 0 {
		add(float64(n)*6, "%d tài xế có thu nhập/tuần vượt ngưỡng thống kê bất thường (mean+3σ) — nên kiểm tra xem có phải lỗi tính toán hay chỉ là tài xế hiệu suất cao thật.", n)
	}
	if n := len(r.RidersVoucherSpam); n > 0 {
		add(float64(n)*7, "%d khách hàng redeem voucher nhiều bất thường (>5 lần) trong lần chạy này.", n)
	}
	for _, z := range r.UndersuppliedZones(3) {
		if z.AverageDemandSupply > 3 {
			add(z.AverageDemandSupply*5, "Khu vực %s thiếu tài xế nghiêm trọng (cầu/cung trung bình %.1fx) — rủi ro ETA cao, khách bỏ đi.", zoneLabel(z.Zone), z.AverageDemandSupply)
		}
	}
	if best, ok := r.HighestETAZone(); ok && best.AverageETAMinutes > 20 {
		add(best.AverageETAMinutes, "Khu vực %s có ETA trung bình cao nhất (%.1f phút) — vượt xa ngưỡng trải nghiệm tốt.", zoneLabel(best.Zone), best.AverageETAMinutes)
	}
	if r.OffPeakProfit.AverageProfitVND < r.PeakHourProfit.AverageProfitVND/2 && r.OffPeakProfit.TripCount > 0 {
		add(20, "Lợi nhuận trung bình/chuyến giờ thấp điểm (%s VND) thấp hơn đáng kể so với giờ cao điểm (%s VND).", formatVND(int64(r.OffPeakProfit.AverageProfitVND)), formatVND(int64(r.PeakHourProfit.AverageProfitVND)))
	}

	sort.Slice(risks, func(i, j int) bool { return risks[i].Severity > risks[j].Severity })
	if len(risks) > 20 {
		risks = risks[:20]
	}
	return risks
}

func RenderTop20RisksMarkdown(risks []Risk) string {
	var b strings.Builder
	b.WriteString("# Panda Business Validation — Top 20 Business Risks\n\n")
	if len(risks) == 0 {
		b.WriteString("_Không phát hiện rủi ro nào vượt ngưỡng đáng chú ý trong lần chạy này._\n")
		return b.String()
	}
	if len(risks) < 20 {
		fmt.Fprintf(&b, "_Lưu ý: chỉ %d rủi ro đủ điều kiện nổi bật trong lần chạy này (ngưỡng tối đa 20)._\n\n", len(risks))
	}
	for i, risk := range risks {
		fmt.Fprintf(&b, "%d. %s\n", i+1, risk.Text)
	}
	b.WriteString("\n---\n_Mỗi rủi ro được tính trực tiếp từ dữ liệu mô phỏng thật (xem business_audit.md và top_50_anomalies.json để có chi tiết)._\n")
	return b.String()
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
