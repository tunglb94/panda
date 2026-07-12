package bi

import (
	"sort"

	"github.com/fairride/ai_simulation/domain/entity"
)

// LeaderboardEntry is one driver's row — PHẦN 13.
type LeaderboardEntry struct {
	DriverID              string  `json:"driver_id"`
	IncomeVND             int64   `json:"income_vnd"` // this run's total net income (sum of DriverNetVND across completed trips)
	Trips                 int     `json:"trips"`
	AcceptanceRatePercent float64 `json:"acceptance_rate_percent"`
	Rating                float64 `json:"rating"`
	// FuelEfficiencyKmPer1000VND is km driven per 1,000đ of estimated fuel
	// cost (see driver_economy.go's fuel-cost-per-km assumption) — higher
	// is more efficient. A simulation-design efficiency metric, not a real
	// fuel-economy measurement.
	FuelEfficiencyKmPer1000VND float64 `json:"fuel_efficiency_km_per_1000_vnd"`
	OnlineHours                float64 `json:"online_hours"`
}

type DriverLeaderboard struct {
	Top100      []LeaderboardEntry `json:"top_100"`
	Assumptions []Assumption       `json:"assumptions"`
}

// ComputeDriverLeaderboard is PHẦN 13, ranked by IncomeVND descending.
func ComputeDriverLeaderboard(in Input) DriverLeaderboard {
	type agg struct {
		income int64
		trips  int
		km     float64
	}
	perDriver := map[string]*agg{}
	for _, t := range in.Trips {
		if t.Outcome != entity.OutcomeCompleted || t.DriverID == "" {
			continue
		}
		a := perDriver[t.DriverID]
		if a == nil {
			a = &agg{}
			perDriver[t.DriverID] = a
		}
		a.income += t.DriverNetVND
		a.trips++
		a.km += t.DistanceKM
	}

	var entries []LeaderboardEntry
	for id, d := range in.Drivers {
		a := perDriver[id]
		if a == nil || a.trips == 0 {
			continue
		}
		e := LeaderboardEntry{DriverID: id, IncomeVND: a.income, Trips: a.trips, Rating: d.Rating, OnlineHours: d.TotalHoursOnline}
		if total := d.OffersAccepted + d.OffersRejected; total > 0 {
			e.AcceptanceRatePercent = 100 * float64(d.OffersAccepted) / float64(total)
		}
		fuelCostVND := a.km * fuelCostPerKmFor(d.VehicleType)
		if fuelCostVND > 0 {
			e.FuelEfficiencyKmPer1000VND = 1000 * a.km / fuelCostVND
		}
		entries = append(entries, e)
	}

	sort.Slice(entries, func(i, j int) bool { return entries[i].IncomeVND > entries[j].IncomeVND })
	if len(entries) > 100 {
		entries = entries[:100]
	}
	return DriverLeaderboard{
		Top100: entries,
		Assumptions: []Assumption{
			{Title: "Fuel Efficiency", Detail: "km/1.000đ chi phí nhiên liệu ước tính (xem driver_economy.json's fuel cost assumption) — không phải số liệu tiêu hao nhiên liệu thật, vì simulation không mô hình mức tiêu hao xe cụ thể."},
		},
	}
}

func fuelCostPerKmFor(v entity.DriverVehicleType) float64 {
	switch v {
	case entity.VehicleCar:
		return fuelCostPerKmCarVND
	case entity.VehicleVan:
		return fuelCostPerKmVanVND
	default:
		return fuelCostPerKmMotorcycleVND
	}
}
