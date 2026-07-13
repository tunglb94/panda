package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	domainerrors "github.com/fairride/shared/errors"
	"github.com/fairride/wallet/domain/entity"
	"github.com/fairride/wallet/domain/repository"
)

const bankAccountFields = `bank_account_id, driver_id, bank_name, account_holder_name, account_number, branch_name, created_at, updated_at`

// BankAccountRepository is the PostgreSQL implementation of repository.BankAccountRepository.
type BankAccountRepository struct {
	pool *pgxpool.Pool
}

var _ repository.BankAccountRepository = (*BankAccountRepository)(nil)

func NewBankAccountRepository(pool *pgxpool.Pool) *BankAccountRepository {
	return &BankAccountRepository{pool: pool}
}

// Save upserts by driver_id — Phần 6's "chỉ 1 tài khoản mặc định" means
// adding a new bank account always replaces the existing one.
func (r *BankAccountRepository) Save(ctx context.Context, b *entity.BankAccount) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO bank_accounts (`+bankAccountFields+`)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		ON CONFLICT (driver_id) DO UPDATE SET
			bank_account_id     = EXCLUDED.bank_account_id,
			bank_name           = EXCLUDED.bank_name,
			account_holder_name = EXCLUDED.account_holder_name,
			account_number      = EXCLUDED.account_number,
			branch_name         = EXCLUDED.branch_name,
			updated_at          = EXCLUDED.updated_at`,
		b.BankAccountID, b.DriverID, b.BankName, b.AccountHolderName, b.AccountNumber, b.BranchName, b.CreatedAt.UTC(), b.UpdatedAt.UTC(),
	)
	if err != nil {
		return domainerrors.Internal("bank_account: save failed").WithMeta("error", err.Error())
	}
	return nil
}

func (r *BankAccountRepository) FindByDriverID(ctx context.Context, driverID string) (*entity.BankAccount, error) {
	row := r.pool.QueryRow(ctx, `SELECT `+bankAccountFields+` FROM bank_accounts WHERE driver_id = $1`, driverID)
	var (
		bankAccountID, dID, bankName, holderName, accountNumber, branchName string
		createdAt, updatedAt                                                time.Time
	)
	err := row.Scan(&bankAccountID, &dID, &bankName, &holderName, &accountNumber, &branchName, &createdAt, &updatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domainerrors.NotFound("bank account not found")
		}
		return nil, domainerrors.Internal("bank_account: scan failed").WithMeta("error", err.Error())
	}
	return entity.ReconstituteBankAccount(bankAccountID, dID, bankName, holderName, accountNumber, branchName, createdAt.UTC(), updatedAt.UTC()), nil
}
