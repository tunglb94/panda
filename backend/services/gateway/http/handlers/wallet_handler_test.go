package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	driverapp "github.com/fairride/driver/app"
	driverentity "github.com/fairride/driver/domain/entity"
	"github.com/fairride/gateway/http/handlers"
	"github.com/fairride/identity/infrastructure/jwt"
	domainerrors "github.com/fairride/shared/errors"
	walletapp "github.com/fairride/wallet/app"
	walletentity "github.com/fairride/wallet/domain/entity"
	walletrepo "github.com/fairride/wallet/domain/repository"
)

// ─── in-memory wallet fakes (mirrors wallet/app's own test fakes) ──────────

type fakeWalletRepo struct {
	byOwner map[string]*walletentity.Wallet
}

func newFakeWalletRepo() *fakeWalletRepo {
	return &fakeWalletRepo{byOwner: map[string]*walletentity.Wallet{}}
}
func (r *fakeWalletRepo) Save(_ context.Context, w *walletentity.Wallet) error {
	r.byOwner[w.OwnerID] = w
	return nil
}
func (r *fakeWalletRepo) FindByID(_ context.Context, id string) (*walletentity.Wallet, error) {
	for _, w := range r.byOwner {
		if w.WalletID == id {
			return w, nil
		}
	}
	return nil, domainerrors.NotFound("wallet not found")
}
func (r *fakeWalletRepo) FindByOwnerID(_ context.Context, ownerID string) (*walletentity.Wallet, error) {
	w, ok := r.byOwner[ownerID]
	if !ok {
		return nil, domainerrors.NotFound("wallet not found")
	}
	return w, nil
}

type fakeLedgerRepo struct {
	entries []walletentity.LedgerEntry
	// wallets/txs are set post-construction (see newWalletTestDeps) so
	// ListOutstandingDrivers can do the same wallet_id -> owner_id / type
	// join the real Postgres query does, instead of always returning empty.
	wallets *fakeWalletRepo
	txs     *fakeWalletTxRepo
}

func newFakeLedgerRepo() *fakeLedgerRepo { return &fakeLedgerRepo{} }
func (r *fakeLedgerRepo) Save(_ context.Context, e *walletentity.LedgerEntry) error {
	r.entries = append(r.entries, *e)
	return nil
}
func (r *fakeLedgerRepo) FindByWalletID(_ context.Context, walletID string) ([]walletentity.LedgerEntry, error) {
	var out []walletentity.LedgerEntry
	for _, e := range r.entries {
		if e.WalletID == walletID {
			out = append(out, e)
		}
	}
	return out, nil
}
func (r *fakeLedgerRepo) FindByTransactionID(_ context.Context, txID string) ([]walletentity.LedgerEntry, error) {
	var out []walletentity.LedgerEntry
	for _, e := range r.entries {
		if e.TransactionID == txID {
			out = append(out, e)
		}
	}
	return out, nil
}
func (r *fakeLedgerRepo) ListOutstandingDrivers(_ context.Context, limit int) ([]walletrepo.OutstandingDriver, error) {
	if r.wallets == nil || r.txs == nil {
		return nil, nil
	}
	byWallet := map[string]*walletentity.Wallet{}
	for _, w := range r.wallets.byOwner {
		byWallet[w.WalletID] = w
	}
	net := map[string]int64{}
	currency := map[string]string{}
	for _, e := range r.entries {
		w, ok := byWallet[e.WalletID]
		if !ok || w.WalletType != walletentity.WalletTypeDriver {
			continue
		}
		tx, ok := r.txs.byID[e.TransactionID]
		if !ok || tx.Type != walletentity.TypeCommission {
			continue
		}
		if e.Direction == walletentity.DirectionDebit {
			net[w.OwnerID] += e.AmountCents
		} else {
			net[w.OwnerID] -= e.AmountCents
		}
		currency[w.OwnerID] = w.Currency
	}
	var out []walletrepo.OutstandingDriver
	for ownerID, amount := range net {
		if amount > 0 {
			out = append(out, walletrepo.OutstandingDriver{DriverID: ownerID, OutstandingCents: amount, Currency: currency[ownerID]})
		}
	}
	return out, nil
}

