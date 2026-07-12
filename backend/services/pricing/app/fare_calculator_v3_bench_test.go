package app_test

import (
	"testing"
	"time"

	"github.com/fairride/pricing/app"
	"github.com/fairride/pricing/config"
	"github.com/fairride/pricing/domain/entity"
)

// Sprint brief PHẦN 15: benchmark at 100 / 1,000 / 10,000 / 100,000 fare
// calculations, measuring latency/allocation/memory. Go's `testing.B`
// already reports ns/op (latency) and, with b.ReportAllocs(), B/op and
// allocs/op (memory/allocation) — run with:
//
//	go test ./app/... -bench BenchmarkFareCalculatorV3 -benchmem -run ^$
//
// The four sub-benchmarks below each compute a batch of N calculations per
// b.N iteration, so `go test -bench` naturally reports "ns per batch of N"
// alongside the standard per-call metrics from the unsized variant.

func benchmarkInput(i int) entity.RideInputV3 {
	vehicles := []entity.VehicleType{entity.VehicleTypeCar, entity.VehicleTypeMotorcycle, entity.VehicleTypeVan}
	return entity.RideInputV3{
		VehicleType:    vehicles[i%len(vehicles)],
		DistanceKM:     float64(1 + i%120),
		DurationMin:    float64(1+i%120) * 2.2,
		WaitingMin:     float64(i % 10),
		RequestTime:    time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC),
		CommissionTier: entity.CommissionTierBronze,
	}
}

func BenchmarkFareCalculatorV3_Estimate(b *testing.B) {
	cfg := config.Default()
	calc := app.NewFareCalculatorV3(cfg.Fare, cfg.Airport, cfg.Commission, cfg.VATRate, app.DefaultRuleConfigs())
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := calc.EstimateV3(benchmarkInput(i)); err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
	}
}

func benchmarkBatch(b *testing.B, n int) {
	cfg := config.Default()
	calc := app.NewFareCalculatorV3(cfg.Fare, cfg.Airport, cfg.Commission, cfg.VATRate, app.DefaultRuleConfigs())
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < n; j++ {
			if _, err := calc.EstimateV3(benchmarkInput(j)); err != nil {
				b.Fatalf("unexpected error: %v", err)
			}
		}
	}
}

func BenchmarkFareCalculatorV3_Batch100(b *testing.B)    { benchmarkBatch(b, 100) }
func BenchmarkFareCalculatorV3_Batch1000(b *testing.B)   { benchmarkBatch(b, 1_000) }
func BenchmarkFareCalculatorV3_Batch10000(b *testing.B)  { benchmarkBatch(b, 10_000) }
func BenchmarkFareCalculatorV3_Batch100000(b *testing.B) { benchmarkBatch(b, 100_000) }

// BenchmarkFareCalculatorV2_Estimate — baseline for comparison, confirming
// V3's added Distance Tier walk / Rule Engine reuse doesn't regress V2's
// already-benchmarked path (see fare_calculator_bench_test.go).
func BenchmarkFareCalculatorV2_Estimate(b *testing.B) {
	calc := app.NewFareCalculator(entity.DefaultFareConfig())
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := calc.Estimate(entity.VehicleTypeCar, float64(1+i%120), float64(1+i%120)*2.2); err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
	}
}
