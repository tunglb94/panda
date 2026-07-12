// Package entity holds the pure data model for the AI Digital Twin
// Simulation: city geography, agents (Driver/Rider), and environmental
// state (weather/traffic/scenario). No business logic lives here — see
// ruleengine, aiengine, integration, and simulation for behavior.
package entity

import "math"

// ZoneType is one of the 8 required city districts.
type ZoneType string

const (
	ZoneCBD           ZoneType = "cbd"           // Central Business District
	ZoneResidential   ZoneType = "residential"   // Khu dân cư
	ZoneIndustrial    ZoneType = "industrial"    // KCN
	ZoneAirport       ZoneType = "airport"       // Sân bay
	ZoneBusStation    ZoneType = "bus_station"   // Bến xe
	ZoneHospital      ZoneType = "hospital"      // Bệnh viện
	ZoneSchool        ZoneType = "school"        // Trường học
	ZoneEntertainment ZoneType = "entertainment" // Khu vui chơi
)

// AllZoneTypes lists every zone the city simulates, in a stable order used
// for deterministic iteration (statistics, dashboard zone ordering).
func AllZoneTypes() []ZoneType {
	return []ZoneType{
		ZoneCBD, ZoneResidential, ZoneIndustrial, ZoneAirport,
		ZoneBusStation, ZoneHospital, ZoneSchool, ZoneEntertainment,
	}
}

// Zone is one district of the simulated city: a rough 2D coordinate (for
// distance/ETA math — not real GPS, just a relative city-plane) and a
// baseline demand weight used to seed how many ride requests originate
// there at a given hour (see simulation/scenario_scheduler.go).
type Zone struct {
	Type ZoneType
	X    float64 // relative city-plane coordinate, km from city center
	Y    float64

	// BaseDemandWeight is a relative popularity multiplier for ride requests
	// originating in this zone under "normal" conditions (no rush hour, no
	// event). 1.0 = average zone. Scenario logic scales this further.
	BaseDemandWeight float64
}

// City is the fixed geography every simulation run operates on — one city,
// per the sprint brief ("hoạt động trong cùng một thành phố").
type City struct {
	Zones map[ZoneType]Zone
}

// NewDefaultCity returns Panda's simulated single city: 8 zones laid out on
// a simple relative plane so distance/ETA math has something real to work
// with. Coordinates and demand weights are simulation-design choices (not
// sourced from any BRB numeric rule — there is no real city being modeled),
// documented here rather than scattered as inline magic numbers.
func NewDefaultCity() *City {
	zones := map[ZoneType]Zone{
		ZoneCBD:           {Type: ZoneCBD, X: 0, Y: 0, BaseDemandWeight: 2.2},
		ZoneResidential:   {Type: ZoneResidential, X: -6, Y: 3, BaseDemandWeight: 1.4},
		ZoneIndustrial:    {Type: ZoneIndustrial, X: 8, Y: -4, BaseDemandWeight: 0.9},
		ZoneAirport:       {Type: ZoneAirport, X: 15, Y: 6, BaseDemandWeight: 0.6},
		ZoneBusStation:    {Type: ZoneBusStation, X: -3, Y: -5, BaseDemandWeight: 1.0},
		ZoneHospital:      {Type: ZoneHospital, X: 2, Y: -2, BaseDemandWeight: 0.8},
		ZoneSchool:        {Type: ZoneSchool, X: -4, Y: 1, BaseDemandWeight: 1.1},
		ZoneEntertainment: {Type: ZoneEntertainment, X: 5, Y: 4, BaseDemandWeight: 1.3},
	}
	return &City{Zones: zones}
}

// DistanceKM returns the straight-line distance between two zones on the
// city plane. Not real-world routing — a deliberately simple stand-in
// consistent with the rest of the simulation being a behavioral model, not
// a routing engine.
func (c *City) DistanceKM(a, b ZoneType) float64 {
	za, ok1 := c.Zones[a]
	zb, ok2 := c.Zones[b]
	if !ok1 || !ok2 {
		return 0
	}
	dx := za.X - zb.X
	dy := za.Y - zb.Y
	return math.Sqrt(dx*dx + dy*dy)
}
