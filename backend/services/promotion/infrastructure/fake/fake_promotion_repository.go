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

type FakePromotionRepository struct {
	mu          sync.Mutex
	vouchers    map[string]*entity.Voucher // by ID
	codeIndex   map[string]string          // lowercased code -> ID
	redemptions map[string]int64           // "voucherID|riderID" -> count
}

func NewFakePromotionRepository() *FakePromotionRepository {
	return &FakePromotionRepository{
		vouchers:    make(map[string]*entity.Voucher),
		codeIndex:   make(map[string]string),
		redemptions: make(map[string]int64),
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

func (f *FakePromotionRepository) UsageCountForRider(_ context.Context, voucherID, riderID string) (int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.redemptions[redemptionKey(voucherID, riderID)], nil
}

func (f *FakePromotionRepository) RecordRedemption(_ context.Context, voucherID, riderID, _ string, discountAmount int64) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	v, ok := f.vouchers[voucherID]
	if !ok {
		return errors.NotFound("voucher not found: " + voucherID)
	}
	if err := v.Reserve(discountAmount, time.Now()); err != nil {
		return err
	}
	f.vouchers[voucherID] = v
	f.redemptions[redemptionKey(voucherID, riderID)]++
	return nil
}

func (f *FakePromotionRepository) ReleaseRedemption(_ context.Context, voucherID, riderID, _ string, discountAmount int64) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	v, ok := f.vouchers[voucherID]
	if !ok {
		return errors.NotFound("voucher not found: " + voucherID)
	}
	key := redemptionKey(voucherID, riderID)
	if f.redemptions[key] <= 0 {
		return errors.PreconditionFailed("no redemption to release for this rider/voucher pair")
	}
	if err := v.Release(discountAmount, time.Now()); err != nil {
		return err
	}
	f.vouchers[voucherID] = v
	f.redemptions[key]--
	return nil
}

func redemptionKey(voucherID, riderID string) string {
	return voucherID + "|" + riderID
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
