package entity

import (
	"strings"
	"time"

	"github.com/fairride/shared/errors"
)

// VoucherStatus is the campaign-level lifecycle state of a Voucher.
//
// This is distinct from BRB §4.2's per-instance lifecycle (Created → Issued →
// Active → Redeemed → Expired/Cancelled), which describes one voucher copy
// issued to one rider. The field set requested for this engine (budget,
// remaining_budget, max_usage, max_usage_per_user) describes a CAMPAIGN/
// definition that many riders redeem against (BRB §3.2.9 Coupon Campaign +
// BRB Part 4 Voucher merged into one generic model, since the requested
// fields span both). VoucherStatus tracks the campaign, not any one rider's
// copy of it.
type VoucherStatus string

const (
	VoucherStatusDraft     VoucherStatus = "draft"
	VoucherStatusActive    VoucherStatus = "active"
	VoucherStatusPaused    VoucherStatus = "paused"
	VoucherStatusExpired   VoucherStatus = "expired"
	VoucherStatusExhausted VoucherStatus = "exhausted" // BRB §3.3 Rule 3: 100% budget consumed -> pauses automatically
	VoucherStatusCancelled VoucherStatus = "cancelled"
)

func (s VoucherStatus) Valid() bool {
	switch s {
	case VoucherStatusDraft, VoucherStatusActive, VoucherStatusPaused,
		VoucherStatusExpired, VoucherStatusExhausted, VoucherStatusCancelled:
		return true
	}
	return false
}

