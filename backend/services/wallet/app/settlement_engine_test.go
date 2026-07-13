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

// ─── stub SettlementRepository ───────────────────────────────────────────────

type stubSettlementRepo struct {
	byTripID map[string]*entity.Settlement
}

var _ repository.SettlementRepository = (*stubSettlementRepo)(nil)

func newSettlementRepo() *stubSettlementRepo {
	return &stubSettlementRepo{byTripID: map[string]*entity.Settlement{}}
}

func (r *stubSettlementRepo) Save(_ context.Context, s *entity.Settlement) error {
	if _, ok := r.byTripID[s.TripID]; ok {
		return domainerrors.AlreadyExists("settlement already exists for this trip")
	}
	r.byTripID[s.TripID] = s
	return nil
}

func (r *stubSettlementRepo) FindByTripID(_ context.Context, tripID string) (*entity.Settlement, error) {
	s, ok := r.byTripID[tripID]
	if !ok {
		return nil, domainerrors.NotFound("settlement not found")
	}
	return s, nil
}

func (r *stubSettlementRepo) MarkPosted(_ context.Context, settlementID string) error {
	for _, s := range r.byTripID {
		if s.SettlementID == settlementID {
			s.Status = entity.SettlementStatusPosted
			return nil
		}
	}
	return domainerrors.NotFound("settlement not found")
}

func (r *stubSettlementRepo) FindByID(_ context.Context, settlementID string) (*entity.Settlement, error) {
	for _, s := range r.byTripID {
		if s.SettlementID == settlementID {
			return s, nil
		}
	}
	return nil, domainerrors.NotFound("settlement not found")
}

func (r *stubSettlementRepo) ListByDriverID(_ context.Context, driverID string, from, to int64, limit int) ([]*entity.Settlement, error) {
	var out []*entity.Settlement
	for _, s := range r.byTripID {
		if s.DriverID == driverID {
			out = append(out, s)
		}
	}
	return out, nil
}

func (r *stubSettlementRepo) ListAll(_ context.Context, driverID string, from, to int64, limit int) ([]*entity.Settlement, error) {
	var out []*entity.Settlement
	for _, s := range r.byTripID {
		if driverID == "" || s.DriverID == driverID {
			out = append(out, s)
		}
	}
	return out, nil
}

// ─── stub BankAccountRepository ──────────────────────────────────────────────

type stubBankAccountRepo struct {
	byDriverID map[string]*entity.BankAccount
}

var _ repository.BankAccountRepository = (*stubBankAccountRepo)(nil)

func newBankAccountRepo() *stubBankAccountRepo {
	return &stubBankAccountRepo{byDriverID: map[string]*entity.BankAccount{}}
}

func (r *stubBankAccountRepo) Save(_ context.Context, b *entity.BankAccount) error {
	r.byDriverID[b.DriverID] = b
	return nil
}

func (r *stubBankAccountRepo) FindByDriverID(_ context.Context, driverID string) (*entity.BankAccount, error) {
	b, ok := r.byDriverID[driverID]
	if !ok {
		return nil, domainerrors.NotFound("bank account not found")
	}
	return b, nil
}

// ─── stub PayoutRequestRepository ────────────────────────────────────────────

type stubPayoutRequestRepo struct {
	byID map[string]*entity.PayoutRequest
}

var _ repository.PayoutRequestRepository = (*stubPayoutRequestRepo)(nil)

func newPayoutRequestRepo() *stubPayoutRequestRepo {
	return &stubPayoutRequestRepo{byID: map[string]*entity.PayoutRequest{}}
}

func (r *stubPayoutRequestRepo) Save(_ context.Context, p *entity.PayoutRequest) error {
	r.byID[p.PayoutRequestID] = p
	return nil
}

func (r *stubPayoutRequestRepo) FindByID(_ context.Context, id string) (*entity.PayoutRequest, error) {
	p, ok := r.byID[id]
	if !ok {
		return nil, domainerrors.NotFound("payout request not found")
	}
	return p, nil
}

func (r *stubPayoutRequestRepo) FindInFlightByDriverID(_ context.Context, driverID string) (*entity.PayoutRequest, error) {
	for _, p := range r.byID {
		if p.DriverID == driverID && p.IsInFlight() {
			return p, nil
		}
	}
	return nil, domainerrors.NotFound("no in-flight payout request")
}

func (r *stubPayoutRequestRepo) ListByDriverID(_ context.Context, driverID string, limit int) ([]*entity.PayoutRequest, error) {
	var out []*entity.PayoutRequest
	for _, p := range r.byID {
		if p.DriverID == driverID {
			out = append(out, p)
		}
	}
	return out, nil
}

