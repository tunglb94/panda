package main

import (
	"log"
	"os"

	sharedconfig "github.com/fairride/shared/config"
	sharedgrpc "github.com/fairride/shared/grpc"
	"github.com/fairride/shared/server"

	"github.com/fairride/pricing/app"
	pricingconfig "github.com/fairride/pricing/config"
	"github.com/fairride/pricing/domain/entity"
	pricinggrpc "github.com/fairride/pricing/grpc"
	"github.com/fairride/pricing/grpc/pricingpb"
)

func main() {
	server.Run("pricing", register)
}

func register(srv *sharedgrpc.Server, ready *server.ReadinessTracker) {
	cfg := sharedconfig.Load("pricing")
	_ = cfg // pricing is a pure compute service; no DB or Redis required

	// ─── Pricing V3 feature flag — PRICING_VERSION env var, "v2" (default)
	// or "v3". docs/business/PRICING_V3_IMPLEMENTATION.md Phần 4/17: NOT
	// bật mặc định — an unset/unrecognised value fails closed to v2 inside
	// app.NewVersionedFareCalculator, so a missing env var is exactly as
	// safe as an explicit PRICING_VERSION=v2. ─────────────────────────────
	version := app.PricingVersion(os.Getenv("PRICING_VERSION"))

	v2Calc := app.NewFareCalculator(entity.DefaultFareConfig())

	var v3Calc *app.FareCalculatorV3
	if version == app.PricingVersionV3 {
		// PRICING_CONFIG_PATH lets operators point at an ops-managed YAML
		// file (see backend/services/pricing/config/pricing_v3.default.yaml
		// for the schema) without a code change or release — falls back to
		// the embedded default when unset, so a bare PRICING_VERSION=v3
		// with no config file is still a valid, runnable configuration.
		v3Config := pricingconfig.Default()
		if path := os.Getenv("PRICING_CONFIG_PATH"); path != "" {
			loaded, err := pricingconfig.Load(path)
			if err != nil {
				log.Fatalf("pricing: PRICING_CONFIG_PATH=%s is invalid, refusing to start with PRICING_VERSION=v3: %v", path, err)
			}
			v3Config = loaded
		}
		v3Calc = app.NewFareCalculatorV3(v3Config.Fare, v3Config.Airport, v3Config.Commission, v3Config.VATRate, app.DefaultRuleConfigs())
	}

	calc := app.NewVersionedFareCalculator(version, v2Calc, v3Calc)
	handler := pricinggrpc.NewHandler(calc)
	pricingpb.RegisterPricingServiceServer(srv.Inner(), handler)
}
