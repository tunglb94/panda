package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fairride/gateway/http/handlers"
	"github.com/fairride/pricing/grpc/pricingpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type stubPricingClient struct {
	resp *pricingpb.FareResponse
	err  error

	lastReq *pricingpb.EstimateFareRequest
}

func (s *stubPricingClient) EstimateFare(_ context.Context, req *pricingpb.EstimateFareRequest, _ ...grpc.CallOption) (*pricingpb.FareResponse, error) {
	s.lastReq = req
	if s.err != nil {
		return nil, s.err
	}
	return s.resp, nil
}

func fareResp() *pricingpb.FareResponse {
	return &pricingpb.FareResponse{
		Fare: &pricingpb.FareBreakdown{
			VehicleType:     "car",
			DistanceKm:      5.2,
			DurationMinutes: 12.5,
			BaseFare:        15000,
			DistanceFare:    33800,
			TimeFare:        5000,
			BookingFee:      2000,
			RideFare:        53800,
			Total:           55800,
			CurrencyCode:    "VND",
			IsFinal:         false,
		},
	}
}

func estimateFareBody(t *testing.T, serviceType string) *bytes.Reader {
	t.Helper()
	body, err := json.Marshal(map[string]any{
		"pickup_lat":      10.762622,
		"pickup_lon":      106.660172,
		"destination_lat": 10.776530,
		"destination_lon": 106.700981,
		"service_type":    serviceType,
		"trip_type":       "ride",
		"promo_code":      "",
	})
	if err != nil {
		t.Fatalf("marshal request body: %v", err)
	}
	return bytes.NewReader(body)
}

func TestEstimateFare_Success(t *testing.T) {
	stub := &stubPricingClient{resp: fareResp()}
	h := handlers.NewPricingHandler(stub)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/rides/estimate-fare", estimateFareBody(t, "car"))
	w := httptest.NewRecorder()

	h.EstimateFare(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 — %s", w.Code, w.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if body["total"] != float64(55800) {
		t.Errorf("total = %v, want 55800", body["total"])
	}
	if body["currency_code"] != "VND" {
		t.Errorf("currency_code = %v, want VND", body["currency_code"])
	}
	if body["service_type"] != "car" {
		t.Errorf("service_type = %v, want car (echoed rider tier, not backend vehicle_type)", body["service_type"])
	}
	if stub.lastReq.VehicleType != "car" {
		t.Errorf("pricing request vehicle_type = %q, want car", stub.lastReq.VehicleType)
	}
	if stub.lastReq.DistanceKm <= 0 {
		t.Errorf("pricing request distance_km = %v, want > 0 (haversine of distinct coordinates)", stub.lastReq.DistanceKm)
	}
}

func TestEstimateFare_MapsBikePlusAndCarXLToBackendVehicleTypes(t *testing.T) {
	cases := map[string]string{
		"motorcycle": "motorcycle",
		"bike_plus":  "bike_plus",
		"car":        "car",
		"car_xl":     "car_xl",
	}
	for serviceType, wantVehicleType := range cases {
		stub := &stubPricingClient{resp: fareResp()}
		h := handlers.NewPricingHandler(stub)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/rides/estimate-fare", estimateFareBody(t, serviceType))
		w := httptest.NewRecorder()

		h.EstimateFare(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("service_type=%s: status = %d, want 200 — %s", serviceType, w.Code, w.Body.String())
		}
		if stub.lastReq.VehicleType != wantVehicleType {
			t.Errorf("service_type=%s: pricing vehicle_type = %q, want %q", serviceType, stub.lastReq.VehicleType, wantVehicleType)
		}
	}
}

func TestEstimateFare_InvalidServiceType(t *testing.T) {
	stub := &stubPricingClient{resp: fareResp()}
	h := handlers.NewPricingHandler(stub)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/rides/estimate-fare", estimateFareBody(t, "helicopter"))
	w := httptest.NewRecorder()

	h.EstimateFare(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestEstimateFare_ServiceUnavailableWhenNotConfigured(t *testing.T) {
	h := handlers.NewPricingHandler(nil)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/rides/estimate-fare", estimateFareBody(t, "car"))
	w := httptest.NewRecorder()

	h.EstimateFare(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("status = %d, want 503", w.Code)
	}
}

func TestEstimateFare_GRPCError(t *testing.T) {
	stub := &stubPricingClient{err: status.Error(codes.Internal, "pricing down")}
	h := handlers.NewPricingHandler(stub)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/rides/estimate-fare", estimateFareBody(t, "car"))
	w := httptest.NewRecorder()

	h.EstimateFare(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", w.Code)
	}
}

func TestEstimateFare_InvalidBody(t *testing.T) {
	stub := &stubPricingClient{resp: fareResp()}
	h := handlers.NewPricingHandler(stub)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/rides/estimate-fare", bytes.NewReader([]byte("not json")))
	w := httptest.NewRecorder()

	h.EstimateFare(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}
