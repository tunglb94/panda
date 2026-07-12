package bi

// DriverBehavior is PHẦN 8 — every one of the sprint brief's 4 named AI
// decision points, reported as raw counts. "Tắt app" and "Đi nghỉ" are
// reported together (FatigueStop) since FatigueDecision only has one Stop
// outcome — see Assumptions.
type DriverBehavior struct {
	ContinueDrivingCount int `json:"continue_driving_count"` // Tiếp tục chạy
	StopCount            int `json:"stop_count"`             // Tắt app / Đi nghỉ (không phân biệt được — xem Assumptions)
	SurgeChaseCount      int `json:"surge_chase_count"`      // Đổi khu vực theo Surge
	SurgeStayCount       int `json:"surge_stay_count"`
	OfferAcceptedCount   int `json:"offer_accepted_count"` // Nhận chuyến
	OfferRejectedCount   int `json:"offer_rejected_count"` // Từ chối chuyến
	Assumptions          []Assumption `json:"assumptions"`
}

// ComputeDriverBehavior is PHẦN 8 — every count comes straight from
// World's decision-outcome counters (see simulation/world.go's
// recordFatigueOutcome/recordSurgeChaseOutcome/recordDriverOfferOutcome),
// passed in via Input; no new computation, purely a reporting view.
func ComputeDriverBehavior(in Input) DriverBehavior {
	return DriverBehavior{
		ContinueDrivingCount: in.FatigueContinueCount,
		StopCount:            in.FatigueStopCount,
		SurgeChaseCount:      in.SurgeChaseCount,
		SurgeStayCount:       in.SurgeStayCount,
		OfferAcceptedCount:   in.DriverOffersAccepted,
		OfferRejectedCount:   in.DriverOffersRejected,
		Assumptions: []Assumption{
			{Title: "\"Tắt app\" và \"Đi nghỉ\" gộp chung", Detail: "ruleengine.FatigueDecision chỉ có 2 kết quả: Continue hoặc Stop — không phân biệt tài xế chủ động nghỉ ngơi khỏi việc tắt app hẳn. StopCount là tổng cả hai."},
		},
	}
}

// PassengerBehavior is PHẦN 9.
type PassengerBehavior struct {
	SwitchToCompetitorPercent float64 `json:"switch_to_competitor_percent"` // Đổi sang Grab
	StayOnPandaPercent        float64 `json:"stay_on_panda_percent"`
	VoucherUsePercent         float64 `json:"voucher_use_percent"`
	VoucherKeepPercent        float64 `json:"voucher_keep_percent"`

	// CancelledByPriceOrLoyaltyPercent is the ONLY rider-initiated
	// cancellation reason this simulation actually models — see
	// Assumptions for why "Huỷ vì ETA"/"Huỷ vì Driver"/"Đợi thêm" are
	// reported as not_modeled rather than invented percentages.
	CancelledByPriceOrLoyaltyPercent float64      `json:"cancelled_by_price_or_loyalty_percent"`
	NotModeled                       []string     `json:"not_modeled"`
	Assumptions                      []Assumption `json:"assumptions"`
}

// ComputePassengerBehavior is PHẦN 9.
func ComputePassengerBehavior(in Input) PassengerBehavior {
	var out PassengerBehavior
	if total := in.SwitchAppCount + in.StayOnPandaCount; total > 0 {
		out.SwitchToCompetitorPercent = 100 * float64(in.SwitchAppCount) / float64(total)
		out.StayOnPandaPercent = 100 * float64(in.StayOnPandaCount) / float64(total)
	}
	if total := in.VoucherUsedCount + in.VoucherKeptCount; total > 0 {
		out.VoucherUsePercent = 100 * float64(in.VoucherUsedCount) / float64(total)
		out.VoucherKeepPercent = 100 * float64(in.VoucherKeptCount) / float64(total)
	}
	if req := in.Bundle.SimulationReport.Dispatch.Requested; req > 0 {
		out.CancelledByPriceOrLoyaltyPercent = 100 * float64(in.Bundle.SimulationReport.Dispatch.Cancelled) / float64(req)
	}
	out.NotModeled = []string{"huỷ vì ETA", "huỷ vì Driver", "đợi thêm (chờ lâu hơn thay vì huỷ)"}
	out.Assumptions = []Assumption{
		{Title: "Chỉ có 1 lý do huỷ chủ động của khách", Detail: "SwitchAppDecision (giá + lòng trung thành) là quyết định huỷ DUY NHẤT ride_flow.go/delivery_flow.go mô hình hoá. Không có logic riêng cho \"huỷ vì ETA cao\" hay \"huỷ vì driver\" — báo cáo trung thực 0/not_modeled thay vì suy diễn tỉ lệ giả."},
	}
	return out
}
