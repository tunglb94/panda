# Panda — Pricing V3 Implementation Report

**Document Classification:** Internal — Confidential
**Authority:** CPO · CFO · CTO
**Effective Date:** 2026-07-11
**Status:** IMPLEMENTED, NOT ACTIVE — `PRICING_VERSION` defaults to `v2` everywhere; V3 code is live in the binary but computes nothing in production until an operator explicitly sets the env var. **Not committed, not pushed** (per sprint instruction).
**Read in order before this report, per sprint instruction:** `docs/business/business-rule-bible-v1.0.md`, `docs/business/MARKET_PRICING_RESEARCH.md`, `docs/business/PRICING_V3_DESIGN.md`, `docs/business/PRICING_V3_REVIEW.md`. Where these conflicted, this implementation followed the stated priority **Review > Design > BRB > Research** — see Phần 6/13 below for the two places that mattered in practice.

---

## 1. ARCHITECTURE

Pricing V3 is an **additive layer on top of the existing Dynamic Pricing Engine**, not a rewrite. The chain the sprint required to survive unchanged did:

```
PricingPipeline → PricingEvaluator → PricingRule → RuleConfig → PricingContext → PricingResult
```

Every one of those six types (`app/pricing_pipeline.go`, `app/pricing_evaluator.go`, `app/pricing_rule.go`, `app/rules_defined.go`, `app/rules_todo.go`, `domain/entity/pricing_context.go`, `domain/entity/pricing_result.go`) is **byte-for-byte unmodified** except one additive field on `PricingContext` (`AirportLeg`, zero-value backward compatible — see Phần 8). `FareCalculator` (V2, `app/fare_calculator.go`) is also completely unmodified.

What V3 adds:

```
                         ┌────────────────────────────┐
config/*.yaml ──Load()──▶│      config.PricingV3Config │
                         └──────────────┬─────────────┘
                                        │
              ┌─────────────────────────┼──────────────────────────┐
              ▼                         ▼                          ▼
   entity.FareConfigV3      entity.AirportFeeConfigV3   entity.CommissionConfigV3
   (Base/DistanceTiers/            (Pickup/Dropoff              (Bronze..Diamond
    TrafficTime/Waiting/            fee by vehicle)                rate by tier)
    Minimum/Booking)
              │                         │
              ▼                         ▼
   entity.DistanceFareForTiers   app.AirportFeeRuleV3 ──registered into──▶ app.NewDefaultPricingPipelineV3()
   (tier-table walk, not                                                  (same 9 rules as V2's pipeline,
    if/else)                                                               AirportFeeRule swapped for V3's)
              │                                                            │
              └─────────────────────┬──────────────────────────────────────┘
                                    ▼
                     app.FareCalculatorV3.EstimateV3 / CalculateFinalV3
                                    │
                        entity.FullFareBreakdownV3
                     (Base/Distance/Moving/Traffic/Waiting/Airport/
                      Surge/Voucher/Commission/VAT/PlatformFee/
                      DriverIncome/PlatformRevenue/FinalFare)
                                    │
                     ┌──────────────┴───────────────┐
                     ▼                               ▼
        app.ValidateFullBreakdown          FullFareBreakdownV3.Explanation()
        (fare<0, commission>100%,          / .ExplanationString()
         discount>fare, NaN/Inf,
         overflow — PHẦN 11)

app.VersionedFareCalculator (PricingVersion "v2"|"v3")
  ├─ v2 → *app.FareCalculator (unmodified)              ──▶ entity.FareBreakdown (unchanged wire shape)
  └─ v3 → *app.FareCalculatorV3.EstimateV3 → downgradeToFareBreakdown() ──▶ entity.FareBreakdown (same shape)
                     │
                     ▼
        app.FareEstimator interface (NEW — extracted from grpc.Handler's
        previously-concrete *app.FareCalculator field)
                     │
                     ▼
        grpc.Handler  →  cmd/server/main.go reads PRICING_VERSION / PRICING_CONFIG_PATH env vars
```

**Why this shape:** the sprint's PHẦN 1 was explicit that the rule engine must not regress to if/else, and the existing engine already does exactly what V3 needs for surge/surcharges (Demand/Peak/Night/Holiday/Rain/Airport) — reusing it via one new pipeline constructor (`NewDefaultPricingPipelineV3`, additive, `pricing_pipeline.go` untouched) was both the safest and the least-code path. Distance Tier and Commission are the two genuinely new calculations V3 introduces; both are pure functions taking config as input, no hardcoded numbers, and both have their own focused unit test files.

