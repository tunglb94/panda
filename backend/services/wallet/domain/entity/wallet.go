// Package entity contains the Wallet bounded-context domain model.
package entity

import (
	"strings"
	"time"

	"github.com/fairride/shared/errors"
)

// WalletType classifies wallet ownership.
type WalletType string

const (
	WalletTypeRider    WalletType = "rider"
	WalletTypeDriver   WalletType = "driver"
	WalletTypePlatform WalletType = "platform"
)

var validWalletTypes = map[WalletType]bool{
	WalletTypeRider:    true,
	WalletTypeDriver:   true,
	WalletTypePlatform: true,
}

// Wallet is the aggregate root for a FAIRRIDE wallet.
// Balance is NEVER stored; it is always derived from the ledger via ComputeBalance.
type Wallet struct {
	WalletID   string
	OwnerID    string
	WalletType WalletType
	Currency   string // ISO 4217 (e.g. "USD")
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// NewWallet creates a validated Wallet.
// walletID and ownerID are required. walletType must be a known WalletType.
// currency must be non-empty (ISO 4217 code).
func NewWallet(walletID, ownerID string, walletType WalletType, currency string, now time.Time) (*Wallet, error) {
	if strings.TrimSpace(walletID) == "" {
		return nil, errors.InvalidArgument("wallet id must not be empty")
	}
	if strings.TrimSpace(ownerID) == "" {
		return nil, errors.InvalidArgument("owner id must not be empty")
	}
	if !validWalletTypes[walletType] {
		return nil, errors.InvalidArgument("invalid wallet type: " + string(walletType))
	}
	if strings.TrimSpace(currency) == "" {
		return nil, errors.InvalidArgument("currency must not be empty")
	}
	return &Wallet{
		WalletID:   walletID,
		OwnerID:    ownerID,
		WalletType: walletType,
		Currency:   currency,
		CreatedAt:  now,
		UpdatedAt:  now,
	}, nil
}

// ReconstituteWallet rebuilds a Wallet from a persistence record without validation.
func ReconstituteWallet(
	walletID, ownerID string,
	walletType WalletType,
	currency string,
	createdAt, updatedAt time.Time,
) *Wallet {
	return &Wallet{
		WalletID:   walletID,
		OwnerID:    ownerID,
		WalletType: walletType,
		Currency:   currency,
		CreatedAt:  createdAt,
		UpdatedAt:  updatedAt,
	}
}

// ComputeBalance derives the current balance from a slice of ledger entries.
// Returns CodeInvalidArgument if any entry does not belong to this wallet
// or carries a different currency.
// Returns 0 with no error when entries is empty (new wallet, zero balance).
func (w *Wallet) ComputeBalance(entries []LedgerEntry) (int64, error) {
	var balance int64
	for i := range entries {
		e := &entries[i]
		if e.WalletID != w.WalletID {
			return 0, errors.InvalidArgument("ledger entry wallet_id mismatch")
		}
		if e.Currency != w.Currency {
			return 0, errors.InvalidArgument("ledger entry currency mismatch")
		}
		switch e.Direction {
		case DirectionCredit:
			balance += e.AmountCents
		case DirectionDebit:
			balance -= e.AmountCents
		default:
			return 0, errors.InvalidArgument("unknown entry direction: " + string(e.Direction))
		}
	}
	return balance, nil
}
