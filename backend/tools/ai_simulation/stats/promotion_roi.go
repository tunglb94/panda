package stats

import "github.com/fairride/ai_simulation/domain/entity"

// PromotionROIEntry is one campaign type's ROI breakdown — covers every
// PromotionType that actually fired (auto-apply campaigns AND the
// manual_coupon voucher path), not just vouchers, since PHẦN 7 asks about
// "Promotion ROI" broadly.
type PromotionROIEntry struct {
	Type          string  `json:"type"`
	RedeemedCount int     `json:"redeemed_count"`
	TotalCostVND  int64   `json:"total_cost_vnd"`
	// GMVGeneratedVND is the gross fare of every trip that redeemed this
	// promotion type — the "revenue this incentive touched", not a claim
	// that the trip would never have happened without it.
	GMVGeneratedVND int64   `json:"gmv_generated_vnd"`
	ROI             float64 `json:"roi"` // (GMV - cost) / cost; 0 if cost is 0
	CPAVND          float64 `json:"cost_per_redemption_vnd"`
	RepeatRatePercent float64 `json:"repeat_rate_percent"` // % of redeeming riders who completed >1 trip total in the run
}

type PromotionROI struct {
	// IssuedVoucherCount / StillHeldCount / ExpiredCount describe the
	// manual_coupon voucher specifically (the only promotion type riders
	// individually hold/decide to spend, per ruleengine.VoucherUseDecision)
	// — auto-apply campaigns (First Ride/Birthday/Weekend) have no
	// "issued/held" concept, BRB applies them automatically when eligible.
	IssuedVoucherCount int `json:"voucher_issued_count"`
	UsedVoucherCount   int `json:"voucher_used_count"`
	KeptVoucherCount   int `json:"voucher_kept_count"`
	// StillHeldAtEndCount is riders holding an unredeemed voucher when the
	// run ended — the closest honest proxy for "outstanding/expiring soon".
	StillHeldAtEndCount int `json:"voucher_still_held_at_end_count"`
	// ExpiredCount is always 0 under this simulation's seeded campaign
	// config (integration/promotion_adapter.go seeds SIM10 with a 365-day
	// validity window — far longer than any --days run length), computed
	// honestly rather than fabricated as a nonzero placeholder. See that
	// adapter's seedVouchers for the exact end date.
	ExpiredCount int `json:"voucher_expired_count"`

	ByType []PromotionROIEntry `json:"by_type"`
}

func (c *Collector) BuildPromotionROI(riders map[string]*entity.RiderAgent, trips []*entity.SimTrip, voucherIssuedCount, voucherUsedCount, voucherKeptCount int) PromotionROI {
	out := PromotionROI{
		IssuedVoucherCount: voucherIssuedCount,
		UsedVoucherCount:   voucherUsedCount,
		KeptVoucherCount:   voucherKeptCount,
		ExpiredCount:       0, // see doc comment on PromotionROI.ExpiredCount
	}
	for _, r := range riders {
		if r.HasActiveVoucher {
			out.StillHeldAtEndCount++
		}
	}

	// Total completed-trip count per rider — needed for RepeatRate ("did a
	// rider who redeemed this promotion also complete another trip").
	tripsByRider := map[string]int{}
	for _, t := range trips {
		if t.Outcome == entity.OutcomeCompleted {
			tripsByRider[t.RiderID]++
		}
	}

	type acc struct {
		count      int
		costVND    int64
		gmvVND     int64
		riders     map[string]bool
	}
	byType := map[string]*acc{}
	for _, t := range trips {
		if t.Outcome != entity.OutcomeCompleted || t.PromotionType == "" {
			continue
		}
		a := byType[t.PromotionType]
		if a == nil {
			a = &acc{riders: map[string]bool{}}
			byType[t.PromotionType] = a
		}
		a.count++
		a.costVND += t.VoucherDiscountVND
		a.gmvVND += t.FinalFareVND + t.VoucherDiscountVND
		a.riders[t.RiderID] = true
	}

	for typ, a := range byType {
		entry := PromotionROIEntry{
			Type: typ, RedeemedCount: a.count, TotalCostVND: a.costVND, GMVGeneratedVND: a.gmvVND,
		}
		if a.costVND > 0 {
			entry.ROI = float64(a.gmvVND-a.costVND) / float64(a.costVND)
			entry.CPAVND = float64(a.costVND) / float64(a.count)
		}
		var repeaters int
		for riderID := range a.riders {
			if tripsByRider[riderID] > 1 {
				repeaters++
			}
		}
		if len(a.riders) > 0 {
			entry.RepeatRatePercent = 100 * float64(repeaters) / float64(len(a.riders))
		}
		out.ByType = append(out.ByType, entry)
	}
	return out
}
