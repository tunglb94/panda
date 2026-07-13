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

const settlementFields = `
	settlement_id, trip_id, driver_id, trip_type, payment_method, fare_amount_cents,
	commission_rate, commission_amount_cents, driver_income_cents,
	promotion_subsidy_cents, voucher_cost_cents, currency, transaction_id, created_at, status,
	voucher_status`

// SettlementRepository is the PostgreSQL implementation of
// repository.SettlementRepository — insert-only, no Update/Delete method
// exists (Phần 2/13 — Settlement is immutable).
type SettlementRepository struct {
	pool *pgxpool.Pool
}

var _ repository.SettlementRepository = (*SettlementRepository)(nil)

func NewSettlementRepository(pool *pgxpool.Pool) *SettlementRepository {
	return &SettlementRepository{pool: pool}
}

func (r *SettlementRepository) Save(ctx context.Context, s *entity.Settlement) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO settlements (`+settlementFields+`)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)`,
		s.SettlementID, s.TripID, s.DriverID, string(s.TripType), string(s.PaymentMethod), s.FareAmountCents,
		s.CommissionRate, s.CommissionAmountCents, s.DriverIncomeCents,
		s.PromotionSubsidyCents, s.VoucherCostCents, s.Currency, s.TransactionID, s.CreatedAt.UTC(),
		string(s.Status), string(s.VoucherStatus),
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domainerrors.AlreadyExists("settlement already exists for this trip")
		}
		return domainerrors.Internal("settlement: save failed").WithMeta("error", err.Error())
	}
	return nil
}

// MarkPosted is the one narrow exception to "insert-only, no Update" — see
// repository.SettlementRepository.MarkPosted.
func (r *SettlementRepository) MarkPosted(ctx context.Context, settlementID string) error {
	_, err := r.pool.Exec(ctx, `UPDATE settlements SET status = $1 WHERE settlement_id = $2`,
		string(entity.SettlementStatusPosted), settlementID)
	if err != nil {
		return domainerrors.Internal("settlement: mark posted failed").WithMeta("error", err.Error())
	}
	return nil
}

func (r *SettlementRepository) FindByTripID(ctx context.Context, tripID string) (*entity.Settlement, error) {
	row := r.pool.QueryRow(ctx, `SELECT `+settlementFields+` FROM settlements WHERE trip_id = $1`, tripID)
	return scanSettlement(row)
}

func (r *SettlementRepository) FindByID(ctx context.Context, settlementID string) (*entity.Settlement, error) {
	row := r.pool.QueryRow(ctx, `SELECT `+settlementFields+` FROM settlements WHERE settlement_id = $1`, settlementID)
	return scanSettlement(row)
}

func (r *SettlementRepository) ListByDriverID(ctx context.Context, driverID string, from, to int64, limit int) ([]*entity.Settlement, error) {
	return r.list(ctx, driverID, from, to, limit)
}

func (r *SettlementRepository) ListAll(ctx context.Context, driverID string, from, to int64, limit int) ([]*entity.Settlement, error) {
	return r.list(ctx, driverID, from, to, limit)
}

func (r *SettlementRepository) list(ctx context.Context, driverID string, from, to int64, limit int) ([]*entity.Settlement, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	q := `SELECT ` + settlementFields + ` FROM settlements WHERE 1=1`
	args := []any{}
	if driverID != "" {
		args = append(args, driverID)
		q += ` AND driver_id = $` + strconv.Itoa(len(args))
	}
	if from > 0 {
		args = append(args, time.Unix(from, 0).UTC())
		q += ` AND created_at >= $` + strconv.Itoa(len(args))
	}
	if to > 0 {
		args = append(args, time.Unix(to, 0).UTC())
		q += ` AND created_at < $` + strconv.Itoa(len(args))
	}
	args = append(args, limit)
	q += ` ORDER BY created_at DESC LIMIT $` + strconv.Itoa(len(args))

	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, domainerrors.Internal("settlement: list failed").WithMeta("error", err.Error())
	}
	defer rows.Close()

	out := []*entity.Settlement{}
	for rows.Next() {
		s, err := scanSettlementRow(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	if err := rows.Err(); err != nil {
		return nil, domainerrors.Internal("settlement: rows error").WithMeta("error", err.Error())
	}
	return out, nil
}

func scanSettlement(row rowScanner) (*entity.Settlement, error) {
	s, err := scanSettlementRow(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domainerrors.NotFound("settlement not found")
		}
		return nil, err
	}
	return s, nil
}

func scanSettlementRow(row rowScanner) (*entity.Settlement, error) {
	var (
		settlementID, tripID, driverID, tripType, paymentMethod string
		fareAmountCents                                         int64
		commissionRate                                          float64
		commissionAmountCents, driverIncomeCents                int64
		promotionSubsidyCents, voucherCostCents                 int64
		currency, transactionID                                 string
		createdAt                                               time.Time
		status, voucherStatus                                   string
	)
	err := row.Scan(
		&settlementID, &tripID, &driverID, &tripType, &paymentMethod, &fareAmountCents,
		&commissionRate, &commissionAmountCents, &driverIncomeCents,
		&promotionSubsidyCents, &voucherCostCents, &currency, &transactionID, &createdAt,
		&status, &voucherStatus,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pgx.ErrNoRows
		}
		return nil, domainerrors.Internal("settlement: scan failed").WithMeta("error", err.Error())
	}
	return entity.ReconstituteSettlement(
		settlementID, tripID, driverID, entity.TripType(tripType), entity.PaymentMethod(paymentMethod),
		fareAmountCents, commissionRate, commissionAmountCents, driverIncomeCents,
		promotionSubsidyCents, voucherCostCents, currency, transactionID, createdAt.UTC(),
		entity.SettlementStatus(status), entity.VoucherStatus(voucherStatus),
	), nil
}
