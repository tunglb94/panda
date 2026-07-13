package handlers

import (
	"context"
	"strconv"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/fairride/booking/grpc/bookingpb"
	"github.com/fairride/trip/grpc/trippb"
	walletapp "github.com/fairride/wallet/app"
	walletentity "github.com/fairride/wallet/domain/entity"
)

// SettlementEngine is the additive gateway-layer hook that turns a paid
// trip into Settlement + Ledger entries (Settlement Engine spec, Phần 2) —
// the exact same pattern as TripEventNotifier (this file's sibling): the
// gateway calls it AFTER BookingHandler.PayRide's existing gRPC call
// already succeeded, so Trip/Booking's own lifecycle code is never touched.
//
// Settlement fires at payment confirmation (PayRide success), not at trip
// completion (FinishTrip) — PaymentMethod (needed to pick Settlement Case
// 1 vs 2) is only known once the rider/driver confirms how the fare was
// collected, which in this codebase happens at PayRide, not FinishTrip. See
// the module report's "Financial Flow" section for the full reasoning.
type SettlementEngine struct {
	bookings BookingClient
	trips    TripStatusClient
	create   *walletapp.CreateSettlementUseCase
}

func NewSettlementEngine(bookings BookingClient, trips TripStatusClient, create *walletapp.CreateSettlementUseCase) *SettlementEngine {
	return &SettlementEngine{bookings: bookings, trips: trips, create: create}
}

func (s *SettlementEngine) configured() bool {
	return s != nil && s.bookings != nil && s.create != nil
}

// SettleResult carries a best-effort outcome back to the HTTP handler —
// never blocks or fails the PayRide response (the payment itself already
// succeeded in Trip/Booking, which remains the source of truth for trip
// state); a non-empty Error is surfaced as an extra, non-fatal JSON field
// so a failed settlement is at least visible rather than silently lost.
type SettleResult struct {
	Settled bool
	Error   string
}

// Settle resolves driver_id and trip_type via the same read-only,
// already-established enrichment calls the rest of this package uses
// (GetBookingDetails for driver_id, NewTripReader for trip_type — see
// chat_handler.go's tripReaderAdapter), then posts the trip's Settlement.
func (s *SettlementEngine) Settle(ctx context.Context, tripID string, fareCents int64, currency, paymentMethod string) SettleResult {
	if !s.configured() || fareCents <= 0 {
		return SettleResult{}
	}
	details, err := s.bookings.GetBookingDetails(ctx, &bookingpb.GetBookingDetailsRequest{TripId: tripID})
	if err != nil || details.GetDriverId() == "" {
		return SettleResult{Error: "could not resolve driver for settlement"}
	}
	tripType := walletentity.TripTypeRide
	method := walletentity.PaymentMethodCash
	if paymentMethod == string(walletentity.PaymentMethodWallet) {
		method = walletentity.PaymentMethodWallet
	}
	var fin financialsFromTrip
	if s.trips != nil {
		if snapshot, err := (tripReaderAdapter{client: s.trips}).GetTrip(ctx, tripID); err == nil && snapshot.TripType == "delivery" {
			tripType = walletentity.TripTypeDelivery
		}
		// Read Trip's OWN persisted PaymentMethod/commission detail rather
		// than trusting only the ephemeral HTTP-request paymentMethod param
		// (critique #3: Settlement should read Trip/Booking, not the
		// request body). paymentMethod is kept as the fallback for a Trip
		// row that predates this fix (has_commission_detail=false and no
		// payment_method persisted yet).
		fin = fetchTripFinancials(ctx, s.trips, tripID)
		if fin.paymentMethod == string(walletentity.PaymentMethodWallet) {
			method = walletentity.PaymentMethodWallet
		} else if fin.paymentMethod == string(walletentity.PaymentMethodCash) {
			method = walletentity.PaymentMethodCash
		}
	}
	_, err = s.create.Execute(ctx, walletapp.CreateSettlementInput{
		TripID: tripID, DriverID: details.GetDriverId(), TripType: tripType,
		PaymentMethod: method, FareAmountCents: fareCents, Currency: currency,
		HasCommissionDetail:  fin.hasCommissionDetail,
		CommissionCents:      fin.commissionCents,
		DriverIncomeCents:    fin.driverIncomeCents,
		CommissionRate:       fin.commissionRate,
		VoucherDiscountCents: fin.voucherDiscountCents,
	})
	if err != nil {
		return SettleResult{Error: err.Error()}
	}
	return SettleResult{Settled: true}
}

// financialsFromTrip is what fetchTripFinancials reads off Trip's response
// header metadata (see trip/grpc/handler.go's setTripFinancialsHeader) — the
// durable source Settlement must use instead of inventing its own commission
// number or trusting only an ephemeral HTTP request field.
type financialsFromTrip struct {
	paymentMethod        string
	hasCommissionDetail  bool
	commissionCents      int64
	driverIncomeCents    int64
	commissionRate       float64
	voucherDiscountCents int64
}

func fetchTripFinancials(ctx context.Context, client TripStatusClient, tripID string) financialsFromTrip {
	var header metadata.MD
	_, err := client.GetTrip(ctx, &trippb.GetTripRequest{TripId: tripID}, grpc.Header(&header))
	if err != nil || header == nil {
		return financialsFromTrip{}
	}
	var fin financialsFromTrip
	if vals := header.Get("x-payment-method"); len(vals) > 0 {
		fin.paymentMethod = vals[0]
	}
	if vals := header.Get("x-has-commission-detail"); len(vals) > 0 && vals[0] == "true" {
		fin.hasCommissionDetail = true
		if v := header.Get("x-commission-cents"); len(v) > 0 {
			fin.commissionCents, _ = strconv.ParseInt(v[0], 10, 64)
		}
		if v := header.Get("x-driver-income-cents"); len(v) > 0 {
			fin.driverIncomeCents, _ = strconv.ParseInt(v[0], 10, 64)
		}
		if v := header.Get("x-commission-rate"); len(v) > 0 {
			fin.commissionRate, _ = strconv.ParseFloat(v[0], 64)
		}
		if v := header.Get("x-voucher-discount-cents"); len(v) > 0 {
			fin.voucherDiscountCents, _ = strconv.ParseInt(v[0], 10, 64)
		}
	}
	return fin
}
