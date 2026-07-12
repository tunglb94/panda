package simulation

// This file implements BƯỚC 4 (Competitive Simulation). Per the sprint
// brief: "Không cần chính xác tuyệt đối. Dùng khoảng giá thị trường" — every
// number below is a directional market-range ESTIMATE derived from the
// qualitative competitor analysis already done in docs/business/
// PRICING_STRATEGY.md §0 (published commission ranges, public positioning),
// not scraped real-time prices from any competitor's API. Treat this as
// "roughly how much more/less would a rider likely pay, and would a driver
// likely earn, on each platform for the same trip" — not a guarantee.

// CompetitorProfile approximates one competitor's pricing behaviour as two
// ratios applied to Panda's own computed fare for the identical trip:
//
//   theirCustomerTotal = pandaCustomerTotal × PriceMultiplier
//   theirDriverIncome  = theirCustomerTotal × (1 − CommissionRate)
//
// This is deliberately simple (BRB §1.2 Principle 4: "Simplicity Scales") —
// a full per-competitor fare formula would need data Panda does not have
// access to, and would create false precision that a "khoảng giá thị
// trường" comparison explicitly does not need.
type CompetitorProfile struct {
	Name            string
	PriceMultiplier float64 // vs Panda's CustomerTotal for the same trip; 1.0 = same price
	CommissionRate  float64 // their approximate take rate, for driver-income comparison
	Note            string  // sourced from PRICING_STRATEGY.md §0 — what this estimate is grounded in
}

// CompetitorProfiles — the 8 platforms BƯỚC 4 names, in the same order.
// Every PriceMultiplier/CommissionRate below traces to a specific
// characterisation already written in PRICING_STRATEGY.md §0.2.
func CompetitorProfiles() []CompetitorProfile {
	return []CompetitorProfile{
		{
			Name: "Grab", PriceMultiplier: 1.15, CommissionRate: 0.25,
			Note: "Hoa hồng thuộc nhóm cao nhất khu vực; surge thuật toán phức tạp, thường đắt hơn vào giờ cao điểm (PRICING_STRATEGY §0.2)",
		},
		{
			Name: "Be", PriceMultiplier: 1.00, CommissionRate: 0.20,
			Note: "Định vị thương hiệu Việt, giá cạnh tranh gần mức thị trường trung bình (PRICING_STRATEGY §0.2)",
		},
		{
			Name: "XanhSM", PriceMultiplier: 1.20, CommissionRate: 0.15,
			Note: "Mô hình tài xế-nhân viên (không phải hoa hồng đối tác thật) — CommissionRate ở đây là ước lượng cho MỤC ĐÍCH SO SÁNH thu nhập, không phản ánh cơ cấu lương thật; giá cao hơn do định vị xe điện cao cấp đồng nhất (PRICING_STRATEGY §0.2)",
		},
		{
			Name: "Maxim", PriceMultiplier: 0.75, CommissionRate: 0.15,
			Note: "Định vị giá rẻ nhất, vận hành tinh gọn, phục vụ thị trường ngách (PRICING_STRATEGY §0.2)",
		},
		{
			Name: "inDrive", PriceMultiplier: 0.80, CommissionRate: 0.10,
			Note: "Mô hình đấu giá ngược — hoa hồng công bố cực thấp; giá thực tế do rider đề xuất nên thường thấp hơn mặt bằng chung (PRICING_STRATEGY §0.2)",
		},
		{
			Name: "Bolt", PriceMultiplier: 0.90, CommissionRate: 0.17,
			Note: "Chiến lược thách thức giá rẻ có kiểm soát, hoa hồng thấp hơn Uber/Grab (PRICING_STRATEGY §0.2)",
		},
		{
			Name: "Uber", PriceMultiplier: 1.20, CommissionRate: 0.27,
			Note: "Upfront pricing + surge mạnh, hoa hồng thuộc nhóm cao nhất ngành toàn cầu (PRICING_STRATEGY §0.2)",
		},
		{
			Name: "Lyft", PriceMultiplier: 1.10, CommissionRate: 0.25,
			Note: "Theo sát giá Uber ở thị trường Mỹ, thường rẻ hơn Uber một chút (PRICING_STRATEGY §0.2)",
		},
	}
}

// CompetitiveRow is one line of the BƯỚC-4 comparison table for a single
// trip against a single competitor estimate.
type CompetitiveRow struct {
	Competitor           string
	Note                 string
	EstimatedCustomerTotal int64
	EstimatedDriverIncome  int64
	CustomerDeltaVsPanda   int64 // positive = competitor is more expensive than Panda
	DriverDeltaVsPanda     int64 // positive = competitor driver earns more than Panda driver
}

// CompareToMarket runs BƯỚC 4 for one already-computed Panda fare: how would
// each of the 8 competitors' estimated price/driver-income compare for the
// identical trip.
func CompareToMarket(panda *FareBreakdown) []CompetitiveRow {
	rows := make([]CompetitiveRow, 0, len(CompetitorProfiles()))
	for _, c := range CompetitorProfiles() {
		theirTotal := roundVND(float64(panda.CustomerTotal) * c.PriceMultiplier)
		theirDriver := roundVND(float64(theirTotal) * (1 - c.CommissionRate))
		rows = append(rows, CompetitiveRow{
			Competitor:             c.Name,
			Note:                   c.Note,
			EstimatedCustomerTotal: theirTotal,
			EstimatedDriverIncome:  theirDriver,
			CustomerDeltaVsPanda:   theirTotal - panda.CustomerTotal,
			DriverDeltaVsPanda:     theirDriver - panda.NetDriver,
		})
	}
	return rows
}

// MarketPosition summarises BƯỚC 5's three ordered objectives for one trip:
// is Panda cheaper for the rider than the market, and does the Panda driver
// earn more than the market, simultaneously?
type MarketPosition struct {
	CheaperThanMarketCount int // how many of the 8 competitors Panda beats on price (lower CustomerTotal)
	DriverEarnsMoreCount   int // how many of the 8 competitors Panda's driver beats on income
	TotalCompetitors       int
}

func EvaluateMarketPosition(panda *FareBreakdown, rows []CompetitiveRow) MarketPosition {
	mp := MarketPosition{TotalCompetitors: len(rows)}
	for _, r := range rows {
		if panda.CustomerTotal <= r.EstimatedCustomerTotal {
			mp.CheaperThanMarketCount++
		}
		if panda.NetDriver >= r.EstimatedDriverIncome {
			mp.DriverEarnsMoreCount++
		}
	}
	return mp
}
