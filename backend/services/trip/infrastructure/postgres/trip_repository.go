package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	domainerrors "github.com/fairride/shared/errors"
	"github.com/fairride/trip/domain/entity"
	"github.com/fairride/trip/domain/repository"
)

// TripRepository is the PostgreSQL implementation of repository.TripRepository.
type TripRepository struct {
	pool *pgxpool.Pool
}

var _ repository.TripRepository = (*TripRepository)(nil)

func NewTripRepository(pool *pgxpool.Pool) *TripRepository {
	return &TripRepository{pool: pool}
}

// Save upserts a trip. On conflict, all mutable fields are updated except rider_id
// and pickup_address / dropoff_address (set at creation time).
func (r *TripRepository) Save(ctx context.Context, trip *entity.Trip) error {
	const q = `
		INSERT INTO trips (
			trip_id, rider_id, driver_id, status,
			pickup_address, dropoff_address, cancellation_reason,
			final_fare_total, fare_currency, payment_method,
			created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
		ON CONFLICT (trip_id) DO UPDATE SET
			driver_id           = EXCLUDED.driver_id,
			status              = EXCLUDED.status,
			cancellation_reason = EXCLUDED.cancellation_reason,
			final_fare_total    = EXCLUDED.final_fare_total,
			fare_currency       = EXCLUDED.fare_currency,
			payment_method      = EXCLUDED.payment_method,
			updated_at          = EXCLUDED.updated_at`

	_, err := r.pool.Exec(ctx, q,
		trip.TripID,
		trip.RiderID,
		trip.DriverID,
		string(trip.Status),
		trip.PickupAddress,
		trip.DropoffAddress,
		trip.CancellationReason,
		trip.FinalFareTotal,
		trip.FareCurrency,
		trip.PaymentMethod,
		trip.CreatedAt.UTC(),
		trip.UpdatedAt.UTC(),
	)
	if err != nil {
		return domainerrors.Internal("trip: save failed").WithMeta("error", err.Error())
	}
	return nil
}

// FindByID returns a single trip or CodeNotFound.
func (r *TripRepository) FindByID(ctx context.Context, tripID string) (*entity.Trip, error) {
	const q = `
		SELECT trip_id, rider_id, driver_id, status,
		       pickup_address, dropoff_address, cancellation_reason,
		       final_fare_total, fare_currency, payment_method,
		       created_at, updated_at
		FROM trips
		WHERE trip_id = $1`

	return r.scanOne(r.pool.QueryRow(ctx, q, tripID))
}

// FindByRiderID returns all trips for a rider, newest first.
func (r *TripRepository) FindByRiderID(ctx context.Context, riderID string) ([]*entity.Trip, error) {
	const q = `
		SELECT trip_id, rider_id, driver_id, status,
		       pickup_address, dropoff_address, cancellation_reason,
		       final_fare_total, fare_currency, payment_method,
		       created_at, updated_at
		FROM trips
		WHERE rider_id = $1
		ORDER BY created_at DESC`

	rows, err := r.pool.Query(ctx, q, riderID)
	if err != nil {
		return nil, domainerrors.Internal("trip: find by rider query failed").WithMeta("error", err.Error())
	}
	defer rows.Close()

	var trips []*entity.Trip
	for rows.Next() {
		trip, err := r.scanOne(rows)
		if err != nil {
			return nil, err
		}
		trips = append(trips, trip)
	}
	if err := rows.Err(); err != nil {
		return nil, domainerrors.Internal("trip: rows iteration failed").WithMeta("error", err.Error())
	}
	return trips, nil
}

// FindByDriverID returns all trips for a driver, newest first.
func (r *TripRepository) FindByDriverID(ctx context.Context, driverID string) ([]*entity.Trip, error) {
	const q = `
		SELECT trip_id, rider_id, driver_id, status,
		       pickup_address, dropoff_address, cancellation_reason,
		       final_fare_total, fare_currency, payment_method,
		       created_at, updated_at
		FROM trips
		WHERE driver_id = $1
		ORDER BY created_at DESC`

	rows, err := r.pool.Query(ctx, q, driverID)
	if err != nil {
		return nil, domainerrors.Internal("trip: find by driver query failed").WithMeta("error", err.Error())
	}
	defer rows.Close()

	var trips []*entity.Trip
	for rows.Next() {
		trip, err := r.scanOne(rows)
		if err != nil {
			return nil, err
		}
		trips = append(trips, trip)
	}
	if err := rows.Err(); err != nil {
		return nil, domainerrors.Internal("trip: rows iteration failed").WithMeta("error", err.Error())
	}
	return trips, nil
}

// ─── private helpers ──────────────────────────────────────────────────────────

type rowScanner interface {
	Scan(dest ...any) error
}

func (r *TripRepository) scanOne(row rowScanner) (*entity.Trip, error) {
	var (
		tripID, riderID, driverID, status            string
		pickupAddress, dropoffAddress, cancellation string
		fareCurrency, paymentMethod                 string
		finalFareTotal                              int64
		createdAt, updatedAt                        time.Time
	)
	err := row.Scan(
		&tripID, &riderID, &driverID, &status,
		&pickupAddress, &dropoffAddress, &cancellation,
		&finalFareTotal, &fareCurrency, &paymentMethod,
		&createdAt, &updatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domainerrors.NotFound("trip not found")
		}
		return nil, domainerrors.Internal("trip: scan failed").WithMeta("error", err.Error())
	}
	return entity.ReconstituteTrip(
		tripID, riderID, driverID,
		entity.TripStatus(status),
		pickupAddress, dropoffAddress, cancellation,
		finalFareTotal, fareCurrency, paymentMethod,
		createdAt.UTC(), updatedAt.UTC(),
	), nil
}
