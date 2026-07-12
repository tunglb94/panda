package entity_test

import (
	"testing"

	"github.com/fairride/pricing/domain/entity"
)

func testAirportConfig() entity.AirportFeeConfigV3 {
	return entity.AirportFeeConfigV3{
		PickupFee:  map[entity.VehicleType]int64{entity.VehicleTypeCar: 15000, entity.VehicleTypeVan: 20000},
		DropoffFee: map[entity.VehicleType]int64{entity.VehicleTypeCar: 5000, entity.VehicleTypeVan: 7000},
	}
}

func TestAirportFeeConfigV3_FeeFor_Pickup(t *testing.T) {
	cfg := testAirportConfig()
	if got := cfg.FeeFor(entity.VehicleTypeCar, entity.AirportLegPickup); got != 15000 {
		t.Errorf("car pickup: got %d, want 15000", got)
	}
	if got := cfg.FeeFor(entity.VehicleTypeVan, entity.AirportLegPickup); got != 20000 {
		t.Errorf("van pickup: got %d, want 20000", got)
	}
}

func TestAirportFeeConfigV3_FeeFor_Dropoff(t *testing.T) {
	cfg := testAirportConfig()
	if got := cfg.FeeFor(entity.VehicleTypeCar, entity.AirportLegDropoff); got != 5000 {
		t.Errorf("car dropoff: got %d, want 5000", got)
	}
}

func TestAirportFeeConfigV3_FeeFor_MotorcycleIsZero(t *testing.T) {
	// Deliberately absent from config — PRICING_V3_DESIGN.md Phần 7 /
	// MARKET_PRICING_RESEARCH.md Phần 1.3: no airport surcharge for bike.
	cfg := testAirportConfig()
	if got := cfg.FeeFor(entity.VehicleTypeMotorcycle, entity.AirportLegPickup); got != 0 {
		t.Errorf("motorcycle pickup: got %d, want 0", got)
	}
	if got := cfg.FeeFor(entity.VehicleTypeMotorcycle, entity.AirportLegDropoff); got != 0 {
		t.Errorf("motorcycle dropoff: got %d, want 0", got)
	}
}

func TestAirportFeeConfigV3_FeeFor_NoLeg(t *testing.T) {
	cfg := testAirportConfig()
	if got := cfg.FeeFor(entity.VehicleTypeCar, entity.AirportLegNone); got != 0 {
		t.Errorf("no leg: got %d, want 0", got)
	}
}
