package main

import (
	"context"

	"github.com/fairride/promotion/app"
	"github.com/fairride/promotion/infrastructure/postgres"
	sharedconfig "github.com/fairride/shared/config"
	sharedgrpc "github.com/fairride/shared/grpc"
	"github.com/fairride/shared/server"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	server.Run("promotion", register)
}

// register wires the Promotion Engine's domain/app/infrastructure layers
// (Postgres repository -> VoucherValidator + RuleRegistry -> PromotionService).
//
// TODO(promotion-engine): this service does not yet expose a gRPC handler —
// no .proto/generated *pb code was written in this sprint (see CHANGELOG /
// report: the sprint's explicit required components — PromotionService,
// VoucherValidator, PromotionRepository, PromotionRule, PromotionResult,
// PromotionError — are all domain/app-layer concepts, and this environment
// has no Go toolchain available to safely hand-verify generated protobuf
// code). Once a promotion.proto + grpc handler exist, register them here on
// srv.Inner(), following the exact pattern used by
// backend/services/driver/cmd/server/main.go.
func register(_ *sharedgrpc.Server, ready *server.ReadinessTracker) {
	cfg := sharedconfig.Load("promotion")

	ready.Set("db", false)
	pool, err := pgxpool.New(context.Background(), cfg.DB.URL)
	if err == nil {
		err = pool.Ping(context.Background())
	}
	ready.Set("db", err == nil)

	voucherRepo := postgres.NewVoucherRepository(pool)
	validator := app.NewVoucherValidator()
	rules := app.NewDefaultRuleRegistry()

	// promotionService is constructed and ready to be handed to a future gRPC
	// handler; it is intentionally unused by this main() until that handler
	// exists.
	_ = app.NewPromotionService(voucherRepo, validator, rules)
}
