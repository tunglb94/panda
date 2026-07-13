package app_test

import (
	"testing"

	"github.com/fairride/shared/errors"
	"github.com/fairride/wallet/app"
	"github.com/fairride/wallet/domain/entity"
)

func settleCashTrip(t *testing.T, h *settlementHarness, tripID, driverID string, fare int64) {
	t.Helper()
	_, err := h.engine.Execute(ctx, app.CreateSettlementInput{
		TripID: tripID, DriverID: driverID, TripType: entity.TripTypeRide,
		PaymentMethod: entity.PaymentMethodWallet, FareAmountCents: fare, Currency: "VND",
	})
	if err != nil {
		t.Fatalf("settle: %v", err)
	}
}

func newPayoutHarness(t *testing.T) (*settlementHarness, *stubBankAccountRepo, *stubPayoutRequestRepo, *app.CreatePayoutRequestUseCase) {
	t.Helper()
	h := newSettlementHarness(0.20)
	banks := newBankAccountRepo()
	payouts := newPayoutRequestRepo()
	summary := app.NewGetWalletSummaryUseCase(h.wallets, h.ledger, h.tx, payouts)
	getOrCreate := app.NewGetOrCreateWalletUseCase(h.wallets)
	create := app.NewCreatePayoutRequestUseCase(payouts, banks, summary, getOrCreate, h.ledger, h.tx, h.audit, 0)
	return h, banks, payouts, create
}

func seedBankAccount(t *testing.T, banks *stubBankAccountRepo, driverID string) {
	t.Helper()
	b, err := entity.NewBankAccount("bank-1", driverID, "Vietcombank", "Nguyen Van A", "0123456789", "", now)
	if err != nil {
		t.Fatal(err)
	}
	_ = banks.Save(ctx, b)
}

func TestCreatePayoutRequest_RejectsWithoutBankAccount(t *testing.T) {
	h, _, _, create := newPayoutHarness(t)
	settleCashTrip(t, h, "trip-1", "d1", 500_000) // 400000 available after 20% commission

	_, err := create.Execute(ctx, app.CreatePayoutRequestInput{DriverID: "d1", AmountCents: 100_000})
	if !errors.IsCode(err, errors.CodePreconditionFailed) {
		t.Errorf("want CodePreconditionFailed for missing bank account, got %v", err)
	}
}

func TestCreatePayoutRequest_RejectsBelowMinimum(t *testing.T) {
	h, banks, _, create := newPayoutHarness(t)
	seedBankAccount(t, banks, "d1")
	settleCashTrip(t, h, "trip-1", "d1", 500_000)

	_, err := create.Execute(ctx, app.CreatePayoutRequestInput{DriverID: "d1", AmountCents: 1_000})
	if !errors.IsCode(err, errors.CodePreconditionFailed) {
		t.Errorf("want CodePreconditionFailed for below-minimum amount, got %v", err)
	}
}

func TestCreatePayoutRequest_RejectsWhenInsufficientAvailable(t *testing.T) {
	h, banks, _, create := newPayoutHarness(t)
	seedBankAccount(t, banks, "d1")
	settleCashTrip(t, h, "trip-1", "d1", 100_000) // 80000 available

	_, err := create.Execute(ctx, app.CreatePayoutRequestInput{DriverID: "d1", AmountCents: 200_000})
	if !errors.IsCode(err, errors.CodePreconditionFailed) {
		t.Errorf("want CodePreconditionFailed for insufficient balance, got %v", err)
	}
}

func TestCreatePayoutRequest_RejectsSecondInFlightRequest(t *testing.T) {
	h, banks, _, create := newPayoutHarness(t)
	seedBankAccount(t, banks, "d1")
	settleCashTrip(t, h, "trip-1", "d1", 1_000_000) // 800000 available

	_, err := create.Execute(ctx, app.CreatePayoutRequestInput{DriverID: "d1", AmountCents: 100_000})
	if err != nil {
		t.Fatalf("first request should succeed: %v", err)
	}
	_, err = create.Execute(ctx, app.CreatePayoutRequestInput{DriverID: "d1", AmountCents: 100_000})
	if !errors.IsCode(err, errors.CodePreconditionFailed) {
		t.Errorf("want CodePreconditionFailed for a second in-flight request, got %v", err)
	}
}