type fakeWalletTxRepo struct {
	byID map[string]*walletentity.Transaction
}

func newFakeWalletTxRepo() *fakeWalletTxRepo {
	return &fakeWalletTxRepo{byID: map[string]*walletentity.Transaction{}}
}
func (r *fakeWalletTxRepo) Save(_ context.Context, tx *walletentity.Transaction) error {
	r.byID[tx.TransactionID] = tx
	return nil
}
func (r *fakeWalletTxRepo) FindByID(_ context.Context, id string) (*walletentity.Transaction, error) {
	tx, ok := r.byID[id]
	if !ok {
		return nil, domainerrors.NotFound("transaction not found")
	}
	return tx, nil
}
func (r *fakeWalletTxRepo) FindByReferenceID(_ context.Context, refID string) ([]*walletentity.Transaction, error) {
	var out []*walletentity.Transaction
	for _, tx := range r.byID {
		if tx.ReferenceID == refID {
			out = append(out, tx)
		}
	}
	return out, nil
}
func (r *fakeWalletTxRepo) FindByIDs(_ context.Context, ids []string) (map[string]*walletentity.Transaction, error) {
	out := map[string]*walletentity.Transaction{}
	for _, id := range ids {
		if tx, ok := r.byID[id]; ok {
			out[id] = tx
		}
	}
	return out, nil
}

type fakeSettlementRepo struct {
	byTripID map[string]*walletentity.Settlement
}

func newFakeSettlementRepo() *fakeSettlementRepo {
	return &fakeSettlementRepo{byTripID: map[string]*walletentity.Settlement{}}
}
func (r *fakeSettlementRepo) Save(_ context.Context, s *walletentity.Settlement) error {
	if _, ok := r.byTripID[s.TripID]; ok {
		return domainerrors.AlreadyExists("settlement already exists")
	}
	r.byTripID[s.TripID] = s
	return nil
}
func (r *fakeSettlementRepo) FindByTripID(_ context.Context, tripID string) (*walletentity.Settlement, error) {
	s, ok := r.byTripID[tripID]
	if !ok {
		return nil, domainerrors.NotFound("settlement not found")
	}
	return s, nil
}
func (r *fakeSettlementRepo) MarkPosted(_ context.Context, settlementID string) error {
	for _, s := range r.byTripID {
		if s.SettlementID == settlementID {
			s.Status = walletentity.SettlementStatusPosted
			return nil
		}
	}
	return domainerrors.NotFound("settlement not found")
}
func (r *fakeSettlementRepo) FindByID(_ context.Context, id string) (*walletentity.Settlement, error) {
	for _, s := range r.byTripID {
		if s.SettlementID == id {
			return s, nil
		}
	}
	return nil, domainerrors.NotFound("settlement not found")
}
func (r *fakeSettlementRepo) ListByDriverID(_ context.Context, driverID string, from, to int64, limit int) ([]*walletentity.Settlement, error) {
	var out []*walletentity.Settlement
	for _, s := range r.byTripID {
		if s.DriverID == driverID {
			out = append(out, s)
		}
	}
	return out, nil
}
func (r *fakeSettlementRepo) ListAll(_ context.Context, driverID string, from, to int64, limit int) ([]*walletentity.Settlement, error) {
	var out []*walletentity.Settlement
	for _, s := range r.byTripID {
		if driverID == "" || s.DriverID == driverID {
			out = append(out, s)
		}
	}
	return out, nil
}

type fakeBankAccountRepo struct {
	byDriverID map[string]*walletentity.BankAccount
}

func newFakeBankAccountRepo() *fakeBankAccountRepo {
	return &fakeBankAccountRepo{byDriverID: map[string]*walletentity.BankAccount{}}
}
func (r *fakeBankAccountRepo) Save(_ context.Context, b *walletentity.BankAccount) error {
	r.byDriverID[b.DriverID] = b
	return nil
}
func (r *fakeBankAccountRepo) FindByDriverID(_ context.Context, driverID string) (*walletentity.BankAccount, error) {
	b, ok := r.byDriverID[driverID]
	if !ok {
		return nil, domainerrors.NotFound("bank account not found")
	}
	return b, nil
}

