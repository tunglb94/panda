package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	driverapp "github.com/fairride/driver/app"
	driverentity "github.com/fairride/driver/domain/entity"
	"github.com/fairride/gateway/http/middleware"
	walletapp "github.com/fairride/wallet/app"
	walletentity "github.com/fairride/wallet/domain/entity"
)

// WalletHandler exposes the Driver Finance module (Settlement Engine spec)
// over HTTP, scoped to the currently-authenticated driver only (Phần 12 —
// "Driver chỉ xem dữ liệu mình"; claims.UserID is always used as the
// driver_id, never a client-supplied one). Runs entirely in-process
// against the wallet service's own Go packages — same pattern as
// KYCHandler/CallHandler (no gRPC surface, no protoc/buf toolchain here).
type WalletHandler struct {
	getSummary       *walletapp.GetWalletSummaryUseCase
	getStatement     *walletapp.GetStatementUseCase
	listTransactions *walletapp.ListWalletTransactionsUseCase
	getBankAccount   *walletapp.GetBankAccountUseCase
	setBankAccount   *walletapp.SetBankAccountUseCase
	createPayout     *walletapp.CreatePayoutRequestUseCase
	listMyPayouts    *walletapp.ListMyPayoutRequestsUseCase
	getDriverKYC     *driverapp.GetDriverVerificationUseCase
}

func NewWalletHandler(
	getSummary *walletapp.GetWalletSummaryUseCase,
	getStatement *walletapp.GetStatementUseCase,
	listTransactions *walletapp.ListWalletTransactionsUseCase,
	getBankAccount *walletapp.GetBankAccountUseCase,
	setBankAccount *walletapp.SetBankAccountUseCase,
	createPayout *walletapp.CreatePayoutRequestUseCase,
	listMyPayouts *walletapp.ListMyPayoutRequestsUseCase,
	getDriverKYC *driverapp.GetDriverVerificationUseCase,
) *WalletHandler {
	return &WalletHandler{
		getSummary: getSummary, getStatement: getStatement, listTransactions: listTransactions,
		getBankAccount: getBankAccount, setBankAccount: setBankAccount,
		createPayout: createPayout, listMyPayouts: listMyPayouts, getDriverKYC: getDriverKYC,
	}
}

func (h *WalletHandler) configured() bool { return h != nil && h.getSummary != nil }

func (h *WalletHandler) unavailable(w http.ResponseWriter) bool {
	if h.configured() {
		return false
	}
	writeJSON(w, http.StatusServiceUnavailable, errorResponse{Error: "wallet service not configured"})
	return true
}

func claimsDriverID(w http.ResponseWriter, r *http.Request) (string, bool) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		writeBadRequest(w, "missing auth claims")
		return "", false
	}
	return claims.UserID, true
}

// ─── GET /api/v1/driver/wallet/summary — Phần 3/7 ──────────────────────────

func (h *WalletHandler) GetSummary(w http.ResponseWriter, r *http.Request) {
	if h.unavailable(w) {
		return
	}
	driverID, ok := claimsDriverID(w, r)
	if !ok {
		return
	}
	summary, err := h.getSummary.Execute(r.Context(), driverID)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, walletSummaryJSON(summary))
}

func walletSummaryJSON(s *walletapp.WalletSummary) map[string]any {
	return map[string]any{
		"currency":                 s.Currency,
		"available_cents":          s.AvailableCents,
		"pending_cents":            s.PendingCents,
		"outstanding_cents":        s.OutstandingCents,
		"net_cents":                s.NetCents,
		"lifetime_earned_cents":    s.LifetimeEarnedCents,
		"lifetime_withdrawn_cents": s.LifetimeWithdrawnCents,
	}
}

// ─── GET /api/v1/driver/wallet/statement?from=&to= — Phần 9 ────────────────

func (h *WalletHandler) GetStatement(w http.ResponseWriter, r *http.Request) {
	if h.unavailable(w) {
		return
	}
	driverID, ok := claimsDriverID(w, r)
	if !ok {
		return
	}
	from, _ := strconv.ParseInt(r.URL.Query().Get("from"), 10, 64)
	to, _ := strconv.ParseInt(r.URL.Query().Get("to"), 10, 64)
	st, err := h.getStatement.Execute(r.Context(), driverID, from, to)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	lines := make([]map[string]any, len(st.Lines))
	for i, l := range st.Lines {
		lines[i] = map[string]any{
			"settlement_id":       l.SettlementID,
			"trip_id":             l.TripID,
			"trip_type":           string(l.TripType),
			"payment_method":      string(l.PaymentMethod),
			"fare_cents":          l.FareAmountCents,
			"commission_cents":    l.CommissionAmountCents,
			"driver_income_cents": l.DriverIncomeCents,
			"created_at":          time.Unix(l.CreatedAt, 0).UTC().Format(time.RFC3339),
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"currency":                st.Currency,
		"ride_income_cents":       st.RideIncomeCents,
		"delivery_income_cents":   st.DeliveryIncomeCents,
		"commission_cents":        st.CommissionCents,
		"promotion_cents":         st.PromotionCents,
		"voucher_cents":           st.VoucherCents,
		"cash_income_cents":       st.CashIncomeCents,
		"electronic_income_cents": st.ElectronicIncomeCents,
		"withdrawal_cents":        st.WithdrawalCents,
		"outstanding_cents":       st.OutstandingCents,
		"lines":                   lines,
	})
}

// ─── GET /api/v1/driver/wallet/transactions?limit= — Phần 6/9 ──────────────

