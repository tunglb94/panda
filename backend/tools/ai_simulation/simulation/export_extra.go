package simulation

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/fairride/ai_simulation/aiengine"
	"github.com/fairride/ai_simulation/audit"
	"github.com/fairride/ai_simulation/benchmark"
	"github.com/fairride/ai_simulation/bi"
	"github.com/fairride/ai_simulation/dashboard"
	"github.com/fairride/ai_simulation/domain/entity"
	"github.com/fairride/ai_simulation/insights"
	"github.com/fairride/ai_simulation/stats"
)

// writeValidationReport runs stats.Collector.Validate (PHẦN 11's self-check
// list), writes it, and returns the computed report so callers (the audit
// package's business_audit.md/CEO_report.html/top_20_business_risks.md)
// can reuse it instead of calling Validate a second time. Always succeeds
// regardless of what Validate finds — anomalies land as warnings inside the
// JSON, never as a crash or a non-zero exit code.
func writeValidationReport(dir string, w *World, trips []*entity.SimTrip, bi stats.BusinessIntelligence) (stats.ValidationReport, error) {
	report := w.Stats.Validate(trips, w.Drivers, bi)
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return report, fmt.Errorf("marshal validation report: %w", err)
	}
	return report, os.WriteFile(filepath.Join(dir, "validation_report.json"), data, 0o644)
}

// writeBenchmarkReport writes the AI-vs-Rule-Engine / cache / performance
// summary the sprint brief's "Benchmark" section asks for.
func writeBenchmarkReport(dir string, report benchmark.Report) error {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal benchmark report: %w", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "benchmark_report.json"), data, 0o644); err != nil {
		return fmt.Errorf("write benchmark report: %w", err)
	}
	return nil
}

// writeDashboard generates the single-file dashboard.html.
func writeDashboard(dir string, bundle stats.Bundle) error {
	html := dashboard.Generate(bundle)
	return os.WriteFile(filepath.Join(dir, "dashboard.html"), []byte(html), 0o644)
}

// writeExecutiveDashboard generates the single-file executive_dashboard.html
// — the CEO-facing summary PHẦN 2 asks for ("CEO chỉ mở 1 file này").
func writeExecutiveDashboard(dir string, bi stats.BusinessIntelligence, bundle stats.Bundle) error {
	html := dashboard.GenerateExecutive(bi, bundle)
	return os.WriteFile(filepath.Join(dir, "executive_dashboard.html"), []byte(html), 0o644)
}

// writeInsightReports generates simulation_summary.md and
// business_recommendation.md (PHẦN 9/10) — findings/recommendations are
// always Rule-Engine-computed from real data (insights.ComputeFindings/
// ComputeRecommendations); ai, if non-nil and reachable, only rephrases
// them (see insights.WriteSummary's doc comment). A nil/unreachable ai
// still produces a complete, fully data-grounded report.
func writeInsightReports(ctx context.Context, dir string, ai *aiengine.DecisionEngine, trips []*entity.SimTrip, bundle stats.Bundle, bi stats.BusinessIntelligence) ([]insights.Finding, []insights.Recommendation, error) {
	findings := insights.ComputeFindings(trips, bundle, bi)
	recs := insights.ComputeRecommendations(trips, bundle, bi)

	summaryMD := insights.WriteSummary(ctx, ai, findings)
	if err := os.WriteFile(filepath.Join(dir, "simulation_summary.md"), []byte(summaryMD), 0o644); err != nil {
		return findings, recs, fmt.Errorf("write simulation_summary.md: %w", err)
	}
	recMD := insights.WriteRecommendations(ctx, ai, recs)
	if err := os.WriteFile(filepath.Join(dir, "business_recommendation.md"), []byte(recMD), 0o644); err != nil {
		return findings, recs, fmt.Errorf("write business_recommendation.md: %w", err)
	}
	return findings, recs, nil
}

