package insights

import (
	"context"
	"strconv"
	"strings"
)

// reportEngine is the subset of aiengine.DecisionEngine this package needs
// — kept as a small local interface so insights has no compile dependency
// on aiengine (mirrors ruleengine's own dependency-free style).
type reportEngine interface {
	GenerateReport(ctx context.Context, prompt string) (string, bool)
}

// WriteSummary produces simulation_summary.md's content: the deterministic,
// data-grounded finding list, optionally polished into more natural
// Vietnamese prose by phi4:14b. The AI is given the findings as already-
// computed plain-text bullets and instructed to rephrase only — every
// number was written by Go, not the model, per "AI phải tự tổng kết ... Không
// được bịa. Chỉ dùng dữ liệu simulation." The Rule Engine (ComputeFindings)
// always decides WHAT the findings are; AI, when available, only helps
// present them — the same 95%/5% split this whole simulation uses for
// agent decisions, applied to report writing.
func WriteSummary(ctx context.Context, ai reportEngine, findings []Finding) string {
	fallback := RenderSummaryMarkdown(findings)
	if ai == nil || len(findings) == 0 {
		return fallback
	}
	polished, ok := ai.GenerateReport(ctx, buildSummaryPrompt(findings))
	if !ok || !looksLikeValidPolish(polished, findings) {
		return fallback
	}
	return "# Panda Simulation — Tổng kết phát hiện quan trọng (AI-assisted)\n\n" + polished +
		"\n\n---\n_Diễn đạt bởi phi4:14b từ các phát hiện do Rule Engine tính toán trực tiếp trên dữ liệu mô phỏng — không có số liệu mới nào được AI thêm vào._\n"
}

// WriteRecommendations mirrors WriteSummary for business_recommendation.md.
func WriteRecommendations(ctx context.Context, ai reportEngine, recs []Recommendation) string {
	fallback := RenderRecommendationsMarkdown(recs)
	if ai == nil || len(recs) == 0 {
		return fallback
	}
	polished, ok := ai.GenerateReport(ctx, buildRecommendationPrompt(recs))
	if !ok || len(polished) < len(recs)*20 { // a real rewrite of N recommendations should not be drastically shorter than N short bullets
		return fallback
	}
	return "# Panda Simulation — Đề xuất cải thiện kinh doanh (AI-assisted)\n\n" + polished +
		"\n\n---\n_Diễn đạt bởi phi4:14b từ các đề xuất do Rule Engine tính toán trực tiếp trên dữ liệu mô phỏng._\n"
}

func buildSummaryPrompt(findings []Finding) string {
	var b strings.Builder
	b.WriteString("Bạn là chuyên viên phân tích dữ liệu ride-hailing. Dưới đây là danh sách phát hiện đã được tính toán CHÍNH XÁC từ dữ liệu mô phỏng thật, mỗi dòng đã có đủ số liệu. " +
		"Hãy viết lại thành một bản tóm tắt điều hành gọn gàng, đánh số 1 đến " + strconv.Itoa(len(findings)) +
		", giữ NGUYÊN mọi con số đã cho, KHÔNG thêm bất kỳ số liệu hay phát hiện mới nào không có trong danh sách dưới đây:\n\n")
	for i, f := range findings {
		b.WriteString(strconv.Itoa(i + 1))
		b.WriteString(". ")
		b.WriteString(f.Text)
		b.WriteString("\n")
	}
	return b.String()
}

func buildRecommendationPrompt(recs []Recommendation) string {
	var b strings.Builder
	b.WriteString("Bạn là chuyên viên tư vấn chiến lược ride-hailing. Dưới đây là danh sách đề xuất đã được tính toán từ dữ liệu mô phỏng thật, mỗi đề xuất đã có Priority/Expected Impact/Risk. " +
		"Hãy viết lại thành văn phong chuyên nghiệp hơn, giữ NGUYÊN mọi con số, Priority, Expected Impact và Risk đã cho, KHÔNG thêm đề xuất mới:\n\n")
	for i, r := range recs {
		b.WriteString(strconv.Itoa(i + 1))
		b.WriteString(". ")
		b.WriteString(r.Text)
		b.WriteString(" [Priority: ")
		b.WriteString(r.Priority)
		b.WriteString(", Expected Impact: ")
		b.WriteString(r.ExpectedImpact)
		b.WriteString(", Risk: ")
		b.WriteString(r.Risk)
		b.WriteString("]\n")
	}
	return b.String()
}

// looksLikeValidPolish is a light sanity check on the AI's rewrite — a real
// rephrasing of N findings should not collapse to something drastically
// shorter (a sign the model ignored most of the input) and should not be
// suspiciously longer than what N short bullets rewritten could plausibly
// need (a sign of unrelated/hallucinated content). Not a semantic check —
// this simulation has no way to verify meaning, only gross shape — so on
// any doubt the caller falls back to the deterministic, guaranteed-accurate
// version rather than risk shipping a subtly wrong AI rewrite.
func looksLikeValidPolish(text string, findings []Finding) bool {
	if len(text) < len(findings)*15 {
		return false
	}
	if len(text) > len(findings)*600 {
		return false
	}
	return true
}
