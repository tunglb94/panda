// Package config loads Pricing V3's rate configuration from YAML — the
// sprint brief's PHẦN 2/19 requirement that "Không hardcode bất kỳ con số
// nào. Toàn bộ phải đọc từ config." No VehicleRatesV3/DistanceTier/
// CommissionConfigV3/AirportFeeConfigV3 value is ever constructed with a Go
// literal anywhere else in this service — every one of them is built here,
// from either an operator-supplied YAML file or the embedded default
// (pricing_v3.default.yaml), which is parsed through this exact same code
// path rather than being a second, hand-written Go source of truth.
package config

import (
	"embed"
	"fmt"
	"os"
	"sort"

	"github.com/fairride/pricing/domain/entity"
	"gopkg.in/yaml.v3"
)

//go:embed pricing_v3.default.yaml
var defaultConfigFS embed.FS

// vehicleTypeKeys maps the YAML "vehicles" map keys to entity.VehicleType —
// a lookup table, not an if/else chain, matching the discipline the rest of
// this service already uses for BRB-derived mappings. Pricing is out of
// scope for the Vehicle/Service Catalog refactor ("Không thay đổi
// Pricing") — exactly the 3 original physical-vehicle keys, unchanged.
var vehicleTypeKeys = map[string]entity.VehicleType{
	"car":        entity.VehicleTypeCar,
	"motorcycle": entity.VehicleTypeMotorcycle,
	"van":        entity.VehicleTypeVan,
}

var commissionTierKeys = map[string]entity.CommissionTier{
	"bronze":   entity.CommissionTierBronze,
	"silver":   entity.CommissionTierSilver,
	"gold":     entity.CommissionTierGold,
	"platinum": entity.CommissionTierPlatinum,
	"diamond":  entity.CommissionTierDiamond,
}

// ─── YAML wire shapes — plain data, no business logic ──────────────────────

type fileDistanceTier struct {
	FromKM    float64 `yaml:"from_km"`
	ToKM      float64 `yaml:"to_km"`
	RatePerKM int64   `yaml:"rate_per_km"`
}

type fileVehicleRates struct {
	BaseFare             int64              `yaml:"base_fare"`
	MinimumFare          int64              `yaml:"minimum_fare"`
	BookingFee           int64              `yaml:"booking_fee"`
	TrafficTimePerMinute int64              `yaml:"traffic_time_per_minute"`
	WaitingFeePerMinute  int64              `yaml:"waiting_fee_per_minute"`
	WaitingGraceMinutes  int                `yaml:"waiting_grace_minutes"`
	DistanceTiers        []fileDistanceTier `yaml:"distance_tiers"`
}

type fileAirportFees struct {
	PickupFee  map[string]int64 `yaml:"pickup_fee"`
	DropoffFee map[string]int64 `yaml:"dropoff_fee"`
}

type fileConfig struct {
	CurrencyCode     string                      `yaml:"currency_code"`
	VATRate          float64                     `yaml:"vat_rate"`
	Commission       map[string]float64          `yaml:"commission"`
	LastTierMinRatio float64                     `yaml:"last_tier_min_ratio"`
	Vehicles         map[string]fileVehicleRates `yaml:"vehicles"`
	Airport          fileAirportFees             `yaml:"airport"`
}

// PricingV3Config is everything FareCalculatorV3 needs, fully resolved into
// domain types.
type PricingV3Config struct {
	Fare       entity.FareConfigV3
	Airport    entity.AirportFeeConfigV3
	Commission entity.CommissionConfigV3
	VATRate    float64
}

// Default parses the embedded pricing_v3.default.yaml — the config Pricing
// V3 falls back to when no operator-supplied file path is given (e.g. in
// tests, or a first deploy before an ops-managed file exists). Panics only
// if the embedded file itself is malformed (a build-time defect, since it
// ships inside the binary — never a runtime/operator error).
func Default() *PricingV3Config {
	raw, err := defaultConfigFS.ReadFile("pricing_v3.default.yaml")
	if err != nil {
		panic("pricing/config: embedded default config missing: " + err.Error())
	}
	cfg, err := parse(raw)
	if err != nil {
		panic("pricing/config: embedded default config is invalid: " + err.Error())
	}
	return cfg
}

