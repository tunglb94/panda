package app

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/fairride/shared/errors"
	"github.com/fairride/wallet/domain/entity"
	"github.com/fairride/wallet/domain/repository"
)

// PlatformOwnerID is the sentinel OwnerID for Panda's own platform wallet
// (WalletTypePlatform) — a single row, lazily created on first use, the
// same way a driver's wallet is. There is exactly one platform wallet.
const PlatformOwnerID = "platform"

// DefaultCurrency is VND (Vietnamese Dong has no decimal subunit — see
// pricing/domain/entity/fare.go's doc comment) — every wallet, transaction,
// and ledger entry this module creates uses VND. Never USD; this is a
// Vietnam-market platform.
const DefaultCurrency = "VND"

// GetOrCreateWalletUseCase idempotently ensures a wallet exists for a given
// owner, creating one on first use. Used by the Settlement Engine (which
// must not fail a trip settlement just because this is a driver's or the
// platform's first-ever transaction) and by the driver-facing wallet
// summary endpoint.
type GetOrCreateWalletUseCase struct {
	wallets repository.WalletRepository
}

func NewGetOrCreateWalletUseCase(wallets repository.WalletRepository) *GetOrCreateWalletUseCase {
	return &GetOrCreateWalletUseCase{wallets: wallets}
}

func (uc *GetOrCreateWalletUseCase) Execute(ctx context.Context, ownerID string, walletType entity.WalletType) (*entity.Wallet, error) {
	if strings.TrimSpace(ownerID) == "" {
		return nil, errors.InvalidArgument("owner_id must not be empty")
	}
	existing, err := uc.wallets.FindByOwnerID(ctx, ownerID)
	if err == nil {
		return existing, nil
	}
	if errors.GetCode(err) != errors.CodeNotFound {
		return nil, err
	}
	now := time.Now().UTC()
	w, err := entity.NewWallet(uuid.NewString(), ownerID, walletType, DefaultCurrency, now)
	if err != nil {
		return nil, err
	}
	if err := uc.wallets.Save(ctx, w); err != nil && errors.GetCode(err) != errors.CodeAlreadyExists {
		return nil, err
	}
	// Always re-read the authoritative row rather than trusting the
	// locally-generated struct — Save is an upsert-by-owner_id, so a
	// concurrent first-creation race could mean the row now on disk has a
	// different WalletID than the one generated in this call.
	return uc.wallets.FindByOwnerID(ctx, ownerID)
}
