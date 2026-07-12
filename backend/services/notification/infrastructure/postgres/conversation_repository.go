package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/fairride/notification/domain/entity"
	"github.com/fairride/notification/domain/repository"
	domainerrors "github.com/fairride/shared/errors"
)

// ConversationRepository is the PostgreSQL implementation of repository.ConversationRepository.
type ConversationRepository struct {
	pool *pgxpool.Pool
}

var _ repository.ConversationRepository = (*ConversationRepository)(nil)

func NewConversationRepository(pool *pgxpool.Pool) *ConversationRepository {
	return &ConversationRepository{pool: pool}
}

func (r *ConversationRepository) Save(ctx context.Context, c *entity.Conversation) error {
	const q = `
		INSERT INTO conversations (id, trip_id, rider_id, driver_id, trip_type, status, created_at, closed_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`
	_, err := r.pool.Exec(ctx, q, c.ID, c.TripID, c.RiderID, c.DriverID, c.TripType, string(c.Status), c.CreatedAt.UTC(), c.ClosedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domainerrors.AlreadyExists("conversation already exists for this trip")
		}
		return domainerrors.Internal("conversation: save failed").WithMeta("error", err.Error())
	}
	return nil
}

func (r *ConversationRepository) Update(ctx context.Context, c *entity.Conversation) error {
	const q = `UPDATE conversations SET status = $2, closed_at = $3 WHERE id = $1`
	tag, err := r.pool.Exec(ctx, q, c.ID, string(c.Status), c.ClosedAt)
	if err != nil {
		return domainerrors.Internal("conversation: update failed").WithMeta("error", err.Error())
	}
	if tag.RowsAffected() == 0 {
		return domainerrors.NotFound("conversation not found")
	}
	return nil
}

func (r *ConversationRepository) FindByTripID(ctx context.Context, tripID string) (*entity.Conversation, error) {
	const q = `
		SELECT id, trip_id, rider_id, driver_id, trip_type, status, created_at, closed_at
		FROM conversations WHERE trip_id = $1`
	return r.scanOne(r.pool.QueryRow(ctx, q, tripID))
}

func (r *ConversationRepository) FindByID(ctx context.Context, id string) (*entity.Conversation, error) {
	const q = `
		SELECT id, trip_id, rider_id, driver_id, trip_type, status, created_at, closed_at
		FROM conversations WHERE id = $1`
	return r.scanOne(r.pool.QueryRow(ctx, q, id))
}

func (r *ConversationRepository) scanOne(row pgx.Row) (*entity.Conversation, error) {
	var (
		id, tripID, riderID, driverID, tripType, status string
		createdAt                                       time.Time
		closedAt                                        *time.Time
	)
	err := row.Scan(&id, &tripID, &riderID, &driverID, &tripType, &status, &createdAt, &closedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domainerrors.NotFound("conversation not found")
		}
		return nil, domainerrors.Internal("conversation: find failed").WithMeta("error", err.Error())
	}
	return entity.ReconstituteConversation(id, tripID, riderID, driverID, tripType, entity.ConversationStatus(status), createdAt.UTC(), closedAt), nil
}
