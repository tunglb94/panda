package stats

import "github.com/fairride/ai_simulation/domain/entity"

// RunConfig echoes the CLI parameters a report was generated from, so a
// saved simulation_report.json is self-describing.
type RunConfig struct {
	Drivers int    `json:"drivers"`
	Riders  int    `json:"riders"`
	Days    int    `json:"days"`
	Model   string `json:"model"`
}

type FinancialSummary struct {
	PlatformRevenueVND int64 `json:"platform_revenue_vnd"` // total commission + booking fees
	DriverRevenueVND   int64 `json:"driver_revenue_vnd"`   // total driver net income
	PassengerSavingVND int64 `json:"passenger_saving_vnd"` // total promotion + voucher discounts
	CommissionVND      int64 `json:"commission_vnd"`
	PromotionCostVND   int64 `json:"promotion_cost_vnd"`
	VoucherCostVND     int64 `json:"voucher_cost_vnd"`
	GMVVND             int64 `json:"gmv_vnd"` // gross fare collected from riders, pre-discount
}

type SimulationReport struct {
	Config    RunConfig             `json:"config"`
	Dispatch  DispatchOutcomeCounts `json:"dispatch"`
	Financial FinancialSummary      `json:"financial"`
}

func (c *Collector) BuildSimulationReport(cfg RunConfig, trips []*entity.SimTrip) SimulationReport {
	report := SimulationReport{Config: cfg, Dispatch: c.DispatchOutcomes(trips)}

	var fin FinancialSummary
	for _, t := range trips {
		if t.Outcome != entity.OutcomeCompleted {
			continue
		}
		fin.GMVVND += t.FinalFareVND + t.VoucherDiscountVND
		fin.CommissionVND += t.CommissionVND
		fin.DriverRevenueVND += t.DriverNetVND
		fin.PassengerSavingVND += t.VoucherDiscountVND
		if t.PromotionType == "manual_coupon" {
			fin.VoucherCostVND += t.VoucherDiscountVND
		} else if t.PromotionType != "" {
			fin.PromotionCostVND += t.VoucherDiscountVND
		}
	}
	fin.PlatformRevenueVND = fin.CommissionVND
	report.Financial = fin
	return report
}
