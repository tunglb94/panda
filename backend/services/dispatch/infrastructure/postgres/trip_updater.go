package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/fairride/dispatch/domain/repository"
	domainerrors "github.com/fairride/shared/errors"
)

// TripUpdater directly updates trip status in the shared trips table.
// Both dispatch and trip services share one PostgreSQL database in the MVP.
type TripUpdater struct {
	pool *pgxpool.Pool
}

var _ repository.TripUpdater = (*TripUpdater)(nil)

func NewTripUpdater(pool *pgxpool.Pool) *TripUpdater {
	return &TripUpdater{pool: pool}
}

// SetSearching transitions the trip to the Searching status.
func (u *TripUpdater) SetSearching(ctx context.Context, tripID string, now time.Time) error {
	const q = `UPDATE trips SET status = 'searching', updated_at = $1 WHERE trip_id = $2`
	_, err := u.pool.Exec(ctx, q, now.UTC(), tripID)
	if err != nil {
		return domainerrors.Internal("dispatch: set trip searching").WithMeta("error", err.Error())
	}
	return nil
}

// AssignDriver transitions the trip to DriverAssigned status and sets the driver.
func (u *TripUpdater) AssignDriver(ctx context.Context, tripID, driverID string, now time.Time) error {
	const q = `UPDATE trips SET status = 'driver_assigned', driver_id = $1, updated_at = $2 WHERE trip_id = $3`
	_, err := u.pool.Exec(ctx, q, driverID, now.UTC(), tripID)
	if err != nil {
		return domainerrors.Internal("dispatch: assign driver to trip").WithMeta("error", err.Error())
	}
	return nil
}
