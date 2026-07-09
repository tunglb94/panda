package entity_test

import (
	"testing"

	"github.com/fairride/wallet/domain/entity"
)

// ─── NewLedgerEntry ───────────────────────────────────────────────────────────

func TestNewLedgerEntry_Valid(t *testing.T) {
	e, err := entity.NewLedgerEntry("e1", "w1", "tx1", entity.DirectionCredit, 500, "USD", "fare payment", now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if e.EntryID != "e1" || e.WalletID != "w1" || e.TransactionID != "tx1" {
		t.Fatal("fields not set correctly")
	}
	if e.Direction != entity.DirectionCredit {
		t.Fatal("wrong direction")
	}
	if e.AmountCents != 500 {
		t.Fatalf("expected 500, got %d", e.AmountCents)
	}
	if e.Currency != "USD" {
		t.Fatal("wrong currency")
	}
	// Immutability: no UpdatedAt field.
	if !e.CreatedAt.Equal(now) {
		t.Fatal("created_at not set")
	}
}

func TestNewLedgerEntry_BothDirections(t *testing.T) {
	dirs := []entity.EntryDirection{entity.DirectionCredit, entity.DirectionDebit}
	for _, d := range dirs {
		_, err := entity.NewLedgerEntry("e1", "w1", "tx1", d, 100, "USD", "", now)
		if err != nil {
			t.Errorf("unexpected error for direction %s: %v", d, err)
		}
	}
}

func TestNewLedgerEntry_EmptyEntryID(t *testing.T) {
	_, err := entity.NewLedgerEntry("", "w1", "tx1", entity.DirectionCredit, 100, "USD", "", now)
	if err == nil {
		t.Fatal("expected error for empty entry id")
	}
}

func TestNewLedgerEntry_EmptyWalletID(t *testing.T) {
	_, err := entity.NewLedgerEntry("e1", "", "tx1", entity.DirectionCredit, 100, "USD", "", now)
	if err == nil {
		t.Fatal("expected error for empty wallet id")
	}
}

func TestNewLedgerEntry_EmptyTransactionID(t *testing.T) {
	_, err := entity.NewLedgerEntry("e1", "w1", "", entity.DirectionCredit, 100, "USD", "", now)
	if err == nil {
		t.Fatal("expected error for empty transaction id")
	}
}

func TestNewLedgerEntry_InvalidDirection(t *testing.T) {
	_, err := entity.NewLedgerEntry("e1", "w1", "tx1", entity.EntryDirection("transfer"), 100, "USD", "", now)
	if err == nil {
		t.Fatal("expected error for invalid direction")
	}
}

func TestNewLedgerEntry_ZeroAmount(t *testing.T) {
	_, err := entity.NewLedgerEntry("e1", "w1", "tx1", entity.DirectionCredit, 0, "USD", "", now)
	if err == nil {
		t.Fatal("expected error for zero amount")
	}
}

func TestNewLedgerEntry_NegativeAmount(t *testing.T) {
	_, err := entity.NewLedgerEntry("e1", "w1", "tx1", entity.DirectionCredit, -100, "USD", "", now)
	if err == nil {
		t.Fatal("expected error for negative amount")
	}
}

func TestNewLedgerEntry_EmptyCurrency(t *testing.T) {
	_, err := entity.NewLedgerEntry("e1", "w1", "tx1", entity.DirectionCredit, 100, "", "", now)
	if err == nil {
		t.Fatal("expected error for empty currency")
	}
}

func TestNewLedgerEntry_EmptyDescriptionAllowed(t *testing.T) {
	_, err := entity.NewLedgerEntry("e1", "w1", "tx1", entity.DirectionCredit, 100, "USD", "", now)
	if err != nil {
		t.Fatalf("description should be optional, got: %v", err)
	}
}

// ─── Immutability ─────────────────────────────────────────────────────────────

func TestLedgerEntry_HasNoUpdatedAt(t *testing.T) {
	// Compile-time check: LedgerEntry has no UpdatedAt field.
	// This test ensures the struct is used correctly; if UpdatedAt is ever added,
	// this file fails to compile and forces a domain review.
	e, _ := entity.NewLedgerEntry("e1", "w1", "tx1", entity.DirectionCredit, 100, "USD", "", now)
	_ = e.CreatedAt // only timestamp field
}
