package app_test

import (
	"context"
	"testing"
	"time"

	domainerrors "github.com/fairride/shared/errors"
	"github.com/fairride/wallet/app"
	"github.com/fairride/wallet/domain/entity"
	"github.com/fairride/wallet/domain/repository"
)

var (
	ctx = context.Background()
	now = time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
)

// ─── stub WalletRepository ───────────────────────────────────────────────────

type stubWalletRepo struct {
	byID    map[string]*entity.Wallet
	byOwner map[string]*entity.Wallet
}

var _ repository.WalletRepository = (*stubWalletRepo)(nil)

func newWalletRepo() *stubWalletRepo {
	return &stubWalletRepo{
		byID:    make(map[string]*entity.Wallet),
		byOwner: make(map[string]*entity.Wallet),
	}
}

func (r *stubWalletRepo) Save(_ context.Context, w *entity.Wallet) error {
	r.byID[w.WalletID] = w
	r.byOwner[w.OwnerID] = w
	return nil
}

func (r *stubWalletRepo) FindByID(_ context.Context, id string) (*entity.Wallet, error) {
	w, ok := r.byID[id]
	if !ok {
		return nil, domainerrors.NotFound("wallet not found: " + id)
	}
	return w, nil
}

func (r *stubWalletRepo) FindByOwnerID(_ context.Context, ownerID string) (*entity.Wallet, error) {
	w, ok := r.byOwner[ownerID]
	if !ok {
		return nil, domainerrors.NotFound("wallet not found for owner: " + ownerID)
	}
	return w, nil
}

// ─── stub LedgerEntryRepository ─────────────────────────────────────────────

type stubLedgerRepo struct {
	entries []entity.LedgerEntry
}

var _ repository.LedgerEntryRepository = (*stubLedgerRepo)(nil)

func newLedgerRepo() *stubLedgerRepo { return &stubLedgerRepo{} }

func (r *stubLedgerRepo) Save(_ context.Context, e *entity.LedgerEntry) error {
	for _, existing := range r.entries {
		if existing.EntryID == e.EntryID {
			return domainerrors.AlreadyExists("ledger entry already exists")
		}
	}
	r.entries = append(r.entries, *e)
	return nil
}

func (r *stubLedgerRepo) FindByWalletID(_ context.Context, walletID string) ([]entity.LedgerEntry, error) {
	var out []entity.LedgerEntry
	for _, e := range r.entries {
		if e.WalletID == walletID {
			out = append(out, e)
		}
	}
	return out, nil
}

func (r *stubLedgerRepo) FindByTransactionID(_ context.Context, txID string) ([]entity.LedgerEntry, error) {
	var out []entity.LedgerEntry
	for _, e := range r.entries {
		if e.TransactionID == txID {
			out = append(out, e)
		}
	}
	return out, nil
}

func (r *stubLedgerRepo) ListOutstandingDrivers(_ context.Context, limit int) ([]repository.OutstandingDriver, error) {
	return nil, nil
}

// ─── stub TransactionRepository ─────────────────────────────────────────────

type stubTxRepo struct {
	byID  map[string]*entity.Transaction
	byRef map[string][]*entity.Transaction
}

var _ repository.TransactionRepository = (*stubTxRepo)(nil)

func newTxRepo() *stubTxRepo {
	return &stubTxRepo{
		byID:  make(map[string]*entity.Transaction),
		byRef: make(map[string][]*entity.Transaction),
	}
}

func (r *stubTxRepo) Save(_ context.Context, tx *entity.Transaction) error {
	if _, ok := r.byID[tx.TransactionID]; ok {
		return domainerrors.AlreadyExists("transaction already exists")
	}
	r.byID[tx.TransactionID] = tx
	if tx.ReferenceID != "" {
		r.byRef[tx.ReferenceID] = append(r.byRef[tx.ReferenceID], tx)
	}
	return nil
}

func (r *stubTxRepo) FindByID(_ context.Context, id string) (*entity.Transaction, error) {
	tx, ok := r.byID[id]
	if !ok {
		return nil, domainerrors.NotFound("transaction not found: " + id)
	}
	return tx, nil
}

func (r *stubTxRepo) FindByIDs(_ context.Context, ids []string) (map[string]*entity.Transaction, error) {
	out := map[string]*entity.Transaction{}
	for _, id := range ids {
		if tx, ok := r.byID[id]; ok {
			out[id] = tx
		}
	}
	return out, nil
}

func (r *stubTxRepo) FindByReferenceID(_ context.Context, refID string) ([]*entity.Transaction, error) {
	return r.byRef[refID], nil
}

