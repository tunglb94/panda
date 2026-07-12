// Package memory provides an in-memory DeliveryRepository implementation for
// Delivery V1 Phase 1 (docs/business/DELIVERY_V1_DESIGN.md), mirroring the
// pattern already used by backend/services/promotion/infrastructure/fake.
// Suitable for unit tests and local development without Postgres; no
// Postgres-backed DeliveryRepository exists yet — persistence wiring is
// deferred to a later phase (see the design doc's Phần 17 Migration).
package memory

import (
	"context"
	"sync"

	"github.com/fairride/shared/errors"
	"github.com/fairride/trip/domain/entity"
	"github.com/fairride/trip/domain/repository"
)

// compile-time assertion: DeliveryRepository satisfies repository.DeliveryRepository.
var _ repository.DeliveryRepository = (*DeliveryRepository)(nil)

// DeliveryRepository is a concurrency-safe in-memory implementation of
// repository.DeliveryRepository.
type DeliveryRepository struct {
	mu         sync.Mutex
	deliveries map[string]*entity.Delivery
}

// NewDeliveryRepository returns an empty in-memory DeliveryRepository.
func NewDeliveryRepository() *DeliveryRepository {
	return &DeliveryRepository{
		deliveries: make(map[string]*entity.Delivery),
	}
}

// Save upserts a delivery, storing a defensive copy.
func (r *DeliveryRepository) Save(_ context.Context, delivery *entity.Delivery) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if delivery == nil {
		return errors.InvalidArgument("delivery must not be nil")
	}
	clone := *delivery
	r.deliveries[delivery.DeliveryID] = &clone
	return nil
}

// FindByID returns CodeNotFound if no delivery has the given ID.
func (r *DeliveryRepository) FindByID(_ context.Context, deliveryID string) (*entity.Delivery, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	d, ok := r.deliveries[deliveryID]
	if !ok {
		return nil, errors.NotFound("delivery not found: " + deliveryID)
	}
	clone := *d
	return &clone, nil
}
