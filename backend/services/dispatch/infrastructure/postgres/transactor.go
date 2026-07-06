package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/fairride/dispatch/domain/entity"
	"github.com/fairride/dispatch/domain/repository"
	domainerrors "github.com/fairride/shared/errors"
)

// Transactor implements repository.Transactor using pgx pool transactions.
type Transactor struct {
	pool *pgxpool.Pool
}

var _ repository.Transactor = (*Transactor)(nil)

func NewTransactor(pool *pgxpool.Pool) *Transactor {
	return &Transactor{pool: pool}
}

// WithinTx begins a PostgreSQL transaction, passes tx-scoped repository
// implementations to fn, and commits on success. If fn returns an error the
// transaction is rolled back and the same error is returned to the caller.
// The deferred Rollback is a no-op after a successful Commit.
func (t *Transactor) WithinTx(
	ctx context.Context,
	fn func(jobs repository.DispatchJobRepository, trips repository.TripUpdater) error,
) error {
	tx, err := t.pool.Begin(ctx)
	if err != nil {
		return domainerrors.Internal("dispatch: begin transaction").WithMeta("error", err.Error())
	}
	defer tx.Rollback(ctx) //nolint:errcheck // rollback error is intentionally ignored after commit

	if err := fn(&txDispatchRepository{tx: tx}, &txTripUpdater{tx: tx}); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return domainerrors.Internal("dispatch: commit transaction").WithMeta("error", err.Error())
	}
	return nil
}

// ─── tx-scoped DispatchJobRepository ─────────────────────────────────────────

type txDispatchRepository struct {
	tx pgx.Tx
}

var _ repository.DispatchJobRepository = (*txDispatchRepository)(nil)

func (r *txDispatchRepository) Save(ctx context.Context, job *entity.DispatchJob) error {
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

	_, err := r.tx.Exec(ctx, q,
		job.JobID, job.TripID, job.RiderID,
		job.PickupLat, job.PickupLon,
		string(job.Status), job.CurrentDriverID, job.AssignedDriverID,
		job.OfferedDriverIDsCSV(), offerExpiresAt,
		job.OfferTimeoutSec, job.MaxAttempts, job.AttemptCount,
		job.CreatedAt.UTC(), job.UpdatedAt.UTC(),
	)
	if err != nil {
		return domainerrors.Internal("dispatch: tx save job").WithMeta("error", err.Error())
	}
	return nil
}

func (r *txDispatchRepository) FindByID(ctx context.Context, jobID string) (*entity.DispatchJob, error) {
	const q = `
		SELECT job_id, trip_id, rider_id, pickup_lat, pickup_lon,
		       status, current_driver_id, assigned_driver_id, offered_driver_ids,
		       offer_expires_at, offer_timeout_sec, max_attempts, attempt_count,
		       created_at, updated_at
		FROM dispatch_jobs WHERE job_id = $1`
	return scanDispatchJob(r.tx.QueryRow(ctx, q, jobID))
}

func (r *txDispatchRepository) FindByTripID(ctx context.Context, tripID string) (*entity.DispatchJob, error) {
	const q = `
		SELECT job_id, trip_id, rider_id, pickup_lat, pickup_lon,
		       status, current_driver_id, assigned_driver_id, offered_driver_ids,
		       offer_expires_at, offer_timeout_sec, max_attempts, attempt_count,
		       created_at, updated_at
		FROM dispatch_jobs WHERE trip_id = $1`
	return scanDispatchJob(r.tx.QueryRow(ctx, q, tripID))
}

func (r *txDispatchRepository) FindCurrentOfferForDriver(ctx context.Context, driverID string) (*entity.DispatchJob, error) {
	const q = `
		SELECT job_id, trip_id, rider_id, pickup_lat, pickup_lon,
		       status, current_driver_id, assigned_driver_id, offered_driver_ids,
		       offer_expires_at, offer_timeout_sec, max_attempts, attempt_count,
		       created_at, updated_at
		FROM dispatch_jobs
		WHERE current_driver_id = $1
		  AND status = 'searching'
		  AND offer_expires_at > NOW()
		LIMIT 1`
	return scanDispatchJob(r.tx.QueryRow(ctx, q, driverID))
}

func (r *txDispatchRepository) FindExpiredOffers(ctx context.Context, now time.Time) ([]*entity.DispatchJob, error) {
	const q = `
		SELECT job_id, trip_id, rider_id, pickup_lat, pickup_lon,
		       status, current_driver_id, assigned_driver_id, offered_driver_ids,
		       offer_expires_at, offer_timeout_sec, max_attempts, attempt_count,
		       created_at, updated_at
		FROM dispatch_jobs
		WHERE status = 'searching' AND offer_expires_at < $1`

	rows, err := r.tx.Query(ctx, q, now.UTC())
	if err != nil {
		return nil, domainerrors.Internal("dispatch: tx find expired offers").WithMeta("error", err.Error())
	}
	defer rows.Close()

	var jobs []*entity.DispatchJob
	for rows.Next() {
		job, err := scanDispatchJob(rows)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, job)
	}
	if err := rows.Err(); err != nil {
		return nil, domainerrors.Internal("dispatch: tx rows iteration").WithMeta("error", err.Error())
	}
	return jobs, nil
}

// ─── tx-scoped TripUpdater ────────────────────────────────────────────────────

type txTripUpdater struct {
	tx pgx.Tx
}

var _ repository.TripUpdater = (*txTripUpdater)(nil)

func (u *txTripUpdater) SetSearching(ctx context.Context, tripID string, now time.Time) error {
	const q = `UPDATE trips SET status = 'searching', updated_at = $1 WHERE trip_id = $2`
	_, err := u.tx.Exec(ctx, q, now.UTC(), tripID)
	if err != nil {
		return domainerrors.Internal("dispatch: tx set trip searching").WithMeta("error", err.Error())
	}
	return nil
}

func (u *txTripUpdater) AssignDriver(ctx context.Context, tripID, driverID string, now time.Time) error {
	const q = `UPDATE trips SET status = 'driver_assigned', driver_id = $1, updated_at = $2 WHERE trip_id = $3`
	_, err := u.tx.Exec(ctx, q, driverID, now.UTC(), tripID)
	if err != nil {
		return domainerrors.Internal("dispatch: tx assign driver to trip").WithMeta("error", err.Error())
	}
	return nil
}

