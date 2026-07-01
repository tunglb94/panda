package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds all runtime configuration for a FAIRRIDE service.
// Values are read from environment variables at startup.
type Config struct {
	ServiceName string
	Environment string
	LogLevel    string

	GRPC GRPCConfig
	HTTP HTTPConfig
	DB   DBConfig
	Redis RedisConfig
	Kafka KafkaConfig
}

type GRPCConfig struct {
	Addr              string
	MaxRecvMsgSizeMB  int
	MaxSendMsgSizeMB  int
	KeepaliveInterval time.Duration
}

type HTTPConfig struct {
	Addr         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

type DBConfig struct {
	URL             string
	MaxConns        int32
	MinConns        int32
	MaxConnLifetime time.Duration
	MaxConnIdleTime time.Duration
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
	PoolSize int
}

type KafkaConfig struct {
	Brokers []string
}

// Load reads configuration from environment variables.
// serviceName is the default SERVICE_NAME if the env var is not set.
func Load(serviceName string) Config {
	return Config{
		ServiceName: envOr("SERVICE_NAME", serviceName),
		Environment: envOr("ENVIRONMENT", "development"),
		LogLevel:    envOr("LOG_LEVEL", "info"),

		GRPC: GRPCConfig{
			Addr:              envOr("GRPC_ADDR", ":50051"),
			MaxRecvMsgSizeMB:  envInt("GRPC_MAX_RECV_MSG_SIZE_MB", 4),
			MaxSendMsgSizeMB:  envInt("GRPC_MAX_SEND_MSG_SIZE_MB", 4),
			KeepaliveInterval: envDuration("GRPC_KEEPALIVE_INTERVAL", 2*time.Hour),
		},

		HTTP: HTTPConfig{
			Addr:         envOr("HTTP_ADDR", ":8080"),
			ReadTimeout:  envDuration("HTTP_READ_TIMEOUT", 10*time.Second),
			WriteTimeout: envDuration("HTTP_WRITE_TIMEOUT", 10*time.Second),
			IdleTimeout:  envDuration("HTTP_IDLE_TIMEOUT", 60*time.Second),
		},

		DB: DBConfig{
			URL:             os.Getenv("DATABASE_URL"),
			MaxConns:        int32(envInt("DATABASE_MAX_CONNS", 5)),
			MinConns:        int32(envInt("DATABASE_MIN_CONNS", 2)),
			MaxConnLifetime: envDuration("DATABASE_MAX_CONN_LIFETIME", time.Hour),
			MaxConnIdleTime: envDuration("DATABASE_MAX_CONN_IDLE_TIME", 30*time.Minute),
		},

		Redis: RedisConfig{
			Addr:     envOr("REDIS_ADDR", "localhost:6379"),
			Password: os.Getenv("REDIS_PASSWORD"),
			DB:       envInt("REDIS_DB", 0),
			PoolSize: envInt("REDIS_POOL_SIZE", 10),
		},

		Kafka: KafkaConfig{
			Brokers: strings.Split(envOr("KAFKA_BROKERS", "localhost:9092"), ","),
		},
	}
}

func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}

func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

func envDuration(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return fallback
}
