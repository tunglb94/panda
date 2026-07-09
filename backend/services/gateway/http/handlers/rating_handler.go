package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/fairride/gateway/http/middleware"
	"github.com/fairride/review/grpc/reviewpb"
	"google.golang.org/grpc"
)

// ReviewClient is the subset of reviewpb.ReviewServiceClient used by the gateway.
type ReviewClient interface {
	SubmitRating(ctx context.Context, in *reviewpb.SubmitRatingRequest, opts ...grpc.CallOption) (*reviewpb.RatingProto, error)
	GetTripRating(ctx context.Context, in *reviewpb.GetRatingRequest, opts ...grpc.CallOption) (*reviewpb.RatingProto, error)
}

// RatingHandler exposes rating operations over HTTP.
type RatingHandler struct {
	client ReviewClient
}

// NewRatingHandler constructs a RatingHandler.
// Passing nil for client makes all endpoints return 503.
func NewRatingHandler(client ReviewClient) *RatingHandler {
	return &RatingHandler{client: client}
}

type submitRatingRequest struct {
	RateeID string `json:"ratee_id"`
	Role    string `json:"role"`
	Stars   int32  `json:"stars"`
	Comment string `json:"comment"`
}

// SubmitRating handles POST /api/v1/rides/{tripID}/rate.
func (h *RatingHandler) SubmitRating(w http.ResponseWriter, r *http.Request) {
	if h.client == nil {
		writeJSON(w, http.StatusServiceUnavailable, errorResponse{Error: "review service not configured"})
		return
	}
	tripID := r.PathValue("tripID")
	if tripID == "" {
		writeBadRequest(w, "tripID is required")
		return
	}
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		writeBadRequest(w, "missing auth claims")
		return
	}
	var req submitRatingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeBadRequest(w, "invalid request body")
		return
	}
	if req.RateeID == "" {
		writeBadRequest(w, "ratee_id is required")
		return
	}
	if req.Role == "" {
		writeBadRequest(w, "role is required")
		return
	}
	if req.Stars < 1 || req.Stars > 5 {
		writeBadRequest(w, "stars must be between 1 and 5")
		return
	}
	rating, err := h.client.SubmitRating(r.Context(), &reviewpb.SubmitRatingRequest{
		TripId:  tripID,
		RaterId: claims.UserID,
		RateeId: req.RateeID,
		Role:    req.Role,
		Stars:   req.Stars,
		Comment: req.Comment,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, toRatingMap(rating))
}

// GetRating handles GET /api/v1/rides/{tripID}/rating?role=rider|driver.
func (h *RatingHandler) GetRating(w http.ResponseWriter, r *http.Request) {
	if h.client == nil {
		writeJSON(w, http.StatusServiceUnavailable, errorResponse{Error: "review service not configured"})
		return
	}
	tripID := r.PathValue("tripID")
	if tripID == "" {
		writeBadRequest(w, "tripID is required")
		return
	}
	role := r.URL.Query().Get("role")
	if role == "" {
		writeBadRequest(w, "role query parameter is required")
		return
	}
	rating, err := h.client.GetTripRating(r.Context(), &reviewpb.GetRatingRequest{
		TripId: tripID,
		Role:   role,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toRatingMap(rating))
}

func toRatingMap(r *reviewpb.RatingProto) map[string]any {
	m := map[string]any{
		"rating_id": r.GetRatingId(),
		"trip_id":   r.GetTripId(),
		"rater_id":  r.GetRaterId(),
		"ratee_id":  r.GetRateeId(),
		"role":      r.GetRole(),
		"stars":     r.GetStars(),
		"comment":   r.GetComment(),
	}
	if ts := r.GetCreatedAt(); ts != nil {
		m["created_at"] = ts.AsTime().UTC().Format(time.RFC3339)
	}
	return m
}
