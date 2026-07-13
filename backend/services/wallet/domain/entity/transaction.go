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
	// Kept for backward compatibility with the Wallet Foundation phase —
	// unused by the Settlement Engine (Phần 1), which uses the more
	// specific types below instead.
	TypeTripPayment TransactionType = "trip_payment"
	// TypeTripEarnings records the driver's gross earnings from a completed
	// trip. Kept for backward compatibility — superseded by
	// TypeRideIncome/TypeDeliveryIncome.
	TypeTripEarnings TransactionType = "trip_earnings"
	// TypePlatformCommission is the Wallet-Foundation-phase commission type.
	// Kept for backward compatibility — superseded by TypeCommission.
	TypePlatformCommission TransactionType = "platform_commission"

	// TypeRideIncome records a driver's income from a completed ride
	// (Settlement Engine Case 1/2 — see settlement.go).
	TypeRideIncome TransactionType = "ride_income"
	// TypeDeliveryIncome records a driver's income from a completed
	// delivery. Reserved: Delivery has no fare/payment hook yet (Known Gap).
	TypeDeliveryIncome TransactionType = "delivery_income"
	// TypeCashCollected is informational — tags a trip whose fare the
	// driver physically collected as cash (see Transaction.PaymentMethod,
	// which is the field actually used to filter Available balance; this
	// type exists so Cash appears as its own Transaction History category
	// per Phần 6/9's "Cash / Electronic" filter).
	TypeCashCollected TransactionType = "cash_collected"
	// TypePlatformReceivable records the platform's claim against a driver
	// who collected cash and owes the platform its commission share
	// (Settlement Case 1 — becomes the driver's Outstanding balance).
	TypePlatformReceivable TransactionType = "platform_receivable"
	// TypePlatformPayable records the platform's obligation to pay a driver
	// their net income after collecting the fare electronically
	// (Settlement Case 2 — becomes the driver's Available balance).
	TypePlatformPayable TransactionType = "platform_payable"
	// TypeCommission records the platform's commission on a trip, in either
	// Settlement case.
	TypeCommission TransactionType = "commission"
	// TypePromotionSubsidy / TypeVoucherSubsidy record the platform
	// absorbing a promotion/voucher discount so driver income is
	// unaffected (Settlement Case 3). Reserved: Promotion/Voucher data
	// does not reach a completed Trip yet (Known Gap) — no code path
	// creates these today, but the ledger model supports them.
	TypePromotionSubsidy TransactionType = "promotion_subsidy"
	TypeVoucherSubsidy   TransactionType = "voucher_subsidy"
	// TypeBonus records a platform-initiated bonus credited to a driver.
	TypeBonus TransactionType = "bonus"
	// TypePenalty records a platform-initiated penalty debited from a driver.
	TypePenalty TransactionType = "penalty"
	// TypeWithdrawal records a paid-out payout request (Phần 5/8 — created
	// only at the Paid transition, never at Pending/Approved).
	TypeWithdrawal TransactionType = "withdrawal"
	// TypePayoutHold freezes Available the moment a PayoutRequest is
	// created (a debit) so the same money can't be claimed by a second
	// request or silently keep counting as withdrawable while one is in
	// flight (critique #8 — Frozen Balance). Reversed with a credit entry
	// of the same amount on Reject, or on Paid (where the matching
	// TypeWithdrawal debit takes over as the permanent reduction).
	TypePayoutHold TransactionType = "payout_hold"
	// TypeRefund records a refund-related ledger movement.
	TypeRefund TransactionType = "refund"
	// TypeAdjustment is a generic manual correction (kept for backward
	// compatibility — TypeManualCredit/TypeManualDebit are preferred for
	// new Admin-initiated adjustments since they carry an explicit sign).
	TypeAdjustment TransactionType = "adjustment"
	// TypeManualCredit / TypeManualDebit record an Admin Manual Adjustment
	// (Phần 10) — always paired with a wallet_audit_logs entry.
	TypeManualCredit TransactionType = "manual_credit"
	TypeManualDebit  TransactionType = "manual_debit"
)

var validTransactionTypes = map[TransactionType]bool{
	TypeTripPayment:        true,
	TypeTripEarnings:       true,
	TypePlatformCommission: true,
	TypeRideIncome:         true,
	TypeDeliveryIncome:     true,
	TypeCashCollected:      true,
	TypePlatformReceivable: true,
	TypePlatformPayable:    true,
	TypeCommission:         true,
	TypePromotionSubsidy:   true,
	TypeVoucherSubsidy:     true,
	TypeBonus:              true,
	TypePenalty:            true,
	TypeWithdrawal:         true,
	TypePayoutHold:         true,
	TypeRefund:             true,
	TypeAdjustment:         true,
	TypeManualCredit:       true,
	TypeManualDebit:        true,
}

// validPaymentMethods — empty string is always allowed (non-trip
// transactions, e.g. Withdrawal/Adjustment, carry no payment method).
var validPaymentMethods = map[string]bool{
	"":       true,
	"cash":   true,
	"wallet": true,
}

// Transaction groups one or more LedgerEntries into a single atomic financial event.
// Transactions are immutable after creation (append-only financial log).
//
// PaymentMethod ("cash" | "wallet" | "") is how the Wallet Projection
// (Phần 3) tells a cash-collected trip (driver already holds the money —
// excluded from Available, tracked as Outstanding via the commission owed)
// apart from an electronically-collected trip (platform holds the money —
// included in Available once paid out). Empty for non-trip transactions
// (Withdrawal, Bonus, Adjustment, ...).
type Transaction struct {
	TransactionID string
	Type          TransactionType
	ReferenceID   string // e.g. trip_id for trip-related transactions; may be empty
	PaymentMethod string // "cash" | "wallet" | ""
	Currency      string // ISO 4217
	Description   string
	CreatedAt     time.Time
	// No UpdatedAt — Transaction is immutable.
}

// NewTransaction creates a validated Transaction.
func NewTransaction(
	transactionID string,
	txType TransactionType,
	referenceID, paymentMethod, currency, description string,
	now time.Time,
) (*Transaction, error) {
	if strings.TrimSpace(transactionID) == "" {
		return nil, errors.InvalidArgument("transaction id must not be empty")
	}
	if !validTransactionTypes[txType] {
		return nil, errors.InvalidArgument("invalid transaction type: " + string(txType))
	}
	if !validPaymentMethods[paymentMethod] {
		return nil, errors.InvalidArgument("invalid payment method: " + paymentMethod)
	}
	if strings.TrimSpace(currency) == "" {
		return nil, errors.InvalidArgument("currency must not be empty")
	}
	return &Transaction{
		TransactionID: transactionID,
		Type:          txType,
		ReferenceID:   referenceID,
		PaymentMethod: paymentMethod,
		Currency:      currency,
		Description:   description,
		CreatedAt:     now,
	}, nil
}

// ReconstituteTransaction rebuilds a Transaction from persistence without validation.
func ReconstituteTransaction(
	transactionID string,
	txType TransactionType,
	referenceID, paymentMethod, currency, description string,
	createdAt time.Time,
) *Transaction {
	return &Transaction{
		TransactionID: transactionID,
		Type:          txType,
		ReferenceID:   referenceID,
		PaymentMethod: paymentMethod,
		Currency:      currency,
		Description:   description,
		CreatedAt:     createdAt,
	}
}
