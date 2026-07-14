package simulation

import (
	"context"
	"fmt"
	"math/rand"
	"sort"
	"time"

	pricingentity "github.com/fairride/pricing/domain/entity"

	"github.com/fairride/ai_simulation/aiengine"
	"github.com/fairride/ai_simulation/benchmark"
	"github.com/fairride/ai_simulation/domain/entity"
	"github.com/fairride/ai_simulation/integration"
	"github.com/fairride/ai_simulation/ruleengine"
	"github.com/fairride/ai_simulation/stats"
)

// Config drives one simulation run — a 1:1 mapping of the CLI flags the
// sprint brief specifies.
type Config struct {
	Drivers   int
	Riders    int
	Days      int
	Model     string
	OllamaURL string
	Seed      int64
}

// Engine owns one simulation run's World and executes its tick loop.
type Engine struct {
	cfg   Config
	world *World
}

// NewEngine builds every adapter (real Pricing/Promotion/Dispatch engines,
// BRB-sourced Driver Economy, the AI Decision Engine against Ollama) and
// populates the World with cfg.Drivers/cfg.Riders randomized agents.
func NewEngine(cfg Config) *Engine {
	if cfg.Seed == 0 {
		cfg.Seed = time.Now().UnixNano()
	}
	rnd := rand.New(rand.NewSource(cfg.Seed))
	// dispatchRng is a separate, independently-seeded source used only to
	// break exact-distance ties in dispatch matching (see
	// dispatch_fakes.go's shuffleTiedGroups) — kept apart from rnd so this
	// fix doesn't perturb the draw sequence every other subsystem already
	// relies on for its own --seed reproducibility.
	dispatchRng := rand.New(rand.NewSource(cfg.Seed + 1))
	start := time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC) // a Monday — makes IsWeekend() predictable across runs

	world := &World{
		City:                 entity.NewDefaultCity(),
		Clock:                entity.NewSimClock(start),
		Rand:                 rnd,
		Drivers:              make(map[string]*entity.DriverAgent, cfg.Drivers),
		Riders:               make(map[string]*entity.RiderAgent, cfg.Riders),
		Weather:              entity.WeatherSunny,
		Traffic:              entity.TrafficClear,
		Scenarios:            entity.ActiveScenarios{},
		ZoneActiveRequests:   make(map[entity.ZoneType]int),
		ZoneAvailableDrivers: make(map[entity.ZoneType]int),
		heatmapSupplySum:     make(map[string]float64),
		heatmapSupplyCount:   make(map[string]int),
		Pricing:              integration.NewPricingAdapter(),
		Promotion:            integration.NewPromotionAdapter(start),
		Dispatch:             integration.NewDispatchAdapter(dispatchRng),
		Economy:              integration.NewDriverEconomy(),
		Delivery:             integration.NewDeliveryAdapter(),
		AI:                   aiengine.NewDecisionEngine(cfg.OllamaURL, cfg.Model, 20*time.Second),
		Stats:                stats.NewCollector(),
		Bench:                benchmark.NewTracker(),
	}

	seedDrivers(world, cfg.Drivers)
	seedRiders(world, cfg.Riders)
	world.DriverIDs = sortedKeys(world.Drivers)
	world.RiderIDs = sortedKeys(world.Riders)

	return &Engine{cfg: cfg, world: world}
}

// AIEnabled reports whether Ollama was reachable at startup.
func (e *Engine) AIEnabled() bool { return e.world.AI.Enabled() }

// Run executes cfg.Days simulated days (1440 ticks each) and returns the
// final statistics bundle. ctx cancellation stops the run early (partial
// results are still returned/exported).
func (e *Engine) Run(ctx context.Context) error {
	w := e.world
	totalTicks := int64(e.cfg.Days) * 24 * 60
	scheduler := NewScenarioScheduler(w.Rand)

	for w.Clock.Tick < totalTicks {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if w.Clock.MinuteOfDay() == 0 {
			onNewDay(w)
			scheduler.RollDailyConditions(w)
		}
		if w.Clock.MinuteOfDay()%180 == 0 { // traffic re-rolls every 3 simulated hours
			scheduler.RollTraffic(w)
		}
		ApplyHourlyScenarios(w)
		w.RefreshZoneCounters()
		w.RecordZoneSupplySample()

		processTick(ctx, w)

		w.Bench.RecordTick()
		if w.Clock.MinuteOfDay() == 0 {
			w.Bench.SampleMemory()
		}
		w.Clock.Advance()
	}
	return nil
}

