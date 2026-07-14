// Package adapters contains concrete gRPC client adapters that implement the
// booking/app client interfaces. Each adapter wraps a generated gRPC client stub.
package adapters

import (
	"context"
	"strconv"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

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

func (a *TripAdapter) CreateTrip(ctx context.Context, in app.CreateTripParams) (*app.CreateTripResult, error) {
	resp, err := a.client.CreateTrip(ctx, &trippb.CreateTripRequest{
		RiderId:            in.RiderID,
		PickupAddress:      in.PickupAddress,
		DropoffAddress:     in.DropoffAddress,
		TripType:           in.TripType,
		PickupContactName:  in.PickupContactName,
		PickupContactPhone: in.PickupContactPhone,
		ReceiverName:       in.ReceiverName,
		ReceiverPhone:      in.ReceiverPhone,
		PackageNote:        in.PackageNote,
		PackageValue:       in.PackageValue,
		PackageWeight:      in.PackageWeightKg,
	})
	if err != nil {
		return nil, err
	}
	return &app.CreateTripResult{
		TripID:     resp.GetTrip().GetTripId(),
		DeliveryID: resp.GetTrip().GetDeliveryId(),
	}, nil
}

func (a *TripAdapter) MarkDriverArrived(ctx context.Context, tripID string) error {
	_, err := a.client.MarkDriverArrived(ctx, &trippb.GetTripRequest{TripId: tripID})
	return err
}

func (a *TripAdapter) StartTrip(ctx context.Context, tripID string) error {
	_, err := a.client.StartTrip(ctx, &trippb.StartTripRequest{TripId: tripID})
	return err
}

// CompleteTrip persists fare on Trip and, when fare carries commission
// detail from Pricing V3, forwards it as OUTGOING gRPC metadata so Trip's
// handler can persist it too — trippb.CompleteTripRequest has no field slots
// for it and no protoc/buf toolchain is available to add one (same
// constraint as the "x-service-type" precedent in booking/grpc/handler.go).
func (a *TripAdapter) CompleteTrip(ctx context.Context, tripID string, fare app.FareInfo) (*app.TripInfo, error) {
	if fare.HasCommissionDetail {
		ctx = metadata.AppendToOutgoingContext(ctx,
			"x-has-commission-detail", "true",
			"x-commission-cents", strconv.FormatInt(fare.CommissionCents, 10),
			"x-driver-income-cents", strconv.FormatInt(fare.DriverIncomeCents, 10),
			"x-commission-rate", strconv.FormatFloat(fare.CommissionRate, 'f', -1, 64),
		)
	}
	// Voucher detail — own gate, independent of commission detail (a trip
	// can have a voucher with or without Pricing V3 commission detail).
	if fare.VoucherID != "" {
		ctx = metadata.AppendToOutgoingContext(ctx,
			"x-has-voucher-detail", "true",
			"x-voucher-id", fare.VoucherID,
			"x-voucher-code", fare.VoucherCode,
			"x-voucher-discount-cents", strconv.FormatInt(fare.VoucherDiscountCents, 10),
		)
	}
	// Trip Summary (Ride Lifecycle Fare Validation) — always forwarded
	// (unlike commission/voucher detail, every completed trip has a summary).
	ctx = metadata.AppendToOutgoingContext(ctx,
		"x-travelled-distance-km", strconv.FormatFloat(fare.TravelledDistanceKm, 'f', -1, 64),
		"x-travelled-duration-min", strconv.FormatFloat(fare.TravelledDurationMin, 'f', -1, 64),
		"x-toll-fee-cents", strconv.FormatInt(fare.TollFeeCents, 10),
		"x-extra-fee-cents", strconv.FormatInt(fare.ExtraFeeCents, 10),
	)
	resp, err := a.client.CompleteTrip(ctx, &trippb.CompleteTripRequest{
		TripId:         tripID,
		FinalFareTotal: fare.Total,
		FareCurrency:   fare.CurrencyCode,
	})
	if err != nil {
		return nil, err
	}
	return protoToTripInfo(resp.GetTrip()), nil
}

// GetTrip reads the commission detail and real payment method back via
// INCOMING response header metadata (see trip/grpc/handler.go's
// setFinancialsResponseHeader) — TripProto has no payment_method or
// commission fields, same proto-extension constraint as above.
func (a *TripAdapter) GetTrip(ctx context.Context, tripID string) (*app.TripInfo, error) {
	var header metadata.MD
	resp, err := a.client.GetTrip(ctx, &trippb.GetTripRequest{TripId: tripID}, grpc.Header(&header))
	if err != nil {
		return nil, err
	}
	info := protoToTripInfo(resp.GetTrip())
	applyTripFinancialsHeader(info, header)
	return info, nil
}

// applyTripFinancialsHeader reads PaymentMethod/commission detail Trip's
// gRPC handler attaches as response header metadata. No-ops (leaves info's
// zero values) when absent or info is nil.
func applyTripFinancialsHeader(info *app.TripInfo, md metadata.MD) {
	if info == nil || md == nil {
		return
	}
	if vals := md.Get("x-payment-method"); len(vals) > 0 {
		info.PaymentMethod = vals[0]
	}
	if vals := md.Get("x-has-commission-detail"); len(vals) > 0 && vals[0] == "true" {
		info.HasCommissionDetail = true
	}
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

func (a *TripAdapter) AcceptDelivery(ctx context.Context, tripID string) error {
	_, err := a.client.AcceptDelivery(ctx, &trippb.GetTripRequest{TripId: tripID})
	return err
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
