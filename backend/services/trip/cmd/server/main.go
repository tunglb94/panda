package main

import (
	"context"

	sharedconfig "github.com/fairride/shared/config"
	sharedgrpc "github.com/fairride/shared/grpc"
	"github.com/fairride/shared/server"
	"github.com/fairride/trip/app"
	tripgrpc "github.com/fairride/trip/grpc"
	"github.com/fairride/trip/grpc/trippb"
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

	handler := tripgrpc.NewHandler(
		app.NewCreateTripUseCase(tripRepo),
		app.NewCancelTripUseCase(tripRepo),
		app.NewGetTripUseCase(tripRepo),
		app.NewStartTripUseCase(tripRepo),
		app.NewCompleteTripUseCase(tripRepo),
	)
	trippb.RegisterTripServiceServer(srv.Inner(), handler)
}
