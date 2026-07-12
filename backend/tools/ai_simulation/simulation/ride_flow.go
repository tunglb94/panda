package simulation

import (
	"context"
	"math/rand"

	dispatchentity "github.com/fairride/dispatch/domain/entity"
	promotionentity "github.com/fairride/promotion/domain/entity"

	"github.com/fairride/ai_simulation/aiengine"
	"github.com/fairride/ai_simulation/domain/entity"
	"github.com/fairride/ai_simulation/integration"
	"github.com/fairride/ai_simulation/ruleengine"
)

const averageCitySpeedKmh = 28.0 // baseline before Traffic.SpeedFactor() scales it

// ProcessRideRequest runs one ride request end-to-end: destination/vehicle
// selection, a real Pricing Engine quote, a real Promotion Engine
// evaluation (auto-apply, then a held voucher if the rider decides to spend
// it), a rider stay-or-switch-app decision, real Dispatch matching, a
// driver accept/reject rule, and — on success — Driver Economy commission
// split. Every statistic later exported comes from the *entity.SimTrip this
// function returns.
func ProcessRideRequest(ctx context.Context, w *World, rider *entity.RiderAgent, tripID string) *entity.SimTrip {
	pickupZone := rider.Zone
	destZone := pickDestinationZone(w, rider)
	distanceKM := w.City.DistanceKM(pickupZone, destZone)
	if distanceKM < 0.8 {
		distanceKM = 0.8
	}
	durationMin := (distanceKM / averageCitySpeedKmh) * 60 / w.Traffic.SpeedFactor()

	serviceType := pickServiceType(w.Rand, rider)

	trip := &entity.SimTrip{
		TripID: tripID, RiderID: rider.ID, Kind: entity.KindRide,
		PickupZone: pickupZone, DestinationZone: destZone, DistanceKM: distanceKM,
		RequestedAtTick: w.Clock.Tick, ServiceType: serviceType,
		Weather: w.Weather, Traffic: w.Traffic,
		Outcome: entity.OutcomePending,
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

	// Rider decision: stay on Panda or switch to a competitor app.
	competitorFare := simulateCompetitorFare(w.Rand, orderAmount)
	switchOutcome := decideSwitchApp(ctx, w, rider, orderAmount, competitorFare)
	if switchOutcome.Decision == ruleengine.DecisionSwitch {
		trip.Outcome = entity.OutcomeCancelled
		w.RecordTrip(trip)
		return trip
	}

	job, err := w.Dispatch.RequestRide(ctx, tripID, rider.ID, w.City.Zones[pickupZone].X, w.City.Zones[pickupZone].Y, serviceType)
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
		driver.CurrentTripID = tripID
	}

	// Commit the promotion redemption now that the trip is truly happening
	// (mirrors the real Estimate-then-Redeem split — see
	// backend/services/promotion's PromotionService doc comment).
	if promoResult != nil && promoResult.Applied {
		_ = w.Promotion.Redeem(ctx, promoResult, rider.ID, tripID)
	}

	commissionVND, driverNetVND := int64(0), int64(0)
	if driver != nil {
		tier := w.Economy.TierForDriver(driver.TotalTrips, driver.Rating)
		driver.AccountType = tier
		// Split on the metered fare actually charged (orderAmount, net of
		// the booking fee) — orderAmount already reflects surge and any
		// voucher/promotion discount. The booking fee is passed to Split
		// separately and added once, unscaled (BRB §2.2.5: a flat fee, not
		// surge-sensitive) — commissionVND+driverNetVND == orderAmount by
		// construction, matching Split's own documented invariant.
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

// evaluatePromotions checks auto-apply campaigns first (First Ride/Birthday/
// Weekend — no rider decision involved, BRB applies them automatically when
// eligible); only if none apply AND the rider is holding a manual coupon
// does the VoucherUseDecision rule/AI decision get invoked at all — an
// idle rider with no voucher never triggers an AI call.
func evaluatePromotions(ctx context.Context, w *World, rider *entity.RiderAgent, serviceType entity.ServiceType, orderAmount int64) *promotionentity.PromotionResult {
	autoResult, err := w.Promotion.Evaluate(ctx, integration.EvaluateInput{
		RiderID: rider.ID, ServiceType: serviceType, City: "default",
		OrderAmountVND: orderAmount, RequestTime: w.Clock.Now(),
		IsNewRider: rider.TripCount == 0, CompletedTripsTotal: int64(rider.TripCount),
		BirthdayToday: false, MembershipTier: string(rider.Membership),
	})
	if err == nil && autoResult.Applied {
		return autoResult
	}

	if !rider.HasActiveVoucher {
		return nil
	}

	useOutcome := decideVoucherUse(ctx, w, rider, orderAmount)
	w.recordDecision(useOutcome.NeedsAI)
	w.recordVoucherOutcome(useOutcome.Decision == ruleengine.DecisionUseVoucher)
	if useOutcome.Decision != ruleengine.DecisionUseVoucher {
		return nil
	}

	couponResult, err := w.Promotion.Evaluate(ctx, integration.EvaluateInput{
		RiderID: rider.ID, ServiceType: serviceType, City: "default",
		OrderAmountVND: orderAmount, RequestTime: w.Clock.Now(),
		VoucherCode: "SIM10", MembershipTier: string(rider.Membership),
	})
	if err != nil || !couponResult.Applied {
		return nil
	}
	rider.HasActiveVoucher = false
	return couponResult
}

func pickDestinationZone(w *World, rider *entity.RiderAgent) entity.ZoneType {
	weights := map[entity.ZoneType]float64{}
	hour := w.Clock.Hour()
	for _, z := range entity.AllZoneTypes() {
		weights[z] = w.City.Zones[z].BaseDemandWeight
	}
	switch rider.Habit {
	case entity.HabitCommuter:
		if hour == rider.WorkStartHour {
			weights[entity.ZoneCBD] *= 4
			weights[entity.ZoneIndustrial] *= 3
		} else if hour == rider.WorkEndHour {
			weights[entity.ZoneResidential] *= 4
		}
	case entity.HabitNightOwl:
		if hour >= 19 || hour < 2 {
			weights[entity.ZoneEntertainment] *= 4
		}
	case entity.HabitBusinessTraveler:
		weights[entity.ZoneAirport] *= 3
	}
	weights[rider.Zone] = 0 // never "travel" to the same zone you're already in
	return pickWeightedZone(w.Rand, weights)
}

func pickWeightedZone(rnd *rand.Rand, weights map[entity.ZoneType]float64) entity.ZoneType {
	var total float64
	for _, wgt := range weights {
		total += wgt
	}
	if total <= 0 {
		return entity.ZoneCBD
	}
	r := rnd.Float64() * total
	for _, z := range entity.AllZoneTypes() {
		r -= weights[z]
		if r <= 0 {
			return z
		}
	}
	return entity.ZoneCBD
}

// pickServiceType picks a Ride service tier in two steps: which family
// (Bike vs Car) the rider leans toward — income-tiered, a simulation-design
// assumption not sourced from any BRB number, same as every other weight in
// this function — then which tier within that family, at the fixed ratios
// the sprint brief specifies (Bike 65% / Bike Plus 35%, Car 75% / Car XL
// 25%) so aggregate statistics reflect the intended product mix from the
// first run, not a slow warm-up.
func pickServiceType(rnd *rand.Rand, rider *entity.RiderAgent) entity.ServiceType {
	var bikeFamily bool
	switch {
	case rider.Income < 8_000_000: // VND/month — lower income leans bike
		bikeFamily = rnd.Float64() < 0.75
	case rider.Income > 30_000_000: // higher income leans car
		bikeFamily = rnd.Float64() < 0.20
	default:
		bikeFamily = rnd.Float64() < 0.40
	}

	if bikeFamily {
		if rnd.Float64() < 0.65 {
			return entity.ServiceBike
		}
		return entity.ServiceBikePlus
	}
	if rnd.Float64() < 0.75 {
		return entity.ServiceCar
	}
	return entity.ServiceCarXL
}

// simulateCompetitorFare is a synthetic reference price standing in for a
// real-time competitor quote (Grab/Be/etc) — no such data source exists,
// see docs/business/PRICING_STRATEGY.md's competitive analysis, which
// itself only uses market-range estimates. Centered on Panda's own price so
// the two are usually close, occasionally meaningfully apart in either
// direction — the exact shape the sprint brief's worked example describes.
func simulateCompetitorFare(rnd *rand.Rand, pandaFareVND int64) int64 {
	factor := 0.82 + rnd.Float64()*0.42 // 0.82x - 1.24x
	return int64(float64(pandaFareVND) * factor)
}

func decideSwitchApp(ctx context.Context, w *World, rider *entity.RiderAgent, pandaFare, competitorFare int64) ruleengine.Outcome {
	outcome := ruleengine.SwitchAppDecision(ruleengine.SwitchAppInput{
		PandaFareVND: pandaFare, CompetitorFareVND: competitorFare,
		PriceSensitivity: rider.PriceSensitivity, Patience: rider.Patience,
		Membership: string(rider.Membership),
	})
	if !outcome.NeedsAI {
		w.recordDecision(false)
		w.recordSwitchAppOutcome(outcome.Decision == ruleengine.DecisionSwitch)
		return outcome
	}
	prompt := aiengine.SwitchAppPrompt(pandaFare, competitorFare, aiengine.BucketLevel3(rider.PriceSensitivity), string(rider.Membership))
	result := w.AI.Decide(ctx, aiengine.DecisionRequest{
		Prompt:   prompt,
		Parse:    aiengine.ParseBinary("SWITCH", ruleengine.DecisionSwitch, "STAY", ruleengine.DecisionStay),
		Fallback: outcome.Decision,
	})
	w.recordDecision(true)
	outcome.Decision = result.Decision
	w.recordSwitchAppOutcome(outcome.Decision == ruleengine.DecisionSwitch)
	return outcome
}

func decideVoucherUse(ctx context.Context, w *World, rider *entity.RiderAgent, orderAmount int64) ruleengine.Outcome {
	outcome := ruleengine.VoucherUseDecision(ruleengine.VoucherUseInput{
		DiscountPercent: rider.VoucherDiscountPercent, PriceSensitivity: rider.PriceSensitivity,
		TripCount: rider.TripCount, OrderAmountVND: orderAmount,
	})
	if !outcome.NeedsAI {
		return outcome
	}
	prompt := aiengine.VoucherUsePrompt(rider.VoucherDiscountPercent, aiengine.BucketLevel3(rider.PriceSensitivity), aiengine.BucketLevel3(1-float64(rider.TripCount)/50))
	result := w.AI.Decide(ctx, aiengine.DecisionRequest{
		Prompt:   prompt,
		Parse:    aiengine.ParseBinary("USE", ruleengine.DecisionUseVoucher, "KEEP", ruleengine.DecisionKeepVoucher),
		Fallback: outcome.Decision,
	})
	outcome.Decision = result.Decision
	return outcome
}

// resolveOffer walks the dispatch job's offer chain: for the currently
// offered driver, run the (rule-only — not one of the sprint brief's 4 AI
// decision points) accept/reject check; on reject, RejectTripUseCase
// (called by DispatchAdapter.Reject) auto-advances to the next nearest
// driver and returns the updated job, so this loop simply keeps asking
// "does the currently offered driver accept?" until one does or the job
// runs out of candidates (Status transitions to Failed, CurrentDriverID
// goes empty).
func resolveOffer(ctx context.Context, w *World, tripID string, job *dispatchentity.DispatchJob) string {
	for attempts := 0; attempts < dispatchentity.DefaultMaxAttempts+1 && job != nil; attempts++ {
		if job.Status == dispatchentity.JobStatusAssigned {
			driverID, ok := w.Dispatch.AssignedDriver(tripID)
			if ok {
				return driverID
			}
			return job.AssignedDriverID
		}
		if job.CurrentDriverID == "" {
			return "" // job failed — no eligible driver left
		}

		offeredDriverID := job.CurrentDriverID
		var err error
		if driverAcceptsOffer(w, offeredDriverID) {
			job, err = w.Dispatch.Accept(ctx, tripID, offeredDriverID)
		} else {
			job, err = w.Dispatch.Reject(ctx, tripID, offeredDriverID)
		}
		if err != nil {
			return ""
		}
	}
	return ""
}

// driverAcceptsOffer is a deliberately simple Rule-Engine-only check (not
// one of the sprint brief's 4 named AI decision points) — high fatigue or a
// very low rating makes a driver more likely to decline an offer.
func driverAcceptsOffer(w *World, driverID string) bool {
	d, ok := w.Drivers[driverID]
	if !ok {
		return false
	}
	acceptProb := 0.9 - 0.3*d.Fatigue
	if d.Rating < 4.0 {
		acceptProb -= 0.15
	}
	accepted := w.Rand.Float64() < acceptProb
	w.recordDriverOfferOutcome(accepted)
	if accepted {
		d.OffersAccepted++
	} else {
		d.OffersRejected++
	}
	return accepted
}