type fakePayoutRequestRepo struct {
	byID map[string]*walletentity.PayoutRequest
}

func newFakePayoutRequestRepo() *fakePayoutRequestRepo {
	return &fakePayoutRequestRepo{byID: map[string]*walletentity.PayoutRequest{}}
}
func (r *fakePayoutRequestRepo) Save(_ context.Context, p *walletentity.PayoutRequest) error {
	r.byID[p.PayoutRequestID] = p
	return nil
}
func (r *fakePayoutRequestRepo) FindByID(_ context.Context, id string) (*walletentity.PayoutRequest, error) {
	p, ok := r.byID[id]
	if !ok {
		return nil, domainerrors.NotFound("payout request not found")
	}
	return p, nil
}
func (r *fakePayoutRequestRepo) FindInFlightByDriverID(_ context.Context, driverID string) (*walletentity.PayoutRequest, error) {
	for _, p := range r.byID {
		if p.DriverID == driverID && p.IsInFlight() {
			return p, nil
		}
	}
	return nil, domainerrors.NotFound("no in-flight payout")
}
func (r *fakePayoutRequestRepo) ListByDriverID(_ context.Context, driverID string, limit int) ([]*walletentity.PayoutRequest, error) {
	var out []*walletentity.PayoutRequest
	for _, p := range r.byID {
		if p.DriverID == driverID {
			out = append(out, p)
		}
	}
	return out, nil
}
func (r *fakePayoutRequestRepo) ListByFilter(_ context.Context, status walletentity.PayoutStatus, driverID string, limit int) ([]*walletentity.PayoutRequest, error) {
	var out []*walletentity.PayoutRequest
	for _, p := range r.byID {
		if status != "" && p.Status != status {
			continue
		}
		if driverID != "" && p.DriverID != driverID {
			continue
		}
		out = append(out, p)
	}
	return out, nil
}

type fakeWalletAuditRepo struct{ entries []*walletentity.AuditLog }

func (r *fakeWalletAuditRepo) Save(_ context.Context, log *walletentity.AuditLog) error {
	r.entries = append(r.entries, log)
	return nil
}
func (r *fakeWalletAuditRepo) ListByDriverID(_ context.Context, driverID string, limit int) ([]*walletentity.AuditLog, error) {
	return nil, nil
}

// ─── builder ─────────────────────────────────────────────────────────────────

type walletTestDeps struct {
	wallets     *fakeWalletRepo
	ledger      *fakeLedgerRepo
	tx          *fakeWalletTxRepo
	settlements *fakeSettlementRepo
	banks       *fakeBankAccountRepo
	payouts     *fakePayoutRequestRepo
	audit       *fakeWalletAuditRepo
}

func buildWalletHandler(deps *walletTestDeps, kyc *driverapp.GetDriverVerificationUseCase) *handlers.WalletHandler {
	summary := walletapp.NewGetWalletSummaryUseCase(deps.wallets, deps.ledger, deps.tx, deps.payouts)
	return handlers.NewWalletHandler(
		summary,
		walletapp.NewGetStatementUseCase(deps.settlements, summary),
		walletapp.NewListWalletTransactionsUseCase(deps.wallets, deps.ledger, deps.tx),
		walletapp.NewGetBankAccountUseCase(deps.banks),
		walletapp.NewSetBankAccountUseCase(deps.banks, deps.audit),
		walletapp.NewCreatePayoutRequestUseCase(deps.payouts, deps.banks, summary, walletapp.NewGetOrCreateWalletUseCase(deps.wallets), deps.ledger, deps.tx, deps.audit, 0),
		walletapp.NewListMyPayoutRequestsUseCase(deps.payouts),
		kyc,
	)
}

