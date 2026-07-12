// Package grpc contains the gRPC handler for the Pricing service.
package grpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/fairride/pricing/app"
	"github.com/fairride/pricing/domain/entity"
	"github.com/fairride/pricing/grpc/pricingpb"
	domainerrors "github.com/fairride/shared/errors"
)

// Handler implements pricingpb.PricingServiceServer.
type Handler struct {
	pricingpb.UnimplementedPricingServiceServer
	calc app.FareEstimator
}

// NewHandler creates a Handler from a FareEstimator — either a plain
// *app.FareCalculator (V2 only, e.g. in tests that don't care about the V3
// feature flag) or an *app.VersionedFareCalculator (production, dispatches
// to V2 or V3 per PRICING_VERSION — see cmd/server/main.go). Accepting the
// interface rather than the concrete *app.FareCalculator is the only change
// this sprint makes to this file — every existing caller passing
// *app.FareCalculator keeps compiling unchanged, since it already satisfies
// FareEstimator.
func NewHandler(calc app.FareEstimator) *Handler {
	if calc == nil {
		panic("pricing grpc: FareEstimator must not be nil")
	}
	return &Handler{calc: calc}
}

// EstimateFare implements PricingServiceServer.EstimateFare.
func (h *Handler) EstimateFare(ctx context.Context, req *pricingpb.EstimateFareRequest) (*pricingpb.FareResponse, error) {
	if req.GetVehicleType() == "" {
		return nil, status.Error(codes.InvalidArgument, "vehicle_type is required")
	}
	fb, err := h.calc.Estimate(entity.VehicleType(req.GetVehicleType()), req.GetDistanceKm(), req.GetDurationMinutes())
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &pricingpb.FareResponse{Fare: toProto(fb)}, nil
}

// CalculateFinalFare implements PricingServiceServer.CalculateFinalFare.
func (h *Handler) CalculateFinalFare(ctx context.Context, req *pricingpb.CalculateFinalFareRequest) (*pricingpb.FareResponse, error) {
	if req.GetVehicleType() == "" {
		return nil, status.Error(codes.InvalidArgument, "vehicle_type is required")
	}
	fb, err := h.calc.CalculateFinal(entity.VehicleType(req.GetVehicleType()), req.GetActualDistanceKm(), req.GetActualDurationMinutes())
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &pricingpb.FareResponse{Fare: toProto(fb)}, nil
}

// ─── private helpers ──────────────────────────────────────────────────────────

func toProto(fb *entity.FareBreakdown) *pricingpb.FareBreakdown {
	return &pricingpb.FareBreakdown{
		VehicleType:     string(fb.VehicleType),
		DistanceKm:      fb.DistanceKM,
		DurationMinutes: fb.DurationMin,
		BaseFare:        fb.BaseFare,
		DistanceFare:    fb.DistanceFare,
		TimeFare:        fb.TimeFare,
		BookingFee:      fb.BookingFee,
		RideFare:        fb.RideFare,
		Total:           fb.Total,
		CurrencyCode:    fb.CurrencyCode,
		IsFinal:         fb.IsFinal,
	}
}

func toGRPCError(err error) error {
	switch domainerrors.GetCode(err) {
	case domainerrors.CodeNotFound:
		return status.Error(codes.NotFound, err.Error())
	case domainerrors.CodeAlreadyExists:
		return status.Error(codes.AlreadyExists, err.Error())
	case domainerrors.CodeInvalidArgument:
		return status.Error(codes.InvalidArgument, err.Error())
	case domainerrors.CodePreconditionFailed:
		return status.Error(codes.FailedPrecondition, err.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
}
