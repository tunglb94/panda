package insights

import (
	"fmt"
	"strings"
)

// RenderSummaryMarkdown is the deterministic fallback for
// simulation_summary.md — always available even with Ollama fully down,
// since every Finding's text was already fully rendered by ComputeFindings.
func RenderSummaryMarkdown(findings []Finding) string {
	var b strings.Builder
	b.WriteString("# Panda Simulation — Tổng kết phát hiện quan trọng\n\n")
	if len(findings) < 20 {
		fmt.Fprintf(&b, "_Lưu ý: chỉ %d phát hiện đủ điều kiện nổi bật trong lần chạy này (ngưỡng tối đa là 20) — số liệu nhỏ hơn phản ánh quy mô/độ dài của lần chạy, không phải giới hạn nhân tạo._\n\n", len(findings))
	}
	for i, f := range findings {
		fmt.Fprintf(&b, "%d. %s\n", i+1, f.Text)
	}
	b.WriteString("\n---\n_Toàn bộ số liệu trên được tính trực tiếp từ dữ liệu mô phỏng thực tế (simulation_report.json và các file thống kê liên quan) — không có số liệu nào được suy đoán hay bịa đặt._\n")
	return b.String()
}

// RenderRecommendationsMarkdown is business_recommendation.md's deterministic
// fallback — same guarantee as RenderSummaryMarkdown.
func RenderRecommendationsMarkdown(recs []Recommendation) string {
	var b strings.Builder
	b.WriteString("# Panda Simulation — Đề xuất cải thiện kinh doanh\n\n")
	if len(recs) < 30 {
		fmt.Fprintf(&b, "_Lưu ý: chỉ %d đề xuất đủ điều kiện nổi bật trong lần chạy này (ngưỡng tối đa là 30)._\n\n", len(recs))
	}
	for i, r := range recs {
		fmt.Fprintf(&b, "## %d. %s\n\n", i+1, r.Text)
		fmt.Fprintf(&b, "- **Priority:** %s\n- **Expected Impact:** %s\n- **Risk:** %s\n\n", r.Priority, r.ExpectedImpact, r.Risk)
	}
	b.WriteString("---\n_Mỗi đề xuất được sinh ra từ một điều kiện thật đo được trong dữ liệu mô phỏng (xem simulation_summary.md và các file *_analytics.json/*.json liên quan); Priority/Expected Impact/Risk là nhận định nghiệp vụ đi kèm, không phải số liệu đo được._\n")
	return b.String()
}