func (r *stubPayoutRequestRepo) ListByFilter(_ context.Context, status entity.PayoutStatus, driverID string, limit int) ([]*entity.PayoutRequest, error) {
	var out []*entity.PayoutRequest
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

// ─── stub AuditLogRepository ─────────────────────────────────────────────────

type stubAuditLogRepo struct {
	entries []*entity.AuditLog
}

var _ repository.AuditLogRepository = (*stubAuditLogRepo)(nil)

func newAuditLogRepo() *stubAuditLogRepo { return &stubAuditLogRepo{} }

func (r *stubAuditLogRepo) Save(_ context.Context, log *entity.AuditLog) error {
	r.entries = append(r.entries, log)
	return nil
}

func (r *stubAuditLogRepo) ListByDriverID(_ context.Context, driverID string, limit int) ([]*entity.AuditLog, error) {
	var out []*entity.AuditLog
	for _, e := range r.entries {
		if e.DriverID == driverID {
			out = append(out, e)
		}
	}
	return out, nil
}

// ─── test harness ─────────────────────────────────────────────────────────────

type settlementHarness struct {
	wallets     *stubWalletRepo
	ledger      *stubLedgerRepo
	tx          *stubTxRepo
	settlements *stubSettlementRepo
	audit       *stubAuditLogRepo
	engine      *app.CreateSettlementUseCase
	summary     *app.GetWalletSummaryUseCase
}

func newSettlementHarness(commissionRate float64) *settlementHarness {
	wr := newWalletRepo()
	lr := newLedgerRepo()
	tr := newTxRepo()
	sr := newSettlementRepo()
	ar := newAuditLogRepo()
	getOrCreate := app.NewGetOrCreateWalletUseCase(wr)
	engine := app.NewCreateSettlementUseCase(sr, getOrCreate, lr, tr, ar, commissionRate)
	summary := app.NewGetWalletSummaryUseCase(wr, lr, tr, newPayoutRequestRepo())
	return &settlementHarness{wallets: wr, ledger: lr, tx: tr, settlements: sr, audit: ar, engine: engine, summary: summary}
}

// ─── Settlement Engine Case 1 — Cash ─────────────────────────────────────────

func TestSettlementEngine_CashCase_DriverOwesOutstanding(t *testing.T) {
	h := newSettlementHarness(0.20)
	s, err := h.engine.Execute(ctx, app.CreateSettlementInput{
		TripID: "trip-1", DriverID: "d1", TripType: entity.TripTypeRide,
		PaymentMethod: entity.PaymentMethodCash, FareAmountCents: 120_000, Currency: "VND",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.CommissionAmountCents != 24_000 {
		t.Errorf("commission = %d, want 24000 (20%% of 120000)", s.CommissionAmountCents)
	}
	if s.DriverIncomeCents != 96_000 {
		t.Errorf("driver income = %d, want 96000", s.DriverIncomeCents)
	}

	summary, err := h.summary.Execute(ctx, "d1")
	if err != nil {
		t.Fatalf("summary: %v", err)
	}
	if summary.OutstandingCents != 24_000 {
		t.Errorf("outstanding = %d, want 24000", summary.OutstandingCents)
	}
	if summary.AvailableCents != 0 {
		t.Errorf("available = %d, want 0 — cash-collected income never enters the app wallet", summary.AvailableCents)
	}
	if summary.LifetimeEarnedCents != 96_000 {
		t.Errorf("lifetime earned = %d, want 96000 (net of commission)", summary.LifetimeEarnedCents)
	}
}

// TestSettlementEngine_UsesPricingCommission_NotFlatRate is critique #1's
// regression guard: when the caller resolved a real commission figure from
// Pricing V3 (HasCommissionDetail=true), Execute must use that number
// verbatim — never its own flat commissionRate — even though the harness is
// configured with a DIFFERENT flat rate (0.20) than the Pricing-supplied one.
func TestSettlementEngine_UsesPricingCommission_NotFlatRate(t *testing.T) {
	h := newSettlementHarness(0.20)
	s, err := h.engine.Execute(ctx, app.CreateSettlementInput{
		TripID: "trip-pricing-v3", DriverID: "d1", TripType: entity.TripTypeRide,
		PaymentMethod: entity.PaymentMethodCash, FareAmountCents: 120_000, Currency: "VND",
		HasCommissionDetail: true,
		CommissionCents:     18_500, // Pricing's own tiered number — not 20% of 120000
		DriverIncomeCents:   101_500,
		CommissionRate:      0.154,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.CommissionAmountCents != 18_500 {
		t.Errorf("commission = %d, want 18500 (Pricing's number, not the flat 20%% rate)", s.CommissionAmountCents)
	}
	if s.DriverIncomeCents != 101_500 {
		t.Errorf("driver income = %d, want 101500", s.DriverIncomeCents)
	}
	if s.CommissionRate != 0.154 {
		t.Errorf("commission rate = %v, want 0.154", s.CommissionRate)
	}
}

// TestSettlementEngine_VoucherStatus is critique #2's regression guard:
// "no voucher" (Pricing V3, VoucherDiscountCents=0), "voucher applied"
// (Pricing V3, VoucherDiscountCents>0), and "unknown" (Pricing V2 /
// HasCommissionDetail=false) must all be distinguishable, not all collapsed
// into a silent VoucherCostCents=0.
func TestSettlementEngine_VoucherStatus(t *testing.T) {
	cases := []struct {
		name        string
		tripID      string
		hasDetail   bool
		voucherCost int64
		wantStatus  entity.VoucherStatus
		wantCost    int64
	}{
		{"unknown_v2", "trip-v2", false, 0, entity.VoucherStatusUnknown, 0},
		{"none_v3", "trip-v3-none", true, 0, entity.VoucherStatusNone, 0},
		{"applied_v3", "trip-v3-applied", true, 15_000, entity.VoucherStatusApplied, 15_000},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			h := newSettlementHarness(0.20)
			s, err := h.engine.Execute(ctx, app.CreateSettlementInput{
				TripID: c.tripID, DriverID: "d1", TripType: entity.TripTypeRide,
				PaymentMethod: entity.PaymentMethodCash, FareAmountCents: 120_000, Currency: "VND",
				HasCommissionDetail:  c.hasDetail,
				CommissionCents:      18_500,
				DriverIncomeCents:    101_500,
				CommissionRate:       0.154,
				VoucherDiscountCents: c.voucherCost,
			})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if s.VoucherStatus != c.wantStatus {
				t.Errorf("VoucherStatus = %q, want %q", s.VoucherStatus, c.wantStatus)
			}
			if s.VoucherCostCents != c.wantCost {
				t.Errorf("VoucherCostCents = %d, want %d", s.VoucherCostCents, c.wantCost)
			}
		})
	}
}

// TestSettlementEngine_ResumesFromPending is critique #7's regression guard:
// simulates a crash right after the Settlement claim + first ledger entry
// were written (Status still Pending) by seeding those rows directly, then
// calls Execute again with the same input. It must resume and finish
// posting (reaching Status=Posted) using the SAME deterministic sub-IDs,
// without duplicating the income entry that was already there.
func TestSettlementEngine_ResumesFromPending(t *testing.T) {
	h := newSettlementHarness(0.20)
	in := app.CreateSettlementInput{
		TripID: "trip-crash", DriverID: "d1", TripType: entity.TripTypeRide,
		PaymentMethod: entity.PaymentMethodCash, FareAmountCents: 120_000, Currency: "VND",
	}

	driverWallet, err := h.wallets.FindByOwnerID(ctx, "d1")
	if err != nil {
		// Wallet not created yet in this fresh harness — create it the same
		// way the engine would on its first pass.
		getOrCreate := app.NewGetOrCreateWalletUseCase(h.wallets)
		driverWallet, err = getOrCreate.Execute(ctx, "d1", entity.WalletTypeDriver)
		if err != nil {
			t.Fatalf("seed driver wallet: %v", err)
		}
	}

	settlementID := "stl-trip-crash"
	incomeTxID := settlementID + "-income-tx"
	pending, err := entity.NewSettlement(
		settlementID, in.TripID, in.DriverID, in.TripType, in.PaymentMethod,
		in.FareAmountCents, 0.20, 24_000, 96_000, 0, 0, in.Currency, incomeTxID, time.Now().UTC(),
		entity.SettlementStatusPending, entity.VoucherStatusUnknown,
	)
	if err != nil {
		t.Fatalf("seed pending settlement: %v", err)
	}
	if err := h.settlements.Save(ctx, pending); err != nil {
		t.Fatalf("seed pending settlement save: %v", err)
	}
	incomeTx, err := entity.NewTransaction(incomeTxID, entity.TypeRideIncome, in.TripID, string(in.PaymentMethod), in.Currency, "seed", time.Now().UTC())
	if err != nil {
		t.Fatalf("seed income tx: %v", err)
	}
	if err := h.tx.Save(ctx, incomeTx); err != nil {
		t.Fatalf("seed income tx save: %v", err)
	}
	incomeEntry, err := entity.NewLedgerEntry(incomeTxID+"-entry", driverWallet.WalletID, incomeTxID, entity.DirectionCredit, 96_000, in.Currency, "seed", time.Now().UTC())
	if err != nil {
		t.Fatalf("seed income entry: %v", err)
	}
	if err := h.ledger.Save(ctx, incomeEntry); err != nil {
		t.Fatalf("seed income entry save: %v", err)
	}

	s, err := h.engine.Execute(ctx, in)
	if err != nil {
		t.Fatalf("resume execute: %v", err)
	}
	if s.Status != entity.SettlementStatusPosted {
		t.Errorf("Status = %q, want posted", s.Status)
	}

	entries, err := h.ledger.FindByWalletID(ctx, driverWallet.WalletID)
	if err != nil {
		t.Fatalf("find entries: %v", err)
	}
	incomeCount := 0
	for _, e := range entries {
		if e.TransactionID == incomeTxID {
			incomeCount++
		}
	}
	if incomeCount != 1 {
		t.Errorf("income entry count = %d, want 1 (resume must not duplicate the seeded entry)", incomeCount)
	}
}

// ─── Settlement Engine Case 2 — Wallet/Card ──────────────────────────────────

func TestSettlementEngine_WalletCase_DriverGetsAvailable(t *testing.T) {
	h := newSettlementHarness(0.20)
	s, err := h.engine.Execute(ctx, app.CreateSettlementInput{
		TripID: "trip-2", DriverID: "d1", TripType: entity.TripTypeRide,
		PaymentMethod: entity.PaymentMethodWallet, FareAmountCents: 120_000, Currency: "VND",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.DriverIncomeCents != 96_000 {
		t.Errorf("driver income = %d, want 96000", s.DriverIncomeCents)
	}

	summary, err := h.summary.Execute(ctx, "d1")
	if err != nil {
		t.Fatalf("summary: %v", err)
	}
	if summary.AvailableCents != 96_000 {
		t.Errorf("available = %d, want 96000 — electronically-collected income is withdrawable", summary.AvailableCents)
	}
	if summary.OutstandingCents != 0 {
		t.Errorf("outstanding = %d, want 0", summary.OutstandingCents)
	}
	if summary.NetCents != 96_000 {
		t.Errorf("net = %d, want 96000", summary.NetCents)
	}
}

// ─── Idempotency ──────────────────────────────────────────────────────────────

func TestSettlementEngine_Idempotent(t *testing.T) {
	h := newSettlementHarness(0.20)
	in := app.CreateSettlementInput{
		TripID: "trip-3", DriverID: "d1", TripType: entity.TripTypeRide,
		PaymentMethod: entity.PaymentMethodCash, FareAmountCents: 100_000, Currency: "VND",
	}
	first, err := h.engine.Execute(ctx, in)
	if err != nil {
		t.Fatalf("first settle: %v", err)
	}
	second, err := h.engine.Execute(ctx, in)
	if err != nil {
		t.Fatalf("second settle should succeed idempotently: %v", err)
	}
	if second.SettlementID != first.SettlementID {
		t.Error("second settle should return the exact same Settlement, not create a new one")
	}

	summary, _ := h.summary.Execute(ctx, "d1")
	if summary.OutstandingCents != 20_000 {
		t.Errorf("outstanding = %d, want 20000 (only ONE settlement's worth, not double-posted)", summary.OutstandingCents)
	}
}

// ─── Mixed cash + wallet trips ────────────────────────────────────────────────

func TestSettlementEngine_MixedCashAndWalletTrips_NetBalance(t *testing.T) {
	h := newSettlementHarness(0.20)
	_, err := h.engine.Execute(ctx, app.CreateSettlementInput{
		TripID: "trip-cash", DriverID: "d1", TripType: entity.TripTypeRide,
		PaymentMethod: entity.PaymentMethodCash, FareAmountCents: 120_000, Currency: "VND",
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = h.engine.Execute(ctx, app.CreateSettlementInput{
		TripID: "trip-wallet", DriverID: "d1", TripType: entity.TripTypeRide,
		PaymentMethod: entity.PaymentMethodWallet, FareAmountCents: 100_000, Currency: "VND",
	})
	if err != nil {
		t.Fatal(err)
	}

	summary, err := h.summary.Execute(ctx, "d1")
	if err != nil {
		t.Fatal(err)
	}
	// cash trip: +24000 outstanding, 0 available
	// wallet trip: +80000 available (100000 - 20% = 80000)
	if summary.OutstandingCents != 24_000 {
		t.Errorf("outstanding = %d, want 24000", summary.OutstandingCents)
	}
	if summary.AvailableCents != 80_000 {
		t.Errorf("available = %d, want 80000", summary.AvailableCents)
	}
	if summary.NetCents != 56_000 {
		t.Errorf("net = %d, want 56000 (80000 available - 24000 outstanding)", summary.NetCents)
	}
	if summary.LifetimeEarnedCents != 176_000 {
		t.Errorf("lifetime earned = %d, want 176000 (96000 + 80000)", summary.LifetimeEarnedCents)
	}
}
