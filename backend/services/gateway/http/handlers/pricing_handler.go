package handlers

import (
	"context"
	"encoding/json"
	"math"
	"net/http"

	"github.com/fairride/pricing/grpc/pricingpb"
	"google.golang.org/grpc"
)

// PricingClient is the subset of pricingpb.PricingServiceClient used by the gateway.
type PricingClient interface {
	EstimateFare(ctx context.Context, in *pricingpb.EstimateFareRequest, opts ...grpc.CallOption) (*pricingpb.FareResponse, error)
}

// PricingHandler exposes a pre-booking fare estimate over HTTP — the Rider
// app sends only pickup/destination coordinates plus the tier the rider
// picked (Phần "Backend là Single Source of Truth" — Flutter must not
// compute any fare itself). Pricing's real EstimateFare RPC has no
// awareness of geography (see fare_calculator.go — it takes only
// vehicle_type/distance_km/duration_minutes), so this handler is the one
// place distance/duration get derived from coordinates before calling the
// real, unmodified Pricing formula.
type PricingHandler struct {
	client PricingClient
}

// NewPricingHandler constructs a PricingHandler. Passing nil for client
// makes EstimateFare return 503.
func NewPricingHandler(client PricingClient) *PricingHandler {
	return &PricingHandler{client: client}
}

// riderServiceTypeToVehicleType maps the Rider app's 4 booking tiers onto
// Pricing's VehicleTypes (see pricing/domain/entity/fare.go) — a 1:1
// identity mapping now that Pricing has its own bike_plus/car_xl rate
// cards (kept as an explicit map, not a passthrough, so an invalid
// service_type is still rejected below rather than forwarded as-is).
var riderServiceTypeToVehicleType = map[string]string{
	"motorcycle": "motorcycle",
	"bike_plus":  "bike_plus",
	"car":        "car",
	"car_xl":     "car_xl",
}

// assumedAvgSpeedKmh is the same fallback average-speed constant the Rider
// app's now-deleted MockTripMetrics used for a straight-line ETA — there is
// still no server-side routing engine (OSRM/Directions calls happen
// client-side, only for the map polyline), so duration is estimated from
// Haversine distance the same way, just computed authoritatively here
// instead of duplicated across Flutter widgets.
const assumedAvgSpeedKmh = 25.0

type estimateFareRequest struct {
	PickupLat      float64 `json:"pickup_lat"`
	PickupLon      float64 `json:"pickup_lon"`
	DestinationLat float64 `json:"destination_lat"`
	DestinationLon float64 `json:"destination_lon"`
	ServiceType    string  `json:"service_type"`
	TripType       string  `json:"trip_type"`
	PromoCode      string  `json:"promo_code"`
}

// EstimateFare handles POST /api/v1/rides/estimate-fare. promo_code is
// accepted but not applied — Pricing's EstimateFare RPC has no
// discount/voucher input at all (confirmed: no promotion engine is wired to
// any RPC), so the response never carries a discount today; the Rider UI
// only renders one if present, never fabricates one.
func (h *PricingHandler) EstimateFare(w http.ResponseWriter, r *http.Request) {
	if h == nil || h.client == nil {
		writeJSON(w, http.StatusServiceUnavailable, errorResponse{Error: "pricing service not configured"})
		return
	}
	var req estimateFareRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeBadRequest(w, "invalid request body")
		return
	}
	vehicleType, ok := riderServiceTypeToVehicleType[req.ServiceType]
	if !ok {
		writeBadRequest(w, "service_type must be one of motorcycle, bike_plus, car, car_xl")
		return
	}

	distanceKM := haversineKM(req.PickupLat, req.PickupLon, req.DestinationLat, req.DestinationLon)
	durationMin := (distanceKM / assumedAvgSpeedKmh) * 60

	resp, err := h.client.EstimateFare(r.Context(), &pricingpb.EstimateFareRequest{
		VehicleType:     vehicleType,
		DistanceKm:      distanceKM,
		DurationMinutes: durationMin,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	fare := resp.GetFare()
	if fare == nil {
		writeJSON(w, http.StatusServiceUnavailable, errorResponse{Error: "pricing service returned no fare"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"service_type":     req.ServiceType,
		"vehicle_type":     fare.GetVehicleType(),
		"distance_km":      fare.GetDistanceKm(),
		"duration_minutes": fare.GetDurationMinutes(),
		"base_fare":        fare.GetBaseFare(),
		"distance_fare":    fare.GetDistanceFare(),
		"time_fare":        fare.GetTimeFare(),
		"booking_fee":      fare.GetBookingFee(),
		"ride_fare":        fare.GetRideFare(),
		"total":            fare.GetTotal(),
		"currency_code":    fare.GetCurrencyCode(),
		"is_final":         fare.GetIsFinal(),
	})
}

// haversineKM returns the great-circle distance between two coordinates in
// kilometers. There is no routing/geo backend anywhere in this system (the
// standalone "geo" service is an empty, uncalled stub) — this is a
// straight-line approximation, same limitation the deleted client-side
// MockTripMetrics had, now computed once, server-side, as the single source
// of truth instead of duplicated per Flutter widget.
func haversineKM(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadiusKM = 6371.0
	toRad := func(deg float64) float64 { return deg * math.Pi / 180 }
	dLat := toRad(lat2 - lat1)
	dLon := toRad(lon2 - lon1)
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(toRad(lat1))*math.Cos(toRad(lat2))*math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return earthRadiusKM * c
}
