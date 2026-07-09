package app

import (
	"context"
	"strings"

	"github.com/fairride/shared/errors"
	"github.com/fairride/wallet/domain/entity"
	"github.com/fairride/wallet/domain/repository"
)

// GetLedgerUseCase returns all ledger entries for a wallet.
type GetLedgerUseCase struct {
	wallets repository.WalletRepository
	ledger  repository.LedgerEntryRepository
}

// NewGetLedgerUseCase wires the use case.
func NewGetLedgerUseCase(
	wallets repository.WalletRepository,
	ledger repository.LedgerEntryRepository,
) *GetLedgerUseCase {
	return &GetLedgerUseCase{wallets: wallets, ledger: ledger}
}

// Execute validates the wallet exists then returns its ledger entries.
// Returns CodeNotFound if walletID does not exist.
// Returns CodeInvalidArgument if walletID is empty.
func (uc *GetLedgerUseCase) Execute(ctx context.Context, walletID string) ([]entity.LedgerEntry, error) {
	if strings.TrimSpace(walletID) == "" {
		return nil, errors.InvalidArgument("wallet_id must not be empty")
	}
	// Verify the wallet exists before querying ledger.
	if _, err := uc.wallets.FindByID(ctx, walletID); err != nil {
		return nil, err
	}
	return uc.ledger.FindByWalletID(ctx, walletID)
}
