package stats

import "github.com/fairride/ai_simulation/domain/entity"

// Per-trip infrastructure cost assumptions (PHẦN 3's "Estimated Cloud/Map/SMS
// Cost"). These are NOT sourced from any real Panda finance/cost-accounting
// data — no such document exists — they are simulation-design placeholders
// set at plausible small-scale Vietnamese SaaS/telecom market rates (a
// managed Postgres+compute footprint amortized per trip; a per-request
// routing/geocoding API call; one OTP/status SMS via a local gateway like
// eSMS/Viettel, typically ~300-350đ/message). Labeled "Estimated" in every
// field name, exactly as the sprint brief itself requests ("Estimated Cloud
// Cost"), signaling these are assumptions, not measured costs.
const (
	assumedCloudCostVND = 400
	assumedMapCostVND   = 250
	assumedSMSCostVND   = 350

	// vatRatePercent is Vietnam's standard VAT rate — a public tax-law fact,
	// not a BRB number. BRB itself explicitly excludes VAT ("Platform taxes
	// ... are calculated and remitted by the Finance team independently of
	// this document", business-rule-bible-v1.0.md line 986), so this
	// simulation applies the standard rate to platform commission revenue
	// as a disclosed assumption rather than leaving VAT unmodeled.
	vatRatePercent = 10
)

// UnitEconomicsSegment is one Ride/Delivery/overall slice of the per-trip
// money waterfall PHẦN 3 asks for: Khách trả -> Voucher -> Promotion ->
// Driver -> Commission -> VAT -> Platform -> Estimated Cloud/Map/SMS Cost ->
// Estimated Profit.
type UnitEconomicsSegment struct {
	SampleSize                int     `json:"sample_size_completed_trips"`
	AverageCustomerPaysVND    float64 `json:"average_customer_pays_vnd"`
	AverageVoucherVND         float64 `json:"average_voucher_vnd"`
	AveragePromotionVND       float64 `json:"average_promotion_vnd"`
	AverageDriverNetVND       float64 `json:"average_driver_vnd"`
	AverageCommissionVND      float64 `json:"average_commission_vnd"`
	AverageVATVND             float64 `json:"average_vat_vnd"`
	AveragePlatformNetVND     float64 `json:"average_platform_net_vnd"` // commission - VAT
	EstimatedCloudCostVND     float64 `json:"estimated_cloud_cost_vnd"`
	EstimatedMapCostVND       float64 `json:"estimated_map_cost_vnd"`
	EstimatedSMSCostVND       float64 `json:"estimated_sms_cost_vnd"`
	EstimatedProfitVND        float64 `json:"estimated_profit_vnd"`
}

type UnitEconomics struct {
	Overall  UnitEconomicsSegment `json:"overall"`
	Ride     UnitEconomicsSegment `json:"ride"`
	Delivery UnitEconomicsSegment `json:"delivery"`
}

func (c *Collector) BuildUnitEconomics(trips []*entity.SimTrip) UnitEconomics {
	return UnitEconomics{
		Overall:  buildUnitEconomicsSegment(trips, ""),
		Ride:     buildUnitEconomicsSegment(trips, entity.KindRide),
		Delivery: buildUnitEconomicsSegment(trips, entity.KindDelivery),
	}
}

// buildUnitEconomicsSegment aggregates over completed trips matching kind;
// an empty kind means "every completed trip regardless of Kind".
func buildUnitEconomicsSegment(trips []*entity.SimTrip, kind entity.TripKind) UnitEconomicsSegment {
	var seg UnitEconomicsSegment
	var customerSum, voucherSum, promoSum, driverSum, commissionSum float64

	for _, t := range trips {
		if t.Outcome != entity.OutcomeCompleted {
			continue
		}
		if kind != "" && t.Kind != kind {
			continue
		}
		seg.SampleSize++
		customerSum += float64(t.FinalFareVND)
		driverSum += float64(t.DriverNetVND)
		commissionSum += float64(t.CommissionVND)
		if t.PromotionType == "manual_coupon" {
			voucherSum += float64(t.VoucherDiscountVND)
		} else if t.PromotionType != "" {
			promoSum += float64(t.VoucherDiscountVND)
		}
	}

	seg.AverageCustomerPaysVND = avg(customerSum, seg.SampleSize)
	seg.AverageVoucherVND = avg(voucherSum, seg.SampleSize)
	seg.AveragePromotionVND = avg(promoSum, seg.SampleSize)
	seg.AverageDriverNetVND = avg(driverSum, seg.SampleSize)
	seg.AverageCommissionVND = avg(commissionSum, seg.SampleSize)
	seg.AverageVATVND = seg.AverageCommissionVND * vatRatePercent / 100
	seg.AveragePlatformNetVND = seg.AverageCommissionVND - seg.AverageVATVND
	if seg.SampleSize > 0 {
		seg.EstimatedCloudCostVND = assumedCloudCostVND
		seg.EstimatedMapCostVND = assumedMapCostVND
		seg.EstimatedSMSCostVND = assumedSMSCostVND
	}
	seg.EstimatedProfitVND = estimatedPlatformProfitVND(seg.AverageCommissionVND, seg.SampleSize > 0)
	return seg
}

// EstimatedInfraCostPerTripVND exports the sum of the 3 per-trip cost
// assumptions above — for callers (bi package) that need the raw platform-
// cost total rather than a profit figure, without duplicating the
// constants a second time.
func EstimatedInfraCostPerTripVND() int64 {
	return assumedCloudCostVND + assumedMapCostVND + assumedSMSCostVND
}

// VATRatePercent exports the standard VAT rate this simulation applies to
// commission revenue — see the const block above for sourcing.
func VATRatePercent() int64 { return vatRatePercent }

// PerTripProfitVND is estimatedPlatformProfitVND's exported form for a
// single completed trip's commission — used by the audit package's
// per-trip profit checks (PHẦN 2/6/16/17/18 of the business audit) so that
// package reuses the exact same VAT/cost-assumption constants instead of
// duplicating them.
func PerTripProfitVND(commissionVND int64) int64 {
	return int64(estimatedPlatformProfitVND(float64(commissionVND), true))
}

// estimatedPlatformProfitVND applies the standard VAT rate then subtracts
// the per-trip infrastructure cost assumptions above — shared by
// unit_economics.json and pricing_analytics.json so the VAT/cost-assumption
// logic exists in exactly one place. hasCosts is false for an empty
// sample (no trips -> no costs to subtract either).
func estimatedPlatformProfitVND(commissionVND float64, hasCosts bool) float64 {
	vat := commissionVND * vatRatePercent / 100
	net := commissionVND - vat
	if !hasCosts {
		return net
	}
	return net - assumedCloudCostVND - assumedMapCostVND - assumedSMSCostVND
}
