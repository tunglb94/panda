package config_test

import (
	"testing"
	"time"

	"github.com/fairride/shared/config"
)

func TestLoad_DefaultValues(t *testing.T) {
	t.Setenv("SERVICE_NAME", "")
	t.Setenv("LOG_LEVEL", "")
	t.Setenv("GRPC_ADDR", "")
	t.Setenv("HTTP_ADDR", "")

	cfg := config.Load("test-service")

	cases := []struct {
		name string
		got  string
		want string
	}{
		{"ServiceName", cfg.ServiceName, "test-service"},
		{"LogLevel", cfg.LogLevel, "info"},
		{"Environment", cfg.Environment, "development"},
		{"GRPCAddr", cfg.GRPC.Addr, ":50051"},
		{"HTTPAddr", cfg.HTTP.Addr, ":8080"},
	}

	for _, tc := range cases {
		if tc.got != tc.want {
			t.Errorf("%s: got %q, want %q", tc.name, tc.got, tc.want)
		}
	}

	if cfg.DB.MaxConns != 5 {
		t.Errorf("DB.MaxConns: got %d, want 5", cfg.DB.MaxConns)
	}
	if cfg.DB.MinConns != 2 {
		t.Errorf("DB.MinConns: got %d, want 2", cfg.DB.MinConns)
	}
}

func TestLoad_EnvOverrides(t *testing.T) {
	t.Setenv("LOG_LEVEL", "debug")
	t.Setenv("GRPC_ADDR", ":9000")
	t.Setenv("HTTP_ADDR", ":9001")
	t.Setenv("ENVIRONMENT", "production")
	t.Setenv("DATABASE_MAX_CONNS", "50")
	t.Setenv("REDIS_ADDR", "redis:6379")
	t.Setenv("KAFKA_BROKERS", "kafka1:9092,kafka2:9092")

	cfg := config.Load("test-service")

	if cfg.LogLevel != "debug" {
		t.Errorf("LogLevel: got %q, want %q", cfg.LogLevel, "debug")
	}
	if cfg.GRPC.Addr != ":9000" {
		t.Errorf("GRPCAddr: got %q, want %q", cfg.GRPC.Addr, ":9000")
	}
	if cfg.HTTP.Addr != ":9001" {
		t.Errorf("HTTPAddr: got %q, want %q", cfg.HTTP.Addr, ":9001")
	}
	if !cfg.IsProduction() {
		t.Error("IsProduction: expected true for environment=production")
	}
	if cfg.DB.MaxConns != 50 {
		t.Errorf("DB.MaxConns: got %d, want 50", cfg.DB.MaxConns)
	}
	if cfg.Redis.Addr != "redis:6379" {
		t.Errorf("Redis.Addr: got %q, want %q", cfg.Redis.Addr, "redis:6379")
	}
	if len(cfg.Kafka.Brokers) != 2 {
		t.Errorf("Kafka.Brokers: got %d brokers, want 2", len(cfg.Kafka.Brokers))
	}
}

func TestLoad_HTTPTimeoutDefaults(t *testing.T) {
	t.Setenv("HTTP_READ_TIMEOUT", "")
	t.Setenv("HTTP_WRITE_TIMEOUT", "")

	cfg := config.Load("test-service")

	if cfg.HTTP.ReadTimeout != 10*time.Second {
		t.Errorf("HTTP.ReadTimeout: got %v, want 10s", cfg.HTTP.ReadTimeout)
	}
	if cfg.HTTP.WriteTimeout != 10*time.Second {
		t.Errorf("HTTP.WriteTimeout: got %v, want 10s", cfg.HTTP.WriteTimeout)
	}
}
