package handlers_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fairride/driver/grpc/driverpb"
	"github.com/fairride/gateway/http/handlers"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ─── stub client ─────────────────────────────────────────────────────────────

type stubAvailabilityClient struct {
	goOnlineResp        *driverpb.AvailabilityResponse
	goOfflineResp       *driverpb.AvailabilityResponse
	getAvailabilityResp *driverpb.AvailabilityResponse
	err                 error
}

func (s *stubAvailabilityClient) GoOnline(_ context.Context, _ *driverpb.GoOnlineRequest, _ ...grpc.CallOption) (*driverpb.AvailabilityResponse, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.goOnlineResp, nil
}
func (s *stubAvailabilityClient) GoOffline(_ context.Context, _ *driverpb.GoOfflineRequest, _ ...grpc.CallOption) (*driverpb.AvailabilityResponse, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.goOfflineResp, nil
}
func (s *stubAvailabilityClient) GetAvailability(_ context.Context, _ *driverpb.GetAvailabilityRequest, _ ...grpc.CallOption) (*driverpb.AvailabilityResponse, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.getAvailabilityResp, nil
}

// ─── helpers ─────────────────────────────────────────────────────────────────

func onlineResp() *driverpb.AvailabilityResponse {
	return &driverpb.AvailabilityResponse{DriverId: "drv-001", IsOnline: true}
}
func offlineResp() *driverpb.AvailabilityResponse {
	return &driverpb.AvailabilityResponse{DriverId: "drv-001", IsOnline: false}
}

// ─── tests ────────────────────────────────────────────────────────────────────

func TestGoOnline_Success(t *testing.T) {
	stub := &stubAvailabilityClient{goOnlineResp: onlineResp()}
	h := handlers.NewAvailabilityHandler(stub)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/driver/go-online", nil)
	req = withClaims(req, "drv-001", "driver")
	w := httptest.NewRecorder()
	h.GoOnline(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 — %s", w.Code, w.Body.String())
	}
}

func TestGoOnline_ServiceUnavailable(t *testing.T) {
	h := handlers.NewAvailabilityHandler(nil)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/driver/go-online", nil)
	req = withClaims(req, "drv-001", "driver")
	w := httptest.NewRecorder()
	h.GoOnline(w, req)
	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("status = %d, want 503", w.Code)
	}
}

func TestGoOnline_GRPCError(t *testing.T) {
	stub := &stubAvailabilityClient{err: status.Error(codes.Internal, "redis down")}
	h := handlers.NewAvailabilityHandler(stub)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/driver/go-online", nil)
	req = withClaims(req, "drv-001", "driver")
	w := httptest.NewRecorder()
	h.GoOnline(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", w.Code)
	}
}

func TestGoOffline_Success(t *testing.T) {
	stub := &stubAvailabilityClient{goOfflineResp: offlineResp()}
	h := handlers.NewAvailabilityHandler(stub)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/driver/go-offline", nil)
	req = withClaims(req, "drv-001", "driver")
	w := httptest.NewRecorder()
	h.GoOffline(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 — %s", w.Code, w.Body.String())
	}
}

func TestGetAvailability_Success(t *testing.T) {
	stub := &stubAvailabilityClient{getAvailabilityResp: onlineResp()}
	h := handlers.NewAvailabilityHandler(stub)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/driver/availability", nil)
	req = withClaims(req, "drv-001", "driver")
	w := httptest.NewRecorder()
	h.GetAvailability(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 — %s", w.Code, w.Body.String())
	}
}
