package stats

import (
	"fmt"

	"github.com/fairride/ai_simulation/domain/entity"
)

// DriverAnalytics is PHẦN 4's requested driver-side deep-dive — a
// population-level distributional/aggregate view, complementing
// driver_statistics.json's per-driver record list.
type DriverAnalytics struct {
	IncomeDistribution     []HistogramBucket `json:"income_week_distribution"`
	FatigueDistribution    []HistogramBucket `json:"fatigue_distribution"`
	AverageOnlineHours     float64           `json:"average_online_hours_total"`
	AverageFuelLevel       float64           `json:"average_fuel_level"`
	AcceptanceRatePercent  float64           `json:"acceptance_rate_percent"`
	CancellationRatePercent float64          `json:"cancellation_rate_percent"`
	AverageRideTripsPerDriver     float64    `json:"average_ride_trips_per_driver"`
	AverageDeliveryTripsPerDriver float64    `json:"average_delivery_trips_per_driver"`
	RideToDeliveryRatio   float64            `json:"ride_to_delivery_ratio"` // ride trips per 1 delivery trip
	RetentionRatePercent  float64            `json:"retention_rate_percent"` // avg(days active) / total simulated days
	AverageSatisfaction   float64            `json:"average_satisfaction"`
}

// BuildDriverAnalytics aggregates driver population state plus the trip
// ledger (for the Ride/Delivery/cancellation splits, which live on SimTrip,
// not DriverAgent) into one distributional report. offersAccepted/
// offersRejected/totalDays come from World, which is the only place that
// tracks them (see world.go's recordDriverOfferOutcome doc comment).
func (c *Collector) BuildDriverAnalytics(drivers map[string]*entity.DriverAgent, trips []*entity.SimTrip, offersAccepted, offersRejected, totalDays int) DriverAnalytics {
	var out DriverAnalytics
	if len(drivers) == 0 {
		return out
	}

	incomes := make([]float64, 0, len(drivers))
	fatigues := make([]float64, 0, len(drivers))
	var hoursSum, fuelSum, satisfactionSum, daysActiveSum float64
	for _, d := range drivers {
		incomes = append(incomes, float64(d.IncomeWeek))
		fatigues = append(fatigues, d.Fatigue)
		hoursSum += d.TotalHoursOnline
		fuelSum += d.Fuel
		satisfactionSum += d.Satisfaction
		daysActiveSum += float64(d.DaysActive)
	}
	n := len(drivers)
	out.AverageOnlineHours = hoursSum / float64(n)
	out.AverageFuelLevel = fuelSum / float64(n)
	out.AverageSatisfaction = satisfactionSum / float64(n)
	if totalDays > 0 {
		out.RetentionRatePercent = 100 * (daysActiveSum / float64(n)) / float64(totalDays)
	}
	out.IncomeDistribution = buildLinearHistogram(incomes, 500_000, 12, "đ")
	out.FatigueDistribution = buildLinearHistogram(fatigues, 0.1, 10, "")

	if total := offersAccepted + offersRejected; total > 0 {
		out.AcceptanceRatePercent = 100 * float64(offersAccepted) / float64(total)
	}

	var rideCompleted, deliveryCompleted, assignedThenCancelled, assignedTotal int
	for _, t := range trips {
		if t.DriverID == "" {
			continue
		}
		assignedTotal++
		if t.Outcome == entity.OutcomeCancelled {
			assignedThenCancelled++
			continue
		}
		if t.Outcome != entity.OutcomeCompleted {
			continue
		}
		if t.Kind == entity.KindDelivery {
			deliveryCompleted++
		} else {
			rideCompleted++
		}
	}
	if assignedTotal > 0 {
		out.CancellationRatePercent = 100 * float64(assignedThenCancelled) / float64(assignedTotal)
	}
	out.AverageRideTripsPerDriver = float64(rideCompleted) / float64(n)
	out.AverageDeliveryTripsPerDriver = float64(deliveryCompleted) / float64(n)
	if deliveryCompleted > 0 {
		out.RideToDeliveryRatio = float64(rideCompleted) / float64(deliveryCompleted)
	}
	return out
}

// buildLinearHistogram buckets values into bucketCount fixed-width bands
// starting at 0 — shared by Income/Fatigue distributions (and reusable by
// any future population-level histogram).
func buildLinearHistogram(values []float64, bucketWidth float64, bucketCount int, unit string) []HistogramBucket {
	counts := make([]int, bucketCount)
	for _, v := range values {
		idx := int(v / bucketWidth)
		if idx < 0 {
			idx = 0
		}
		if idx >= bucketCount {
			idx = bucketCount - 1
		}
		counts[idx]++
	}
	out := make([]HistogramBucket, bucketCount)
	for i := 0; i < bucketCount; i++ {
		lo := formatBucketBound(float64(i)*bucketWidth, unit)
		if i == bucketCount-1 {
			out[i] = HistogramBucket{RangeLabel: lo + "+", Count: counts[i]}
			continue
		}
		hi := formatBucketBound(float64(i+1)*bucketWidth, unit)
		out[i] = HistogramBucket{RangeLabel: lo + "-" + hi, Count: counts[i]}
	}
	return out
}

func formatBucketBound(v float64, unit string) string {
	if unit == "đ" {
		return fmt.Sprintf("%dk", int64(v/1000))
	}
	return fmt.Sprintf("%.1f", v)
}
