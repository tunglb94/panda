package simulation

import (
	"math/rand"

	"github.com/fairride/ai_simulation/domain/entity"
)

// ScenarioScheduler decides Weather/Traffic/ActiveScenarios. Weather and
// traffic are redrawn once per simulated day (not every tick) so conditions
// persist realistically instead of flickering minute to minute; time-of-day
// scenarios (rush hour, lunch) are a pure function of the current hour and
// need no scheduling state at all.
type ScenarioScheduler struct {
	rnd *rand.Rand
}

func NewScenarioScheduler(rnd *rand.Rand) *ScenarioScheduler {
	return &ScenarioScheduler{rnd: rnd}
}

// RollDailyConditions redraws weather/traffic-base/rare-event state for the
// day that just started. Probabilities here are simulation-design choices
// (not sourced from any BRB number — there is no real weather/traffic
// dataset to calibrate against), documented inline rather than left as
// unexplained magic numbers.
func (s *ScenarioScheduler) RollDailyConditions(w *World) {
	w.Weather = s.rollWeather()
	w.Scenarios = entity.ActiveScenarios{}

	if w.Clock.IsWeekend() {
		w.Scenarios[entity.ScenarioWeekend] = true
	}
	if w.Weather.IsRainy() {
		w.Scenarios[entity.ScenarioRain] = true
	}

	// Rare events: ~4% chance of a concert, ~2% chance of a festival, ~3%
	// chance of an accident somewhere in the city, on any given day.
	if s.rnd.Float64() < 0.04 {
		w.Scenarios[entity.ScenarioConcert] = true
	}
	if s.rnd.Float64() < 0.02 {
		w.Scenarios[entity.ScenarioFestival] = true
	}
	if s.rnd.Float64() < 0.03 {
		w.Scenarios[entity.ScenarioAccident] = true
	}
	// Every 10th simulated day is treated as a public holiday — a simple,
	// deterministic stand-in so Holiday-surcharge/promotion logic gets
	// exercised across a multi-day run without a real VN holiday calendar.
	if w.Clock.Day()%10 == 9 {
		w.Scenarios[entity.ScenarioHoliday] = true
	}
}

func (s *ScenarioScheduler) rollWeather() entity.Weather {
	r := s.rnd.Float64()
	switch {
	case r < 0.60:
		return entity.WeatherSunny
	case r < 0.85:
		return entity.WeatherLightRain
	case r < 0.97:
		return entity.WeatherHeavyRain
	default:
		return entity.WeatherFlooded
	}
}

// RollTraffic redraws city traffic — called a few times per simulated day
// (see engine.go) rather than once, since congestion realistically shifts
// faster than weather. Rush hours and rain both bias toward heavier traffic.
func (s *ScenarioScheduler) RollTraffic(w *World) {
	hour := w.Clock.Hour()
	isRush := (hour >= 7 && hour < 9) || (hour >= 17 && hour < 20)

	r := s.rnd.Float64()
	switch {
	case w.Scenarios.Has(entity.ScenarioAccident) && r < 0.5:
		w.Traffic = entity.TrafficAccident
	case isRush && r < 0.55:
		w.Traffic = entity.TrafficJammed
	case isRush || w.Weather.IsRainy():
		w.Traffic = entity.TrafficBusy
	case r < 0.6:
		w.Traffic = entity.TrafficClear
	default:
		w.Traffic = entity.TrafficBusy
	}
}

// ApplyHourlyScenarios sets/clears the time-of-day scenario flags for the
// current hour — a pure function of hour, re-evaluated every tick.
func ApplyHourlyScenarios(w *World) {
	hour := w.Clock.Hour()
	delete(w.Scenarios, entity.ScenarioMorningRush)
	delete(w.Scenarios, entity.ScenarioLunch)
	delete(w.Scenarios, entity.ScenarioEveningRush)

	switch {
	case hour >= 7 && hour < 9:
		w.Scenarios[entity.ScenarioMorningRush] = true
	case hour >= 11 && hour < 13:
		w.Scenarios[entity.ScenarioLunch] = true
	case hour >= 17 && hour < 20:
		w.Scenarios[entity.ScenarioEveningRush] = true
	}
}

// baseDeliveryShare is each zone's baseline share of "what does a rider in
// this zone request" that is a Delivery rather than a Ride — a simulation-
// design assumption (not a BRB number; no real Ride/Delivery request-mix
// dataset exists to calibrate against), set per the sprint brief's own
// worked examples: CBD/Airport lean Ride (people travelling), Industrial
// (KCN)/Residential lean Delivery (parcels sent to/from home and workplaces).
var baseDeliveryShare = map[entity.ZoneType]float64{
	entity.ZoneCBD:           0.10,
	entity.ZoneResidential:   0.45,
	entity.ZoneIndustrial:    0.55,
	entity.ZoneAirport:       0.05,
	entity.ZoneBusStation:    0.15,
	entity.ZoneHospital:      0.20,
	entity.ZoneSchool:        0.15,
	entity.ZoneEntertainment: 0.10,
}

// DeliveryProbability returns the probability that a rider requesting
// something from zone this tick wants a Delivery rather than a Ride.
// Weekend and rain both skew further toward Delivery — a documented
// simulation-design assumption (parcel/e-commerce demand rising on
// weekends/bad weather is a real, commonly observed pattern in the
// industry research this tool's sibling design doc cites, but the specific
// multipliers here are not sourced from any BRB number).
func DeliveryProbability(w *World, zone entity.ZoneType) float64 {
	p := baseDeliveryShare[zone]
	if w.Scenarios.Has(entity.ScenarioWeekend) {
		p *= 1.3
	}
	if w.Weather.IsRainy() {
		p *= 1.2
	}
	if p > 0.9 {
		p = 0.9 // always leave some Ride share even under maximum skew
	}
	return p
}

// DemandMultiplier scales how likely riders are to request trips this tick,
// derived from active scenarios — used by engine.go's trip-generation step.
func DemandMultiplier(w *World) float64 {
	m := 1.0
	if w.Scenarios.Has(entity.ScenarioMorningRush) || w.Scenarios.Has(entity.ScenarioEveningRush) {
		m *= 1.8
	}
	if w.Scenarios.Has(entity.ScenarioLunch) {
		m *= 1.3
	}
	if w.Scenarios.Has(entity.ScenarioWeekend) {
		m *= 1.15
	}
	if w.Scenarios.Has(entity.ScenarioRain) {
		m *= 1.25 // more people want a ride, fewer want to walk/bike
	}
	if w.Scenarios.Has(entity.ScenarioConcert) || w.Scenarios.Has(entity.ScenarioFestival) {
		m *= 1.6
	}
	if w.Scenarios.Has(entity.ScenarioHoliday) {
		m *= 0.8 // fewer commute trips, partially offset by leisure trips elsewhere
	}
	return m
}
