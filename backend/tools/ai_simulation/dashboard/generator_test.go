package dashboard

import (
	"strings"
	"testing"

	"github.com/fairride/ai_simulation/stats"
)

// TestGenerate_EmptyMapsDoNotProduceJSNull guards against a real bug found
// during manual verification: encoding/json marshals a nil map/slice as
// `null`, and `data: null` throws inside Chart.js — which aborts every
// later `new Chart(...)` call in the same <script> block. An empty
// promotion-cost map (no promotions fired yet) must still render as `[]`.
func TestGenerate_EmptyMapsDoNotProduceJSNull(t *testing.T) {
	bundle := stats.Bundle{
		DriverStatistics:    stats.DriverStatistics{AccountTypeCounts: map[string]int{}},
		PromotionStatistics: stats.PromotionStatistics{CostByType: map[string]int64{}},
		PricingStatistics:   stats.PricingStatistics{FareHistogram: nil},
		Heatmap:             stats.Heatmap{Cells: nil},
	}
	html := Generate(bundle)

	if strings.Contains(html, "= null") {
		t.Errorf("dashboard.html must never assign a JS array as null (breaks every later Chart.js call); got:\n%s", html)
	}
	for _, want := range []string{"promoLabels = []", "promoData = []", "tierLabels = []", "tierData = []", "histLabels = []", "histData = []"} {
		if !strings.Contains(html, want) {
			t.Errorf("expected dashboard.html to contain %q for an empty series", want)
		}
	}
}

func TestGenerate_ProducesWellFormedHTMLDocument(t *testing.T) {
	bundle := stats.Bundle{
		DriverStatistics:    stats.DriverStatistics{AccountTypeCounts: map[string]int{"bronze": 5}},
		PromotionStatistics: stats.PromotionStatistics{CostByType: map[string]int64{"first_ride": 10000}},
	}
	html := Generate(bundle)

	if !strings.HasPrefix(html, "<!doctype html>") {
		t.Errorf("expected the document to start with a doctype")
	}
	if !strings.HasSuffix(html, "</body></html>") {
		t.Errorf("expected the document to end with a closed body/html")
	}
	if strings.Count(html, "<script") != 2 { // Chart.js CDN + inline script
		t.Errorf("expected exactly 2 <script> tags, got %d", strings.Count(html, "<script"))
	}
	if !strings.Contains(html, "new Chart(document.getElementById('dispatchPie'") {
		t.Errorf("expected the dispatch pie chart to be wired up")
	}
}
