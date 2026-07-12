package entity

import (
	"fmt"
	"strings"
)

// FullFareBreakdownV3 is the complete, itemised Pricing V3 result — every
// line docs/business/PRICING_V3_DESIGN.md Phần 9/11 and the implementation
// sprint brief ask for: Base Fare, Distance Fare, Moving/Traffic/Waiting
// Time, Airport Fee, Surge, Voucher/Discount, Commission, VAT, Platform Fee,
// Driver Income, Platform Revenue, Final Fare.
//
// This is deliberately a SEPARATE type from FareBreakdown (the V2 struct in
// fare.go), not an extension of it — FareBreakdown is a wire-compatible
// shape the gRPC handler has returned since before this sprint (see
// grpc/handler.go); adding fields to it is safe (Go/protobuf both tolerate
// new fields), but Commission/VAT/DriverIncome/PlatformRevenue are NEW
// responsibilities the Pricing service has never had (see commission_v3.go
// doc comment) — keeping them on a new type makes that scope expansion
// explicit and reviewable, rather than silently growing the old type's
// meaning.
type FullFareBreakdownV3 struct {
	VehicleType  VehicleType
	DistanceKM   float64
	DurationMin  float64
	CurrencyCode string
	IsFinal      bool

	// ─── Metered components (PRICING_V3_DESIGN.md Phần 3/4/5) ────────────
	BaseFare     int64
	DistanceFare int64 // sum across DistanceTier bands, see DistanceFareForTiers

	// MovingTimeFare is always 0 — Moving Time (speed >= the slow-traffic
	// threshold) is already priced through DistanceFare, never billed a
	// second time. The field exists (rather than being omitted) so a caller
	// rendering an Explanation can show *why* it is zero instead of the line
	// simply being absent — PRICING_V3_DESIGN.md Phần 5.1.
	MovingTimeFare int64
	// TrafficTimeFare bills only the minutes spent below the slow-traffic
	// speed threshold (renamed from V2's "TimeFare" — same formula, clearer
	// name; see VehicleRatesV3.TrafficTimePerMinute).
	TrafficTimeFare int64
	// WaitingFee bills minutes after the driver marks "Arrived", beyond the
	// configured grace period — new to production in V3 (V2 never modelled
	// pre-trip waiting at all).
	WaitingFee int64

	// AirportFee is the Pickup or Dropoff fee actually applied (0 if the
	// trip is not an airport leg, or the vehicle class has no configured
	// fee — see AirportFeeConfigV3.FeeFor).
	AirportFee int64
	AirportLeg AirportLeg

	// RideFare = max(BaseFare+DistanceFare+MovingTimeFare+TrafficTimeFare,
	// MinimumFare) x SurgeMultiplier, + AirportFee + WaitingFee — see
	// app/fare_calculator_v3.go for the exact order of operations, which
	// matches BRB §2.13.4's stated order (surge applies to Base+Distance+
	// Time+Airport; Waiting/Booking are never surged).
	RideFare          int64
	MinimumFareForced bool

	// ─── Surge (computed by the EXISTING PricingPipeline/PricingRule engine
	// — this struct only records the outcome, it does not duplicate the
	// combination logic; see app/fare_calculator_v3.go) ────────────────────
	SurgeMultiplier float64
	SurgeLabel      string

	// ─── Promotion (Pricing V3 does not decide voucher eligibility — that
	// stays the Promotion Engine's job, BRB Part 3/4 — it only applies a
	// pre-resolved discount amount the caller supplies, exactly the same
	// division of responsibility backend/services/pricing/simulation
	// already uses for PromotionInput) ──────────────────────────────────────
	VoucherLabel     string
	VoucherRequested int64
	VoucherDiscount  int64 // clamped so it can never exceed the pre-discount total (BRB §4.9)

	// ─── Commission / Driver / Platform (new production responsibility —
	// see commission_v3.go) ─────────────────────────────────────────────────
	CommissionTier CommissionTier
	CommissionRate float64
	Commission     int64 // on RideFare only, never on BookingFee/pass-through (BRB §2.2.6)
	PlatformFee    int64 // = BookingFee, renamed per PRICING_V3_DESIGN.md Phần 3 ("Booking Fee đóng vai trò Platform Fee")
	VATRate        float64
	VAT            int64 // on (Commission+PlatformFee), ASSUMPTION per MARKET_PRICING_RESEARCH.md — no BRB VAT rule exists

	DriverIncome    int64 // RideFare - Commission (before any minimum-earning top-up — not modelled in V3 Pricing service; that guarantee is Settlement's responsibility per ECONOMY_ENGINE §4)
	PlatformRevenue int64 // Commission + PlatformFee - VAT - VoucherDiscount

	FinalFare int64 // what the rider actually pays: RideFare + PlatformFee - VoucherDiscount
}

