package stats

import "github.com/fairride/ai_simulation/domain/entity"

type DispatchStatistics struct {
	DispatchOutcomeCounts
	AcceptRatePercent  float64 `json:"accept_rate_percent"`
	AverageETAMinutes  float64 `json:"average_eta_minutes"`
	AveragePickupMinutes float64 `json:"average_pickup_minutes"`
}

func (c *Collector) BuildDispatchStatistics(trips []*entity.SimTrip) DispatchStatistics {
	counts := c.DispatchOutcomes(trips)
	out := DispatchStatistics{DispatchOutcomeCounts: counts}

	var etaSum, pickupSum float64
	var etaN, pickupN int
	for _, t := range trips {
		if t.ETAMinutes > 0 {
			etaSum += t.ETAMinutes
			etaN++
		}
		if t.PickupMinutes > 0 {
			pickupSum += t.PickupMinutes
			pickupN++
		}
	}
	out.AverageETAMinutes = avg(etaSum, etaN)
	out.AveragePickupMinutes = avg(pickupSum, pickupN)
	if counts.Requested > 0 {
		out.AcceptRatePercent = 100 * float64(counts.Accepted) / float64(counts.Requested)
	}
	return out
}
