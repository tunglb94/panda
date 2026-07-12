// Package bi turns the AI Digital Twin Simulation into full Business
// Intelligence reporting: every function here only READS data the
// simulation already produced (SimTrip records, DriverAgent/RiderAgent
// final state, stats.Bundle) — none of it changes simulation behavior, and
// none of it duplicates a computation stats/insights/audit already do
// (those are reused directly where a section overlaps). Per the sprint
// brief's explicit "Không dùng fake data nếu đã có dữ liệu thật. Nếu thiếu
// dữ liệu thì ghi rõ ASSUMPTION" — any metric the simulation genuinely has
// no data for is reported as zero/not-modeled with an explicit ASSUMPTION
// note, never invented.
package bi

import (
	"github.com/fairride/ai_simulation/domain/entity"
	"github.com/fairride/ai_simulation/stats"
)

// profitFor is the shared per-trip profit helper every bi section uses —
// wraps stats.PerTripProfitVND (VAT + estimated infra cost, see
// stats/unit_economics.go's own doc comment for the exact assumptions) so
// no section in this package invents its own profit formula.
func profitFor(t *entity.SimTrip) int64 {
	return stats.PerTripProfitVND(t.CommissionVND)
}

// Input bundles everything every Compute* function in this package needs —
// avoids each one taking a dozen parameters. Built once per run by
// simulation/export_extra.go (the only caller with access to World's
// unexported decision-outcome counters).
type Input struct {
	Trips   []*entity.SimTrip
	Drivers map[string]*entity.DriverAgent
	Riders  map[string]*entity.RiderAgent
	Bundle  stats.Bundle
	BI      stats.BusinessIntelligence
	Days    int

	FatigueContinueCount int
	FatigueStopCount     int
	SwitchAppCount       int // rider switched to the simulated competitor
	StayOnPandaCount     int
	SurgeChaseCount      int
	SurgeStayCount       int
	DriverOffersAccepted int
	DriverOffersRejected int
	VoucherUsedCount     int
	VoucherKeptCount     int
}

// Assumption mirrors audit.Assumption's shape (kept as its own type rather
// than importing audit, so bi has no dependency on the audit package —
// they are siblings, not layered on each other).
type Assumption struct {
	Title  string
	Detail string
}

// driving/waiting minutes derivation shared by several sections — a trip's
// stored ETAMinutes is PickupMinutes + actual driving/transit duration (see
// domain/entity/trip.go's doc comment), so the driving duration itself is
// always ETAMinutes-PickupMinutes; not stored as its own field to avoid a
// second source of truth.
func drivingMinutes(t *entity.SimTrip) float64 {
	d := t.ETAMinutes - t.PickupMinutes
	if d < 0 {
		return 0
	}
	return d
}

// requestedHour/requestedDay/isWeekendTrip/isHolidayTrip derive time-of-day
// and calendar facts purely from RequestedAtTick — engine.go's NewEngine
// fixes the simulated start date to a Monday specifically so this
// derivation is possible without carrying a *entity.SimClock reference
// around (see that constructor's own doc comment), and
// scenario_scheduler.go's RollDailyConditions defines "every 10th day is a
// holiday" — both reproduced here read-only, not redefined.
func requestedHour(t *entity.SimTrip) int {
	return int((t.RequestedAtTick % (24 * 60)) / 60)
}

func requestedDay(t *entity.SimTrip) int64 {
	return t.RequestedAtTick / (24 * 60)
}

func isWeekendTrip(t *entity.SimTrip) bool {
	dow := requestedDay(t) % 7 // 0=Monday (see engine.go's NewEngine start-date comment)
	return dow == 5 || dow == 6
}

func isHolidayTrip(t *entity.SimTrip) bool {
	return requestedDay(t)%10 == 9
}

func daypart(hour int) string {
	switch {
	case hour >= 5 && hour < 11:
		return "morning" // Sáng
	case hour >= 11 && hour < 14:
		return "noon" // Trưa
	case hour >= 14 && hour < 18:
		return "afternoon" // Chiều
	default:
		return "night" // Đêm
	}
}