// Load reads and validates a Pricing V3 config from path. Operators change
// rates by editing this file (or pointing PRICING_CONFIG_PATH at a new one,
// see cmd/server/main.go) — never by editing Go source.
func Load(path string) (*PricingV3Config, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("pricing/config: reading %s: %w", path, err)
	}
	return parse(raw)
}

func parse(raw []byte) (*PricingV3Config, error) {
	var fc fileConfig
	if err := yaml.Unmarshal(raw, &fc); err != nil {
		return nil, fmt.Errorf("pricing/config: parsing yaml: %w", err)
	}
	return resolve(fc)
}

func resolve(fc fileConfig) (*PricingV3Config, error) {
	if fc.CurrencyCode == "" {
		return nil, fmt.Errorf("pricing/config: currency_code is required")
	}
	if fc.VATRate < 0 || fc.VATRate > 1 {
		return nil, fmt.Errorf("pricing/config: vat_rate must be within [0,1], got %v", fc.VATRate)
	}
	if len(fc.Vehicles) == 0 {
		return nil, fmt.Errorf("pricing/config: at least one vehicle class is required")
	}

	// last_tier_min_ratio has no implicit default: degressive tiering is the
	// entire point of Pricing V3 (PRICING_V3_DESIGN.md Phần 4), so silently
	// defaulting to something permissive would reopen the unbounded-discount
	// risk this guardrail exists to close (PRICING_V3_REVIEW.md Phần 13 mục
	// 2), while silently defaulting to something strict (e.g. 1.0, "no
	// discount allowed") would silently break ordinary tiering. Forcing the
	// operator to state this explicitly is safer than guessing either way.
	if fc.LastTierMinRatio <= 0 {
		return nil, fmt.Errorf("pricing/config: last_tier_min_ratio is required and must be > 0 (PRICING_V3_REVIEW.md Phần 13 mục 2 guardrail — e.g. 0.35 permits the last tier to be as low as 35%% of the first)")
	}
	lastTierMinRatio := fc.LastTierMinRatio

	rates := make(map[entity.VehicleType]entity.VehicleRatesV3, len(fc.Vehicles))
	for key, fv := range fc.Vehicles {
		vt, ok := vehicleTypeKeys[key]
		if !ok {
			return nil, fmt.Errorf("pricing/config: unknown vehicle key %q (want one of car/motorcycle/van)", key)
		}
		tiers, err := resolveTiers(key, fv.DistanceTiers, lastTierMinRatio)
		if err != nil {
			return nil, err
		}
		if fv.BaseFare < 0 || fv.MinimumFare < 0 || fv.BookingFee < 0 ||
			fv.TrafficTimePerMinute < 0 || fv.WaitingFeePerMinute < 0 || fv.WaitingGraceMinutes < 0 {
			return nil, fmt.Errorf("pricing/config: vehicle %q has a negative rate field", key)
		}
		rates[vt] = entity.VehicleRatesV3{
			BaseFare:             fv.BaseFare,
			DistanceTiers:        tiers,
			TrafficTimePerMinute: fv.TrafficTimePerMinute,
			WaitingFeePerMinute:  fv.WaitingFeePerMinute,
			WaitingGraceMinutes:  fv.WaitingGraceMinutes,
			MinimumFare:          fv.MinimumFare,
			BookingFee:           fv.BookingFee,
		}
	}

	commissionRates := make(map[entity.CommissionTier]float64, len(fc.Commission))
	for key, rate := range fc.Commission {
		tier, ok := commissionTierKeys[key]
		if !ok {
			return nil, fmt.Errorf("pricing/config: unknown commission tier %q", key)
		}
		if rate < 0 || rate > 1 {
			return nil, fmt.Errorf("pricing/config: commission tier %q rate must be within [0,1], got %v", key, rate)
		}
		commissionRates[tier] = rate
	}
	if _, ok := commissionRates[entity.CommissionTierBronze]; !ok {
		return nil, fmt.Errorf("pricing/config: commission.bronze is required (fallback rate for unrecognised tiers)")
	}

	airport := entity.AirportFeeConfigV3{
		PickupFee:  resolveAirportFees(fc.Airport.PickupFee),
		DropoffFee: resolveAirportFees(fc.Airport.DropoffFee),
	}

	return &PricingV3Config{
		Fare: entity.FareConfigV3{
			CurrencyCode: fc.CurrencyCode,
			Rates:        rates,
		},
		Airport:    airport,
		Commission: entity.CommissionConfigV3{RateByTier: commissionRates},
		VATRate:    fc.VATRate,
	}, nil
}

