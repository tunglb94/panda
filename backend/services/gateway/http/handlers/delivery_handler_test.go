package handlers_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fairride/gateway/http/handlers"
	"github.com/fairride/trip/grpc/trippb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type stubDeliveryTripClient struct {
	pickupParcel     func(ctx context.Context, in *trippb.GetTripRequest, opts ...grpc.CallOption) (*trippb.TripResponse, error)
	startDelivery    func(ctx context.Context, in *trippb.GetTripRequest, opts ...grpc.CallOption) (*trippb.TripResponse, error)
	completeDelivery func(ctx context.Context, in *trippb.GetTripRequest, opts ...grpc.CallOption) (*trippb.TripResponse, error)
}

func (s *stubDeliveryTripClient) PickupParcel(ctx context.Context, in *trippb.GetTripRequest, opts ...grpc.CallOption) (*trippb.TripResponse, error) {
	return s.pickupParcel(ctx, in, opts...)
}
func (s *stubDeliveryTripClient) StartDelivery(ctx context.Context, in *trippb.GetTripRequest, opts ...grpc.CallOption) (*trippb.TripResponse, error) {
	return s.startDelivery(ctx, in, opts...)
}
func (s *stubDeliveryTripClient) CompleteDelivery(ctx context.Context, in *trippb.GetTripRequest, opts ...grpc.CallOption) (*trippb.TripResponse, error) {
	return s.completeDelivery(ctx, in, opts...)
}

func TestDeliveryHandler_PickupParcel_Success(t *testing.T) {
	stub := &stubDeliveryTripClient{
		pickupParcel: func(_ context.Context, in *trippb.GetTripRequest, _ ...grpc.CallOption) (*trippb.TripResponse, error) {
			return &trippb.TripResponse{Trip: &trippb.TripProto{
				TripId:         in.GetTripId(),
				Status:         "in_progress",
				DeliveryStatus: "PARCEL_PICKED_UP",
			}}, nil
		},
	}
	h := handlers.NewDeliveryHandler(stub)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/rides/trip-1/pickup-parcel", nil)
	req.SetPathValue("tripID", "trip-1")
	w := httptest.NewRecorder()
	h.PickupParcel(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d — body: %s", w.Code, w.Body.String())
	}
	var resp map[string]string
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp["delivery_status"] != "PARCEL_PICKED_UP" {
		t.Fatalf("want PARCEL_PICKED_UP, got %q", resp["delivery_status"])
	}
}

func TestDeliveryHandler_StartDelivery_Success(t *testing.T) {
	stub := &stubDeliveryTripClient{
		startDelivery: func(_ context.Context, in *trippb.GetTripRequest, _ ...grpc.CallOption) (*trippb.TripResponse, error) {
			return &trippb.TripResponse{Trip: &trippb.TripProto{TripId: in.GetTripId(), DeliveryStatus: "IN_DELIVERY"}}, nil
		},
	}
	h := handlers.NewDeliveryHandler(stub)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/rides/trip-1/start-delivery", nil)
	req.SetPathValue("tripID", "trip-1")
	w := httptest.NewRecorder()
	h.StartDelivery(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d — body: %s", w.Code, w.Body.String())
	}
}

func TestDeliveryHandler_CompleteDelivery_Success(t *testing.T) {
	stub := &stubDeliveryTripClient{
		completeDelivery: func(_ context.Context, in *trippb.GetTripRequest, _ ...grpc.CallOption) (*trippb.TripResponse, error) {
			return &trippb.TripResponse{Trip: &trippb.TripProto{TripId: in.GetTripId(), DeliveryStatus: "COMPLETED"}}, nil
		},
	}
	h := handlers.NewDeliveryHandler(stub)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/rides/trip-1/complete-delivery", nil)
	req.SetPathValue("tripID", "trip-1")
	w := httptest.NewRecorder()
	h.CompleteDelivery(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d — body: %s", w.Code, w.Body.String())
	}
}

func TestDeliveryHandler_GRPCError_MapsToHTTP(t *testing.T) {
	stub := &stubDeliveryTripClient{
		pickupParcel: func(_ context.Context, _ *trippb.GetTripRequest, _ ...grpc.CallOption) (*trippb.TripResponse, error) {
			return nil, status.Error(codes.FailedPrecondition, "trip not in ACCEPTED state")
		},
	}
	h := handlers.NewDeliveryHandler(stub)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/rides/trip-1/pickup-parcel", nil)
	req.SetPathValue("tripID", "trip-1")
	w := httptest.NewRecorder()
	h.PickupParcel(w, req)

	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("want 422, got %d", w.Code)
	}
}

func TestDeliveryHandler_MissingTripID(t *testing.T) {
	h := handlers.NewDeliveryHandler(&stubDeliveryTripClient{})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/rides//pickup-parcel", nil)
	w := httptest.NewRecorder()
	h.PickupParcel(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", w.Code)
	}
}

func TestDeliveryHandler_NilClient_ReturnsServiceUnavailable(t *testing.T) {
	h := handlers.NewDeliveryHandler(nil)

	for _, call := range []func(http.ResponseWriter, *http.Request){h.PickupParcel, h.StartDelivery, h.CompleteDelivery} {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/rides/trip-1/pickup-parcel", nil)
		req.SetPathValue("tripID", "trip-1")
		w := httptest.NewRecorder()
		call(w, req)
		if w.Code != http.StatusServiceUnavailable {
			t.Fatalf("want 503, got %d", w.Code)
		}
	}
}
