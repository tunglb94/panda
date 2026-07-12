package bi

import "github.com/fairride/ai_simulation/domain/entity"

// AirportAnalysis is PHẦN 15 — a Sân bay-only deep dive. Trips include
// both pickup-at-airport and dropoff-at-airport (a trip touching Airport
// either way is "airport traffic"), matching audit.Report's own Airport
// Profit definition (§16 of the earlier Business Audit phase) so the two
// numbers stay comparable.
type AirportAnalysis struct {
	TripCount               int     `json:"trip_count"`
	AverageETAMinutes       float64 `json:"average_eta_minutes"`
	AveragePassengerWaitingMinutes float64 `json:"average_passenger_waiting_minutes"` // PickupMinutes for airport trips
	RevenueVND               int64   `json:"revenue_vnd"`
	CancelRatePercent        float64 `json:"cancel_rate_percent"`

	NotModeled  []string     `json:"not_modeled"`
	Assumptions []Assumption `json:"assumptions"`
}

// ComputeAirportAnalysis is PHẦN 15.
func ComputeAirportAnalysis(in Input) AirportAnalysis {
	var out AirportAnalysis
	var requested, cancelled, completed int
	var etaSum, waitSum float64
	var revenue int64

	for _, t := range in.Trips {
		if t.PickupZone != entity.ZoneAirport && t.DestinationZone != entity.ZoneAirport {
			continue
		}
		requested++
		switch t.Outcome {
		case entity.OutcomeCompleted:
			completed++
			etaSum += t.ETAMinutes
			waitSum += t.PickupMinutes
			revenue += t.FinalFareVND
		case entity.OutcomeCancelled:
			cancelled++
		}
	}

	out.TripCount = requested
	out.RevenueVND = revenue
	if completed > 0 {
		out.AverageETAMinutes = etaSum / float64(completed)
		out.AveragePassengerWaitingMinutes = waitSum / float64(completed)
	}
	if requested > 0 {
		out.CancelRatePercent = 100 * float64(cancelled) / float64(requested)
	}
	out.NotModeled = []string{"queue (không có hàng đợi vật lý tại sân bay)", "driver waiting (tài xế không \"chờ\" tại 1 khu vực — được ghép ngay khi có yêu cầu qua Dispatch gần nhất)"}
	out.Assumptions = []Assumption{
		{Title: "Airport trips = pickup HOẶC dropoff tại Sân bay", Detail: "Khớp với định nghĩa Airport Profit (§16) trong Business Audit phase trước — một chuyến chỉ cần chạm Sân bay ở một đầu là được tính, không chỉ pickup."},
	}
	return out
}
