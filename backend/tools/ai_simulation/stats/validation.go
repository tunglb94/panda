package stats

import (
	"fmt"

	"github.com/fairride/ai_simulation/domain/entity"
)

// ValidationWarning is one anomaly Validate found — never fatal. PHẦN 11
// asks for "ghi warning, không crash", so this package never returns an
// error; a Validate call always produces a (possibly empty) warning list.
type ValidationWarning struct {
	Check    string `json:"check"`
	Severity string `json:"severity"` // "info" | "warning" | "critical"
	Message  string `json:"message"`
}

type ValidationReport struct {
	Passed   bool                `json:"passed"` // true iff no "critical" warnings
	Warnings []ValidationWarning `json:"warnings"`
}

// Validate runs every PHẦN 11 self-check (Revenue Balance, Driver Income,
// Commission, Voucher, Promotion, Profit) against the final trip ledger and
// driver population. Every check is a plain comparison against values this
// same run already computed — no external ground truth is assumed, since
// none exists; the point is catching internally inconsistent output (e.g. a
// future code change that breaks conservation of money), not validating
// against a business target.
func (c *Collector) Validate(trips []*entity.SimTrip, drivers map[string]*entity.DriverAgent, bi BusinessIntelligence) ValidationReport {
	var report ValidationReport
	warn := func(check, severity, format string, args ...any) {
		report.Warnings = append(report.Warnings, ValidationWarning{Check: check, Severity: severity, Message: fmt.Sprintf(format, args...)})
	}

	// --- Revenue Balance ---
	// Money conservation: what riders paid (GMV, pre-discount gross) must
	// equal driver net + platform commission + the discounts riders didn't
	// pay. A mismatch beyond float-rounding noise means some money is being
	// created or destroyed somewhere in the pipeline.
	accountedVND := bi.DriverRevenueVND + bi.PlatformRevenueVND + bi.VoucherCostVND + bi.PromotionCostVND
	if bi.GMVVND > 0 {
		driftPercent := 100 * absFloat(float64(bi.GMVVND-accountedVND)) / float64(bi.GMVVND)
		if driftPercent > 1.0 {
			warn("revenue_balance", "critical", "GMV (%d) khác biệt %.2f%% so với tổng đã phân bổ (driver+platform+voucher+promotion = %d) — vượt ngưỡng sai số 1%%, có khả năng lỗi tính toán.", bi.GMVVND, driftPercent, accountedVND)
		} else if driftPercent > 0.1 {
			warn("revenue_balance", "info", "GMV lệch %.3f%% so với tổng đã phân bổ — trong ngưỡng sai số làm tròn, không đáng lo ngại.", driftPercent)
		}
	}
	if bi.PlatformRevenueVND < 0 {
		warn("revenue_balance", "critical", "Platform revenue âm (%d VND) — không hợp lệ.", bi.PlatformRevenueVND)
	}
	if bi.DriverRevenueVND < 0 {
		warn("revenue_balance", "critical", "Driver revenue âm (%d VND) — không hợp lệ.", bi.DriverRevenueVND)
	}

	// --- Driver Income ---
	var negativeIncomeDrivers int
	for _, d := range drivers {
		if d.IncomeWeek < 0 || d.IncomeToday < 0 {
			negativeIncomeDrivers++
		}
	}
	if negativeIncomeDrivers > 0 {
		warn("driver_income", "critical", "%d tài xế có thu nhập âm (IncomeToday/IncomeWeek < 0) — không hợp lệ, tài xế không thể có thu nhập ròng âm.", negativeIncomeDrivers)
	}

	// --- Commission ---
	// A tight BRB §7.1 commission/base-fare band (12%-20%) is NOT a valid
	// per-trip check here: commission is deliberately scaled by surge and
	// promotion adjustments (see ride_flow.go's "Apply surge/promotion
	// adjustment proportionally" comment), so a surged trip's
	// commission/BaseFareVND ratio legitimately exceeds the base rate —
	// comparing against the unsurged BaseFareVND denominator produced a
	// false-positive on ~70% of trips in manual testing once surge was
	// exercised. The real, surge-independent invariant is simpler:
	// commission and driver net must both be non-negative, and together
	// must equal FinalFareVND — what the rider actually paid after any
	// voucher/promotion discount (commission+driverNet is a split of money
	// actually collected, not of the pre-discount sticker price; the
	// discount itself is foregone revenue, not part of either driver or
	// platform's share). A >1 VND mismatch beyond integer-division rounding
	// flags a real accounting bug in the split logic.
	var negativeCommission, negativeDriverNet, commissionExceedsFare int
	for _, t := range trips {
		if t.Outcome != entity.OutcomeCompleted {
			continue
		}
		if t.CommissionVND < 0 {
			negativeCommission++
		}
		if t.DriverNetVND < 0 {
			negativeDriverNet++
		}
		if diff := t.CommissionVND + t.DriverNetVND - t.FinalFareVND; diff > 2 || diff < -2 {
			commissionExceedsFare++
		}
	}
	if negativeCommission > 0 {
		warn("commission", "critical", "%d chuyến có commission âm — không hợp lệ.", negativeCommission)
	}
	if negativeDriverNet > 0 {
		warn("commission", "critical", "%d chuyến có driver net âm — không hợp lệ.", negativeDriverNet)
	}
	if commissionExceedsFare > 0 {
		warn("commission", "warning", "%d chuyến có commission+driver_net lệch quá 2 VND so với giá cuối cùng khách trả (FinalFareVND) — kiểm tra lại phần chia surge/booking fee.", commissionExceedsFare)
	}

	// --- Voucher ---
	// FinalFareVND is the post-discount amount the rider actually pays — it
	// must never go negative, i.e. a voucher discount must never exceed the
	// pre-discount order amount (FinalFareVND + VoucherDiscountVND, the
	// same convention BuildBusinessIntelligence's GMV calc uses).
	var voucherExceedsOrder int
	for _, t := range trips {
		if t.PromotionType != "manual_coupon" || t.VoucherDiscountVND <= 0 {
			continue
		}
		if t.FinalFareVND < 0 {
			voucherExceedsOrder++
		}
	}
	if voucherExceedsOrder > 0 {
		warn("voucher", "critical", "%d chuyến có giá sau khi trừ voucher bị âm (voucher discount vượt quá giá trị đơn hàng) — không hợp lệ.", voucherExceedsOrder)
	}

	// --- Promotion ---
	if bi.VoucherCostVND+bi.PromotionCostVND > bi.GMVVND && bi.GMVVND > 0 {
		warn("promotion", "critical", "Tổng chi phí voucher+promotion (%d VND) vượt quá GMV (%d VND) — không hợp lệ.", bi.VoucherCostVND+bi.PromotionCostVND, bi.GMVVND)
	}

	// --- Profit ---
	if bi.GMVVND > 0 {
		marginOfProfit := 100 * float64(bi.ProfitVND) / float64(bi.GMVVND)
		if marginOfProfit < -20 {
			warn("profit", "warning", "Estimated Profit âm sâu (%.1f%% GMV) — nếu duy trì ở quy mô lớn, mô hình kinh doanh hiện tại không bền vững với giả định chi phí hạ tầng/VAT đang dùng (xem unit_economics.json).", marginOfProfit)
		}
		if bi.PlatformMarginPercent > 50 {
			warn("profit", "warning", "Platform Margin bất thường cao (%.1f%% GMV) — vượt xa dải hoa hồng BRB §7.1 (12%%-20%%), có khả năng lỗi tính toán thay vì hiệu quả kinh doanh thật.", bi.PlatformMarginPercent)
		}
	}

	report.Passed = true
	for _, w := range report.Warnings {
		if w.Severity == "critical" {
			report.Passed = false
			break
		}
	}
	return report
}

func absFloat(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}
