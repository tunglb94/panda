package stats

import "github.com/fairride/ai_simulation/domain/entity"

// PricingCategoryStat is one row of pricing_analytics.json.
type PricingCategoryStat struct {
	Category               string  `json:"category"`
	TripCount              int     `json:"trip_count"`
	AverageFareVND         float64 `json:"average_fare_vnd"`
	AverageCommissionVND   float64 `json:"average_commission_vnd"`
	AverageDriverIncomeVND float64 `json:"average_driver_income_vnd"`
	AverageProfitVND       float64 `json:"average_profit_vnd"`
	MarketSavingVND        float64 `json:"average_market_saving_vnd"` // avg voucher+promotion discount
}

type PricingAnalytics struct {
	Categories []PricingCategoryStat `json:"categories"`
}

// allServiceTypes and allTripKinds are iterated together, not hardcoded as
// a flat category list — Vehicle/Service Catalog refactor Part 7 ("được
// tạo bằng tổ hợp TripType + ServiceType, không hardcode"). Every one of
// the 2x4=8 combinations (Ride Bike, Ride Bike Plus, Ride Car, Ride Car
// XL, Delivery Bike, Delivery Bike Plus, Delivery Car, Delivery Car XL)
// now has a real simulated rate (integration.NewPricingAdapter's own rate
// table covers all 4 ServiceTypes), so none of them are permanently
// zero/NotModeled the way Bike Plus/Car XL were before that rate table
// existed.
var allServiceTypes = []entity.ServiceType{
	entity.ServiceBike, entity.ServiceBikePlus, entity.ServiceCar, entity.ServiceCarXL,
}

var allTripKinds = []entity.TripKind{entity.KindRide, entity.KindDelivery}

func serviceTypeLabel(s entity.ServiceType) string {
	switch s {
	case entity.ServiceBike:
		return "Bike"
	case entity.ServiceBikePlus:
		return "Bike Plus"
	case entity.ServiceCar:
		return "Car"
	case entity.ServiceCarXL:
		return "Car XL"
	default:
		return string(s)
	}
}

func tripKindLabel(k entity.TripKind) string {
	switch k {
	case entity.KindDelivery:
		return "Delivery"
	default:
		return "Ride"
	}
}

func (c *Collector) BuildPricingAnalytics(trips []*entity.SimTrip) PricingAnalytics {
	var out []PricingCategoryStat
	for _, kind := range allTripKinds {
		for _, st := range allServiceTypes {
			label := tripKindLabel(kind) + " " + serviceTypeLabel(st)
			out = append(out, buildPricingCategoryStat(trips, label, kind, st))
		}
	}
	return PricingAnalytics{Categories: out}
}

func buildPricingCategoryStat(trips []*entity.SimTrip, label string, kind entity.TripKind, st entity.ServiceType) PricingCategoryStat {
	stat := PricingCategoryStat{Category: label}
	var fareSum, commissionSum, driverSum, savingSum float64
	for _, t := range trips {
		if t.Outcome != entity.OutcomeCompleted || t.ServiceType != st || t.Kind != kind {
			continue
		}
		stat.TripCount++
		fareSum += float64(t.FinalFareVND)
		commissionSum += float64(t.CommissionVND)
		driverSum += float64(t.DriverNetVND)
		savingSum += float64(t.VoucherDiscountVND)
	}
	stat.AverageFareVND = avg(fareSum, stat.TripCount)
	stat.AverageCommissionVND = avg(commissionSum, stat.TripCount)
	stat.AverageDriverIncomeVND = avg(driverSum, stat.TripCount)
	stat.MarketSavingVND = avg(savingSum, stat.TripCount)
	stat.AverageProfitVND = estimatedPlatformProfitVND(stat.AverageCommissionVND, stat.TripCount > 0)
	return stat
}
