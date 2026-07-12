package app

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/fairride/promotion/domain/entity"
	"github.com/fairride/promotion/domain/repository"
	sharederrors "github.com/fairride/shared/errors"
)

// PromotionService is the engine's single entry point. Pricing Service is
// meant to call Evaluate() to ask "how much discount applies to this trip,"
// exactly matching the sprint goal: "Pricing Service chỉ hỏi Promotion
// Engine được giảm bao nhiêu."
//
// Evaluate is read-only and safe to call repeatedly for fare estimates (it
// does not touch budget/usage counters). Once a trip is actually booked,
// callers must call Redeem to commit the reservation (decrement budget,
// increment usage) — mirroring the Estimate-vs-Commit split already used by
// the Pricing Service and BRB §5.10's Locked Balance pattern, so that a rider
// viewing multiple fare estimates does not consume voucher budget/usage more
// than once.
//
// Scope note: BRB §3.2.8 Cashback is a POST-trip credit (24h delay, BRB
// §3.4 #7 "does not conflict with upfront discounts"). It is not part of
// this pre-trip Evaluate() contract and is intentionally out of scope —
// TODO(promotion-engine): a separate PostTripPromotionService should own
// Cashback once the settlement/wallet services exist to credit it.
type PromotionService struct {
	repo      repository.PromotionRepository
	validator *VoucherValidator
	rules     RuleRegistry
}

func NewPromotionService(repo repository.PromotionRepository, validator *VoucherValidator, rules RuleRegistry) *PromotionService {
	return &PromotionService{repo: repo, validator: validator, rules: rules}
}

// candidateOutcome pairs a voucher with either its computed discount or the
// reason it was rejected, used to build PromotionResult.Warnings.
type candidateOutcome struct {
	voucher  *entity.Voucher
	discount int64
	rejected string
}

// Evaluate returns the single best promotion applicable to req, per BRB §3.4
// Campaign Priority Rules and §4.7 (maximum one voucher per trip).
func (s *PromotionService) Evaluate(ctx context.Context, req *entity.PromotionRequest, now time.Time) (*entity.PromotionResult, error) {
	candidates, err := s.gatherCandidates(ctx, req)
	if err != nil {
		return nil, err
	}

	var warnings []string
	var eligible []candidateOutcome

	for _, v := range candidates {
		outcome, err := s.evaluateOne(ctx, v, req, now)
		if err != nil {
			return nil, err
		}
		if outcome.rejected != "" {
			warnings = append(warnings, fmt.Sprintf("%s (%s): %s", v.Name, v.Type, outcome.rejected))
			continue
		}
		eligible = append(eligible, outcome)
	}

	if len(eligible) == 0 {
		reason := "no eligible promotion found for this trip"
		if req.VoucherCode != "" {
			reason = "voucher code is not valid for this trip"
		}
		return entity.NoDiscount(req.OrderAmount, reason, warnings), nil
	}

	sort.SliceStable(eligible, func(i, j int) bool {
		a, b := eligible[i], eligible[j]
		if a.voucher.Priority != b.voucher.Priority {
			return a.voucher.Priority < b.voucher.Priority // BRB §3.4: lower number = higher priority
		}
		if a.discount != b.discount {
			return a.discount > b.discount // BRB §3.5: equal priority -> higher monetary value wins
		}
		return a.voucher.EndTime.Before(b.voucher.EndTime) // BRB §3.5: equal value -> earlier expiry wins
	})

	best := eligible[0]

	// BRB §4.7: maximum ONE voucher per trip by default. Other eligible
	// candidates are surfaced as warnings, not applied. True multi-voucher
	// stacking (Voucher.Stackable) is a documented TODO — see voucher.go.
	for _, other := range eligible[1:] {
		warnings = append(warnings, fmt.Sprintf(
			"%s (%s) was also eligible (discount %d VND) but not applied: BRB §4.7 allows at most one voucher per trip",
			other.voucher.Name, other.voucher.Type, other.discount))
	}

	finalAmount := req.OrderAmount - best.discount
	if finalAmount < 0 {
		finalAmount = 0
	}

	return &entity.PromotionResult{
		Applied:             true,
		VoucherID:           best.voucher.ID,
		VoucherCode:         best.voucher.Code,
		VoucherName:         best.voucher.Name,
		Type:                best.voucher.Type,
		DiscountType:        best.voucher.DiscountType,
		DiscountAmount:      best.discount,
		OriginalOrderAmount: req.OrderAmount,
		FinalOrderAmount:    finalAmount,
		Reason:              "applied " + best.voucher.Name,
		Warnings:            warnings,
	}, nil
}

