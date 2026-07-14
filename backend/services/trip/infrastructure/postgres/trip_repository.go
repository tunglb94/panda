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
			created_at, updated_at,
			has_commission_detail, commission_cents, driver_income_cents,
			voucher_discount_cents, commission_rate, voucher_id, voucher_code,
			arrived_at, started_at, travelled_distance_km, travelled_duration_min,
			waiting_duration_min, toll_fee_cents, extra_fee_cents
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24,$25,$26)
		ON CONFLICT (trip_id) DO UPDATE SET
			driver_id               = EXCLUDED.driver_id,
			status                  = EXCLUDED.status,
			cancellation_reason     = EXCLUDED.cancellation_reason,
			final_fare_total        = EXCLUDED.final_fare_total,
			fare_currency           = EXCLUDED.fare_currency,
			payment_method          = EXCLUDED.payment_method,
			updated_at              = EXCLUDED.updated_at,
			has_commission_detail   = EXCLUDED.has_commission_detail,
			commission_cents        = EXCLUDED.commission_cents,
			driver_income_cents     = EXCLUDED.driver_income_cents,
			voucher_discount_cents  = EXCLUDED.voucher_discount_cents,
			commission_rate         = EXCLUDED.commission_rate,
			voucher_id              = EXCLUDED.voucher_id,
			voucher_code            = EXCLUDED.voucher_code,
			arrived_at              = EXCLUDED.arrived_at,
			started_at              = EXCLUDED.started_at,
			travelled_distance_km   = EXCLUDED.travelled_distance_km,
			travelled_duration_min  = EXCLUDED.travelled_duration_min,
			waiting_duration_min    = EXCLUDED.waiting_duration_min,
			toll_fee_cents          = EXCLUDED.toll_fee_cents,
			extra_fee_cents         = EXCLUDED.extra_fee_cents`

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
		trip.HasCommissionDetail,
		trip.CommissionCents,
		trip.DriverIncomeCents,
		trip.VoucherDiscountCents,
		trip.CommissionRate,
		trip.VoucherID,
		trip.VoucherCode,
		utcOrNil(trip.ArrivedAt),
		utcOrNil(trip.StartedAt),
		trip.TravelledDistanceKm,
		trip.TravelledDurationMin,
		trip.WaitingDurationMin,
		trip.TollFeeCents,
		trip.ExtraFeeCents,
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
		       created_at, updated_at,
		       has_commission_detail, commission_cents, driver_income_cents,
		       voucher_discount_cents, commission_rate, voucher_id, voucher_code,
		       arrived_at, started_at, travelled_distance_km, travelled_duration_min,
		       waiting_duration_min, toll_fee_cents, extra_fee_cents
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
		       created_at, updated_at,
		       has_commission_detail, commission_cents, driver_income_cents,
		       voucher_discount_cents, commission_rate, voucher_id, voucher_code,
		       arrived_at, started_at, travelled_distance_km, travelled_duration_min,
		       waiting_duration_min, toll_fee_cents, extra_fee_cents
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
		       created_at, updated_at,
		       has_commission_detail, commission_cents, driver_income_cents,
		       voucher_discount_cents, commission_rate, voucher_id, voucher_code,
		       arrived_at, started_at, travelled_distance_km, travelled_duration_min,
		       waiting_duration_min, toll_fee_cents, extra_fee_cents
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
		tripID, riderID, driverID, status           string
		pickupAddress, dropoffAddress, cancellation string
		fareCurrency, paymentMethod                 string
		finalFareTotal                              int64
		createdAt, updatedAt                        time.Time
		hasCommissionDetail                         bool
		commissionCents, driverIncomeCents          int64
		voucherDiscountCents                        int64
		commissionRate                              float64
		voucherID, voucherCode                      string
		arrivedAt, startedAt                        *time.Time
		travelledDistanceKm, travelledDurationMin   float64
		waitingDurationMin                          float64
		tollFeeCents, extraFeeCents                 int64
	)
	err := row.Scan(
		&tripID, &riderID, &driverID, &status,
		&pickupAddress, &dropoffAddress, &cancellation,
		&finalFareTotal, &fareCurrency, &paymentMethod,
		&createdAt, &updatedAt,
		&hasCommissionDetail, &commissionCents, &driverIncomeCents,
		&voucherDiscountCents, &commissionRate, &voucherID, &voucherCode,
		&arrivedAt, &startedAt, &travelledDistanceKm, &travelledDurationMin,
		&waitingDurationMin, &tollFeeCents, &extraFeeCents,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domainerrors.NotFound("trip not found")
		}
		return nil, domainerrors.Internal("trip: scan failed").WithMeta("error", err.Error())
	}
	if arrivedAt != nil {
		t := arrivedAt.UTC()
		arrivedAt = &t
	}
	if startedAt != nil {
		t := startedAt.UTC()
		startedAt = &t
	}
	return entity.ReconstituteTrip(
		tripID, riderID, driverID,
		entity.TripStatus(status),
		pickupAddress, dropoffAddress, cancellation,
		finalFareTotal, fareCurrency, paymentMethod,
		createdAt.UTC(), updatedAt.UTC(),
		entity.CompleteFinancials{
			HasCommissionDetail:  hasCommissionDetail,
			CommissionCents:      commissionCents,
			DriverIncomeCents:    driverIncomeCents,
			VoucherDiscountCents: voucherDiscountCents,
			CommissionRate:       commissionRate,
			VoucherID:            voucherID,
			VoucherCode:          voucherCode,
		},
		arrivedAt, startedAt,
		waitingDurationMin,
		entity.TripSummary{
			TravelledDistanceKm:  travelledDistanceKm,
			TravelledDurationMin: travelledDurationMin,
			TollFeeCents:         tollFeeCents,
			ExtraFeeCents:        extraFeeCents,
		},
	), nil
}

// utcOrNil normalizes a *time.Time to UTC for storage, preserving nil.
func utcOrNil(t *time.Time) *time.Time {
	if t == nil {
		return nil
	}
	u := t.UTC()
	return &u
}
