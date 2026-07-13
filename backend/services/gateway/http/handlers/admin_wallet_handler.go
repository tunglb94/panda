package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/fairride/gateway/http/middleware"
	walletapp "github.com/fairride/wallet/app"
	walletentity "github.com/fairride/wallet/domain/entity"
)

// AdminWalletHandler exposes the Financial Core's admin surface (Phần 10)
// — every route this handler serves must be wrapped in both `auth` and
// `middleware.RequireAdmin` by the router; this handler itself does not
// re-check the role, it trusts the middleware chain (same convention as
// AdminKYCHandler). "Không cần UI. Chỉ API." — no admin UI is built here.
type AdminWalletHandler struct {
	listSettlements        *walletapp.ListSettlementsUseCase
	getSettlementDetail    *walletapp.GetSettlementDetailUseCase
	listOutstandingDrivers *walletapp.ListOutstandingDriversUseCase
	listPayoutRequests     *walletapp.ListPayoutRequestsUseCase
	approvePayout          *walletapp.ApprovePayoutRequestUseCase
	rejectPayout           *walletapp.RejectPayoutRequestUseCase
	markPayoutPaid         *walletapp.MarkPayoutPaidUseCase
	manualAdjustment       *walletapp.ManualAdjustmentUseCase
}

func NewAdminWalletHandler(
	listSettlements *walletapp.ListSettlementsUseCase,
	getSettlementDetail *walletapp.GetSettlementDetailUseCase,
	listOutstandingDrivers *walletapp.ListOutstandingDriversUseCase,
	listPayoutRequests *walletapp.ListPayoutRequestsUseCase,
	approvePayout *walletapp.ApprovePayoutRequestUseCase,
	rejectPayout *walletapp.RejectPayoutRequestUseCase,
	markPayoutPaid *walletapp.MarkPayoutPaidUseCase,
	manualAdjustment *walletapp.ManualAdjustmentUseCase,
) *AdminWalletHandler {
	return &AdminWalletHandler{
		listSettlements: listSettlements, getSettlementDetail: getSettlementDetail,
		listOutstandingDrivers: listOutstandingDrivers, listPayoutRequests: listPayoutRequests,
		approvePayout: approvePayout, rejectPayout: rejectPayout, markPayoutPaid: markPayoutPaid,
		manualAdjustment: manualAdjustment,
	}
}

func (h *AdminWalletHandler) configured() bool { return h != nil && h.listSettlements != nil }

func (h *AdminWalletHandler) unavailable(w http.ResponseWriter) bool {
	if h.configured() {
		return false
	}
	writeJSON(w, http.StatusServiceUnavailable, errorResponse{Error: "admin wallet service not configured"})
	return true
}

// ─── Settlement List / Detail (Phần 10) ─────────────────────────────────────

// ListSettlements handles GET /api/v1/admin/settlements?driver_id=&from=&to=&limit=.
func (h *AdminWalletHandler) ListSettlements(w http.ResponseWriter, r *http.Request) {
	if h.unavailable(w) {
		return
	}
	driverID := r.URL.Query().Get("driver_id")
	from, _ := strconv.ParseInt(r.URL.Query().Get("from"), 10, 64)
	to, _ := strconv.ParseInt(r.URL.Query().Get("to"), 10, 64)
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	list, err := h.listSettlements.Execute(r.Context(), driverID, from, to, limit)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	items := make([]map[string]any, len(list))
	for i, s := range list {
		items[i] = settlementJSON(s)
	}
	writeJSON(w, http.StatusOK, map[string]any{"settlements": items})
}

// GetSettlementDetail handles GET /api/v1/admin/settlements/{settlementID}.
func (h *AdminWalletHandler) GetSettlementDetail(w http.ResponseWriter, r *http.Request) {
	if h.unavailable(w) {
		return
	}
	id := r.PathValue("settlementID")
	s, err := h.getSettlementDetail.Execute(r.Context(), id)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, settlementJSON(s))
}

func settlementJSON(s *walletentity.Settlement) map[string]any {
	return map[string]any{
		"settlement_id":           s.SettlementID,
		"trip_id":                 s.TripID,
		"driver_id":               s.DriverID,
		"trip_type":               string(s.TripType),
		"payment_method":          string(s.PaymentMethod),
		"fare_amount_cents":       s.FareAmountCents,
		"commission_rate":         s.CommissionRate,
		"commission_amount_cents": s.CommissionAmountCents,
		"driver_income_cents":     s.DriverIncomeCents,
		"promotion_subsidy_cents": s.PromotionSubsidyCents,
		"voucher_cost_cents":      s.VoucherCostCents,
		"voucher_status":          string(s.VoucherStatus),
		"status":                  string(s.Status),
		"currency":                s.Currency,
		"created_at":              s.CreatedAt.UTC().Format(time.RFC3339),
	}
}

// ─── Outstanding Drivers (Phần 10) ───────────────────────────────────────────

