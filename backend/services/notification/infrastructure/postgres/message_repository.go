package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/fairride/notification/domain/entity"
	"github.com/fairride/notification/domain/repository"
	domainerrors "github.com/fairride/shared/errors"
)

// MessageRepository is the PostgreSQL implementation of repository.MessageRepository.
type MessageRepository struct {
	pool *pgxpool.Pool
}

var _ repository.MessageRepository = (*MessageRepository)(nil)

func NewMessageRepository(pool *pgxpool.Pool) *MessageRepository {
	return &MessageRepository{pool: pool}
}

func (r *MessageRepository) Save(ctx context.Context, m *entity.Message) error {
	const q = `
		INSERT INTO messages (id, conversation_id, sender_id, sender_role, body, quick_reply_key, created_at, delivered_at, read_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		RETURNING seq`
	err := r.pool.QueryRow(ctx, q,
		m.ID, m.ConversationID, m.SenderID, string(m.SenderRole), m.Body, m.QuickReplyKey,
		m.CreatedAt.UTC(), m.DeliveredAt, m.ReadAt,
	).Scan(&m.Seq)
	if err != nil {
		return domainerrors.Internal("message: save failed").WithMeta("error", err.Error())
	}
	return nil
}

func (r *MessageRepository) ListSince(ctx context.Context, conversationID string, sinceSeq int64, limit int) ([]*entity.Message, error) {
	const q = `
		SELECT seq, id, conversation_id, sender_id, sender_role, body, quick_reply_key, created_at, delivered_at, read_at
		FROM messages
		WHERE conversation_id = $1 AND seq > $2
		ORDER BY seq ASC
		LIMIT $3`
	rows, err := r.pool.Query(ctx, q, conversationID, sinceSeq, limit)
	if err != nil {
		return nil, domainerrors.Internal("message: list failed").WithMeta("error", err.Error())
	}
	defer rows.Close()

	var out []*entity.Message
	for rows.Next() {
		var (
			seq                                         int64
			id, convID, senderID, senderRole, body, qrk string
			createdAt                                   time.Time
			deliveredAt, readAt                         *time.Time
		)
		if err := rows.Scan(&seq, &id, &convID, &senderID, &senderRole, &body, &qrk, &createdAt, &deliveredAt, &readAt); err != nil {
			return nil, domainerrors.Internal("message: scan failed").WithMeta("error", err.Error())
		}
		out = append(out, entity.ReconstituteMessage(id, seq, convID, senderID, entity.SenderRole(senderRole), body, qrk, createdAt.UTC(), deliveredAt, readAt))
	}
	if err := rows.Err(); err != nil {
		return nil, domainerrors.Internal("message: rows error").WithMeta("error", err.Error())
	}
	return out, nil
}

func (r *MessageRepository) MarkReadByRecipient(ctx context.Context, conversationID, recipientID string, now time.Time) error {
	const q = `
		UPDATE messages SET read_at = $3
		WHERE conversation_id = $1 AND sender_id != $2 AND read_at IS NULL`
	_, err := r.pool.Exec(ctx, q, conversationID, recipientID, now.UTC())
	if err != nil {
		return domainerrors.Internal("message: mark read failed").WithMeta("error", err.Error())
	}
	return nil
}

func (r *MessageRepository) CountUnread(ctx context.Context, conversationID, recipientID string) (int, error) {
	const q = `
		SELECT COUNT(*) FROM messages
		WHERE conversation_id = $1 AND sender_id != $2 AND read_at IS NULL`
	var count int
	if err := r.pool.QueryRow(ctx, q, conversationID, recipientID).Scan(&count); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, nil
		}
		return 0, domainerrors.Internal("message: count unread failed").WithMeta("error", err.Error())
	}
	return count, nil
}
