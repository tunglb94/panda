package stats

import "github.com/fairride/ai_simulation/domain/entity"

type VoucherStatistics struct {
	TripsWithVoucher    int     `json:"trips_with_voucher"`
	TotalVoucherCostVND int64   `json:"total_voucher_cost_vnd"`
	AverageDiscountVND  float64 `json:"average_discount_vnd"`

	// From ruleengine.VoucherUseDecision — how often riders chose to spend a
	// held voucher now vs keep it, including the ambiguous cases AI decided.
	UsedCount int `json:"used_count"`
	KeptCount int `json:"kept_count"`
}

// BuildVoucherStatistics covers manual_coupon (code-redeemed) trips — see
// BuildPromotionStatistics's doc comment for the Part 3 / Part 4 split this
// mirrors. UsedCount/KeptCount are populated by the simulation engine
// directly (ride_flow.go), not derivable from SimTrip alone, so they're set
// via the returned struct's fields by the caller after Build returns.
func (c *Collector) BuildVoucherStatistics(trips []*entity.SimTrip) VoucherStatistics {
	out := VoucherStatistics{}
	for _, t := range trips {
		if t.PromotionType != "manual_coupon" {
			continue
		}
		out.TripsWithVoucher++
		out.TotalVoucherCostVND += t.VoucherDiscountVND
	}
	out.AverageDiscountVND = avg(float64(out.TotalVoucherCostVND), out.TripsWithVoucher)
	return out
}
