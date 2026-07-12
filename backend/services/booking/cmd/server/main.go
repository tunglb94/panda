package main

import (
	"context"
	"fmt"
	"os"

	sharedconfig "github.com/fairride/shared/config"
	sharedgrpc "github.com/fairride/shared/grpc"
	"github.com/fairride/shared/idempotency"
	"github.com/fairride/shared/server"

	"github.com/fairride/booking/app"
	bookinggrpc "github.com/fairride/booking/grpc"
	"github.com/fairride/booking/grpc/adapters"
	"github.com/fairride/booking/grpc/bookingpb"
	dispatchpb "github.com/fairride/dispatch/grpc/dispatchpb"
	pricingpb "github.com/fairride/pricing/grpc/pricingpb"
	trippb "github.com/fairride/trip/grpc/trippb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	server.Run("booking", register)
}

func register(srv *sharedgrpc.Server, ready *server.ReadinessTracker) {
	cfg := sharedconfig.Load("booking")

	tripAddr := envOrDefault("TRIP_ADDR", cfg.GRPC.Addr)
	dispatchAddr := envOrDefault("DISPATCH_ADDR", cfg.GRPC.Addr)
	pricingAddr := envOrDefault("PRICING_ADDR", cfg.GRPC.Addr)

	ready.Set("trip", false)
	tripConn, tripErr := grpc.NewClient(tripAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	ready.Set("trip", tripErr == nil)

	ready.Set("dispatch", false)
	dispatchConn, dispatchErr := grpc.NewClient(dispatchAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	ready.Set("dispatch", dispatchErr == nil)

	ready.Set("pricing", false)
	pricingConn, pricingErr := grpc.NewClient(pricingAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	ready.Set("pricing", pricingErr == nil)

	if tripErr != nil || dispatchErr != nil || pricingErr != nil {
		fmt.Fprintf(os.Stderr, "booking: upstream connection error(s) — trip=%v dispatch=%v pricing=%v\n",
			tripErr, dispatchErr, pricingErr)
	}

	tripAdapter := adapters.NewTripAdapter(trippb.NewTripServiceClient(tripConn))
	dispatchAdapter := adapters.NewDispatchAdapter(dispatchpb.NewDispatchServiceClient(dispatchConn))
	pricingAdapter := adapters.NewPricingAdapter(pricingpb.NewPricingServiceClient(pricingConn))

	// Idempotency store: requires a PostgreSQL DB for the booking service.
	// The DB URL is read from the shared config (BOOKING_DB_URL env var or config file).
	// If the DB is unavailable the service starts without idempotency rather than refusing to boot.
	ready.Set("db", false)
	var idemStore app.IdempotencyStore
	if cfg.DB.URL != "" {
		store, closePool, err := idempotency.NewPostgresStoreFromURL(context.Background(), cfg.DB.URL)
		if err == nil {
			if initErr := store.Init(context.Background()); initErr == nil {
				idemStore = store
				ready.Set("db", true)
				// closePool is intentionally not deferred — the pool lives for the process lifetime.
				_ = closePool
			} else {
				closePool() // init failed; release the connection pool
			}
		}
	}

	bookRide := app.NewBookRideUseCase(tripAdapter, dispatchAdapter)
	acceptOffer := app.NewAcceptDispatchOfferUseCase(dispatchAdapter, tripAdapter)
	finishTrip := app.NewFinishTripUseCase(pricingAdapter, tripAdapter)
	if idemStore != nil {
		bookRide = bookRide.WithIdempotency(idemStore)
		acceptOffer = acceptOffer.WithIdempotency(idemStore)
		finishTrip = finishTrip.WithIdempotency(idemStore)
	}

	handler := bookinggrpc.NewHandler(
		bookRide,
		acceptOffer,
		app.NewRejectDispatchOfferUseCase(dispatchAdapter),
		app.NewArriveAtPickupUseCase(tripAdapter),
		app.NewStartTripUseCase(tripAdapter),
		finishTrip,
		app.NewGetBookingDetailsUseCase(tripAdapter, dispatchAdapter),
		app.NewGetDriverCurrentOfferUseCase(dispatchAdapter, tripAdapter),
		app.NewCancelRideUseCase(tripAdapter),
		app.NewPayRideUseCase(tripAdapter),
		app.NewListRiderTripsUseCase(tripAdapter),
		app.NewListDriverTripsUseCase(tripAdapter),
	)
	bookingpb.RegisterBookingServiceServer(srv.Inner(), handler)
}

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