func resolveAirportFees(raw map[string]int64) map[entity.VehicleType]int64 {
	out := make(map[entity.VehicleType]int64, len(raw))
	for key, fee := range raw {
		if vt, ok := vehicleTypeKeys[key]; ok {
			out[vt] = fee
		}
	}
	return out
}

// resolveTiers validates a vehicle's distance tier table: sorted ascending,
// no gaps, exactly one open-ended (to_km<=0) tier and it must be last, and
// the last tier's rate must not fall below lastTierMinRatio x the first
// tier's rate (docs/business/PRICING_V3_REVIEW.md Phần 13 mục 2 guardrail —
// see the yaml file's header comment).
func resolveTiers(vehicleKey string, raw []fileDistanceTier, lastTierMinRatio float64) ([]entity.DistanceTier, error) {
	if len(raw) == 0 {
		return nil, fmt.Errorf("pricing/config: vehicle %q has no distance_tiers", vehicleKey)
	}
	sorted := make([]fileDistanceTier, len(raw))
	copy(sorted, raw)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].FromKM < sorted[j].FromKM })

	tiers := make([]entity.DistanceTier, 0, len(sorted))
	nextFrom := 0.0
	for i, t := range sorted {
		if t.RatePerKM < 0 {
			return nil, fmt.Errorf("pricing/config: vehicle %q tier %d has a negative rate_per_km", vehicleKey, i)
		}
		if t.FromKM != nextFrom {
			return nil, fmt.Errorf("pricing/config: vehicle %q distance_tiers must be contiguous starting at 0 (tier %d starts at %v, expected %v)", vehicleKey, i, t.FromKM, nextFrom)
		}
		openEnded := t.ToKM <= 0
		if openEnded && i != len(sorted)-1 {
			return nil, fmt.Errorf("pricing/config: vehicle %q has an open-ended tier (to_km<=0) that is not the last tier", vehicleKey)
		}
		if !openEnded && t.ToKM <= t.FromKM {
			return nil, fmt.Errorf("pricing/config: vehicle %q tier %d has to_km <= from_km", vehicleKey, i)
		}
		if i == len(sorted)-1 && !openEnded {
			return nil, fmt.Errorf("pricing/config: vehicle %q's last distance tier must be open-ended (to_km: 0)", vehicleKey)
		}
		tiers = append(tiers, entity.DistanceTier{FromKM: t.FromKM, ToKM: t.ToKM, RatePerKM: t.RatePerKM})
		if !openEnded {
			nextFrom = t.ToKM
		}
	}

	first := tiers[0].RatePerKM
	last := tiers[len(tiers)-1].RatePerKM
	if first > 0 && float64(last) < float64(first)*lastTierMinRatio {
		return nil, fmt.Errorf(
			"pricing/config: vehicle %q's last tier rate (%d/km) is below %.0f%% of its first tier rate (%d/km) — "+
				"this is the unbounded long-distance discount guardrail from docs/business/PRICING_V3_REVIEW.md Phần 13 mục 2; "+
				"raise the last tier's rate_per_km or last_tier_min_ratio if this is an intentional, CFO-approved long-haul rate",
			vehicleKey, last, lastTierMinRatio*100, first)
	}
	return tiers, nil
}
