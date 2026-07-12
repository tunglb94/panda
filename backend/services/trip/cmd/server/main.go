package main

import (
	"context"

	sharedconfig "github.com/fairride/shared/config"
	sharedgrpc "github.com/fairride/shared/grpc"
	"github.com/fairride/shared/server"
	"github.com/fairride/trip/app"
	tripgrpc "github.com/fairride/trip/grpc"
	"github.com/fairride/trip/grpc/trippb"
	"github.com/fairride/trip/infrastructure/memory"
	"github.com/fairride/trip/infrastructure/postgres"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	server.Run("trip", register)
}

func register(srv *sharedgrpc.Server, ready *server.ReadinessTracker) {
	cfg := sharedconfig.Load("trip")

	ready.Set("db", false)
	pool, err := pgxpool.New(context.Background(), cfg.DB.URL)
	if err == nil {
		err = pool.Ping(context.Background())
	}
	ready.Set("db", err == nil)

	tripRepo := postgres.NewTripRepository(pool)
	// Delivery V1 Phase 2: in-memory only — no Postgres-backed
	// DeliveryRepository exists yet (docs/business/DELIVERY_V1_DESIGN.md
	// Phần 17 defers that migration; Phase 1 deliberately scoped the
	// persistence layer to in-memory for the same reason). Deliveries
	// created via BookRide do not survive a server restart yet.
	deliveryRepo := memory.NewDeliveryRepository()

	handler := tripgrpc.NewHandler(
		app.NewCreateTripUseCase(tripRepo, deliveryRepo),
		app.NewCancelTripUseCase(tripRepo),
		app.NewGetTripUseCase(tripRepo),
		app.NewMarkDriverArrivedUseCase(tripRepo),
		app.NewStartTripUseCase(tripRepo),
		app.NewCompleteTripUseCase(tripRepo),
		app.NewInitiatePaymentUseCase(tripRepo),
		app.NewPayTripUseCase(tripRepo),
		app.NewListTripsByRiderUseCase(tripRepo),
		app.NewListTripsByDriverUseCase(tripRepo),
		app.NewPickupParcelUseCase(tripRepo, deliveryRepo),
		app.NewStartDeliveryUseCase(tripRepo, deliveryRepo),
		app.NewCompleteDeliveryUseCase(tripRepo, deliveryRepo),
		app.NewAcceptDeliveryUseCase(tripRepo, deliveryRepo),
	)
	trippb.RegisterTripServiceServer(srv.Inner(), handler)
}
