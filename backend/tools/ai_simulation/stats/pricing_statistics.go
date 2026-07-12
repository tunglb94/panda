package stats

import (
	"fmt"

	"github.com/fairride/ai_simulation/domain/entity"
)

type PricingStatistics struct {
	TotalTripsQuoted       int                        `json:"total_trips_quoted"`
	AverageBaseFareVND     float64                    `json:"average_base_fare_vnd"`
	AverageFinalFareVND    float64                    `json:"average_final_fare_vnd"`
	AverageSurgeMultiplier float64                    `json:"average_surge_multiplier"`
	SurgedTripCount        int                        `json:"surged_trip_count"`
	SurgedTripPercent      float64                    `json:"surged_trip_percent"`
	MaxSurgeObserved       float64                    `json:"max_surge_observed"`
	FareByVehicleType      map[string]VehicleFareStat `json:"fare_by_vehicle_type"`
	FareHistogram          []HistogramBucket          `json:"fare_histogram"`
}

// HistogramBucket is one bar of the dashboard's fare-distribution histogram.
type HistogramBucket struct {
	RangeLabel string `json:"range_label"`
	Count      int    `json:"count"`
}

type VehicleFareStat struct {
	Count               int     `json:"count"`
	AverageFinalFareVND float64 `json:"average_final_fare_vnd"`
}

func (c *Collector) BuildPricingStatistics(trips []*entity.SimTrip) PricingStatistics {
	out := PricingStatistics{FareByVehicleType: map[string]VehicleFareStat{}}
	var baseSum, finalSum, surgeSum float64
	byVehicle := map[string]*struct {
		count int
		sum   float64
	}{}

	for _, t := range trips {
		if t.BaseFareVND == 0 && t.FinalFareVND == 0 {
			continue // never quoted (e.g. rider abandoned before pricing)
		}
		out.TotalTripsQuoted++
		baseSum += float64(t.BaseFareVND)
		finalSum += float64(t.FinalFareVND)
		surgeSum += t.SurgeMultiplier
		if t.SurgeMultiplier > 1.0 {
			out.SurgedTripCount++
		}
		if t.SurgeMultiplier > out.MaxSurgeObserved {
			out.MaxSurgeObserved = t.SurgeMultiplier
		}
		key := string(t.ServiceType)
		if byVehicle[key] == nil {
			byVehicle[key] = &struct {
				count int
				sum   float64
			}{}
		}
		byVehicle[key].count++
		byVehicle[key].sum += float64(t.FinalFareVND)
	}

	out.AverageBaseFareVND = avg(baseSum, out.TotalTripsQuoted)
	out.AverageFinalFareVND = avg(finalSum, out.TotalTripsQuoted)
	out.AverageSurgeMultiplier = avg(surgeSum, out.TotalTripsQuoted)
	if out.TotalTripsQuoted > 0 {
		out.SurgedTripPercent = 100 * float64(out.SurgedTripCount) / float64(out.TotalTripsQuoted)
	}
	for k, v := range byVehicle {
		out.FareByVehicleType[k] = VehicleFareStat{Count: v.count, AverageFinalFareVND: avg(v.sum, v.count)}
	}
	out.FareHistogram = buildFareHistogram(trips)
	return out
}

// buildFareHistogram buckets final fares into fixed 20,000 VND-wide bands
// (0-20k, 20k-40k, ... capped at a 200k+ overflow bucket) — wide enough to
// stay readable with BRB's launch-tier fare scale (§2.2.1-§2.2.5).
func buildFareHistogram(trips []*entity.SimTrip) []HistogramBucket {
	const bucketWidth = 20_000
	const bucketCount = 10 // last bucket is "200k+"
	counts := make([]int, bucketCount)

	for _, t := range trips {
		if t.FinalFareVND <= 0 {
			continue
		}
		idx := int(t.FinalFareVND / bucketWidth)
		if idx >= bucketCount {
			idx = bucketCount - 1
		}
		counts[idx]++
	}

	out := make([]HistogramBucket, bucketCount)
	for i := 0; i < bucketCount; i++ {
		lo := i * bucketWidth / 1000
		if i == bucketCount-1 {
			out[i] = HistogramBucket{RangeLabel: fmt.Sprintf("%dk+", lo), Count: counts[i]}
			continue
		}
		hi := (i + 1) * bucketWidth / 1000
		out[i] = HistogramBucket{RangeLabel: fmt.Sprintf("%d-%dk", lo, hi), Count: counts[i]}
	}
	return out
}
