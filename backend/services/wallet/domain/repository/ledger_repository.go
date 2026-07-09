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
}
