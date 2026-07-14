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
// This service has no gRPC handler and is not meant to run as a standalone
// process in production — no .proto/generated *pb code exists for it (this
// environment has no protoc/buf toolchain), so gateway/cmd/server/main.go
// imports these same domain/app/infrastructure packages directly and
// in-process instead, sharing the gateway's own Postgres pool — the exact
// pattern already used there for Identity/User/Wallet. This main() and its
// standalone binary exist only so `go build ./...`/`go vet ./...` cover the
// module the same way every other service's does; it is never actually
// deployed/run on its own.
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
