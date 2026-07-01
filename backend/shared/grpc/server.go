// Package grpc provides a production-ready gRPC server builder for FAIRRIDE services.
package grpc

import (
	"fmt"
	"net"
	"time"

	"github.com/rs/zerolog"
	stdgrpc "google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"
)

// Options configures the gRPC server at construction time.
type Options struct {
	// EnableReflection enables the gRPC server reflection API (grpcurl, gRPC UI).
	// Must be false in production — reflection exposes the full API surface.
	EnableReflection bool

	// MaxRecvMsgSizeMB caps incoming message size. Default: 4 MB.
	MaxRecvMsgSizeMB int

	// MaxSendMsgSizeMB caps outgoing message size. Default: 4 MB.
	MaxSendMsgSizeMB int
}

// DefaultOptions returns sensible defaults for non-production environments.
func DefaultOptions() Options {
	return Options{
		EnableReflection: true,
		MaxRecvMsgSizeMB: 4,
		MaxSendMsgSizeMB: 4,
	}
}

// Server wraps a gRPC server with standard health check and reflection endpoints.
type Server struct {
	inner  *stdgrpc.Server
	health *health.Server
	addr   string
	log    zerolog.Logger
}

// NewServer creates a gRPC server pre-configured with:
//   - Logging and panic-recovery interceptors (unary + stream)
//   - keepalive settings tuned for mobile clients
//   - Health check service (grpc.health.v1)
//   - Server reflection — only when opts.EnableReflection is true
func NewServer(addr string, log zerolog.Logger, opts Options) *Server {
	if opts.MaxRecvMsgSizeMB <= 0 {
		opts.MaxRecvMsgSizeMB = 4
	}
	if opts.MaxSendMsgSizeMB <= 0 {
		opts.MaxSendMsgSizeMB = 4
	}

	base := []stdgrpc.ServerOption{
		stdgrpc.ChainUnaryInterceptor(
			RecoveryInterceptor(log),
			LoggingInterceptor(log),
		),
		stdgrpc.ChainStreamInterceptor(
			StreamRecoveryInterceptor(log),
			StreamLoggingInterceptor(log),
		),
		stdgrpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle:     5 * time.Minute,
			MaxConnectionAge:      2 * time.Hour, // mobile-friendly: long-lived connections
			MaxConnectionAgeGrace: 30 * time.Second,
			Time:                  2 * time.Hour,
			Timeout:               20 * time.Second,
		}),
		stdgrpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             30 * time.Second,
			PermitWithoutStream: true,
		}),
		stdgrpc.MaxRecvMsgSize(opts.MaxRecvMsgSizeMB * 1024 * 1024),
		stdgrpc.MaxSendMsgSize(opts.MaxSendMsgSizeMB * 1024 * 1024),
	}

	srv := stdgrpc.NewServer(base...)
	h := health.NewServer()
	healthpb.RegisterHealthServer(srv, h)

	if opts.EnableReflection {
		reflection.Register(srv)
	}

	h.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)

	return &Server{inner: srv, health: h, addr: addr, log: log}
}

// Register calls fn with the underlying *grpc.Server for registering service handlers.
func (s *Server) Register(fn func(*stdgrpc.Server)) {
	fn(s.inner)
}

// Inner returns the underlying *grpc.Server for advanced use cases.
func (s *Server) Inner() *stdgrpc.Server {
	return s.inner
}

// Serve starts accepting connections. This call blocks until the server stops.
func (s *Server) Serve() error {
	lis, err := net.Listen("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("grpc listen on %s: %w", s.addr, err)
	}
	s.log.Info().Str("addr", s.addr).Msg("gRPC server listening")
	return s.inner.Serve(lis)
}

// GracefulStop signals the health check as NOT_SERVING then drains in-flight RPCs.
func (s *Server) GracefulStop() {
	s.health.SetServingStatus("", healthpb.HealthCheckResponse_NOT_SERVING)
	s.health.Shutdown()
	s.inner.GracefulStop()
}

// SetServingStatus updates the health check status for a named service.
func (s *Server) SetServingStatus(service string, status healthpb.HealthCheckResponse_ServingStatus) {
	s.health.SetServingStatus(service, status)
}
