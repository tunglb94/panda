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
	dph *handlers.DriverProfileHandler,
	rh *handlers.RatingHandler,
	authMiddleware func(http.Handler) http.Handler,
	log zerolog.Logger,
) http.Handler {
	mux := http.NewServeMux()

	// Health — no auth required.
	mux.HandleFunc("GET /health", handleHealth)

	// Auth — no JWT required (issues the token).
	mux.HandleFunc("POST /api/v1/auth/login", ah.Login)
	mux.HandleFunc("POST /api/v1/auth/rider/login", ah.RiderLogin)

	auth := authMiddleware

	// Driver availability — auth required.
	mux.Handle("POST /api/v1/driver/go-online", auth(http.HandlerFunc(avh.GoOnline)))
	mux.Handle("POST /api/v1/driver/go-offline", auth(http.HandlerFunc(avh.GoOffline)))
	mux.Handle("GET /api/v1/driver/availability", auth(http.HandlerFunc(avh.GetAvailability)))

	// Driver location — auth required.
	mux.Handle("POST /api/v1/driver/location", auth(http.HandlerFunc(lh.UpdateLocation)))
	mux.Handle("GET /api/v1/driver/{driverID}/location", auth(http.HandlerFunc(lh.GetLocation)))

	// Driver profile — auth required (rider reads assigned driver's profile).
	mux.Handle("GET /api/v1/drivers/{driverID}/profile", auth(http.HandlerFunc(dph.GetDriverProfile)))

	// Driver trip offer — auth required (driver polls this endpoint).
	mux.Handle("GET /api/v1/driver/current-offer", auth(http.HandlerFunc(bh.GetDriverOffer)))

	// Trip history — auth required.
	mux.Handle("GET /api/v1/rider/trips", auth(http.HandlerFunc(bh.ListRiderTrips)))
	mux.Handle("GET /api/v1/driver/trips", auth(http.HandlerFunc(bh.ListDriverTrips)))

	// Booking API — all routes require authentication.
	mux.Handle("POST /api/v1/rides", auth(http.HandlerFunc(bh.BookRide)))
	mux.Handle("GET /api/v1/rides/{tripID}", auth(http.HandlerFunc(bh.GetBooking)))
	mux.Handle("POST /api/v1/rides/{tripID}/accept", auth(http.HandlerFunc(bh.AcceptDispatchOffer)))
	mux.Handle("POST /api/v1/rides/{tripID}/reject", auth(http.HandlerFunc(bh.RejectDispatchOffer)))
	mux.Handle("POST /api/v1/rides/{tripID}/arrive", auth(http.HandlerFunc(bh.ArriveAtPickup)))
	mux.Handle("POST /api/v1/rides/{tripID}/start", auth(http.HandlerFunc(bh.StartTrip)))
	mux.Handle("POST /api/v1/rides/{tripID}/finish", auth(http.HandlerFunc(bh.FinishTrip)))
	mux.Handle("POST /api/v1/rides/{tripID}/pay", auth(http.HandlerFunc(bh.PayRide)))
	mux.Handle("POST /api/v1/rides/{tripID}/cancel", auth(http.HandlerFunc(bh.CancelRide)))
	mux.Handle("POST /api/v1/rides/{tripID}/rate", auth(http.HandlerFunc(rh.SubmitRating)))
	mux.Handle("GET /api/v1/rides/{tripID}/rating", auth(http.HandlerFunc(rh.GetRating)))

	return middleware.Logging(log)(mux)
}

func handleHealth(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}
