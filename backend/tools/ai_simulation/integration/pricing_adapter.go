// Package integration adapts the simulation's own domain/entity types to the
// REAL production engines (pricing, promotion, dispatch) so the simulation
// calls the actual business logic instead of reimplementing it, per the
// sprint brief's "PHẢI DÙNG Pricing Engine, Promotion Engine, Voucher
// Engine, Dispatch, Driver Economy đang có sẵn. KHÔNG viết lại."
package integration

import (
	"time"

	pricingapp "github.com/fairride/pricing/app"
	pricingentity "github.com/fairride/pricing/domain/entity"

	"github.com/fairride/ai_simulation/domain/entity"
)

// PricingAdapter wraps the real backend/services/pricing FareCalculator +
// Dynamic Pricing Engine (see that service's CHANGELOG entry). The
// simulation is the first real caller to exercise the engine with several
// surge rules actually enabled — production still ships them disabled by
// default, per that engine's own backward-compatibility guarantee.
type PricingAdapter struct {
	calc *pricingapp.FareCalculator
}

// NewPricingAdapter builds a FareCalculator over BRB §2.2.1-§2.2.5's exact
// published VND rates, keyed by the simulation's own rate table — entirely
// separate from (and never loaded from) production's
// pricing_v3.default.yaml, so this carries no "never fabricate a
// production rate" obligation; it already had an established precedent for
// exactly this before the Vehicle/Service Catalog refactor (motorcycle's
// rate is derived, not a BRB number — see the entry below).
//
// pricingentity.VehicleType is declared `type VehicleType string` with no
// runtime enum enforcement, so this map is free to use "bike_plus"/"car_xl"
// as keys even though they aren't among production's 3 named VehicleType
// constants (car/motorcycle/van) — Rates is a plain map, and
// FareCalculator only ever does a map lookup, never a switch over the
// named consts. This gives the simulation 4 independently-priced service
// tiers while still routing every quote through the real Dynamic Pricing
// Engine (surge/peak/rain/night/holiday/airport rules, all enabled below).
//
//   - car/van/motorcycle: BRB §2.2.1-§2.2.5 Standard/XL, and motorcycle at
//     60% of car (typical SEA market ratio, not a BRB number) — the
//     existing, pre-refactor mapping (Pricing Simulation Report "Ghi chú
//     phương pháp").
//   - "bike_plus": simulation-only assumption, +20% over the motorcycle
//     rate (a newer-bike/higher-rated-driver premium) — no BRB number
//     exists for this tier at all yet.
//   - "car_xl": reuses the van rate verbatim — Car XL is the rider-facing
//     name for the same 7-seat/XL tier BRB's van rate already represents
//     (driver.ServiceType.RequiredVehicleType maps ServiceCarXL to
//     VehicleTypeVan for the same reason).
func NewPricingAdapter() *PricingAdapter {
	motorcycleRate := pricingentity.VehicleRates{
		BaseFare: 6_000, PerKmRate: 2_400, PerMinuteRate: 240,
		MinimumFare: 15_000, BookingFee: 2_000,
	}
	vanRate := pricingentity.VehicleRates{
		BaseFare: 18_000, PerKmRate: 5_000, PerMinuteRate: 500,
		MinimumFare: 40_000, BookingFee: 2_000,
	}
	config := pricingentity.FareConfig{
		CurrencyCode: "VND",
		Rates: map[pricingentity.VehicleType]pricingentity.VehicleRates{
			pricingentity.VehicleTypeCar: { // BRB §2.2.1-§2.2.5 Standard
				BaseFare: 10_000, PerKmRate: 4_000, PerMinuteRate: 400,
				MinimumFare: 25_000, BookingFee: 2_000,
			},
			pricingentity.VehicleTypeVan:        vanRate,        // BRB §2.2.1-§2.2.5 XL
			pricingentity.VehicleTypeMotorcycle: motorcycleRate, // 60% of car — assumption, not a BRB number
			pricingentity.VehicleType("bike_plus"): { // +20% over motorcycle — simulation-only assumption
				BaseFare: motorcycleRate.BaseFare * 12 / 10, PerKmRate: motorcycleRate.PerKmRate * 12 / 10,
				PerMinuteRate: motorcycleRate.PerMinuteRate * 12 / 10, MinimumFare: motorcycleRate.MinimumFare * 12 / 10,
				BookingFee: motorcycleRate.BookingFee,
			},
			pricingentity.VehicleType("car_xl"): vanRate, // same XL tier as van, different rider-facing name
		},
	}

	pipeline := pricingapp.NewDefaultPricingPipeline()
	configs := pricingapp.DefaultRuleConfigs()
	enable := func(name string, maxSurge float64) {
		cfg := configs[name]
		cfg.Enabled = true
		cfg.Weight = 1.0
		if cfg.MinSurge == 0 {
			cfg.MinSurge = 1.0
		}
		cfg.MaxSurge = maxSurge
		configs[name] = cfg
	}
	enable(pricingapp.RuleNameDemandSurge, pricingentity.MaxDemandSurgeMultiplier)
	enable(pricingapp.RuleNamePeakHour, pricingentity.PeakHourSurchargeMultiplier)
	enable(pricingapp.RuleNameNight, pricingentity.NightSurchargeMultiplier)
	enable(pricingapp.RuleNameHoliday, pricingentity.HolidaySurchargeMultiplier)
	enable(pricingapp.RuleNameRain, pricingentity.RainSurchargeMultiplier)
	// Airport is a flat-fee rule; MaxSurge caps the VND amount, not a ratio.
	airportCfg := configs[pricingapp.RuleNameAirport]
	airportCfg.Enabled = true
	airportCfg.Weight = 1.0
	airportCfg.MaxSurge = float64(pricingentity.AirportFeeVND)
	configs[pricingapp.RuleNameAirport] = airportCfg

	calc := pricingapp.NewFareCalculatorWithPipeline(config, pipeline, configs)
	return &PricingAdapter{calc: calc}
}

