package repository

import (
	"context"

	"github.com/fairride/wallet/domain/entity"
)

// LedgerEntryRepository persists and retrieves immutable LedgerEntry records.
// Entries are never updated or deleted.
type LedgerEntryRepository interface {
	// Save appends a new ledger entry. Never updates an existing one.
	// Returns CodeAlreadyExists if an entry with the same EntryID already exists.
	Save(ctx context.Context, entry *entity.LedgerEntry) error

	// FindByWalletID returns all ledger entries for a wallet, ordered by created_at ASC.
	// Returns an empty slice (not an error) when none exist.
	FindByWalletID(ctx context.Context, walletID string) ([]entity.LedgerEntry, error)

	// FindByTransactionID returns all ledger entries that belong to a transaction.
	// Returns an empty slice (not an error) when none exist.
	FindByTransactionID(ctx context.Context, transactionID string) ([]entity.LedgerEntry, error)

	// ListOutstandingDrivers aggregates every driver wallet's net
	// TypeCommission entries (debit - credit) and returns those still
	// greater than zero (Phần 10 — admin "Outstanding Drivers" view).
	// Reads only from the ledger (Phần 13 — "Wallet không được tự tính.
	// Mọi số liệu phải đọc từ Ledger."), never from Settlement directly.
	ListOutstandingDrivers(ctx context.Context, limit int) ([]OutstandingDriver, error)
}

// OutstandingDriver is one row of the admin Outstanding Drivers report.
type OutstandingDriver struct {
	DriverID         string
	OutstandingCents int64
	Currency         string
}
