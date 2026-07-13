package app

import "github.com/fairride/pricing/domain/entity"

// PricingVersion selects which fare engine VersionedFareCalculator dispatches
// to — the sprint's PHẦN 17 Migration/Feature Flag requirement ("Không bật
// Pricing V3 mặc định... pricing.version v2/v3... để rollback tức thì").
type PricingVersion string

const (
	// PricingVersionV2 uses the pre-existing FareCalculator (fare_calculator.go,
	// completely unmodified by this sprint) — the DEFAULT, per "không bật
	// Pricing V3 mặc định".
	PricingVersionV2 PricingVersion = "v2"
	// PricingVersionV3 uses FareCalculatorV3 (fare_calculator_v3.go).
	PricingVersionV3 PricingVersion = "v3"
)

// VersionedFareCalculator is what cmd/server/main.go and grpc/handler.go
// actually hold — a single calculator whose *behaviour* (not API) switches
// on Version. Its exported methods return the exact same
// *entity.FareBreakdown shape the gRPC handler has always consumed (see
// grpc/handler.go), so switching Version is a pure runtime toggle: no
// caller code changes, no proto change, and (per PHẦN 17) an instant
// rollback by flipping the flag back to v2.
type VersionedFareCalculator struct {
	Version PricingVersion
	v2      *FareCalculator
	v3      *FareCalculatorV3
}

// NewVersionedFareCalculator builds the dispatcher. An empty/unrecognised
// version fails closed to v2 — the same "never silently activate an
// unreviewed code path" discipline RuleConfigMap.Get already uses for
// disabled rules.
func NewVersionedFareCalculator(version PricingVersion, v2 *FareCalculator, v3 *FareCalculatorV3) *VersionedFareCalculator {
	if version != PricingVersionV3 {
		version = PricingVersionV2
	}
	return &VersionedFareCalculator{Version: version, v2: v2, v3: v3}
}

// Estimate dispatches to v2 or v3 depending on Version, always returning the
// V2-shaped entity.FareBreakdown (see downgradeToFareBreakdown for how a V3
// result maps onto it without loss of the fields V2 callers already read:
// Total, CurrencyCode, per-component amounts).
func (c *VersionedFareCalculator) Estimate(vehicleType entity.VehicleType, distanceKM, durationMin float64) (*entity.FareBreakdown, error) {
	if c.Version == PricingVersionV3 && c.v3 != nil {
		full, err := c.v3.EstimateV3(entity.RideInputV3{
			VehicleType:    vehicleType,
			DistanceKM:     distanceKM,
			DurationMin:    durationMin,
			CommissionTier: entity.CommissionTierBronze,
		})
		if err != nil {
			return nil, err
		}
		return downgradeToFareBreakdown(full), nil
	}
	return c.v2.Estimate(vehicleType, distanceKM, durationMin)
}

// CalculateFinal mirrors Estimate for the post-trip final fare.
func (c *VersionedFareCalculator) CalculateFinal(vehicleType entity.VehicleType, distanceKM, durationMin float64) (*entity.FareBreakdown, error) {
	if c.Version == PricingVersionV3 && c.v3 != nil {
		full, err := c.v3.CalculateFinalV3(entity.RideInputV3{
			VehicleType:    vehicleType,
			DistanceKM:     distanceKM,
			DurationMin:    durationMin,
			CommissionTier: entity.CommissionTierBronze,
		})
		if err != nil {
			return nil, err
		}
		return downgradeToFareBreakdown(full), nil
	}
	return c.v2.CalculateFinal(vehicleType, distanceKM, durationMin)
}

// EstimateV3Detailed exposes the full V3 breakdown (Explanation, Commission,
// VAT, Driver Income, Platform Revenue...) for callers that know they want
// it — e.g. internal tooling, or a future proto extension (see
// docs/business/PRICING_V3_IMPLEMENTATION.md "Next Phase"). Returns an error
// if the calculator is not running in v3 mode, so a caller can't
// accidentally read a V3 breakdown while the platform is configured for v2.
func (c *VersionedFareCalculator) EstimateV3Detailed(input entity.RideInputV3) (*entity.FullFareBreakdownV3, error) {
	if c.Version != PricingVersionV3 || c.v3 == nil {
		return nil, errPricingV3NotActive
	}
	return c.v3.EstimateV3(input)
}

// CalculateFinalDetailed mirrors EstimateV3Detailed for the post-trip final
// fare — exposes Commission, DriverIncome, VoucherDiscount, CommissionRate
// etc. for callers (Settlement) that must never invent their own commission
// number. Returns an error if the calculator is not running in v3 mode, so a
// caller can't accidentally read a V3 breakdown while configured for v2.
func (c *VersionedFareCalculator) CalculateFinalDetailed(input entity.RideInputV3) (*entity.FullFareBreakdownV3, error) {
	if c.Version != PricingVersionV3 || c.v3 == nil {
		return nil, errPricingV3NotActive
	}
	return c.v3.CalculateFinalV3(input)
}

var errPricingV3NotActive = fareCalculatorV3NotActiveError{}

type fareCalculatorV3NotActiveError struct{}

func (fareCalculatorV3NotActiveError) Error() string {
	return "pricing: PricingVersionV3 is not active for this calculator (PRICING_VERSION env/config is not \"v3\")"
}

// downgradeToFareBreakdown maps a FullFareBreakdownV3 onto the pre-existing
// V2 wire shape. Total uses FinalFare (RideFare + WaitingFee + PlatformFee -
// VoucherDiscount) — the actual amount the rider pays — rather than
// RideFare+BookingFee alone, since V2's Total always meant "what the rider
// pays" and V3 adds components (Waiting Fee, Voucher) that legitimately
// change that amount.
func downgradeToFareBreakdown(full *entity.FullFareBreakdownV3) *entity.FareBreakdown {
	return &entity.FareBreakdown{
		VehicleType:  full.VehicleType,
		DistanceKM:   full.DistanceKM,
		DurationMin:  full.DurationMin,
		BaseFare:     full.BaseFare,
		DistanceFare: full.DistanceFare,
		TimeFare:     full.TrafficTimeFare,
		BookingFee:   full.PlatformFee,
		RideFare:     full.RideFare,
		Total:        full.FinalFare,
		CurrencyCode: full.CurrencyCode,
		IsFinal:      full.IsFinal,
	}
}