// Export writes every JSON statistics file plus dashboard.html to dir.
func (e *Engine) Export(dir string) error {
	w := e.world
	trips := w.Trips()

	bundle := stats.Bundle{
		SimulationReport: w.Stats.BuildSimulationReport(stats.RunConfig{
			Drivers: e.cfg.Drivers, Riders: e.cfg.Riders, Days: e.cfg.Days, Model: e.cfg.Model,
		}, trips),
		DriverStatistics:    w.Stats.BuildDriverStatistics(w.Drivers),
		RiderStatistics:     w.Stats.BuildRiderStatistics(w.Riders),
		PricingStatistics:   w.Stats.BuildPricingStatistics(trips),
		PromotionStatistics: w.Stats.BuildPromotionStatistics(trips),
		VoucherStatistics:   w.Stats.BuildVoucherStatistics(trips),
		DispatchStatistics:  w.Stats.BuildDispatchStatistics(trips),
		DeliveryStatistics:  w.Stats.BuildDeliveryStatistics(trips),
		Heatmap:             w.Stats.BuildHeatmap(trips, w.SupplyByZoneHour()),
		UnitEconomics:       w.Stats.BuildUnitEconomics(trips),
		DriverAnalytics:     w.Stats.BuildDriverAnalytics(w.Drivers, trips, w.driverOffersAccepted, w.driverOffersRejected, e.cfg.Days),
		RiderAnalytics:      w.Stats.BuildRiderAnalytics(w.Riders, trips, e.cfg.Days),
		PricingAnalytics:    w.Stats.BuildPricingAnalytics(trips),
		PromotionROI:        w.Stats.BuildPromotionROI(w.Riders, trips, w.voucherIssuedCount, w.voucherUsedCount, w.voucherKeptCount),
	}
	bundle.VoucherStatistics.UsedCount = w.voucherUsedCount
	bundle.VoucherStatistics.KeptCount = w.voucherKeptCount
	if err := bundle.Export(dir); err != nil {
		return fmt.Errorf("export statistics: %w", err)
	}

	bi := w.Stats.BuildBusinessIntelligence(trips, bundle.DriverAnalytics.RetentionRatePercent, bundle.RiderAnalytics.RetentionRatePercent)
	if err := writeExecutiveDashboard(dir, bi, bundle); err != nil {
		return err
	}
	findings, recs, err := writeInsightReports(context.Background(), dir, w.AI, trips, bundle, bi)
	if err != nil {
		return err
	}
	validation, err := writeValidationReport(dir, w, trips, bi)
	if err != nil {
		return err
	}
	if err := writeBusinessAudit(dir, w, trips, bundle, bi, validation, findings, recs); err != nil {
		return err
	}
	if err := writeBusinessIntelligence(dir, w, e.cfg, trips, bundle, bi, recs); err != nil {
		return err
	}

	aiStats := w.AI.StatsSnapshot()
	totalLookups := aiStats.CacheHits + aiStats.AICalls + aiStats.RuleFallbackUsed
	var hitPercent float64
	if totalLookups > 0 {
		hitPercent = 100 * float64(aiStats.CacheHits) / float64(totalLookups)
	}
	report := w.Bench.Build(benchmark.AIEngineStats{
		AICalls: aiStats.AICalls, CacheHits: aiStats.CacheHits, CacheHitPercent: hitPercent,
		CacheSize: aiStats.CacheSize, RuleFallbackUsed: aiStats.RuleFallbackUsed,
		Timeouts: aiStats.Timeouts, AvgAILatencyMS: aiStats.AvgAILatencyMS, CircuitOpen: aiStats.CircuitOpen,
	})
	if err := writeBenchmarkReport(dir, report); err != nil {
		return err
	}

	return writeDashboard(dir, bundle)
}

func onNewDay(w *World) {
	for _, d := range w.Drivers {
		if d.Online || d.HoursOnlineToday > 0 {
			d.DaysActive++
		}
		d.IncomeToday = 0
		d.HoursOnlineToday = 0
	}
	if w.Clock.Day()%7 == 0 {
		for _, d := range w.Drivers {
			d.IncomeWeek = 0
		}
	}
}

// processTick generates ride requests for this minute, evaluates periodic
// driver fatigue/surge-chase decisions, and lets online idle drivers drift
// toward higher-demand zones.
func processTick(ctx context.Context, w *World) {
	demandMult := DemandMultiplier(w)

	// Ride-request generation: each rider has a small independent
	// probability of requesting a trip this minute, scaled by their habit's
	// typical activity window and the current scenario demand multiplier.
	// Expected value is calibrated so a rider with average habits requests
	// roughly 1-3 trips across a simulated day, not every tick.
	for _, riderID := range w.RiderIDs {
		rider := w.Riders[riderID]
		if rider.CurrentTripID != "" {
			continue
		}
		baseProb := 0.0009 * demandMult
		if w.Rand.Float64() >= baseProb {
			continue
		}
		var trip *entity.SimTrip
		if w.Rand.Float64() < DeliveryProbability(w, rider.Zone) {
			rider.CurrentTripID = "pending-delivery"
			trip = ProcessDeliveryRequest(ctx, w, rider)
		} else {
			tripID := fmt.Sprintf("trip-%d-%s", w.Clock.Tick, rider.ID)
			rider.CurrentTripID = tripID
			trip = ProcessRideRequest(ctx, w, rider, tripID)
		}
		w.Bench.RecordTrip()
		applyTripOutcomeToSatisfaction(rider, trip)
		rider.CurrentTripID = ""
	}

	// Periodic per-driver decisions — every 15 simulated minutes per driver
	// is enough to model fatigue/relocation choices realistically without
	// evaluating the rule engine 500 times every single minute.
	if w.Clock.Tick%15 == 0 {
		for _, driverID := range w.DriverIDs {
			evaluateDriverState(ctx, w, w.Drivers[driverID])
		}
	}
}

