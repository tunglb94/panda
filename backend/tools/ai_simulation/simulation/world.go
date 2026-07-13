// Package simulation is the orchestration layer: the tick loop, the world
// state every agent/adapter reads and writes, and the per-ride-request flow
// that ties Rule Engine + AI Decision Engine + the real Pricing/Promotion/
// Dispatch/Driver-Economy engines together.
package simulation

import (
	"math/rand"

	"github.com/fairride/ai_simulation/aiengine"
	"github.com/fairride/ai_simulation/benchmark"
	"github.com/fairride/ai_simulation/domain/entity"
	"github.com/fairride/ai_simulation/integration"
	"github.com/fairride/ai_simulation/stats"
)

// World is the single mutable state container for one simulation run.
type World struct {
	City   *entity.City
	Clock  *entity.SimClock
	Rand   *rand.Rand

	Drivers map[string]*entity.DriverAgent
	Riders  map[string]*entity.RiderAgent

	// DriverIDs/RiderIDs are Drivers/Riders' keys, sorted once after
	// seeding — Go randomizes map iteration order per-process, so any
	// per-tick loop that both ranges over Drivers/Riders AND draws from
	// Rand (processTick's rider loop, evaluateDriverState's driver loop)
	// must iterate these slices instead of the maps directly, or the same
	// --seed produces a different draw sequence (and therefore different
	// results) on every run. Populated once in NewEngine; never mutated
	// after (no driver/rider is added or removed mid-run).
	DriverIDs []string
	RiderIDs  []string

	Weather   entity.Weather
	Traffic   entity.Traffic
	Scenarios entity.ActiveScenarios

	// Zone-level live demand/supply counters, reset and rebuilt every tick —
	// this is what feeds PricingAdapter's Demand Surge signal
	// (ActiveRequests/AvailableDrivers), the same shape the real Dynamic
	// Pricing Engine expects in production.
	ZoneActiveRequests   map[entity.ZoneType]int
	ZoneAvailableDrivers map[entity.ZoneType]int

	Pricing    *integration.PricingAdapter
	Promotion  *integration.PromotionAdapter
	Dispatch   *integration.DispatchAdapter
	Economy    *integration.DriverEconomy
	Delivery   *integration.DeliveryAdapter

	AI    *aiengine.DecisionEngine
	Stats *stats.Collector
	Bench *benchmark.Tracker

	trips []*entity.SimTrip

	voucherUsedCount int
	voucherKeptCount int

	driverOffersAccepted int
	driverOffersRejected int

	voucherIssuedCount int

	// Decision-outcome counters for bi/driver_behavior.json and
	// bi/passenger_behavior.json (PHẦN 8/9 of the Business Intelligence
	// brief) — none of these are derivable from SimTrip alone, since a
	// decision doesn't always leave a distinct trip-level trace (e.g. a
	// FatigueDecision=Stop just takes the driver offline; no trip record is
	// created for that event at all).
	fatigueContinueCount int
	fatigueStopCount     int
	switchAppCount       int // rider chose to switch to the simulated competitor
	stayOnPandaCount     int
	surgeChaseCount      int // driver chose to relocate toward a surged zone
	surgeStayCount       int

	// heatmapSupplySamples accumulates ZoneAvailableDrivers snapshots keyed
	// by "zone|hour" — heatmap.json's Supply/Driver layer needs a
	// time-of-day average, which isn't derivable from the trip ledger
	// alone (ZoneAvailableDrivers is a live per-tick counter, overwritten
	// every tick by RefreshZoneCounters, not retained historically).
	heatmapSupplySum   map[string]float64
	heatmapSupplyCount map[string]int
}

// RecordZoneSupplySample snapshots this tick's per-zone available-driver
// count into the hour-of-day bucket it falls in — called once per tick
// (see engine.go's Run loop) right after RefreshZoneCounters.
func (w *World) RecordZoneSupplySample() {
	hour := w.Clock.Hour()
	for _, z := range entity.AllZoneTypes() {
		key := zoneHourKey(z, hour)
		w.heatmapSupplySum[key] += float64(w.ZoneAvailableDrivers[z])
		w.heatmapSupplyCount[key]++
	}
}