// QuoteInput carries the simulated signals the real Dynamic Pricing Engine
// needs — a direct translation of the simulation's own entities into
// pricing's entity.PricingContext, performing no pricing math itself.
type QuoteInput struct {
	ServiceType      entity.ServiceType
	DistanceKM       float64
	DurationMin      float64
	RequestTime      time.Time
	ActiveRequests   int
	AvailableDrivers int
	IsAirportZone    bool
	Weather          entity.Weather
	IsHoliday        bool
}

// Quote calls the real FareCalculator.EstimateWithContext — this IS the
// production Dynamic Pricing Engine, not a simulation-local reimplementation.
func (a *PricingAdapter) Quote(in QuoteInput) (*pricingentity.FareBreakdown, error) {
	vt := toPricingRateKey(in.ServiceType)
	ctx := pricingentity.NeutralContext(vt, in.RequestTime)
	ctx.ActiveRequests = in.ActiveRequests
	ctx.AvailableDrivers = in.AvailableDrivers
	ctx.IsAirportZone = in.IsAirportZone
	ctx.IsRainActive = in.Weather.IsRainy()
	ctx.IsHoliday = in.IsHoliday
	if in.Weather == entity.WeatherHeavyRain || in.Weather == entity.WeatherFlooded {
		ctx.TrafficLevel = pricingentity.TrafficLevelHeavy
	}
	return a.calc.EstimateWithContext(vt, in.DistanceKM, in.DurationMin, ctx)
}

// toPricingRateKey maps the simulation's ServiceType to this adapter's own
// Rates map key — "bike_plus"/"car_xl" are simulation-only keys (see
// NewPricingAdapter's doc comment), not production pricingentity.VehicleType
// constants.
func toPricingRateKey(s entity.ServiceType) pricingentity.VehicleType {
	switch s {
	case entity.ServiceBike:
		return pricingentity.VehicleTypeMotorcycle
	case entity.ServiceBikePlus:
		return pricingentity.VehicleType("bike_plus")
	case entity.ServiceCarXL:
		return pricingentity.VehicleType("car_xl")
	default:
		return pricingentity.VehicleTypeCar
	}
}
