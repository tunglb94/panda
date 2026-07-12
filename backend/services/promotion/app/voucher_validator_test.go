package app_test

import (
	"testing"
	"time"

	"github.com/fairride/promotion/app"
	"github.com/fairride/promotion/domain/entity"
)

var now = time.Date(2026, 7, 10, 12, 0, 0, 0, time.UTC)

func activeVoucher(t *testing.T, mutate func(v *entity.Voucher)) *entity.Voucher {
	t.Helper()
	v, err := entity.NewVoucher(
		"v-1", "", "Test Voucher", "fixture", 30,
		now.Add(-time.Hour), now.Add(time.Hour),
		0, 1, 100_000,
		entity.DiscountTypePercentage, 50, 30_000, 20_000,
		nil, nil, nil,
		true, false, false,
		entity.PromotionTypeFirstRide, now,
	)
	if err != nil {
		t.Fatalf("build voucher: %v", err)
	}
	if err := v.Activate(now); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if mutate != nil {
		mutate(v)
	}
	return v
}

func baseRequest() *entity.PromotionRequest {
	return &entity.PromotionRequest{
		RiderID:     "rider-1",
		VehicleType: "car",
		City:        "Ho Chi Minh City",
		OrderAmount: 100_000,
		RequestTime: now,
	}
}

func TestVoucherValidator_Valid(t *testing.T) {
	v := activeVoucher(t, nil)
	req := baseRequest()
	if err := app.NewVoucherValidator().Validate(v, req, now); err != nil {
		t.Fatalf("expected valid, got %v", err)
	}
}

func TestVoucherValidator_InvalidStatus(t *testing.T) {
	v := activeVoucher(t, func(v *entity.Voucher) { v.Status = entity.VoucherStatusPaused })
	err := app.NewVoucherValidator().Validate(v, baseRequest(), now)
	if entity.ReasonOf(err) != entity.ReasonInvalidStatus {
		t.Fatalf("expected ReasonInvalidStatus, got %v (%v)", entity.ReasonOf(err), err)
	}
}

func TestVoucherValidator_Expired(t *testing.T) {
	v := activeVoucher(t, func(v *entity.Voucher) { v.EndTime = now.Add(-time.Minute) })
	err := app.NewVoucherValidator().Validate(v, baseRequest(), now)
	if entity.ReasonOf(err) != entity.ReasonExpired {
		t.Fatalf("expected ReasonExpired, got %v (%v)", entity.ReasonOf(err), err)
	}
}

func TestVoucherValidator_NotYetActive(t *testing.T) {
	v := activeVoucher(t, func(v *entity.Voucher) { v.StartTime = now.Add(time.Hour) })
	err := app.NewVoucherValidator().Validate(v, baseRequest(), now)
	if entity.ReasonOf(err) != entity.ReasonWrongTiming {
		t.Fatalf("expected ReasonWrongTiming, got %v (%v)", entity.ReasonOf(err), err)
	}
}

func TestVoucherValidator_BudgetExhausted(t *testing.T) {
	v := activeVoucher(t, func(v *entity.Voucher) { v.RemainingBudget = 0 })
	err := app.NewVoucherValidator().Validate(v, baseRequest(), now)
	if entity.ReasonOf(err) != entity.ReasonBudgetExhausted {
		t.Fatalf("expected ReasonBudgetExhausted, got %v (%v)", entity.ReasonOf(err), err)
	}
}

func TestVoucherValidator_UsageExhausted(t *testing.T) {
	v := activeVoucher(t, func(v *entity.Voucher) {
		v.MaxUsage = 5
		v.UsageCount = 5
	})
	err := app.NewVoucherValidator().Validate(v, baseRequest(), now)
	if entity.ReasonOf(err) != entity.ReasonUsageExhausted {
		t.Fatalf("expected ReasonUsageExhausted, got %v (%v)", entity.ReasonOf(err), err)
	}
}

