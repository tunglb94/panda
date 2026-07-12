package bi

import "sort"

// Distribution is the Median/Mean/P95/Min/Max view PHẦN 4 asks every
// pricing breakdown to carry — shared by any section needing the same
// 5-number summary over a set of real observed values.
type Distribution struct {
	Count  int     `json:"count"`
	Mean   float64 `json:"mean"`
	Median float64 `json:"median"`
	P95    float64 `json:"p95"`
	Min    float64 `json:"min"`
	Max    float64 `json:"max"`
}

// ComputeDistribution sorts values (a copy — never mutates the caller's
// slice) and derives all 5 figures in one pass.
func ComputeDistribution(values []float64) Distribution {
	if len(values) == 0 {
		return Distribution{}
	}
	sorted := append([]float64(nil), values...)
	sort.Float64s(sorted)

	var sum float64
	for _, v := range sorted {
		sum += v
	}
	n := len(sorted)
	return Distribution{
		Count:  n,
		Mean:   sum / float64(n),
		Median: percentile(sorted, 0.50),
		P95:    percentile(sorted, 0.95),
		Min:    sorted[0],
		Max:    sorted[n-1],
	}
}

// percentile uses nearest-rank on an already-sorted slice — simple, no
// interpolation, adequate for a business-reporting percentile rather than a
// statistically rigorous one.
func percentile(sorted []float64, p float64) float64 {
	if len(sorted) == 0 {
		return 0
	}
	idx := int(p * float64(len(sorted)-1))
	if idx < 0 {
		idx = 0
	}
	if idx >= len(sorted) {
		idx = len(sorted) - 1
	}
	return sorted[idx]
}
