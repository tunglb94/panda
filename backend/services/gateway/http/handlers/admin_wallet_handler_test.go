package handlers_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fairride/gateway/http/handlers"
	"github.com/fairride/identity/infrastructure/jwt"
	walletapp "github.com/fairride/wallet/app"
	walletentity "github.com/fairride/wallet/domain/entity"
)

func buildAdminWalletHandler(deps *walletTestDeps) *handlers.AdminWalletHandler {
	getOrCreate := walletapp.NewGetOrCreateWalletUseCase(deps.wallets)
	return handlers.NewAdminWalletHandler(
		walletapp.NewListSettlementsUseCase(deps.settlements),
		walletapp.NewGetSettlementDetailUseCase(deps.settlements),
		walletapp.NewListOutstandingDriversUseCase(deps.ledger),
		walletapp.NewListPayoutRequestsUseCase(deps.payouts),
		walletapp.NewApprovePayoutRequestUseCase(deps.payouts, deps.audit),
		walletapp.NewRejectPayoutRequestUseCase(deps.payouts, getOrCreate, deps.ledger, deps.tx, deps.audit),
		walletapp.NewMarkPayoutPaidUseCase(deps.payouts, getOrCreate, deps.ledger, deps.tx, deps.audit),
		walletapp.NewManualAdjustmentUseCase(getOrCreate, deps.ledger, deps.tx, deps.audit),
	)
}

func adminWalletRequest(t *testing.T, method, path string, body any) *http.Request {
	t.Helper()
	r := walletRequest(t, method, path, "admin1", body)
	return injectClaims(r, &jwt.AccessClaims{UserID: "admin1", UserType: "admin"})
}

func TestAdminWalletHandler_ServiceUnavailableWhenNotConfigured(t *testing.T) {
	h := handlers.NewAdminWalletHandler(nil, nil, nil, nil, nil, nil, nil, nil)
	w := httptest.NewRecorder()
	h.ListSettlements(w, adminWalletRequest(t, http.MethodGet, "/api/v1/admin/settlements", nil))
	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("status = %d, want 503", w.Code)
	}
}

