package main

import (
	"context"

	sharedconfig "github.com/fairride/shared/config"
	sharedgrpc "github.com/fairride/shared/grpc"
	sharedrds "github.com/fairride/shared/redis"
	"github.com/fairride/shared/server"
	"github.com/fairride/driver/app"
	drivergrpc "github.com/fairride/driver/grpc"
	"github.com/fairride/driver/grpc/driverpb"
	"github.com/fairride/driver/infrastructure/postgres"
	driverredis "github.com/fairride/driver/infrastructure/redis"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	server.Run("driver", register)
}

func register(srv *sharedgrpc.Server, ready *server.ReadinessTracker) {
	cfg := sharedconfig.Load("driver")

	ready.Set("db", false)
	pool, err := pgxpool.New(context.Background(), cfg.DB.URL)
	if err == nil {
		err = pool.Ping(context.Background())
	}
	ready.Set("db", err == nil)

	driverRepo := postgres.NewDriverRepository(pool)

	getUC := app.NewGetDriverProfileUseCase(driverRepo)
	getByUserUC := app.NewGetDriverProfileByUserIDUseCase(driverRepo)
	updateUC := app.NewUpdateDriverProfileUseCase(driverRepo)
	onlineUC := app.NewUpdateOnlineStatusUseCase(driverRepo)
	verificationUC := app.NewUpdateVerificationStatusUseCase(driverRepo)

	handler := drivergrpc.NewHandler(getUC, getByUserUC, updateUC, onlineUC, verificationUC)
	driverpb.RegisterDriverProfileServiceServer(srv.Inner(), handler)

	vehicleRepo := postgres.NewVehicleRepository(pool)
	vehicleHandler := drivergrpc.NewVehicleHandler(
		app.NewCreateVehicleUseCase(vehicleRepo),
		app.NewUpdateVehicleUseCase(vehicleRepo),
		app.NewDeleteVehicleUseCase(vehicleRepo),
		app.NewListVehiclesUseCase(vehicleRepo),
	)
	driverpb.RegisterVehicleServiceServer(srv.Inner(), vehicleHandler)

	ready.Set("redis", false)
	redisClient, redisErr := sharedrds.Connect(context.Background(), sharedrds.Config{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
		PoolSize: cfg.Redis.PoolSize,
	})
	ready.Set("redis", redisErr == nil)
	if redisErr == nil {
		availRepo := driverredis.NewAvailabilityRepository(redisClient)
		availHandler := drivergrpc.NewAvailabilityHandler(
			app.NewGoOnlineUseCase(availRepo),
			app.NewGoOfflineUseCase(availRepo),
			app.NewHeartbeatUseCase(availRepo),
			app.NewGetAvailabilityUseCase(availRepo),
		)
		driverpb.RegisterDriverAvailabilityServiceServer(srv.Inner(), availHandler)
	}
}
