package simulation

// This file implements BƯỚC 5 (Optimization). It never modifies BRB's
// approved rates (pricing_constants.go stays the single source of truth for
// what this package treats as "current") — it sweeps a small grid of
// ALTERNATIVE booking-fee/commission configurations and scores each one
// against the sprint's three ordered priorities, so the result is a
// recommendation to put in front of the CPO/CFO, not a change this package
// applies to itself.
//
// Ordered priorities (BƯỚC 5, verbatim):
//   1. Driver kiếm nhiều hơn đối thủ         (maximise DriverEarnsMoreCount)
//   2. Khách trả thấp hơn đối thủ            (maximise CheaperThanMarketCount)
//   3. Platform hòa vốn hoặc lời nhẹ         (AggregateMarginPct close to 0, never far negative)
// Explicitly NOT optimised: platform profit maximisation.

// ParameterCandidate is one alternative (BookingFee, Bronze-commission-anchor)
// configuration to test. The commission *spread* between tiers (BRB §7.1:
// Bronze→Diamond is always −2 percentage points per tier) is preserved —
// only the Bronze anchor moves, so a candidate is still a single coherent
// tier ladder, not an arbitrary flat rate.
type ParameterCandidate struct {
	Label            string
	BookingFeeVND    int64
	BronzeCommission float64
}

// DefaultCandidateGrid — small, deliberately narrow grid centered on the
// BRB-approved values (BookingFee=2,000 VND, Bronze commission=20%) so every
// candidate is a plausible adjustment, not a wild guess.
func DefaultCandidateGrid() []ParameterCandidate {
	fees := []int64{1_500, 2_000, 2_500, 3_000}
	commissions := []float64{0.16, 0.18, 0.20, 0.22}
	var grid []ParameterCandidate
	for _, f := range fees {
		for _, c := range commissions {
			grid = append(grid, ParameterCandidate{
				Label:            "Fee" + moneyLabel(f) + "_Bronze" + percentLabel(c),
				BookingFeeVND:    f,
				BronzeCommission: c,
			})
		}
	}
	return grid
}

// commissionRateForCandidate reproduces the BRB §7.1 tier ladder (a fixed
// −2pp step per tier) anchored at the candidate's Bronze rate, instead of
// the package-level CommissionRate(), so the optimizer can test an anchor
// shift without touching the real constant table.
func commissionRateForCandidate(cand ParameterCandidate, tier DriverTier) float64 {
	step := 0.02
	switch tier {
	case TierSilver:
		return cand.BronzeCommission - step
	case TierGold:
		return cand.BronzeCommission - 2*step
	case TierPlatinum:
		return cand.BronzeCommission - 3*step
	case TierDiamond:
		return cand.BronzeCommission - 4*step
	default:
		return cand.BronzeCommission
	}
}

// simulateWithCandidate re-runs one TripInput with a candidate's
// BookingFee/commission substituted for the defaults, by cloning the
// simulator's rate table and post-adjusting the commission-dependent fields.
// It intentionally reuses Simulator.Simulate for every BRB-approved rule
// (surge/surcharges/rounding/safety) and only overrides the two knobs the
// optimizer is allowed to move.
func simulateWithCandidate(sim *Simulator, cand ParameterCandidate, in TripInput) (*FareBreakdown, error) {
	adjustedRates := make(map[VehicleType]VehicleRates, len(sim.Rates))
	for k, v := range sim.Rates {
		adjustedRates[k] = v
	}
	adjusted := &Simulator{Rates: adjustedRates, PeakWindows: sim.PeakWindows, SurgeBands: sim.SurgeBands}

	fb, err := adjusted.Simulate(in)
	if err != nil {
		return nil, err
	}

	// Re-derive every field downstream of BookingFee/Commission using the
	// candidate's values instead of the package defaults.
	fb.ServiceFee = cand.BookingFeeVND
	preDiscountTotal := fb.RideFare + fb.AirportFeeApplied + fb.WaitingFee + fb.BridgeFee + fb.ParkingFee + fb.ServiceFee
	fb.PromotionApplied = clampInt64(fb.PromotionRequested, 0, preDiscountTotal)
	fb.CustomerTotal = roundUpToUnit(preDiscountTotal-fb.PromotionApplied, RiderRoundingUnit)

	commissionBase := fb.RideFare + fb.AirportFeeApplied + fb.WaitingFee
	rate := commissionRateForCandidate(cand, in.DriverTier)
	fb.Commission = roundVND(float64(commissionBase) * rate)
	driverShareOfMetered := commissionBase - fb.Commission
	fb.DriverIncomeGross = driverShareOfMetered + fb.BridgeFee + fb.ParkingFee
	fb.MinimumEarningTopUp = 0
	if fb.DriverIncomeGross < MinimumDriverEarningVND {
		fb.MinimumEarningTopUp = MinimumDriverEarningVND - fb.DriverIncomeGross
	}
	fb.NetDriver = roundUpToUnit(fb.DriverIncomeGross+fb.MinimumEarningTopUp+fb.LongPickupCompensation, DriverRoundingUnit)

	platformRevenueGross := fb.Commission + fb.ServiceFee
	platformAfterPromo := platformRevenueGross - fb.PromotionApplied
	platformAfterProtection := platformAfterPromo - fb.MinimumEarningTopUp - fb.LongPickupCompensation - AssumedInsuranceCostVND
	vatBase := platformAfterProtection
	if vatBase < 0 {
		vatBase = 0
	}
	fb.VAT = roundVND(float64(vatBase) * AssumedVATRate)
	fb.PlatformRevenue = platformAfterProtection
	fb.Profit = platformAfterProtection - fb.VAT
	if fb.CustomerTotal > 0 {
		fb.MarginPct = float64(fb.Profit) / float64(fb.CustomerTotal)
	}
	applySafetyClamps(fb)
	return fb, nil
}

