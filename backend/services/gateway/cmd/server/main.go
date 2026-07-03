package main

import (
	"errors"
	"net/http"
	"os"
	"time"

	"github.com/fairride/booking/grpc/bookingpb"
	httpgateway "github.com/fairride/gateway/http"
	"github.com/fairride/gateway/http/handlers"
	"github.com/fairride/gateway/http/middleware"
	"github.com/fairride/identity/infrastructure/jwt"
	sharedconfig "github.com/fairride/shared/config"
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

	bookingAddr := envOrDefault("BOOKING_ADDR", cfg.GRPC.Addr)
	conn, err := grpc.NewClient(bookingAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal().Err(err).Str("addr", bookingAddr).Msg("failed to connect to booking service")
	}
	defer conn.Close()

	bh := handlers.NewBookingHandler(bookingpb.NewBookingServiceClient(conn))
	authMW := middleware.Auth(tokenSvc)
	router := httpgateway.NewRouter(bh, authMW, log)

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
