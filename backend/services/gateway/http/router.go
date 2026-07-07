package httpgateway

import (
	"net/http"

	"github.com/fairride/gateway/http/handlers"
	"github.com/fairride/gateway/http/middleware"
	"github.com/rs/zerolog"
)

// NewRouter constructs the gateway HTTP mux with all routes and middleware wired up.
// /health and /api/v1/auth/* are unauthenticated. All other /api/v1/* routes require
// a valid Bearer JWT. Every request is wrapped in the logging middleware.
func NewRouter(
	bh *handlers.BookingHandler,
	ah *handlers.AuthHandler,
	avh *handlers.AvailabilityHandler,
	lh *handlers.LocationHandler,
	authMiddleware func(http.Handler) http.Handler,
	log zerolog.Logger,
) http.Handler {
	mux := http.NewServeMux()

	// Health — no auth required.
	mux.HandleFunc("GET /health", handleHealth)

	// Auth — no JWT required (issues the token).
	mux.HandleFunc("POST /api/v1/auth/login", ah.Login)

	// Driver availability — auth required.
	auth := authMiddleware
	mux.Handle("POST /api/v1/driver/go-online", auth(http.HandlerFunc(avh.GoOnline)))
	mux.Handle("POST /api/v1/driver/go-offline", auth(http.HandlerFunc(avh.GoOffline)))
	mux.Handle("GET /api/v1/driver/availability", auth(http.HandlerFunc(avh.GetAvailability)))

	// Driver location — auth required.
	// POST: driver uploads their current coordinates (Phase 24).
	// GET:  rider polls the assigned driver's coordinates (Phase 25).
	mux.Handle("POST /api/v1/driver/location", auth(http.HandlerFunc(lh.UpdateLocation)))
	mux.Handle("GET /api/v1/driver/{driverID}/location", auth(http.HandlerFunc(lh.GetLocation)))

	// Booking API — all routes require authentication.
	mux.Handle("POST /api/v1/rides", auth(http.HandlerFunc(bh.BookRide)))
	mux.Handle("GET /api/v1/rides/{tripID}", auth(http.HandlerFunc(bh.GetBooking)))
	mux.Handle("POST /api/v1/rides/{tripID}/accept", auth(http.HandlerFunc(bh.AcceptDispatchOffer)))
	mux.Handle("POST /api/v1/rides/{tripID}/reject", auth(http.HandlerFunc(bh.RejectDispatchOffer)))
	mux.Handle("POST /api/v1/rides/{tripID}/start", auth(http.HandlerFunc(bh.StartTrip)))
	mux.Handle("POST /api/v1/rides/{tripID}/finish", auth(http.HandlerFunc(bh.FinishTrip)))

	// Driver trip offer — auth required (driver polls this endpoint).
	mux.Handle("GET /api/v1/driver/current-offer", auth(http.HandlerFunc(bh.GetDriverOffer)))

	return middleware.Logging(log)(mux)
}

func handleHealth(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}
