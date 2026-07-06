package postgres

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/fairride/dispatch/domain/entity"
	"github.com/fairride/dispatch/domain/repository"
	domainerrors "github.com/fairride/shared/errors"
)

// DispatchRepository is the PostgreSQL implementation of repository.DispatchJobRepository.
type DispatchRepository struct {
	pool *pgxpool.Pool
}

var _ repository.DispatchJobRepository = (*DispatchRepository)(nil)

func NewDispatchRepository(pool *pgxpool.Pool) *DispatchRepository {
	return &DispatchRepository{pool: pool}
}

// Save upserts a dispatch job. trip_id is immutable after the first insert.
func (r *DispatchRepository) Save(ctx context.Context, job *entity.DispatchJob) error {
	var offerExpiresAt *time.Time
	if !job.OfferExpiresAt.IsZero() {
		t := job.OfferExpiresAt.UTC()
		offerExpiresAt = &t
	}

	const q = `
		INSERT INTO dispatch_jobs (
			job_id, trip_id, rider_id, pickup_lat, pickup_lon,
			status, current_driver_id, assigned_driver_id, offered_driver_ids,
			offer_expires_at, offer_timeout_sec, max_attempts, attempt_count,
			created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)
		ON CONFLICT (job_id) DO UPDATE SET
			status             = EXCLUDED.status,
			current_driver_id  = EXCLUDED.current_driver_id,
			assigned_driver_id = EXCLUDED.assigned_driver_id,
			offered_driver_ids = EXCLUDED.offered_driver_ids,
			offer_expires_at   = EXCLUDED.offer_expires_at,
			attempt_count      = EXCLUDED.attempt_count,
			updated_at         = EXCLUDED.updated_at`

	_, err := r.pool.Exec(ctx, q,
		job.JobID,
		job.TripID,
		job.RiderID,
		job.PickupLat,
		job.PickupLon,
		string(job.Status),
		job.CurrentDriverID,
		job.AssignedDriverID,
		job.OfferedDriverIDsCSV(),
		offerExpiresAt,
		job.OfferTimeoutSec,
		job.MaxAttempts,
		job.AttemptCount,
		job.CreatedAt.UTC(),
		job.UpdatedAt.UTC(),
	)
	if err != nil {
		return domainerrors.Internal("dispatch: save job").WithMeta("error", err.Error())
	}
	return nil
}

// FindByID returns a single dispatch job or CodeNotFound.
func (r *DispatchRepository) FindByID(ctx context.Context, jobID string) (*entity.DispatchJob, error) {
	const q = `
		SELECT job_id, trip_id, rider_id, pickup_lat, pickup_lon,
		       status, current_driver_id, assigned_driver_id, offered_driver_ids,
		       offer_expires_at, offer_timeout_sec, max_attempts, attempt_count,
		       created_at, updated_at
		FROM dispatch_jobs WHERE job_id = $1`
	return r.scanOne(r.pool.QueryRow(ctx, q, jobID))
}

// FindByTripID returns the dispatch job for a trip or CodeNotFound.
func (r *DispatchRepository) FindByTripID(ctx context.Context, tripID string) (*entity.DispatchJob, error) {
	const q = `
		SELECT job_id, trip_id, rider_id, pickup_lat, pickup_lon,
		       status, current_driver_id, assigned_driver_id, offered_driver_ids,
		       offer_expires_at, offer_timeout_sec, max_attempts, attempt_count,
		       created_at, updated_at
		FROM dispatch_jobs WHERE trip_id = $1`
	return r.scanOne(r.pool.QueryRow(ctx, q, tripID))
}

// FindExpiredOffers returns searching jobs whose offer has timed out.
func (r *DispatchRepository) FindExpiredOffers(ctx context.Context, now time.Time) ([]*entity.DispatchJob, error) {
	const q = `
		SELECT job_id, trip_id, rider_id, pickup_lat, pickup_lon,
		       status, current_driver_id, assigned_driver_id, offered_driver_ids,
		       offer_expires_at, offer_timeout_sec, max_attempts, attempt_count,
		       created_at, updated_at
		FROM dispatch_jobs
		WHERE status = 'searching' AND offer_expires_at < $1`

	rows, err := r.pool.Query(ctx, q, now.UTC())
	if err != nil {
		return nil, domainerrors.Internal("dispatch: find expired offers").WithMeta("error", err.Error())
	}
	defer rows.Close()

	var jobs []*entity.DispatchJob
	for rows.Next() {
		job, err := r.scanOne(rows)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, job)
	}
	if err := rows.Err(); err != nil {
		return nil, domainerrors.Internal("dispatch: rows iteration").WithMeta("error", err.Error())
	}
	return jobs, nil
}

// ─── private helpers ──────────────────────────────────────────────────────────

type rowScanner interface {
	Scan(dest ...any) error
}

func (r *DispatchRepository) scanOne(row rowScanner) (*entity.DispatchJob, error) {
	return scanDispatchJob(row)
}

// scanDispatchJob is a package-level helper used by both DispatchRepository and
// txDispatchRepository to reconstruct a DispatchJob from a scanned row.
func scanDispatchJob(row rowScanner) (*entity.DispatchJob, error) {
	var (
		jobID, tripID, riderID            string
		pickupLat, pickupLon              float64
		status                            string
		currentDriverID, assignedDriverID string
		offeredDriverIDsCSV               string
		offerExpiresAt                    *time.Time
		offerTimeoutSec, maxAttempts, attemptCount int
		createdAt, updatedAt              time.Time
	)

	err := row.Scan(
		&jobID, &tripID, &riderID, &pickupLat, &pickupLon,
		&status, &currentDriverID, &assignedDriverID, &offeredDriverIDsCSV,
		&offerExpiresAt, &offerTimeoutSec, &maxAttempts, &attemptCount,
		&createdAt, &updatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domainerrors.NotFound("dispatch job not found")
		}
		return nil, domainerrors.Internal("dispatch: scan job").WithMeta("error", err.Error())
	}

	var oea time.Time
	if offerExpiresAt != nil {
		oea = offerExpiresAt.UTC()
	}

	var offered []string
	if offeredDriverIDsCSV != "" {
		offered = strings.Split(offeredDriverIDsCSV, ",")
	}

	return entity.ReconstituteDispatchJob(
		jobID, tripID, riderID,
		pickupLat, pickupLon,
		entity.JobStatus(status),
		currentDriverID, assignedDriverID,
		offered,
		oea,
		offerTimeoutSec, maxAttempts, attemptCount,
		createdAt.UTC(), updatedAt.UTC(),
	), nil
}
