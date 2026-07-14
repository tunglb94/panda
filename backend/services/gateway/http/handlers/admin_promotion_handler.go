package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	promotionapp "github.com/fairride/promotion/app"
	promotionentity "github.com/fairride/promotion/domain/entity"
)

// AdminPromotionHandler implements the Admin app's Voucher CRUD — every
// route is RequireAdmin-gated (see router.go).
type AdminPromotionHandler struct {
	create *promotionapp.CreateVoucherUseCase
	update *promotionapp.UpdateVoucherUseCase
	list   *promotionapp.ListVouchersUseCase
	get    *promotionapp.GetVoucherUseCase
	review *promotionapp.ReviewVoucherUseCase
	stats  *promotionapp.VoucherStatsUseCase
}

func NewAdminPromotionHandler(
	create *promotionapp.CreateVoucherUseCase,
	update *promotionapp.UpdateVoucherUseCase,
	list *promotionapp.ListVouchersUseCase,
	get *promotionapp.GetVoucherUseCase,
	review *promotionapp.ReviewVoucherUseCase,
	stats *promotionapp.VoucherStatsUseCase,
) *AdminPromotionHandler {
	return &AdminPromotionHandler{create: create, update: update, list: list, get: get, review: review, stats: stats}
}

func (h *AdminPromotionHandler) unavailable(w http.ResponseWriter) bool {
	if h.create == nil || h.update == nil || h.list == nil || h.get == nil || h.review == nil {
		writeJSON(w, http.StatusServiceUnavailable, errorResponse{Error: "promotion service not configured"})
		return true
	}
	return false
}

// ListVouchers handles GET /api/v1/admin/vouchers.
func (h *AdminPromotionHandler) ListVouchers(w http.ResponseWriter, r *http.Request) {
	if h.unavailable(w) {
		return
	}
	vouchers, err := h.list.Execute(r.Context())
	if err != nil {
		writeDomainError(w, err)
		return
	}
	now := time.Now()
	out := make([]map[string]any, 0, len(vouchers))
	for _, v := range vouchers {
		out = append(out, h.voucherJSONWithStats(r.Context(), v, now))
	}
	writeJSON(w, http.StatusOK, map[string]any{"vouchers": out})
}

// GetVoucher handles GET /api/v1/admin/vouchers/{id}.
func (h *AdminPromotionHandler) GetVoucher(w http.ResponseWriter, r *http.Request) {
	if h.unavailable(w) {
		return
	}
	id := r.PathValue("id")
	v, err := h.get.Execute(r.Context(), id)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, h.voucherJSONWithStats(r.Context(), v, time.Now()))
}

// voucherJSONWithStats adds Issued/Redeemed/Expired (Phase 4 admin stats) to
// adminVoucherJSON — best-effort, nil-safe: a stats lookup failure just
// omits the "stats" key rather than failing the whole voucher response
// (không cần dashboard đẹp, but the CRUD list must never break over it).
func (h *AdminPromotionHandler) voucherJSONWithStats(ctx context.Context, v *promotionentity.Voucher, now time.Time) map[string]any {
	out := adminVoucherJSON(v, now)
	if h.stats == nil {
		return out
	}
	if s, err := h.stats.Execute(ctx, v.ID); err == nil {
		out["stats"] = map[string]any{
			"issued":    s.Issued,
			"redeemed":  s.Redeemed,
			"remaining": v.RemainingBudget,
			"expired":   s.Expired,
		}
	}
	return out
}

type voucherFormRequest struct {
	Code            string   `json:"code"`
	Title           string   `json:"title"`
	Description     string   `json:"description"`
	Type            string   `json:"type"` // "fixed" or "percentage" — mapped to entity.DiscountType below
	Value           int64    `json:"value"`
	MaxDiscount     int64    `json:"max_discount"`
	MinOrder        int64    `json:"min_order"`
	Start           string   `json:"start"` // RFC3339
	End             string   `json:"end"`   // RFC3339
	UsageLimit      int64    `json:"usage_limit"`
	PerUserLimit    int64    `json:"per_user_limit"`
	Budget          int64    `json:"budget"`
	ServiceType     []string `json:"service_type"`
	TripType        []string `json:"trip_type"`
	Campaign        string   `json:"campaign"`
	Enabled         bool     `json:"enabled"`
}

// discountTypeFromWire maps the Admin form's "fixed"/"percentage" onto
// entity.DiscountType's "flat"/"percentage" — the form's field is literally
// named "type" per the sprint brief, kept distinct from entity naming here
// rather than renaming the domain type to match a UI label.
func discountTypeFromWire(raw string) promotionentity.DiscountType {
	if raw == "fixed" {
		return promotionentity.DiscountTypeFlat
	}
	return promotionentity.DiscountType(raw) // "percentage" (or an invalid value — validated downstream)
}

