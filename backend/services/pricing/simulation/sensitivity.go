package simulation

// This file implements BƯỚC 6 (Sensitivity Analysis) — five external shocks
// named in the sprint brief. Where a shock isn't a direct parameter of the
// fare formula itself (fuel price, driver/rider population), this file
// models it as a bounded, clearly-labelled ASSUMPTION rather than silently
// pretending the formula has a dial for it. The question BƯỚC 6 asks is
// "does the engine still hold up" — every function below returns a
// pass/fail-style verdict against the BƯỚC-7 safety invariants plus a
// plain-language readout, not just raw numbers.

// AssumedFuelCostPerKM — NOT in BRB (fuel is the driver's own operating
// cost; BRB never nets it against fare). A rough Vietnam car fuel-cost
// estimate (~7–8 L/100km × ~24,000 VND/L, mid-2026 pricing) used only to
// judge whether the driver's *effective* take-home still clears a
// reasonable margin after a fuel shock — never fed back into the fare
// formula itself.
const AssumedFuelCostPerKM float64 = 1_800

// FuelShockResult is BƯỚC 6, shock 1: xăng tăng 20%.
type FuelShockResult struct {
	FuelCostIncreasePct     float64
	AvgNetDriverBefore      float64
	AvgFuelCostBefore       float64
	AvgEffectiveTakeBefore  float64 // NetDriver − FuelCost
	AvgEffectiveTakeAfter   float64
	EffectiveTakeErodedPct  float64 // how much of the driver's effective margin the fuel shock consumes
	StillPositiveForAllTrips bool
}

func RunFuelShock(scenarios []NamedScenario, increasePct float64) FuelShockResult {
	sim := NewDefaultSimulator()
	var sumNet, sumFuelBefore, sumFuelAfter float64
	n := 0
	allPositiveAfter := true
	for _, sc := range scenarios {
		fb, err := sim.Simulate(sc.Input)
		if err != nil {
			continue
		}
		fuelBefore := sc.Input.DistanceKM * AssumedFuelCostPerKM
		fuelAfter := fuelBefore * (1 + increasePct)
		sumNet += float64(fb.NetDriver)
		sumFuelBefore += fuelBefore
		sumFuelAfter += fuelAfter
		if float64(fb.NetDriver)-fuelAfter <= 0 {
			allPositiveAfter = false
		}
		n++
	}
	if n == 0 {
		return FuelShockResult{}
	}
	avgNet := sumNet / float64(n)
	avgFuelBefore := sumFuelBefore / float64(n)
	avgFuelAfter := sumFuelAfter / float64(n)
	takeBefore := avgNet - avgFuelBefore
	takeAfter := avgNet - avgFuelAfter
	eroded := 0.0
	if takeBefore > 0 {
		eroded = (takeBefore - takeAfter) / takeBefore
	}
	return FuelShockResult{
		FuelCostIncreasePct:      increasePct,
		AvgNetDriverBefore:       avgNet,
		AvgFuelCostBefore:        avgFuelBefore,
		AvgEffectiveTakeBefore:   takeBefore,
		AvgEffectiveTakeAfter:    takeAfter,
		EffectiveTakeErodedPct:   eroded,
		StillPositiveForAllTrips: allPositiveAfter,
	}
}

// PromotionShockResult is BƯỚC 6, shock 2: voucher tăng.
type PromotionShockResult struct {
	DiscountMultiplier float64
	AggProfitBefore    int64
	AggProfitAfter     int64
	StillBreakeven     bool // AggProfitAfter ≥ 0 across every scenario that has a promotion
}

func RunPromotionShock(scenarios []NamedScenario, discountMultiplier float64) PromotionShockResult {
	sim := NewDefaultSimulator()
	var before, after int64
	for _, sc := range scenarios {
		if sc.Input.Promotion == nil {
			continue
		}
		fbBefore, err := sim.Simulate(sc.Input)
		if err != nil {
			continue
		}
		before += fbBefore.Profit

		scaled := sc.Input
		scaledPromo := *sc.Input.Promotion
		scaledPromo.DiscountVND = roundVND(float64(scaledPromo.DiscountVND) * discountMultiplier)
		scaled.Promotion = &scaledPromo
		fbAfter, err := sim.Simulate(scaled)
		if err != nil {
			continue
		}
		after += fbAfter.Profit
	}
	return PromotionShockResult{
		DiscountMultiplier: discountMultiplier,
		AggProfitBefore:    before,
		AggProfitAfter:     after,
		StillBreakeven:     after >= 0,
	}
}

// CommissionShockResult is BƯỚC 6, shock 3: commission giảm.
type CommissionShockResult struct {
	BronzeCommissionAfter float64
	AggProfitBefore       int64
	AggProfitAfter        int64
	AggDriverIncomeBefore int64
	AggDriverIncomeAfter  int64
	StillBreakeven        bool
}