---

## 2. FILES CHANGED

### 2.1 New files (this sprint)

| File | Purpose |
|---|---|
| `domain/entity/distance_tier.go` | `DistanceTier`, `VehicleRatesV3`, `FareConfigV3`, `DistanceFareForTiers` (degressive tier walk) |
| `domain/entity/airport_v3.go` | `AirportLeg`, `AirportFeeConfigV3` |
| `domain/entity/commission_v3.go` | `CommissionTier`, `CommissionConfigV3` |
| `domain/entity/ride_input_v3.go` | `RideInputV3` — FareCalculatorV3's input DTO |
| `domain/entity/full_breakdown_v3.go` | `FullFareBreakdownV3`, `ExplanationLine`, `.Explanation()`, `.ExplanationString()` |
| `domain/entity/pricing_context.go` | **Modified** (pre-existing file from the Dynamic Pricing Engine sprint) — added one field, `AirportLeg AirportLeg` |
| `app/rules_airport_v3.go` | `AirportFeeRuleV3` — a `PricingRule` implementation, leg/vehicle-aware |
| `app/pricing_pipeline_v3.go` | `NewDefaultPricingPipelineV3` — same 9 rules as V2's pipeline, Airport swapped |
| `app/fare_calculator_v3.go` | `FareCalculatorV3`, `EstimateV3`, `CalculateFinalV3` |
| `app/validation_v3.go` | `validateRideInput`, `ValidateFullBreakdown` |
| `app/feature_flag.go` | `PricingVersion`, `VersionedFareCalculator`, `downgradeToFareBreakdown` |
| `app/fare_estimator.go` | `FareEstimator` interface (both `*FareCalculator` and `*VersionedFareCalculator` satisfy it) |
| `config/config.go` | YAML loader + validator (`Load`, `Default`) |
| `config/pricing_v3.default.yaml` | Embedded default rate config — every V3 number lives here, not in Go |
| `simulation/simulator_v3.go` | `SimulatorV3` — thin wrapper around `app.FareCalculatorV3`, for scenario testing against the real engine |
| 10 new `*_test.go` files | See Phần 11/12 |

### 2.2 Modified files (pre-existing, this sprint)

| File | Change |
|---|---|
| `grpc/handler.go` | `Handler.calc` / `NewHandler` param type changed from concrete `*app.FareCalculator` to the new `app.FareEstimator` interface. **No behavioural change** — `*app.FareCalculator` already satisfies the interface, so every existing caller compiles unchanged. |
| `cmd/server/main.go` | Reads `PRICING_VERSION` / `PRICING_CONFIG_PATH` env vars, constructs `VersionedFareCalculator`, passes it to the (now interface-typed) `pricinggrpc.NewHandler`. |
| `go.mod` / `go.sum` | Added `gopkg.in/yaml.v3 v3.0.1` (the only new dependency — network-verified reachable before adding, see Phần 5). |

### 2.3 Explicitly NOT touched

- `backend/services/pricing/domain/entity/fare.go`, `app/fare_calculator.go`, and every V2 test file — the pre-existing V2 engine, byte-for-byte unmodified.
- `backend/services/pricing/simulation/pricing_simulator.go`, `pricing_constants.go`, `scenarios.go`, `competitive.go`, `optimizer.go`, `sensitivity.go`, and their tests — the pre-existing 111-scenario V2 simulator (`docs/business/PRICING_SIMULATION_REPORT.md`) remains valid exactly as before; see Phần 13's "Known Limitations" for why V3 got a *new* `SimulatorV3` instead of modifying this one.
- `backend/services/promotion/**` — audited, not modified (Phần 7).
- `backend/proto/pricing/*.proto`, `grpc/pricingpb/*.pb.go` — not regenerated (Phần 13, 15).
- Every Driver/Rider Flutter file, Dispatch, Booking, Trip, Identity, Map service — out of scope per the sprint's explicit `PHẠM VI`.

---

## 3. MIGRATION

