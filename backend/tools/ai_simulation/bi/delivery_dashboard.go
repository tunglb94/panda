package bi

import "github.com/fairride/ai_simulation/domain/entity"

// DeliveryDashboard is PHẦN 10 — reuses stats.DeliveryStatistics (request/
// accept/reject/cancel/pickup-time/delivery-time/distance/weight) directly
// rather than recomputing those; only DeliveryProfitVND is newly derived
// here (stats.DeliveryStatistics has no profit field).
type DeliveryDashboard struct {
	RideCount             int     `json:"ride_count"`
	DeliveryCount         int     `json:"delivery_count"`
	DeliveryRevenueVND    int64   `json:"delivery_revenue_vnd"`
	DeliveryProfitVND     int64   `json:"delivery_profit_vnd"`
	AverageWeightKg       float64 `json:"average_weight_kg"`
	AverageDistanceKM     float64 `json:"average_distance_km"`
	AveragePickupMinutes  float64 `json:"average_pickup_minutes"`
	AverageDropoffMinutes float64 `json:"average_dropoff_minutes"` // pickup -> delivered transit time
	DeliveryAcceptanceRatePercent float64 `json:"delivery_acceptance_rate_percent"`
	DeliveryCancelRatePercent     float64 `json:"delivery_cancel_rate_percent"`

	NotModeled  []string     `json:"not_modeled"`
	Assumptions []Assumption `json:"assumptions"`
}

// ComputeDeliveryDashboard is PHẦN 10.
func ComputeDeliveryDashboard(in Input) DeliveryDashboard {
	ds := in.Bundle.DeliveryStatistics

	var rideCount, deliveryCount int
	var deliveryProfit int64
	for _, t := range in.Trips {
		if t.Outcome != entity.OutcomeCompleted {
			continue
		}
		if t.Kind == entity.KindDelivery {
			deliveryCount++
			deliveryProfit += profitFor(t)
		} else {
			rideCount++
		}
	}

	out := DeliveryDashboard{
		RideCount: rideCount, DeliveryCount: deliveryCount,
		DeliveryRevenueVND: in.BI.DeliveryRevenueVND, DeliveryProfitVND: deliveryProfit,
		AverageWeightKg: ds.AverageWeightKg, AverageDistanceKM: ds.AverageDistanceKM,
		AveragePickupMinutes: ds.AveragePickupMinutes, AverageDropoffMinutes: ds.AverageDeliveryMinutes,
		NotModeled: []string{
			"restaurant waiting (không có khái niệm nhà hàng, chỉ có điểm lấy hàng chung)",
			"food cold % (không theo dõi nhiệt độ hàng hoá)",
			"late delivery % (không có SLA thời gian giao được định nghĩa để so sánh)",
		},
		Assumptions: []Assumption{
			{Title: "Delivery Profit dùng chung công thức Unit Economics", Detail: "PerTripProfitVND (VAT 10% + chi phí hạ tầng ước tính) áp dụng như Ride — Delivery không có mô hình chi phí riêng."},
		},
	}
	if ds.Requested > 0 {
		out.DeliveryAcceptanceRatePercent = 100 * float64(ds.Accepted) / float64(ds.Requested)
		out.DeliveryCancelRatePercent = 100 * float64(ds.Cancelled) / float64(ds.Requested)
	}
	return out
}
