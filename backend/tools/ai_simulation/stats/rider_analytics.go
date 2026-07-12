package stats

import "github.com/fairride/ai_simulation/domain/entity"

// RiderAnalytics is PHẦN 5's requested rider-side deep-dive — a
// population-level aggregate view over both the rider agents' own state
// (price sensitivity, membership, income) and their trip ledger (fare,
// cancellations, retention), complementing rider_statistics.json.
type RiderAnalytics struct {
	AverageTrips         float64          `json:"average_ride_trips"`
	AverageDeliveries    float64          `json:"average_deliveries"`
	TotalVoucherSavedVND int64            `json:"total_voucher_saved_vnd"`
	TotalPromotionUsedCount int           `json:"total_promotion_used_count"`
	AverageFareVND       float64          `json:"average_fare_vnd"`
	AveragePriceSensitivity float64       `json:"average_price_sensitivity"`
	MembershipCounts     map[string]int   `json:"membership_counts"`
	RetentionRatePercent float64          `json:"retention_rate_percent"` // avg(distinct active days) / total simulated days
	CancelledRequestsTotal int            `json:"cancelled_requests_total"`
	CancellationRatePercent float64       `json:"cancellation_rate_percent"`
}

// BuildRiderAnalytics aggregates rider agent state plus the trip ledger.
// totalDays is the run's configured --days, used to normalize retention.
func (c *Collector) BuildRiderAnalytics(riders map[string]*entity.RiderAgent, trips []*entity.SimTrip, totalDays int) RiderAnalytics {
	out := RiderAnalytics{MembershipCounts: map[string]int{}}
	if len(riders) == 0 {
		return out
	}

	var sensitivitySum float64
	for _, r := range riders {
		sensitivitySum += r.PriceSensitivity
		out.MembershipCounts[string(r.Membership)]++
	}
	n := len(riders)
	out.AveragePriceSensitivity = sensitivitySum / float64(n)

	activeDaysByRider := map[string]map[int64]bool{}
	var rideCount, deliveryCount, cancelledCount, requestedCount int
	var fareSum float64
	for _, t := range trips {
		if t.RiderID == "" {
			continue
		}
		requestedCount++
		if activeDaysByRider[t.RiderID] == nil {
			activeDaysByRider[t.RiderID] = map[int64]bool{}
		}
		activeDaysByRider[t.RiderID][t.RequestedAtTick/(24*60)] = true

		switch t.Outcome {
		case entity.OutcomeCompleted:
			fareSum += float64(t.FinalFareVND)
			if t.Kind == entity.KindDelivery {
				deliveryCount++
			} else {
				rideCount++
			}
			if t.PromotionType == "manual_coupon" {
				out.TotalVoucherSavedVND += t.VoucherDiscountVND
			} else if t.PromotionType != "" {
				out.TotalPromotionUsedCount++
			}
		case entity.OutcomeCancelled:
			cancelledCount++
		}
	}

	out.AverageTrips = float64(rideCount) / float64(n)
	out.AverageDeliveries = float64(deliveryCount) / float64(n)
	out.AverageFareVND = avg(fareSum, rideCount+deliveryCount)
	out.CancelledRequestsTotal = cancelledCount
	if requestedCount > 0 {
		out.CancellationRatePercent = 100 * float64(cancelledCount) / float64(requestedCount)
	}

	if totalDays > 0 {
		var activeDaysSum float64
		for _, days := range activeDaysByRider {
			activeDaysSum += float64(len(days))
		}
		out.RetentionRatePercent = 100 * (activeDaysSum / float64(n)) / float64(totalDays)
	}
	return out
}
