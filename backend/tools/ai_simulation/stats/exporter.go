package stats

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Bundle holds every exported statistics document for one simulation run —
// exactly the 8 files the sprint brief names.
type Bundle struct {
	SimulationReport    SimulationReport
	DriverStatistics    DriverStatistics
	RiderStatistics     RiderStatistics
	PricingStatistics   PricingStatistics
	PromotionStatistics PromotionStatistics
	VoucherStatistics   VoucherStatistics
	DispatchStatistics  DispatchStatistics
	DeliveryStatistics  DeliveryStatistics
	Heatmap             Heatmap

	UnitEconomics    UnitEconomics
	DriverAnalytics  DriverAnalytics
	RiderAnalytics   RiderAnalytics
	PricingAnalytics PricingAnalytics
	PromotionROI     PromotionROI
}

// Export writes every document in the bundle to its own JSON file under dir.
func (b Bundle) Export(dir string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create output dir: %w", err)
	}
	files := map[string]any{
		"simulation_report.json":    b.SimulationReport,
		"driver_statistics.json":    b.DriverStatistics,
		"rider_statistics.json":     b.RiderStatistics,
		"pricing_statistics.json":   b.PricingStatistics,
		"promotion_statistics.json": b.PromotionStatistics,
		"voucher_statistics.json":   b.VoucherStatistics,
		"dispatch_statistics.json":  b.DispatchStatistics,
		"delivery_statistics.json":  b.DeliveryStatistics,
		"heatmap.json":              b.Heatmap,
		"unit_economics.json":       b.UnitEconomics,
		"driver_analytics.json":     b.DriverAnalytics,
		"rider_analytics.json":      b.RiderAnalytics,
		"pricing_analytics.json":    b.PricingAnalytics,
		"promotion_roi.json":        b.PromotionROI,
	}
	for name, doc := range files {
		data, err := json.MarshalIndent(doc, "", "  ")
		if err != nil {
			return fmt.Errorf("marshal %s: %w", name, err)
		}
		if err := os.WriteFile(filepath.Join(dir, name), data, 0o644); err != nil {
			return fmt.Errorf("write %s: %w", name, err)
		}
	}
	return nil
}
