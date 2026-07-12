package entity

// DiscountType identifies how Voucher.DiscountValue is interpreted.
type DiscountType string

const (
	// DiscountTypePercentage: DiscountValue is a whole-number percentage (0-100)
	// of the order amount, capped by MaxDiscount (BRB §3.2.1-3.2.6 all express
	// discounts this way, e.g. "50% off, maximum 30,000 VND").
	DiscountTypePercentage DiscountType = "percentage"

	// DiscountTypeFlat: DiscountValue is a fixed VND amount (BRB §3.2.7 Referral:
	// "Rider B receives 30,000 VND off").
	DiscountTypeFlat DiscountType = "flat"
)

func (d DiscountType) Valid() bool {
	return d == DiscountTypePercentage || d == DiscountTypeFlat
}

// ComputeDiscount returns the raw discount amount for orderAmount (VND) before
// the BRB §4.9 "cannot exceed order amount" clamp is applied by the caller.
func ComputeDiscount(discountType DiscountType, discountValue, maxDiscount, orderAmount int64) int64 {
	var raw int64
	switch discountType {
	case DiscountTypePercentage:
		raw = orderAmount * discountValue / 100
	case DiscountTypeFlat:
		raw = discountValue
	default:
		return 0
	}
	if maxDiscount > 0 && raw > maxDiscount {
		raw = maxDiscount
	}
	if raw < 0 {
		raw = 0
	}
	return raw
}
