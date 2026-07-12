package stats

import "github.com/fairride/ai_simulation/domain/entity"

// HeatmapCell is one (zone, hour-of-day) bucket's full PHẦN 8 layer set —
// Ride/Delivery/Driver(supply)/Demand/Supply/Surge/Cancelled/ETA, all keyed
// identically so the dashboard can render any one of them as a heatmap
// layer from the same cell list.
type HeatmapCell struct {
	Zone  string `json:"zone"`
	Hour  int    `json:"hour"`

	// Count is every request (Ride+Delivery) originating in this zone at
	// this hour — kept as the top-level field for backward compatibility
	// with the original dashboard.html heatmap, which reads it directly.
	Count int `json:"count"`

	RideCount      int `json:"ride_count"`
	DeliveryCount  int `json:"delivery_count"`
	CancelledCount int `json:"cancelled_count"`

	// DemandCount is an explicit alias for Count — PHẦN 8 names "Demand"
	// as its own layer; this is the same number under the name the brief
	// uses, not a second independent measurement.
	DemandCount int `json:"demand_count"`

	// AverageSupply is the average number of available (online, idle)
	// drivers in this zone during this hour-of-day across the whole run —
	// sampled every tick (see simulation/world.go's RecordZoneSupplySample),
	// 0 if this zone/hour combination was never sampled with any online
	// driver present.
	AverageSupply float64 `json:"average_supply"`

	AverageSurgeMultiplier float64 `json:"average_surge_multiplier"`
	AverageETAMinutes      float64 `json:"average_eta_minutes"`

	// CompletedCount/AverageFareVND/AverageDriverIncomeVND extend the
	// original PHẦN 8 layer set for city_dashboard.json's Business
	// Intelligence view (Completed Trips / Fare / Driver Income per zone).
	CompletedCount         int     `json:"completed_count"`
	AverageFareVND         float64 `json:"average_fare_vnd"`
	AverageDriverIncomeVND float64 `json:"average_driver_income_vnd"`
}

type Heatmap struct {
	Cells []HeatmapCell `json:"cells"`
}

// BuildHeatmap buckets every requested trip by pickup zone and the hour of
// day (0-23) it was requested — hour is derived from RequestedAtTick
// (1 tick = 1 minute, see domain/entity/clock.go). supplyByZoneHour comes
// from World.SupplyByZoneHour() (see that method's doc comment for why the
// hand-off is a plain map rather than a direct dependency).
func (c *Collector) BuildHeatmap(trips []*entity.SimTrip, supplyByZoneHour map[string]float64) Heatmap {
	type accumulator struct {
		count, rideCount, deliveryCount, cancelledCount, completedCount int
		surgeSum                                        float64
		surgeN                                          int
		etaSum                                           float64
		etaN                                             int
		fareSum, incomeSum                              float64
	}
	acc := map[string]*accumulator{}
	cellKey := func(zone string, hour int) string { return zone + "|" + string(rune('0'+hour/10)) + string(rune('0'+hour%10)) }

	for _, t := range trips {
		hour := int((t.RequestedAtTick % (24 * 60)) / 60)
		zone := string(t.PickupZone)
		key := cellKey(zone, hour)
		a := acc[key]
		if a == nil {
			a = &accumulator{}
			acc[key] = a
		}
		a.count++
		if t.Kind == entity.KindDelivery {
			a.deliveryCount++
		} else {
			a.rideCount++
		}
		if t.Outcome == entity.OutcomeCancelled {
			a.cancelledCount++
		}
		if t.Outcome == entity.OutcomeCompleted {
			a.completedCount++
			a.fareSum += float64(t.FinalFareVND)
			a.incomeSum += float64(t.DriverNetVND)
		}
		if t.SurgeMultiplier > 0 {
			a.surgeSum += t.SurgeMultiplier
			a.surgeN++
		}
		if t.ETAMinutes > 0 {
			a.etaSum += t.ETAMinutes
			a.etaN++
		}
	}

	var out Heatmap
	for _, z := range entity.AllZoneTypes() {
		for h := 0; h < 24; h++ {
			key := cellKey(string(z), h)
			a := acc[key]
			cell := HeatmapCell{Zone: string(z), Hour: h, AverageSupply: supplyByZoneHour[key]}
			if a != nil {
				cell.Count = a.count
				cell.DemandCount = a.count
				cell.RideCount = a.rideCount
				cell.DeliveryCount = a.deliveryCount
				cell.CancelledCount = a.cancelledCount
				cell.AverageSurgeMultiplier = avg(a.surgeSum, a.surgeN)
				cell.AverageETAMinutes = avg(a.etaSum, a.etaN)
				cell.CompletedCount = a.completedCount
				cell.AverageFareVND = avg(a.fareSum, a.completedCount)
				cell.AverageDriverIncomeVND = avg(a.incomeSum, a.completedCount)
			}
			out.Cells = append(out.Cells, cell)
		}
	}
	return out
}
