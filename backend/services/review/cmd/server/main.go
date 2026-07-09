package main

import (
	"context"

	sharedconfig "github.com/fairride/shared/config"
	sharedgrpc "github.com/fairride/shared/grpc"
	"github.com/fairride/shared/server"
	"github.com/fairride/review/app"
	reviewgrpc "github.com/fairride/review/grpc"
	"github.com/fairride/review/grpc/reviewpb"
	"github.com/fairride/review/infrastructure/postgres"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	server.Run("review", register)
}

func register(srv *sharedgrpc.Server, ready *server.ReadinessTracker) {
	cfg := sharedconfig.Load("review")

	ready.Set("db", false)
	pool, err := pgxpool.New(context.Background(), cfg.DB.URL)
	if err == nil {
		err = pool.Ping(context.Background())
	}
	if err == nil {
		err = postgres.CreateSchema(context.Background(), pool)
	}
	ready.Set("db", err == nil)

	ratingRepo := postgres.NewRatingRepository(pool)
	handler := reviewgrpc.NewHandler(
		app.NewSubmitRatingUseCase(ratingRepo),
		app.NewGetTripRatingUseCase(ratingRepo),
	)
	reviewpb.RegisterReviewServiceServer(srv.Inner(), handler)
}
