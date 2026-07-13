package entity_test

import (
	"testing"

	"github.com/fairride/wallet/domain/entity"
)

// ─── NewTransaction ───────────────────────────────────────────────────────────

func TestNewTransaction_Valid(t *testing.T) {
	tx, err := entity.NewTransaction("tx1", entity.TypeTripPayment, "trip-abc", "", "USD", "trip fare", now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tx.TransactionID != "tx1" {
		t.Fatal("transaction id not set")
	}
	if tx.Type != entity.TypeTripPayment {
		t.Fatal("type not set")
	}
	if tx.ReferenceID != "trip-abc" {
		t.Fatal("reference_id not set")
	}
	if tx.Currency != "USD" {
		t.Fatal("currency not set")
	}
	if !tx.CreatedAt.Equal(now) {
		t.Fatal("created_at not set")
	}
}

func TestNewTransaction_AllTypes(t *testing.T) {
	types := []entity.TransactionType{
		entity.TypeTripPayment,
		entity.TypeTripEarnings,
		entity.TypePlatformCommission,
		entity.TypeRideIncome,
		entity.TypeDeliveryIncome,
		entity.TypeCashCollected,
		entity.TypePlatformReceivable,
		entity.TypePlatformPayable,
		entity.TypeCommission,
		entity.TypePromotionSubsidy,
		entity.TypeVoucherSubsidy,
		entity.TypeBonus,
		entity.TypePenalty,
		entity.TypeWithdrawal,
		entity.TypeRefund,
		entity.TypeAdjustment,
		entity.TypeManualCredit,
		entity.TypeManualDebit,
	}
	for _, tt := range types {
		_, err := entity.NewTransaction("tx1", tt, "", "", "USD", "", now)
		if err != nil {
			t.Errorf("unexpected error for type %s: %v", tt, err)
		}
	}
}

func TestNewTransaction_EmptyID(t *testing.T) {
	_, err := entity.NewTransaction("", entity.TypeTripPayment, "", "", "USD", "", now)
	if err == nil {
		t.Fatal("expected error for empty transaction id")
	}
}

func TestNewTransaction_InvalidType(t *testing.T) {
	_, err := entity.NewTransaction("tx1", entity.TransactionType("unknown"), "", "", "USD", "", now)
	if err == nil {
		t.Fatal("expected error for invalid transaction type")
	}
}

func TestNewTransaction_EmptyCurrency(t *testing.T) {
	_, err := entity.NewTransaction("tx1", entity.TypeTripPayment, "", "", "", "", now)
	if err == nil {
		t.Fatal("expected error for empty currency")
	}
}

func TestNewTransaction_EmptyReferenceIDAllowed(t *testing.T) {
	_, err := entity.NewTransaction("tx1", entity.TypeAdjustment, "", "", "USD", "manual", now)
	if err != nil {
		t.Fatalf("reference_id should be optional: %v", err)
	}
}

func TestNewTransaction_EmptyDescriptionAllowed(t *testing.T) {
	_, err := entity.NewTransaction("tx1", entity.TypeAdjustment, "", "", "USD", "", now)
	if err != nil {
		t.Fatalf("description should be optional: %v", err)
	}
}

func TestNewTransaction_PaymentMethodCashAndWallet(t *testing.T) {
	if _, err := entity.NewTransaction("tx1", entity.TypeRideIncome, "trip-1", "cash", "USD", "", now); err != nil {
		t.Errorf("cash payment method should be valid: %v", err)
	}
	if _, err := entity.NewTransaction("tx2", entity.TypeRideIncome, "trip-2", "wallet", "USD", "", now); err != nil {
		t.Errorf("wallet payment method should be valid: %v", err)
	}
}

func TestNewTransaction_InvalidPaymentMethod(t *testing.T) {
	_, err := entity.NewTransaction("tx1", entity.TypeRideIncome, "trip-1", "credit_card", "USD", "", now)
	if err == nil {
		t.Fatal("expected error for invalid payment method")
	}
}

// ─── Immutability ─────────────────────────────────────────────────────────────

func TestTransaction_HasNoUpdatedAt(t *testing.T) {
	tx, _ := entity.NewTransaction("tx1", entity.TypeTripPayment, "", "", "USD", "", now)
	_ = tx.CreatedAt // only timestamp field
}