// ─── helpers ─────────────────────────────────────────────────────────────────

func seedWallet(t *testing.T, repo *stubWalletRepo, walletID, ownerID string, wt entity.WalletType) *entity.Wallet {
	t.Helper()
	w, err := entity.NewWallet(walletID, ownerID, wt, "USD", now)
	if err != nil {
		t.Fatal(err)
	}
	_ = repo.Save(ctx, w)
	return w
}

func seedEntry(t *testing.T, repo *stubLedgerRepo, walletID, txID string, dir entity.EntryDirection, cents int64) {
	t.Helper()
	e, err := entity.NewLedgerEntry("e-"+walletID+"-"+txID, walletID, txID, dir, cents, "USD", "", now)
	if err != nil {
		t.Fatal(err)
	}
	_ = repo.Save(ctx, e)
}

func seedTransaction(t *testing.T, repo *stubTxRepo, txID, refID string, txType entity.TransactionType) *entity.Transaction {
	t.Helper()
	tx, err := entity.NewTransaction(txID, txType, refID, "", "USD", "", now)
	if err != nil {
		t.Fatal(err)
	}
	_ = repo.Save(ctx, tx)
	return tx
}

// ─── GetWalletUseCase ────────────────────────────────────────────────────────

func TestGetWallet_Found(t *testing.T) {
	wr := newWalletRepo()
	seedWallet(t, wr, "w1", "rider-1", entity.WalletTypeRider)

	uc := app.NewGetWalletUseCase(wr)
	w, err := uc.Execute(ctx, "rider-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if w.WalletID != "w1" {
		t.Fatalf("expected w1, got %s", w.WalletID)
	}
}

func TestGetWallet_NotFound(t *testing.T) {
	uc := app.NewGetWalletUseCase(newWalletRepo())
	_, err := uc.Execute(ctx, "nobody")
	if err == nil {
		t.Fatal("expected error")
	}
	de, ok := err.(*domainerrors.DomainError)
	if !ok || de.Code != domainerrors.CodeNotFound {
		t.Fatalf("expected CodeNotFound, got %v", err)
	}
}

func TestGetWallet_EmptyOwnerID(t *testing.T) {
	uc := app.NewGetWalletUseCase(newWalletRepo())
	_, err := uc.Execute(ctx, "")
	if err == nil {
		t.Fatal("expected error")
	}
	de, ok := err.(*domainerrors.DomainError)
	if !ok || de.Code != domainerrors.CodeInvalidArgument {
		t.Fatalf("expected CodeInvalidArgument, got %v", err)
	}
}

// ─── GetBalanceUseCase ───────────────────────────────────────────────────────

func TestGetBalance_EmptyLedger(t *testing.T) {
	wr := newWalletRepo()
	lr := newLedgerRepo()
	seedWallet(t, wr, "w1", "rider-1", entity.WalletTypeRider)

	uc := app.NewGetBalanceUseCase(wr, lr)
	result, err := uc.Execute(ctx, "rider-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.BalanceCents != 0 {
		t.Fatalf("expected 0 balance, got %d", result.BalanceCents)
	}
	if result.Currency != "USD" {
		t.Fatalf("wrong currency: %s", result.Currency)
	}
}

func TestGetBalance_WithEntries(t *testing.T) {
	wr := newWalletRepo()
	lr := newLedgerRepo()
	seedWallet(t, wr, "w1", "driver-1", entity.WalletTypeDriver)
	seedEntry(t, lr, "w1", "tx1", entity.DirectionCredit, 1500)
	seedEntry(t, lr, "w1", "tx2", entity.DirectionDebit, 300)

	uc := app.NewGetBalanceUseCase(wr, lr)
	result, err := uc.Execute(ctx, "driver-1")
	if err != nil {
		t.Fatal(err)
	}
	if result.BalanceCents != 1200 {
		t.Fatalf("expected 1200, got %d", result.BalanceCents)
	}
}

func TestGetBalance_OwnerNotFound(t *testing.T) {
	uc := app.NewGetBalanceUseCase(newWalletRepo(), newLedgerRepo())
	_, err := uc.Execute(ctx, "ghost")
	if err == nil {
		t.Fatal("expected error")
	}
	de := err.(*domainerrors.DomainError)
	if de.Code != domainerrors.CodeNotFound {
		t.Fatalf("expected CodeNotFound, got %s", de.Code)
	}
}

