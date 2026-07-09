package entity_test

import (
	"testing"
	"time"

	"github.com/fairride/wallet/domain/entity"
)

var now = time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)

// ─── NewWallet ────────────────────────────────────────────────────────────────

func TestNewWallet_Valid(t *testing.T) {
	w, err := entity.NewWallet("w1", "owner1", entity.WalletTypeRider, "USD", now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if w.WalletID != "w1" || w.OwnerID != "owner1" {
		t.Fatalf("fields not set correctly")
	}
	if w.WalletType != entity.WalletTypeRider {
		t.Fatalf("wrong wallet type")
	}
	if w.Currency != "USD" {
		t.Fatalf("wrong currency")
	}
	if !w.CreatedAt.Equal(now) || !w.UpdatedAt.Equal(now) {
		t.Fatalf("timestamps not set")
	}
}

func TestNewWallet_AllTypes(t *testing.T) {
	types := []entity.WalletType{entity.WalletTypeRider, entity.WalletTypeDriver, entity.WalletTypePlatform}
	for _, wt := range types {
		_, err := entity.NewWallet("w1", "o1", wt, "USD", now)
		if err != nil {
			t.Errorf("unexpected error for type %s: %v", wt, err)
		}
	}
}

func TestNewWallet_EmptyID(t *testing.T) {
	_, err := entity.NewWallet("", "owner1", entity.WalletTypeRider, "USD", now)
	if err == nil {
		t.Fatal("expected error for empty wallet id")
	}
}

func TestNewWallet_EmptyOwner(t *testing.T) {
	_, err := entity.NewWallet("w1", "", entity.WalletTypeRider, "USD", now)
	if err == nil {
		t.Fatal("expected error for empty owner id")
	}
}

func TestNewWallet_InvalidType(t *testing.T) {
	_, err := entity.NewWallet("w1", "o1", entity.WalletType("admin"), "USD", now)
	if err == nil {
		t.Fatal("expected error for invalid wallet type")
	}
}

func TestNewWallet_EmptyCurrency(t *testing.T) {
	_, err := entity.NewWallet("w1", "o1", entity.WalletTypeRider, "", now)
	if err == nil {
		t.Fatal("expected error for empty currency")
	}
}

// ─── ComputeBalance ───────────────────────────────────────────────────────────

func makeWallet(t *testing.T) *entity.Wallet {
	t.Helper()
	w, err := entity.NewWallet("w1", "o1", entity.WalletTypeRider, "USD", now)
	if err != nil {
		t.Fatal(err)
	}
	return w
}

func makeEntry(t *testing.T, dir entity.EntryDirection, cents int64) entity.LedgerEntry {
	t.Helper()
	e, err := entity.NewLedgerEntry("e1", "w1", "tx1", dir, cents, "USD", "", now)
	if err != nil {
		t.Fatal(err)
	}
	return *e
}

func TestComputeBalance_Empty(t *testing.T) {
	w := makeWallet(t)
	bal, err := w.ComputeBalance(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if bal != 0 {
		t.Fatalf("expected 0 balance for empty ledger, got %d", bal)
	}
}

func TestComputeBalance_CreditOnly(t *testing.T) {
	w := makeWallet(t)
	entries := []entity.LedgerEntry{
		makeEntry(t, entity.DirectionCredit, 1000),
		makeEntry(t, entity.DirectionCredit, 500),
	}
	bal, err := w.ComputeBalance(entries)
	if err != nil {
		t.Fatal(err)
	}
	if bal != 1500 {
		t.Fatalf("expected 1500, got %d", bal)
	}
}

func TestComputeBalance_DebitOnly(t *testing.T) {
	w := makeWallet(t)
	entries := []entity.LedgerEntry{
		makeEntry(t, entity.DirectionDebit, 300),
	}
	bal, err := w.ComputeBalance(entries)
	if err != nil {
		t.Fatal(err)
	}
	if bal != -300 {
		t.Fatalf("expected -300, got %d", bal)
	}
}

func TestComputeBalance_CreditAndDebit(t *testing.T) {
	w := makeWallet(t)
	entries := []entity.LedgerEntry{
		makeEntry(t, entity.DirectionCredit, 2000),
		makeEntry(t, entity.DirectionDebit, 750),
		makeEntry(t, entity.DirectionCredit, 250),
	}
	bal, err := w.ComputeBalance(entries)
	if err != nil {
		t.Fatal(err)
	}
	if bal != 1500 {
		t.Fatalf("expected 1500, got %d", bal)
	}
}

func TestComputeBalance_WalletIDMismatch(t *testing.T) {
	w := makeWallet(t)
	e, _ := entity.NewLedgerEntry("e1", "wrong-wallet", "tx1", entity.DirectionCredit, 100, "USD", "", now)
	_, err := w.ComputeBalance([]entity.LedgerEntry{*e})
	if err == nil {
		t.Fatal("expected error for wallet_id mismatch")
	}
}

func TestComputeBalance_CurrencyMismatch(t *testing.T) {
	w := makeWallet(t)
	e, _ := entity.NewLedgerEntry("e1", "w1", "tx1", entity.DirectionCredit, 100, "EUR", "", now)
	_, err := w.ComputeBalance([]entity.LedgerEntry{*e})
	if err == nil {
		t.Fatal("expected error for currency mismatch")
	}
}
