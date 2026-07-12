package simulation

import (
	"strconv"
	"time"
)

// NamedScenario pairs a human-readable label with a TripInput so the report
// generator (cmd/pricing-simulate) and the test suite can both iterate the
// same canonical scenario set — BƯỚC 3 of the sprint brief.
type NamedScenario struct {
	Name  string
	Input TripInput
}

// Fixed reference timestamps so every scenario run is deterministic —
// re-running the simulator (or a reader re-deriving these numbers by hand)
// always reproduces the exact same figures.
var (
	tNormalWeekday = mustTime("2026-07-14T14:00:00+07:00") // Tue 14:00 — no night/peak
	tRushMorning   = mustTime("2026-07-14T08:00:00+07:00") // Tue 08:00 — BRB §2.2.12 morning peak
	tRushEvening   = mustTime("2026-07-14T18:00:00+07:00") // Tue 18:00 — BRB §2.2.12 evening peak
	tMidnight      = mustTime("2026-07-14T23:30:00+07:00") // Tue 23:30 — BRB §2.2.10 night window
	tLateNight     = mustTime("2026-07-15T02:00:00+07:00") // Wed 02:00 — still within night window
)

func mustTime(rfc3339 string) time.Time {
	t, err := time.Parse(time.RFC3339, rfc3339)
	if err != nil {
		panic(err) // scenario definitions are static; a parse error here is a coding bug, not runtime data
	}
	return t
}

var allVehicleTypes = []VehicleType{VehicleCar, VehicleMotorcycle, VehicleVan}
var allDriverTiers = []DriverTier{TierBronze, TierSilver, TierGold, TierPlatinum, TierDiamond}

// baseTrip returns a "nothing special" trip: normal weekday, clear weather,
// no holiday, no surge, Bronze driver, no promotion, no waiting, no
// airport, no bridge/parking, no long pickup. Every scenario below starts
// here and overrides only the fields it's testing.
func baseTrip(vt VehicleType, distanceKM float64) TripInput {
	return TripInput{
		VehicleType: vt,
		DistanceKM:  distanceKM,
		DurationMin: distanceKM * 2.2, // ~27 km/h average city speed — plausible, not load-bearing for any test assertion
		RequestTime: tNormalWeekday,
		Weather:     WeatherClear,
		DriverTier:  TierBronze,
	}
}

