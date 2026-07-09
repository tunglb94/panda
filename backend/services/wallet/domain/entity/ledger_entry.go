package entity

import (
	"strings"
	"time"

	"github.com/fairride/shared/errors"
)

// EntryDirection classifies whether an entry increases or decreases the balance.
type EntryDirection string

const (
	DirectionCredit EntryDirection = "credit" // increases balance
	DirectionDebit  EntryDirection = "debit"  // decreases balance
)

var validDirections = map[EntryDirection]bool{
	DirectionCredit: true,
	DirectionDebit:  true,
}

// LedgerEntry is an immutable financial record.
// Entries are never updated or deleted; they are the source of truth for wallet balances.
type LedgerEntry struct {
	EntryID       string
	WalletID      string
	TransactionID string
	Direction     EntryDirection
	AmountCents   int64  // always positive; direction carries the sign
	Currency      string // ISO 4217
	Description   string
	CreatedAt     time.Time
	// No UpdatedAt — LedgerEntry is immutable after creation.
}

// NewLedgerEntry creates a validated, immutable LedgerEntry.
// AmountCents must be strictly positive.
func NewLedgerEntry(
	entryID, walletID, transactionID string,
	direction EntryDirection,
	amountCents int64,
	currency, description string,
	now time.Time,
) (*LedgerEntry, error) {
	if strings.TrimSpace(entryID) == "" {
		return nil, errors.InvalidArgument("entry id must not be empty")
	}
	if strings.TrimSpace(walletID) == "" {
		return nil, errors.InvalidArgument("wallet id must not be empty")
	}
	if strings.TrimSpace(transactionID) == "" {
		return nil, errors.InvalidArgument("transaction id must not be empty")
	}
	if !validDirections[direction] {
		return nil, errors.InvalidArgument("invalid direction: " + string(direction))
	}
	if amountCents <= 0 {
		return nil, errors.InvalidArgument("amount_cents must be greater than zero")
	}
	if strings.TrimSpace(currency) == "" {
		return nil, errors.InvalidArgument("currency must not be empty")
	}
	return &LedgerEntry{
		EntryID:       entryID,
		WalletID:      walletID,
		TransactionID: transactionID,
		Direction:     direction,
		AmountCents:   amountCents,
		Currency:      currency,
		Description:   description,
		CreatedAt:     now,
	}, nil
}

// ReconstituteLedgerEntry rebuilds a LedgerEntry from persistence without validation.
func ReconstituteLedgerEntry(
	entryID, walletID, transactionID string,
	direction EntryDirection,
	amountCents int64,
	currency, description string,
	createdAt time.Time,
) *LedgerEntry {
	return &LedgerEntry{
		EntryID:       entryID,
		WalletID:      walletID,
		TransactionID: transactionID,
		Direction:     direction,
		AmountCents:   amountCents,
		Currency:      currency,
		Description:   description,
		CreatedAt:     createdAt,
	}
}
