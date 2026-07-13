package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	domainerrors "github.com/fairride/shared/errors"
	"github.com/fairride/wallet/domain/entity"
	"github.com/fairride/wallet/domain/repository"
)

const transactionFields = `transaction_id, type, reference_id, payment_method, currency, description, created_at`

// TransactionRepository is the PostgreSQL implementation of repository.TransactionRepository.
type TransactionRepository struct {
	pool *pgxpool.Pool
}

var _ repository.TransactionRepository = (*TransactionRepository)(nil)

func NewTransactionRepository(pool *pgxpool.Pool) *TransactionRepository {
	return &TransactionRepository{pool: pool}
}

func (r *TransactionRepository) Save(ctx context.Context, tx *entity.Transaction) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO wallet_transactions (`+transactionFields+`)
		VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		tx.TransactionID, string(tx.Type), tx.ReferenceID, tx.PaymentMethod, tx.Currency, tx.Description, tx.CreatedAt.UTC(),
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domainerrors.AlreadyExists("transaction already exists")
		}
		return domainerrors.Internal("transaction: save failed").WithMeta("error", err.Error())
	}
	return nil
}

func (r *TransactionRepository) FindByID(ctx context.Context, transactionID string) (*entity.Transaction, error) {
	row := r.pool.QueryRow(ctx, `SELECT `+transactionFields+` FROM wallet_transactions WHERE transaction_id = $1`, transactionID)
	return scanTransaction(row)
}

func (r *TransactionRepository) FindByReferenceID(ctx context.Context, referenceID string) ([]*entity.Transaction, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT `+transactionFields+` FROM wallet_transactions
		WHERE reference_id = $1
		ORDER BY created_at ASC`, referenceID)
	if err != nil {
		return nil, domainerrors.Internal("transaction: list failed").WithMeta("error", err.Error())
	}
	defer rows.Close()

	out := []*entity.Transaction{}
	for rows.Next() {
		var (
			transactionID, txType, refID, paymentMethod, currency, description string
			createdAt                                                          time.Time
		)
		if err := rows.Scan(&transactionID, &txType, &refID, &paymentMethod, &currency, &description, &createdAt); err != nil {
			return nil, domainerrors.Internal("transaction: scan failed").WithMeta("error", err.Error())
		}
		out = append(out, entity.ReconstituteTransaction(transactionID, entity.TransactionType(txType), refID, paymentMethod, currency, description, createdAt.UTC()))
	}
	if err := rows.Err(); err != nil {
		return nil, domainerrors.Internal("transaction: rows error").WithMeta("error", err.Error())
	}
	return out, nil
}

// FindByIDs fetches every transaction in transactionIDs with one query
// (ANY($1)) instead of GetWalletSummaryUseCase's previous per-ID FindByID
// loop — fixes the N+1 pattern flagged for the wallet projection at scale.
func (r *TransactionRepository) FindByIDs(ctx context.Context, transactionIDs []string) (map[string]*entity.Transaction, error) {
	out := map[string]*entity.Transaction{}
	if len(transactionIDs) == 0 {
		return out, nil
	}
	rows, err := r.pool.Query(ctx, `
		SELECT `+transactionFields+` FROM wallet_transactions
		WHERE transaction_id = ANY($1)`, transactionIDs)
	if err != nil {
		return nil, domainerrors.Internal("transaction: batch find failed").WithMeta("error", err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		var (
			transactionID, txType, refID, paymentMethod, currency, description string
			createdAt                                                          time.Time
		)
		if err := rows.Scan(&transactionID, &txType, &refID, &paymentMethod, &currency, &description, &createdAt); err != nil {
			return nil, domainerrors.Internal("transaction: scan failed").WithMeta("error", err.Error())
		}
		out[transactionID] = entity.ReconstituteTransaction(transactionID, entity.TransactionType(txType), refID, paymentMethod, currency, description, createdAt.UTC())
	}
	if err := rows.Err(); err != nil {
		return nil, domainerrors.Internal("transaction: rows error").WithMeta("error", err.Error())
	}
	return out, nil
}

func scanTransaction(row rowScanner) (*entity.Transaction, error) {
	var (
		transactionID, txType, refID, paymentMethod, currency, description string
		createdAt                                                          time.Time
	)
	err := row.Scan(&transactionID, &txType, &refID, &paymentMethod, &currency, &description, &createdAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domainerrors.NotFound("transaction not found")
		}
		return nil, domainerrors.Internal("transaction: scan failed").WithMeta("error", err.Error())
	}
	return entity.ReconstituteTransaction(transactionID, entity.TransactionType(txType), refID, paymentMethod, currency, description, createdAt.UTC()), nil
}
