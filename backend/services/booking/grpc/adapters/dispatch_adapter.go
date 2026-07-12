package adapters

import (
	"context"

	"github.com/fairride/booking/app"
	"github.com/fairride/dispatch/grpc/dispatchpb"
)

// DispatchAdapter implements app.DispatchClient using the Dispatch gRPC client.
type DispatchAdapter struct {
	client dispatchpb.DispatchServiceClient
}

func NewDispatchAdapter(client dispatchpb.DispatchServiceClient) *DispatchAdapter {
	return &DispatchAdapter{client: client}
}

func (a *DispatchAdapter) RequestDispatch(ctx context.Context, tripID, riderID, tripType, serviceType string, pickupLat, pickupLon float64) error {
	_, err := a.client.RequestDispatch(ctx, &dispatchpb.RequestDispatchRequest{
		TripId:    tripID,
		RiderId:   riderID,
		PickupLat: pickupLat,
		PickupLon: pickupLon,
		TripType:  tripType,
		// The wire field is still named "vehicle_type" (added during an
		// earlier Delivery phase) — its VALUE now carries a ServiceType.
		// See dispatch/grpc/handler.go's matching comment.
		VehicleType: serviceType,
	})
	return err
}

func (a *DispatchAdapter) AcceptTrip(ctx context.Context, tripID, driverID string) error {
	_, err := a.client.AcceptTrip(ctx, &dispatchpb.AcceptTripRequest{
		TripId:   tripID,
		DriverId: driverID,
	})
	return err
}

func (a *DispatchAdapter) RejectTrip(ctx context.Context, tripID, driverID string) error {
	_, err := a.client.RejectTrip(ctx, &dispatchpb.RejectTripRequest{
		TripId:   tripID,
		DriverId: driverID,
	})
	return err
}

func (a *DispatchAdapter) GetDispatchStatus(ctx context.Context, tripID string) (*app.DispatchInfo, error) {
	resp, err := a.client.GetDispatchStatus(ctx, &dispatchpb.GetDispatchStatusRequest{TripId: tripID})
	if err != nil {
		return nil, err
	}
	job := resp.GetJob()
	if job == nil {
		return nil, nil
	}
	return &app.DispatchInfo{
		TripID:           job.GetTripId(),
		Status:           job.GetStatus(),
		AssignedDriverID: job.GetAssignedDriverId(),
	}, nil
}

func (a *DispatchAdapter) GetDriverOffer(ctx context.Context, driverID string) (*app.DriverOfferInfo, error) {
	resp, err := a.client.GetDriverOffer(ctx, &dispatchpb.GetDriverOfferRequest{DriverId: driverID})
	if err != nil {
		return nil, err
	}
	if !resp.GetHasOffer() {
		return nil, nil
	}
	info := &app.DriverOfferInfo{
		TripID: resp.GetTripId(),
	}
	if ts := resp.GetOfferExpiresAt(); ts != nil {
		info.OfferExpiresAt = ts.AsTime()
	}
	return info, nil
}