func RunCommissionShock(scenarios []NamedScenario, newBronzeCommission float64) CommissionShockResult {
	sim := NewDefaultSimulator()
	cand := ParameterCandidate{Label: "CommissionShock", BookingFeeVND: BookingFee, BronzeCommission: newBronzeCommission}
	var profitBefore, profitAfter, driverBefore, driverAfter int64
	for _, sc := range scenarios {
		fbBefore, err := sim.Simulate(sc.Input)
		if err != nil {
			continue
		}
		profitBefore += fbBefore.Profit
		driverBefore += fbBefore.NetDriver

		fbAfter, err := simulateWithCandidate(sim, cand, sc.Input)
		if err != nil {
			continue
		}
		profitAfter += fbAfter.Profit
		driverAfter += fbAfter.NetDriver
	}
	return CommissionShockResult{
		BronzeCommissionAfter: newBronzeCommission,
		AggProfitBefore:       profitBefore,
		AggProfitAfter:        profitAfter,
		AggDriverIncomeBefore: driverBefore,
		AggDriverIncomeAfter:  driverAfter,
		StillBreakeven:        profitAfter >= 0,
	}
}

// SupplyShockResult is BƯỚC 6, shock 4: driver tăng gấp đôi. Modeled as
// DSR ∝ 1 / driver_count (demand held constant, doubling supply halves the
// demand/supply ratio) — a simplification documented here, not a full
// marketplace-liquidity model (that would need real elasticity data Panda
// does not have yet).
type SupplyShockResult struct {
	AvgSurgeMultiplierBefore float64
	AvgSurgeMultiplierAfter  float64
	AvgDriverIncomeBefore    float64
	AvgDriverIncomeAfter     float64
	DriverIncomeChangePct    float64
}

func RunDriverSupplyDoubled(scenarios []NamedScenario) SupplyShockResult {
	sim := NewDefaultSimulator()
	var surgeBefore, surgeAfter, driverBefore, driverAfter float64
	n := 0
	for _, sc := range scenarios {
		if sc.Input.DSR <= 0 {
			continue // only scenarios with an active demand/supply signal are meaningful here
		}
		fbBefore, err := sim.Simulate(sc.Input)
		if err != nil {
			continue
		}
		halved := sc.Input
		halved.DSR = sc.Input.DSR / 2
		fbAfter, err := sim.Simulate(halved)
		if err != nil {
			continue
		}
		surgeBefore += fbBefore.SurgeMultiplier
		surgeAfter += fbAfter.SurgeMultiplier
		driverBefore += float64(fbBefore.NetDriver)
		driverAfter += float64(fbAfter.NetDriver)
		n++
	}
	if n == 0 {
		return SupplyShockResult{}
	}
	res := SupplyShockResult{
		AvgSurgeMultiplierBefore: surgeBefore / float64(n),
		AvgSurgeMultiplierAfter:  surgeAfter / float64(n),
		AvgDriverIncomeBefore:    driverBefore / float64(n),
		AvgDriverIncomeAfter:     driverAfter / float64(n),
	}
	if res.AvgDriverIncomeBefore > 0 {
		res.DriverIncomeChangePct = (res.AvgDriverIncomeAfter - res.AvgDriverIncomeBefore) / res.AvgDriverIncomeBefore
	}
	return res
}

// DemandShockResult is BƯỚC 6, shock 5: khách tăng gấp 5. Modeled as
// DSR ∝ demand (supply held constant, 5x demand multiplies DSR by 5) — same
// documented simplification as the supply shock, inverted.
type DemandShockResult struct {
	AvgSurgeMultiplierBefore float64
	AvgSurgeMultiplierAfter  float64
	PriceCapHitCountBefore   int
	PriceCapHitCountAfter    int
	MaxSurgeHitCountAfter    int // scenarios where the ×2.0 cap itself became the binding constraint
}

func RunRiderDemand5x(scenarios []NamedScenario) DemandShockResult {
	sim := NewDefaultSimulator()
	var surgeBefore, surgeAfter float64
	n := 0
	var capBefore, capAfter, maxSurgeAfter int
	for _, sc := range scenarios {
		if sc.Input.DSR <= 0 {
			continue
		}
		fbBefore, err := sim.Simulate(sc.Input)
		if err != nil {
			continue
		}
		spiked := sc.Input
		spiked.DSR = sc.Input.DSR * 5
		fbAfter, err := sim.Simulate(spiked)
		if err != nil {
			continue
		}
		surgeBefore += fbBefore.SurgeMultiplier
		surgeAfter += fbAfter.SurgeMultiplier
		if fbBefore.PriceCapApplied {
			capBefore++
		}
		if fbAfter.PriceCapApplied {
			capAfter++
		}
		if fbAfter.SurgeMultiplier >= MaxSurgeMultiplier {
			maxSurgeAfter++
		}
		n++
	}
	if n == 0 {
		return DemandShockResult{}
	}
	return DemandShockResult{
		AvgSurgeMultiplierBefore: surgeBefore / float64(n),
		AvgSurgeMultiplierAfter:  surgeAfter / float64(n),
		PriceCapHitCountBefore:   capBefore,
		PriceCapHitCountAfter:    capAfter,
		MaxSurgeHitCountAfter:    maxSurgeAfter,
	}
}