// applyTripOutcomeToSatisfaction is a simulation-only behavioral signal (not
// a BRB rule): a completed trip with a short pickup wait nudges a rider's
// implicit satisfaction upward (reflected in whether they keep choosing
// Panda over the simulated competitor in future SwitchAppDecision calls,
// via PriceSensitivity/Patience drift), a rejected/cancelled trip nudges it
// down.
func applyTripOutcomeToSatisfaction(rider *entity.RiderAgent, trip *entity.SimTrip) {
	switch trip.Outcome {
	case entity.OutcomeCompleted:
		if trip.PickupMinutes <= 5 {
			rider.Patience = clamp01(rider.Patience + 0.01)
		}
	case entity.OutcomeRejected:
		rider.Patience = clamp01(rider.Patience - 0.03)
	}
}

// evaluateDriverState runs the two driver-side AI-eligible decisions
// (Fatigue continue/stop, Surge-chase relocate/stay) plus fatigue/battery/
// fuel drift for one driver.
func evaluateDriverState(ctx context.Context, w *World, d *entity.DriverAgent) {
	if d.Online {
		d.HoursOnlineToday += 0.25
		d.TotalHoursOnline += 0.25
		if d.HoursOnlineToday > d.MaxHoursOnlineContinuous {
			d.MaxHoursOnlineContinuous = d.HoursOnlineToday
		}
		d.Fatigue = clamp01(d.Fatigue + d.FatigueGainPerTick)
		d.PhoneBattery = clamp01(d.PhoneBattery - 0.006)
		d.Fuel = clamp01(d.Fuel - 0.01)
	} else {
		// Recovery while offline: fatigue eases, phone charges, tank/battery
		// refills — without this, a driver who ever hits FatigueDecision's
		// Stop floor would stay offline for the rest of the simulation, since
		// nothing else ever brings them back online.
		d.Fatigue = clamp01(d.Fatigue - 0.02)
		d.PhoneBattery = clamp01(d.PhoneBattery + 0.02)
		d.Fuel = clamp01(d.Fuel + 0.015)
		tryStartShift(ctx, w, d)
	}

	if d.Online && d.IsAvailable() {
		fatigueOutcome := ruleengine.FatigueDecision(ruleengine.FatigueInput{
			Fatigue: d.Fatigue, HoursOnlineToday: d.HoursOnlineToday,
			PhoneBattery: d.PhoneBattery, Fuel: d.Fuel,
			IncomeTodayVND: d.IncomeToday, DailyTargetVND: dailyTargetForTier(d.AccountType),
		})
		decision := resolveFatigue(ctx, w, d, fatigueOutcome)
		if decision == ruleengine.DecisionStop {
			d.Online = false
			_ = w.Dispatch.RemoveDriver(ctx, d.ID)
			return
		}

		if w.Rand.Float64() < 0.3 { // not every idle tick considers chasing surge — avoids constant relocation churn
			considerSurgeChase(ctx, w, d)
		}
	}

	// Publish current position for dispatch matching regardless of the
	// decisions above (position may have changed via surge chase).
	if d.Online {
		pos := w.City.Zones[d.Zone]
		_ = w.Dispatch.SetDriverPosition(ctx, d.ID, pos.X, pos.Y, d.ServiceType, d.RideEnabled, d.DeliveryEnabled)
	}
}

// tryStartShift is the counterpart to the Stop decision in evaluateDriverState:
// an offline driver who has rested (low fatigue) and has enough phone/fuel to
// work again probabilistically starts a new shift. A flat 8% chance per
// 15-simulated-minute check gives an expected ~1h52m offline break, a
// simulation-design constant (not a BRB rule — no real shift-start behavior
// data exists to sample from).
func tryStartShift(ctx context.Context, w *World, d *entity.DriverAgent) {
	if d.Fatigue > 0.3 || d.PhoneBattery < 0.3 || d.Fuel < 0.3 {
		return
	}
	if w.Rand.Float64() >= 0.08 {
		return
	}
	d.Online = true
	d.HoursOnlineToday = 0
	pos := w.City.Zones[d.Zone]
	_ = w.Dispatch.SetDriverPosition(ctx, d.ID, pos.X, pos.Y, d.ServiceType, d.RideEnabled, d.DeliveryEnabled)
}

