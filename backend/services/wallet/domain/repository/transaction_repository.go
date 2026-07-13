package repository

import (
	"context"

	"github.com/fairride/wallet/domain/entity"
)

// TransactionRepository persists and retrieves Transaction records.
// Transactions are immutable after creation.
type TransactionRepository interface {
	// Save appends a new transaction.
	// Returns CodeAlreadyExists if a transaction with the same TransactionID already exists.
	Save(ctx context.Context, tx *entity.Transaction) error

	// FindByID returns the transaction with the given ID.
	// Returns CodeNotFound if no transaction has that ID.
	FindByID(ctx context.Context, transactionID string) (*entity.Transaction, error)

	// FindByReferenceID returns all transactions that reference a given entity (e.g. a trip).
	// Returns an empty slice (not an error) when none exist.
	FindByReferenceID(ctx context.Context, referenceID string) ([]*entity.Transaction, error)

	// FindByIDs returns every transaction in transactionIDs in a single
	// round-trip, keyed by TransactionID. Missing IDs are simply absent from
	// the returned map (not an error) — GetWalletSummaryUseCase's ledger
	// entries can then treat them as orphaned exactly as FindByID's
	// best-effort loop always has.
	FindByIDs(ctx context.Context, transactionIDs []string) (map[string]*entity.Transaction, error)
}
