package httpgateway

import (
	"net/http"

	"github.com/fairride/gateway/http/handlers"
	"github.com/fairride/gateway/http/middleware"
	"github.com/rs/zerolog"
)

// NewRouter constructs the gateway HTTP mux with all routes and middleware wired up.
// /health is unauthenticated. All /api/v1/* routes require a valid Bearer JWT.
// Every request is wrapped in the logging middleware.
func NewRouter(
	bh *handlers.BookingHandler,
	authMiddleware func(http.Handler) http.Handler,
	log zerolog.Logger,
) http.Handler {
	mux := http.NewServeMux()

	// Health — no auth required.
	mux.HandleFunc("GET /health", handleHealth)

	// Booking API — all routes require authentication.
	auth := authMiddleware
	mux.Handle("POST /api/v1/rides", auth(http.HandlerFunc(bh.BookRide)))
	mux.Handle("GET /api/v1/rides/{tripID}", auth(http.HandlerFunc(bh.GetBooking)))
	mux.Handle("POST /api/v1/rides/{tripID}/accept", auth(http.HandlerFunc(bh.AcceptDispatchOffer)))
	mux.Handle("POST /api/v1/rides/{tripID}/reject", auth(http.HandlerFunc(bh.RejectDispatchOffer)))
	mux.Handle("POST /api/v1/rides/{tripID}/start", auth(http.HandlerFunc(bh.StartTrip)))
	mux.Handle("POST /api/v1/rides/{tripID}/finish", auth(http.HandlerFunc(bh.FinishTrip)))

	return middleware.Logging(log)(mux)
}

func handleHealth(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}
