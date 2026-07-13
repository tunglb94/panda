package repository

import (
	"context"

	"github.com/fairride/wallet/domain/entity"
)

// BankAccountRepository persists a driver's single default BankAccount (Phần 6).
type BankAccountRepository interface {
	// Save upserts by DriverID (one row per driver — see the migration's
	// UNIQUE(driver_id)).
	Save(ctx context.Context, b *entity.BankAccount) error
	FindByDriverID(ctx context.Context, driverID string) (*entity.BankAccount, error)
}
