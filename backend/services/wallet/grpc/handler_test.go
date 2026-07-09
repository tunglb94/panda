package grpc_test

import (
	"context"
	"testing"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/fairride/wallet/app"
	"github.com/fairride/wallet/domain/entity"
	"github.com/fairride/wallet/domain/repository"
	walletgrpc "github.com/fairride/wallet/grpc"
	"github.com/fairride/wallet/grpc/walletpb"
	domainerrors "github.com/fairride/shared/errors"
)

var (
	ctx = context.Background()
	now = time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
)

// ─── stubs (reuse from app_test patterns) ───────────────────────────────────

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
		return nil, domainerrors.NotFound("wallet not found")
	}
	return w, nil
}
func (r *stubWalletRepo) FindByOwnerID(_ context.Context, ownerID string) (*entity.Wallet, error) {
	w, ok := r.byOwner[ownerID]
	if !ok {
		return nil, domainerrors.NotFound("wallet not found")
	}
	return w, nil
}

type stubLedgerRepo struct{ entries []entity.LedgerEntry }

var _ repository.LedgerEntryRepository = (*stubLedgerRepo)(nil)

func newLedgerRepo() *stubLedgerRepo { return &stubLedgerRepo{} }
func (r *stubLedgerRepo) Save(_ context.Context, e *entity.LedgerEntry) error {
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

type stubTxRepo struct {
	byID map[string]*entity.Transaction
}

var _ repository.TransactionRepository = (*stubTxRepo)(nil)

func newTxRepo() *stubTxRepo { return &stubTxRepo{byID: make(map[string]*entity.Transaction)} }
func (r *stubTxRepo) Save(_ context.Context, tx *entity.Transaction) error {
	r.byID[tx.TransactionID] = tx
	return nil
}
func (r *stubTxRepo) FindByID(_ context.Context, id string) (*entity.Transaction, error) {
	tx, ok := r.byID[id]
	if !ok {
		return nil, domainerrors.NotFound("transaction not found")
	}
	return tx, nil
}
func (r *stubTxRepo) FindByReferenceID(_ context.Context, _ string) ([]*entity.Transaction, error) {
	return nil, nil
}

// ─── handler factory ─────────────────────────────────────────────────────────

func newHandler(wr *stubWalletRepo, lr *stubLedgerRepo, tr *stubTxRepo) *walletgrpc.Handler {
	return walletgrpc.NewHandler(
		app.NewGetWalletUseCase(wr),
		app.NewGetBalanceUseCase(wr, lr),
		app.NewGetLedgerUseCase(wr, lr),
		app.NewGetTransactionUseCase(tr, lr),
	)
}

func seedWallet(t *testing.T, r *stubWalletRepo) *entity.Wallet {
	t.Helper()
	w, err := entity.NewWallet("w1", "owner-1", entity.WalletTypeRider, "USD", now)
	if err != nil {
		t.Fatal(err)
	}
	_ = r.Save(ctx, w)
	return w
}

func seedEntry(t *testing.T, r *stubLedgerRepo, walletID, txID string, dir entity.EntryDirection, cents int64) {
	t.Helper()
	e, err := entity.NewLedgerEntry("e-"+txID, walletID, txID, dir, cents, "USD", "", now)
	if err != nil {
		t.Fatal(err)
	}
	_ = r.Save(ctx, e)
}

func seedTx(t *testing.T, r *stubTxRepo) *entity.Transaction {
	t.Helper()
	tx, err := entity.NewTransaction("tx1", entity.TypeTripPayment, "trip-1", "USD", "", now)
	if err != nil {
		t.Fatal(err)
	}
	_ = r.Save(ctx, tx)
	return tx
}

// ─── GetWallet ───────────────────────────────────────────────────────────────

func TestGetWallet_OK(t *testing.T) {
	wr, lr, tr := newWalletRepo(), newLedgerRepo(), newTxRepo()
	seedWallet(t, wr)
	h := newHandler(wr, lr, tr)

	resp, err := h.GetWallet(ctx, &walletpb.GetWalletRequest{OwnerId: "owner-1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Wallet.WalletId != "w1" {
		t.Fatalf("expected w1, got %s", resp.Wallet.WalletId)
	}
	if resp.Wallet.WalletType != "rider" {
		t.Fatalf("expected rider, got %s", resp.Wallet.WalletType)
	}
	if resp.Wallet.Currency != "USD" {
		t.Fatal("wrong currency")
	}
}

func TestGetWallet_NotFound(t *testing.T) {
	h := newHandler(newWalletRepo(), newLedgerRepo(), newTxRepo())
	_, err := h.GetWallet(ctx, &walletpb.GetWalletRequest{OwnerId: "ghost"})
	if err == nil {
		t.Fatal("expected error")
	}
	if status.Code(err) != codes.NotFound {
		t.Fatalf("expected NotFound, got %v", status.Code(err))
	}
}

func TestGetWallet_EmptyOwnerID(t *testing.T) {
	h := newHandler(newWalletRepo(), newLedgerRepo(), newTxRepo())
	_, err := h.GetWallet(ctx, &walletpb.GetWalletRequest{OwnerId: ""})
	if err == nil {
		t.Fatal("expected error")
	}
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("expected InvalidArgument, got %v", status.Code(err))
	}
}

// ─── GetBalance ──────────────────────────────────────────────────────────────

func TestGetBalance_OK(t *testing.T) {
	wr, lr, tr := newWalletRepo(), newLedgerRepo(), newTxRepo()
	seedWallet(t, wr)
	seedEntry(t, lr, "w1", "tx1", entity.DirectionCredit, 2000)
	seedEntry(t, lr, "w1", "tx2", entity.DirectionDebit, 500)
	h := newHandler(wr, lr, tr)

	resp, err := h.GetBalance(ctx, &walletpb.GetBalanceRequest{OwnerId: "owner-1"})
	if err != nil {
		t.Fatal(err)
	}
	if resp.BalanceCents != 1500 {
		t.Fatalf("expected 1500, got %d", resp.BalanceCents)
	}
	if resp.Currency != "USD" {
		t.Fatal("wrong currency")
	}
}

func TestGetBalance_ZeroBalance(t *testing.T) {
	wr, lr, tr := newWalletRepo(), newLedgerRepo(), newTxRepo()
	seedWallet(t, wr)
	h := newHandler(wr, lr, tr)

	resp, err := h.GetBalance(ctx, &walletpb.GetBalanceRequest{OwnerId: "owner-1"})
	if err != nil {
		t.Fatal(err)
	}
	if resp.BalanceCents != 0 {
		t.Fatalf("expected 0, got %d", resp.BalanceCents)
	}
}

func TestGetBalance_NotFound(t *testing.T) {
	h := newHandler(newWalletRepo(), newLedgerRepo(), newTxRepo())
	_, err := h.GetBalance(ctx, &walletpb.GetBalanceRequest{OwnerId: "nobody"})
	if status.Code(err) != codes.NotFound {
		t.Fatalf("expected NotFound, got %v", err)
	}
}

// ─── GetLedger ───────────────────────────────────────────────────────────────

func TestGetLedger_OK(t *testing.T) {
	wr, lr, tr := newWalletRepo(), newLedgerRepo(), newTxRepo()
	seedWallet(t, wr)
	seedEntry(t, lr, "w1", "tx1", entity.DirectionCredit, 1000)
	seedEntry(t, lr, "w1", "tx2", entity.DirectionCredit, 500)
	h := newHandler(wr, lr, tr)

	resp, err := h.GetLedger(ctx, &walletpb.GetLedgerRequest{WalletId: "w1"})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(resp.Entries))
	}
}