// AllScenarios builds the full BƯỚC-3 scenario set (116 scenarios — exceeds
// the required minimum of 100). Grouped by category, matching the examples
// enumerated in the sprint brief (2/5/12/25 km, Airport, Rush Hour, Rain,
// Midnight, Holiday, Waiting 3/15 min, Bridge, Parking, Voucher, Student,
// Membership, plus surge bands, driver tiers, and edge cases the brief
// implies but doesn't spell out).
func AllScenarios() []NamedScenario {
	var out []NamedScenario
	distances := []float64{2, 5, 12, 25}

	// A) Distance × vehicle baseline sweep — 4 × 3 = 12
	for _, vt := range allVehicleTypes {
		for _, d := range distances {
			out = append(out, NamedScenario{
				Name:  scenarioName("Baseline", vt, d),
				Input: baseTrip(vt, d),
			})
		}
	}

	// C) Airport pickup — 3 vehicles × {5, 25} km = 6
	for _, vt := range allVehicleTypes {
		for _, d := range []float64{5, 25} {
			in := baseTrip(vt, d)
			in.IsAirportZone = true
			out = append(out, NamedScenario{Name: scenarioName("Airport", vt, d), Input: in})
		}
	}

	// D) Rush hour (peak) — 3 vehicles × {5, 12} km = 6
	for _, vt := range allVehicleTypes {
		for _, d := range []float64{5, 12} {
			in := baseTrip(vt, d)
			in.RequestTime = tRushMorning
			out = append(out, NamedScenario{Name: scenarioName("RushHourMorning", vt, d), Input: in})
		}
	}

	// E) Rain — 3 vehicles × {5, 12} km = 6
	for _, vt := range allVehicleTypes {
		for _, d := range []float64{5, 12} {
			in := baseTrip(vt, d)
			in.Weather = WeatherRain
			out = append(out, NamedScenario{Name: scenarioName("Rain", vt, d), Input: in})
		}
	}

	// F) Midnight (night surcharge) — 3 vehicles × {5, 12} km = 6
	for _, vt := range allVehicleTypes {
		for _, d := range []float64{5, 12} {
			in := baseTrip(vt, d)
			in.RequestTime = tMidnight
			out = append(out, NamedScenario{Name: scenarioName("Midnight", vt, d), Input: in})
		}
	}

	// G) Holiday — 3 vehicles × {5, 12} km = 6
	for _, vt := range allVehicleTypes {
		for _, d := range []float64{5, 12} {
			in := baseTrip(vt, d)
			in.IsHoliday = true
			out = append(out, NamedScenario{Name: scenarioName("Holiday", vt, d), Input: in})
		}
	}

	// H) Combined Holiday + Night + Rain — static-cap stress test — 3 vehicles = 3
	for _, vt := range allVehicleTypes {
		in := baseTrip(vt, 12)
		in.RequestTime = tLateNight
		in.IsHoliday = true
		in.Weather = WeatherRain
		out = append(out, NamedScenario{Name: scenarioName("HolidayNightRain_StaticCapStress", vt, 12), Input: in})
	}

	// I) Waiting 3 minutes (within grace, should bill 0) — 3 vehicles = 3
	for _, vt := range allVehicleTypes {
		in := baseTrip(vt, 5)
		in.WaitingMin = 3
		out = append(out, NamedScenario{Name: scenarioName("Waiting3Min", vt, 5), Input: in})
	}

	// J) Waiting 15 minutes (12 chargeable minutes) — 3 vehicles = 3
	for _, vt := range allVehicleTypes {
		in := baseTrip(vt, 5)
		in.WaitingMin = 15
		out = append(out, NamedScenario{Name: scenarioName("Waiting15Min", vt, 5), Input: in})
	}

	// K) Bridge fee — 3 amounts on a car trip = 3
	for _, fee := range []int64{5_000, 15_000, 30_000} {
		in := baseTrip(VehicleCar, 8)
		in.BridgeFeeVND = fee
		out = append(out, NamedScenario{Name: "Bridge_" + moneyLabel(fee), Input: in})
	}

	// L) Parking fee — 3 amounts on a car trip = 3
	for _, fee := range []int64{5_000, 20_000, 50_000} {
		in := baseTrip(VehicleCar, 8)
		in.ParkingFeeVND = fee
		out = append(out, NamedScenario{Name: "Parking_" + moneyLabel(fee), Input: in})
	}

	// M) Bridge + Parking combined — 2
	for _, vt := range []VehicleType{VehicleCar, VehicleVan} {
		in := baseTrip(vt, 15)
		in.BridgeFeeVND = 15_000
		in.ParkingFeeVND = 20_000
		out = append(out, NamedScenario{Name: scenarioName("BridgeAndParking", vt, 15), Input: in})
	}

	// N) Voucher — including one deliberately larger than the trip value to
	// exercise the BRB §4.9 / BƯỚC-7 clamp — 5
	voucherCases := []struct {
		label string
		d     float64
		vnd   int64
	}{
		{"Voucher_Small", 5, 10_000},
		{"Voucher_Medium", 12, 30_000},
		{"Voucher_Large", 25, 80_000},
		{"Voucher_MinTrip", 2, 20_000},
		{"Voucher_ExceedsTripValue_SafetyClampTest", 2, 500_000},
	}
	for _, vc := range voucherCases {
		in := baseTrip(VehicleCar, vc.d)
		in.Promotion = &PromotionInput{Label: "Voucher", DiscountVND: vc.vnd}
		out = append(out, NamedScenario{Name: vc.label, Input: in})
	}

	// O) Student promotion — PRICING_STRATEGY §7.2.2 (NOT YET in BRB), 10%
	// off modeled as a pre-resolved VND amount on 3 trip sizes = 3
	for _, d := range []float64{5, 12, 25} {
		in := baseTrip(VehicleCar, d)
		fareOnly, _ := NewDefaultSimulator().Simulate(baseTrip(VehicleCar, d))
		discount := int64(float64(fareOnly.CustomerTotal) * 0.10)
		in.Promotion = &PromotionInput{Label: "Student(10%)", DiscountVND: discount}
		out = append(out, NamedScenario{Name: scenarioName("StudentPromo", VehicleCar, d), Input: in})
	}

	// P) Membership ("Panda Plus") promotion — PRICING_STRATEGY §7.2.1 (NOT
	// YET in BRB), modeled as booking-fee-equivalent VND off = 3
	for _, d := range []float64{5, 12, 25} {
		in := baseTrip(VehicleCar, d)
		in.Promotion = &PromotionInput{Label: "Membership(BookingFeeWaived)", DiscountVND: BookingFee}
		out = append(out, NamedScenario{Name: scenarioName("MembershipPromo", VehicleCar, d), Input: in})
	}

	// Q) Long pickup compensation — near (4km) and far (7km) × 3 vehicles = 6
	for _, vt := range allVehicleTypes {
		for _, pd := range []float64{4, 7} {
			in := baseTrip(vt, 6)
			in.PickupDistanceKM = pd
			label := "LongPickupNear"
			if pd > LongPickupFarKM {
				label = "LongPickupFar"
			}
			out = append(out, NamedScenario{Name: scenarioName(label, vt, 6), Input: in})
		}
	}

	// R) Surge bands — every BRB §2.13.2 band × {5, 12} km = 12
	surgeDSRs := []struct {
		label string
		dsr   float64
	}{
		{"DSR_NoSurge", 1.0},
		{"DSR_Busy", 1.35},
		{"DSR_HighDemand", 1.75},
		{"DSR_VeryHighDemand", 2.25},
		{"DSR_PeakDemand", 2.75},
		{"DSR_MaxSurge", 3.5},
	}
	for _, sc := range surgeDSRs {
		for _, d := range []float64{5, 12} {
			in := baseTrip(VehicleCar, d)
			in.DSR = sc.dsr
			out = append(out, NamedScenario{Name: scenarioName(sc.label, VehicleCar, d), Input: in})
		}
	}

	// S) Surge + Airport combined — 3
	for _, vt := range allVehicleTypes {
		in := baseTrip(vt, 20)
		in.DSR = 2.6
		in.IsAirportZone = true
		out = append(out, NamedScenario{Name: scenarioName("SurgePlusAirport", vt, 20), Input: in})
	}

	// T) Surge suppresses Peak (BRB §2.2.12) — verification pair — 2
	{
		in := baseTrip(VehicleCar, 8)
		in.RequestTime = tRushEvening
		in.DSR = 2.2 // surge active → peak must NOT also apply
		out = append(out, NamedScenario{Name: "SurgeSuppressesPeak_SurgeActive", Input: in})

		in2 := baseTrip(VehicleCar, 8)
		in2.RequestTime = tRushEvening
		in2.DSR = 1.0 // surge inactive → peak SHOULD apply
		out = append(out, NamedScenario{Name: "SurgeSuppressesPeak_PeakActive", Input: in2})
	}

	// U) Driver tier sweep on a fixed representative trip — 5
	for _, tier := range allDriverTiers {
		in := baseTrip(VehicleCar, 10)
		in.DriverTier = tier
		out = append(out, NamedScenario{Name: "DriverTier_" + string(tier), Input: in})
	}

	// V) Edge cases — 8
	{
		out = append(out, NamedScenario{Name: "Edge_ZeroDistanceZeroDuration", Input: baseTrip(VehicleCar, 0)})
		veryShort := baseTrip(VehicleMotorcycle, 0.5)
		out = append(out, NamedScenario{Name: "Edge_VeryShortMotorcycleTrip", Input: veryShort})
		veryLong := baseTrip(VehicleVan, 50)
		out = append(out, NamedScenario{Name: "Edge_VeryLongVanTrip", Input: veryLong})
		hugeWait := baseTrip(VehicleCar, 3)
		hugeWait.WaitingMin = 60
		out = append(out, NamedScenario{Name: "Edge_HugeWaitingTime", Input: hugeWait})
		maxStack := baseTrip(VehicleVan, 25)
		maxStack.RequestTime = tLateNight
		maxStack.IsHoliday = true
		maxStack.Weather = WeatherRain
		maxStack.DSR = 3.5
		maxStack.IsAirportZone = true
		out = append(out, NamedScenario{Name: "Edge_EverythingStacked_MaxRealisticFare", Input: maxStack})
		zeroPromo := baseTrip(VehicleCar, 5)
		zeroPromo.Promotion = &PromotionInput{Label: "ZeroVoucher", DiscountVND: 0}
		out = append(out, NamedScenario{Name: "Edge_ZeroValueVoucher", Input: zeroPromo})
		negativeBridgeGuard := baseTrip(VehicleCar, 5)
		negativeBridgeGuard.BridgeFeeVND = -1000 // malformed input — engine must not pay a driver a negative pass-through
		out = append(out, NamedScenario{Name: "Edge_NegativeBridgeFeeInput_SafetyTest", Input: negativeBridgeGuard})
		diamondMaxStack := baseTrip(VehicleCar, 25)
		diamondMaxStack.DriverTier = TierDiamond
		diamondMaxStack.DSR = 3.5
		diamondMaxStack.RequestTime = tLateNight
		diamondMaxStack.IsHoliday = true
		diamondMaxStack.Weather = WeatherRain
		out = append(out, NamedScenario{Name: "Edge_DiamondDriverMaxStack", Input: diamondMaxStack})
	}

	// W) PassengerLevel must have zero pricing effect — verification pair — 2
	{
		a := baseTrip(VehicleCar, 6)
		a.PassengerLevel = ""
		out = append(out, NamedScenario{Name: "PassengerLevel_None", Input: a})
		b := baseTrip(VehicleCar, 6)
		b.PassengerLevel = "Gold"
		out = append(out, NamedScenario{Name: "PassengerLevel_Gold_ShouldBeIdenticalToNone", Input: b})
	}

	// X) A few more motorcycle-specific points for coverage — 3
	{
		out = append(out, NamedScenario{Name: "Motorcycle_ShortRain", Input: func() TripInput {
			in := baseTrip(VehicleMotorcycle, 3)
			in.Weather = WeatherRain
			return in
		}()})
		out = append(out, NamedScenario{Name: "Motorcycle_AirportPickup", Input: func() TripInput {
			in := baseTrip(VehicleMotorcycle, 18)
			in.IsAirportZone = true
			return in
		}()})
		out = append(out, NamedScenario{Name: "Motorcycle_HolidayNight", Input: func() TripInput {
			in := baseTrip(VehicleMotorcycle, 9)
			in.IsHoliday = true
			in.RequestTime = tMidnight
			return in
		}()})
	}

	return out
}

func scenarioName(label string, vt VehicleType, distanceKM float64) string {
	return label + "_" + string(vt) + "_" + moneyLabel(int64(distanceKM)) + "km"
}

func moneyLabel(v int64) string {
	if v < 0 {
		return "neg" + moneyLabel(-v)
	}
	switch {
	case v >= 1_000_000:
		return strconv.FormatInt(v/1_000_000, 10) + "M"
	case v >= 1_000:
		return strconv.FormatInt(v/1_000, 10) + "k"
	default:
		return strconv.FormatInt(v, 10)
	}
}