// AverageZoneSupply returns the average available-driver count sampled for
// zone at hour across the whole run — 0 if never sampled.
func (w *World) AverageZoneSupply(z entity.ZoneType, hour int) float64 {
	key := zoneHourKey(z, hour)
	n := w.heatmapSupplyCount[key]
	if n == 0 {
		return 0
	}
	return w.heatmapSupplySum[key] / float64(n)
}

func zoneHourKey(z entity.ZoneType, hour int) string {
	return string(z) + "|" + string(rune('0'+hour/10)) + string(rune('0'+hour%10))
}

// SupplyByZoneHour returns every sampled zone/hour's average available-driver
// count, keyed identically to stats.HeatmapCell's own Zone+Hour (via the
// same zoneHourKey format) — the hand-off point that lets the stats package
// build heatmap.json's Supply layer without importing simulation (which
// would create an import cycle, since simulation already imports stats).
func (w *World) SupplyByZoneHour() map[string]float64 {
	out := make(map[string]float64, len(w.heatmapSupplySum))
	for key, sum := range w.heatmapSupplySum {
		out[key] = sum / float64(w.heatmapSupplyCount[key])
	}
	return out
}

// RecordTrip appends a completed/failed ride-request record — the single
// source every exporter (stats, dashboard, heatmap) reads from afterward.
func (w *World) RecordTrip(t *entity.SimTrip) {
	w.trips = append(w.trips, t)
}

// Trips returns every ride-request record from the run so far.
func (w *World) Trips() []*entity.SimTrip {
	return w.trips
}

// RefreshZoneCounters recomputes live per-zone demand/supply — called once
// per tick before any ride requests are processed for that tick.
func (w *World) RefreshZoneCounters() {
	for _, z := range entity.AllZoneTypes() {
		w.ZoneActiveRequests[z] = 0
		w.ZoneAvailableDrivers[z] = 0
	}
	for _, d := range w.Drivers {
		if d.IsAvailable() {
			w.ZoneAvailableDrivers[d.Zone]++
		}
	}
}

// recordDecision tallies one ambiguous-decision resolution into the
// benchmark tracker — usedAI distinguishes an AI-resolved decision from one
// the rule engine settled confidently on its own.
func (w *World) recordDecision(usedAI bool) {
	if usedAI {
		w.Bench.RecordAIDecision()
	} else {
		w.Bench.RecordRuleDecision()
	}
}

// recordVoucherOutcome tallies whether a held voucher was spent or kept —
// feeds VoucherStatistics.UsedCount/KeptCount, which aren't derivable from
// SimTrip alone (a "kept" decision produces no trip-level voucher field at
// all).
func (w *World) recordVoucherOutcome(used bool) {
	if used {
		w.voucherUsedCount++
	} else {
		w.voucherKeptCount++
	}
}

// recordDriverOfferOutcome tallies whether a driver accepted or rejected an
// offered trip (see ride_flow.go's driverAcceptsOffer) — feeds
// driver_analytics.json's aggregate Acceptance rate, which isn't derivable
// from SimTrip alone (resolveOffer's retry loop can reject several drivers
// before one finally accepts; only the final assignment survives onto the
// trip record).
func (w *World) recordDriverOfferOutcome(accepted bool) {
	if accepted {
		w.driverOffersAccepted++
	} else {
		w.driverOffersRejected++
	}
}

// recordFatigueOutcome/recordSwitchAppOutcome/recordSurgeChaseOutcome tally
// the 3 remaining named decision types PHẦN 8/9 asks to see counted — the
// 4th (VoucherUse) already has recordVoucherOutcome, and driver offer
// accept/reject already has recordDriverOfferOutcome above.
func (w *World) recordFatigueOutcome(continued bool) {
	if continued {
		w.fatigueContinueCount++
	} else {
		w.fatigueStopCount++
	}
}

func (w *World) recordSwitchAppOutcome(switched bool) {
	if switched {
		w.switchAppCount++
	} else {
		w.stayOnPandaCount++
	}
}

func (w *World) recordSurgeChaseOutcome(chased bool) {
	if chased {
		w.surgeChaseCount++
	} else {
		w.surgeStayCount++
	}
}
