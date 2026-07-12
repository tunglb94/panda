package bi

import (
	"fmt"

	"github.com/fairride/ai_simulation/domain/entity"
	"github.com/fairride/ai_simulation/insights"
)

// ComputeAdditionalRecommendations adds PHẦN 18's own named examples
// ("Tăng 5% driver quanh Airport", "Giảm voucher khu CBD", "Tăng surge
// threshold cuối tuần", "Giảm commission Bike", "Tăng Delivery bonus") as
// concrete rule candidates gated on real data this bi package computed —
// merged with insights.ComputeRecommendations' existing catalog by the
// caller (export_extra.go), not a replacement for it. Reuses
// insights.Recommendation directly rather than a parallel type.
func ComputeAdditionalRecommendations(in Input, driverEconomy DriverEconomyReport, pricing PricingBreakdown, delivery DeliveryDashboard) []insights.Recommendation {
	var recs []insights.Recommendation
	add := func(sev float64, priority, impact, risk, format string, args ...any) {
		if sev <= 0 {
			return
		}
		recs = append(recs, insights.Recommendation{
			Text: fmt.Sprintf(format, args...), ExpectedImpact: impact, Risk: risk, Priority: priority, Signal: sev,
		})
	}

	if gap := airportDemandSupplyGap(in); gap > 5 {
		add(gap, "High", "Giảm ETA và tăng acceptance rate tại Sân bay", "Chi phí incentive tăng",
			"Tăng 5%% tài xế hoạt động quanh khu vực Sân bay — tỉ lệ cầu/cung hiện tại %.1fx.", gap)
	}

	if cbdVoucher := voucherCostForZone(in, entity.ZoneCBD); cbdVoucher > 0 && in.BI.VoucherCostVND > 0 {
		share := 100 * float64(cbdVoucher) / float64(in.BI.VoucherCostVND)
		if share > 30 {
			add(share, "Medium", "Giảm chi phí voucher, tăng lợi nhuận ròng", "Có thể giảm chuyển đổi ở CBD nếu cắt quá mạnh",
				"Giảm ngân sách voucher khu Trung tâm (CBD) — khu vực này chiếm %.1f%% tổng chi phí voucher.", share)
		}
	}

	if weekendDist, ok := pricing.ByCalendar["weekend"]; ok && weekendDist.Count > 0 {
		if normalDist, ok2 := pricing.ByCalendar["normal"]; ok2 && normalDist.Mean > 0 && weekendDist.Mean < normalDist.Mean*1.1 {
			add(20, "Low", "Tăng doanh thu giờ cao điểm cuối tuần", "Có thể giảm nhu cầu nếu surge quá cao",
				"Xem xét tăng ngưỡng surge cuối tuần — giá/km cuối tuần (%.0fđ) chỉ nhỉnh hơn ngày thường (%.0fđ) không đáng kể dù nhu cầu cuối tuần thường cao hơn.", weekendDist.Mean, normalDist.Mean)
		}
	}

	if bikeProfit, carProfit := profitPerKmByVehicle(in); bikeProfit > 0 && carProfit > 0 && bikeProfit < carProfit*0.5 {
		add(30, "Low", "Cân bằng lợi nhuận giữa các hạng xe", "Giảm doanh thu hoa hồng từ Bike ngắn hạn",
			"Xem xét giảm commission cho Bike — lợi nhuận/km hiện tại (%s VND) chưa bằng nửa Car (%s VND).", formatVNDBi(int64(bikeProfit)), formatVNDBi(int64(carProfit)))
	}

	if delivery.DeliveryCount > 0 && in.BI.DeliveryPercent > 0 && in.BI.DeliveryPercent < 30 {
		add(30-in.BI.DeliveryPercent, "Low", "Tăng khối lượng Delivery, đa dạng hoá doanh thu", "Chi phí thưởng tăng ngắn hạn",
			"Tăng thưởng (bonus) cho tài xế nhận đơn Delivery — hiện chỉ chiếm %.1f%% tổng số chuyến dù lợi nhuận/chuyến Delivery (%s VND) không thua kém Ride.", in.BI.DeliveryPercent, formatVNDBi(int64(delivery.DeliveryProfitVND)))
	}

	for _, seg := range driverEconomy.Segments {
		if seg.DriverCount > 0 && seg.NetIncomePerDayVND > 0 && seg.NetIncomePerDayVND < livingWagePerDayVND {
			add(30, "High", "Tăng thu nhập tài xế nhóm thu nhập thấp", "Giảm biên lợi nhuận nền tảng nếu tăng ưu đãi",
				"Xem xét ưu đãi thu nhập tối thiểu cho nhóm %s — thu nhập ròng/ngày hiện tại (%s VND) dưới mức tham chiếu sinh hoạt.", seg.Category, formatVNDBi(int64(seg.NetIncomePerDayVND)))
		}
	}

	return recs
}

func voucherCostForZone(in Input, zone entity.ZoneType) int64 {
	var sum int64
	for _, t := range in.Trips {
		if t.Outcome == entity.OutcomeCompleted && t.PickupZone == zone && t.PromotionType == "manual_coupon" {
			sum += t.VoucherDiscountVND
		}
	}
	return sum
}

func profitPerKmByVehicle(in Input) (bikePerKm, carPerKm float64) {
	var bikeProfit, carProfit, bikeKm, carKm float64
	for _, t := range in.Trips {
		if t.Outcome != entity.OutcomeCompleted || t.Kind != entity.KindRide || t.DistanceKM <= 0 {
			continue
		}
		p := float64(profitFor(t))
		switch t.ServiceType {
		case entity.ServiceBike, entity.ServiceBikePlus:
			bikeProfit += p
			bikeKm += t.DistanceKM
		case entity.ServiceCar, entity.ServiceCarXL:
			carProfit += p
			carKm += t.DistanceKM
		}
	}
	if bikeKm > 0 {
		bikePerKm = bikeProfit / bikeKm
	}
	if carKm > 0 {
		carPerKm = carProfit / carKm
	}
	return bikePerKm, carPerKm
}