// Voucher is a promotion campaign definition. Field list is exactly the 20
// fields requested by the sprint brief.
type Voucher struct {
	ID          string
	Code        string // empty = system auto-applies based on Type eligibility (no code entry); non-empty = rider must enter this code (BRB §3.2.9 Coupon / Part 4 Voucher)
	Name        string
	Description string
	Status      VoucherStatus
	Priority    int // lower = evaluated first / wins ties, per BRB §3.4 Campaign Priority Rules

	StartTime time.Time
	EndTime   time.Time

	MaxUsage         int64 // 0 = uncapped by count (still budget-capped). BRB §4.6 total issuance quota.
	MaxUsagePerUser  int64 // BRB §4.6 "most vouchers have a usage limit of 1"; 0 is treated as 1 (single-use) by the validator, never as unlimited.

	Budget          int64 // VND, total platform cost approved for this campaign. BRB §3.3 Rule 1: "No budget = no campaign."
	RemainingBudget int64 // VND, decremented on each Redeem.

	DiscountType  DiscountType
	DiscountValue int64 // percentage (0-100) if DiscountType=percentage, VND amount if flat
	MaxDiscount   int64 // VND cap, 0 = no cap. BRB §3.2.x "maximum X VND" fields.
	MinOrder      int64 // VND, minimum trip fare required. BRB §4.8.

	VehicleTypes []string // empty = all vehicle types. BRB §4.11.
	Cities       []string // empty = nationwide. BRB §4.10.
	Membership   []string // empty = no membership restriction. Eligibility filter only — see PromotionTypeMembership doc comment. Never used to compute a discount amount.

	NewUserOnly bool // BRB §3.2.1 First Ride: "zero completed trips"

	// Combinable: may this voucher coexist with a post-trip mechanism (BRB §4.7
	// exception / §3.4 #7: Cashback always applies post-trip regardless of the
	// upfront discount used). Does NOT allow combining with another Voucher.
	Combinable bool

	// Stackable: may this voucher be combined with ANOTHER voucher/promotion on
	// the same trip. BRB §4.7 default is a hard "maximum ONE voucher per trip";
	// Stackable=true marks an explicit CPO-approved exception (BRB §3.4
	// "Override" clause). PromotionService's current Evaluate() always applies
	// at most one voucher regardless of this flag — true multi-voucher stacking
	// requires a campaign-pair configuration model not covered by this field set.
	// TODO(promotion-engine): implement pair-level stacking overrides once BRB
	// defines how approved pairs are recorded.
	Stackable bool

	Type PromotionType

	UsageCount int64 // total redemptions so far, across all riders

	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewVoucher validates and constructs a brand-new campaign definition
// (Status starts Draft; RemainingBudget starts equal to Budget).
func NewVoucher(
	id, code, name, description string,
	priority int,
	startTime, endTime time.Time,
	maxUsage, maxUsagePerUser, budget int64,
	discountType DiscountType, discountValue, maxDiscount, minOrder int64,
	vehicleTypes, cities, membership []string,
	newUserOnly, combinable, stackable bool,
	promoType PromotionType,
	now time.Time,
) (*Voucher, error) {
	if id == "" {
		return nil, errors.InvalidArgument("voucher id is required")
	}
	if name == "" {
		return nil, errors.InvalidArgument("voucher name is required")
	}
	if !discountType.Valid() {
		return nil, errors.InvalidArgument("invalid discount_type: " + string(discountType))
	}
	if discountValue < 0 {
		return nil, errors.InvalidArgument("discount_value cannot be negative")
	}
	if discountType == DiscountTypePercentage && discountValue > 100 {
		return nil, errors.InvalidArgument("percentage discount_value cannot exceed 100")
	}
	if maxDiscount < 0 || minOrder < 0 {
		return nil, errors.InvalidArgument("max_discount and min_order cannot be negative")
	}
	if maxUsage < 0 || maxUsagePerUser < 0 {
		return nil, errors.InvalidArgument("max_usage and max_usage_per_user cannot be negative")
	}
	if budget <= 0 {
		// BRB §3.3 Rule 1: "Every promotion campaign must have an approved
		// budget before activation. No budget = no campaign."
		return nil, errors.InvalidArgument("budget must be > 0 (BRB §3.3 Rule 1: no budget = no campaign)")
	}
	if !endTime.After(startTime) {
		return nil, errors.InvalidArgument("end_time must be after start_time")
	}
	if endTime.IsZero() {
		// BRB §3.6: "Campaigns created without an end date cannot be activated."
		return nil, errors.InvalidArgument("end_time is required (BRB §3.6: no open-ended campaigns)")
	}
	if promoType == "" {
		return nil, errors.InvalidArgument("promotion type is required")
	}

	v := &Voucher{
		ID:              id,
		Code:            strings.TrimSpace(code),
		Name:            name,
		Description:     description,
		Status:          VoucherStatusDraft,
		Priority:        priority,
		StartTime:       startTime,
		EndTime:         endTime,
		MaxUsage:        maxUsage,
		MaxUsagePerUser: maxUsagePerUser,
		Budget:          budget,
		RemainingBudget: budget,
		DiscountType:    discountType,
		DiscountValue:   discountValue,
		MaxDiscount:     maxDiscount,
		MinOrder:        minOrder,
		VehicleTypes:    vehicleTypes,
		Cities:          cities,
		Membership:      membership,
		NewUserOnly:     newUserOnly,
		Combinable:      combinable,
		Stackable:       stackable,
		Type:            promoType,
		UsageCount:      0,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	return v, nil
}

// ReconstituteVoucher builds a Voucher from persisted data without
// re-validating business invariants (used when hydrating from Postgres).
func ReconstituteVoucher(
	id, code, name, description string,
	status VoucherStatus,
	priority int,
	startTime, endTime time.Time,
	maxUsage, maxUsagePerUser, budget, remainingBudget int64,
	discountType DiscountType, discountValue, maxDiscount, minOrder int64,
	vehicleTypes, cities, membership []string,
	newUserOnly, combinable, stackable bool,
	promoType PromotionType,
	usageCount int64,
	createdAt, updatedAt time.Time,
) *Voucher {
	return &Voucher{
		ID:              id,
		Code:            code,
		Name:            name,
		Description:     description,
		Status:          status,
		Priority:        priority,
		StartTime:       startTime,
		EndTime:         endTime,
		MaxUsage:        maxUsage,
		MaxUsagePerUser: maxUsagePerUser,
		Budget:          budget,
		RemainingBudget: remainingBudget,
		DiscountType:    discountType,
		DiscountValue:   discountValue,
		MaxDiscount:     maxDiscount,
		MinOrder:        minOrder,
		VehicleTypes:    vehicleTypes,
		Cities:          cities,
		Membership:      membership,
		NewUserOnly:     newUserOnly,
		Combinable:      combinable,
		Stackable:       stackable,
		Type:            promoType,
		UsageCount:      usageCount,
		CreatedAt:       createdAt,
		UpdatedAt:       updatedAt,
	}
}

// Activate transitions Draft -> Active.
func (v *Voucher) Activate(now time.Time) *errors.DomainError {
	if v.Status != VoucherStatusDraft && v.Status != VoucherStatusPaused {
		return errors.PreconditionFailed("only draft or paused vouchers can be activated, current status: " + string(v.Status))
	}
	v.Status = VoucherStatusActive
	v.UpdatedAt = now
	return nil
}

// Pause transitions Active -> Paused (manual operator action, distinct from
// automatic Exhaust).
func (v *Voucher) Pause(now time.Time) *errors.DomainError {
	if v.Status != VoucherStatusActive {
		return errors.PreconditionFailed("only active vouchers can be paused, current status: " + string(v.Status))
	}
	v.Status = VoucherStatusPaused
	v.UpdatedAt = now
	return nil
}

// Cancel transitions any non-terminal state -> Cancelled (e.g. fraud detected,
// BRB §4.2).
func (v *Voucher) Cancel(now time.Time) *errors.DomainError {
	if v.Status == VoucherStatusCancelled {
		return errors.PreconditionFailed("voucher already cancelled")
	}
	v.Status = VoucherStatusCancelled
	v.UpdatedAt = now
	return nil
}

// ExhaustIfBudgetSpent transitions Active -> Exhausted once RemainingBudget
// reaches 0, per BRB §3.3 Rule 3 ("At 100%, the campaign pauses automatically").
func (v *Voucher) ExhaustIfBudgetSpent(now time.Time) {
	if v.Status == VoucherStatusActive && v.RemainingBudget <= 0 {
		v.Status = VoucherStatusExhausted
		v.UpdatedAt = now
	}
}

// HasBudgetFor reports whether the voucher's remaining budget can cover a
// discount of the given amount.
func (v *Voucher) HasBudgetFor(discountAmount int64) bool {
	return v.RemainingBudget >= discountAmount
}

// IsUsageAvailable reports whether the total usage cap still has room.
// MaxUsage == 0 means uncapped by count (still budget-capped separately).
func (v *Voucher) IsUsageAvailable() bool {
	if v.MaxUsage == 0 {
		return true
	}
	return v.UsageCount < v.MaxUsage
}

// PerUserLimit returns the effective per-user usage cap. BRB §4.6: "Most
// vouchers have a usage limit of 1 (single-use)" — 0 is treated as 1, never
// as unlimited, so an operator cannot accidentally create an unlimited-use
// voucher by leaving the field at its zero value.
func (v *Voucher) PerUserLimit() int64 {
	if v.MaxUsagePerUser <= 0 {
		return 1
	}
	return v.MaxUsagePerUser
}

// Reserve applies a redemption: decrements RemainingBudget, increments
// UsageCount, and auto-exhausts if budget hits 0. Callers must have already
// validated eligibility (VoucherValidator + PromotionRule) before calling this.
func (v *Voucher) Reserve(discountAmount int64, now time.Time) *errors.DomainError {
	if discountAmount < 0 {
		return errors.InvalidArgument("discount amount cannot be negative")
	}
	if !v.HasBudgetFor(discountAmount) {
		return errors.ResourceExhausted("voucher budget exhausted")
	}
	if !v.IsUsageAvailable() {
		return errors.ResourceExhausted("voucher usage quota exhausted")
	}
	v.RemainingBudget -= discountAmount
	v.UsageCount++
	v.UpdatedAt = now
	v.ExhaustIfBudgetSpent(now)
	return nil
}

// Release reverses a Reserve (BRB §4.13 Refund Behaviour / §4.14 Cancellation
// Behaviour: voucher is reinstated when the trip did not consume it through
// rider fault).
func (v *Voucher) Release(discountAmount int64, now time.Time) *errors.DomainError {
	if discountAmount < 0 {
		return errors.InvalidArgument("discount amount cannot be negative")
	}
	v.RemainingBudget += discountAmount
	if v.RemainingBudget > v.Budget {
		v.RemainingBudget = v.Budget
	}
	if v.UsageCount > 0 {
		v.UsageCount--
	}
	if v.Status == VoucherStatusExhausted && v.RemainingBudget > 0 {
		v.Status = VoucherStatusActive
	}
	v.UpdatedAt = now
	return nil
}
