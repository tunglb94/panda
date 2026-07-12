package audit

import (
	"sort"

	"github.com/fairride/ai_simulation/domain/entity"
	"github.com/fairride/ai_simulation/stats"
)

// computeZoneStats builds per-pickup-zone figures feeding #11 (dispatch
// bias), #12 (highest ETA), #13/#14 (under/oversupplied zones) — one pass
// over trips for accept-rate/ETA/profit, plus heatmap.json's own supply
// samples (already collected during the run, not recomputed here) for the
// demand/supply ratio.
func computeZoneStats(r *Report, trips []*entity.SimTrip, bundle stats.Bundle) {
	byZone := map[string]*ZoneStat{}
	get := func(zone string) *ZoneStat {
		z := byZone[zone]
		if z == nil {
			z = &ZoneStat{Zone: zone}
			byZone[zone] = z
		}
		return z
	}

	for _, t := range trips {
		z := get(string(t.PickupZone))
		z.Requested++
		switch t.Outcome {
		case entity.OutcomeCompleted:
			z.Accepted++
			z.TripCount++
			z.ProfitVND += stats.PerTripProfitVND(t.CommissionVND)
		case entity.OutcomeRejected:
			z.Rejected++
		}
	}
	for _, z := range byZone {
		if z.Requested > 0 {
			z.AcceptRatePercent = 100 * float64(z.Accepted) / float64(z.Requested)
		}
	}

	// ETA + demand/supply ratio come from heatmap.json's cells (already
	// aggregated per zone/hour during the run, including the
	// RecordZoneSupplySample data trips alone can't reconstruct).
	etaSumByZone := map[string]float64{}
	etaNByZone := map[string]int{}
	ratioSumByZone := map[string]float64{}
	ratioNByZone := map[string]int{}
	for _, c := range bundle.Heatmap.Cells {
		if c.AverageETAMinutes > 0 {
			etaSumByZone[c.Zone] += c.AverageETAMinutes * float64(c.Count)
			etaNByZone[c.Zone] += c.Count
		}
		if c.DemandCount > 0 {
			supply := c.AverageSupply
			if supply < 0.1 {
				supply = 0.1
			}
			ratioSumByZone[c.Zone] += float64(c.DemandCount) / supply
			ratioNByZone[c.Zone]++
		}
	}
	for zone, z := range byZone {
		if n := etaNByZone[zone]; n > 0 {
			z.AverageETAMinutes = etaSumByZone[zone] / float64(n)
		}
		if n := ratioNByZone[zone]; n > 0 {
			z.AverageDemandSupply = ratioSumByZone[zone] / float64(n)
		}
	}

	r.ZoneStats = make([]ZoneStat, 0, len(byZone))
	for _, z := range byZone {
		r.ZoneStats = append(r.ZoneStats, *z)
	}
	sort.Slice(r.ZoneStats, func(i, j int) bool { return r.ZoneStats[i].Zone < r.ZoneStats[j].Zone })
}

// HighestETAZone/UndersuppliedZones/OversuppliedZones are convenience views
// over ZoneStats for the report renderers (kept as methods rather than
// extra Report fields, since they're pure derivations of ZoneStats).
func (r Report) HighestETAZone() (ZoneStat, bool) {
	var best ZoneStat
	found := false
	for _, z := range r.ZoneStats {
		if !found || z.AverageETAMinutes > best.AverageETAMinutes {
			best, found = z, true
		}
	}
	return best, found
}

func (r Report) UndersuppliedZones(n int) []ZoneStat {
	sorted := append([]ZoneStat(nil), r.ZoneStats...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].AverageDemandSupply > sorted[j].AverageDemandSupply })
	if len(sorted) > n {
		sorted = sorted[:n]
	}
	return sorted
}

func (r Report) OversuppliedZones(n int) []ZoneStat {
	sorted := append([]ZoneStat(nil), r.ZoneStats...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].AverageDemandSupply < sorted[j].AverageDemandSupply })
	if len(sorted) > n {
		sorted = sorted[:n]
	}
	return sorted
}

