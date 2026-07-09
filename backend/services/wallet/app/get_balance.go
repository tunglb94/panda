package app

import (
	"context"
	"strings"

	"github.com/fairride/shared/errors"
	"github.com/fairride/wallet/domain/repository"
)

// GetBalanceResult carries the computed wallet balance.
type GetBalanceResult struct {
	WalletID     string
	OwnerID      string
	BalanceCents int64
	Currency     string
}

// GetBalanceUseCase computes the current balance for a wallet owner.
// Balance is always derived from the ledger — never read from a stored field.
type GetBalanceUseCase struct {
	wallets repository.WalletRepository
	ledger  repository.LedgerEntryRepository
}

// NewGetBalanceUseCase wires the use case.
func NewGetBalanceUseCase(
	wallets repository.WalletRepository,
	ledger repository.LedgerEntryRepository,
) *GetBalanceUseCase {
	return &GetBalanceUseCase{wallets: wallets, ledger: ledger}
}

// Execute fetches the wallet and its ledger entries, then derives the balance.
// Returns CodeNotFound if the owner has no wallet.
// Returns CodeInvalidArgument if ownerID is empty.
func (uc *GetBalanceUseCase) Execute(ctx context.Context, ownerID string) (*GetBalanceResult, error) {
	if strings.TrimSpace(ownerID) == "" {
		return nil, errors.InvalidArgument("owner_id must not be empty")
	}
	wallet, err := uc.wallets.FindByOwnerID(ctx, ownerID)
	if err != nil {
		return nil, err
	}
	entries, err := uc.ledger.FindByWalletID(ctx, wallet.WalletID)
	if err != nil {
		return nil, err
	}
	balance, err := wallet.ComputeBalance(entries)
	if err != nil {
		return nil, err
	}
	return &GetBalanceResult{
		WalletID:     wallet.WalletID,
		OwnerID:      wallet.OwnerID,
		BalanceCents: balance,
		Currency:     wallet.Currency,
	}, nil
}
