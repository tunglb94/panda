package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fairride/dispatch/grpc/dispatchpb"
	"github.com/fairride/gateway/http/handlers"
	"github.com/fairride/gateway/http/middleware"
	"github.com/fairride/identity/infrastructure/jwt"
	"google.golang.org/grpc"
	grpccodes "google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"
)

// injectClaims mirrors how the Auth middleware stores claims in context.
func injectClaims(r *http.Request, claims *jwt.AccessClaims) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), middleware.ClaimsKey, claims))
}

// ─── stub ─────────────────────────────────────────────────────────────────────

type stubDispatchLocationClient struct {
	updateErr error
	getLat    float64
	getLon    float64
	getActive bool
	getErr    error
}

var _ handlers.DispatchLocationClient = (*stubDispatchLocationClient)(nil)

func (s *stubDispatchLocationClient) UpdateDriverLocation(_ context.Context, _ *dispatchpb.UpdateDriverLocationRequest, _ ...grpc.CallOption) (*dispatchpb.UpdateDriverLocationResponse, error) {
	return &dispatchpb.UpdateDriverLocationResponse{}, s.updateErr
}

func (s *stubDispatchLocationClient) GetDriverLocation(_ context.Context, _ *dispatchpb.GetDriverLocationRequest, _ ...grpc.CallOption) (*dispatchpb.GetDriverLocationResponse, error) {
	if s.getErr != nil {
		return nil, s.getErr
	}
	return &dispatchpb.GetDriverLocationResponse{
		Lat:      s.getLat,
		Lon:      s.getLon,
		IsActive: s.getActive,
	}, nil
}

// ─── helpers ─────────────────────────────────────────────────────────────────

func authedRequest(t *testing.T, method, path string, body any) *http.Request {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&buf).Encode(body)
	}
	r := httptest.NewRequest(method, path, &buf)
	r.Header.Set("Content-Type", "application/json")
	return injectClaims(r, &jwt.AccessClaims{UserID: "d1"})
}

func noAuthRequest(t *testing.T, method, path string, body any) *http.Request {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&buf).Encode(body)
	}
	return httptest.NewRequest(method, path, &buf)
}

// ─── UpdateLocation ──────────────────────────────────────────────────────────

func TestUpdateLocation_OK(t *testing.T) {
	h := handlers.NewLocationHandler(&stubDispatchLocationClient{})
	w := httptest.NewRecorder()
	h.UpdateLocation(w, authedRequest(t, http.MethodPost, "/api/v1/driver/location",
		map[string]float64{"lat": 10.5, "lon": 106.5}))

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestUpdateLocation_ServiceUnavailable(t *testing.T) {
	h := handlers.NewLocationHandler(nil)
	w := httptest.NewRecorder()
	h.UpdateLocation(w, authedRequest(t, http.MethodPost, "/api/v1/driver/location",
		map[string]float64{"lat": 10.0, "lon": 106.0}))

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("status = %d, want 503", w.Code)
	}
}

func TestUpdateLocation_MissingClaims(t *testing.T) {
	h := handlers.NewLocationHandler(&stubDispatchLocationClient{})
	w := httptest.NewRecorder()
	h.UpdateLocation(w, noAuthRequest(t, http.MethodPost, "/api/v1/driver/location",
		map[string]float64{"lat": 10.0, "lon": 106.0}))

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestUpdateLocation_GRPCError(t *testing.T) {
	stub := &stubDispatchLocationClient{
		updateErr: grpcstatus.Error(grpccodes.Internal, "redis down"),
	}
	h := handlers.NewLocationHandler(stub)
	w := httptest.NewRecorder()
	h.UpdateLocation(w, authedRequest(t, http.MethodPost, "/api/v1/driver/location",
		map[string]float64{"lat": 10.0, "lon": 106.0}))

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", w.Code)
	}
}

// ─── GetLocation ─────────────────────────────────────────────────────────────

func TestGetLocation_Active(t *testing.T) {
	stub := &stubDispatchLocationClient{getLat: 10.5, getLon: 106.5, getActive: true}
	h := handlers.NewLocationHandler(stub)
	w := httptest.NewRecorder()

	r := authedRequest(t, http.MethodGet, "/api/v1/driver/d1/location", nil)
	r.SetPathValue("driverID", "d1")
	h.GetLocation(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
	var body map[string]any
	_ = json.NewDecoder(w.Body).Decode(&body)
	if body["is_active"] != true {
		t.Errorf("is_active = %v, want true", body["is_active"])
	}
}

func TestGetLocation_NotActive(t *testing.T) {
	stub := &stubDispatchLocationClient{getActive: false}
	h := handlers.NewLocationHandler(stub)
	w := httptest.NewRecorder()

	r := authedRequest(t, http.MethodGet, "/api/v1/driver/d99/location", nil)
	r.SetPathValue("driverID", "d99")
	h.GetLocation(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200 (inactive driver is not an error)", w.Code)
	}
	var body map[string]any
	_ = json.NewDecoder(w.Body).Decode(&body)
	if body["is_active"] != false {
		t.Errorf("is_active = %v, want false", body["is_active"])
	}
}

func TestGetLocation_ServiceUnavailable(t *testing.T) {
	h := handlers.NewLocationHandler(nil)
	w := httptest.NewRecorder()

	r := authedRequest(t, http.MethodGet, "/api/v1/driver/d1/location", nil)
	r.SetPathValue("driverID", "d1")
	h.GetLocation(w, r)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("status = %d, want 503", w.Code)
	}
}
