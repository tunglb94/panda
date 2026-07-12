package entity

// Weather is one of the 4 required conditions.
type Weather string

const (
	WeatherSunny       Weather = "sunny"       // Nắng
	WeatherLightRain   Weather = "light_rain"  // Mưa nhỏ
	WeatherHeavyRain   Weather = "heavy_rain"  // Mưa lớn
	WeatherFlooded     Weather = "flooded"     // Ngập
)

// IsRainy reports whether this weather activates BRB §2.2.13's Rain
// Surcharge signal in the real Pricing Engine (see integration/pricing_adapter.go).
func (w Weather) IsRainy() bool {
	return w == WeatherLightRain || w == WeatherHeavyRain || w == WeatherFlooded
}
