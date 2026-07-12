package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/fairride/notification/domain/entity"
	"github.com/fairride/notification/domain/repository"
	domainerrors "github.com/fairride/shared/errors"
)

// CallSessionRepository is the PostgreSQL implementation of repository.CallSessionRepository.
type CallSessionRepository struct {
	pool *pgxpool.Pool
}

var _ repository.CallSessionRepository = (*CallSessionRepository)(nil)

func NewCallSessionRepository(pool *pgxpool.Pool) *CallSessionRepository {
	return &CallSessionRepository{pool: pool}
}

func (r *CallSessionRepository) Save(ctx context.Context, cs *entity.CallSession) error {
	const q = `
		INSERT INTO call_sessions (id, trip_id, caller_id, callee_id, created_at)
		VALUES ($1,$2,$3,$4,$5)`
	_, err := r.pool.Exec(ctx, q, cs.ID, cs.TripID, cs.CallerID, cs.CalleeID, cs.CreatedAt.UTC())
	if err != nil {
		return domainerrors.Internal("call_session: save failed").WithMeta("error", err.Error())
	}
	return nil
}