func computeRideDeliveryRatio(r *Report, trips []*entity.SimTrip) {
	for _, t := range trips {
		if t.Outcome != entity.OutcomeCompleted {
			continue
		}
		if t.Kind == entity.KindDelivery {
			r.DeliveryCount++
		} else {
			r.RideCount++
		}
	}
	if r.DeliveryCount > 0 {
		r.RideToDeliveryRatio = float64(r.RideCount) / float64(r.DeliveryCount)
	}
}

const (
	peakHourStartMorning = 7
	peakHourEndMorning   = 9
	peakHourStartEvening = 17
	peakHourEndEvening   = 20
)

func isPeakHour(hour int) bool {
	return (hour >= peakHourStartMorning && hour < peakHourEndMorning) || (hour >= peakHourStartEvening && hour < peakHourEndEvening)
}

// computeSegmentProfits covers #16 (Airport), #17 (Peak Hour), #18 (Off
// Peak), #19 (Weather).
func computeSegmentProfits(r *Report, trips []*entity.SimTrip) {
	airport := &segmentAccumulator{}
	peak := &segmentAccumulator{}
	offPeak := &segmentAccumulator{}
	byWeather := map[entity.Weather]*segmentAccumulator{}

	for _, t := range trips {
		hour := int((t.RequestedAtTick % (24 * 60)) / 60)
		isAirport := t.PickupZone == entity.ZoneAirport || t.DestinationZone == entity.ZoneAirport
		if isAirport {
			airport.add(t)
		}
		if isPeakHour(hour) {
			peak.add(t)
		} else {
			offPeak.add(t)
		}
		if byWeather[t.Weather] == nil {
			byWeather[t.Weather] = &segmentAccumulator{}
		}
		byWeather[t.Weather].add(t)
	}

	r.AirportProfit = airport.build("Airport")
	r.PeakHourProfit = peak.build("Peak Hour")
	r.OffPeakProfit = offPeak.build("Off Peak")
	for weather, acc := range byWeather {
		r.WeatherImpact = append(r.WeatherImpact, acc.build(string(weather)))
	}
	sort.Slice(r.WeatherImpact, func(i, j int) bool { return r.WeatherImpact[i].Label < r.WeatherImpact[j].Label })
}

type segmentAccumulator struct {
	requested, completed, cancelled int
	fareSum, surgeSum               float64
	profitSum                       int64
}

func (a *segmentAccumulator) add(t *entity.SimTrip) {
	a.requested++
	switch t.Outcome {
	case entity.OutcomeCompleted:
		a.completed++
		a.fareSum += float64(t.FinalFareVND)
		a.surgeSum += t.SurgeMultiplier
		a.profitSum += stats.PerTripProfitVND(t.CommissionVND)
	case entity.OutcomeCancelled:
		a.cancelled++
	}
}

func (a *segmentAccumulator) build(label string) SegmentProfit {
	seg := SegmentProfit{Label: label, TripCount: a.completed, TotalProfitVND: a.profitSum}
	if a.completed > 0 {
		seg.AverageProfitVND = float64(a.profitSum) / float64(a.completed)
		seg.AverageFareVND = a.fareSum / float64(a.completed)
		seg.AverageSurge = a.surgeSum / float64(a.completed)
	}
	if a.requested > 0 {
		seg.CancellationPercent = 100 * float64(a.cancelled) / float64(a.requested)
	}
	return seg
}

// computeDriverIncomeStats fills the Executive Summary's mean/median driver
// weekly income — median is not computed anywhere else in this tool.
func computeDriverIncomeStats(r *Report, drivers map[string]*entity.DriverAgent) {
	if len(drivers) == 0 {
		return
	}
	incomes := make([]float64, 0, len(drivers))
	var sum float64
	for _, d := range drivers {
		incomes = append(incomes, float64(d.IncomeWeek))
		sum += float64(d.IncomeWeek)
	}
	r.AverageDriverIncomeWeekVND = sum / float64(len(incomes))
	sort.Float64s(incomes)
	mid := len(incomes) / 2
	if len(incomes)%2 == 0 {
		r.MedianDriverIncomeWeekVND = (incomes[mid-1] + incomes[mid]) / 2
	} else {
		r.MedianDriverIncomeWeekVND = incomes[mid]
	}
}
