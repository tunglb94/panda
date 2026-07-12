// Package dashboard renders one self-contained dashboard.html for a
// simulation run — Chart.js loaded from CDN, all data embedded inline as
// JSON, no local server required (open the file directly in a browser).
package dashboard

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/fairride/ai_simulation/domain/entity"
	"github.com/fairride/ai_simulation/stats"
)

// Generate returns the full dashboard.html document for bundle.
func Generate(bundle stats.Bundle) string {
	hourly := hourlyTripCounts(bundle.Heatmap)

	var b strings.Builder
	b.WriteString(htmlHead)
	writeSummaryCards(&b, bundle)
	b.WriteString(`<div class="grid">`)
	writeCanvas(&b, "dispatchPie", "Dispatch Outcomes (Pie)")
	writeCanvas(&b, "driverTierBar", "Driver Tiers (Bar)")
	writeCanvas(&b, "promotionCostBar", "Promotion Cost by Type (Bar)")
	writeCanvas(&b, "hourlyLine", "Trips per Hour (Line / Timeline)")
	writeCanvas(&b, "fareHistogram", "Fare Distribution (Histogram)")
	b.WriteString(`</div>`)
	writeHeatmap(&b, bundle.Heatmap)
	writeScripts(&b, bundle, hourly)
	b.WriteString(htmlFoot)
	return b.String()
}

func hourlyTripCounts(h stats.Heatmap) [24]int {
	var out [24]int
	for _, c := range h.Cells {
		if c.Hour >= 0 && c.Hour < 24 {
			out[c.Hour] += c.Count
		}
	}
	return out
}

func writeSummaryCards(b *strings.Builder, bundle stats.Bundle) {
	r := bundle.SimulationReport
	fmt.Fprintf(b, `<div class="cards">
  <div class="card"><div class="v">%d</div><div class="l">Requested</div></div>
  <div class="card"><div class="v">%d</div><div class="l">Completed</div></div>
  <div class="card"><div class="v">%d</div><div class="l">Rejected</div></div>
  <div class="card"><div class="v">%d</div><div class="l">Cancelled</div></div>
  <div class="card"><div class="v">%s</div><div class="l">Platform Revenue (VND)</div></div>
  <div class="card"><div class="v">%s</div><div class="l">Driver Revenue (VND)</div></div>
  <div class="card"><div class="v">%s</div><div class="l">Passenger Saving (VND)</div></div>
  <div class="card"><div class="v">%s</div><div class="l">GMV (VND)</div></div>
</div>`,
		r.Dispatch.Requested, r.Dispatch.Completed, r.Dispatch.Rejected, r.Dispatch.Cancelled,
		formatVND(r.Financial.PlatformRevenueVND), formatVND(r.Financial.DriverRevenueVND),
		formatVND(r.Financial.PassengerSavingVND), formatVND(r.Financial.GMVVND),
	)
}

