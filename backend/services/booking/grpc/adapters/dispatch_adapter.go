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

func (a *DispatchAdapter) RequestDispatch(ctx context.Context, tripID, riderID string, pickupLat, pickupLon float64) error {
	_, err := a.client.RequestDispatch(ctx, &dispatchpb.RequestDispatchRequest{
		TripId:   tripID,
		RiderId:  riderID,
		PickupLat: pickupLat,
		PickupLon: pickupLon,
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
