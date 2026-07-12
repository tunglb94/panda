package stats

import "github.com/fairride/ai_simulation/domain/entity"

type DriverStatEntry struct {
	ID           string  `json:"id"`
	AccountType  string  `json:"account_type"`
	Rating       float64 `json:"rating"`
	TotalTrips   int     `json:"total_trips"`
	IncomeTodayVND int64 `json:"income_today_vnd"`
	IncomeWeekVND  int64 `json:"income_week_vnd"`
	Satisfaction float64 `json:"satisfaction"`
	Fatigue      float64 `json:"fatigue"`
	OnlineAtEnd  bool    `json:"online_at_end"`
}

type DriverStatistics struct {
	TotalDrivers      int               `json:"total_drivers"`
	OnlineAtEnd       int               `json:"online_at_end"`
	OfflineAtEnd      int               `json:"offline_at_end"`
	AverageRating     float64           `json:"average_rating"`
	AverageIncomeVND  float64           `json:"average_income_week_vnd"`
	AverageFatigue    float64           `json:"average_fatigue"`
	AccountTypeCounts map[string]int    `json:"account_type_counts"`
	Drivers           []DriverStatEntry `json:"drivers"`
}

func (c *Collector) BuildDriverStatistics(drivers map[string]*entity.DriverAgent) DriverStatistics {
	out := DriverStatistics{AccountTypeCounts: map[string]int{}}
	var ratingSum, incomeSum, fatigueSum float64
	for _, d := range drivers {
		out.TotalDrivers++
		if d.Online {
			out.OnlineAtEnd++
		} else {
			out.OfflineAtEnd++
		}
		ratingSum += d.Rating
		incomeSum += float64(d.IncomeWeek)
		fatigueSum += d.Fatigue
		out.AccountTypeCounts[string(d.AccountType)]++
		out.Drivers = append(out.Drivers, DriverStatEntry{
			ID: d.ID, AccountType: string(d.AccountType), Rating: d.Rating,
			TotalTrips: d.TotalTrips, IncomeTodayVND: d.IncomeToday, IncomeWeekVND: d.IncomeWeek,
			Satisfaction: d.Satisfaction, Fatigue: d.Fatigue, OnlineAtEnd: d.Online,
		})
	}
	out.AverageRating = avg(ratingSum, out.TotalDrivers)
	out.AverageIncomeVND = avg(incomeSum, out.TotalDrivers)
	out.AverageFatigue = avg(fatigueSum, out.TotalDrivers)
	return out
}
