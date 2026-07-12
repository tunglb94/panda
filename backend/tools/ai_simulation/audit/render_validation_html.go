package audit

import (
	"fmt"
	"strings"

	"github.com/fairride/ai_simulation/stats"
)

// RenderValidationHTML is validation_report.json rendered as a readable,
// self-contained HTML page (no server, no external assets except the
// Chart.js CDN script other dashboards in this tool already use — omitted
// here since a warning list needs no chart).
func RenderValidationHTML(v stats.ValidationReport) string {
	var b strings.Builder
	b.WriteString(`<!doctype html><html><head><meta charset="utf-8"><title>Panda Validation Report</title><style>
body{font-family:system-ui,sans-serif;background:#0b0f0d;color:#e5e7eb;margin:0;padding:24px}
h1{color:#1A8C4E} .status{font-size:1.4rem;font-weight:700;padding:12px 20px;border-radius:10px;display:inline-block;margin-bottom:20px}
.status.pass{background:rgba(26,140,78,0.2);color:#3FCB85} .status.fail{background:rgba(220,38,38,0.2);color:#F87171}
table{border-collapse:collapse;width:100%} th,td{text-align:left;padding:10px 14px;border-bottom:1px solid #22302a}
th{color:#9CA3AF;font-size:0.78rem;text-transform:uppercase} .sev{padding:2px 10px;border-radius:99px;font-size:0.72rem;font-weight:700;text-transform:uppercase}
.sev.critical{background:rgba(220,38,38,0.2);color:#F87171} .sev.warning{background:rgba(245,158,11,0.2);color:#FBBF24} .sev.info{background:rgba(37,99,235,0.2);color:#93C5FD}
</style></head><body>
<h1>Panda — Validation Report</h1>`)

	if v.Passed {
		b.WriteString(`<div class="status pass">✓ PASSED — không có critical warning</div>`)
	} else {
		b.WriteString(`<div class="status fail">✗ FAILED — có critical warning</div>`)
	}

	if len(v.Warnings) == 0 {
		b.WriteString(`<p>Không có warning nào.</p>`)
	} else {
		b.WriteString(`<table><thead><tr><th>Check</th><th>Severity</th><th>Message</th></tr></thead><tbody>`)
		for _, w := range v.Warnings {
			fmt.Fprintf(&b, `<tr><td>%s</td><td><span class="sev %s">%s</span></td><td>%s</td></tr>`, w.Check, w.Severity, w.Severity, htmlEscape(w.Message))
		}
		b.WriteString(`</tbody></table>`)
	}
	b.WriteString(`</body></html>`)
	return b.String()
}

func htmlEscape(s string) string {
	r := strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;")
	return r.Replace(s)
}
