package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/fairride/identity/domain/entity"
	domainerrors "github.com/fairride/shared/errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	otpFields = `id, phone_number, code_hash, purpose, expires_at, attempts, consumed, created_at`
	otpTable  = `identity_otp_challenges`
	otpSelect = `SELECT ` + otpFields + ` FROM ` + otpTable
)

// OTPRepository is the PostgreSQL implementation of repository.OTPRepository.
type OTPRepository struct {
	pool *pgxpool.Pool
}

// NewOTPRepository constructs an OTPRepository backed by pool.
func NewOTPRepository(pool *pgxpool.Pool) *OTPRepository {
	return &OTPRepository{pool: pool}
}

// Save upserts a challenge. Rows are matched by ID.
func (r *OTPRepository) Save(ctx context.Context, c *entity.OTPChallenge) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO `+otpTable+`
			(id, phone_number, code_hash, purpose, expires_at, attempts, consumed, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (id) DO UPDATE SET
			attempts = EXCLUDED.attempts,
			consumed = EXCLUDED.consumed
	`, c.ID, c.PhoneNumber, c.CodeHash, c.Purpose, c.ExpiresAt, c.Attempts, c.Consumed, c.CreatedAt)
	if err != nil {
		return domainerrors.Internal("save otp challenge").WithMeta("error", err.Error())
	}
	return nil
}

// FindLatestByPhone returns the most recently created challenge for phoneNumber.
func (r *OTPRepository) FindLatestByPhone(ctx context.Context, phoneNumber string) (*entity.OTPChallenge, error) {
	var id, phone, codeHash, purpose string
	var expiresAt, createdAt time.Time
	var attempts int
	var consumed bool

	err := r.pool.QueryRow(ctx,
		otpSelect+` WHERE phone_number = $1 ORDER BY created_at DESC LIMIT 1`, phoneNumber,
	).Scan(&id, &phone, &codeHash, &purpose, &expiresAt, &attempts, &consumed, &createdAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domainerrors.NotFound("no otp challenge found for phone")
		}
		return nil, domainerrors.Internal("query otp challenge").WithMeta("error", err.Error())
	}
	return entity.ReconstituteOTPChallenge(id, phone, codeHash, purpose, expiresAt, attempts, consumed, createdAt), nil
}
