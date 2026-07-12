package dashboard

import (
	"fmt"
	"strings"

	"github.com/fairride/ai_simulation/stats"
)

// GenerateExecutive renders executive_dashboard.html — PHẦN 2's CEO-facing
// one-pager ("CEO chỉ mở 1 file này"): every field business_intelligence
// computes, plus 2 charts (Ride vs Delivery split, Revenue waterfall) built
// from the same real trip-derived data every other export uses.
func GenerateExecutive(bi stats.BusinessIntelligence, bundle stats.Bundle) string {
	var b strings.Builder
	b.WriteString(executiveHTMLHead)
	writeExecutiveHeader(&b, bundle)
	writeExecutiveCards(&b, bi)
	b.WriteString(`<div class="grid">`)
	writeCanvas(&b, "revenueWaterfall", "Revenue Waterfall (Bar)")
	writeCanvas(&b, "rideDeliverySplit", "Ride vs Delivery (Pie)")
	b.WriteString(`</div>`)
	writeExecutiveScripts(&b, bi)
	b.WriteString(htmlFoot)
	return b.String()
}

func writeExecutiveHeader(b *strings.Builder, bundle stats.Bundle) {
	cfg := bundle.SimulationReport.Config
	fmt.Fprintf(b, `<h1>Panda — Executive Dashboard</h1>
<p class="subtitle">%d drivers · %d riders · %d simulated day(s) · model %s</p>`,
		cfg.Drivers, cfg.Riders, cfg.Days, cfg.Model)
}

func writeExecutiveCards(b *strings.Builder, bi stats.BusinessIntelligence) {
	fmt.Fprintf(b, `<div class="cards">
  <div class="card big"><div class="v">%s</div><div class="l">GMV (VND)</div></div>
  <div class="card big"><div class="v">%s</div><div class="l">Net Revenue (VND)</div></div>
  <div class="card big"><div class="v">%s</div><div class="l">Estimated Profit (VND)</div></div>
  <div class="card"><div class="v">%s</div><div class="l">Platform Revenue (VND)</div></div>
  <div class="card"><div class="v">%s</div><div class="l">Driver Revenue (VND)</div></div>
  <div class="card"><div class="v">%.1f%%</div><div class="l">Platform Margin</div></div>
  <div class="card"><div class="v">%s</div><div class="l">Voucher Cost (VND)</div></div>
  <div class="card"><div class="v">%s</div><div class="l">Promotion Cost (VND)</div></div>
  <div class="card"><div class="v">%.1f%%</div><div class="l">Driver Retention</div></div>
  <div class="card"><div class="v">%.1f%%</div><div class="l">Passenger Retention</div></div>
  <div class="card"><div class="v">%s</div><div class="l">Ride Revenue (VND)</div></div>
  <div class="card"><div class="v">%s</div><div class="l">Delivery Revenue (VND)</div></div>
  <div class="card"><div class="v">%.1f%%</div><div class="l">Ride %%</div></div>
  <div class="card"><div class="v">%.1f%%</div><div class="l">Delivery %%</div></div>
  <div class="card"><div class="v">%s</div><div class="l">Average Fare (VND)</div></div>
  <div class="card"><div class="v">%.1f min</div><div class="l">Average ETA</div></div>
  <div class="card"><div class="v">%.1f%%</div><div class="l">Acceptance Rate</div></div>
  <div class="card"><div class="v">%.1f%%</div><div class="l">Cancellation Rate</div></div>
</div>`,
		formatVND(bi.GMVVND), formatVND(bi.NetRevenueVND), formatVND(bi.ProfitVND),
		formatVND(bi.PlatformRevenueVND), formatVND(bi.DriverRevenueVND), bi.PlatformMarginPercent,
		formatVND(bi.VoucherCostVND), formatVND(bi.PromotionCostVND),
		bi.DriverRetentionPercent, bi.PassengerRetentionPercent,
		formatVND(bi.RideRevenueVND), formatVND(bi.DeliveryRevenueVND),
		bi.RidePercent, bi.DeliveryPercent,
		formatVND(int64(bi.AverageFareVND)), bi.AverageETAMinutes,
		bi.AcceptanceRatePercent, bi.CancellationRatePercent,
	)
}

func writeExecutiveScripts(b *strings.Builder, bi stats.BusinessIntelligence) {
	fmt.Fprintf(b, `<script src="https://cdn.jsdelivr.net/npm/chart.js@4"></script>
<script>
new Chart(document.getElementById('revenueWaterfall'), {type:'bar', data:{
  labels: ['GMV', 'Platform Revenue', 'Driver Revenue', 'Voucher Cost', 'Promotion Cost', 'Net Revenue', 'Est. Profit'],
  datasets: [{data: [%d,%d,%d,%d,%d,%d,%d], backgroundColor: ['#1A8C4E','#2563EB','#F59E0B','#DC2626','#DC2626','#0F5C33','#22C55E']}]
}, options: {plugins:{legend:{display:false}}}});

new Chart(document.getElementById('rideDeliverySplit'), {type:'pie', data:{
  labels: ['Ride', 'Delivery'],
  datasets: [{data: [%.2f,%.2f], backgroundColor: ['#1A8C4E','#2563EB']}]
}});
</script>`,
		bi.GMVVND, bi.PlatformRevenueVND, bi.DriverRevenueVND, bi.VoucherCostVND, bi.PromotionCostVND, bi.NetRevenueVND, bi.ProfitVND,
		bi.RidePercent, bi.DeliveryPercent,
	)
}

const executiveHTMLHead = `<!doctype html>
<html><head><meta charset="utf-8"><title>Panda Executive Dashboard</title>
<style>
body{font-family:system-ui,sans-serif;background:#0b0f0d;color:#e5e7eb;margin:0;padding:24px}
h1{color:#1A8C4E;margin-bottom:4px} .subtitle{color:#9CA3AF;margin-top:0;margin-bottom:24px}
.cards{display:grid;grid-template-columns:repeat(auto-fit,minmax(200px,1fr));gap:12px;margin-bottom:24px}
.card{background:#141a17;border-radius:12px;padding:16px 20px}
.card.big{background:#16241c;border:1px solid #1A8C4E}
.card .v{font-size:22px;font-weight:700;color:#1A8C4E} .card.big .v{font-size:26px}
.card .l{font-size:12px;color:#9CA3AF;margin-top:4px}
.grid{display:grid;grid-template-columns:repeat(auto-fit,minmax(380px,1fr));gap:16px;margin-bottom:24px}
.chart-box{background:#141a17;border-radius:12px;padding:16px}
.chart-box h3{color:#e5e7eb;margin:0 0 8px}
</style></head><body>
`