func TestGetBalance_EmptyOwnerID(t *testing.T) {
	uc := app.NewGetBalanceUseCase(newWalletRepo(), newLedgerRepo())
	_, err := uc.Execute(ctx, "")
	if err == nil {
		t.Fatal("expected error")
	}
	de := err.(*domainerrors.DomainError)
	if de.Code != domainerrors.CodeInvalidArgument {
		t.Fatalf("expected CodeInvalidArgument, got %s", de.Code)
	}
}

// ─── GetLedgerUseCase ────────────────────────────────────────────────────────

func TestGetLedger_ReturnsEntries(t *testing.T) {
	wr := newWalletRepo()
	lr := newLedgerRepo()
	seedWallet(t, wr, "w1", "rider-1", entity.WalletTypeRider)
	seedEntry(t, lr, "w1", "tx1", entity.DirectionCredit, 500)
	seedEntry(t, lr, "w1", "tx2", entity.DirectionDebit, 200)

	uc := app.NewGetLedgerUseCase(wr, lr)
	entries, err := uc.Execute(ctx, "w1")
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
}

func TestGetLedger_WalletNotFound(t *testing.T) {
	uc := app.NewGetLedgerUseCase(newWalletRepo(), newLedgerRepo())
	_, err := uc.Execute(ctx, "missing-wallet")
	if err == nil {
		t.Fatal("expected error")
	}
	de := err.(*domainerrors.DomainError)
	if de.Code != domainerrors.CodeNotFound {
		t.Fatalf("expected CodeNotFound, got %s", de.Code)
	}
}

func TestGetLedger_EmptyWalletID(t *testing.T) {
	uc := app.NewGetLedgerUseCase(newWalletRepo(), newLedgerRepo())
	_, err := uc.Execute(ctx, "")
	if err == nil {
		t.Fatal("expected error")
	}
	de := err.(*domainerrors.DomainError)
	if de.Code != domainerrors.CodeInvalidArgument {
		t.Fatalf("expected CodeInvalidArgument, got %s", de.Code)
	}
}

func TestGetLedger_EmptyForNewWallet(t *testing.T) {
	wr := newWalletRepo()
	lr := newLedgerRepo()
	seedWallet(t, wr, "w1", "rider-1", entity.WalletTypeRider)

	uc := app.NewGetLedgerUseCase(wr, lr)
	entries, err := uc.Execute(ctx, "w1")
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatalf("expected empty ledger, got %d entries", len(entries))
	}
}

// ─── GetTransactionUseCase ───────────────────────────────────────────────────

func TestGetTransaction_Found(t *testing.T) {
	tr := newTxRepo()
	lr := newLedgerRepo()
	seedTransaction(t, tr, "tx1", "trip-abc", entity.TypeTripPayment)
	seedEntry(t, lr, "w1", "tx1", entity.DirectionDebit, 1000)
	seedEntry(t, lr, "w2", "tx1", entity.DirectionCredit, 850)

	uc := app.NewGetTransactionUseCase(tr, lr)
	result, err := uc.Execute(ctx, "tx1")
	if err != nil {
		t.Fatal(err)
	}
	if result.Transaction.TransactionID != "tx1" {
		t.Fatal("wrong transaction id")
	}
	if len(result.Entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(result.Entries))
	}
}

func TestGetTransaction_NotFound(t *testing.T) {
	uc := app.NewGetTransactionUseCase(newTxRepo(), newLedgerRepo())
	_, err := uc.Execute(ctx, "missing-tx")
	if err == nil {
		t.Fatal("expected error")
	}
	de := err.(*domainerrors.DomainError)
	if de.Code != domainerrors.CodeNotFound {
		t.Fatalf("expected CodeNotFound, got %s", de.Code)
	}
}

func TestGetTransaction_EmptyID(t *testing.T) {
	uc := app.NewGetTransactionUseCase(newTxRepo(), newLedgerRepo())
	_, err := uc.Execute(ctx, "")
	if err == nil {
		t.Fatal("expected error")
	}
	de := err.(*domainerrors.DomainError)
	if de.Code != domainerrors.CodeInvalidArgument {
		t.Fatalf("expected CodeInvalidArgument, got %s", de.Code)
	}
}

func TestGetTransaction_NoEntriesYet(t *testing.T) {
	tr := newTxRepo()
	lr := newLedgerRepo()
	seedTransaction(t, tr, "tx1", "", entity.TypeAdjustment)

	uc := app.NewGetTransactionUseCase(tr, lr)
	result, err := uc.Execute(ctx, "tx1")
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Entries) != 0 {
		t.Fatalf("expected 0 entries, got %d", len(result.Entries))
	}
}