// ExplanationLine is one row of a rider/driver-facing fare explanation —
// PRICING_V3_DESIGN.md Phần 10 / sprint brief Part 10 ("để frontend chỉ
// việc render").
type ExplanationLine struct {
	Label  string
	Amount int64 // negative for discounts
}

// Explanation returns an ordered, human-readable line-item breakdown of fb,
// skipping components that did not apply (0 airport fee, no surge, no
// voucher) so the rendered explanation never shows a confusing "0đ" line
// for something that simply does not apply to this trip.
func (fb FullFareBreakdownV3) Explanation() []ExplanationLine {
	lines := []ExplanationLine{
		{Label: "Base Fare", Amount: fb.BaseFare},
		{Label: "Distance", Amount: fb.DistanceFare},
	}
	if fb.TrafficTimeFare > 0 {
		lines = append(lines, ExplanationLine{Label: "Traffic", Amount: fb.TrafficTimeFare})
	}
	if fb.WaitingFee > 0 {
		lines = append(lines, ExplanationLine{Label: "Waiting", Amount: fb.WaitingFee})
	}
	if fb.AirportFee > 0 {
		label := "Airport"
		if fb.AirportLeg == AirportLegPickup {
			label = "Airport Pickup"
		} else if fb.AirportLeg == AirportLegDropoff {
			label = "Airport Dropoff"
		}
		lines = append(lines, ExplanationLine{Label: label, Amount: fb.AirportFee})
	}
	if fb.SurgeMultiplier > 1.0 {
		label := "Surge"
		if fb.SurgeLabel != "" {
			label = fb.SurgeLabel
		}
		surgeAmount := fb.RideFare - (fb.BaseFare + fb.DistanceFare + fb.TrafficTimeFare)
		if surgeAmount < 0 {
			surgeAmount = 0
		}
		lines = append(lines, ExplanationLine{Label: label, Amount: surgeAmount})
	}
	if fb.PlatformFee > 0 {
		lines = append(lines, ExplanationLine{Label: "Booking Fee", Amount: fb.PlatformFee})
	}
	if fb.VoucherDiscount > 0 {
		label := "Voucher"
		if fb.VoucherLabel != "" {
			label = fb.VoucherLabel
		}
		lines = append(lines, ExplanationLine{Label: label, Amount: -fb.VoucherDiscount})
	}
	lines = append(lines, ExplanationLine{Label: "Final", Amount: fb.FinalFare})
	return lines
}

// ExplanationString renders Explanation() as an aligned, fixed-width text
// block — the exact format the sprint brief's Part 10 example shows
// ("Base Fare........25.000"), so a frontend that has no rich breakdown UI
// yet can render this string directly.
func (fb FullFareBreakdownV3) ExplanationString() string {
	lines := fb.Explanation()
	var b strings.Builder
	const width = 16
	for _, l := range lines {
		label := l.Label
		if len(label) > width {
			label = label[:width]
		}
		dots := strings.Repeat(".", width-len(label)+1)
		sign := ""
		if l.Amount < 0 {
			sign = "-"
		}
		b.WriteString(fmt.Sprintf("%s%s%s%s\n", label, dots, sign, formatThousands(abs64(l.Amount))))
	}
	return b.String()
}

func abs64(v int64) int64 {
	if v < 0 {
		return -v
	}
	return v
}

// formatThousands renders v with "." as the thousands separator, matching
// the Vietnamese VND convention used throughout docs/business/*.md.
func formatThousands(v int64) string {
	s := fmt.Sprintf("%d", v)
	n := len(s)
	if n <= 3 {
		return s
	}
	var b strings.Builder
	rem := n % 3
	if rem > 0 {
		b.WriteString(s[:rem])
		if n > rem {
			b.WriteString(".")
		}
	}
	for i := rem; i < n; i += 3 {
		b.WriteString(s[i : i+3])
		if i+3 < n {
			b.WriteString(".")
		}
	}
	return b.String()
}