// Phần 4 — a cash-collected driver with Outstanding > Available cannot withdraw.
func TestCreatePayoutRequest_RejectsWhenOutstandingExceedsAvailable(t *testing.T) {
	h, banks, _, create := newPayoutHarness(t)
	seedBankAccount(t, banks, "d1")
	// Cash trip: 0 available, 200000 outstanding.
	_, err := h.engine.Execute(ctx, app.CreateSettlementInput{
		TripID: "trip-1", DriverID: "d1", TripType: entity.TripTypeRide,
		PaymentMethod: entity.PaymentMethodCash, FareAmountCents: 1_000_000, Currency: "VND",
	})
	if err != nil {
		t.Fatal(err)
	}

	_, err = create.Execute(ctx, app.CreatePayoutRequestInput{DriverID: "d1", AmountCents: 50_000})
	if !errors.IsCode(err, errors.CodePreconditionFailed) {
		t.Errorf("want CodePreconditionFailed when outstanding exceeds available, got %v", err)
	}
}

func TestPayoutFlow_ApproveThenMarkPaid_CreatesWithdrawalLedgerEntry(t *testing.T) {
	h, banks, payouts, create := newPayoutHarness(t)
	seedBankAccount(t, banks, "d1")
	settleCashTrip(t, h, "trip-1", "d1", 1_000_000) // 800000 available

	req, err := create.Execute(ctx, app.CreatePayoutRequestInput{DriverID: "d1", AmountCents: 300_000})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if req.Status != entity.PayoutPending {
		t.Fatalf("status = %v, want pending", req.Status)
	}

	approve := app.NewApprovePayoutRequestUseCase(payouts, h.audit)
	req, err = approve.Execute(ctx, app.ReviewPayoutRequestInput{PayoutRequestID: req.PayoutRequestID, Reviewer: "admin1"})
	if err != nil {
		t.Fatalf("approve: %v", err)
	}
	if req.Status != entity.PayoutApproved {
		t.Fatalf("status = %v, want approved", req.Status)
	}

	getOrCreate := app.NewGetOrCreateWalletUseCase(h.wallets)
	markPaid := app.NewMarkPayoutPaidUseCase(payouts, getOrCreate, h.ledger, h.tx, h.audit)
	req, err = markPaid.Execute(ctx, req.PayoutRequestID, "admin1")
	if err != nil {
		t.Fatalf("mark paid: %v", err)
	}
	if req.Status != entity.PayoutPaid {
		t.Fatalf("status = %v, want paid", req.Status)
	}
	if req.TransactionID == "" {
		t.Error("expected a transaction_id linking to the Withdrawal ledger entry")
	}

	summary, err := h.summary.Execute(ctx, "d1")
	if err != nil {
		t.Fatal(err)
	}
	if summary.AvailableCents != 500_000 {
		t.Errorf("available = %d, want 500000 (800000 - 300000 withdrawn)", summary.AvailableCents)
	}
	if summary.LifetimeWithdrawnCents != 300_000 {
		t.Errorf("lifetime withdrawn = %d, want 300000", summary.LifetimeWithdrawnCents)
	}
}

// TestCreatePayoutRequest_FreezesAvailableImmediately is critique #8's
// regression guard: Available must drop the moment a request is CREATED
// (Pending), not only once it's Paid — otherwise the same money could be
// claimed again by anything else reading Available while the request sits
// in flight.
func TestCreatePayoutRequest_FreezesAvailableImmediately(t *testing.T) {
	h, banks, _, create := newPayoutHarness(t)
	seedBankAccount(t, banks, "d1")
	settleCashTrip(t, h, "trip-1", "d1", 1_000_000) // 800000 available

	before, err := h.summary.Execute(ctx, "d1")
	if err != nil {
		t.Fatal(err)
	}
	if before.AvailableCents != 800_000 {
		t.Fatalf("available before request = %d, want 800000", before.AvailableCents)
	}

	if _, err := create.Execute(ctx, app.CreatePayoutRequestInput{DriverID: "d1", AmountCents: 300_000}); err != nil {
		t.Fatalf("create: %v", err)
	}

	after, err := h.summary.Execute(ctx, "d1")
	if err != nil {
		t.Fatal(err)
	}
	if after.AvailableCents != 500_000 {
		t.Errorf("available after request = %d, want 500000 (frozen immediately, before Approve/Paid)", after.AvailableCents)
	}
}

// TestRejectPayoutRequest_RestoresAvailable is critique #8's reversal guard.
func TestRejectPayoutRequest_RestoresAvailable(t *testing.T) {
	h, banks, payouts, create := newPayoutHarness(t)
	seedBankAccount(t, banks, "d1")
	settleCashTrip(t, h, "trip-1", "d1", 1_000_000) // 800000 available

	req, err := create.Execute(ctx, app.CreatePayoutRequestInput{DriverID: "d1", AmountCents: 300_000})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	reject := app.NewRejectPayoutRequestUseCase(payouts, app.NewGetOrCreateWalletUseCase(h.wallets), h.ledger, h.tx, h.audit)
	if _, err := reject.Execute(ctx, app.ReviewPayoutRequestInput{PayoutRequestID: req.PayoutRequestID, Reviewer: "admin1", Reason: "Sai thông tin ngân hàng"}); err != nil {
		t.Fatalf("reject: %v", err)
	}

	summary, err := h.summary.Execute(ctx, "d1")
	if err != nil {
		t.Fatal(err)
	}
	if summary.AvailableCents != 800_000 {
		t.Errorf("available after reject = %d, want 800000 (hold fully released)", summary.AvailableCents)
	}
}

