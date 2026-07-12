package simulation

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/fairride/ai_simulation/domain/entity"
	"github.com/fairride/ai_simulation/integration"
)

// ProcessDeliveryRequest runs one delivery (parcel) request end-to-end: a
// real Pricing Engine estimate (Delivery has no fare formula of its own in
// production — see integration.DeliveryAdapter's doc comment — so this
// reuses the same Pricing Engine call Ride uses, exactly mirroring the
// precedent already established by the Rider app's own DeliveryFormPage),
// the same Promotion/Voucher evaluation Ride uses, real Dispatch matching
// (TripType=delivery), and then the REAL production Delivery state machine
// (Created -> Accepted -> ParcelPickedUp -> InDelivery -> Delivered ->
// Completed) via integration.DeliveryAdapter — never a simulation-local
// reimplementation of that lifecycle.
//
// Unlike ProcessRideRequest, the TripID is not chosen by the caller: the
// real CreateTripUseCase generates it (see DeliveryAdapter.CreateDelivery),
// and that single ID is then used for every subsequent call (Dispatch,
// Accept/Pickup/Start/Complete) — the same ID a real Booking service would
// learn from CreateTrip before requesting dispatch.
func ProcessDeliveryRequest(ctx context.Context, w *World, rider *entity.RiderAgent) *entity.SimTrip {
	pickupZone := rider.Zone
	destZone := pickDestinationZone(w, rider)
	distanceKM := w.City.DistanceKM(pickupZone, destZone)
	if distanceKM < 0.8 {
		distanceKM = 0.8
	}
	durationMin := (distanceKM / averageCitySpeedKmh) * 60 / w.Traffic.SpeedFactor()

	weightKg := randomPackageWeightKg(w.Rand)
	declaredValue := int64(50_000 + w.Rand.Intn(2_000_000))
	serviceType := pickDeliveryServiceType(w.Rand)

	tripEntity, err := w.Delivery.CreateDelivery(ctx, integration.CreateDeliveryInput{
		RiderID: rider.ID, PickupAddress: string(pickupZone), DropoffAddress: string(destZone),
		SenderName: "Sender-" + rider.ID, SenderPhone: integration.FakeVNPhone(w.Rand),
		ReceiverName: fmt.Sprintf("Receiver-%d", w.Rand.Int63()), ReceiverPhone: integration.FakeVNPhone(w.Rand),
		PackageNote: "simulated parcel", PackageValue: declaredValue, WeightKg: weightKg,
	})
	if err != nil {
		// Booking-time failure (e.g. a validation edge case) — no real
		// TripID exists yet, so record it under a synthetic local one.
		trip := &entity.SimTrip{
			TripID:  fmt.Sprintf("delivery-failed-%d-%s", w.Clock.Tick, rider.ID),
			RiderID: rider.ID, Kind: entity.KindDelivery, Outcome: entity.OutcomeCancelled,
			RequestedAtTick: w.Clock.Tick, Weather: w.Weather, Traffic: w.Traffic,
			PackageWeightKg: weightKg,
		}
		w.RecordTrip(trip)
		return trip
	}
	tripID := tripEntity.TripID

	trip := &entity.SimTrip{
		TripID: tripID, RiderID: rider.ID, Kind: entity.KindDelivery,
		PickupZone: pickupZone, DestinationZone: destZone, DistanceKM: distanceKM,
		RequestedAtTick: w.Clock.Tick, ServiceType: serviceType,
		Weather: w.Weather, Traffic: w.Traffic,
		PackageWeightKg: weightKg,
		Outcome:         entity.OutcomePending,
	}

	quote, err := w.Pricing.Quote(integration.QuoteInput{
		ServiceType: serviceType, DistanceKM: distanceKM, DurationMin: durationMin,
		RequestTime:    w.Clock.Now(),
		ActiveRequests: w.ZoneActiveRequests[pickupZone], AvailableDrivers: w.ZoneAvailableDrivers[pickupZone],
		IsAirportZone: pickupZone == entity.ZoneAirport || destZone == entity.ZoneAirport,
		Weather:       w.Weather,
		IsHoliday:     w.Scenarios.Has(entity.ScenarioHoliday),
	})
	if err != nil {
		trip.Outcome = entity.OutcomeCancelled
		w.RecordTrip(trip)
		return trip
	}
	trip.BaseFareVND = quote.BaseFare + quote.DistanceFare + quote.TimeFare
	if quote.Total > 0 && trip.BaseFareVND > 0 {
		trip.SurgeMultiplier = float64(quote.RideFare) / float64(trip.BaseFareVND)
	} else {
		trip.SurgeMultiplier = 1.0
	}

	orderAmount := quote.Total
	promoResult := evaluatePromotions(ctx, w, rider, serviceType, orderAmount)
	if promoResult != nil && promoResult.Applied {
		trip.VoucherDiscountVND = promoResult.DiscountAmount
		trip.PromotionType = string(promoResult.Type)
		orderAmount = promoResult.FinalOrderAmount
	}
	trip.FinalFareVND = orderAmount

	job, err := w.Dispatch.RequestDelivery(ctx, tripID, rider.ID, w.City.Zones[pickupZone].X, w.City.Zones[pickupZone].Y, serviceType)
	if err != nil || job == nil {
		trip.Outcome = entity.OutcomeRejected
		w.RecordTrip(trip)
		return trip
	}

	assignedDriverID := resolveOffer(ctx, w, tripID, job)
	if assignedDriverID == "" {
		trip.Outcome = entity.OutcomeRejected
		w.RecordTrip(trip)
		return trip
	}

	driver := w.Drivers[assignedDriverID]
	trip.DriverID = assignedDriverID
	trip.AssignedAtTick = w.Clock.Tick
	if driver != nil {
		pickupDist := w.City.DistanceKM(driver.Zone, pickupZone)
		trip.PickupMinutes = (pickupDist / averageCitySpeedKmh) * 60 / w.Traffic.SpeedFactor()
		trip.ETAMinutes = trip.PickupMinutes + durationMin
		trip.DeliveryTransitMinutes = durationMin
		driver.CurrentTripID = tripID
	}

	if promoResult != nil && promoResult.Applied {
		_ = w.Promotion.Redeem(ctx, promoResult, rider.ID, tripID)
	}

	// Drive the REAL production Delivery state machine — every call below
	// is a real backend/services/trip use case, not simulation logic.
	if err := w.Delivery.AssignDriver(tripID, assignedDriverID); err != nil {
		trip.Outcome = entity.OutcomeCancelled
		w.RecordTrip(trip)
		return trip
	}
	if err := w.Delivery.MarkDriverArrived(ctx, tripID); err != nil {
		trip.Outcome = entity.OutcomeCancelled
		w.RecordTrip(trip)
		return trip
	}
	if _, err := w.Delivery.AcceptDelivery(ctx, tripID); err != nil {
		trip.Outcome = entity.OutcomeCancelled
		w.RecordTrip(trip)
		return trip
	}
	if _, err := w.Delivery.PickupParcel(ctx, tripID); err != nil {
		trip.Outcome = entity.OutcomeCancelled
		w.RecordTrip(trip)
		return trip
	}
	if _, err := w.Delivery.StartDelivery(ctx, tripID); err != nil {
		trip.Outcome = entity.OutcomeCancelled
		w.RecordTrip(trip)
		return trip
	}
	if _, err := w.Delivery.CompleteDelivery(ctx, tripID); err != nil {
		trip.Outcome = entity.OutcomeCancelled
		w.RecordTrip(trip)
		return trip
	}

	commissionVND, driverNetVND := int64(0), int64(0)
	if driver != nil {
		tier := w.Economy.TierForDriver(driver.TotalTrips, driver.Rating)
		driver.AccountType = tier
		// See ride_flow.go's identical block for why Split runs on the
		// already-scaled metered fare with the booking fee added unscaled.
		scaledMeteredFareVND := orderAmount - quote.BookingFee
		commissionVND, driverNetVND = w.Economy.Split(tier, scaledMeteredFareVND, quote.BookingFee)
		driver.IncomeToday += driverNetVND
		driver.IncomeWeek += driverNetVND
		driver.TotalTrips++
		driver.TripsThisRun++
		driver.Zone = destZone
		driver.CurrentTripID = ""
	}
	trip.CommissionVND = commissionVND
	trip.DriverNetVND = driverNetVND
	trip.CompletedAtTick = w.Clock.Tick
	trip.Outcome = entity.OutcomeCompleted

	rider.TripCount++
	rider.Zone = destZone

	w.RecordTrip(trip)
	return trip
}

// randomPackageWeightKg is a simulation-design assumption (0.5-15kg,
// roughly a document up to a small-appliance parcel) — no real package
// weight distribution exists to sample from.
func randomPackageWeightKg(rnd *rand.Rand) float64 {
	return 0.5 + rnd.Float64()*14.5
}

// pickDeliveryServiceType picks a Delivery service tier from the SAME
// 4-value ServiceType catalog Ride uses (Vehicle/Service Catalog refactor —
// there is no "delivery_bike"/"delivery_car" ServiceType; TripType=delivery
// is what distinguishes a Delivery request from a Ride one). Only
// Bike/Car — the two-wheeled/four-wheeled base tiers, not the Plus/XL
// premium variants — are realistic for parcel delivery, at the fixed 80/20
// ratio the sprint brief specifies.
func pickDeliveryServiceType(rnd *rand.Rand) entity.ServiceType {
	if rnd.Float64() < 0.80 {
		return entity.ServiceBike
	}
	return entity.ServiceCar
}
