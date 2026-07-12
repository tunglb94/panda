package audit

import "github.com/fairride/ai_simulation/stats"

// BugFinding is one documented defect — per this task's explicit
// instruction, ComputeReport/this whole package never fixes anything; a
// BugFinding is reported in exactly this shape (Bug/Cause/Impact/Repro/File)
// and left for a human to act on.
type BugFinding struct {
	Title        string
	Cause        string
	Impact       string
	Reproduction string
	File         string
}

// Assumption is one place this audit found existing logic that is a
// deliberate simulation-design choice or an open question — not a bug, but
// worth a human's attention before trusting the numbers derived from it.
type Assumption struct {
	Title  string
	Detail string
}

// DetectBugs inspects the validation report for anything this run's
// self-check caught. As of this audit, the commission-scaling bug found in
// the previous session's testing (see CHANGELOG "commission double-counted
// the booking fee on surged trips") is already fixed — DetectBugs will
// report nothing for it unless a regression reintroduces the same signature
// (a commission conservation warning), in which case it surfaces here
// rather than being silently re-fixed, per this task's explicit
// "KHÔNG tự sửa" instruction.
func DetectBugs(validation stats.ValidationReport) []BugFinding {
	bugs := append([]BugFinding{}, StructuralBugs()...)
	for _, w := range validation.Warnings {
		if w.Severity != "critical" {
			continue
		}
		bugs = append(bugs, BugFinding{
			Title:        "Validation critical: " + w.Check,
			Cause:        w.Message,
			Impact:       "Vi phạm bất biến tài chính cơ bản (" + w.Check + ") — số liệu GMV/commission/driver income trong các báo cáo khác có thể không đáng tin cậy cho đến khi khắc phục.",
			Reproduction: "Chạy `go run ./backend/tools/ai_simulation` với cùng cấu hình và seed, kiểm tra validation_report.json.",
			File:         "backend/tools/ai_simulation/stats/validation.go (self-check); nguồn dữ liệu thực tế nằm ở simulation/ride_flow.go và simulation/delivery_flow.go",
		})
	}
	return bugs
}

// StructuralBugs are defects found by direct investigation during this
// audit (not derivable from a single run's validation_report.json) —
// reported here every time, since they are properties of the code, not of
// one run's random draws. Per the task's explicit "KHÔNG tự sửa" (do not
// fix) instruction, none of these has been patched.
func StructuralBugs() []BugFinding {
	return []BugFinding{
		{
			Title: "--seed không tạo ra kết quả xác định (deterministic) giữa các lần chạy",
			Cause: "World.Drivers/World.Riders là map[string]*entity.DriverAgent/RiderAgent — Go cố ý ngẫu nhiên hoá thứ tự duyệt map giữa mỗi lần chạy chương trình. Mọi vòng lặp `for _, d := range w.Drivers` / `for _, r := range w.Riders` (simulation/engine.go's processTick/evaluateDriverState, onNewDay, RefreshZoneCounters...) tiêu thụ số ngẫu nhiên từ *rand.Rand đã seed theo đúng thứ tự đó — thứ tự duyệt khác nhau → chuỗi số ngẫu nhiên được rút ra khác nhau → kết quả khác nhau, dù --seed giống hệt nhau.",
			Impact: "Đã tự kiểm chứng trực tiếp: chạy `--seed=777` hai lần với cấu hình giống hệt nhau (30 driver/200 rider/2 ngày) cho ra 666 vs 586 requested (~12% lệch), GMV lệch ~20%. Điều này có nghĩa: (1) không có lần chạy nào trong báo cáo audit này có thể tái hiện chính xác byte-for-byte chỉ bằng cách chạy lại cùng --seed; (2) --seed hiện tại chỉ hữu ích để tạo dữ liệu ngẫu nhiên có kiểm soát ở mức phân phối thống kê, không phải để debug/regression-test một kết quả cụ thể. Không ảnh hưởng đến tính hợp lệ của các kết luận kinh doanh (vẫn là dữ liệu thật từ một lần chạy thật), nhưng ảnh hưởng đến khả năng tái lập số liệu chính xác.",
			Reproduction: "go run ./backend/tools/ai_simulation --drivers=30 --riders=200 --days=2 --seed=777 --out=/tmp/a && go run ./backend/tools/ai_simulation --drivers=30 --riders=200 --days=2 --seed=777 --out=/tmp/b && diff /tmp/a/simulation_report.json /tmp/b/simulation_report.json",
			File: "backend/tools/ai_simulation/simulation/world.go (Drivers/Riders map fields), engine.go (mọi range trên 2 map này), seed.go",
		},
	}
}

