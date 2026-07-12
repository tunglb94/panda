package repository

import (
	"context"

	"github.com/fairride/trip/domain/entity"
)

// DeliveryRepository persists and retrieves Delivery aggregates.
// All methods return *errors.DomainError on failure, matching
// TripRepository's convention (trip_repository.go).
type DeliveryRepository interface {
	// Save upserts a delivery.
	Save(ctx context.Context, delivery *entity.Delivery) error

	// FindByID returns CodeNotFound if no delivery has the given ID.
	FindByID(ctx context.Context, deliveryID string) (*entity.Delivery, error)
}