func TestGetLedger_WalletNotFound(t *testing.T) {
	h := newHandler(newWalletRepo(), newLedgerRepo(), newTxRepo())
	_, err := h.GetLedger(ctx, &walletpb.GetLedgerRequest{WalletId: "missing"})
	if status.Code(err) != codes.NotFound {
		t.Fatalf("expected NotFound, got %v", err)
	}
}

func TestGetLedger_EmptyWalletID(t *testing.T) {
	h := newHandler(newWalletRepo(), newLedgerRepo(), newTxRepo())
	_, err := h.GetLedger(ctx, &walletpb.GetLedgerRequest{WalletId: ""})
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("expected InvalidArgument, got %v", err)
	}
}

// ─── GetTransaction ──────────────────────────────────────────────────────────

func TestGetTransaction_OK(t *testing.T) {
	wr, lr, tr := newWalletRepo(), newLedgerRepo(), newTxRepo()
	seedTx(t, tr)
	seedEntry(t, lr, "w1", "tx1", entity.DirectionDebit, 1000)
	seedEntry(t, lr, "w2", "tx1", entity.DirectionCredit, 850)
	h := newHandler(wr, lr, tr)

	resp, err := h.GetTransaction(ctx, &walletpb.GetTransactionRequest{TransactionId: "tx1"})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Transaction.TransactionId != "tx1" {
		t.Fatal("wrong transaction id")
	}
	if resp.Transaction.Type != "trip_payment" {
		t.Fatalf("wrong type: %s", resp.Transaction.Type)
	}
	if len(resp.Entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(resp.Entries))
	}
}

func TestGetTransaction_NotFound(t *testing.T) {
	h := newHandler(newWalletRepo(), newLedgerRepo(), newTxRepo())
	_, err := h.GetTransaction(ctx, &walletpb.GetTransactionRequest{TransactionId: "ghost"})
	if status.Code(err) != codes.NotFound {
		t.Fatalf("expected NotFound, got %v", err)
	}
}

func TestGetTransaction_EmptyID(t *testing.T) {
	h := newHandler(newWalletRepo(), newLedgerRepo(), newTxRepo())
	_, err := h.GetTransaction(ctx, &walletpb.GetTransactionRequest{TransactionId: ""})
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("expected InvalidArgument, got %v", err)
	}
}
