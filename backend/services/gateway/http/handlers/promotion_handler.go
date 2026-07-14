package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/fairride/gateway/http/middleware"
	promotionapp "github.com/fairride/promotion/app"
	promotionentity "github.com/fairride/promotion/domain/entity"
)

// PromotionHandler exposes the Promotion Engine over HTTP — wired
// in-process into the gateway (no gRPC/.proto surface exists for
// Promotion; this environment has no protoc/buf toolchain — same
// constraint and same in-process pattern already used for
// Identity/User/Wallet, see gateway/cmd/server/main.go).
type PromotionHandler struct {
	service    *promotionapp.PromotionService
	myVouchers *promotionapp.MyVouchersUseCase
}

// NewPromotionHandler constructs a PromotionHandler. A nil service/myVouchers
// makes the corresponding endpoints return 503 — same graceful-degrade
// pattern as every other optional dependency in this gateway.
func NewPromotionHandler(service *promotionapp.PromotionService, myVouchers *promotionapp.MyVouchersUseCase) *PromotionHandler {
	return &PromotionHandler{service: service, myVouchers: myVouchers}
}

type promoEvaluateRequest struct {
	Code        string `json:"code"`
	OrderAmount int64  `json:"order_amount"`
	ServiceType string `json:"service_type"`
	TripType    string `json:"trip_type"`
	City        string `json:"city"`
}

// Validate handles POST /api/v1/promo/validate — a read-only check-as-you-type,
// no side effects (doesn't touch the voucher's budget/usage).
func (h *PromotionHandler) Validate(w http.ResponseWriter, r *http.Request) {
	h.evaluate(w, r)
}

// Apply handles POST /api/v1/promo/apply — the rider's explicit "Áp dụng"
// action. Same read-only Evaluate call as Validate; the distinction is
// purely on the client (this is the result the booking screen pins and
// later resends as voucher_code on POST /api/v1/rides). Actual redemption
// (budget/usage consumption) only happens once a trip is really booked —
// see BookingHandler.BookRide.
func (h *PromotionHandler) Apply(w http.ResponseWriter, r *http.Request) {
	h.evaluate(w, r)
}

func (h *PromotionHandler) evaluate(w http.ResponseWriter, r *http.Request) {
	if h.service == nil {
		writeJSON(w, http.StatusServiceUnavailable, errorResponse{Error: "promotion service not configured"})
		return
	}
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		writeBadRequest(w, "missing auth claims")
		return
	}
	var req promoEvaluateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeBadRequest(w, "invalid request body")
		return
	}
	if strings.TrimSpace(req.Code) == "" {
		writeBadRequest(w, "code is required")
		return
	}
	if req.OrderAmount <= 0 {
		writeBadRequest(w, "order_amount must be positive")
		return
	}

	result, err := h.service.Evaluate(r.Context(), &promotionentity.PromotionRequest{
		RiderID:     claims.UserID,
		ServiceType: req.ServiceType,
		TripType:    req.TripType,
		City:        req.City,
		OrderAmount: req.OrderAmount,
		VoucherCode: strings.TrimSpace(req.Code),
		RequestTime: time.Now(),
	}, time.Now())
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, promotionResultJSON(result))
}

func promotionResultJSON(result *promotionentity.PromotionResult) map[string]any {
	return map[string]any{
		"applied":                result.Applied,
		"voucher_id":             result.VoucherID,
		"voucher_code":           result.VoucherCode,
		"voucher_name":           result.VoucherName,
		"discount_type":          string(result.DiscountType),
		"discount_amount":        result.DiscountAmount,
		"original_order_amount":  result.OriginalOrderAmount,
		"final_order_amount":     result.FinalOrderAmount,
		"reason":                 result.Reason,
		"warnings":               result.Warnings,
	}
}

// MyVouchers handles GET /api/v1/rider/vouchers — the voucher wallet
// (Available/Used/Expired tabs).
func (h *PromotionHandler) MyVouchers(w http.ResponseWriter, r *http.Request) {
	if h.myVouchers == nil {
		writeJSON(w, http.StatusServiceUnavailable, errorResponse{Error: "promotion service not configured"})
		return
	}
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		writeBadRequest(w, "missing auth claims")
		return
	}
	result, err := h.myVouchers.Execute(r.Context(), claims.UserID)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	now := time.Now()
	available := make([]map[string]any, 0, len(result.Available))
	for _, v := range result.Available {
		available = append(available, riderVoucherJSON(v, now))
	}
	expired := make([]map[string]any, 0, len(result.Expired))
	for _, v := range result.Expired {
		expired = append(expired, riderVoucherJSON(v, now))
	}
	used := make([]map[string]any, 0, len(result.Used))
	for _, rec := range result.Used {
		used = append(used, map[string]any{
			"voucher_id":      rec.VoucherID,
			"voucher_code":    rec.VoucherCode,
			"voucher_name":    rec.VoucherName,
			"trip_id":         rec.TripID,
			"discount_amount": rec.DiscountAmount,
			"redeemed_at":     rec.RedeemedAt.UTC().Format(time.RFC3339),
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"available": available,
		"used":      used,
		"expired":   expired,
	})
}

// riderVoucherJSON is the rider-facing shape of a voucher — no budget/
// remaining-budget internals, just what a rider needs to decide whether to use it.
func riderVoucherJSON(v *promotionentity.Voucher, now time.Time) map[string]any {
	return map[string]any{
		"id":              v.ID,
		"code":            v.Code,
		"name":            v.Name,
		"description":     v.Description,
		"campaign":        v.Campaign,
		"discount_type":   string(v.DiscountType),
		"discount_value":  v.DiscountValue,
		"max_discount":    v.MaxDiscount,
		"min_order":       v.MinOrder,
		"end_time":        v.EndTime.UTC().Format(time.RFC3339),
		"service_types":   v.ServiceTypes,
		"trip_types":      v.TripTypes,
		"state":           v.EffectiveState(now),
	}
}
