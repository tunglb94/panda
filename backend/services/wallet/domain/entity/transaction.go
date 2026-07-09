package entity

import (
	"strings"
	"time"

	"github.com/fairride/shared/errors"
)

// TransactionType classifies the financial event that caused the ledger entries.
type TransactionType string

const (
	// TypeTripPayment records the rider's payment for a completed trip.
	TypeTripPayment TransactionType = "trip_payment"
	// TypeTripEarnings records the driver's gross earnings from a completed trip.
	TypeTripEarnings TransactionType = "trip_earnings"
	// TypePlatformCommission records the platform's commission from a completed trip.
	TypePlatformCommission TransactionType = "platform_commission"
	// TypeRefund records a refund credited back to the rider.
	TypeRefund TransactionType = "refund"
	// TypeAdjustment records a manual credit or debit by platform operations.
	TypeAdjustment TransactionType = "adjustment"
)

var validTransactionTypes = map[TransactionType]bool{
	TypeTripPayment:        true,
	TypeTripEarnings:       true,
	TypePlatformCommission: true,
	TypeRefund:             true,
	TypeAdjustment:         true,
}

// Transaction groups one or more LedgerEntries into a single atomic financial event.
// Transactions are immutable after creation (append-only financial log).
type Transaction struct {
	TransactionID string
	Type          TransactionType
	ReferenceID   string // e.g. trip_id for trip-related transactions; may be empty
	Currency      string // ISO 4217
	Description   string
	CreatedAt     time.Time
	// No UpdatedAt — Transaction is immutable.
}

// NewTransaction creates a validated Transaction.
func NewTransaction(
	transactionID string,
	txType TransactionType,
	referenceID, currency, description string,
	now time.Time,
) (*Transaction, error) {
	if strings.TrimSpace(transactionID) == "" {
		return nil, errors.InvalidArgument("transaction id must not be empty")
	}
	if !validTransactionTypes[txType] {
		return nil, errors.InvalidArgument("invalid transaction type: " + string(txType))
	}
	if strings.TrimSpace(currency) == "" {
		return nil, errors.InvalidArgument("currency must not be empty")
	}
	return &Transaction{
		TransactionID: transactionID,
		Type:          txType,
		ReferenceID:   referenceID,
		Currency:      currency,
		Description:   description,
		CreatedAt:     now,
	}, nil
}

// ReconstituteTransaction rebuilds a Transaction from persistence without validation.
func ReconstituteTransaction(
	transactionID string,
	txType TransactionType,
	referenceID, currency, description string,
	createdAt time.Time,
) *Transaction {
	return &Transaction{
		TransactionID: transactionID,
		Type:          txType,
		ReferenceID:   referenceID,
		Currency:      currency,
		Description:   description,
		CreatedAt:     createdAt,
	}
}
