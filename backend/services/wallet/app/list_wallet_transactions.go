package app

import (
	"context"
	"sort"
	"strings"

	"github.com/fairride/shared/errors"
	"github.com/fairride/wallet/domain/entity"
	"github.com/fairride/wallet/domain/repository"
)

// WalletTransactionLine is one row of the driver-facing Transaction History
// (Phần 6/9) — every ledger movement on a driver's own wallet, across every
// category (Ride/Delivery income, Commission, Bonus, Withdrawal, Refund,
// Adjustment, ...), not just trip settlements.
type WalletTransactionLine struct {
	TransactionID string
	Type          entity.TransactionType
	Direction     entity.EntryDirection
	AmountCents   int64
	Currency      string
	Description   string
	PaymentMethod string
	CreatedAt     int64 // unix seconds
}

// ListWalletTransactionsUseCase returns a driver's own wallet ledger,
// newest first, joined with each entry's Transaction metadata.
type ListWalletTransactionsUseCase struct {
	wallets repository.WalletRepository
	ledger  repository.LedgerEntryRepository
	tx      repository.TransactionRepository
}

func NewListWalletTransactionsUseCase(wallets repository.WalletRepository, ledger repository.LedgerEntryRepository, tx repository.TransactionRepository) *ListWalletTransactionsUseCase {
	return &ListWalletTransactionsUseCase{wallets: wallets, ledger: ledger, tx: tx}
}

func (uc *ListWalletTransactionsUseCase) Execute(ctx context.Context, driverID string, limit int) ([]WalletTransactionLine, error) {
	if strings.TrimSpace(driverID) == "" {
		return nil, errors.InvalidArgument("driver_id must not be empty")
	}
	wallet, err := uc.wallets.FindByOwnerID(ctx, driverID)
	if err != nil {
		if errors.GetCode(err) == errors.CodeNotFound {
			return []WalletTransactionLine{}, nil
		}
		return nil, err
	}
	entries, err := uc.ledger.FindByWalletID(ctx, wallet.WalletID)
	if err != nil {
		return nil, err
	}
	out := make([]WalletTransactionLine, 0, len(entries))
	for i := range entries {
		e := &entries[i]
		t, err := uc.tx.FindByID(ctx, e.TransactionID)
		if err != nil {
			continue
		}
		out = append(out, WalletTransactionLine{
			TransactionID: e.TransactionID, Type: t.Type, Direction: e.Direction,
			AmountCents: e.AmountCents, Currency: e.Currency, Description: e.Description,
			PaymentMethod: t.PaymentMethod, CreatedAt: e.CreatedAt.Unix(),
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt > out[j].CreatedAt })
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}