No migration is required to **not** activate V3 — the default (`PRICING_VERSION` unset) is byte-for-byte V2 behaviour, verified by `TestVersionedFareCalculator_V2ModeMatchesDirectFareCalculator`.

To activate V3 in an environment:

1. (Optional) Prepare an ops-managed `pricing.yaml` following `config/pricing_v3.default.yaml`'s schema, and run it through `go test ./config/...` locally against a copy of `config_test.go`'s validation cases, or simply call `config.Load(path)` in a throwaway script — `Load` returns a descriptive error for every malformed shape (gaps between tiers, non-terminal open-ended tier, missing `commission.bronze`, `last_tier_min_ratio` unset, negative rates, commission >100%). If skipped, `PRICING_CONFIG_PATH` unset falls back to the embedded default.
2. Set `PRICING_VERSION=v3` (and `PRICING_CONFIG_PATH=/path/to/pricing.yaml` if using a custom file) on the `pricing` service.
3. Restart the service. `cmd/server/main.go` calls `log.Fatalf` and refuses to start if `PRICING_CONFIG_PATH` points at an invalid file — a bad config can never silently fall back to serving wrong prices.
4. Monitor — see Rollback Plan (Phần 14).

---

## 4. FEATURE FLAG

`app.PricingVersion` — `"v2"` (default) or `"v3"`, read from the `PRICING_VERSION` environment variable in `cmd/server/main.go`. `app.NewVersionedFareCalculator` fails closed: any value other than the literal string `"v3"` (including empty, typos, `"V3"`, `"3"`) resolves to `v2`. This is covered by `TestVersionedFareCalculator_DefaultsToV2` and `TestVersionedFareCalculator_UnrecognisedVersionFailsClosedToV2`.

