package bi

import "github.com/fairride/ai_simulation/domain/entity"

// PassengerEconomy is PHẦN 3 — reuses stats.RiderAnalytics/PromotionROI for
// the fields that already exist there (Membership, Voucher used/unused,
// Average Fare) rather than recomputing them a second way; only the
// genuinely new dimensions (Km/ETA/Waiting/Cancel per rider, Trips/month)
// are computed here directly from the trip ledger.
type PassengerEconomy struct {
	AverageMonthlyIncomeVND float64        `json:"average_monthly_income_vnd"` // RiderAgent.Income — a seeded rider attribute, not derived from trips
	VoucherUsedCount        int            `json:"voucher_used_count"`
	VoucherUnusedCount      int            `json:"voucher_unused_count"` // issued but never redeemed by end of run — "bỏ phí"
	MembershipCounts        map[string]int `json:"membership_counts"`
	TripsPerMonth           float64        `json:"trips_per_month"` // scaled from this run's trip count by (30/Days)
	AverageSpendVND         float64        `json:"average_spend_vnd"`
	AverageFareVND          float64        `json:"average_fare_vnd"`
	AverageKm               float64        `json:"average_km"`
	AverageETAMinutes       float64        `json:"average_eta_minutes"`
	AverageWaitingMinutes   float64        `json:"average_waiting_minutes"` // pickup wait, i.e. PickupMinutes
	AverageCancelPercent    float64        `json:"average_cancel_percent"`  // per-rider mean of (own cancelled / own requested)
}

// ComputePassengerEconomy is PHẦN 3.
func ComputePassengerEconomy(in Input) PassengerEconomy {
	out := PassengerEconomy{MembershipCounts: in.Bundle.RiderAnalytics.MembershipCounts}
	out.VoucherUsedCount = in.Bundle.PromotionROI.UsedVoucherCount
	out.VoucherUnusedCount = in.Bundle.PromotionROI.IssuedVoucherCount - in.Bundle.PromotionROI.UsedVoucherCount

	var incomeSum float64
	for _, r := range in.Riders {
		incomeSum += float64(r.Income)
	}
	if len(in.Riders) > 0 {
		out.AverageMonthlyIncomeVND = incomeSum / float64(len(in.Riders))
	}

	type riderAgg struct {
		requested, cancelled, completed int
		spendSum                        int64
		kmSum, etaSum, waitSum          float64
	}
	perRider := map[string]*riderAgg{}
	for _, t := range in.Trips {
		if t.RiderID == "" {
			continue
		}
		a := perRider[t.RiderID]
		if a == nil {
			a = &riderAgg{}
			perRider[t.RiderID] = a
		}
		a.requested++
		switch t.Outcome {
		case entity.OutcomeCompleted:
			a.completed++
			a.spendSum += t.FinalFareVND
			a.kmSum += t.DistanceKM
			a.etaSum += t.ETAMinutes
			a.waitSum += t.PickupMinutes
		case entity.OutcomeCancelled:
			a.cancelled++
		}
	}

	var spendSum, kmSum, etaSum, waitSum, cancelPctSum float64
	var completedTotal int
	var ridersWithRequests int
	for _, a := range perRider {
		spendSum += float64(a.spendSum)
		kmSum += a.kmSum
		etaSum += a.etaSum
		waitSum += a.waitSum
		completedTotal += a.completed
		if a.requested > 0 {
			cancelPctSum += 100 * float64(a.cancelled) / float64(a.requested)
			ridersWithRequests++
		}
	}
	if completedTotal > 0 {
		out.AverageSpendVND = spendSum / float64(completedTotal)
		out.AverageFareVND = out.AverageSpendVND
		out.AverageKm = kmSum / float64(completedTotal)
		out.AverageETAMinutes = etaSum / float64(completedTotal)
		out.AverageWaitingMinutes = waitSum / float64(completedTotal)
	}
	if ridersWithRequests > 0 {
		out.AverageCancelPercent = cancelPctSum / float64(ridersWithRequests)
	}
	if in.Days > 0 {
		out.TripsPerMonth = in.Bundle.RiderAnalytics.AverageTrips * (30.0 / float64(in.Days))
	}
	return out
}