func TestPayoutFlow_CannotMarkPaidWithoutApproval(t *testing.T) {
	h, banks, payouts, create := newPayoutHarness(t)
	seedBankAccount(t, banks, "d1")
	settleCashTrip(t, h, "trip-1", "d1", 1_000_000)

	req, err := create.Execute(ctx, app.CreatePayoutRequestInput{DriverID: "d1", AmountCents: 100_000})
	if err != nil {
		t.Fatal(err)
	}

	getOrCreate := app.NewGetOrCreateWalletUseCase(h.wallets)
	markPaid := app.NewMarkPayoutPaidUseCase(payouts, getOrCreate, h.ledger, h.tx, h.audit)
	_, err = markPaid.Execute(ctx, req.PayoutRequestID, "admin1")
	if !errors.IsCode(err, errors.CodePreconditionFailed) {
		t.Errorf("want CodePreconditionFailed — Không tự Paid, must go through Approve first, got %v", err)
	}
}

func TestRejectPayoutRequest_RequiresReason(t *testing.T) {
	h, banks, payouts, create := newPayoutHarness(t)
	seedBankAccount(t, banks, "d1")
	settleCashTrip(t, h, "trip-1", "d1", 1_000_000)

	req, err := create.Execute(ctx, app.CreatePayoutRequestInput{DriverID: "d1", AmountCents: 100_000})
	if err != nil {
		t.Fatal(err)
	}

	reject := app.NewRejectPayoutRequestUseCase(payouts, app.NewGetOrCreateWalletUseCase(h.wallets), h.ledger, h.tx, h.audit)
	_, err = reject.Execute(ctx, app.ReviewPayoutRequestInput{PayoutRequestID: req.PayoutRequestID, Reviewer: "admin1", Reason: ""})
	if !errors.IsCode(err, errors.CodeInvalidArgument) {
		t.Errorf("want CodeInvalidArgument for empty reject reason, got %v", err)
	}

	req, err = reject.Execute(ctx, app.ReviewPayoutRequestInput{PayoutRequestID: req.PayoutRequestID, Reviewer: "admin1", Reason: "Sai thông tin ngân hàng"})
	if err != nil {
		t.Fatalf("reject with reason should succeed: %v", err)
	}
	if req.Status != entity.PayoutRejected {
		t.Errorf("status = %v, want rejected", req.Status)
	}

	// After rejection, driver should be able to request again (no longer in-flight).
	_, err = create.Execute(ctx, app.CreatePayoutRequestInput{DriverID: "d1", AmountCents: 100_000})
	if err != nil {
		t.Errorf("should be able to request again after rejection: %v", err)
	}
}

// ─── Manual Adjustment (Phần 10/12) ─────────────────────────────────────────

func TestManualAdjustment_RequiresReason(t *testing.T) {
	h := newSettlementHarness(0.20)
	getOrCreate := app.NewGetOrCreateWalletUseCase(h.wallets)
	uc := app.NewManualAdjustmentUseCase(getOrCreate, h.ledger, h.tx, h.audit)

	_, err := uc.Execute(ctx, app.ManualAdjustmentInput{
		DriverID: "d1", AmountCents: 50_000, Direction: entity.DirectionCredit, Reason: "", ActorID: "admin1",
	})
	if !errors.IsCode(err, errors.CodeInvalidArgument) {
		t.Errorf("want CodeInvalidArgument for empty reason, got %v", err)
	}
}

func TestManualAdjustment_CreditIncreasesAvailable(t *testing.T) {
	h := newSettlementHarness(0.20)
	getOrCreate := app.NewGetOrCreateWalletUseCase(h.wallets)
	uc := app.NewManualAdjustmentUseCase(getOrCreate, h.ledger, h.tx, h.audit)

	_, err := uc.Execute(ctx, app.ManualAdjustmentInput{
		DriverID: "d1", AmountCents: 50_000, Direction: entity.DirectionCredit, Reason: "Bồi thường sự cố", ActorID: "admin1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	summary, err := h.summary.Execute(ctx, "d1")
	if err != nil {
		t.Fatal(err)
	}
	if summary.AvailableCents != 50_000 {
		t.Errorf("available = %d, want 50000", summary.AvailableCents)
	}
	if len(h.audit.entries) != 1 {
		t.Errorf("expected 1 audit entry, got %d", len(h.audit.entries))
	}
}
