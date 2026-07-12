package app

import "github.com/fairride/pricing/domain/entity"

// NewDefaultPricingPipelineV3 builds the same 9-rule pipeline as
// NewDefaultPricingPipeline (pricing_pipeline.go — completely unmodified by
// this file), except AirportFeeRule is replaced with AirportFeeRuleV3
// (rules_airport_v3.go). Every other rule (Demand Surge, Supply Surge, Peak
// Hour, Night, Holiday, Rain, Traffic, Special Event) is reused unchanged —
// per the sprint brief PHẦN 5 ("Các rule TODO vẫn giữ TODO. Không tự thêm
// rule."), V3 adds no new surge rule beyond the airport fee swap.
func NewDefaultPricingPipelineV3(airportConfig entity.AirportFeeConfigV3, vehicleType entity.VehicleType) *PricingPipeline {
	return NewPricingPipeline([]PricingRule{
		NewDemandSurgeRule(entity.DefaultDSRTiers()),
		NewSupplySurgeRule(),
		NewPeakHourRule(entity.DefaultPeakHourWindows()),
		NewNightSurchargeRule(entity.NightWindowStartHour, entity.NightWindowEndHour),
		NewHolidaySurchargeRule(),
		NewRainSurchargeRule(),
		NewAirportFeeRuleV3(airportConfig, vehicleType),
		NewTrafficSurgeRule(),
		NewSpecialEventRule(),
	})
}
