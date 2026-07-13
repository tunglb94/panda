package app

import (
	"context"
	"strings"

	"github.com/fairride/shared/errors"
	"github.com/fairride/wallet/domain/entity"
	"github.com/fairride/wallet/domain/repository"
)

// WalletSummary is the Phần 3 Wallet View — every field is derived from
// the ledger at read time (Phần 13 — "Wallet không được tự tính. Mọi số
// liệu phải đọc từ Ledger."); nothing here is a stored column.
type WalletSummary struct {
	WalletID               string
	OwnerID                string
	Currency               string
	AvailableCents         int64 // withdrawable now — electronically-collected income minus withdrawals/penalties
	PendingCents           int64 // sum of the driver's own in-flight (Pending/Approved) payout requests
	OutstandingCents       int64 // commission owed from cash-collected trips (Phần 4)
	NetCents               int64 // Available - Outstanding — what Phần 5 actually gates payout against
	LifetimeEarnedCents    int64
	LifetimeWithdrawnCents int64
}

// creditIncomeTypes are the ledger transaction types that count toward
// Lifetime Earnings (Phần 3) regardless of collection method.
var creditIncomeTypes = map[entity.TransactionType]bool{
	entity.TypeRideIncome:     true,
	entity.TypeDeliveryIncome: true,
	entity.TypeBonus:          true,
	entity.TypeManualCredit:   true,
	entity.TypeAdjustment:     true,
}

// GetWalletSummaryUseCase computes a driver's full Wallet Projection.
type GetWalletSummaryUseCase struct {
	wallets repository.WalletRepository
	ledger  repository.LedgerEntryRepository
	tx      repository.TransactionRepository
	payouts repository.PayoutRequestRepository
}

func NewGetWalletSummaryUseCase(
	wallets repository.WalletRepository,
	ledger repository.LedgerEntryRepository,
	tx repository.TransactionRepository,
	payouts repository.PayoutRequestRepository,
) *GetWalletSummaryUseCase {
	return &GetWalletSummaryUseCase{wallets: wallets, ledger: ledger, tx: tx, payouts: payouts}
}

func (uc *GetWalletSummaryUseCase) Execute(ctx context.Context, driverID string) (*WalletSummary, error) {
	if strings.TrimSpace(driverID) == "" {
		return nil, errors.InvalidArgument("driver_id must not be empty")
	}
	wallet, err := uc.wallets.FindByOwnerID(ctx, driverID)
	if err != nil {
		if errors.GetCode(err) == errors.CodeNotFound {
			// A driver with no trips yet has no wallet row — that's a
			// legitimate zero-everything state, not an error.
			return &WalletSummary{OwnerID: driverID, Currency: DefaultCurrency}, nil
		}
		return nil, err
	}
	entries, err := uc.ledger.FindByWalletID(ctx, wallet.WalletID)
	if err != nil {
		return nil, err
	}

	// Resolve each entry's owning Transaction (Type + PaymentMethod) — the
	// categorization Available/Outstanding/Lifetime lives on the
	// Transaction, not the LedgerEntry itself. One batch query instead of
	// one FindByID per unique transaction (was N+1 at scale).
	uniqueIDs := make([]string, 0, len(entries))
	seen := map[string]bool{}
	for i := range entries {
		txID := entries[i].TransactionID
		if !seen[txID] {
			seen[txID] = true
			uniqueIDs = append(uniqueIDs, txID)
		}
	}
	txByID, err := uc.tx.FindByIDs(ctx, uniqueIDs)
	if err != nil {
		return nil, err
	}

	summary := &WalletSummary{WalletID: wallet.WalletID, OwnerID: wallet.OwnerID, Currency: wallet.Currency}
	for i := range entries {
		e := &entries[i]
		t, ok := txByID[e.TransactionID]
		if !ok {
			continue
		}
		switch t.Type {
		case entity.TypeCommission:
			// Outstanding = net debit Commission entries (Phần 4). Only
			// ever debited on the driver's wallet for cash-collected
			// trips; a credit here (Admin manual reversal) reduces it.
			if e.Direction == entity.DirectionDebit {
				summary.OutstandingCents += e.AmountCents
			} else {
				summary.OutstandingCents -= e.AmountCents
			}
		case entity.TypeWithdrawal, entity.TypePenalty, entity.TypeManualDebit, entity.TypePayoutHold:
			// TypePayoutHold's debit freezes Available the moment a payout
			// request is created; its reversal credit (Reject) restores it —
			// same signed-sum treatment as Withdrawal/Penalty/ManualDebit.
			if e.Direction == entity.DirectionDebit {
				summary.AvailableCents -= e.AmountCents
			} else {
				summary.AvailableCents += e.AmountCents
			}
		default:
			if creditIncomeTypes[t.Type] && e.Direction == entity.DirectionCredit {
				summary.LifetimeEarnedCents += e.AmountCents
				// Cash-collected income never entered the app wallet — the
				// driver already holds it physically — so it's excluded
				// from Available (Phần 3), only electronically-collected
				// income is.
				if t.PaymentMethod != string(entity.PaymentMethodCash) {
					summary.AvailableCents += e.AmountCents
				}
			}
		}
		if t.Type == entity.TypeWithdrawal {
			summary.LifetimeWithdrawnCents += e.AmountCents
		}
	}
	if summary.OutstandingCents < 0 {
		summary.OutstandingCents = 0
	}

	inFlight, err := uc.payouts.FindInFlightByDriverID(ctx, driverID)
	if err == nil {
		summary.PendingCents = inFlight.AmountCents
	} else if errors.GetCode(err) != errors.CodeNotFound {
		return nil, err
	}

	summary.NetCents = summary.AvailableCents - summary.OutstandingCents
	return summary, nil
}