func (s *PromotionService) gatherCandidates(ctx context.Context, req *entity.PromotionRequest) ([]*entity.Voucher, error) {
	if req.VoucherCode != "" {
		v, err := s.repo.FindByCode(ctx, req.VoucherCode)
		if err != nil {
			if sharederrors.IsCode(err, sharederrors.CodeNotFound) {
				return nil, nil
			}
			return nil, err
		}
		return []*entity.Voucher{v}, nil
	}
	return s.repo.FindAutoApplyCandidates(ctx, req.City, req.VehicleType, entity.AllPromotionTypes())
}

func (s *PromotionService) evaluateOne(ctx context.Context, v *entity.Voucher, req *entity.PromotionRequest, now time.Time) (candidateOutcome, error) {
	if err := s.validator.Validate(v, req, now); err != nil {
		return candidateOutcome{voucher: v, rejected: err.Error()}, nil
	}

	rule, ok := s.rules[v.Type]
	if !ok {
		return candidateOutcome{voucher: v, rejected: "no rule registered for this promotion type"}, nil
	}
	if eligible, reason := rule.IsEligible(ctx, v, req, now); !eligible {
		return candidateOutcome{voucher: v, rejected: reason}, nil
	}

	usage, err := s.repo.UsageCountForRider(ctx, v.ID, req.RiderID)
	if err != nil {
		return candidateOutcome{}, err
	}
	if usage >= v.PerUserLimit() {
		return candidateOutcome{voucher: v, rejected: "rider has already used this voucher the maximum number of times"}, nil
	}

	discount := entity.ComputeDiscount(v.DiscountType, v.DiscountValue, v.MaxDiscount, req.OrderAmount)
	if discount > req.OrderAmount {
		discount = req.OrderAmount // BRB §4.9: cannot reduce amount paid below 0
	}
	if discount <= 0 {
		return candidateOutcome{voucher: v, rejected: "computed discount is zero for this order amount"}, nil
	}
	if !v.HasBudgetFor(discount) {
		return candidateOutcome{voucher: v, rejected: "voucher campaign budget cannot cover this discount"}, nil
	}

	return candidateOutcome{voucher: v, discount: discount}, nil
}

// Redeem commits a previously-evaluated PromotionResult against a confirmed
// trip: decrements the voucher's remaining budget and increments usage
// counters. No-op if result.Applied is false.
func (s *PromotionService) Redeem(ctx context.Context, result *entity.PromotionResult, riderID, tripID string) error {
	if result == nil || !result.Applied {
		return nil
	}
	return s.repo.RecordRedemption(ctx, result.VoucherID, riderID, tripID, result.DiscountAmount)
}

// ReleaseRedemption reverses a Redeem — BRB §4.13 Refund Behaviour / §4.14
// Cancellation Behaviour: voucher is reinstated when the trip did not consume
// it through rider fault. No-op if result.Applied is false.
func (s *PromotionService) ReleaseRedemption(ctx context.Context, result *entity.PromotionResult, riderID, tripID string) error {
	if result == nil || !result.Applied {
		return nil
	}
	return s.repo.ReleaseRedemption(ctx, result.VoucherID, riderID, tripID, result.DiscountAmount)
}