// writeBusinessAudit produces the Full Business Validation's 6 required
// artifacts (validation_report.html, CEO_report.html, business_audit.md,
// top_50_anomalies.json, top_20_business_risks.md, top_30_optimization.md).
// Every number comes from ComputeReport (backend/tools/ai_simulation/audit)
// — this function only renders/writes, per the task's explicit "chỉ chạy
// simulation và tạo báo cáo, không tự sửa bug" instruction.
func writeBusinessAudit(
	dir string, w *World, trips []*entity.SimTrip, bundle stats.Bundle, bi stats.BusinessIntelligence,
	validation stats.ValidationReport, findings []insights.Finding, recs []insights.Recommendation,
) error {
	report := audit.ComputeReport(trips, w.Drivers, w.Riders, bundle, validation)
	anomalies := audit.ComputeTop50Anomalies(trips, w.Drivers, report)
	bugs := audit.DetectBugs(validation)
	assumptions := audit.KnownAssumptions()
	risks := audit.ComputeTop20Risks(report)

	writeFile := func(name string, data []byte) error {
		return os.WriteFile(filepath.Join(dir, name), data, 0o644)
	}

	if err := writeFile("validation_report.html", []byte(audit.RenderValidationHTML(validation))); err != nil {
		return fmt.Errorf("write validation_report.html: %w", err)
	}
	if err := writeFile("CEO_report.html", []byte(audit.RenderCEOHTML(bi, report, bundle.DriverAnalytics.RetentionRatePercent, bundle.RiderAnalytics.RetentionRatePercent, findings, bugs, recs))); err != nil {
		return fmt.Errorf("write CEO_report.html: %w", err)
	}
	if err := writeFile("business_audit.md", []byte(audit.RenderBusinessAuditMarkdown(report, validation, assumptions, bugs))); err != nil {
		return fmt.Errorf("write business_audit.md: %w", err)
	}
	anomaliesJSON, err := json.MarshalIndent(anomalies, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal top_50_anomalies.json: %w", err)
	}
	if err := writeFile("top_50_anomalies.json", anomaliesJSON); err != nil {
		return fmt.Errorf("write top_50_anomalies.json: %w", err)
	}
	if err := writeFile("top_20_business_risks.md", []byte(audit.RenderTop20RisksMarkdown(risks))); err != nil {
		return fmt.Errorf("write top_20_business_risks.md: %w", err)
	}
	// top_30_optimization.md reuses the exact same Recommendation set
	// business_recommendation.md already rendered (same AI-optional
	// polish pattern) — this task's exact requested filename, no
	// duplicated recommendation logic.
	optMD := insights.WriteRecommendations(context.Background(), w.AI, recs)
	if err := writeFile("top_30_optimization.md", []byte(optMD)); err != nil {
		return fmt.Errorf("write top_30_optimization.md: %w", err)
	}
	return nil
}

// driverEconomyExport/passengerEconomyExport/businessAlertsExport are the
// small wrapper structs that let §2 (Driver Distribution) ride along inside
// driver_economy.json, §14 (Passenger Segment) inside passenger_economy.json,
// and §20 (Realism Score) inside business_alerts.json — PHẦN 19's file list
// names exactly 12 files, and these 3 sections don't have their own —
// folding them into the closest-related file avoids inventing filenames
// the brief never asked for.
type driverEconomyExport struct {
	bi.DriverEconomyReport
	Distribution bi.DriverDistribution `json:"distribution"`
}

type passengerEconomyExport struct {
	bi.PassengerEconomy
	Segments bi.PassengerSegmentReport `json:"segments"`
}

type businessAlertsExport struct {
	Alerts      []bi.Alert          `json:"alerts"`
	Realism     bi.RealismScoreReport `json:"realism_score"`
}

