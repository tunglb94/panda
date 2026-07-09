package app

import (
	"context"
	"strings"

	"github.com/fairride/shared/errors"
	"github.com/fairride/wallet/domain/entity"
	"github.com/fairride/wallet/domain/repository"
)

// GetTransactionResult bundles a transaction with its associated ledger entries.
type GetTransactionResult struct {
	Transaction *entity.Transaction
	Entries     []entity.LedgerEntry
}

// GetTransactionUseCase returns a transaction and all its ledger entries.
type GetTransactionUseCase struct {
	transactions repository.TransactionRepository
	ledger       repository.LedgerEntryRepository
}

// NewGetTransactionUseCase wires the use case.
func NewGetTransactionUseCase(
	transactions repository.TransactionRepository,
	ledger repository.LedgerEntryRepository,
) *GetTransactionUseCase {
	return &GetTransactionUseCase{transactions: transactions, ledger: ledger}
}

// Execute fetches the transaction by ID and its associated ledger entries.
// Returns CodeNotFound if no transaction has that ID.
// Returns CodeInvalidArgument if transactionID is empty.
func (uc *GetTransactionUseCase) Execute(ctx context.Context, transactionID string) (*GetTransactionResult, error) {
	if strings.TrimSpace(transactionID) == "" {
		return nil, errors.InvalidArgument("transaction_id must not be empty")
	}
	tx, err := uc.transactions.FindByID(ctx, transactionID)
	if err != nil {
		return nil, err
	}
	entries, err := uc.ledger.FindByTransactionID(ctx, transactionID)
	if err != nil {
		return nil, err
	}
	return &GetTransactionResult{
		Transaction: tx,
		Entries:     entries,
	}, nil
}
