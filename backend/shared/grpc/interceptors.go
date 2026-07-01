package grpc

import (
	"context"
	"runtime/debug"
	"time"

	"github.com/rs/zerolog"
	stdgrpc "google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ─── Unary interceptors ───────────────────────────────────────────────────────

// LoggingInterceptor logs each unary RPC with method, duration, and error (if any).
func LoggingInterceptor(log zerolog.Logger) stdgrpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *stdgrpc.UnaryServerInfo,
		handler stdgrpc.UnaryHandler,
	) (any, error) {
		start := time.Now()
		resp, err := handler(ctx, req)
		dur := time.Since(start)

		ev := log.Info()
		if err != nil {
			ev = log.Error().Err(err)
		}
		ev.Str("method", info.FullMethod).
			Dur("duration_ms", dur).
			Msg("rpc")

		return resp, err
	}
}

// RecoveryInterceptor catches panics in unary handlers, logs them, and returns INTERNAL.
func RecoveryInterceptor(log zerolog.Logger) stdgrpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *stdgrpc.UnaryServerInfo,
		handler stdgrpc.UnaryHandler,
	) (resp any, err error) {
		defer func() {
			if r := recover(); r != nil {
				log.Error().
					Interface("panic", r).
					Str("method", info.FullMethod).
					Bytes("stack", debug.Stack()).
					Msg("gRPC panic recovered")
				err = status.Errorf(codes.Internal, "internal server error")
			}
		}()
		return handler(ctx, req)
	}
}

// ─── Stream interceptors ──────────────────────────────────────────────────────

// StreamLoggingInterceptor logs each streaming RPC at the point of establishment.
func StreamLoggingInterceptor(log zerolog.Logger) stdgrpc.StreamServerInterceptor {
	return func(
		srv any,
		ss stdgrpc.ServerStream,
		info *stdgrpc.StreamServerInfo,
		handler stdgrpc.StreamHandler,
	) error {
		start := time.Now()
		err := handler(srv, ss)
		dur := time.Since(start)

		ev := log.Info()
		if err != nil {
			ev = log.Error().Err(err)
		}
		ev.Str("method", info.FullMethod).
			Dur("duration_ms", dur).
			Msg("stream")

		return err
	}
}

// StreamRecoveryInterceptor catches panics in streaming handlers and returns INTERNAL.
func StreamRecoveryInterceptor(log zerolog.Logger) stdgrpc.StreamServerInterceptor {
	return func(
		srv any,
		ss stdgrpc.ServerStream,
		info *stdgrpc.StreamServerInfo,
		handler stdgrpc.StreamHandler,
	) (err error) {
		defer func() {
			if r := recover(); r != nil {
				log.Error().
					Interface("panic", r).
					Str("method", info.FullMethod).
					Bytes("stack", debug.Stack()).
					Msg("gRPC stream panic recovered")
				err = status.Errorf(codes.Internal, "internal server error")
			}
		}()
		return handler(srv, ss)
	}
}
