package entity

// CommissionTier is the driver performance tier from BRB §7.1 (Bronze →
// Diamond, assigned by the Driver Performance Engine — BRB Part 9 — based on
// Acceptance Rate/Completion Rate/Rating/Trip Volume/Trust Score). Pricing
// V3 is the first time the Pricing service itself computes a driver's
// commission share — see docs/business/PRICING_SIMULATION_REPORT.md §1.3
// ("Không có hoa hồng/commission nào trong service" — the production
// Pricing Service never had this before this sprint).
type CommissionTier string

const (
	CommissionTierBronze   CommissionTier = "bronze"
	CommissionTierSilver   CommissionTier = "silver"
	CommissionTierGold     CommissionTier = "gold"
	CommissionTierPlatinum CommissionTier = "platinum"
	CommissionTierDiamond  CommissionTier = "diamond"
)

// CommissionConfigV3 maps each tier to its take rate (BRB §7.1: 20% Bronze
// down to 12% Diamond by default) — loaded from config, never hardcoded in
// Go. Unlike backend/services/pricing/simulation/pricing_constants.go's
// CommissionRate function (a hardcoded switch statement, acceptable there
// because the simulation package is explicitly isolated from production —
// see its own package doc), production Pricing V3 must read this from a
// config map so a rate change never requires a code change or release,
// matching docs/business/ECONOMY_ENGINE.md Phần 11 ("Rule Engine — không
// hardcode").
type CommissionConfigV3 struct {
	RateByTier map[CommissionTier]float64
}

// Rate returns the configured commission rate for tier. Unknown/unconfigured
// tiers fail closed to the Bronze rate (the platform's default, most
// conservative-for-the-driver entry rate — BRB §7.1) rather than 0% (which
// would silently give the platform no revenue) or 100% (which would
// silently take a driver's entire fare) — both of those failure directions
// are worse than "treat as a brand-new driver," which is what an
// unrecognised tier value actually implies operationally.
func (c CommissionConfigV3) Rate(tier CommissionTier) float64 {
	if rate, ok := c.RateByTier[tier]; ok {
		return rate
	}
	return c.RateByTier[CommissionTierBronze]
}
