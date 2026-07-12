// Package stats aggregates simulation output into the JSON exports the
// sprint brief requires (driver/rider/pricing/promotion/voucher/dispatch
// statistics, plus the overall simulation_report.json and heatmap.json).
package stats

import (
	"github.com/fairride/ai_simulation/domain/entity"
)

// Collector derives every statistic from the trip records + final agent
// state — it holds no state of its own beyond what's handed to Finalize,
// keeping it a pure aggregation step over the World's authoritative data.
type Collector struct{}

func NewCollector() *Collector { return &Collector{} }

// DispatchOutcomeCounts is shared by SimulationReport and DispatchStatistics.
type DispatchOutcomeCounts struct {
	Requested int `json:"requested"`
	Accepted  int `json:"accepted"`
	Rejected  int `json:"rejected"`
	Cancelled int `json:"cancelled"`
	Completed int `json:"completed"`
}

func (c *Collector) DispatchOutcomes(trips []*entity.SimTrip) DispatchOutcomeCounts {
	var out DispatchOutcomeCounts
	for _, t := range trips {
		out.Requested++
		switch t.Outcome {
		case entity.OutcomeCompleted:
			out.Completed++
			out.Accepted++
		case entity.OutcomeRejected:
			out.Rejected++
		case entity.OutcomeCancelled:
			out.Cancelled++
		}
	}
	return out
}

func avg(sum float64, n int) float64 {
	if n == 0 {
		return 0
	}
	return sum / float64(n)
}
