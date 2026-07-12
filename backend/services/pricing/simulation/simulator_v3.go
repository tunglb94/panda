// Pricing V3 simulation — sprint brief PHẦN 13: "Simulation Engine phải đọc
// đúng Pricing V3. Không dùng Pricing cũ."
//
// This file is deliberately the OPPOSITE of pricing_simulator.go's isolation
// discipline. Simulator (pricing_simulator.go) intentionally never touches
// production types — its own doc comment says so — because its job was to
// let Product/Finance test proposed BRB rules BEFORE they existed in
// production, without any risk of a simulation parameter accidentally
// affecting a real fare. SimulatorV3 has a different job: prove that
// Pricing V3 (already real production code, once PRICING_VERSION=v3 is
// set — see cmd/server/main.go) behaves as documented across many
// scenarios. For that job, simulating anything OTHER than the real
// app.FareCalculatorV3 against the real config.PricingV3Config would prove
// nothing — so SimulatorV3 wraps the production calculator directly instead
// of re-implementing its formula a second time (avoiding exactly the kind
// of "duplicate logic" the sprint brief's top-level goal prohibits).
//
// pricing_simulator.go, Simulator, FareBreakdown, and the 111 scenarios in
// scenarios.go / docs/business/PRICING_SIMULATION_REPORT.md are completely
// unmodified by this file and remain valid exactly as before.
package simulation

import (
	"github.com/fairride/pricing/app"
	pricingconfig "github.com/fairride/pricing/config"
	"github.com/fairride/pricing/domain/entity"
)

// SimulatorV3 runs scenarios through the real production Pricing V3 engine.
type SimulatorV3 struct {
	calc *app.FareCalculatorV3
}

// NewSimulatorV3FromProductionConfig builds a SimulatorV3 from the same
// embedded default config production falls back to (config.Default()) —
// the default choice for "what would Pricing V3 charge for this trip today."
func NewSimulatorV3FromProductionConfig() *SimulatorV3 {
	cfg := pricingconfig.Default()
	return NewSimulatorV3(cfg, app.DefaultRuleConfigs())
}

// NewSimulatorV3 builds a SimulatorV3 from an explicit config and rule set —
// e.g. to simulate a proposed config change (a candidate
// PRICING_CONFIG_PATH file) before it is ever loaded by cmd/server/main.go.
func NewSimulatorV3(cfg *pricingconfig.PricingV3Config, ruleConfigs app.RuleConfigMap) *SimulatorV3 {
	calc := app.NewFareCalculatorV3(cfg.Fare, cfg.Airport, cfg.Commission, cfg.VATRate, ruleConfigs)
	return &SimulatorV3{calc: calc}
}

// Simulate computes one scenario using the exact production V3 formula —
// literally app.FareCalculatorV3.EstimateV3, not a reimplementation.
func (s *SimulatorV3) Simulate(in entity.RideInputV3) (*entity.FullFareBreakdownV3, error) {
	return s.calc.EstimateV3(in)
}

// SimulateFinal is Simulate's post-trip counterpart (CalculateFinalV3).
func (s *SimulatorV3) SimulateFinal(in entity.RideInputV3) (*entity.FullFareBreakdownV3, error) {
	return s.calc.CalculateFinalV3(in)
}
