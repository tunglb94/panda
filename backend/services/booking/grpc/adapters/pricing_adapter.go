package adapters

import (
	"context"
	"strconv"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/fairride/booking/app"
	"github.com/fairride/pricing/grpc/pricingpb"
)

// PricingAdapter implements app.PricingClient using the Pricing gRPC client.
type PricingAdapter struct {
	client pricingpb.PricingServiceClient
}

func NewPricingAdapter(client pricingpb.PricingServiceClient) *PricingAdapter {
	return &PricingAdapter{client: client}
}

func (a *PricingAdapter) CalculateFinalFare(ctx context.Context, vehicleType string, distanceKM, durationMin float64) (*app.FareInfo, error) {
	var header metadata.MD
	resp, err := a.client.CalculateFinalFare(ctx, &pricingpb.CalculateFinalFareRequest{
		VehicleType:           vehicleType,
		ActualDistanceKm:      distanceKM,
		ActualDurationMinutes: durationMin,
	}, grpc.Header(&header))
	if err != nil {
		return nil, err
	}
	fare := resp.GetFare()
	if fare == nil {
		return &app.FareInfo{}, nil
	}
	info := &app.FareInfo{
		Total:        fare.GetTotal(),
		CurrencyCode: fare.GetCurrencyCode(),
	}
	applyCommissionHeader(info, header)
	return info, nil
}

// applyCommissionHeader reads the commission detail Pricing's gRPC handler
// attaches as response header metadata when V3 is active (see
// pricing/grpc/handler.go's setCommissionResponseHeader) — the only way to
// carry this data across the service boundary without a proto field, since
// no protoc/buf toolchain is available to add one. Leaves info's commission
// fields at their zero value (HasCommissionDetail=false) when absent, e.g.
// Pricing is running V2.
func applyCommissionHeader(info *app.FareInfo, md metadata.MD) {
	if md == nil {
		return
	}
	if vals := md.Get("x-pricing-v3"); len(vals) == 0 || vals[0] != "true" {
		return
	}
	info.HasCommissionDetail = true
	if vals := md.Get("x-commission-cents"); len(vals) > 0 {
		info.CommissionCents, _ = strconv.ParseInt(vals[0], 10, 64)
	}
	if vals := md.Get("x-driver-income-cents"); len(vals) > 0 {
		info.DriverIncomeCents, _ = strconv.ParseInt(vals[0], 10, 64)
	}
	if vals := md.Get("x-voucher-discount-cents"); len(vals) > 0 {
		info.VoucherDiscountCents, _ = strconv.ParseInt(vals[0], 10, 64)
	}
	if vals := md.Get("x-commission-rate"); len(vals) > 0 {
		info.CommissionRate, _ = strconv.ParseFloat(vals[0], 64)
	}
}
