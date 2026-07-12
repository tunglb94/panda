package bi

import "github.com/fairride/ai_simulation/domain/entity"

// DriverDistribution is PHẦN 2.
type DriverDistribution struct {
	// VehicleMixPercent classifies each driver who completed at least one
	// trip this run by what they actually did: "bike"/"car" (Ride-only, by
	// VehicleType), "delivery" (Delivery-only), "hybrid" (both this run).
	// Drivers with zero completed trips are excluded (not fabricated into a
	// category) — see ZeroTripDriverCount.
	VehicleMixPercent  map[string]float64 `json:"vehicle_mix_percent"`
	VehicleMixCount    map[string]int     `json:"vehicle_mix_count"`
	ZeroTripDriverCount int               `json:"zero_trip_driver_count"`

	// StatusCounts — only states this simulation actually models are real
	// counts; anything else is 0 with a matching Assumption entry (see
	// Assumptions) rather than a fabricated proportion.
	StatusCounts map[string]int `json:"status_counts"`
	Assumptions  []Assumption   `json:"assumptions"`
}

// ComputeDriverDistribution is PHẦN 2.
func ComputeDriverDistribution(in Input) DriverDistribution {
	out := DriverDistribution{
		VehicleMixPercent: map[string]float64{},
		VehicleMixCount:   map[string]int{"bike": 0, "car": 0, "delivery": 0, "hybrid": 0},
		StatusCounts: map[string]int{
			"online": 0, "offline": 0, "resting": 0, "eating": 0,
			"refueling": 0, "low_battery": 0, "no_network": 0, "app_off": 0,
		},
	}

	rideByDriver := map[string]bool{}
	deliveryByDriver := map[string]bool{}
	for _, t := range in.Trips {
		if t.Outcome != entity.OutcomeCompleted || t.DriverID == "" {
			continue
		}
		if t.Kind == entity.KindDelivery {
			deliveryByDriver[t.DriverID] = true
		} else {
			rideByDriver[t.DriverID] = true
		}
	}

	for id, d := range in.Drivers {
		hasRide := rideByDriver[id]
		hasDelivery := deliveryByDriver[id]
		switch {
		case hasRide && hasDelivery:
			out.VehicleMixCount["hybrid"]++
		case hasDelivery:
			out.VehicleMixCount["delivery"]++
		case hasRide: // ride-only
			if d.VehicleType == entity.VehicleMotorcycle {
				out.VehicleMixCount["bike"]++
			} else {
				out.VehicleMixCount["car"]++
			}
		default:
			out.ZeroTripDriverCount++
		}

		if d.Online {
			out.StatusCounts["online"]++
			if d.PhoneBattery < 0.15 {
				out.StatusCounts["low_battery"]++
			}
		} else {
			out.StatusCounts["offline"]++
			out.StatusCounts["app_off"]++ // ASSUMPTION: no distinct "app off but backgrounded" state exists — see Assumptions
			if d.Fatigue > 0.15 {
				out.StatusCounts["resting"]++
			}
		}
	}

	total := len(in.Drivers) - out.ZeroTripDriverCount
	if total > 0 {
		for k, v := range out.VehicleMixCount {
			out.VehicleMixPercent[k] = 100 * float64(v) / float64(total)
		}
	}

	out.Assumptions = []Assumption{
		{Title: "\"Đang ăn\" không được mô hình hoá", Detail: "Không có state nghỉ-ăn riêng biệt trong simulation — driver chỉ có Online/Offline liên tục, không có sub-state theo hoạt động. Luôn báo cáo 0, không suy diễn."},
		{Title: "\"Đang đổ xăng\" không được mô hình hoá", Detail: "Fuel tự hồi phục khi offline (xem simulation/engine.go's evaluateDriverState) — không có hành động \"đổ xăng\" rời rạc để đếm. Luôn báo cáo 0."},
		{Title: "\"Mất mạng\" không được mô hình hoá", Detail: "Không có khái niệm kết nối mạng trong simulation. Luôn báo cáo 0."},
		{Title: "\"Tắt app\" = Offline", Detail: "Simulation không phân biệt tài xế chủ động tắt app với tài xế offline vì lý do khác (hết ca, factor an toàn...) — dùng chung 1 trạng thái Offline."},
		{Title: "\"Đang nghỉ\" là suy luận, không phải state thật", Detail: "Xấp xỉ bằng Offline + Fatigue>0.15 (đang trong giai đoạn hồi phục) — không phải một trạng thái driver tự chọn \"tôi đang nghỉ\" trong simulation."},
	}
	return out
}