// ListOutstandingDrivers handles GET /api/v1/admin/settlements/outstanding?limit=.
func (h *AdminWalletHandler) ListOutstandingDrivers(w http.ResponseWriter, r *http.Request) {
	if h.unavailable(w) {
		return
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	list, err := h.listOutstandingDrivers.Execute(r.Context(), limit)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	items := make([]map[string]any, len(list))
	for i, d := range list {
		items[i] = map[string]any{
			"driver_id":         d.DriverID,
			"outstanding_cents": d.OutstandingCents,
			"currency":          d.Currency,
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"drivers": items})
}

// ─── Withdrawal List / Approve / Reject / Paid (Phần 8/10) ──────────────────

// ListPayoutRequests handles GET /api/v1/admin/payouts?status=&driver_id=&limit=.
func (h *AdminWalletHandler) ListPayoutRequests(w http.ResponseWriter, r *http.Request) {
	if h.unavailable(w) {
		return
	}
	status := walletentity.PayoutStatus(r.URL.Query().Get("status"))
	driverID := r.URL.Query().Get("driver_id")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	list, err := h.listPayoutRequests.Execute(r.Context(), status, driverID, limit)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	items := make([]map[string]any, len(list))
	for i, p := range list {
		items[i] = payoutRequestJSON(p)
		items[i]["driver_id"] = p.DriverID
	}
	writeJSON(w, http.StatusOK, map[string]any{"payout_requests": items})
}

type reviewPayoutRequest struct {
	Reason string `json:"reason"`
}

// ApprovePayoutRequest handles POST /api/v1/admin/payouts/{payoutRequestID}/approve.
func (h *AdminWalletHandler) ApprovePayoutRequest(w http.ResponseWriter, r *http.Request) {
	if h.unavailable(w) {
		return
	}
	id := r.PathValue("payoutRequestID")
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		writeBadRequest(w, "missing auth claims")
		return
	}
	p, err := h.approvePayout.Execute(r.Context(), walletapp.ReviewPayoutRequestInput{PayoutRequestID: id, Reviewer: claims.UserID})
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, payoutRequestJSON(p))
}

// RejectPayoutRequest handles POST /api/v1/admin/payouts/{payoutRequestID}/reject.
func (h *AdminWalletHandler) RejectPayoutRequest(w http.ResponseWriter, r *http.Request) {
	if h.unavailable(w) {
		return
	}
	id := r.PathValue("payoutRequestID")
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		writeBadRequest(w, "missing auth claims")
		return
	}
	var req reviewPayoutRequest
	_ = json.NewDecoder(r.Body).Decode(&req)
	p, err := h.rejectPayout.Execute(r.Context(), walletapp.ReviewPayoutRequestInput{PayoutRequestID: id, Reviewer: claims.UserID, Reason: req.Reason})
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, payoutRequestJSON(p))
}

// MarkPayoutPaid handles POST /api/v1/admin/payouts/{payoutRequestID}/paid
// (Phần 8 — "Không tự Paid": this explicit Admin action is the only way a
// PayoutRequest ever reaches Paid, and the only place a Withdrawal ledger
// entry is created).
func (h *AdminWalletHandler) MarkPayoutPaid(w http.ResponseWriter, r *http.Request) {
	if h.unavailable(w) {
		return
	}
	id := r.PathValue("payoutRequestID")
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		writeBadRequest(w, "missing auth claims")
		return
	}
	p, err := h.markPayoutPaid.Execute(r.Context(), id, claims.UserID)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, payoutRequestJSON(p))
}

// ─── Manual Adjustment (Phần 10/12) ──────────────────────────────────────────

type manualAdjustmentRequest struct {
	DriverID    string `json:"driver_id"`
	AmountCents int64  `json:"amount_cents"`
	Direction   string `json:"direction"` // "credit" | "debit"
	Reason      string `json:"reason"`
}

// ManualAdjustment handles POST /api/v1/admin/wallet/adjustments. Always
// audit-logged (Phần 12 — "Adjustment phải có Audit Log").
func (h *AdminWalletHandler) ManualAdjustment(w http.ResponseWriter, r *http.Request) {
	if h.unavailable(w) {
		return
	}
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		writeBadRequest(w, "missing auth claims")
		return
	}
	var req manualAdjustmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeBadRequest(w, "invalid request body")
		return
	}
	entry, err := h.manualAdjustment.Execute(r.Context(), walletapp.ManualAdjustmentInput{
		DriverID: req.DriverID, AmountCents: req.AmountCents,
		Direction: walletentity.EntryDirection(req.Direction), Reason: req.Reason, ActorID: claims.UserID,
	})
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{
		"entry_id":     entry.EntryID,
		"direction":    string(entry.Direction),
		"amount_cents": entry.AmountCents,
		"currency":     entry.Currency,
		"created_at":   entry.CreatedAt.UTC().Format(time.RFC3339),
	})
}