func resolveFatigue(ctx context.Context, w *World, d *entity.DriverAgent, outcome ruleengine.Outcome) ruleengine.Decision {
	decision := outcome.Decision
	if !outcome.NeedsAI {
		w.recordDecision(false)
	} else {
		prompt := aiengine.FatiguePrompt(
			aiengine.BucketLevel3(d.Fatigue), int(d.HoursOnlineToday),
			aiengine.BucketIncomeProgress(d.IncomeToday, dailyTargetForTier(d.AccountType)),
			d.IncomeToday >= dailyTargetForTier(d.AccountType),
		)
		result := w.AI.Decide(ctx, aiengine.DecisionRequest{
			Prompt:   prompt,
			Parse:    aiengine.ParseBinary("CONTINUE", ruleengine.DecisionContinue, "STOP", ruleengine.DecisionStop),
			Fallback: outcome.Decision,
		})
		w.recordDecision(true)
		decision = result.Decision
	}
	w.recordFatigueOutcome(decision == ruleengine.DecisionContinue)
	return decision
}

func considerSurgeChase(ctx context.Context, w *World, d *entity.DriverAgent) {
	bestZone, bestMultiplier := d.Zone, 1.0
	for _, z := range entity.AllZoneTypes() {
		if z == d.Zone {
			continue
		}
		dsr := demandSupplyRatioForZone(w, z)
		mult := dsrMultiplier(dsr)
		if mult > bestMultiplier {
			bestMultiplier, bestZone = mult, z
		}
	}
	if bestZone == d.Zone {
		return
	}

	outcome := ruleengine.SurgeChaseDecision(ruleengine.SurgeChaseInput{
		SurgeMultiplier: bestMultiplier, DistanceKM: w.City.DistanceKM(d.Zone, bestZone),
		Fatigue: d.Fatigue, Fuel: d.Fuel,
		IncomeTodayVND: d.IncomeToday, DailyTargetVND: dailyTargetForTier(d.AccountType),
	})
	decision := outcome.Decision
	if outcome.NeedsAI {
		prompt := aiengine.SurgeChasePrompt(bestMultiplier, w.City.DistanceKM(d.Zone, bestZone), aiengine.BucketLevel3(d.Fatigue), d.IncomeToday >= dailyTargetForTier(d.AccountType))
		result := w.AI.Decide(ctx, aiengine.DecisionRequest{
			Prompt:   prompt,
			Parse:    aiengine.ParseBinary("CHASE", ruleengine.DecisionChaseSurge, "STAY", ruleengine.DecisionStayPut),
			Fallback: outcome.Decision,
		})
		w.recordDecision(true)
		decision = result.Decision
	} else {
		w.recordDecision(false)
	}

	w.recordSurgeChaseOutcome(decision == ruleengine.DecisionChaseSurge)
	if decision == ruleengine.DecisionChaseSurge {
		d.Zone = bestZone
	}
}

func demandSupplyRatioForZone(w *World, z entity.ZoneType) float64 {
	drivers := w.ZoneAvailableDrivers[z]
	if drivers == 0 {
		if w.ZoneActiveRequests[z] == 0 {
			return 0
		}
		return 999
	}
	return float64(w.ZoneActiveRequests[z]) / float64(drivers)
}

// dsrMultiplier looks up BRB §2.13.2's exact Demand-Supply Ratio table (the
// same one backend/services/pricing's DemandSurgeRule uses) to translate a
// zone's live DSR into the surge multiplier a driver would actually see
// there — reused via the real pricing entity package rather than
// duplicating the table.
func dsrMultiplier(dsr float64) float64 {
	best := 1.0
	for _, tier := range pricingentity.DefaultDSRTiers() {
		if dsr >= tier.MinDSR && tier.Multiplier > best {
			best = tier.Multiplier
		}
	}
	return best
}

// dailyTargetForTier is a simulation-only convenience (not a BRB number):
// higher-tier drivers are modeled as having a somewhat higher personal
// daily income goal, reflecting greater platform reliance/experience.
func dailyTargetForTier(tier entity.AccountType) int64 {
	switch tier {
	case entity.AccountDiamond, entity.AccountPlatinum:
		return 900_000
	case entity.AccountGold, entity.AccountSilver:
		return 700_000
	default:
		return 500_000
	}
}

// sortedKeys returns m's keys in a fixed, deterministic order — the fix for
// the --seed non-determinism bug (see World.DriverIDs/RiderIDs doc comment).
func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}
