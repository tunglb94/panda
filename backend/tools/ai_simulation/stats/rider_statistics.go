package stats

import "github.com/fairride/ai_simulation/domain/entity"

type RiderStatEntry struct {
	ID             string `json:"id"`
	Membership     string `json:"membership"`
	TripCount      int    `json:"trip_count"`
	HasActiveVoucher bool `json:"has_active_voucher"`
}

type RiderStatistics struct {
	TotalRiders        int              `json:"total_riders"`
	AverageTripCount   float64          `json:"average_trip_count"`
	MembershipCounts   map[string]int   `json:"membership_counts"`
	Riders             []RiderStatEntry `json:"riders"`
}

func (c *Collector) BuildRiderStatistics(riders map[string]*entity.RiderAgent) RiderStatistics {
	out := RiderStatistics{MembershipCounts: map[string]int{}}
	var tripSum float64
	for _, r := range riders {
		out.TotalRiders++
		tripSum += float64(r.TripCount)
		out.MembershipCounts[string(r.Membership)]++
		out.Riders = append(out.Riders, RiderStatEntry{
			ID: r.ID, Membership: string(r.Membership), TripCount: r.TripCount,
			HasActiveVoucher: r.HasActiveVoucher,
		})
	}
	out.AverageTripCount = avg(tripSum, out.TotalRiders)
	return out
}