func (h *WalletHandler) ListTransactions(w http.ResponseWriter, r *http.Request) {
	if h.unavailable(w) {
		return
	}
	driverID, ok := claimsDriverID(w, r)
	if !ok {
		return
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	txs, err := h.listTransactions.Execute(r.Context(), driverID, limit)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	items := make([]map[string]any, len(txs))
	for i, t := range txs {
		items[i] = map[string]any{
			"type":           string(t.Type),
			"direction":      string(t.Direction),
			"amount_cents":   t.AmountCents,
			"currency":       t.Currency,
			"description":    t.Description,
			"payment_method": t.PaymentMethod,
			"created_at":     time.Unix(t.CreatedAt, 0).UTC().Format(time.RFC3339),
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"transactions": items})
}

// ─── Bank Account — Phần 6 ───────────────────────────────────────────────────

// GetBankAccount handles GET /api/v1/driver/wallet/bank-account. Never
// exposes the raw account number (Phần 10/12) — only the masked form.
func (h *WalletHandler) GetBankAccount(w http.ResponseWriter, r *http.Request) {
	if h.unavailable(w) {
		return
	}
	driverID, ok := claimsDriverID(w, r)
	if !ok {
		return
	}
	b, err := h.getBankAccount.Execute(r.Context(), driverID)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, bankAccountJSON(b))
}

type setBankAccountRequest struct {
	BankName          string `json:"bank_name"`
	AccountHolderName string `json:"account_holder_name"`
	AccountNumber     string `json:"account_number"`
	BranchName        string `json:"branch_name"`
}

// SetBankAccount handles POST /api/v1/driver/wallet/bank-account.
func (h *WalletHandler) SetBankAccount(w http.ResponseWriter, r *http.Request) {
	if h.unavailable(w) {
		return
	}
	driverID, ok := claimsDriverID(w, r)
	if !ok {
		return
	}
	var req setBankAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeBadRequest(w, "invalid request body")
		return
	}
	b, err := h.setBankAccount.Execute(r.Context(), walletapp.SetBankAccountInput{
		DriverID: driverID, BankName: req.BankName, AccountHolderName: req.AccountHolderName,
		AccountNumber: req.AccountNumber, BranchName: req.BranchName,
	})
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, bankAccountJSON(b))
}

func bankAccountJSON(b *walletentity.BankAccount) map[string]any {
	return map[string]any{
		"bank_name":             b.BankName,
		"account_holder_name":   b.AccountHolderName,
		"masked_account_number": b.MaskedAccountNumber(),
		"branch_name":           b.BranchName,
		"updated_at":            b.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

// ─── Payout Request — Phần 5/8 ───────────────────────────────────────────────

type createPayoutRequestRequest struct {
	AmountCents int64 `json:"amount_cents"`
}

// CreatePayoutRequest handles POST /api/v1/driver/wallet/payouts (Phần 5).
// KYC eligibility is checked here (the gateway already owns the driver KYC
// use cases — see the module report's Kiến trúc section) before delegating
// to CreatePayoutRequestUseCase, which independently validates the
// remaining checks (in-flight request, bank account, balance, minimum).
func (h *WalletHandler) CreatePayoutRequest(w http.ResponseWriter, r *http.Request) {
	if h.unavailable(w) {
		return
	}
	driverID, ok := claimsDriverID(w, r)
	if !ok {
		return
	}
	if h.getDriverKYC != nil {
		dv, err := h.getDriverKYC.Execute(r.Context(), driverID)
		if err != nil || dv.Status != driverentity.KYCApproved {
			msg := "Bạn cần hoàn thành xác minh KYC trước khi rút tiền"
			if err == nil && dv.Status == driverentity.KYCExpired {
				msg = "KYC của bạn đã hết hạn, vui lòng xác minh lại"
			}
			writeJSON(w, http.StatusUnprocessableEntity, errorResponse{Error: msg})
			return
		}
	}
	var req createPayoutRequestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeBadRequest(w, "invalid request body")
		return
	}
	p, err := h.createPayout.Execute(r.Context(), walletapp.CreatePayoutRequestInput{DriverID: driverID, AmountCents: req.AmountCents})
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, payoutRequestJSON(p))
}

// ListMyPayoutRequests handles GET /api/v1/driver/wallet/payouts.
func (h *WalletHandler) ListMyPayoutRequests(w http.ResponseWriter, r *http.Request) {
	if h.unavailable(w) {
		return
	}
	driverID, ok := claimsDriverID(w, r)
	if !ok {
		return
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	list, err := h.listMyPayouts.Execute(r.Context(), driverID, limit)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	items := make([]map[string]any, len(list))
	for i, p := range list {
		items[i] = payoutRequestJSON(p)
	}
	writeJSON(w, http.StatusOK, map[string]any{"payout_requests": items})
}

func payoutRequestJSON(p *walletentity.PayoutRequest) map[string]any {
	body := map[string]any{
		"payout_request_id":     p.PayoutRequestID,
		"amount_cents":          p.AmountCents,
		"currency":              p.Currency,
		"bank_name":             p.BankName,
		"masked_account_number": p.AccountNumberMasked,
		"status":                string(p.Status),
		"requested_at":          p.RequestedAt.UTC().Format(time.RFC3339),
	}
	if p.ReviewedAt != nil {
		body["reviewed_at"] = p.ReviewedAt.UTC().Format(time.RFC3339)
	}
	if p.RejectReason != "" {
		body["reject_reason"] = p.RejectReason
	}
	if p.PaidAt != nil {
		body["paid_at"] = p.PaidAt.UTC().Format(time.RFC3339)
	}
	return body
}
