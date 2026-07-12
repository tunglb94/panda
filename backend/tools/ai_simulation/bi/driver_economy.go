package bi

import "github.com/fairride/ai_simulation/domain/entity"

// Per-km operating cost assumptions (Fuel Cost, Phone Cost) — PHẦN 1 asks
// for both but no such cost model exists anywhere in this simulation or in
// BRB (BRB's Part 7 covers commission, not a driver's own vehicle/phone
// running costs). Set at plausible Vietnamese market rates for a
// motorcycle/car respectively (fuel ~1,200-2,200đ/km depending on vehicle,
// mobile data ~1,500đ/online-hour) — simulation-design ASSUMPTIONS, not
// measured or BRB-sourced, exactly like unit_economics.go's Cloud/Map/SMS
// cost assumptions this mirrors.
const (
	fuelCostPerKmMotorcycleVND = 700
	fuelCostPerKmCarVND        = 2200
	fuelCostPerKmVanVND        = 2800
	phoneCostPerOnlineHourVND  = 1500
)

// ShiftCategory is PHẦN 1's driver self-classification by average daily
// online hours.
type ShiftCategory string

const (
	ShiftPartTime ShiftCategory = "part_time" // <4h/ngày
	ShiftRegular  ShiftCategory = "regular"   // 4-8h/ngày
	ShiftFullTime ShiftCategory = "full_time" // 8-10h/ngày
	ShiftHardcore ShiftCategory = "hardcore"  // >10h/ngày
)

func classifyShift(avgHoursPerDay float64) ShiftCategory {
	switch {
	case avgHoursPerDay < 4:
		return ShiftPartTime
	case avgHoursPerDay < 8:
		return ShiftRegular
	case avgHoursPerDay <= 10:
		return ShiftFullTime
	default:
		return ShiftHardcore
	}
}

// DriverEconomySegment is one shift category's aggregate figures.
type DriverEconomySegment struct {
	Category            string  `json:"category"`
	DriverCount         int     `json:"driver_count"`
	AverageIncomePerDayVND   float64 `json:"average_income_per_day_vnd"`
	AverageIncomePerMonthVND float64 `json:"average_income_per_month_vnd"` // per-day x 30, a projection not a measured monthly figure
	TripsPerDay          float64 `json:"trips_per_day"`
	KmPerDay             float64 `json:"km_per_day"`
	OnlineHoursPerDay    float64 `json:"online_hours_per_day"`
	DrivingHoursPerDay   float64 `json:"driving_hours_per_day"`
	WaitingHoursPerDay   float64 `json:"waiting_hours_per_day"` // online but not actively driving a trip
	IdlePercent          float64 `json:"idle_percent"`
	FuelCostPerDayVND    float64 `json:"fuel_cost_per_day_vnd"`
	PhoneCostPerDayVND   float64 `json:"phone_cost_per_day_vnd"`
	NetIncomePerDayVND   float64 `json:"net_income_per_day_vnd"`   // after platform commission (driver's actual take-home)
	GrossIncomePerDayVND float64 `json:"gross_income_per_day_vnd"` // before platform commission (what rider paid for the metered portion)
	ROI                  float64 `json:"roi"`                      // net income / (fuel+phone cost) — return on the driver's own operating outlay
	AcceptanceRatePercent float64 `json:"acceptance_rate_percent"`
	CancelRatePercent    float64 `json:"cancel_rate_percent"`
	AverageRating        float64 `json:"average_rating"`
	AverageFatigue       float64 `json:"average_fatigue"`
	RetentionPercent     float64 `json:"retention_percent"` // days active / total simulated days
}

type DriverEconomyReport struct {
	Segments    []DriverEconomySegment `json:"segments"`
	Assumptions []Assumption            `json:"assumptions"`
}

type driverAgg struct {
	trips, cancelled, assigned int
	kmSum, drivingMinSum       float64
	netSum, grossSum           int64
}

// ComputeDriverEconomy is PHẦN 1 — every number is either read directly off
// DriverAgent (TotalHoursOnline/DaysActive/Fatigue/Rating/OffersAccepted+
// Rejected) or aggregated from that driver's own completed trips (income,
// km, driving time) — no invented data.
func ComputeDriverEconomy(in Input) DriverEconomyReport {
	perDriver := map[string]*driverAgg{}
	get := func(id string) *driverAgg {
		a := perDriver[id]
		if a == nil {
			a = &driverAgg{}
			perDriver[id] = a
		}
		return a
	}
	for _, t := range in.Trips {
		if t.DriverID == "" {
			continue
		}
		a := get(t.DriverID)
		a.assigned++
		switch t.Outcome {
		case entity.OutcomeCompleted:
			a.trips++
			a.kmSum += t.DistanceKM
			a.drivingMinSum += drivingMinutes(t)
			a.netSum += t.DriverNetVND
			a.grossSum += t.FinalFareVND
		case entity.OutcomeCancelled:
			a.cancelled++
		}
	}

	buckets := map[ShiftCategory]*struct {
		drivers []*entity.DriverAgent
		agg     []*driverAgg
	}{}
	for id, d := range in.Drivers {
		if d.DaysActive == 0 {
			continue // never went online this run — excluded from shift classification, not fabricated into a bucket
		}
		avgHours := d.TotalHoursOnline / float64(d.DaysActive)
		cat := classifyShift(avgHours)
		b := buckets[cat]
		if b == nil {
			b = &struct {
				drivers []*entity.DriverAgent
				agg     []*driverAgg
			}{}
			buckets[cat] = b
		}
		a := perDriver[id]
		if a == nil {
			a = &driverAgg{}
		}
		b.drivers = append(b.drivers, d)
		b.agg = append(b.agg, a)
	}

	var segments []DriverEconomySegment
	for _, cat := range []ShiftCategory{ShiftPartTime, ShiftRegular, ShiftFullTime, ShiftHardcore} {
		b := buckets[cat]
		if b == nil || len(b.drivers) == 0 {
			segments = append(segments, DriverEconomySegment{Category: string(cat)})
			continue
		}
		segments = append(segments, buildSegment(string(cat), b.drivers, b.agg, in.Days))
	}

	return DriverEconomyReport{
		Segments: segments,
		Assumptions: []Assumption{
			{Title: "Fuel cost/km", Detail: "Motorcycle 700đ, Car 2,200đ, Van 2,800đ — giả định thị trường VN, không đến từ BRB hay dữ liệu tài chính thật của Panda."},
			{Title: "Phone cost/giờ online", Detail: "1,500đ/giờ — giả định chi phí data di động, không đo lường thật."},
			{Title: "ROI công thức", Detail: "Net Income / (Fuel Cost + Phone Cost) trong cùng kỳ — định nghĩa tự chọn cho mô phỏng này, không phải công thức ROI chuẩn ngành."},
			{Title: "Tài xế chưa từng online bị loại khỏi phân loại", Detail: "Không xếp vào bucket nào (không có giờ online để tính trung bình/ngày) — tránh chia cho 0 hoặc bịa số liệu."},
		},
	}
}