func newWalletTestDeps() *walletTestDeps {
	wallets := newFakeWalletRepo()
	txs := newFakeWalletTxRepo()
	ledger := newFakeLedgerRepo()
	ledger.wallets = wallets
	ledger.txs = txs
	return &walletTestDeps{
		wallets: wallets, ledger: ledger, tx: txs,
		settlements: newFakeSettlementRepo(), banks: newFakeBankAccountRepo(), payouts: newFakePayoutRequestRepo(),
		audit: &fakeWalletAuditRepo{},
	}
}

func settleTripForTest(t *testing.T, deps *walletTestDeps, tripID, driverID string, fareCents int64, method walletentity.PaymentMethod) {
	t.Helper()
	getOrCreate := walletapp.NewGetOrCreateWalletUseCase(deps.wallets)
	engine := walletapp.NewCreateSettlementUseCase(deps.settlements, getOrCreate, deps.ledger, deps.tx, deps.audit, 0.20)
	_, err := engine.Execute(context.Background(), walletapp.CreateSettlementInput{
		TripID: tripID, DriverID: driverID, TripType: walletentity.TripTypeRide,
		PaymentMethod: method, FareAmountCents: fareCents, Currency: "VND",
	})
	if err != nil {
		t.Fatalf("settle: %v", err)
	}
}

func walletRequest(t *testing.T, method, path string, driverID string, body any) *http.Request {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&buf).Encode(body)
	}
	r := httptest.NewRequest(method, path, &buf)
	r.Header.Set("Content-Type", "application/json")
	return injectClaims(r, &jwt.AccessClaims{UserID: driverID, UserType: "driver"})
}

// ─── tests ────────────────────────────────────────────────────────────────────

func TestWalletHandler_ServiceUnavailableWhenNotConfigured(t *testing.T) {
	h := handlers.NewWalletHandler(nil, nil, nil, nil, nil, nil, nil, nil)
	w := httptest.NewRecorder()
	h.GetSummary(w, walletRequest(t, http.MethodGet, "/api/v1/driver/wallet/summary", "d1", nil))
	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("status = %d, want 503", w.Code)
	}
}

func TestWalletHandler_GetSummary_CashAndWalletMix(t *testing.T) {
	deps := newWalletTestDeps()
	settleTripForTest(t, deps, "trip-cash", "d1", 120_000, walletentity.PaymentMethodCash)
	settleTripForTest(t, deps, "trip-wallet", "d1", 100_000, walletentity.PaymentMethodWallet)
	h := buildWalletHandler(deps, nil)

	w := httptest.NewRecorder()
	h.GetSummary(w, walletRequest(t, http.MethodGet, "/api/v1/driver/wallet/summary", "d1", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200, body=%s", w.Code, w.Body.String())
	}
	var body map[string]any
	_ = json.NewDecoder(w.Body).Decode(&body)
	if body["outstanding_cents"] != float64(24_000) {
		t.Errorf("outstanding = %v, want 24000", body["outstanding_cents"])
	}
	if body["available_cents"] != float64(80_000) {
		t.Errorf("available = %v, want 80000", body["available_cents"])
	}
}

func TestWalletHandler_GetSummary_OnlySeesOwnDriverData(t *testing.T) {
	deps := newWalletTestDeps()
	settleTripForTest(t, deps, "trip-1", "d1", 500_000, walletentity.PaymentMethodWallet)
	settleTripForTest(t, deps, "trip-2", "d2", 900_000, walletentity.PaymentMethodWallet)
	h := buildWalletHandler(deps, nil)

	w := httptest.NewRecorder()
	h.GetSummary(w, walletRequest(t, http.MethodGet, "/api/v1/driver/wallet/summary", "d1", nil))
	var body map[string]any
	_ = json.NewDecoder(w.Body).Decode(&body)
	if body["available_cents"] != float64(400_000) {
		t.Errorf("d1 available = %v, want 400000 (d1's own trip only, not d2's)", body["available_cents"])
	}
}

func TestWalletHandler_SetAndGetBankAccount_NeverExposesRawNumber(t *testing.T) {
	deps := newWalletTestDeps()
	h := buildWalletHandler(deps, nil)

	w := httptest.NewRecorder()
	h.SetBankAccount(w, walletRequest(t, http.MethodPost, "/api/v1/driver/wallet/bank-account", "d1", map[string]any{
		"bank_name": "Vietcombank", "account_holder_name": "Nguyen Van A", "account_number": "0123456789",
	}))
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200, body=%s", w.Code, w.Body.String())
	}
	var body map[string]any
	_ = json.NewDecoder(w.Body).Decode(&body)
	if body["masked_account_number"] != "••••6789" {
		t.Errorf("masked_account_number = %v, want ••••6789", body["masked_account_number"])
	}
	raw, _ := json.Marshal(body)
	if bytes.Contains(raw, []byte("0123456789")) {
		t.Error("response must never contain the raw account number")
	}
}