// KnownAssumptions catalogs every simulation-design assumption this audit's
// numbers depend on — compiled from doc comments already in the codebase
// (not re-derived), so a reader doesn't have to go file-hunting to know
// what to distrust vs. what's a real BRB-sourced number.
func KnownAssumptions() []Assumption {
	return []Assumption{
		{Title: "VAT 10%", Detail: "BRB tự loại VAT khỏi phạm vi (\"calculated and remitted by the Finance team independently\", business-rule-bible-v1.0.md dòng 986) — 10% là mức chuẩn thuế Việt Nam, không phải số BRB."},
		{Title: "Chi phí Cloud/Map/SMS mỗi chuyến (400đ/250đ/350đ)", Detail: "Giả định thiết kế, không có dữ liệu tài chính thật của Panda để đối chiếu — xem stats/unit_economics.go."},
		{Title: "Motorcycle fare = 60% giá Car", Detail: "BRB không định nghĩa rate cho xe máy — giả định tỉ lệ điển hình thị trường Đông Nam Á, không phải số BRB — xem integration/pricing_adapter.go."},
		{Title: "\"Bike Plus\" không tồn tại trong production", Detail: "pricing_analytics.json báo cáo trung thực not_modeled=true thay vì bịa công thức giá — Panda chưa từng xây dựng hạng xe máy cao cấp."},
		{Title: "Dispatch \"bias\" phản ánh cung tài xế, không phải thuật toán", Detail: "offerNextDriver chỉ ghép theo khoảng cách gần nhất — chênh lệch accept-rate theo khu vực (§11) là artifact của phân bố cung, không phải lỗi thiên vị trong code Dispatch."},
		{Title: "Ngưỡng \"voucher spam\" (>5 lần/lần chạy)", Detail: "Ngưỡng thiết kế mô phỏng, không phải số liệu chống gian lận thật — không có dữ liệu lịch sử để hiệu chỉnh."},
		{Title: "Ngưỡng driver income outlier (mean + 3×stddev)", Detail: "Ngưỡng thống kê tiêu chuẩn, không phải chính sách BRB nào."},
		{Title: "Zone/thời tiết/traffic là hằng số thiết kế mô phỏng", Detail: "Toạ độ khu vực, base demand weight, xác suất thời tiết/traffic không đến từ dữ liệu thành phố thật — xem domain/entity/city.go, simulation/scenario_scheduler.go."},
		{Title: "Delivery fare dùng chung Pricing Engine với Ride", Detail: "Production chưa tích hợp Pricing cho Delivery (xác nhận qua backend/services/trip/app/complete_delivery.go's doc comment) — simulation ước tính fare Delivery bằng cùng công thức Ride, giống đúng tiền lệ Rider app's DeliveryFormPage đã làm."},
		{Title: "Peak Hour = 07:00-09:00 và 17:00-20:00", Detail: "Định nghĩa thiết kế mô phỏng dùng cho §17/§18 — không phải khung giờ BRB chính thức nào (BRB có Peak Hour Surcharge riêng nhưng không định nghĩa khung giờ cố định theo cách này)."},
		{Title: "Hai \"Acceptance Rate\" khác nhau dùng cùng một tên", Detail: "driver_analytics.json's acceptance_rate_percent đo tỉ lệ MỘT tài xế đồng ý khi được đề nghị một chuyến cụ thể (offersAccepted/(offersAccepted+offersRejected), driverAcceptsOffer trong ride_flow.go); CEO_report.html/executive_dashboard.html's \"Acceptance Rate\" đo tỉ lệ MỘT yêu cầu cuối cùng có tài xế nhận (completed/requested, sau khi đã thử nhiều tài xế qua vòng lặp resolveOffer). Cả hai đều tính đúng theo định nghĩa riêng, nhưng tên gọi giống nhau dễ gây nhầm lẫn khi đối chiếu hai file — quan sát được trong lần chạy full-scale: 81.4% (per-offer) vs 99.8% (per-request)."},
		{Title: "Demand/Supply ratio ở scale 1000/10000 không nên đọc như tỉ lệ dân số", Detail: "Ở cấu hình 1000 tài xế/10.000 khách (tỉ lệ dân số 1:10), heatmap ghi nhận cầu/cung TỨC THỜI (theo tick) ở mức đồng đều 42-44x tại MỌI khu vực — kể cả các khu vực được xếp \"thừa driver\" trong §14 (so sánh tương đối, không phải oversupply thật, ratio luôn >>1). Vẫn accept rate 99.8% nhờ cơ chế retry nhiều tài xế gần nhất, không phải vì có đủ cung tại một thời điểm — cần đọc con số 42-44x như \"áp lực cầu tức thời cao\", không phải \"khách phải chờ 42x lâu hơn bình thường\"."},
	}
}