func buildSegment(category string, drivers []*entity.DriverAgent, aggs []*driverAgg, totalDays int) DriverEconomySegment {
	seg := DriverEconomySegment{Category: category, DriverCount: len(drivers)}
	var onlineHoursSum, ratingSum, fatigueSum, retentionSum float64
	var acceptedSum, rejectedSum, tripsSum, cancelledSum, assignedSum int
	var kmSum, drivingMinSum float64
	var netSum, grossSum int64

	for i, d := range drivers {
		days := float64(d.DaysActive)
		onlineHoursSum += d.TotalHoursOnline / days
		ratingSum += d.Rating
		fatigueSum += d.Fatigue
		if totalDays > 0 {
			retentionSum += 100 * days / float64(totalDays)
		}
		acceptedSum += d.OffersAccepted
		rejectedSum += d.OffersRejected

		a := aggs[i]
		tripsSum += a.trips
		cancelledSum += a.cancelled
		assignedSum += a.assigned
		kmSum += a.kmSum
		drivingMinSum += a.drivingMinSum
		netSum += a.netSum
		grossSum += a.grossSum
	}

	n := float64(len(drivers))
	seg.OnlineHoursPerDay = onlineHoursSum / n
	seg.AverageRating = ratingSum / n
	seg.AverageFatigue = fatigueSum / n
	seg.RetentionPercent = retentionSum / n

	// Per-driver-day denominators — sum each driver's own DaysActive so a
	// mixed-tenure segment (some drivers active 3 days, others 30) isn't
	// skewed by treating everyone as if they had the same number of days.
	var totalDriverDays float64
	for _, d := range drivers {
		totalDriverDays += float64(d.DaysActive)
	}
	if totalDriverDays == 0 {
		return seg
	}

	seg.TripsPerDay = float64(tripsSum) / totalDriverDays
	seg.KmPerDay = kmSum / totalDriverDays
	seg.DrivingHoursPerDay = (drivingMinSum / 60) / totalDriverDays
	if seg.DrivingHoursPerDay > seg.OnlineHoursPerDay {
		seg.DrivingHoursPerDay = seg.OnlineHoursPerDay // driving-time estimate is a lower bound; never let rounding push it above online hours
	}
	seg.WaitingHoursPerDay = seg.OnlineHoursPerDay - seg.DrivingHoursPerDay
	if seg.OnlineHoursPerDay > 0 {
		seg.IdlePercent = 100 * seg.WaitingHoursPerDay / seg.OnlineHoursPerDay
	}

	fuelCostPerKm := blendedFuelCostPerKm(drivers)
	seg.FuelCostPerDayVND = seg.KmPerDay * fuelCostPerKm
	seg.PhoneCostPerDayVND = seg.OnlineHoursPerDay * phoneCostPerOnlineHourVND

	seg.NetIncomePerDayVND = float64(netSum) / totalDriverDays
	seg.GrossIncomePerDayVND = float64(grossSum) / totalDriverDays
	seg.AverageIncomePerDayVND = seg.NetIncomePerDayVND
	seg.AverageIncomePerMonthVND = seg.NetIncomePerDayVND * 30

	if opCost := seg.FuelCostPerDayVND + seg.PhoneCostPerDayVND; opCost > 0 {
		seg.ROI = seg.NetIncomePerDayVND / opCost
	}

	if total := acceptedSum + rejectedSum; total > 0 {
		seg.AcceptanceRatePercent = 100 * float64(acceptedSum) / float64(total)
	}
	if assignedSum > 0 {
		seg.CancelRatePercent = 100 * float64(cancelledSum) / float64(assignedSum)
	}
	return seg
}

func blendedFuelCostPerKm(drivers []*entity.DriverAgent) float64 {
	var sum float64
	for _, d := range drivers {
		switch d.VehicleType {
		case entity.VehicleCar:
			sum += fuelCostPerKmCarVND
		case entity.VehicleVan:
			sum += fuelCostPerKmVanVND
		default:
			sum += fuelCostPerKmMotorcycleVND
		}
	}
	if len(drivers) == 0 {
		return 0
	}
	return sum / float64(len(drivers))
}
