package main

import (
	"fmt"
	"os"

	sharedconfig "github.com/fairride/shared/config"
	sharedgrpc "github.com/fairride/shared/grpc"
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

	handler := bookinggrpc.NewHandler(
		app.NewBookRideUseCase(tripAdapter, dispatchAdapter),
		app.NewAcceptDispatchOfferUseCase(dispatchAdapter),
		app.NewRejectDispatchOfferUseCase(dispatchAdapter),
		app.NewStartTripUseCase(tripAdapter),
		app.NewFinishTripUseCase(pricingAdapter, tripAdapter),
		app.NewGetBookingDetailsUseCase(tripAdapter, dispatchAdapter),
	)
	bookingpb.RegisterBookingServiceServer(srv.Inner(), handler)
}

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
