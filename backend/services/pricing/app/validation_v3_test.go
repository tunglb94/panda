package app_test

import (
	"testing"

	"github.com/fairride/pricing/app"
	"github.com/fairride/pricing/domain/entity"
)

func TestValidateFullBreakdown_NegativeFinalFare(t *testing.T) {
	err := app.ValidateFullBreakdown(&entity.FullFareBreakdownV3{FinalFare: -1})
	if err == nil {
		t.Fatal("expected an error for a negative FinalFare")
	}
}

func TestValidateFullBreakdown_NegativeRideFare(t *testing.T) {
	err := app.ValidateFullBreakdown(&entity.FullFareBreakdownV3{RideFare: -1})
	if err == nil {
		t.Fatal("expected an error for a negative RideFare")
	}
}

func TestValidateFullBreakdown_CommissionRateOver100Percent(t *testing.T) {
	err := app.ValidateFullBreakdown(&entity.FullFareBreakdownV3{CommissionRate: 1.5})
	if err == nil {
		t.Fatal("expected an error for a commission rate > 100%")
	}
}

func TestValidateFullBreakdown_CommissionRateNegative(t *testing.T) {
	err := app.ValidateFullBreakdown(&entity.FullFareBreakdownV3{CommissionRate: -0.1})
	if err == nil {
		t.Fatal("expected an error for a negative commission rate")
	}
}

func TestValidateFullBreakdown_VoucherExceedsFare(t *testing.T) {
	err := app.ValidateFullBreakdown(&entity.FullFareBreakdownV3{
		RideFare: 10000, WaitingFee: 0, PlatformFee: 2000, VoucherDiscount: 999999,
	})
	if err == nil {
		t.Fatal("expected an error: voucher discount exceeds the pre-discount total")
	}
}

func TestValidateFullBreakdown_VoucherExactlyEqualToTotalIsFine(t *testing.T) {
	err := app.ValidateFullBreakdown(&entity.FullFareBreakdownV3{
		RideFare: 10000, WaitingFee: 0, PlatformFee: 2000, VoucherDiscount: 12000,
		CommissionRate: 0.16,
	})
	if err != nil {
		t.Errorf("unexpected error for a voucher exactly matching the pre-discount total: %v", err)
	}
}

func TestValidateFullBreakdown_NegativeCommission(t *testing.T) {
	err := app.ValidateFullBreakdown(&entity.FullFareBreakdownV3{Commission: -1})
	if err == nil {
		t.Fatal("expected an error for negative commission")
	}
}

func TestValidateFullBreakdown_NegativeVAT(t *testing.T) {
	err := app.ValidateFullBreakdown(&entity.FullFareBreakdownV3{VAT: -1})
	if err == nil {
		t.Fatal("expected an error for negative VAT")
	}
}

func TestValidateFullBreakdown_NegativeDriverIncome(t *testing.T) {
	err := app.ValidateFullBreakdown(&entity.FullFareBreakdownV3{DriverIncome: -1})
	if err == nil {
		t.Fatal("expected an error for negative driver income")
	}
}

func TestValidateFullBreakdown_OverflowGuard(t *testing.T) {
	err := app.ValidateFullBreakdown(&entity.FullFareBreakdownV3{
		FinalFare: 2_000_000_000_000_000, RideFare: 2_000_000_000_000_000,
	})
	if err == nil {
		t.Fatal("expected an error for a fare exceeding the sane upper bound")
	}
}

func TestValidateFullBreakdown_ValidBreakdownPasses(t *testing.T) {
	err := app.ValidateFullBreakdown(&entity.FullFareBreakdownV3{
		RideFare: 50000, WaitingFee: 1000, PlatformFee: 3000, VoucherDiscount: 5000,
		FinalFare: 49000, Commission: 8000, VAT: 500, DriverIncome: 42000,
		CommissionRate: 0.16,
	})
	if err != nil {
		t.Errorf("unexpected error for a well-formed breakdown: %v", err)
	}
}