// writeBusinessIntelligence produces PHẦN 1-20's full Business Intelligence
// layer (backend/tools/ai_simulation/bi) — 12 new JSON exports, every
// number read from real SimTrip/DriverAgent/RiderAgent records or reused
// from stats.Bundle/insights, never fabricated (see bi package's own doc
// comment). recs is the same Recommendation set writeBusinessAudit already
// rendered — reused here, not recomputed, then merged with this package's
// own additional rule candidates (PHẦN 18's named examples).
func writeBusinessIntelligence(dir string, w *World, cfg Config, trips []*entity.SimTrip, bundle stats.Bundle, businessIntel stats.BusinessIntelligence, recs []insights.Recommendation) error {
	in := bi.Input{
		Trips: trips, Drivers: w.Drivers, Riders: w.Riders, Bundle: bundle, BI: businessIntel, Days: cfg.Days,
		FatigueContinueCount: w.fatigueContinueCount, FatigueStopCount: w.fatigueStopCount,
		SwitchAppCount: w.switchAppCount, StayOnPandaCount: w.stayOnPandaCount,
		SurgeChaseCount: w.surgeChaseCount, SurgeStayCount: w.surgeStayCount,
		DriverOffersAccepted: w.driverOffersAccepted, DriverOffersRejected: w.driverOffersRejected,
		VoucherUsedCount: w.voucherUsedCount, VoucherKeptCount: w.voucherKeptCount,
	}

	driverEconomy := bi.ComputeDriverEconomy(in)
	driverDistribution := bi.ComputeDriverDistribution(in)
	passengerEconomy := bi.ComputePassengerEconomy(in)
	passengerSegments := bi.ComputePassengerSegments(in)
	pricingBreakdown := bi.ComputePricingBreakdown(in)
	surgeAnalysis := bi.ComputeSurgeAnalysis(in)
	weatherAnalysis := bi.ComputeWeatherAnalysis(in)
	trafficAnalysis := bi.ComputeTrafficAnalysis(in)
	driverBehavior := bi.ComputeDriverBehavior(in)
	passengerBehavior := bi.ComputePassengerBehavior(in)
	deliveryDashboard := bi.ComputeDeliveryDashboard(in)
	financeDashboard := bi.ComputeFinanceDashboard(in)
	driverLeaderboard := bi.ComputeDriverLeaderboard(in)
	airportAnalysis := bi.ComputeAirportAnalysis(in)
	alerts := bi.ComputeBusinessAlerts(in, driverEconomy)

	assumptionCounts := map[string]int{
		"pricing": 0, "promotion": 0, "dispatch": 0, "delivery": len(deliveryDashboard.Assumptions),
		"driver_economy": len(driverEconomy.Assumptions), "demand": 1, "supply": len(driverDistribution.Assumptions),
		"passenger_behavior": len(passengerBehavior.Assumptions) + len(passengerSegments.Assumptions),
		"traffic": 0, "weather": len(weatherAnalysis.Assumptions),
	}
	realism := bi.ComputeRealismScore(assumptionCounts)

	extraRecs := bi.ComputeAdditionalRecommendations(in, driverEconomy, pricingBreakdown, deliveryDashboard)
	allRecs := append(append([]insights.Recommendation{}, recs...), extraRecs...)
	sort.Slice(allRecs, func(i, j int) bool { return allRecs[i].Signal > allRecs[j].Signal })
	if len(allRecs) > 30 {
		allRecs = allRecs[:30]
	}

	writeJSON := func(name string, v any) error {
		data, err := json.MarshalIndent(v, "", "  ")
		if err != nil {
			return fmt.Errorf("marshal %s: %w", name, err)
		}
		return os.WriteFile(filepath.Join(dir, name), data, 0o644)
	}

	files := map[string]any{
		"driver_economy.json":         driverEconomyExport{DriverEconomyReport: driverEconomy, Distribution: driverDistribution},
		"passenger_economy.json":      passengerEconomyExport{PassengerEconomy: passengerEconomy, Segments: passengerSegments},
		"pricing_breakdown.json":      pricingBreakdown,
		"weather_analysis.json":       weatherAnalysis,
		"traffic_analysis.json":       trafficAnalysis,
		"surge_analysis.json":         surgeAnalysis,
		"city_dashboard.json":         bundle.Heatmap, // same enriched heatmap data city_dashboard.json asks for — see stats/heatmap.go's extended fields
		"finance_dashboard.json":      financeDashboard,
		"airport_analysis.json":       airportAnalysis,
		"business_alerts.json":        businessAlertsExport{Alerts: alerts, Realism: realism},
		"business_recommendations.json": allRecs,
		"driver_leaderboard.json":     driverLeaderboard,
	}
	for name, v := range files {
		if err := writeJSON(name, v); err != nil {
			return err
		}
	}

	// driver_behavior.json/passenger_behavior.json aren't in PHẦN 19's
	// explicit file list but PHẦN 8/9 explicitly ask to "thống kê" them —
	// exported as small additive files rather than silently dropped.
	if err := writeJSON("driver_behavior.json", driverBehavior); err != nil {
		return err
	}
	if err := writeJSON("passenger_behavior.json", passengerBehavior); err != nil {
		return err
	}
	return nil
}
