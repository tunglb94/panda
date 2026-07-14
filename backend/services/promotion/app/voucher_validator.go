package app

import (
	"time"

	"github.com/fairride/promotion/domain/entity"
)

// VoucherValidator performs the generic, type-agnostic checks every voucher
// must pass regardless of PromotionType. This is exactly the set of checks
// the sprint brief lists under "ENGINE PHẢI TÍNH ĐƯỢC" (8 checks), plus
// MinOrder — required because Voucher.MinOrder is one of the 20 requested
// fields and BRB §4.8 defines its behaviour, even though it wasn't named
// among the 8 examples.
//
// VoucherValidator does NOT decide type-specific eligibility (e.g. "is this
// really the rider's birthday") — that is PromotionRule's job. VoucherValidator
// only checks facts intrinsic to the Voucher record itself plus the request's
// city/vehicle/membership/order-amount.
type VoucherValidator struct{}

func NewVoucherValidator() *VoucherValidator {
	return &VoucherValidator{}
}

// Validate returns nil if v is structurally valid for req at time now.
// Otherwise it returns a *errors.DomainError built via entity.PromotionError
// with a ReasonCode identifying exactly which check failed.
func (val *VoucherValidator) Validate(v *entity.Voucher, req *entity.PromotionRequest, now time.Time) error {
	if err := val.checkStatus(v); err != nil {
		return err
	}
	if err := val.checkTiming(v, now); err != nil {
		return err
	}
	if err := val.checkBudget(v); err != nil {
		return err
	}
	if err := val.checkUsage(v); err != nil {
		return err
	}
	if err := val.checkCity(v, req.City); err != nil {
		return err
	}
	if err := val.checkVehicleType(v, req.VehicleType); err != nil {
		return err
	}
	if err := val.checkServiceType(v, req.ServiceType); err != nil {
		return err
	}
	if err := val.checkTripType(v, req.TripType); err != nil {
		return err
	}
	if err := val.checkMembership(v, req.MembershipTier); err != nil {
		return err
	}
	if err := val.checkMinOrder(v, req.OrderAmount); err != nil {
		return err
	}
	return nil
}

// checkStatus: "voucher hợp lệ" — must be in a redeemable status.
func (val *VoucherValidator) checkStatus(v *entity.Voucher) error {
	if v.Status != entity.VoucherStatusActive {
		return entity.PromotionError(entity.ReasonInvalidStatus,
			"voucher is not active, current status: "+string(v.Status))
	}
	return nil
}

// checkTiming: "voucher hết hạn" + "voucher sai thời gian" — outside the
// campaign's [start_time, end_time] window. BRB §4.5: cannot redeem after
// 23:59:59 on the expiry date.
func (val *VoucherValidator) checkTiming(v *entity.Voucher, now time.Time) error {
	if now.Before(v.StartTime) {
		return entity.PromotionError(entity.ReasonWrongTiming, "voucher is not yet active")
	}
	if now.After(v.EndTime) {
		return entity.PromotionError(entity.ReasonExpired, "voucher has expired")
	}
	return nil
}

// checkBudget: "voucher hết ngân sách" — BRB §3.3 Rule 3: at 100% consumed,
// campaign pauses automatically.
func (val *VoucherValidator) checkBudget(v *entity.Voucher) error {
	if v.RemainingBudget <= 0 || v.Status == entity.VoucherStatusExhausted {
		return entity.PromotionError(entity.ReasonBudgetExhausted, "voucher campaign budget exhausted")
	}
	return nil
}

// checkUsage: "voucher hết lượt" — BRB §4.6 total issuance quota. Per-rider
// usage is checked separately by PromotionService (it needs a repository
// lookup this validator does not have access to).
func (val *VoucherValidator) checkUsage(v *entity.Voucher) error {
	if !v.IsUsageAvailable() {
		return entity.PromotionError(entity.ReasonUsageExhausted, "voucher total usage quota exhausted")
	}
	return nil
}

// checkCity: "voucher sai thành phố" — BRB §4.10, empty list = nationwide.
func (val *VoucherValidator) checkCity(v *entity.Voucher, city string) error {
	if len(v.Cities) == 0 {
		return nil
	}
	if !containsCI(v.Cities, city) {
		return entity.PromotionError(entity.ReasonWrongCity, "voucher is not valid in city: "+city)
	}
	return nil
}

// checkVehicleType: "voucher sai loại xe" — BRB §4.11, empty list = all classes.
func (val *VoucherValidator) checkVehicleType(v *entity.Voucher, vehicleType string) error {
	if len(v.VehicleTypes) == 0 {
		return nil
	}
	if !containsCI(v.VehicleTypes, vehicleType) {
		return entity.PromotionError(entity.ReasonWrongVehicle, "voucher is not valid for vehicle type: "+vehicleType)
	}
	return nil
}

// checkServiceType: voucher scoped to specific Rider-app tiers (bike/bike_plus/
// car/car_xl) — empty list on the voucher = all tiers. See Voucher.ServiceTypes.
func (val *VoucherValidator) checkServiceType(v *entity.Voucher, serviceType string) error {
	if len(v.ServiceTypes) == 0 {
		return nil
	}
	if !containsCI(v.ServiceTypes, serviceType) {
		return entity.PromotionError(entity.ReasonWrongServiceType, "voucher is not valid for service type: "+serviceType)
	}
	return nil
}

// checkTripType: voucher scoped to "ride" and/or "delivery" — empty list on
// the voucher = both. See Voucher.TripTypes.
func (val *VoucherValidator) checkTripType(v *entity.Voucher, tripType string) error {
	if len(v.TripTypes) == 0 {
		return nil
	}
	if !containsCI(v.TripTypes, tripType) {
		return entity.PromotionError(entity.ReasonWrongTripType, "voucher is not valid for trip type: "+tripType)
	}
	return nil
}

// checkMembership: "voucher sai membership" — eligibility gate only, per
// ECONOMY_ENGINE §8.1 (membership never changes the fare formula, so this
// check never produces or scales a discount — it only allows/denies).
// Empty list = no membership restriction.
func (val *VoucherValidator) checkMembership(v *entity.Voucher, membershipTier string) error {
	if len(v.Membership) == 0 {
		return nil
	}
	if !containsCI(v.Membership, membershipTier) {
		return entity.PromotionError(entity.ReasonWrongMembership, "voucher requires membership tier: "+joinOr(v.Membership))
	}
	return nil
}

// checkMinOrder: BRB §4.8 — voucher defines a minimum trip fare below which
// it cannot be applied.
func (val *VoucherValidator) checkMinOrder(v *entity.Voucher, orderAmount int64) error {
	if v.MinOrder > 0 && orderAmount < v.MinOrder {
		return entity.PromotionError(entity.ReasonMinOrderNotMet, "order amount below voucher minimum fare")
	}
	return nil
}
