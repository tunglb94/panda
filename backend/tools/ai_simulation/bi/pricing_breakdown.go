package bi

import "github.com/fairride/ai_simulation/domain/entity"

// PricingBreakdown is PHẦN 4 — giá/km (VND per km) segmented across every
// dimension the brief lists, each as a full Median/Mean/P95/Min/Max
// distribution over real completed trips. Every trip contributes to
// exactly one bucket per dimension (a trip is never split/duplicated
// across buckets within the same dimension).
type PricingBreakdown struct {
	ByServiceType map[string]Distribution `json:"by_service_type"` // bike/bike_plus/car/car_xl
	ByProductKind map[string]Distribution `json:"by_product_kind"` // ride vs delivery
	ByDaypart     map[string]Distribution `json:"by_daypart"`      // morning/noon/afternoon/night
	ByWeather     map[string]Distribution `json:"by_weather"`
	ByCalendar    map[string]Distribution `json:"by_calendar"` // holiday / weekend / normal
	ByZone        map[string]Distribution `json:"by_zone"`
}

// ComputePricingBreakdown is PHẦN 4. fareVNDPerKm is only defined for
// trips with DistanceKM > 0 (always true in practice — ride_flow.go/
// delivery_flow.go both clamp distance to a minimum of 0.8km).
func ComputePricingBreakdown(in Input) PricingBreakdown {
	byVehicle := map[string][]float64{}
	byProduct := map[string][]float64{}
	byDaypart := map[string][]float64{}
	byWeather := map[string][]float64{}
	byCalendar := map[string][]float64{}
	byZone := map[string][]float64{}

	for _, t := range in.Trips {
		if t.Outcome != entity.OutcomeCompleted || t.DistanceKM <= 0 {
			continue
		}
		perKm := float64(t.FinalFareVND) / t.DistanceKM

		byVehicle[string(t.ServiceType)] = append(byVehicle[string(t.ServiceType)], perKm)
		byProduct[string(t.Kind)] = append(byProduct[string(t.Kind)], perKm)
		byDaypart[daypart(requestedHour(t))] = append(byDaypart[daypart(requestedHour(t))], perKm)
		byWeather[string(t.Weather)] = append(byWeather[string(t.Weather)], perKm)
		byZone[string(t.PickupZone)] = append(byZone[string(t.PickupZone)], perKm)

		cal := "normal"
		if isHolidayTrip(t) {
			cal = "holiday"
		} else if isWeekendTrip(t) {
			cal = "weekend"
		}
		byCalendar[cal] = append(byCalendar[cal], perKm)
	}

	build := func(m map[string][]float64) map[string]Distribution {
		out := make(map[string]Distribution, len(m))
		for k, v := range m {
			out[k] = ComputeDistribution(v)
		}
		return out
	}

	return PricingBreakdown{
		ByServiceType: build(byVehicle),
		ByProductKind: build(byProduct),
		ByDaypart:     build(byDaypart),
		ByWeather:     build(byWeather),
		ByCalendar:    build(byCalendar),
		ByZone:        build(byZone),
	}
}
