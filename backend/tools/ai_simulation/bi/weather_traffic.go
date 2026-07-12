package bi

import "github.com/fairride/ai_simulation/domain/entity"

// WeatherImpact is one weather condition's effect on demand/supply/ETA/
// cancel/fare — PHẦN 6. "Driver Online" is the average number of drivers
// online while this weather was active, sampled the same way heatmap.json's
// Supply layer is (see simulation/world.go's RecordZoneSupplySample) is
// NOT segmented by weather anywhere today, so this field instead reports
// the number of DISTINCT drivers who completed at least one trip while this
// weather was active — a real, derivable proxy — disclosed in Assumptions.
type WeatherImpact struct {
	Weather               string  `json:"weather"`
	DemandCount           int     `json:"demand_count"` // requested trips
	CancelledCount        int     `json:"cancelled_count"`
	CancelRatePercent     float64 `json:"cancel_rate_percent"`
	AverageETAMinutes     float64 `json:"average_eta_minutes"`
	AverageFareVND        float64 `json:"average_fare_vnd"`
	DriversActiveCount    int     `json:"drivers_active_count"` // distinct drivers who completed a trip during this weather
}

type WeatherAnalysis struct {
	ByWeather   []WeatherImpact `json:"by_weather"`
	Assumptions []Assumption    `json:"assumptions"`
}

// ComputeWeatherAnalysis is PHẦN 6.
func ComputeWeatherAnalysis(in Input) WeatherAnalysis {
	type acc struct {
		requested, cancelled, completed int
		etaSum, fareSum                 float64
		drivers                         map[string]bool
	}
	byWeather := map[entity.Weather]*acc{}
	get := func(w entity.Weather) *acc {
		a := byWeather[w]
		if a == nil {
			a = &acc{drivers: map[string]bool{}}
			byWeather[w] = a
		}
		return a
	}

	for _, t := range in.Trips {
		a := get(t.Weather)
		a.requested++
		switch t.Outcome {
		case entity.OutcomeCompleted:
			a.completed++
			a.etaSum += t.ETAMinutes
			a.fareSum += float64(t.FinalFareVND)
			if t.DriverID != "" {
				a.drivers[t.DriverID] = true
			}
		case entity.OutcomeCancelled:
			a.cancelled++
		}
	}

	var out WeatherAnalysis
	for weather, a := range byWeather {
		wi := WeatherImpact{Weather: string(weather), DemandCount: a.requested, CancelledCount: a.cancelled, DriversActiveCount: len(a.drivers)}
		if a.requested > 0 {
			wi.CancelRatePercent = 100 * float64(a.cancelled) / float64(a.requested)
		}
		if a.completed > 0 {
			wi.AverageETAMinutes = a.etaSum / float64(a.completed)
			wi.AverageFareVND = a.fareSum / float64(a.completed)
		}
		out.ByWeather = append(out.ByWeather, wi)
	}
	out.Assumptions = []Assumption{
		{Title: "\"Driver Online\" theo thời tiết", Detail: "Simulation không lưu số tài xế online tại từng thời điểm theo thời tiết cụ thể — dùng proxy \"số tài xế có ít nhất 1 chuyến hoàn tất trong điều kiện thời tiết này\" thay vì bịa số liệu online tức thời."},
		{Title: "\"Supply\" theo thời tiết không tách riêng", Detail: "heatmap.json's Supply layer lấy mẫu theo zone/giờ, không theo thời tiết — không tổng hợp lại thành 1 số riêng cho từng loại thời tiết vì sẽ trộn lẫn nhiều ngày có thời tiết khác nhau ở cùng khung giờ, dễ gây hiểu lầm."},
	}
	return out
}

// TrafficImpact is PHẦN 7.
type TrafficImpact struct {
	Traffic              string  `json:"traffic"`
	TripCount            int     `json:"trip_count"`
	AverageETAMinutes    float64 `json:"average_eta_minutes"`
	AveragePickupMinutes float64 `json:"average_pickup_minutes"`
	AverageTripTimeMinutes float64 `json:"average_trip_time_minutes"` // ETA - Pickup, i.e. actual driving/transit duration
	AverageDriverIncomeVND float64 `json:"average_driver_income_vnd"`
}

type TrafficAnalysis struct {
	ByTraffic []TrafficImpact `json:"by_traffic"`
}

// ComputeTrafficAnalysis is PHẦN 7.
func ComputeTrafficAnalysis(in Input) TrafficAnalysis {
	type acc struct {
		count                     int
		etaSum, pickupSum, tripSum float64
		incomeSum                 int64
	}
	byTraffic := map[entity.Traffic]*acc{}
	for _, t := range in.Trips {
		if t.Outcome != entity.OutcomeCompleted {
			continue
		}
		a := byTraffic[t.Traffic]
		if a == nil {
			a = &acc{}
			byTraffic[t.Traffic] = a
		}
		a.count++
		a.etaSum += t.ETAMinutes
		a.pickupSum += t.PickupMinutes
		a.tripSum += drivingMinutes(t)
		a.incomeSum += t.DriverNetVND
	}

	var out TrafficAnalysis
	for traffic, a := range byTraffic {
		if a.count == 0 {
			continue
		}
		out.ByTraffic = append(out.ByTraffic, TrafficImpact{
			Traffic: string(traffic), TripCount: a.count,
			AverageETAMinutes: a.etaSum / float64(a.count), AveragePickupMinutes: a.pickupSum / float64(a.count),
			AverageTripTimeMinutes: a.tripSum / float64(a.count), AverageDriverIncomeVND: float64(a.incomeSum) / float64(a.count),
		})
	}
	return out
}
