package postgres

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	domainerrors "github.com/fairride/shared/errors"
	"github.com/fairride/wallet/domain/entity"
	"github.com/fairride/wallet/domain/repository"
)

const payoutRequestFields = `
	payout_request_id, driver_id, amount_cents, currency, bank_account_id, bank_name, account_number_masked,
	status, requested_at, reviewed_at, reviewed_by, reject_reason, paid_at, transaction_id, created_at, updated_at`

// PayoutRequestRepository is the PostgreSQL implementation of repository.PayoutRequestRepository.
type PayoutRequestRepository struct {
	pool *pgxpool.Pool
}

var _ repository.PayoutRequestRepository = (*PayoutRequestRepository)(nil)

func NewPayoutRequestRepository(pool *pgxpool.Pool) *PayoutRequestRepository {
	return &PayoutRequestRepository{pool: pool}
}

func (r *PayoutRequestRepository) Save(ctx context.Context, p *entity.PayoutRequest) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO payout_requests (`+payoutRequestFields+`)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)
		ON CONFLICT (payout_request_id) DO UPDATE SET
			status        = EXCLUDED.status,
			reviewed_at   = EXCLUDED.reviewed_at,
			reviewed_by   = EXCLUDED.reviewed_by,
			reject_reason = EXCLUDED.reject_reason,
			paid_at       = EXCLUDED.paid_at,
			transaction_id = EXCLUDED.transaction_id,
			updated_at    = EXCLUDED.updated_at`,
		p.PayoutRequestID, p.DriverID, p.AmountCents, p.Currency, p.BankAccountID, p.BankName, p.AccountNumberMasked,
		string(p.Status), p.RequestedAt.UTC(), p.ReviewedAt, p.ReviewedBy, p.RejectReason, p.PaidAt, p.TransactionID,
		p.CreatedAt.UTC(), p.UpdatedAt.UTC(),
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domainerrors.AlreadyExists("payout request already exists")
		}
		return domainerrors.Internal("payout_request: save failed").WithMeta("error", err.Error())
	}
	return nil
}

func (r *PayoutRequestRepository) FindByID(ctx context.Context, id string) (*entity.PayoutRequest, error) {
	row := r.pool.QueryRow(ctx, `SELECT `+payoutRequestFields+` FROM payout_requests WHERE payout_request_id = $1`, id)
	return scanPayoutRequest(row)
}

func (r *PayoutRequestRepository) FindInFlightByDriverID(ctx context.Context, driverID string) (*entity.PayoutRequest, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT `+payoutRequestFields+` FROM payout_requests
		WHERE driver_id = $1 AND status IN ('pending','approved')
		ORDER BY created_at DESC LIMIT 1`, driverID)
	return scanPayoutRequest(row)
}

func (r *PayoutRequestRepository) ListByDriverID(ctx context.Context, driverID string, limit int) ([]*entity.PayoutRequest, error) {
	return r.list(ctx, "", driverID, limit)
}

func (r *PayoutRequestRepository) ListByFilter(ctx context.Context, status entity.PayoutStatus, driverID string, limit int) ([]*entity.PayoutRequest, error) {
	return r.list(ctx, string(status), driverID, limit)
}

func (r *PayoutRequestRepository) list(ctx context.Context, status, driverID string, limit int) ([]*entity.PayoutRequest, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	q := `SELECT ` + payoutRequestFields + ` FROM payout_requests WHERE 1=1`
	args := []any{}
	if status != "" {
		args = append(args, status)
		q += ` AND status = $` + strconv.Itoa(len(args))
	}
	if driverID != "" {
		args = append(args, driverID)
		q += ` AND driver_id = $` + strconv.Itoa(len(args))
	}
	args = append(args, limit)
	q += ` ORDER BY created_at DESC LIMIT $` + strconv.Itoa(len(args))

	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, domainerrors.Internal("payout_request: list failed").WithMeta("error", err.Error())
	}
	defer rows.Close()

	out := []*entity.PayoutRequest{}
	for rows.Next() {
		p, err := scanPayoutRequestRow(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	if err := rows.Err(); err != nil {
		return nil, domainerrors.Internal("payout_request: rows error").WithMeta("error", err.Error())
	}
	return out, nil
}

func scanPayoutRequest(row rowScanner) (*entity.PayoutRequest, error) {
	p, err := scanPayoutRequestRow(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domainerrors.NotFound("payout request not found")
		}
		return nil, err
	}
	return p, nil
}

func scanPayoutRequestRow(row rowScanner) (*entity.PayoutRequest, error) {
	var (
		id, driverID, currency, bankAccountID, bankName, accountNumberMasked string
		amountCents                                                          int64
		status                                                               string
		requestedAt                                                          time.Time
		reviewedAt                                                           *time.Time
		reviewedBy, rejectReason                                             string
		paidAt                                                               *time.Time
		transactionID                                                        string
		createdAt, updatedAt                                                 time.Time
	)
	err := row.Scan(
		&id, &driverID, &amountCents, &currency, &bankAccountID, &bankName, &accountNumberMasked,
		&status, &requestedAt, &reviewedAt, &reviewedBy, &rejectReason, &paidAt, &transactionID,
		&createdAt, &updatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pgx.ErrNoRows
		}
		return nil, domainerrors.Internal("payout_request: scan failed").WithMeta("error", err.Error())
	}
	return entity.ReconstitutePayoutRequest(
		id, driverID, amountCents, currency, bankAccountID, bankName, accountNumberMasked,
		entity.PayoutStatus(status), requestedAt.UTC(), reviewedAt, reviewedBy, rejectReason, paidAt, transactionID,
		createdAt.UTC(), updatedAt.UTC(),
	), nil
}