func formatVND(v int64) string {
	s := fmt.Sprintf("%d", v)
	neg := strings.HasPrefix(s, "-")
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

func writeCanvas(b *strings.Builder, id, title string) {
	fmt.Fprintf(b, `<div class="chart-box"><h3>%s</h3><canvas id="%s"></canvas></div>`, title, id)
}

func writeHeatmap(b *strings.Builder, h stats.Heatmap) {
	max := 0
	byZoneHour := map[string]int{}
	for _, c := range h.Cells {
		byZoneHour[c.Zone+"|"+fmt.Sprint(c.Hour)] = c.Count
		if c.Count > max {
			max = c.Count
		}
	}
	if max == 0 {
		max = 1
	}

	b.WriteString(`<div class="chart-box"><h3>Zone x Hour Demand (Heatmap)</h3><table class="heatmap"><thead><tr><th></th>`)
	for hh := 0; hh < 24; hh++ {
		fmt.Fprintf(b, "<th>%d</th>", hh)
	}
	b.WriteString("</tr></thead><tbody>")
	for _, z := range entity.AllZoneTypes() {
		fmt.Fprintf(b, "<tr><th>%s</th>", z)
		for hh := 0; hh < 24; hh++ {
			count := byZoneHour[string(z)+"|"+fmt.Sprint(hh)]
			intensity := float64(count) / float64(max)
			fmt.Fprintf(b, `<td style="background-color: rgba(26,140,78,%.2f)" title="%s %02d:00 - %d trips">%d</td>`, intensity, z, hh, count, count)
		}
		b.WriteString("</tr>")
	}
	b.WriteString("</tbody></table></div>")
}

func writeScripts(b *strings.Builder, bundle stats.Bundle, hourly [24]int) {
	dispatchLabels := []string{"Completed", "Rejected", "Cancelled"}
	dispatchData := []int{bundle.DispatchStatistics.Completed, bundle.DispatchStatistics.Rejected, bundle.DispatchStatistics.Cancelled}

	tierLabels, tierData := mapToSeries(bundle.DriverStatistics.AccountTypeCounts)
	promoLabels, promoData := mapInt64ToSeries(bundle.PromotionStatistics.CostByType)

	histLabels := make([]string, len(bundle.PricingStatistics.FareHistogram))
	histData := make([]int, len(bundle.PricingStatistics.FareHistogram))
	for i, hb := range bundle.PricingStatistics.FareHistogram {
		histLabels[i] = hb.RangeLabel
		histData[i] = hb.Count
	}

	fmt.Fprintf(b, `<script src="https://cdn.jsdelivr.net/npm/chart.js@4"></script>
<script>
const dispatchLabels = %s, dispatchData = %s;
const tierLabels = %s, tierData = %s;
const promoLabels = %s, promoData = %s;
const histLabels = %s, histData = %s;
const hourlyData = %s;

new Chart(document.getElementById('dispatchPie'), {type:'pie', data:{labels:dispatchLabels, datasets:[{data:dispatchData, backgroundColor:['#1A8C4E','#DC2626','#F59E0B']}]}});
new Chart(document.getElementById('driverTierBar'), {type:'bar', data:{labels:tierLabels, datasets:[{label:'Drivers', data:tierData, backgroundColor:'#1A8C4E'}]}});
new Chart(document.getElementById('promotionCostBar'), {type:'bar', data:{labels:promoLabels, datasets:[{label:'Cost (VND)', data:promoData, backgroundColor:'#2563EB'}]}});
new Chart(document.getElementById('hourlyLine'), {type:'line', data:{labels:[...Array(24).keys()], datasets:[{label:'Trips', data:hourlyData, borderColor:'#1A8C4E', fill:false}]}});
new Chart(document.getElementById('fareHistogram'), {type:'bar', data:{labels:histLabels, datasets:[{label:'Trips', data:histData, backgroundColor:'#F59E0B'}]}});
</script>`,
		jsArr(dispatchLabels), jsIntArr(dispatchData),
		jsArr(tierLabels), jsIntArr(tierData),
		jsArr(promoLabels), jsInt64Arr(promoData),
		jsArr(histLabels), jsIntArr(histData),
		jsIntArr(hourly[:]),
	)
}

func mapToSeries(m map[string]int) ([]string, []int) {
	var labels []string
	var data []int
	for k, v := range m {
		labels = append(labels, k)
		data = append(data, v)
	}
	return labels, data
}

func mapInt64ToSeries(m map[string]int64) ([]string, []int64) {
	var labels []string
	var data []int64
	for k, v := range m {
		labels = append(labels, k)
		data = append(data, v)
	}
	return labels, data
}

// jsArr/jsIntArr/jsInt64Arr always emit a JS array literal, never `null` —
// encoding/json marshals a nil slice as `null`, and a `data: null` dataset
// throws inside Chart.js, which would abort every later `new Chart(...)`
// call in the same <script> block (e.g. an empty promotion-cost map before
// any promotions have fired).
func jsArr(v []string) string {
	if v == nil {
		v = []string{}
	}
	b, _ := json.Marshal(v)
	return string(b)
}

func jsIntArr(v []int) string {
	if v == nil {
		v = []int{}
	}
	b, _ := json.Marshal(v)
	return string(b)
}

func jsInt64Arr(v []int64) string {
	if v == nil {
		v = []int64{}
	}
	b, _ := json.Marshal(v)
	return string(b)
}

const htmlHead = `<!doctype html>
<html><head><meta charset="utf-8"><title>Panda AI Digital Twin Simulation</title>
<style>
body{font-family:system-ui,sans-serif;background:#0b0f0d;color:#e5e7eb;margin:0;padding:24px}
h1{color:#1A8C4E} h3{color:#e5e7eb;margin:0 0 8px}
.cards{display:flex;flex-wrap:wrap;gap:12px;margin-bottom:24px}
.card{background:#141a17;border-radius:12px;padding:16px 20px;min-width:160px}
.card .v{font-size:22px;font-weight:700;color:#1A8C4E} .card .l{font-size:12px;color:#9CA3AF}
.grid{display:grid;grid-template-columns:repeat(auto-fit,minmax(380px,1fr));gap:16px;margin-bottom:24px}
.chart-box{background:#141a17;border-radius:12px;padding:16px}
table.heatmap{border-collapse:collapse;width:100%;font-size:11px}
table.heatmap th,table.heatmap td{padding:4px 6px;text-align:center;color:#e5e7eb}
table.heatmap th{color:#9CA3AF;font-weight:500}
</style></head><body>
<h1>Panda — AI Digital Twin Simulation Dashboard</h1>
`

const htmlFoot = `</body></html>`
