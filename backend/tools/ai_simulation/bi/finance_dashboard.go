package bi

import (
	"github.com/fairride/ai_simulation/domain/entity"
	"github.com/fairride/ai_simulation/stats"
)

// FinanceDashboard is PHẦN 11 — reuses stats.BusinessIntelligence for every
// field that already exists there (GMV/PlatformRevenue/DriverRevenue/
// VoucherCost/PromotionCost/Profit/Margin); only PlatformCostVND (infra),
// TaxVND (=VAT, aliased), NetMarginPercent, and Profit/Ride vs Profit/
// Delivery are newly derived here.
type FinanceDashboard struct {
	PlatformRevenueVND int64 `json:"platform_revenue_vnd"`
	PlatformCostVND     int64 `json:"platform_cost_vnd"` // estimated infra cost only — see unit_economics.json's cost assumptions
	DriverRevenueVND    int64 `json:"driver_revenue_vnd"`
	VoucherCostVND      int64 `json:"voucher_cost_vnd"`
	PromotionCostVND    int64 `json:"promotion_cost_vnd"`
	RefundVND           int64 `json:"refund_vnd"` // always 0 — see Assumptions
	CommissionVND       int64 `json:"commission_vnd"`
	TaxVND              int64 `json:"tax_vnd"` // VAT at the standard 10% rate — see unit_economics.json
	GrossMarginPercent  float64 `json:"gross_margin_percent"` // (GMV - driver revenue) / GMV
	NetMarginPercent    float64 `json:"net_margin_percent"`   // profit / GMV
	ProfitPerRideVND    float64 `json:"profit_per_ride_vnd"`
	ProfitPerDeliveryVND float64 `json:"profit_per_delivery_vnd"`

	Assumptions []Assumption `json:"assumptions"`
}

// ComputeFinanceDashboard is PHẦN 11.
func ComputeFinanceDashboard(in Input) FinanceDashboard {
	bi := in.BI
	out := FinanceDashboard{
		PlatformRevenueVND: bi.PlatformRevenueVND, DriverRevenueVND: bi.DriverRevenueVND,
		VoucherCostVND: bi.VoucherCostVND, PromotionCostVND: bi.PromotionCostVND,
		CommissionVND: bi.PlatformRevenueVND, // Commission IS platform revenue in this simulation — no separate non-commission revenue line exists
		TaxVND:        int64(float64(bi.PlatformRevenueVND) * float64(stats.VATRatePercent()) / 100),
	}

	var rideProfitSum, deliveryProfitSum int64
	var rideCount, deliveryCount, completedCount int
	for _, t := range in.Trips {
		if t.Outcome != entity.OutcomeCompleted {
			continue
		}
		completedCount++
		p := profitFor(t)
		if t.Kind == entity.KindDelivery {
			deliveryProfitSum += p
			deliveryCount++
		} else {
			rideProfitSum += p
			rideCount++
		}
	}
	out.PlatformCostVND = int64(completedCount) * stats.EstimatedInfraCostPerTripVND()
	if rideCount > 0 {
		out.ProfitPerRideVND = float64(rideProfitSum) / float64(rideCount)
	}
	if deliveryCount > 0 {
		out.ProfitPerDeliveryVND = float64(deliveryProfitSum) / float64(deliveryCount)
	}

	if bi.GMVVND > 0 {
		out.GrossMarginPercent = 100 * float64(bi.GMVVND-bi.DriverRevenueVND) / float64(bi.GMVVND)
		out.NetMarginPercent = 100 * float64(bi.ProfitVND) / float64(bi.GMVVND)
	}

	out.Assumptions = []Assumption{
		{Title: "Refund luôn = 0", Detail: "Simulation không có cơ chế hoàn tiền (refund) — Promotion/Voucher Engine chỉ áp giảm giá tại thời điểm đặt, không có luồng hoàn tiền sau khi hoàn tất chuyến."},
		{Title: "Platform Cost chỉ gồm chi phí hạ tầng ước tính", Detail: "Không bao gồm chi phí nhân sự/marketing/vận hành thật của Panda — chỉ Cloud/Map/SMS như unit_economics.json đã định nghĩa."},
	}
	return out
}
