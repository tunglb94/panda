package entity

// Traffic is one of the 4 required road conditions.
type Traffic string

const (
	TrafficClear    Traffic = "clear"    // Thông thoáng
	TrafficBusy     Traffic = "busy"     // Đông
	TrafficJammed   Traffic = "jammed"   // Kẹt
	TrafficAccident Traffic = "accident" // Tai nạn
)

// SpeedFactor scales a zone's normal travel speed — used to derive ETA/pickup
// time from distance (see simulation/ride_flow.go). 1.0 = normal speed.
func (t Traffic) SpeedFactor() float64 {
	switch t {
	case TrafficClear:
		return 1.15
	case TrafficBusy:
		return 0.85
	case TrafficJammed:
		return 0.5
	case TrafficAccident:
		return 0.3
	default:
		return 1.0
	}
}
