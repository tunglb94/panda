// Package postgres is the PostgreSQL implementation of the Wallet domain's
// repository interfaces (Financial Core / Settlement Engine phase — the
// Wallet Foundation phase left this package unwritten, "Phase B3" per the
// original cmd/server/main.go comment).
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

const walletFields = `wallet_id, owner_id, wallet_type, currency, created_at, updated_at`

// WalletRepository is the PostgreSQL implementation of repository.WalletRepository.
type WalletRepository struct {
	pool *pgxpool.Pool
}

var _ repository.WalletRepository = (*WalletRepository)(nil)

func NewWalletRepository(pool *pgxpool.Pool) *WalletRepository {
	return &WalletRepository{pool: pool}
}

func (r *WalletRepository) Save(ctx context.Context, w *entity.Wallet) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO wallets (`+walletFields+`)
		VALUES ($1,$2,$3,$4,$5,$6)
		ON CONFLICT (owner_id) DO UPDATE SET
			updated_at = EXCLUDED.updated_at`,
		w.WalletID, w.OwnerID, string(w.WalletType), w.Currency, w.CreatedAt.UTC(), w.UpdatedAt.UTC(),
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domainerrors.AlreadyExists("wallet already exists for this owner")
		}
		return domainerrors.Internal("wallet: save failed").WithMeta("error", err.Error())
	}
	return nil
}

func (r *WalletRepository) FindByID(ctx context.Context, walletID string) (*entity.Wallet, error) {
	row := r.pool.QueryRow(ctx, `SELECT `+walletFields+` FROM wallets WHERE wallet_id = $1`, walletID)
	return scanWallet(row)
}

func (r *WalletRepository) FindByOwnerID(ctx context.Context, ownerID string) (*entity.Wallet, error) {
	row := r.pool.QueryRow(ctx, `SELECT `+walletFields+` FROM wallets WHERE owner_id = $1`, ownerID)
	return scanWallet(row)
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanWallet(row rowScanner) (*entity.Wallet, error) {
	var (
		walletID, ownerID, walletType, currency string
		createdAt, updatedAt                    time.Time
	)
	err := row.Scan(&walletID, &ownerID, &walletType, &currency, &createdAt, &updatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domainerrors.NotFound("wallet not found")
		}
		return nil, domainerrors.Internal("wallet: scan failed").WithMeta("error", err.Error())
	}
	return entity.ReconstituteWallet(walletID, ownerID, entity.WalletType(walletType), currency, createdAt.UTC(), updatedAt.UTC()), nil
}
