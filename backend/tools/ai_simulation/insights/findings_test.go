package insights

import (
	"context"
	"strings"
	"testing"

	"github.com/fairride/ai_simulation/stats"
)

func TestComputeFindings_CapsAt20AndRanksBySignal(t *testing.T) {
	bundle := stats.Bundle{
		DriverAnalytics: stats.DriverAnalytics{RetentionRatePercent: 20, AverageOnlineHours: 5}, // low retention -> high signal
		RiderAnalytics:  stats.RiderAnalytics{RetentionRatePercent: 60},                          // healthy -> low signal
	}
	bi := stats.BusinessIntelligence{AcceptanceRatePercent: 50, CancellationRatePercent: 10, PlatformMarginPercent: 18}

	findings := ComputeFindings(nil, bundle, bi)
	if len(findings) == 0 {
		t.Fatalf("expected at least one finding to fire")
	}
	if len(findings) > 20 {
		t.Errorf("expected at most 20 findings, got %d", len(findings))
	}
	for i := 1; i < len(findings); i++ {
		if findings[i].Signal > findings[i-1].Signal {
			t.Errorf("expected findings ranked by descending signal, index %d (%.1f) > index %d (%.1f)", i, findings[i].Signal, i-1, findings[i-1].Signal)
		}
	}
	// Low driver retention (20%) is explicitly boosted to a high signal —
	// it should be the top (or near-top) finding.
	if !strings.Contains(findings[0].Text, "giữ chân") && !strings.Contains(findings[0].Text, "chấp nhận") {
		t.Errorf("expected the top finding to be one of the flagged-critical metrics, got %q", findings[0].Text)
	}
}

func TestComputeFindings_NoDataProducesNoFindings(t *testing.T) {
	findings := ComputeFindings(nil, stats.Bundle{}, stats.BusinessIntelligence{})
	if len(findings) != 0 {
		t.Errorf("expected zero findings from an all-zero bundle, got %d: %+v", len(findings), findings)
	}
}

func TestComputeRecommendations_SinglePromotionTypeIsNotSelfContradictory(t *testing.T) {
	// Regression test: when best-ROI and worst-ROI promotion are the same
	// entry (only one type redeemed), the recommendation set must not
	// contain both "expand budget" and "cut budget" for it.
	bundle := stats.Bundle{
		PromotionROI: stats.PromotionROI{
			ByType: []stats.PromotionROIEntry{
				{Type: "manual_coupon", RedeemedCount: 14, TotalCostVND: 77_006, GMVGeneratedVND: 770_117, ROI: 9.0, CPAVND: 5500, RepeatRatePercent: 92},
			},
		},
	}
	recs := ComputeRecommendations(nil, bundle, stats.BusinessIntelligence{})

	var expandFound, cutFound bool
	for _, r := range recs {
		if strings.Contains(r.Text, "Mở rộng ngân sách") && strings.Contains(r.Text, "manual_coupon") {
			expandFound = true
		}
		if strings.Contains(r.Text, "manual_coupon") == false {
			continue
		}
		if strings.Contains(r.Text, "giảm ngân sách") || strings.Contains(r.Text, "thắt chặt") {
			cutFound = true
		}
	}
	if expandFound && cutFound {
		t.Fatalf("expected only one of expand/cut recommendations for the same single promotion type, got both:\n%+v", recs)
	}
}

func TestComputeRecommendations_CapsAt30(t *testing.T) {
	bundle := stats.Bundle{
		DriverStatistics: stats.DriverStatistics{TotalDrivers: 100, AccountTypeCounts: map[string]int{"bronze": 90}},
		DriverAnalytics:  stats.DriverAnalytics{RetentionRatePercent: 10, AverageOnlineHours: 1},
		RiderAnalytics:   stats.RiderAnalytics{RetentionRatePercent: 5, MembershipCounts: map[string]int{"free": 90, "gold": 10}},
	}
	bi := stats.BusinessIntelligence{
		AcceptanceRatePercent: 40, CancellationRatePercent: 20, AverageETAMinutes: 30, PlatformMarginPercent: 5,
	}
	recs := ComputeRecommendations(nil, bundle, bi)
	if len(recs) > 30 {
		t.Errorf("expected at most 30 recommendations, got %d", len(recs))
	}
	for _, r := range recs {
		if r.Priority != "High" && r.Priority != "Medium" && r.Priority != "Low" {
			t.Errorf("unexpected priority value %q", r.Priority)
		}
		if r.ExpectedImpact == "" || r.Risk == "" {
			t.Errorf("expected every recommendation to have ExpectedImpact and Risk set, got %+v", r)
		}
	}
}

func TestRenderSummaryMarkdown_NotesWhenFewerThan20(t *testing.T) {
	md := RenderSummaryMarkdown([]Finding{{Text: "chỉ một phát hiện", Signal: 1}})
	if !strings.Contains(md, "chỉ 1 phát hiện") {
		t.Errorf("expected the under-20 note to mention the actual count, got:\n%s", md)
	}
	if !strings.Contains(md, "1. chỉ một phát hiện") {
		t.Errorf("expected the finding to be numbered, got:\n%s", md)
	}
}

// fakeReportEngine lets the test control WriteSummary/WriteRecommendations'
// AI path without a real Ollama server.
type fakeReportEngine struct {
	text string
	ok   bool
}

func (f fakeReportEngine) GenerateReport(_ context.Context, _ string) (string, bool) {
	return f.text, f.ok
}

func TestWriteSummary_FallsBackWhenAIUnavailable(t *testing.T) {
	findings := []Finding{{Text: "Zone X có nhu cầu cao nhất.", Signal: 10}}
	got := WriteSummary(context.Background(), fakeReportEngine{ok: false}, findings)
	want := RenderSummaryMarkdown(findings)
	if got != want {
		t.Errorf("expected the deterministic fallback when AI is unavailable, got a different result")
	}
}

func TestWriteSummary_FallsBackOnSuspiciouslyShortAIOutput(t *testing.T) {
	findings := make([]Finding, 5)
	for i := range findings {
		findings[i] = Finding{Text: "một phát hiện dài với nhiều số liệu thật", Signal: float64(10 - i)}
	}
	// AI "succeeds" but returns something implausibly short for 5 findings.
	got := WriteSummary(context.Background(), fakeReportEngine{text: "ok", ok: true}, findings)
	want := RenderSummaryMarkdown(findings)
	if got != want {
		t.Errorf("expected fallback to the deterministic version when AI output looks too short/suspicious")
	}
}

func TestWriteSummary_UsesAIOutputWhenPlausible(t *testing.T) {
	findings := []Finding{{Text: "Zone X có nhu cầu cao nhất với 100 yêu cầu.", Signal: 10}}
	plausible := strings.Repeat("Đây là bản tóm tắt điều hành hợp lý với đầy đủ số liệu thật. ", 3)
	got := WriteSummary(context.Background(), fakeReportEngine{text: plausible, ok: true}, findings)
	if !strings.Contains(got, plausible) {
		t.Errorf("expected the AI-polished text to be used when it passes the plausibility check, got:\n%s", got)
	}
	if !strings.Contains(got, "AI-assisted") {
		t.Errorf("expected the AI-assisted header, got:\n%s", got)
	}
}
