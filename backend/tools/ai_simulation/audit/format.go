package audit

import "fmt"

// formatVND adds thousands separators to a VND amount for readable
// report/anomaly text.
func formatVND(v int64) string {
	s := fmt.Sprintf("%d", v)
	neg := len(s) > 0 && s[0] == '-'
	if neg {
		s = s[1:]
	}
	var out []byte
	for i, c := range []byte(s) {
		if i > 0 && (len(s)-i)%3 == 0 {
			out = append(out, ',')
		}
		out = append(out, c)
	}
	if neg {
		return "-" + string(out)
	}
	return string(out)
}

func formatFloat(v float64) string {
	return fmt.Sprintf("%.2f", v)
}
