package entity_test

import (
	"testing"

	"github.com/fairride/pricing/domain/entity"
)

// TestDefaultFareConfig verifies the default config is self-consistent.
func TestDefaultFareConfig_HasAllVehicleTypes(t *testing.T) {
	cfg := entity.DefaultFareConfig()

	required := []entity.VehicleType{
		entity.VehicleTypeCar,
		entity.VehicleTypeMotorcycle,
		entity.VehicleTypeVan,
	}
	for _, vt := range required {
		if _, ok := cfg.Rates[vt]; !ok {
			t.Errorf("DefaultFareConfig missing rates for vehicle type %q", vt)
		}
	}
}

func TestDefaultFareConfig_PositiveRates(t *testing.T) {
	cfg := entity.DefaultFareConfig()
	for vt, r := range cfg.Rates {
		if r.BaseFare <= 0 {
			t.Errorf("%s: BaseFare must be positive, got %d", vt, r.BaseFare)
		}
		if r.PerKmRate <= 0 {
			t.Errorf("%s: PerKmRate must be positive, got %d", vt, r.PerKmRate)
		}
		if r.PerMinuteRate <= 0 {
			t.Errorf("%s: PerMinuteRate must be positive, got %d", vt, r.PerMinuteRate)
		}
		if r.MinimumFare <= 0 {
			t.Errorf("%s: MinimumFare must be positive, got %d", vt, r.MinimumFare)
		}
		if r.BookingFee <= 0 {
			t.Errorf("%s: BookingFee must be positive, got %d", vt, r.BookingFee)
		}
	}
}

func TestDefaultFareConfig_MinimumFareCoversBaseFare(t *testing.T) {
	cfg := entity.DefaultFareConfig()
	for vt, r := range cfg.Rates {
		if r.MinimumFare < r.BaseFare {
			t.Errorf("%s: MinimumFare (%d) should be ≥ BaseFare (%d)", vt, r.MinimumFare, r.BaseFare)
		}
	}
}

func TestVehicleTypeConstants(t *testing.T) {
	cases := []struct {
		vt   entity.VehicleType
		want string
	}{
		{entity.VehicleTypeCar, "car"},
		{entity.VehicleTypeMotorcycle, "motorcycle"},
		{entity.VehicleTypeVan, "van"},
	}
	for _, tc := range cases {
		if string(tc.vt) != tc.want {
			t.Errorf("VehicleType %q, want %q", tc.vt, tc.want)
		}
	}
}
