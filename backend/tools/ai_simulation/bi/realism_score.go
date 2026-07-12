package bi

// RealismScore is PHẦN 20 — a self-assessment of how faithfully each
// subsystem models the real Panda business, 0-100. This is inherently a
// qualitative judgment (there is no real Panda operational dataset to
// compare simulated output against), so each score is computed from 2
// concrete, countable facts rather than picked arbitrarily:
//
//  1. EngineReuseScore (0-60): does this subsystem call the REAL production
//     engine (Pricing/Promotion/Dispatch/Delivery state machine), a new-but-
//     BRB-sourced implementation (Driver Economy), or pure simulation-design
//     logic with no real system behind it (demand generation, driver
//     online/offline behavior, weather/traffic probabilities)?
//  2. AssumptionPenalty (0-40, subtracted): how many disclosed ASSUMPTIONs
//     does this subsystem carry in this run's own reports? More disclosed
//     unknowns -> lower confidence the numbers reflect real-world behavior.
//
// The exact weights (60/40 split, -4 points per assumption) are this
// package's own scoring convention — not a BRB or industry-standard
// realism metric, since none exists to adopt.
type RealismCategory struct {
	Category         string `json:"category"`
	Score            int    `json:"score"`
	EngineReuseBasis string `json:"engine_reuse_basis"`
	AssumptionCount  int    `json:"assumption_count"`
}

type RealismScoreReport struct {
	Categories     []RealismCategory `json:"categories"`
	OverallScore   int               `json:"overall_score"`
	Methodology    string            `json:"methodology"`
}

// engineReuseBaseline is each category's EngineReuseScore component (see
// doc comment above) — set once here, not per-run, since it reflects a
// structural property of the code (does this category call a real engine),
// not something that varies by simulation output.
var engineReuseBaseline = map[string]struct {
	score int
	basis string
}{
	"pricing":    {60, "Gọi trực tiếp FareCalculator/Dynamic Pricing Engine thật (BRB §2.2.1-§2.2.13), không viết lại công thức."},
	"promotion":  {60, "Gọi trực tiếp Promotion Engine thật (backend/services/promotion), voucher/campaign BRB-sourced."},
	"dispatch":   {60, "Gọi trực tiếp RequestDispatchUseCase/AcceptTripUseCase thật, thuật toán offerNextDriver không đổi."},
	"delivery":   {60, "Chạy qua đúng state machine Delivery thật (backend/services/trip) — Created→Accepted→...→Completed."},
	"driver_economy": {35, "Code MỚI (không có trong production) nhưng số liệu lấy từ BRB §7.1/§7.2 công bố — không phải suy đoán tự do."},
	"demand":     {10, "Công thức sinh nhu cầu (baseProb, demand multiplier theo giờ/scenario) là thiết kế mô phỏng thuần tuý, không hiệu chỉnh theo dữ liệu Panda thật."},
	"supply":     {10, "Xác suất online/offline, tryStartShift là hằng số thiết kế, không có dữ liệu ca làm việc tài xế thật."},
	"passenger_behavior": {10, "Habit/PriceSensitivity/SwitchApp là mô hình hành vi giả định, không có dữ liệu hành vi khách thật."},
	"traffic":    {10, "3 mức traffic xác suất đơn giản, không dùng dữ liệu GPS/giao thông thật."},
	"weather":    {10, "Xác suất thời tiết cố định, không theo mùa/khí hậu VN thật."},
}

// ComputeRealismScore is PHẦN 20. assumptionCounts maps each category key
// (matching engineReuseBaseline's keys) to how many Assumption entries this
// run's own reports carry for it — passed in by the caller (export_extra.go)
// after every other bi.Compute* function has already run, so the count
// reflects this run's REAL disclosed-assumption tally, not a guess.
func ComputeRealismScore(assumptionCounts map[string]int) RealismScoreReport {
	order := []string{"pricing", "promotion", "dispatch", "delivery", "driver_economy", "demand", "supply", "passenger_behavior", "traffic", "weather"}
	var out RealismScoreReport
	var sum int
	for _, key := range order {
		base := engineReuseBaseline[key]
		n := assumptionCounts[key]
		penalty := n * 4
		if penalty > 40 {
			penalty = 40
		}
		score := base.score + (40 - penalty) // start the "confidence" half at 40, subtract 4 per assumption
		if score > 100 {
			score = 100
		}
		if score < 0 {
			score = 0
		}
		out.Categories = append(out.Categories, RealismCategory{Category: key, Score: score, EngineReuseBasis: base.basis, AssumptionCount: n})
		sum += score
	}
	if len(order) > 0 {
		out.OverallScore = sum / len(order)
	}
	out.Methodology = "Score = EngineReuseBasis (0-60, gọi engine production thật hay không) + (40 - 4*số ASSUMPTION đã công bố cho hạng mục đó, tối thiểu 0). Không phải phép đo khách quan tuyệt đối — là quy ước chấm điểm riêng của package này, không có chuẩn ngành để đối chiếu."
	return out
}