func TestVoucherValidator_WrongCity(t *testing.T) {
	v := activeVoucher(t, func(v *entity.Voucher) { v.Cities = []string{"Hanoi"} })
	err := app.NewVoucherValidator().Validate(v, baseRequest(), now) // request city = Ho Chi Minh City
	if entity.ReasonOf(err) != entity.ReasonWrongCity {
		t.Fatalf("expected ReasonWrongCity, got %v (%v)", entity.ReasonOf(err), err)
	}
}

func TestVoucherValidator_WrongVehicleType(t *testing.T) {
	v := activeVoucher(t, func(v *entity.Voucher) { v.VehicleTypes = []string{"van"} })
	err := app.NewVoucherValidator().Validate(v, baseRequest(), now) // request vehicle = car
	if entity.ReasonOf(err) != entity.ReasonWrongVehicle {
		t.Fatalf("expected ReasonWrongVehicle, got %v (%v)", entity.ReasonOf(err), err)
	}
}

func TestVoucherValidator_WrongMembership(t *testing.T) {
	v := activeVoucher(t, func(v *entity.Voucher) { v.Membership = []string{"gold", "diamond"} })
	req := baseRequest()
	req.MembershipTier = "silver"
	err := app.NewVoucherValidator().Validate(v, req, now)
	if entity.ReasonOf(err) != entity.ReasonWrongMembership {
		t.Fatalf("expected ReasonWrongMembership, got %v (%v)", entity.ReasonOf(err), err)
	}
}

func TestVoucherValidator_MinOrderNotMet(t *testing.T) {
	v := activeVoucher(t, nil) // MinOrder = 20,000 from fixture
	req := baseRequest()
	req.OrderAmount = 10_000
	err := app.NewVoucherValidator().Validate(v, req, now)
	if entity.ReasonOf(err) != entity.ReasonMinOrderNotMet {
		t.Fatalf("expected ReasonMinOrderNotMet, got %v (%v)", entity.ReasonOf(err), err)
	}
}

// TestVoucherValidator_VehicleCatalogExpansionTypes locks in Part 9's own
// example ("Voucher chỉ áp dụng Bike Plus hoặc Car XL"): voucher_repository/
// voucher_validator never had a fixed vehicle-type enum (VehicleTypes is a
// free-form []string, matched case-insensitively — see
// voucher_validator.go's checkVehicleType), so the 4 new tier names work
// with zero code change. This test proves it rather than just asserting it.
func TestVoucherValidator_VehicleCatalogExpansionTypes(t *testing.T) {
	v := activeVoucher(t, func(v *entity.Voucher) { v.VehicleTypes = []string{"bike_plus", "car_xl"} })

	req := baseRequest()
	req.VehicleType = "bike_plus"
	if err := app.NewVoucherValidator().Validate(v, req, now); err != nil {
		t.Fatalf("expected bike_plus to pass a [bike_plus, car_xl] restriction, got %v", err)
	}

	req.VehicleType = "car_xl"
	if err := app.NewVoucherValidator().Validate(v, req, now); err != nil {
		t.Fatalf("expected car_xl to pass a [bike_plus, car_xl] restriction, got %v", err)
	}

	req.VehicleType = "car" // not in the restriction list
	err := app.NewVoucherValidator().Validate(v, req, now)
	if entity.ReasonOf(err) != entity.ReasonWrongVehicle {
		t.Fatalf("expected ReasonWrongVehicle for car against a [bike_plus, car_xl] restriction, got %v (%v)", entity.ReasonOf(err), err)
	}
}

func TestVoucherValidator_NationwideAndAllVehicleDefaults(t *testing.T) {
	// empty Cities/VehicleTypes/Membership means no restriction (BRB §4.10, §4.11)
	v := activeVoucher(t, nil)
	req := baseRequest()
	req.City = "Da Nang"
	req.VehicleType = "motorcycle"
	req.MembershipTier = "" // no membership
	if err := app.NewVoucherValidator().Validate(v, req, now); err != nil {
		t.Fatalf("expected no restriction to pass, got %v", err)
	}
}
