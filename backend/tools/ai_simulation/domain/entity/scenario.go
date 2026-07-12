package entity

// ScenarioKind is one of the required simulated situations. Multiple can be
// active at once (e.g. Rain + EveningRush) — see simulation/scenario_scheduler.go
// for how active scenarios combine to influence demand/supply.
type ScenarioKind string

const (
	ScenarioMorningRush   ScenarioKind = "morning_rush"   // Giờ cao điểm sáng
	ScenarioLunch         ScenarioKind = "lunch"          // Buổi trưa
	ScenarioEveningRush   ScenarioKind = "evening_rush"   // Tan ca
	ScenarioWeekend       ScenarioKind = "weekend"        // Cuối tuần
	ScenarioHoliday       ScenarioKind = "holiday"        // Lễ
	ScenarioRain          ScenarioKind = "rain"           // Mưa
	ScenarioConcert       ScenarioKind = "concert"        // Sự kiện: Concert
	ScenarioFestival      ScenarioKind = "festival"       // Sự kiện: Festival
	ScenarioAccident      ScenarioKind = "accident"       // Tai nạn
)

// ActiveScenarios is the set of scenarios in effect at a given tick.
type ActiveScenarios map[ScenarioKind]bool

func (a ActiveScenarios) Has(k ScenarioKind) bool { return a[k] }
