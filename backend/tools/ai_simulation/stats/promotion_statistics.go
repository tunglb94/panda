package stats

import "github.com/fairride/ai_simulation/domain/entity"

type PromotionStatistics struct {
	TripsWithPromotion int              `json:"trips_with_promotion"`
	TotalPromotionCostVND int64         `json:"total_promotion_cost_vnd"`
	AverageDiscountVND float64          `json:"average_discount_vnd"`
	CostByType         map[string]int64 `json:"cost_by_type_vnd"`
	CountByType        map[string]int   `json:"count_by_type"`
}

// BuildPromotionStatistics covers every applied promotion type EXCEPT
// manual_coupon, which is broken out separately in voucher_statistics.json
// (BRB distinguishes campaign-wide Promotions, Part 3, from individually
// redeemed Vouchers, Part 4 — manual_coupon is this simulation's Part-4-style
// code-redeemed case).
func (c *Collector) BuildPromotionStatistics(trips []*entity.SimTrip) PromotionStatistics {
	out := PromotionStatistics{CostByType: map[string]int64{}, CountByType: map[string]int{}}
	for _, t := range trips {
		if t.PromotionType == "" || t.PromotionType == "manual_coupon" {
			continue
		}
		out.TripsWithPromotion++
		out.TotalPromotionCostVND += t.VoucherDiscountVND
		out.CostByType[t.PromotionType] += t.VoucherDiscountVND
		out.CountByType[t.PromotionType]++
	}
	out.AverageDiscountVND = avg(float64(out.TotalPromotionCostVND), out.TripsWithPromotion)
	return out
}