`VersionedFareCalculator.Estimate`/`CalculateFinal` always return the **V2-shaped** `entity.FareBreakdown` regardless of which engine actually computed it (`downgradeToFareBreakdown` maps V3's richer `FullFareBreakdownV3` onto it) — so the gRPC wire contract is identical in both modes, and flipping the flag is a pure runtime behaviour change, zero API surface change.

For callers that specifically want the full V3 breakdown (Explanation, Commission, VAT, Driver Income, Platform Revenue — not yet exposed over gRPC, see Phần 15), `VersionedFareCalculator.EstimateV3Detailed` returns it, but only while `Version == "v3"` — calling it in v2 mode returns an explicit error rather than a stale/zero-valued struct.

---

## 5. PRICING CONFIG

**No number in `DistanceTier`/`VehicleRatesV3`/`AirportFeeConfigV3`/`CommissionConfigV3` is a Go literal anywhere in this service.** Every value is parsed from YAML by `config.Load`/`config.Default` — including the *default* config, which is an embedded YAML file (`//go:embed pricing_v3.default.yaml`) parsed through the exact same code path as an operator-supplied file, not a second hand-written Go source of truth.

`config.Load` validates, and rejects with a descriptive error, before any bad config can reach a running calculator:
- distance tiers must be contiguous from 0, sorted, with exactly one open-ended (`to_km: 0`) tier and it must be last;
- the open-ended tier's rate must be ≥ `last_tier_min_ratio` × the first tier's rate — **no implicit default** for this ratio (Phần 6 explains why);
- `commission.bronze` is required (the fallback rate for unrecognised tiers);
- every rate field ≥ 0; every commission rate ∈ [0,1]; `vat_rate` ∈ [0,1].

`gopkg.in/yaml.v3` is the one new dependency this sprint adds — network reachability to `proxy.golang.org` was verified before adding it (`go list -m -versions gopkg.in/yaml.v3` succeeded), and `go mod tidy` / `go build` / `go test ./...` all pass with it in place.

---

## 6. DISTANCE TIER

Implemented exactly as `PRICING_V3_DESIGN.md` Phần 4 specified (7 bands, degressive), **with two numeric changes sourced from `PRICING_V3_REVIEW.md`, per the sprint's explicit conflict-priority rule (Review > Design):**

| # | What changed | From (Design) | To (this implementation) | Why |
|---|---|---|---|---|
| 1 | Car `MinimumFare` | 30,000 VND | **25,000 VND** | `PRICING_V3_REVIEW.md` Phần 3 (W1) / Phần 13 mục 1 found the Design's 30,000 minimum made a 1km Car trip **+23%** more expensive than the market average — directly contradicting this sprint's stated goal ("giá đủ rẻ để khách chuyển từ Grab/Be/Xanh"). 25,000 is not an invented number — it is BRB §2.2.4's own already-approved Standard-class minimum, reverted rather than replaced with a new guess. `TestFareCalculatorV3_1kmCarNoLongerSeverelyOverpriced` guards this. |
| 2 | Last-tier floor | None (open-ended, unbounded discount) | `last_tier_min_ratio` config guardrail (no implicit default — `config.Load` requires it, defaulting to a **fail-closed error**, not a permissive number) | `PRICING_V3_REVIEW.md` Phần 2 / Phần 13 mục 2 found the unbounded last tier makes the fare/market gap widen without limit on very long trips (−37.7% at 100km, worsening further beyond). Picking a "correct" long-haul rate is a CFO/CPO commercial decision this implementation cannot make unilaterally — so it ships as a **structural guardrail** (any config whose last tier falls below the configured ratio is rejected at load time) rather than a specific invented number. The shipped default (`last_tier_min_ratio: 0.35`) is satisfied by every vehicle class's Design-specified tiers unchanged. |

Everything else (the 7 km bands and their rates, per vehicle class) is exactly what `PRICING_V3_DESIGN.md` Phần 4.2 specified — Bike/XL tiers were not touched, since the Review's quantified findings (Phần 3/5/10) were specifically about Car; auditing Bike/XL for the same 1km effect is listed as a follow-up in Phần 13 below, not silently assumed fixed.

Commission was also changed from BRB §7.1's 20%/18%/16%/14%/12% to `PRICING_V3_REVIEW.md` Phần 6.2's recommended **16%/15%/14%/13%/12%** — the Review found the old 20% Bronze rate, applied to V3's now-higher fares, takes a disproportionately large absolute amount from drivers (flagged P0 in Phần 13 mục 3).

---

## 7. RULE MATRIX

| Rule | V2 (`pricing_pipeline.go`) | V3 (`pricing_pipeline_v3.go`) | Changed? |
|---|---|---|---|
| Demand Surge | `NewDemandSurgeRule(entity.DefaultDSRTiers())` | same | No |
| Supply Surge | `NewSupplySurgeRule()` — TODO stub, `Applied` always false | same | No — **PHẦN 5 explicit instruction: "Các rule TODO vẫn giữ TODO. Không tự thêm rule."** |
| Peak Hour | `NewPeakHourRule(entity.DefaultPeakHourWindows())` | same | No |
| Night | `NewNightSurchargeRule(...)` | same | No |
| Holiday | `NewHolidaySurchargeRule()` | same | No |
| Rain | `NewRainSurchargeRule()` | same | No |
| **Airport** | `NewAirportFeeRule()` — flat `entity.AirportFeeVND` (10,000, all vehicles) | `NewAirportFeeRuleV3(airportConfig, vehicleType)` — leg-specific, vehicle-specific, config-driven | **Yes — the one rule V3 changes**, per `PRICING_V3_DESIGN.md` Phần 7 |
| Traffic | `NewTrafficSurgeRule()` — TODO stub | same | No |
| Special Event | `NewSpecialEventRule()` — TODO stub | same | No |

No rule was added or removed; `TestFareCalculatorV3_NoRulesEnabledIsNeutral` and `TestFareCalculatorV3_NightSurgeUsesExistingRuleEngine` confirm V3's surge behaviour is driven by the same `RuleConfigMap`/`PricingEvaluator` mechanics as V2, not a parallel implementation.

---

## 8. BACKWARD COMPATIBILITY

| Guarantee | How it's enforced | Test |
|---|---|---|
| V2 `FareCalculator.Estimate`/`CalculateFinal` output unchanged | File byte-for-byte unmodified | `TestFareCalculatorV2_UnaffectedByV3Additions`, plus all 26 pre-existing V2 tests still pass unmodified |
| `PricingContext` extension is additive | New `AirportLeg` field defaults to `AirportLegNone` (Go zero value) — V2's `AirportFeeRule` never reads it | Implicit: every V2 surge/pipeline test still passes |
| `grpc.Handler`'s public contract unchanged | `NewHandler` now takes an interface (`FareEstimator`) instead of a concrete type — Go interfaces are structural, so `*app.FareCalculator` satisfies it with zero changes to that type | `go test ./grpc/...` — pre-existing handler tests pass unmodified |
| gRPC wire shape unchanged in both v2 and v3 mode | `downgradeToFareBreakdown` maps V3's breakdown onto the exact same `entity.FareBreakdown`/`pricingpb.FareBreakdown` shape `grpc/handler.go`'s `toProto` has always produced | `TestVersionedFareCalculator_V3ModeReturnsV2Shape` |
| No API deleted | `Estimate`, `CalculateFinal`, `EstimateWithContext`, `CalculateFinalWithContext` (V2) all still exist and are called exactly as before | `go build ./...` succeeds; nothing referencing these symbols was touched |

---

## 9. PERFORMANCE

Both engines run in low single-digit microseconds per call — several orders of magnitude below anything that could bottleneck a booking flow. See Phần 10 for raw numbers.

**One known cost, documented rather than silently accepted:** `FareCalculatorV3.calculate` constructs a fresh `PricingPipeline` (9 `PricingRule` objects, including a per-vehicle-type `AirportFeeRuleV3`) on **every call**, because `AirportFeeRuleV3` needs to know the vehicle type at construction time and the existing `PricingRule` interface has no per-call context parameter for that. V2's `FareCalculator` builds its pipeline **once**, in its constructor, and reuses it. This is why V3 shows more allocations/op than V2 in the benchmark (22 vs 10) — despite that, V3 is still faster in wall-clock terms in this benchmark run (see Phần 10), so this was left as a documented "Known Limitation" (Phần 13) / "Next Phase" (Phần 15) item rather than a blocking issue, consistent with the sprint's stated priority "Đúng > Ổn định > Dễ rollback > Nhanh."

---

## 10. BENCHMARK

Run: `go test ./app/... -bench "BenchmarkFareCalculatorV3|BenchmarkFareCalculatorV2" -benchmem -run "^$"`
Machine: 12th Gen Intel Core i5-12400F, Windows.

```
BenchmarkFareCalculatorV3_Estimate-12       476196      2702 ns/op      5672 B/op      22 allocs/op
BenchmarkFareCalculatorV3_Batch100-12         4334    273887 ns/op    567200 B/op    2200 allocs/op
BenchmarkFareCalculatorV3_Batch1000-12         388   3020246 ns/op   5672007 B/op   22000 allocs/op
BenchmarkFareCalculatorV3_Batch10000-12         36  46830858 ns/op  56720105 B/op  220000 allocs/op
BenchmarkFareCalculatorV3_Batch100000-12         2 531911500 ns/op 567200728 B/op 2200006 allocs/op
BenchmarkFareCalculatorV2_Estimate-12       259792      5132 ns/op      4800 B/op      10 allocs/op
```

Reading this: a single `EstimateV3` call costs ~2.7µs / ~5.7KB / 22 allocations; scaling linearly, 100,000 calculations cost ~532ms wall-clock and ~567MB of (garbage-collected) allocation — well within what a single gRPC instance handles comfortably for a synchronous, stateless compute call. `BenchmarkFareCalculatorV2_Estimate` is included as the pre-existing baseline for comparison, confirming V3's additions did not regress V2's own already-benchmarked path (`fare_calculator_bench_test.go`, pre-existing, untouched).

---

## 11. TEST COVERAGE

`go test ./... -cover`:

```
github.com/fairride/pricing/app              90.4% of statements
github.com/fairride/pricing/config            85.0% of statements
github.com/fairride/pricing/domain/entity     86.4% of statements
github.com/fairride/pricing/grpc              72.7% of statements
github.com/fairride/pricing/simulation        89.7% of statements
github.com/fairride/pricing/cmd/pricing-simulate   0.0% (pre-existing binary entry point, not unit tested)
github.com/fairride/pricing/cmd/server             0.0% (binary entry point — env var wiring, exercised manually per Phần 3 Migration steps, not unit tested)
github.com/fairride/pricing/grpc/pricingpb         0.0% (generated code)
```

New test files this sprint (11 files, ~60 new test functions + 105 golden sub-tests):

`domain/entity`: `distance_tier_test.go`, `airport_v3_test.go`, `commission_v3_test.go`, `full_breakdown_v3_test.go`
`config`: `config_test.go`
`app`: `fare_calculator_v3_test.go`, `fare_calculator_v3_golden_test.go`, `validation_v3_test.go`, `feature_flag_test.go`, `fare_calculator_v3_bench_test.go`
`simulation`: `simulator_v3_test.go`

---

## 12. GOLDEN CASES

**105 frozen scenarios** (≥ 100 required) in `app/fare_calculator_v3_golden_test.go` — every combination of {car, motorcycle, van} × 35 distances (1km to 100km) × all 5 commission tiers (cycled), each computed once against `config.Default()` with a fixed, neutral request time (no night/peak/weekend sensitivity), and frozen as literal expected values for `BaseFare`, `DistanceFare`, `RideFare`, `Commission`, `PlatformRevenue`, `FinalFare`.

**Generation method (documented for reproducibility, not hidden):** a temporary generator test (`golden_gen_test.go`, deleted after use — its exact contents are preserved in this session's history for anyone who needs to regenerate the table after an intentional formula/config change) ran `FareCalculatorV3.EstimateV3` across the scenario grid and printed Go struct literals, which were then pasted into the permanent golden test file. This is the same "compute once, freeze, diff on every future run" discipline `PRICING_SIMULATION_REPORT.md` used for its own scenario tables.

To regenerate after an intentional change: temporarily restore a generator of the same shape (loop the grid, call `EstimateV3`, `fmt.Printf` the struct literal), run it, and replace `goldenCases`' body — never hand-edit individual frozen numbers to "make the test pass" without first confirming the *reason* the numbers moved was the intended change.

---

## 13. KNOWN LIMITATIONS

Documented explicitly rather than silently left implicit, per this project's established discipline (matching `PRICING_SIMULATION_REPORT.md`'s own "ASSUMPTION" flags):

1. **Pipeline reconstruction cost per call** (Phần 9) — not a correctness issue, a minor allocation cost; candidate for `AirportFeeRuleV3` to read vehicle type from `PricingContext` instead of its constructor in a future pass, letting `FareCalculatorV3` cache one pipeline like V2 does.
2. **Bike/XL Minimum Fare not re-audited** — `PRICING_V3_REVIEW.md`'s quantified 1km-overpricing finding (Phần 3) was specifically about Car; this implementation fixed Car (Phần 6) but did not independently re-derive whether Bike (9,000) or Van/XL (48,000) have the same defect, since the Review didn't quantify it for those classes and inventing a fix without that quantification would be guessing, not implementing.
3. **Voucher/Discount is caller-supplied, not Promotion-Engine-integrated** — `RideInputV3.VoucherDiscountVND` is a pre-resolved amount (identical division of responsibility to `simulation/pricing_simulator.go`'s pre-existing `PromotionInput` pattern). Pricing V3 does not call the Promotion Engine's gRPC surface. This matches the sprint's PHẦN 7 instruction to *verify*, not integrate, Promotion Engine rules (Phần 7 below) — live cross-service wiring (Booking Service calling both Pricing and Promotion, then passing the resolved discount to Pricing) is unchanged from how the system already works today and was explicitly out of this sprint's file scope (`backend/services/promotion/**` only "khi cần wiring" — no wiring was found to be necessary since Pricing already only ever consumed a pre-resolved amount).
4. **Driver minimum-earning guarantee not modelled in V3 Pricing** — BRB §2.14's 20,000 VND floor (already implemented in `simulation/pricing_simulator.go`'s `MinimumDriverEarningVND` top-up) is a Settlement-time guarantee, not a Pricing-time one; `FullFareBreakdownV3.DriverIncome` is the gross commission-split amount before any such top-up. Documented in the field's doc comment, not silently omitted.
5. **`SimulatorV3` is a thin wrapper, not a scenario-generator like the V2 `Simulator`** — it proves consistency with production (Phần 12's cross-check test) but does not ship its own 100+/300+ scenario table the way `scenarios.go` does for V2; `docs/business/MARKET_PRICING_RESEARCH.md`'s 312-scenario table already exists for that purpose and used Python, not this Go package, to compute its numbers.
6. **Pre-existing `gofmt` debt untouched** — `gofmt -l .` on this package also lists several files this sprint never touched (`app/fare_calculator_test.go`, `app/pricing_pipeline_test.go`, `domain/entity/fare.go`, `domain/entity/fare_test.go`, `grpc/handler_test.go`, generated `pricingpb/*.pb.go`, and 4 `simulation/*.go` files) — these were already not gofmt-clean before this sprint (Windows CRLF artifacts) and were deliberately left as-is to keep this sprint's diff scoped to what it actually changed; every file this sprint created or modified is gofmt-clean (`gofmt -l` returns empty for that file set).

---

## 14. ROLLBACK PLAN

Because `VersionedFareCalculator` fails closed to v2 for anything other than the exact string `"v3"`:

- **Instant rollback:** unset `PRICING_VERSION` (or set it to anything other than `"v3"`, e.g. `"v2"`) and restart the `pricing` service. No code change, no redeploy of a different binary, no data migration — the v3 code paths are simply never entered again.
- **No data to roll back:** Pricing V3 is a pure, stateless compute path (same as V2) — no database rows, no cache entries, no persisted state exists that a rollback would need to reconcile.
- **Startup-time safety net:** if `PRICING_VERSION=v3` is set with a broken `PRICING_CONFIG_PATH`, the service refuses to start at all (`log.Fatalf` in `cmd/server/main.go`) rather than starting with a partially-invalid config — so a bad rollout is caught before it ever serves a single price, not discovered after.

---

## 15. NEXT PHASE

1. **Proto extension** — expose `FullFareBreakdownV3`'s new fields (Commission, VAT, Driver Income, Platform Revenue, Explanation) over gRPC. Not done this sprint (`PHẠM VI` allowed it "nếu thật sự cần" — judged not strictly necessary to prove the V3 engine works, and regenerating `.pb.go` files carries more risk than the Go-level API alone). `VersionedFareCalculator.EstimateV3Detailed` is ready to be called by a new RPC method once the proto is extended.
2. **DB-backed config** — `config.Load` reads a file path today; `PRICING_V3_DESIGN.md` Phần 19.3 discusses Postgres-versioned config with effective-dating (BRB §1.4's "30-day advance notice" requirement) as the production-grade next step.
3. **Cache the per-call pipeline** (Phần 9/13 mục 1).
4. **Bike/XL 1km-equivalent audit** (Phần 13 mục 2) — quantify whether the same overpricing defect exists for those classes before assuming it doesn't.
5. **`last_tier_min_ratio` sign-off** — the shipped default (0.35) is a guardrail threshold, not a commercial long-haul rate; CFO/CPO should confirm it (or a different value) explicitly, per `PRICING_V3_REVIEW.md` Phần 13 mục 2's original recommendation.
6. **Live Promotion Engine wiring** at the Booking Service layer (unchanged scope this sprint, Phần 13 mục 3).
7. **Driver minimum-earning top-up** at Settlement time, consistent with how `simulation/pricing_simulator.go` already models it (Phần 13 mục 4).

---

## OUTPUT SUMMARY (as required by "YÊU CẦU QUAN TRỌNG" §5)

**Danh sách file sửa:** Phần 2 above (12 new source files + 2 new packages' worth of tests + 3 modified files — `grpc/handler.go`, `cmd/server/main.go`, `go.mod`/`go.sum` — full list with purpose in Phần 2.1/2.2).

**Kết quả `go test ./...`:** all packages `ok` — `app`, `config`, `domain/entity`, `grpc`, `simulation` (see Phần 11 for the exact coverage numbers; zero failing tests, including every pre-existing V2 test, unmodified).

**Benchmark:** Phần 10 (raw `go test -bench -benchmem` output).

**Coverage:** Phần 11.

**`flutter analyze`:** not run — this sprint touched no Flutter/Rider/Driver UI code (`PHẠM VI` explicitly excluded it), so there is nothing for it to analyze.

**CHANGELOG:** see `CHANGELOG.md` — a new `[Unreleased] → Added` entry for this sprint was added in the same pass as this document.

---

*Kết thúc tài liệu — Panda Pricing V3 Implementation — v0.1.*
*Không commit. Không push. Pricing V3 is implemented and tested but NOT active in any environment (`PRICING_VERSION` defaults to `v2` everywhere) until CPO/CFO/CTO explicitly decide to flip it, per the Migration steps in Phần 3.*
