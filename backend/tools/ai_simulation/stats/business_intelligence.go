package stats

import "github.com/fairride/ai_simulation/domain/entity"

// BusinessIntelligence is the exact PHẦN 2 field list — a CEO-facing
// summary computed directly from the trip ledger, the single source of
// truth every other JSON export also derives from (no double bookkeeping).
type BusinessIntelligence struct {
	GMVVND             int64   `json:"gmv_vnd"`
	NetRevenueVND      int64   `json:"net_revenue_vnd"` // platform commission - voucher cost - promotion cost
	PlatformRevenueVND int64   `json:"platform_revenue_vnd"`
	DriverRevenueVND   int64   `json:"driver_revenue_vnd"`
	ProfitVND          int64   `json:"profit_vnd"` // net revenue after VAT and estimated infra cost — see unit_economics.json's cost assumptions
	PlatformMarginPercent float64 `json:"platform_margin_percent"` // platform revenue / GMV
	VoucherCostVND     int64   `json:"voucher_cost_vnd"`
	PromotionCostVND   int64   `json:"promotion_cost_vnd"`

	DriverRetentionPercent    float64 `json:"driver_retention_percent"`
	PassengerRetentionPercent float64 `json:"passenger_retention_percent"`

	RideRevenueVND     int64   `json:"ride_revenue_vnd"`
	DeliveryRevenueVND int64   `json:"delivery_revenue_vnd"`
	RidePercent        float64 `json:"ride_percent"`     // share of completed trips
	DeliveryPercent    float64 `json:"delivery_percent"` // share of completed trips

	AverageFareVND          float64 `json:"average_fare_vnd"`
	AverageETAMinutes       float64 `json:"average_eta_minutes"`
	AcceptanceRatePercent   float64 `json:"acceptance_rate_percent"`
	CancellationRatePercent float64 `json:"cancellation_rate_percent"`
}

// BuildBusinessIntelligence computes the executive summary directly from
// trips plus the two retention figures (already computed by
// BuildDriverAnalytics/BuildRiderAnalytics — not re-derived here, to avoid
// two slightly different retention formulas existing in the codebase).
func (c *Collector) BuildBusinessIntelligence(trips []*entity.SimTrip, driverRetentionPercent, riderRetentionPercent float64) BusinessIntelligence {
	var bi BusinessIntelligence
	bi.DriverRetentionPercent = driverRetentionPercent
	bi.PassengerRetentionPercent = riderRetentionPercent

	var requested, completed, cancelled int
	var fareSum, etaSum float64
	var etaN int
	var rideCompleted, deliveryCompleted int

	for _, t := range trips {
		requested++
		switch t.Outcome {
		case entity.OutcomeCompleted:
			completed++
			gmv := t.FinalFareVND + t.VoucherDiscountVND
			bi.GMVVND += gmv
			bi.PlatformRevenueVND += t.CommissionVND
			bi.DriverRevenueVND += t.DriverNetVND
			fareSum += float64(t.FinalFareVND)
			if t.PromotionType == "manual_coupon" {
				bi.VoucherCostVND += t.VoucherDiscountVND
			} else if t.PromotionType != "" {
				bi.PromotionCostVND += t.VoucherDiscountVND
			}
			if t.Kind == entity.KindDelivery {
				deliveryCompleted++
				bi.DeliveryRevenueVND += gmv
			} else {
				rideCompleted++
				bi.RideRevenueVND += gmv
			}
			if t.ETAMinutes > 0 {
				etaSum += t.ETAMinutes
				etaN++
			}
		case entity.OutcomeCancelled:
			cancelled++
		}
	}

	bi.NetRevenueVND = bi.PlatformRevenueVND - bi.VoucherCostVND - bi.PromotionCostVND
	// Profit = platform commission, less VAT (see unit_economics.go's
	// vatRatePercent doc comment), less estimated per-trip infra cost
	// scaled by trip volume, less voucher/promotion spend. Uses the same
	// per-trip cost assumptions unit_economics.json/pricing_analytics.json
	// use, scaled here by completed trip count rather than averaged.
	totalVATVND := float64(bi.PlatformRevenueVND) * vatRatePercent / 100
	totalInfraCostVND := float64(completed) * (assumedCloudCostVND + assumedMapCostVND + assumedSMSCostVND)
	bi.ProfitVND = bi.PlatformRevenueVND - int64(totalVATVND) - int64(totalInfraCostVND) - bi.VoucherCostVND - bi.PromotionCostVND
	if bi.GMVVND > 0 {
		bi.PlatformMarginPercent = 100 * float64(bi.PlatformRevenueVND) / float64(bi.GMVVND)
	}
	bi.AverageFareVND = avg(fareSum, completed)
	bi.AverageETAMinutes = avg(etaSum, etaN)
	if requested > 0 {
		bi.AcceptanceRatePercent = 100 * float64(completed) / float64(requested)
		bi.CancellationRatePercent = 100 * float64(cancelled) / float64(requested)
	}
	if completed > 0 {
		bi.RidePercent = 100 * float64(rideCompleted) / float64(completed)
		bi.DeliveryPercent = 100 * float64(deliveryCompleted) / float64(completed)
	}
	return bi
}
