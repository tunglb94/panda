package audit

import (
	"fmt"
	"strings"

	"github.com/fairride/ai_simulation/insights"
	"github.com/fairride/ai_simulation/stats"
)

// RenderCEOHTML is the Executive Summary the task's final section asks
// for, rendered as one self-contained HTML page — every field listed
// verbatim in the task brief, plus the top-20/20/30 lists embedded inline
// (not just linked) so a CEO opening only this one file still sees
// everything.
func RenderCEOHTML(bi stats.BusinessIntelligence, r Report, driverRetentionPercent, riderRetentionPercent float64, findings []insights.Finding, bugs []BugFinding, recs []insights.Recommendation) string {
	var b strings.Builder
	b.WriteString(ceoHTMLHead)
	fmt.Fprintf(&b, `<h1>Panda — Executive Summary</h1>
<p class="subtitle">Full Business Validation — AI Digital Twin Simulation (Business Simulation, không phải load test)</p>
<div class="cards">
  <div class="card big"><div class="v">%s</div><div class="l">Tổng GMV (VND)</div></div>
  <div class="card big"><div class="v">%s</div><div class="l">Doanh thu nền tảng (VND)</div></div>
  <div class="card big"><div class="v">%s</div><div class="l">Lợi nhuận ước tính (VND)</div></div>
  <div class="card"><div class="v">%s</div><div class="l">Thu nhập TB tài xế/tuần (VND)</div></div>
  <div class="card"><div class="v">%s</div><div class="l">Thu nhập trung vị tài xế/tuần (VND)</div></div>
  <div class="card"><div class="v">%s</div><div class="l">Chi phí Voucher (VND)</div></div>
  <div class="card"><div class="v">%s</div><div class="l">Chi phí Promotion (VND)</div></div>
  <div class="card"><div class="v">%.2f:1</div><div class="l">Ride/Delivery Ratio</div></div>
  <div class="card"><div class="v">%.1f%%</div><div class="l">Acceptance Rate</div></div>
  <div class="card"><div class="v">%.1f%%</div><div class="l">Cancel Rate</div></div>
  <div class="card"><div class="v">%.1f phút</div><div class="l">ETA trung bình</div></div>
  <div class="card"><div class="v">%s</div><div class="l">ROI (blended, promotion+voucher)</div></div>
  <div class="card"><div class="v">%.1f%%</div><div class="l">Driver Retention</div></div>
  <div class="card"><div class="v">%.1f%%</div><div class="l">Passenger Retention</div></div>
</div>`,
		formatVND(bi.GMVVND), formatVND(bi.PlatformRevenueVND), formatVND(bi.ProfitVND),
		formatVND(int64(r.AverageDriverIncomeWeekVND)), formatVND(int64(r.MedianDriverIncomeWeekVND)),
		formatVND(bi.VoucherCostVND), formatVND(bi.PromotionCostVND),
		r.RideToDeliveryRatio, bi.AcceptanceRatePercent, bi.CancellationRatePercent, bi.AverageETAMinutes,
		formatBlendedROI(r.PromotionROI),
		driverRetentionPercent, riderRetentionPercent,
	)

	b.WriteString(`<div class="section"><h2>Top 20 Business Insights</h2><ol>`)
	for _, f := range findings {
		fmt.Fprintf(&b, "<li>%s</li>", htmlEscape(f.Text))
	}
	b.WriteString(`</ol></div>`)

	b.WriteString(`<div class="section"><h2>Top 20 Critical Bugs</h2>`)
	if len(bugs) == 0 {
		b.WriteString(`<p class="ok">Không phát hiện critical bug nào trong lần chạy này.</p>`)
	} else {
		b.WriteString(`<ol>`)
		for _, bug := range bugs {
			fmt.Fprintf(&b, "<li><strong>%s</strong> — %s <em>(%s)</em></li>", htmlEscape(bug.Title), htmlEscape(bug.Cause), htmlEscape(bug.File))
		}
		b.WriteString(`</ol>`)
	}
	b.WriteString(`</div>`)

	b.WriteString(`<div class="section"><h2>Top 30 Đề xuất tối ưu</h2><ol>`)
	for _, rec := range recs {
		fmt.Fprintf(&b, "<li>%s <span class=\"badge\">%s</span></li>", htmlEscape(rec.Text), rec.Priority)
	}
	b.WriteString(`</ol></div>`)

	b.WriteString(`</body></html>`)
	return b.String()
}

func formatBlendedROI(entries []stats.PromotionROIEntry) string {
	var totalGMV, totalCost int64
	for _, e := range entries {
		totalGMV += e.GMVGeneratedVND
		totalCost += e.TotalCostVND
	}
	if totalCost == 0 {
		return "N/A (chưa có chi phí voucher/promotion)"
	}
	return fmt.Sprintf("%.2fx", float64(totalGMV-totalCost)/float64(totalCost))
}

const ceoHTMLHead = `<!doctype html>
<html><head><meta charset="utf-8"><title>Panda — Executive Summary</title>
<style>
body{font-family:system-ui,sans-serif;background:#0b0f0d;color:#e5e7eb;margin:0;padding:24px;max-width:1100px}
h1{color:#1A8C4E;margin-bottom:4px} .subtitle{color:#9CA3AF;margin-top:0;margin-bottom:24px}
.cards{display:grid;grid-template-columns:repeat(auto-fit,minmax(210px,1fr));gap:12px;margin-bottom:28px}
.card{background:#141a17;border-radius:12px;padding:16px 20px}
.card.big{background:#16241c;border:1px solid #1A8C4E}
.card .v{font-size:22px;font-weight:700;color:#1A8C4E} .card.big .v{font-size:26px}
.card .l{font-size:12px;color:#9CA3AF;margin-top:4px}
.section{background:#141a17;border-radius:12px;padding:20px 24px;margin-bottom:20px}
.section h2{color:#1A8C4E;margin-top:0}
.section ol{padding-left:22px} .section li{margin:6px 0;color:#e5e7eb}
.badge{display:inline-block;font-size:11px;font-weight:700;text-transform:uppercase;background:rgba(37,99,235,0.25);color:#93C5FD;padding:2px 8px;border-radius:99px;margin-left:6px}
.ok{color:#3FCB85}
</style></head><body>
`
