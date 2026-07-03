package adapters

import (
	"context"

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
	resp, err := a.client.CalculateFinalFare(ctx, &pricingpb.CalculateFinalFareRequest{
		VehicleType:           vehicleType,
		ActualDistanceKm:      distanceKM,
		ActualDurationMinutes: durationMin,
	})
	if err != nil {
		return nil, err
	}
	fare := resp.GetFare()
	if fare == nil {
		return &app.FareInfo{}, nil
	}
	return &app.FareInfo{
		Total:        fare.GetTotal(),
		CurrencyCode: fare.GetCurrencyCode(),
	}, nil
}
