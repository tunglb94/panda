package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"time"

	"github.com/fairride/booking/grpc/bookingpb"
	"github.com/fairride/dispatch/grpc/dispatchpb"
	driverpostgres "github.com/fairride/driver/infrastructure/postgres"
	"github.com/fairride/driver/grpc/driverpb"
	httpgateway "github.com/fairride/gateway/http"
	"github.com/fairride/gateway/http/handlers"
	"github.com/fairride/gateway/http/middleware"
	identitypostgres "github.com/fairride/identity/infrastructure/postgres"
	"github.com/fairride/identity/infrastructure/jwt"
	sharedconfig "github.com/fairride/shared/config"
	"github.com/fairride/shared/database"
	"github.com/fairride/shared/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	cfg := sharedconfig.Load("gateway")
	log := logger.FromConfig(cfg.LogLevel, "gateway", cfg.Environment)

	jwtCfg := jwt.Config{
		AccessSecret:    mustEnv("JWT_ACCESS_SECRET"),
		RefreshSecret:   mustEnv("JWT_REFRESH_SECRET"),
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 7 * 24 * time.Hour,
	}
	tokenSvc, err := jwt.NewTokenService(jwtCfg)
	if err != nil {
		log.Fatal().Err(err).Msg("invalid JWT config")
	}

	// Booking service.
	bookingAddr := envOrDefault("BOOKING_ADDR", cfg.GRPC.Addr)
	bookingConn, err := grpc.NewClient(bookingAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal().Err(err).Str("addr", bookingAddr).Msg("failed to connect to booking service")
	}
	defer bookingConn.Close()
	bh := handlers.NewBookingHandler(bookingpb.NewBookingServiceClient(bookingConn))

	// Auth handler: requires a shared DB for user + driver lookups.
	// If DB_URL is unset or the connection fails, auth returns 503 gracefully.
	var ah *handlers.AuthHandler
	if dbURL := os.Getenv("DB_URL"); dbURL != "" {
		pool, dbErr := database.Connect(context.Background(), database.Config{
			URL:      dbURL,
			MaxConns: 5,
			MinConns: 1,
		})
		if dbErr != nil {
			log.Warn().Err(dbErr).Msg("gateway: DB connection failed — auth will return 503")
		} else {
			ah = handlers.NewAuthHandler(
				identitypostgres.NewUserRepository(pool),
				driverpostgres.NewDriverRepository(pool),
				tokenSvc,
			)
		}
	}
	if ah == nil {
		ah = handlers.NewAuthHandler(nil, nil, tokenSvc)
	}

	// Driver availability: proxies to the driver gRPC service.
	// If DRIVER_ADDR is unset or the connection fails, availability returns 503 gracefully.
	var avh *handlers.AvailabilityHandler
	if driverAddr := os.Getenv("DRIVER_ADDR"); driverAddr != "" {
		driverConn, connErr := grpc.NewClient(driverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if connErr != nil {
			log.Warn().Err(connErr).Str("addr", driverAddr).Msg("gateway: driver service connection failed — availability will return 503")
		} else {
			defer driverConn.Close()
			avh = handlers.NewAvailabilityHandler(driverpb.NewDriverAvailabilityServiceClient(driverConn))
		}
	}
	if avh == nil {
		avh = handlers.NewAvailabilityHandler(nil)
	}

	// Dispatch service: driver location upload (Phase 24) + rider location query (Phase 25).
	// If DISPATCH_ADDR is unset, location endpoints return 503 gracefully.
	var lh *handlers.LocationHandler
	if dispatchAddr := os.Getenv("DISPATCH_ADDR"); dispatchAddr != "" {
		dispatchConn, connErr := grpc.NewClient(dispatchAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if connErr != nil {
			log.Warn().Err(connErr).Str("addr", dispatchAddr).Msg("gateway: dispatch service connection failed — location will return 503")
		} else {
			defer dispatchConn.Close()
			lh = handlers.NewLocationHandler(dispatchpb.NewDispatchServiceClient(dispatchConn))
		}
	}
	if lh == nil {
		lh = handlers.NewLocationHandler(nil)
	}

	authMW := middleware.Auth(tokenSvc)
	router := httpgateway.NewRouter(bh, ah, avh, lh, authMW, log)

	addr := cfg.HTTP.Addr
	srv := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
		IdleTimeout:  cfg.HTTP.IdleTimeout,
	}
	log.Info().Str("addr", addr).Msg("gateway listening")
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal().Err(err).Msg("gateway exited with error")
	}
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic("required environment variable not set: " + key)
	}
	return v
}

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
