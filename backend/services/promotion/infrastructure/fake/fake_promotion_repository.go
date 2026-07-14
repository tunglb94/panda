// Package fake provides an in-memory PromotionRepository implementation,
// explicitly requested by the sprint brief ("PHẢI VIẾT ... fake repository").
// It is safe for concurrent use and is suitable both for unit/integration
// tests and for local development without a Postgres instance.
package fake

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/fairride/promotion/domain/entity"
	"github.com/fairride/shared/errors"
)

// riderRedemption is one logged reservation/redemption/release event —
// status moves 'reserved' -> 'redeemed' | 'released' in place, mirroring
// the postgres voucher_redemptions row lifecycle.
type riderRedemption struct {
	record entity.RedemptionRecord
}

// issuanceEntry is one (voucher, rider) wallet grant.
type issuanceEntry struct {
	status   entity.VoucherIssuanceStatus
	issuedAt time.Time
	usedAt   *time.Time
}

type FakePromotionRepository struct {
	mu            sync.Mutex
	vouchers      map[string]*entity.Voucher   // by ID
	codeIndex     map[string]string             // lowercased code -> ID
	redemptionLog []riderRedemption
	issuances     map[string]*issuanceEntry // "voucherID|riderID" -> entry
}

func NewFakePromotionRepository() *FakePromotionRepository {
	return &FakePromotionRepository{
		vouchers:  make(map[string]*entity.Voucher),
		codeIndex: make(map[string]string),
		issuances: make(map[string]*issuanceEntry),
	}
}

// Seed inserts a voucher directly, bypassing Save's normal flow. Convenience
// for test setup.
func (f *FakePromotionRepository) Seed(v *entity.Voucher) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.store(v)
}

func (f *FakePromotionRepository) FindByID(_ context.Context, id string) (*entity.Voucher, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	v, ok := f.vouchers[id]
	if !ok {
		return nil, errors.NotFound("voucher not found: " + id)
	}
	return clone(v), nil
}

func (f *FakePromotionRepository) FindByCode(_ context.Context, code string) (*entity.Voucher, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	id, ok := f.codeIndex[strings.ToLower(code)]
	if !ok {
		return nil, errors.NotFound("voucher not found for code: " + code)
	}
	return clone(f.vouchers[id]), nil
}

func (f *FakePromotionRepository) FindAutoApplyCandidates(_ context.Context, city, vehicleType string, types []entity.PromotionType) ([]*entity.Voucher, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	typeSet := make(map[entity.PromotionType]bool, len(types))
	for _, t := range types {
		typeSet[t] = true
	}

	var result []*entity.Voucher
	for _, v := range f.vouchers {
		if v.Code != "" {
			continue
		}
		if v.Status != entity.VoucherStatusActive {
			continue
		}
		if len(typeSet) > 0 && !typeSet[v.Type] {
			continue
		}
		if len(v.Cities) > 0 && !containsFold(v.Cities, city) {
			continue
		}
		if len(v.VehicleTypes) > 0 && !containsFold(v.VehicleTypes, vehicleType) {
			continue
		}
		result = append(result, clone(v))
	}
	return result, nil
}

func (f *FakePromotionRepository) Save(_ context.Context, v *entity.Voucher) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.store(v)
	return nil
}

func (f *FakePromotionRepository) store(v *entity.Voucher) {
	f.vouchers[v.ID] = clone(v)
	if v.Code != "" {
		f.codeIndex[strings.ToLower(v.Code)] = v.ID
	}
}

// UsageCountForRider counts 'reserved' + 'redeemed' entries — an in-flight
// reservation already counts against the per-rider limit.
func (f *FakePromotionRepository) UsageCountForRider(_ context.Context, voucherID, riderID string) (int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var count int64
	for _, e := range f.redemptionLog {
		if e.record.VoucherID == voucherID && e.record.RiderID == riderID &&
			(e.record.Status == "reserved" || e.record.Status == "redeemed") {
			count++
		}
	}
	return count, nil
}

// Reserve is idempotent: retrying for the same (voucher, rider, trip) is a no-op.
func (f *FakePromotionRepository) Reserve(_ context.Context, voucherID, riderID, tripID string, discountAmount int64) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	for _, e := range f.redemptionLog {
		if e.record.VoucherID == voucherID && e.record.RiderID == riderID && e.record.TripID == tripID {
			return nil
		}
	}

	v, ok := f.vouchers[voucherID]
	if !ok {
		return errors.NotFound("voucher not found: " + voucherID)
	}
	now := time.Now()
	if err := v.Reserve(discountAmount, now); err != nil {
		return err
	}
	f.vouchers[voucherID] = v
	f.redemptionLog = append(f.redemptionLog, riderRedemption{
		record: entity.RedemptionRecord{
			VoucherID: voucherID, VoucherCode: v.Code, VoucherName: v.Name, RiderID: riderID,
			TripID: tripID, DiscountAmount: discountAmount, Status: "reserved", RedeemedAt: now,
		},
	})
	return nil
}

// ConfirmRedeem transitions a 'reserved' entry to 'redeemed'. Idempotent.
func (f *FakePromotionRepository) ConfirmRedeem(_ context.Context, voucherID, riderID, tripID string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	for i := range f.redemptionLog {
		e := &f.redemptionLog[i]
		if e.record.VoucherID == voucherID && e.record.RiderID == riderID && e.record.TripID == tripID {
			if e.record.Status == "redeemed" {
				return nil
			}
			if e.record.Status != "reserved" {
				return errors.PreconditionFailed("voucher reservation is not in a redeemable state: " + e.record.Status)
			}
			e.record.Status = "redeemed"
			return nil
		}
	}
	return errors.NotFound("no voucher reservation found for this trip")
}

