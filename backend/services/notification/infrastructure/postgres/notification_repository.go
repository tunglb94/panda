package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/fairride/notification/domain/entity"
	"github.com/fairride/notification/domain/repository"
	domainerrors "github.com/fairride/shared/errors"
)

// NotificationRepository is the PostgreSQL implementation of repository.NotificationRepository.
type NotificationRepository struct {
	pool *pgxpool.Pool
}

var _ repository.NotificationRepository = (*NotificationRepository)(nil)

func NewNotificationRepository(pool *pgxpool.Pool) *NotificationRepository {
	return &NotificationRepository{pool: pool}
}

func (r *NotificationRepository) Save(ctx context.Context, n *entity.Notification) error {
	const q = `
		INSERT INTO notifications (id, user_id, category, title, body, trip_id, conversation_id, created_at, read_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`
	_, err := r.pool.Exec(ctx, q,
		n.ID, n.UserID, string(n.Category), n.Title, n.Body, n.TripID, n.ConversationID, n.CreatedAt.UTC(), n.ReadAt,
	)
	if err != nil {
		return domainerrors.Internal("notification: save failed").WithMeta("error", err.Error())
	}
	return nil
}

func (r *NotificationRepository) ListByUser(ctx context.Context, userID string, limit int) ([]*entity.Notification, error) {
	const q = `
		SELECT id, user_id, category, title, body, trip_id, conversation_id, created_at, read_at
		FROM notifications
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2`
	rows, err := r.pool.Query(ctx, q, userID, limit)
	if err != nil {
		return nil, domainerrors.Internal("notification: list failed").WithMeta("error", err.Error())
	}
	defer rows.Close()

	var out []*entity.Notification
	for rows.Next() {
		var (
			id, uid, category, title, body, tripID, convID string
			createdAt                                      time.Time
			readAt                                         *time.Time
		)
		if err := rows.Scan(&id, &uid, &category, &title, &body, &tripID, &convID, &createdAt, &readAt); err != nil {
			return nil, domainerrors.Internal("notification: scan failed").WithMeta("error", err.Error())
		}
		out = append(out, entity.ReconstituteNotification(id, uid, entity.Category(category), title, body, tripID, convID, createdAt.UTC(), readAt))
	}
	if err := rows.Err(); err != nil {
		return nil, domainerrors.Internal("notification: rows error").WithMeta("error", err.Error())
	}
	return out, nil
}

func (r *NotificationRepository) MarkRead(ctx context.Context, id, userID string, now time.Time) error {
	const q = `UPDATE notifications SET read_at = $3 WHERE id = $1 AND user_id = $2 AND read_at IS NULL`
	_, err := r.pool.Exec(ctx, q, id, userID, now.UTC())
	if err != nil {
		return domainerrors.Internal("notification: mark read failed").WithMeta("error", err.Error())
	}
	return nil
}

func (r *NotificationRepository) CountUnread(ctx context.Context, userID string) (int, error) {
	const q = `SELECT COUNT(*) FROM notifications WHERE user_id = $1 AND read_at IS NULL`
	var count int
	if err := r.pool.QueryRow(ctx, q, userID).Scan(&count); err != nil {
		return 0, domainerrors.Internal("notification: count unread failed").WithMeta("error", err.Error())
	}
	return count, nil
}
