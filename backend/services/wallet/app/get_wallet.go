// Package app contains the Wallet service application layer (use cases).
package app

import (
	"context"
	"strings"

	"github.com/fairride/shared/errors"
	"github.com/fairride/wallet/domain/entity"
	"github.com/fairride/wallet/domain/repository"
)

// GetWalletUseCase returns the wallet owned by a given ownerID.
type GetWalletUseCase struct {
	wallets repository.WalletRepository
}

// NewGetWalletUseCase wires the use case.
func NewGetWalletUseCase(wallets repository.WalletRepository) *GetWalletUseCase {
	return &GetWalletUseCase{wallets: wallets}
}

// Execute finds and returns the wallet for ownerID.
// Returns CodeNotFound if the owner has no wallet.
// Returns CodeInvalidArgument if ownerID is empty.
func (uc *GetWalletUseCase) Execute(ctx context.Context, ownerID string) (*entity.Wallet, error) {
	if strings.TrimSpace(ownerID) == "" {
		return nil, errors.InvalidArgument("owner_id must not be empty")
	}
	return uc.wallets.FindByOwnerID(ctx, ownerID)
}