// Release transitions a 'reserved' entry to 'released' and reinstates
// discountAmount to the voucher's budget. Idempotent.
func (f *FakePromotionRepository) Release(_ context.Context, voucherID, riderID, tripID string, discountAmount int64) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	for i := range f.redemptionLog {
		e := &f.redemptionLog[i]
		if e.record.VoucherID == voucherID && e.record.RiderID == riderID && e.record.TripID == tripID {
			if e.record.Status == "released" {
				return nil
			}
			if e.record.Status != "reserved" {
				return errors.PreconditionFailed("no reserved voucher redemption found for this rider/voucher/trip")
			}
			v, ok := f.vouchers[voucherID]
			if !ok {
				return errors.NotFound("voucher not found: " + voucherID)
			}
			if err := v.Release(discountAmount, time.Now()); err != nil {
				return err
			}
			f.vouchers[voucherID] = v
			e.record.Status = "released"
			return nil
		}
	}
	return errors.PreconditionFailed("no reserved voucher redemption found for this rider/voucher/trip")
}

// FindReservationByTrip returns the redemption entry for tripID regardless of status.
func (f *FakePromotionRepository) FindReservationByTrip(_ context.Context, tripID string) (*entity.RedemptionRecord, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, e := range f.redemptionLog {
		if e.record.TripID == tripID {
			rec := e.record
			return &rec, nil
		}
	}
	return nil, errors.NotFound("no voucher reservation found for this trip")
}

// FindAll returns every voucher campaign.
func (f *FakePromotionRepository) FindAll(_ context.Context) ([]*entity.Voucher, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	result := make([]*entity.Voucher, 0, len(f.vouchers))
	for _, v := range f.vouchers {
		result = append(result, clone(v))
	}
	return result, nil
}

// ListRedemptionsByRider returns riderID's full redemption history, newest first.
func (f *FakePromotionRepository) ListRedemptionsByRider(_ context.Context, riderID string) ([]*entity.RedemptionRecord, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var result []*entity.RedemptionRecord
	for i := len(f.redemptionLog) - 1; i >= 0; i-- {
		e := f.redemptionLog[i]
		if e.record.RiderID == riderID {
			rec := e.record
			result = append(result, &rec)
		}
	}
	return result, nil
}

// ─── Per-rider issuance ─────────────────────────────────────────────────────

func (f *FakePromotionRepository) IssueToRider(_ context.Context, voucherID, riderID string, now time.Time) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	key := redemptionKey(voucherID, riderID)
	if _, exists := f.issuances[key]; exists {
		return nil
	}
	f.issuances[key] = &issuanceEntry{status: entity.IssuanceStatusIssued, issuedAt: now}
	return nil
}

func (f *FakePromotionRepository) ListIssuancesForRider(_ context.Context, riderID string) ([]*entity.VoucherIssuance, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var result []*entity.VoucherIssuance
	for key, e := range f.issuances {
		voucherID, rID, ok := splitRedemptionKey(key)
		if !ok || rID != riderID {
			continue
		}
		v := f.vouchers[voucherID]
		result = append(result, &entity.VoucherIssuance{
			VoucherID: voucherID, RiderID: riderID, Status: e.status,
			IssuedAt: e.issuedAt, UsedAt: e.usedAt, Voucher: clone(v),
		})
	}
	return result, nil
}

func (f *FakePromotionRepository) MarkIssuanceUsed(_ context.Context, voucherID, riderID string, now time.Time) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	e, ok := f.issuances[redemptionKey(voucherID, riderID)]
	if !ok || e.status != entity.IssuanceStatusIssued {
		return nil
	}
	e.status = entity.IssuanceStatusUsed
	t := now
	e.usedAt = &t
	return nil
}

func (f *FakePromotionRepository) CountIssued(_ context.Context, voucherID string) (int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var count int64
	for key := range f.issuances {
		if vID, _, ok := splitRedemptionKey(key); ok && vID == voucherID {
			count++
		}
	}
	return count, nil
}

func (f *FakePromotionRepository) CountRedeemed(_ context.Context, voucherID string) (int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var count int64
	for _, e := range f.redemptionLog {
		if e.record.VoucherID == voucherID && e.record.Status == "redeemed" {
			count++
		}
	}
	return count, nil
}

func (f *FakePromotionRepository) CountExpiredIssuances(_ context.Context, voucherID string, now time.Time) (int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	v, ok := f.vouchers[voucherID]
	if !ok || !now.After(v.EndTime) {
		return 0, nil
	}
	var count int64
	for key, e := range f.issuances {
		if vID, _, ok := splitRedemptionKey(key); ok && vID == voucherID && e.status == entity.IssuanceStatusIssued {
			count++
		}
	}
	return count, nil
}

func redemptionKey(voucherID, riderID string) string {
	return voucherID + "|" + riderID
}

func splitRedemptionKey(key string) (voucherID, riderID string, ok bool) {
	idx := strings.Index(key, "|")
	if idx < 0 {
		return "", "", false
	}
	return key[:idx], key[idx+1:], true
}

func containsFold(list []string, target string) bool {
	for _, item := range list {
		if strings.EqualFold(item, target) {
			return true
		}
	}
	return false
}

func clone(v *entity.Voucher) *entity.Voucher {
	if v == nil {
		return nil
	}
	c := *v
	c.VehicleTypes = append([]string(nil), v.VehicleTypes...)
	c.Cities = append([]string(nil), v.Cities...)
	c.Membership = append([]string(nil), v.Membership...)
	return &c
}