// CandidateScore is the aggregate result of running one candidate against
// every scenario and every competitor.
type CandidateScore struct {
	Candidate              ParameterCandidate
	AggCustomerTotal       int64
	AggDriverIncome        int64
	AggPlatformProfit      int64
	AggMarginPct           float64
	CheaperThanMarketCount int // summed across all scenarios × 8 competitors
	DriverEarnsMoreCount   int
	TotalComparisons       int
}

// Optimize runs BƯỚC 5: every candidate in grid against every scenario,
// aggregates the three ordered objectives, and returns candidates sorted
// best-first. "Best" is lexicographic: (1) DriverEarnsMoreCount desc,
// (2) CheaperThanMarketCount desc, (3) |AggMarginPct| asc among candidates
// with AggMarginPct ≥ 0 (breakeven-or-slightly-profitable preferred over
// both a loss and over a large profit — BƯỚC 5 explicitly rules out
// maximising profit).
func Optimize(scenarios []NamedScenario, grid []ParameterCandidate) []CandidateScore {
	sim := NewDefaultSimulator()
	scores := make([]CandidateScore, 0, len(grid))

	for _, cand := range grid {
		var score CandidateScore
		score.Candidate = cand
		for _, sc := range scenarios {
			fb, err := simulateWithCandidate(sim, cand, sc.Input)
			if err != nil {
				continue
			}
			score.AggCustomerTotal += fb.CustomerTotal
			score.AggDriverIncome += fb.NetDriver
			score.AggPlatformProfit += fb.Profit

			rows := CompareToMarket(fb)
			for _, r := range rows {
				score.TotalComparisons++
				if fb.CustomerTotal <= r.EstimatedCustomerTotal {
					score.CheaperThanMarketCount++
				}
				if fb.NetDriver >= r.EstimatedDriverIncome {
					score.DriverEarnsMoreCount++
				}
			}
		}
		if score.AggCustomerTotal > 0 {
			score.AggMarginPct = float64(score.AggPlatformProfit) / float64(score.AggCustomerTotal)
		}
		scores = append(scores, score)
	}

	sortScoresByPriority(scores)
	return scores
}

func sortScoresByPriority(scores []CandidateScore) {
	// Simple insertion sort — grid is small (16 candidates by default), no
	// need for sort.Slice's extra import weight in a file whose entire
	// point is being easy to audit line by line.
	for i := 1; i < len(scores); i++ {
		j := i
		for j > 0 && less(scores[j], scores[j-1]) {
			scores[j], scores[j-1] = scores[j-1], scores[j]
			j--
		}
	}
}

// less reports whether a should be ranked ahead of b.
func less(a, b CandidateScore) bool {
	if a.DriverEarnsMoreCount != b.DriverEarnsMoreCount {
		return a.DriverEarnsMoreCount > b.DriverEarnsMoreCount
	}
	if a.CheaperThanMarketCount != b.CheaperThanMarketCount {
		return a.CheaperThanMarketCount > b.CheaperThanMarketCount
	}
	aBreakeven := a.AggMarginPct >= 0
	bBreakeven := b.AggMarginPct >= 0
	if aBreakeven != bBreakeven {
		return aBreakeven // a candidate that doesn't lose money in aggregate always outranks one that does
	}
	return absFloat(a.AggMarginPct) < absFloat(b.AggMarginPct)
}

func absFloat(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}

func percentLabel(v float64) string {
	return moneyLabel(int64(v*100)) + "pct"
}
