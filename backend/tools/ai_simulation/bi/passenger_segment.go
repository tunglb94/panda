package bi

import "github.com/fairride/ai_simulation/domain/entity"

// SegmentStat is one passenger segment's aggregate figures — PHẦN 14.
type SegmentStat struct {
	Segment          string  `json:"segment"`
	RiderCount       int     `json:"rider_count"`
	AverageTrips     float64 `json:"average_trips"`
	AverageSpendVND  float64 `json:"average_spend_vnd"`
	AverageVoucherUsedCount float64 `json:"average_voucher_used_count"`
	RetentionPercent float64 `json:"retention_percent"`
}

type PassengerSegmentReport struct {
	Segments    []SegmentStat `json:"segments"`
	Assumptions []Assumption  `json:"assumptions"`
}

// classifySegment assigns exactly one segment per rider by a fixed
// priority order (a rider matching several criteria — e.g. both high
// membership and high price sensitivity — lands in whichever is checked
// first) — RiderAgent has no real "is this person a tourist" signal, so
// segments are approximated from the fields that actually exist
// (Membership/PriceSensitivity/Habit/Income/TripCount), documented in
// Assumptions rather than presented as if directly measured.
func classifySegment(r *entity.RiderAgent, avgTripCount float64) string {
	switch {
	case r.Membership == entity.MembershipDiamond:
		return "VIP"
	case avgTripCount > 0 && float64(r.TripCount) > 2*avgTripCount:
		return "Heavy User"
	case r.PriceSensitivity > 0.75:
		return "Cheap"
	case r.Habit == entity.HabitCommuter:
		return "Office Worker"
	case r.Income < 8_000_000:
		return "Student"
	case r.Habit == entity.HabitBusinessTraveler:
		return "Tourist/Business" // closest proxy — see Assumptions
	default:
		return "Normal"
	}
}

// ComputePassengerSegments is PHẦN 14.
func ComputePassengerSegments(in Input) PassengerSegmentReport {
	var totalTrips int
	for _, r := range in.Riders {
		totalTrips += r.TripCount
	}
	var avgTripCount float64
	if len(in.Riders) > 0 {
		avgTripCount = float64(totalTrips) / float64(len(in.Riders))
	}

	spendByRider := map[string]int64{}
	voucherByRider := map[string]int{}
	daysActiveByRider := map[string]map[int64]bool{}
	for _, t := range in.Trips {
		if t.RiderID == "" {
			continue
		}
		if daysActiveByRider[t.RiderID] == nil {
			daysActiveByRider[t.RiderID] = map[int64]bool{}
		}
		daysActiveByRider[t.RiderID][requestedDay(t)] = true
		if t.Outcome != entity.OutcomeCompleted {
			continue
		}
		spendByRider[t.RiderID] += t.FinalFareVND
		if t.PromotionType == "manual_coupon" {
			voucherByRider[t.RiderID]++
		}
	}

	type agg struct {
		riders                  []string
		tripsSum, voucherSum    int
		spendSum                int64
		retentionSum            float64
	}
	bySegment := map[string]*agg{}
	for id, r := range in.Riders {
		seg := classifySegment(r, avgTripCount)
		a := bySegment[seg]
		if a == nil {
			a = &agg{}
			bySegment[seg] = a
		}
		a.riders = append(a.riders, id)
		a.tripsSum += r.TripCount
		a.voucherSum += voucherByRider[id]
		a.spendSum += spendByRider[id]
		if in.Days > 0 {
			a.retentionSum += 100 * float64(len(daysActiveByRider[id])) / float64(in.Days)
		}
	}

	var out PassengerSegmentReport
	for seg, a := range bySegment {
		n := float64(len(a.riders))
		out.Segments = append(out.Segments, SegmentStat{
			Segment: seg, RiderCount: len(a.riders),
			AverageTrips: float64(a.tripsSum) / n, AverageSpendVND: float64(a.spendSum) / n,
			AverageVoucherUsedCount: float64(a.voucherSum) / n, RetentionPercent: a.retentionSum / n,
		})
	}
	out.Assumptions = []Assumption{
		{Title: "Phân khúc là suy luận, không phải nhãn thật", Detail: "RiderAgent không có trường \"segment\"/\"is_tourist\" — VIP=Membership Diamond, Heavy User=TripCount>2x trung bình, Cheap=PriceSensitivity>0.75, Office Worker=Habit Commuter, Student=Income<8tr VND/tháng, Tourist/Business=Habit BusinessTraveler (proxy gần nhất, không phân biệt được khách du lịch thật với khách công tác). Mỗi rider chỉ thuộc 1 phân khúc theo thứ tự ưu tiên trên."},
	}
	return out
}