func TestWalletHandler_CreatePayoutRequest_RequiresKYCApproved(t *testing.T) {
	deps := newWalletTestDeps()
	settleTripForTest(t, deps, "trip-1", "d1", 1_000_000, walletentity.PaymentMethodWallet)
	_ = deps.banks.Save(context.Background(), mustBankAccount(t, "d1"))

	driverRepo := &fakeDVRepoForWallet{status: driverentity.KYCPending}
	kyc := driverapp.NewGetDriverVerificationUseCase(driverRepo)
	h := buildWalletHandler(deps, kyc)

	w := httptest.NewRecorder()
	h.CreatePayoutRequest(w, walletRequest(t, http.MethodPost, "/api/v1/driver/wallet/payouts", "d1", map[string]any{"amount_cents": 100_000}))
	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("status = %d, want 422 when KYC not approved, body=%s", w.Code, w.Body.String())
	}
}

func TestWalletHandler_CreatePayoutRequest_SucceedsWhenKYCApproved(t *testing.T) {
	deps := newWalletTestDeps()
	settleTripForTest(t, deps, "trip-1", "d1", 1_000_000, walletentity.PaymentMethodWallet)
	_ = deps.banks.Save(context.Background(), mustBankAccount(t, "d1"))

	driverRepo := &fakeDVRepoForWallet{status: driverentity.KYCApproved}
	kyc := driverapp.NewGetDriverVerificationUseCase(driverRepo)
	h := buildWalletHandler(deps, kyc)

	w := httptest.NewRecorder()
	h.CreatePayoutRequest(w, walletRequest(t, http.MethodPost, "/api/v1/driver/wallet/payouts", "d1", map[string]any{"amount_cents": 100_000}))
	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201, body=%s", w.Code, w.Body.String())
	}
}

func mustBankAccount(t *testing.T, driverID string) *walletentity.BankAccount {
	t.Helper()
	b, err := walletentity.NewBankAccount("bank-1", driverID, "Vietcombank", "Nguyen Van A", "0123456789", "", time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}
	return b
}

// fakeDVRepoForWallet is a minimal DriverVerificationRepository stub scoped
// to this test file only (the KYC phase's own fakes live in a different
// package and aren't exported).
type fakeDVRepoForWallet struct {
	status driverentity.KYCStatus
}

func (r *fakeDVRepoForWallet) Save(context.Context, *driverentity.DriverVerification) error {
	return nil
}
func (r *fakeDVRepoForWallet) FindByDriverID(_ context.Context, driverID string) (*driverentity.DriverVerification, error) {
	return &driverentity.DriverVerification{DriverID: driverID, Status: r.status}, nil
}
func (r *fakeDVRepoForWallet) FindByNationalIDNumber(context.Context, string) (*driverentity.DriverVerification, error) {
	return nil, domainerrors.NotFound("not found")
}
func (r *fakeDVRepoForWallet) FindByLicenseNumber(context.Context, string) (*driverentity.DriverVerification, error) {
	return nil, domainerrors.NotFound("not found")
}
func (r *fakeDVRepoForWallet) ListByStatus(context.Context, driverentity.KYCStatus, int) ([]*driverentity.DriverVerification, error) {
	return nil, nil
}

func (r *fakeDVRepoForWallet) CountByStatus(context.Context, driverentity.KYCStatus) (int, error) {
	return 0, nil
}