func TestAdminWalletHandler_ListOutstandingDrivers(t *testing.T) {
	deps := newWalletTestDeps()
	settleTripForTest(t, deps, "trip-1", "d1", 120_000, walletentity.PaymentMethodCash)
	h := buildAdminWalletHandler(deps)

	w := httptest.NewRecorder()
	h.ListOutstandingDrivers(w, adminWalletRequest(t, http.MethodGet, "/api/v1/admin/settlements/outstanding", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200, body=%s", w.Code, w.Body.String())
	}
	var body map[string]any
	_ = json.NewDecoder(w.Body).Decode(&body)
	drivers, _ := body["drivers"].([]any)
	if len(drivers) != 1 {
		t.Fatalf("expected 1 outstanding driver, got %d", len(drivers))
	}
	d := drivers[0].(map[string]any)
	if d["driver_id"] != "d1" || d["outstanding_cents"] != float64(24_000) {
		t.Errorf("unexpected outstanding driver row: %+v", d)
	}
}

func TestAdminWalletHandler_ApproveThenPaid_CreatesWithdrawalEntry(t *testing.T) {
	deps := newWalletTestDeps()
	settleTripForTest(t, deps, "trip-1", "d1", 1_000_000, walletentity.PaymentMethodWallet)
	_ = deps.banks.Save(context.Background(), mustBankAccount(t, "d1"))
	wh := buildWalletHandler(deps, nil)
	ah := buildAdminWalletHandler(deps)

	w := httptest.NewRecorder()
	wh.CreatePayoutRequest(w, walletRequest(t, http.MethodPost, "/api/v1/driver/wallet/payouts", "d1", map[string]any{"amount_cents": 200_000}))
	var created map[string]any
	_ = json.NewDecoder(w.Body).Decode(&created)
	payoutID := created["payout_request_id"].(string)

	w2 := httptest.NewRecorder()
	r2 := adminWalletRequest(t, http.MethodPost, "/api/v1/admin/payouts/"+payoutID+"/approve", nil)
	r2.SetPathValue("payoutRequestID", payoutID)
	ah.ApprovePayoutRequest(w2, r2)
	if w2.Code != http.StatusOK {
		t.Fatalf("approve status = %d, want 200, body=%s", w2.Code, w2.Body.String())
	}

	w3 := httptest.NewRecorder()
	r3 := adminWalletRequest(t, http.MethodPost, "/api/v1/admin/payouts/"+payoutID+"/paid", nil)
	r3.SetPathValue("payoutRequestID", payoutID)
	ah.MarkPayoutPaid(w3, r3)
	if w3.Code != http.StatusOK {
		t.Fatalf("mark paid status = %d, want 200, body=%s", w3.Code, w3.Body.String())
	}
	var paid map[string]any
	_ = json.NewDecoder(w3.Body).Decode(&paid)
	if paid["status"] != "paid" {
		t.Errorf("status = %v, want paid", paid["status"])
	}

	w4 := httptest.NewRecorder()
	wh.GetSummary(w4, walletRequest(t, http.MethodGet, "/api/v1/driver/wallet/summary", "d1", nil))
	var summary map[string]any
	_ = json.NewDecoder(w4.Body).Decode(&summary)
	if summary["available_cents"] != float64(600_000) {
		t.Errorf("available = %v, want 600000 (800000 - 200000 withdrawn)", summary["available_cents"])
	}
}

func TestAdminWalletHandler_RejectRequiresReason(t *testing.T) {
	deps := newWalletTestDeps()
	settleTripForTest(t, deps, "trip-1", "d1", 1_000_000, walletentity.PaymentMethodWallet)
	_ = deps.banks.Save(context.Background(), mustBankAccount(t, "d1"))
	wh := buildWalletHandler(deps, nil)
	ah := buildAdminWalletHandler(deps)

	w := httptest.NewRecorder()
	wh.CreatePayoutRequest(w, walletRequest(t, http.MethodPost, "/api/v1/driver/wallet/payouts", "d1", map[string]any{"amount_cents": 100_000}))
	var created map[string]any
	_ = json.NewDecoder(w.Body).Decode(&created)
	payoutID := created["payout_request_id"].(string)

	w2 := httptest.NewRecorder()
	r2 := adminWalletRequest(t, http.MethodPost, "/api/v1/admin/payouts/"+payoutID+"/reject", map[string]any{"reason": ""})
	r2.SetPathValue("payoutRequestID", payoutID)
	ah.RejectPayoutRequest(w2, r2)
	if w2.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400 for empty reject reason, body=%s", w2.Code, w2.Body.String())
	}
}

func TestAdminWalletHandler_ManualAdjustment_RequiresReason(t *testing.T) {
	deps := newWalletTestDeps()
	ah := buildAdminWalletHandler(deps)

	w := httptest.NewRecorder()
	ah.ManualAdjustment(w, adminWalletRequest(t, http.MethodPost, "/api/v1/admin/wallet/adjustments", map[string]any{
		"driver_id": "d1", "amount_cents": 50_000, "direction": "credit", "reason": "",
	}))
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400 for empty reason, body=%s", w.Code, w.Body.String())
	}
}

func TestAdminWalletHandler_ManualAdjustment_OK(t *testing.T) {
	deps := newWalletTestDeps()
	ah := buildAdminWalletHandler(deps)
	wh := buildWalletHandler(deps, nil)

	w := httptest.NewRecorder()
	ah.ManualAdjustment(w, adminWalletRequest(t, http.MethodPost, "/api/v1/admin/wallet/adjustments", map[string]any{
		"driver_id": "d1", "amount_cents": 50_000, "direction": "credit", "reason": "Bồi thường sự cố hệ thống",
	}))
	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201, body=%s", w.Code, w.Body.String())
	}

	w2 := httptest.NewRecorder()
	wh.GetSummary(w2, walletRequest(t, http.MethodGet, "/api/v1/driver/wallet/summary", "d1", nil))
	var summary map[string]any
	_ = json.NewDecoder(w2.Body).Decode(&summary)
	if summary["available_cents"] != float64(50_000) {
		t.Errorf("available = %v, want 50000", summary["available_cents"])
	}
}
