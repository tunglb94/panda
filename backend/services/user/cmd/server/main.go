package main

import (
	"context"

	"github.com/fairride/shared/server"
	sharedgrpc "github.com/fairride/shared/grpc"
	"github.com/fairride/user/app"
	usergrpc "github.com/fairride/user/grpc"
	"github.com/fairride/user/grpc/userpb"
	"github.com/fairride/user/infrastructure/postgres"
	"github.com/jackc/pgx/v5/pgxpool"

	sharedconfig "github.com/fairride/shared/config"
)

func main() {
	server.Run("user", register)
}

func register(srv *sharedgrpc.Server, ready *server.ReadinessTracker) {
	cfg := sharedconfig.Load("user")

	ready.Set("db", false)
	pool, err := pgxpool.New(context.Background(), cfg.DB.URL)
	if err == nil {
		err = pool.Ping(context.Background())
	}
	ready.Set("db", err == nil)

	profileRepo := postgres.NewProfileRepository(pool)
	getUC := app.NewGetProfileUseCase(profileRepo)
	updateUC := app.NewUpdateProfileUseCase(profileRepo)
	handler := usergrpc.NewHandler(getUC, updateUC)

	userpb.RegisterUserProfileServiceServer(srv.Inner(), handler)
}