// CreateVoucher handles POST /api/v1/admin/vouchers.
func (h *AdminPromotionHandler) CreateVoucher(w http.ResponseWriter, r *http.Request) {
	if h.unavailable(w) {
		return
	}
	var req voucherFormRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeBadRequest(w, "invalid request body")
		return
	}
	start, end, err := parseVoucherWindow(req.Start, req.End)
	if err != nil {
		writeBadRequest(w, err.Error())
		return
	}
	v, err := h.create.Execute(r.Context(), promotionapp.CreateVoucherInput{
		Code:            req.Code,
		Name:            req.Title,
		Description:     req.Description,
		DiscountType:    discountTypeFromWire(req.Type),
		DiscountValue:   req.Value,
		MaxDiscount:     req.MaxDiscount,
		MinOrder:        req.MinOrder,
		StartTime:       start,
		EndTime:         end,
		MaxUsage:        req.UsageLimit,
		MaxUsagePerUser: req.PerUserLimit,
		Budget:          req.Budget,
		ServiceTypes:    req.ServiceType,
		TripTypes:       req.TripType,
		Campaign:        req.Campaign,
		Enabled:         req.Enabled,
	})
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, adminVoucherJSON(v, time.Now()))
}

// UpdateVoucher handles PUT /api/v1/admin/vouchers/{id}.
func (h *AdminPromotionHandler) UpdateVoucher(w http.ResponseWriter, r *http.Request) {
	if h.unavailable(w) {
		return
	}
	id := r.PathValue("id")
	var req voucherFormRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeBadRequest(w, "invalid request body")
		return
	}
	start, end, err := parseVoucherWindow(req.Start, req.End)
	if err != nil {
		writeBadRequest(w, err.Error())
		return
	}
	v, err := h.update.Execute(r.Context(), promotionapp.UpdateVoucherInput{
		ID:              id,
		Code:            req.Code,
		Name:            req.Title,
		Description:     req.Description,
		DiscountType:    discountTypeFromWire(req.Type),
		DiscountValue:   req.Value,
		MaxDiscount:     req.MaxDiscount,
		MinOrder:        req.MinOrder,
		StartTime:       start,
		EndTime:         end,
		MaxUsage:        req.UsageLimit,
		MaxUsagePerUser: req.PerUserLimit,
		ServiceTypes:    req.ServiceType,
		TripTypes:       req.TripType,
		Campaign:        req.Campaign,
	})
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, adminVoucherJSON(v, time.Now()))
}

// EnableVoucher handles POST /api/v1/admin/vouchers/{id}/enable.
func (h *AdminPromotionHandler) EnableVoucher(w http.ResponseWriter, r *http.Request) {
	h.reviewAction(w, r, h.review.Enable)
}

// DisableVoucher handles POST /api/v1/admin/vouchers/{id}/disable.
func (h *AdminPromotionHandler) DisableVoucher(w http.ResponseWriter, r *http.Request) {
	h.reviewAction(w, r, h.review.Disable)
}

// DeleteVoucher handles DELETE /api/v1/admin/vouchers/{id} — soft-cancels
// (see ReviewVoucherUseCase.Delete's doc comment); never a hard SQL DELETE.
func (h *AdminPromotionHandler) DeleteVoucher(w http.ResponseWriter, r *http.Request) {
	h.reviewAction(w, r, h.review.Delete)
}

func (h *AdminPromotionHandler) reviewAction(w http.ResponseWriter, r *http.Request, action func(ctx context.Context, id string) (*promotionentity.Voucher, error)) {
	if h.unavailable(w) {
		return
	}
	id := r.PathValue("id")
	v, err := action(r.Context(), id)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, adminVoucherJSON(v, time.Now()))
}

func parseVoucherWindow(startRaw, endRaw string) (time.Time, time.Time, error) {
	start, err := time.Parse(time.RFC3339, startRaw)
	if err != nil {
		return time.Time{}, time.Time{}, errors.New("start must be an RFC3339 timestamp")
	}
	end, err := time.Parse(time.RFC3339, endRaw)
	if err != nil {
		return time.Time{}, time.Time{}, errors.New("end must be an RFC3339 timestamp")
	}
	return start, end, nil
}

// adminVoucherJSON is the full Admin-facing shape of a voucher — includes
// budget internals a rider never sees (see riderVoucherJSON for the
// stripped-down rider-facing shape).
func adminVoucherJSON(v *promotionentity.Voucher, now time.Time) map[string]any {
	discountTypeWire := "percentage"
	if v.DiscountType == promotionentity.DiscountTypeFlat {
		discountTypeWire = "fixed"
	}
	return map[string]any{
		"id":                v.ID,
		"code":              v.Code,
		"title":             v.Name,
		"description":       v.Description,
		"type":              discountTypeWire,
		"value":             v.DiscountValue,
		"max_discount":      v.MaxDiscount,
		"min_order":         v.MinOrder,
		"start":             v.StartTime.UTC().Format(time.RFC3339),
		"end":               v.EndTime.UTC().Format(time.RFC3339),
		"usage_limit":       v.MaxUsage,
		"per_user_limit":    v.MaxUsagePerUser,
		"budget":            v.Budget,
		"remaining_budget":  v.RemainingBudget,
		"service_type":      v.ServiceTypes,
		"trip_type":         v.TripTypes,
		"campaign":          v.Campaign,
		"enabled":           v.Status == promotionentity.VoucherStatusActive,
		"status":            string(v.Status),
		"state":             v.EffectiveState(now),
		"usage_count":       v.UsageCount,
		"created_at":        v.CreatedAt.UTC().Format(time.RFC3339),
		"updated_at":        v.UpdatedAt.UTC().Format(time.RFC3339),
	}
}
