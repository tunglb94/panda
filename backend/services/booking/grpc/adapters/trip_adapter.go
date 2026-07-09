// Package adapters contains concrete gRPC client adapters that implement the
// booking/app client interfaces. Each adapter wraps a generated gRPC client stub.
package adapters

import (
	"context"
	"time"

	"github.com/fairride/booking/app"
	"github.com/fairride/trip/grpc/trippb"
)

// TripAdapter implements app.TripClient using the Trip gRPC client.
type TripAdapter struct {
	client trippb.TripServiceClient
}

func NewTripAdapter(client trippb.TripServiceClient) *TripAdapter {
	return &TripAdapter{client: client}
}

func (a *TripAdapter) CreateTrip(ctx context.Context, riderID, pickup, dropoff string) (string, error) {
	resp, err := a.client.CreateTrip(ctx, &trippb.CreateTripRequest{
		RiderId:        riderID,
		PickupAddress:  pickup,
		DropoffAddress: dropoff,
	})
	if err != nil {
		return "", err
	}
	return resp.GetTrip().GetTripId(), nil
}

func (a *TripAdapter) MarkDriverArrived(ctx context.Context, tripID string) error {
	_, err := a.client.MarkDriverArrived(ctx, &trippb.GetTripRequest{TripId: tripID})
	return err
}

func (a *TripAdapter) StartTrip(ctx context.Context, tripID string) error {
	_, err := a.client.StartTrip(ctx, &trippb.StartTripRequest{TripId: tripID})
	return err
}

func (a *TripAdapter) CompleteTrip(ctx context.Context, tripID string, finalFare int64, currency string) (*app.TripInfo, error) {
	resp, err := a.client.CompleteTrip(ctx, &trippb.CompleteTripRequest{
		TripId:         tripID,
		FinalFareTotal: finalFare,
		FareCurrency:   currency,
	})
	if err != nil {
		return nil, err
	}
	return protoToTripInfo(resp.GetTrip()), nil
}

func (a *TripAdapter) GetTrip(ctx context.Context, tripID string) (*app.TripInfo, error) {
	resp, err := a.client.GetTrip(ctx, &trippb.GetTripRequest{TripId: tripID})
	if err != nil {
		return nil, err
	}
	return protoToTripInfo(resp.GetTrip()), nil
}

func (a *TripAdapter) CancelTrip(ctx context.Context, tripID, reason string) error {
	_, err := a.client.CancelTrip(ctx, &trippb.CancelTripRequest{TripId: tripID, Reason: reason})
	return err
}

func (a *TripAdapter) InitiatePayment(ctx context.Context, tripID string) error {
	_, err := a.client.InitiatePayment(ctx, &trippb.GetTripRequest{TripId: tripID})
	return err
}

func (a *TripAdapter) PayTrip(ctx context.Context, tripID, paymentMethod string) (*app.TripInfo, error) {
	resp, err := a.client.PayTrip(ctx, &trippb.CancelTripRequest{TripId: tripID, Reason: paymentMethod})
	if err != nil {
		return nil, err
	}
	return protoToTripInfo(resp.GetTrip()), nil
}

func (a *TripAdapter) ListByRider(ctx context.Context, riderID string) ([]app.TripSummary, error) {
	resp, err := a.client.ListTripsByRider(ctx, &trippb.ListTripsRequest{PartyId: riderID})
	if err != nil {
		return nil, err
	}
	return protoToSummaries(resp.GetTrips()), nil
}

func (a *TripAdapter) ListByDriver(ctx context.Context, driverID string) ([]app.TripSummary, error) {
	resp, err := a.client.ListTripsByDriver(ctx, &trippb.ListTripsRequest{PartyId: driverID})
	if err != nil {
		return nil, err
	}
	return protoToSummaries(resp.GetTrips()), nil
}

func protoToSummaries(trips []*trippb.TripProto) []app.TripSummary {
	out := make([]app.TripSummary, len(trips))
	for i, t := range trips {
		var createdAt time.Time
		if ts := t.GetCreatedAt(); ts != nil {
			createdAt = ts.AsTime()
		}
		out[i] = app.TripSummary{
			TripID:         t.GetTripId(),
			Status:         t.GetStatus(),
			PickupAddress:  t.GetPickupAddress(),
			DropoffAddress: t.GetDropoffAddress(),
			FinalFare:      t.GetFinalFareTotal(),
			Currency:       t.GetFareCurrency(),
			CreatedAt:      createdAt,
		}
	}
	return out
}

func protoToTripInfo(t *trippb.TripProto) *app.TripInfo {
	if t == nil {
		return nil
	}
	return &app.TripInfo{
		TripID:             t.GetTripId(),
		RiderID:            t.GetRiderId(),
		DriverID:           t.GetDriverId(),
		Status:             t.GetStatus(),
		PickupAddress:      t.GetPickupAddress(),
		DropoffAddress:     t.GetDropoffAddress(),
		CancellationReason: t.GetCancellationReason(),
		FinalFareTotal:     t.GetFinalFareTotal(),
		FareCurrency:       t.GetFareCurrency(),
	}
}
