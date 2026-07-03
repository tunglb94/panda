package main

import (
	sharedconfig "github.com/fairride/shared/config"
	sharedgrpc "github.com/fairride/shared/grpc"
	"github.com/fairride/shared/server"

	"github.com/fairride/pricing/app"
	"github.com/fairride/pricing/domain/entity"
	pricinggrpc "github.com/fairride/pricing/grpc"
	"github.com/fairride/pricing/grpc/pricingpb"
)

func main() {
	server.Run("pricing", register)
}

func register(srv *sharedgrpc.Server, ready *server.ReadinessTracker) {
	cfg := sharedconfig.Load("pricing")
	_ = cfg // pricing is a pure compute service; no DB or Redis required

	calc := app.NewFareCalculator(entity.DefaultFareConfig())
	handler := pricinggrpc.NewHandler(calc)
	pricingpb.RegisterPricingServiceServer(srv.Inner(), handler)
}
