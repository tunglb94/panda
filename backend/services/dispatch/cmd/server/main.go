package main

import (
	"context"

	sharedconfig "github.com/fairride/shared/config"
	sharedgrpc "github.com/fairride/shared/grpc"
	sharedrds "github.com/fairride/shared/redis"
	"github.com/fairride/shared/server"
	"github.com/fairride/dispatch/app"
	dispatchgrpc "github.com/fairride/dispatch/grpc"
	"github.com/fairride/dispatch/grpc/dispatchpb"
	dispatchpostgres "github.com/fairride/dispatch/infrastructure/postgres"
	dispatchredis "github.com/fairride/dispatch/infrastructure/redis"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	server.Run("dispatch", register)
}

func register(srv *sharedgrpc.Server, ready *server.ReadinessTracker) {
	cfg := sharedconfig.Load("dispatch")

	ready.Set("db", false)
	pool, err := pgxpool.New(context.Background(), cfg.DB.URL)
	if err == nil {
		err = pool.Ping(context.Background())
	}
	ready.Set("db", err == nil)

	ready.Set("redis", false)
	redisClient, redisErr := sharedrds.Connect(context.Background(), sharedrds.Config{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
		PoolSize: cfg.Redis.PoolSize,
	})
	ready.Set("redis", redisErr == nil)

	jobRepo := dispatchpostgres.NewDispatchRepository(pool)
	tripUpdater := dispatchpostgres.NewTripUpdater(pool)
	transactor := dispatchpostgres.NewTransactor(pool)

	var locationRepo *dispatchredis.DriverLocationRepository
	if redisErr == nil {
		locationRepo = dispatchredis.NewDriverLocationRepository(redisClient)
	} else {
		locationRepo = dispatchredis.NewDriverLocationRepository(nil) // will fail at runtime
	}

	requestDispatch := app.NewRequestDispatchUseCase(jobRepo, locationRepo, transactor)
	acceptTrip := app.NewAcceptTripUseCase(jobRepo, transactor)
	rejectTrip := app.NewRejectTripUseCase(jobRepo, locationRepo, tripUpdater)
	updateLocation := app.NewUpdateDriverLocationUseCase(locationRepo)
	getStatus := app.NewGetDispatchStatusUseCase(jobRepo)
	getDriverOffer := app.NewGetDriverOfferUseCase(jobRepo)
	getDriverLocation := app.NewGetDriverLocationUseCase(locationRepo)

	handler := dispatchgrpc.NewHandler(
		requestDispatch,
		acceptTrip,
		rejectTrip,
		updateLocation,
		getStatus,
		getDriverOffer,
		getDriverLocation,
	)
	dispatchpb.RegisterDispatchServiceServer(srv.Inner(), handler)

	// Start the background engine only when both DB and Redis are available.
	if err == nil && redisErr == nil {
		engine := app.NewDispatchEngine(jobRepo, locationRepo, tripUpdater)
		engine.Start()
	}
}
