package bi

import (
	"github.com/fairride/ai_simulation/domain/entity"
	"github.com/fairride/ai_simulation/stats"
)

// SurgeAnalysis is PHẦN 5.
type SurgeAnalysis struct {
	TotalSurgedTrips int `json:"total_surged_trips"` // SurgeMultiplier > 1.0
	// TierCounts buckets every surged trip into the named multiplier tier
	// it falls into — "x1.2" means 1.2 <= multiplier < 1.5, etc.; a trip
	// contributes to exactly one tier.
	TierCounts map[string]int `json:"tier_counts"`

	// AverageEpisodeHours is derived from heatmap.json's own zone/hour
	// average-surge cells: a "surge episode" is a maximal run of
	// consecutive hours in one zone where the average surge multiplier
	// exceeds 1.2 — this is the closest real signal to "how long did surge
	// last" the simulation's tick-level (not event-level) surge model
	// supports; see Assumptions.
	AverageEpisodeHours float64 `json:"average_episode_hours"`
	EpisodeCount        int     `json:"episode_count"`

	// RevenueUpliftVND is the extra metered-fare revenue surge pricing
	// generated: sum over surged trips of BaseFareVND*(SurgeMultiplier-1) —
	// isolates the surge contribution specifically, independent of
	// voucher/promotion discounts applied afterward.
	RevenueUpliftVND int64 `json:"revenue_uplift_vnd"`

	DriverSurgeChaseCount   int     `json:"driver_surge_chase_count"`
	DriverSurgeStayCount    int     `json:"driver_surge_stay_count"`
	DriverChaseRatePercent  float64 `json:"driver_chase_rate_percent"`

	Assumptions []Assumption `json:"assumptions"`
}

// ComputeSurgeAnalysis is PHẦN 5.
func ComputeSurgeAnalysis(in Input) SurgeAnalysis {
	out := SurgeAnalysis{TierCounts: map[string]int{"x1.2": 0, "x1.5": 0, "x2": 0, "x3": 0}}

	for _, t := range in.Trips {
		if t.Outcome != entity.OutcomeCompleted || t.SurgeMultiplier <= 1.0 {
			continue
		}
		out.TotalSurgedTrips++
		out.RevenueUpliftVND += int64(float64(t.BaseFareVND) * (t.SurgeMultiplier - 1))

		switch {
		case t.SurgeMultiplier >= 3.0:
			out.TierCounts["x3"]++
		case t.SurgeMultiplier >= 2.0:
			out.TierCounts["x2"]++
		case t.SurgeMultiplier >= 1.5:
			out.TierCounts["x1.5"]++
		case t.SurgeMultiplier >= 1.2:
			out.TierCounts["x1.2"]++
		}
	}

	out.AverageEpisodeHours, out.EpisodeCount = surgeEpisodes(in.Bundle.Heatmap.Cells)

	out.DriverSurgeChaseCount = in.SurgeChaseCount
	out.DriverSurgeStayCount = in.SurgeStayCount
	if total := in.SurgeChaseCount + in.SurgeStayCount; total > 0 {
		out.DriverChaseRatePercent = 100 * float64(in.SurgeChaseCount) / float64(total)
	}

	out.Assumptions = []Assumption{
		{Title: "Surge episode duration", Detail: "Suy ra từ heatmap.json (chuỗi giờ liên tiếp có average_surge_multiplier > 1.2 trong cùng 1 khu vực) — simulation không có khái niệm \"sự kiện surge\" với thời điểm bắt đầu/kết thúc rõ ràng như production thật, đây là proxy hợp lý nhất từ dữ liệu tick-level hiện có."},
	}
	return out
}

// surgeEpisodes scans each zone's 24 hourly cells for maximal runs of
// consecutive hours with AverageSurgeMultiplier > 1.2, returns the average
// run length (in hours) and how many such runs were found across all zones.
func surgeEpisodes(cells []stats.HeatmapCell) (float64, int) {
	byZone := map[string][24]float64{}
	for _, c := range cells {
		if c.Hour < 0 || c.Hour > 23 {
			continue
		}
		arr := byZone[c.Zone]
		arr[c.Hour] = c.AverageSurgeMultiplier
		byZone[c.Zone] = arr
	}

	const threshold = 1.2
	var totalLength, episodeCount int
	for _, hours := range byZone {
		runLen := 0
		for h := 0; h < 24; h++ {
			if hours[h] > threshold {
				runLen++
				continue
			}
			if runLen > 0 {
				totalLength += runLen
				episodeCount++
				runLen = 0
			}
		}
		if runLen > 0 {
			totalLength += runLen
			episodeCount++
		}
	}
	if episodeCount == 0 {
		return 0, 0
	}
	return float64(totalLength) / float64(episodeCount), episodeCount
}
