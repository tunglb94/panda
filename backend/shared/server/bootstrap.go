// Package server provides the standard FAIRRIDE service bootstrap.
// Every service calls server.Run() from its main(), which starts gRPC + HTTP,
// handles graceful shutdown, and manages readiness state.
package server

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/fairride/shared/config"
	sharedgrpc "github.com/fairride/shared/grpc"
	"github.com/fairride/shared/logger"
)

// ReadinessTracker tracks named readiness checks.
// A service is ready when every registered check passes.
// An empty tracker (no checks) is always ready — correct for Phase 1 skeletons.
//
// Usage (Phase 2+):
//
//	ready.Set("db", false)
//	pool, err := database.Connect(ctx, cfg)
//	ready.Set("db", err == nil)
type ReadinessTracker struct {
	mu     sync.RWMutex
	checks map[string]bool
}

func NewReadinessTracker() *ReadinessTracker {
	return &ReadinessTracker{checks: make(map[string]bool)}
}

// Set registers or updates a named readiness check.
func (r *ReadinessTracker) Set(name string, ok bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.checks[name] = ok
}

// IsReady returns true only if all registered checks pass (or no checks are registered).
func (r *ReadinessTracker) IsReady() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, ok := range r.checks {
		if !ok {
			return false
		}
	}
	return true
}

// RegisterFunc is called once after the gRPC server is created.
// Use it to register service-specific gRPC handlers and populate readiness checks.
// May be nil for services with no gRPC handlers (health check only).
type RegisterFunc func(srv *sharedgrpc.Server, ready *ReadinessTracker)

// Run is the standard FAIRRIDE service entry point.
//
// It:
//  1. Loads config from environment
//  2. Creates a gRPC server (reflection disabled in production)
//  3. Calls register (if non-nil) to attach service handlers
//  4. Starts an HTTP server with /health and /ready probes
//  5. Blocks until SIGINT or SIGTERM, then gracefully shuts down
func Run(serviceName string, register RegisterFunc) {
	cfg := config.Load(serviceName)
	log := logger.FromConfig(cfg.LogLevel, cfg.ServiceName, cfg.Environment)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	grpcSrv := sharedgrpc.NewServer(cfg.GRPC.Addr, log, sharedgrpc.Options{
		EnableReflection: !cfg.IsProduction(),
		MaxRecvMsgSizeMB: cfg.GRPC.MaxRecvMsgSizeMB,
		MaxSendMsgSizeMB: cfg.GRPC.MaxSendMsgSizeMB,
	})

	ready := NewReadinessTracker()
	if register != nil {
		register(grpcSrv, ready)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler(serviceName))
	mux.HandleFunc("/ready", readyHandler(ready))

	httpSrv := &http.Server{
		Addr:         cfg.HTTP.Addr,
		Handler:      mux,
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
		IdleTimeout:  cfg.HTTP.IdleTimeout,
	}

	go func() {
		if err := grpcSrv.Serve(); err != nil {
			log.Error().Err(err).Msg("gRPC server stopped unexpectedly")
			stop()
		}
	}()

	go func() {
		log.Info().Str("addr", cfg.HTTP.Addr).Msg("HTTP server listening")
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error().Err(err).Msg("HTTP server stopped unexpectedly")
			stop()
		}
	}()

	log.Info().
		Str("grpc_addr", cfg.GRPC.Addr).
		Str("http_addr", cfg.HTTP.Addr).
		Str("env", cfg.Environment).
		Bool("reflection", !cfg.IsProduction()).
		Msg("service started")

	<-ctx.Done()
	log.Info().Msg("shutdown signal received")

	grpcSrv.GracefulStop()

	shutCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := httpSrv.Shutdown(shutCtx); err != nil {
		log.Error().Err(err).Msg("HTTP server shutdown error")
	}

	log.Info().Msg("shutdown complete")
}

func healthHandler(name string) http.HandlerFunc {
	body := []byte(`{"status":"ok","service":"` + name + `"}`)
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(body)
	}
}

func readyHandler(ready *ReadinessTracker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if ready.IsReady() {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"ready":true}`))
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte(`{"ready":false}`))
		}
	}
}
