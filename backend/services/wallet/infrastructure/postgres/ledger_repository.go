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

const ledgerEntryFields = `entry_id, wallet_id, transaction_id, direction, amount_cents, currency, description, created_at`

// LedgerEntryRepository is the PostgreSQL implementation of
// repository.LedgerEntryRepository — insert-only, no Update/Delete method
// exists (Phần 1 — "Ledger là immutable. Không UPDATE. Không DELETE.").
type LedgerEntryRepository struct {
	pool *pgxpool.Pool
}

var _ repository.LedgerEntryRepository = (*LedgerEntryRepository)(nil)

func NewLedgerEntryRepository(pool *pgxpool.Pool) *LedgerEntryRepository {
	return &LedgerEntryRepository{pool: pool}
}

func (r *LedgerEntryRepository) Save(ctx context.Context, e *entity.LedgerEntry) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO wallet_ledger_entries (`+ledgerEntryFields+`)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		e.EntryID, e.WalletID, e.TransactionID, string(e.Direction), e.AmountCents, e.Currency, e.Description, e.CreatedAt.UTC(),
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domainerrors.AlreadyExists("ledger entry already exists")
		}
		return domainerrors.Internal("ledger_entry: save failed").WithMeta("error", err.Error())
	}
	return nil
}

func (r *LedgerEntryRepository) FindByWalletID(ctx context.Context, walletID string) ([]entity.LedgerEntry, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT `+ledgerEntryFields+` FROM wallet_ledger_entries
		WHERE wallet_id = $1
		ORDER BY created_at ASC`, walletID)
	if err != nil {
		return nil, domainerrors.Internal("ledger_entry: list failed").WithMeta("error", err.Error())
	}
	defer rows.Close()
	return scanLedgerEntries(rows)
}

func (r *LedgerEntryRepository) FindByTransactionID(ctx context.Context, transactionID string) ([]entity.LedgerEntry, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT `+ledgerEntryFields+` FROM wallet_ledger_entries
		WHERE transaction_id = $1
		ORDER BY created_at ASC`, transactionID)
	if err != nil {
		return nil, domainerrors.Internal("ledger_entry: list failed").WithMeta("error", err.Error())
	}
	defer rows.Close()
	return scanLedgerEntries(rows)
}

// ListOutstandingDrivers aggregates each driver wallet's net TypeCommission
// entries (debit owed - credit repaid) directly from the ledger (Phần 13),
// joining wallets (to resolve owner_id/currency for wallet_type='driver')
// and wallet_transactions (to filter type='commission').
func (r *LedgerEntryRepository) ListOutstandingDrivers(ctx context.Context, limit int) ([]repository.OutstandingDriver, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	rows, err := r.pool.Query(ctx, `
		SELECT w.owner_id, w.currency,
			SUM(CASE WHEN e.direction = 'debit' THEN e.amount_cents ELSE -e.amount_cents END) AS outstanding
		FROM wallet_ledger_entries e
		JOIN wallets w ON w.wallet_id = e.wallet_id
		JOIN wallet_transactions t ON t.transaction_id = e.transaction_id
		WHERE w.wallet_type = 'driver' AND t.type = 'commission'
		GROUP BY w.owner_id, w.currency
		HAVING SUM(CASE WHEN e.direction = 'debit' THEN e.amount_cents ELSE -e.amount_cents END) > 0
		ORDER BY outstanding DESC
		LIMIT $1`, limit)
	if err != nil {
		return nil, domainerrors.Internal("ledger_entry: outstanding drivers query failed").WithMeta("error", err.Error())
	}
	defer rows.Close()

	out := []repository.OutstandingDriver{}
	for rows.Next() {
		var d repository.OutstandingDriver
		if err := rows.Scan(&d.DriverID, &d.Currency, &d.OutstandingCents); err != nil {
			return nil, domainerrors.Internal("ledger_entry: outstanding drivers scan failed").WithMeta("error", err.Error())
		}
		out = append(out, d)
	}
	if err := rows.Err(); err != nil {
		return nil, domainerrors.Internal("ledger_entry: outstanding drivers rows error").WithMeta("error", err.Error())
	}
	return out, nil
}

func scanLedgerEntries(rows pgx.Rows) ([]entity.LedgerEntry, error) {
	out := []entity.LedgerEntry{}
	for rows.Next() {
		var (
			entryID, walletID, transactionID, direction, currency, description string
			amountCents                                                        int64
			createdAt                                                          time.Time
		)
		if err := rows.Scan(&entryID, &walletID, &transactionID, &direction, &amountCents, &currency, &description, &createdAt); err != nil {
			return nil, domainerrors.Internal("ledger_entry: scan failed").WithMeta("error", err.Error())
		}
		out = append(out, *entity.ReconstituteLedgerEntry(entryID, walletID, transactionID, entity.EntryDirection(direction), amountCents, currency, description, createdAt.UTC()))
	}
	if err := rows.Err(); err != nil {
		return nil, domainerrors.Internal("ledger_entry: rows error").WithMeta("error", err.Error())
	}
	return out, nil
}
