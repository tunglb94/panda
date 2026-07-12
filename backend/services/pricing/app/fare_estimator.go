package app

import "github.com/fairride/pricing/domain/entity"

// FareEstimator is the shape grpc.Handler depends on — extracted so
// cmd/server/main.go can hand the handler either the pre-existing
// *FareCalculator (V2) or the new *VersionedFareCalculator (feature-flag
// dispatch between V2/V3) without grpc/handler.go's behaviour changing at
// all. Both concrete types already have this exact method shape, so this
// interface is a pure, additive extraction — no existing caller of either
// type needs to change (PHẦN 12: "Không xóa API cũ").
type FareEstimator interface {
	Estimate(vehicleType entity.VehicleType, distanceKM, durationMin float64) (*entity.FareBreakdown, error)
	CalculateFinal(vehicleType entity.VehicleType, distanceKM, durationMin float64) (*entity.FareBreakdown, error)
}

var (
	_ FareEstimator = (*FareCalculator)(nil)
	_ FareEstimator = (*VersionedFareCalculator)(nil)
)
