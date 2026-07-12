package redis_test

import (
	"context"
	"testing"
	"time"

	"github.com/fairride/dispatch/domain/entity"
	dispatchredis "github.com/fairride/dispatch/infrastructure/redis"
)

func setupLocationTest(t *testing.T) *dispatchredis.DriverLocationRepository {
	t.Helper()
	flushKeys(t, "fairride:dispatch:drv:*")
	return dispatchredis.NewDriverLocationRepository(testClient)
}

func TestUpdateLocation_StoresAndActive(t *testing.T) {
	repo := setupLocationTest(t)
	ctx := context.Background()

	if err := repo.UpdateLocation(ctx, "d1", 10.762, 106.660, "", true, false); err != nil {
		t.Fatalf("UpdateLocation: %v", err)
	}

	active, err := repo.IsActive(ctx, "d1")
	if err != nil {
		t.Fatalf("IsActive: %v", err)
	}
	if !active {
		t.Error("expected driver to be active after location update")
	}
}

func TestFindNearby_ReturnsSorted(t *testing.T) {
	repo := setupLocationTest(t)
	ctx := context.Background()

	// Ho Chi Minh City area
	_ = repo.UpdateLocation(ctx, "d_near", 10.762, 106.660, "", true, false)    // ~0 km from search point
	_ = repo.UpdateLocation(ctx, "d_far", 10.800, 106.710, "", true, false)     // ~6 km away
	_ = repo.UpdateLocation(ctx, "d_farther", 10.900, 106.800, "", true, false) // ~18 km away

	drivers, err := repo.FindNearby(ctx, 10.762, 106.660, 10.0, 10)
	if err != nil {
		t.Fatalf("FindNearby: %v", err)
	}
	if len(drivers) < 2 {
		t.Fatalf("expected at least 2 drivers within 10 km, got %d", len(drivers))
	}
	// Nearest should be first
	if drivers[0].DriverID != "d_near" {
		t.Errorf("nearest driver = %q, want d_near", drivers[0].DriverID)
	}
	// d_farther is beyond 10 km, so it should not appear
	for _, d := range drivers {
		if d.DriverID == "d_farther" {
			t.Error("d_farther is beyond radius and should not appear in results")
		}
	}
}

func TestFindNearby_EmptyWhenNoDrivers(t *testing.T) {
	repo := setupLocationTest(t)
	drivers, err := repo.FindNearby(context.Background(), 10.0, 106.0, 10.0, 10)
	if err != nil {
		t.Fatalf("FindNearby: %v", err)
	}
	if len(drivers) != 0 {
		t.Errorf("expected 0 drivers, got %d", len(drivers))
	}
}

func TestIsActive_FalseAfterTTLExpiry(t *testing.T) {
	// Use a very short TTL so we can verify expiry without sleeping long
	repo := dispatchredis.NewDriverLocationRepositoryWithTTL(testClient, 100*time.Millisecond)
	flushKeys(t, "fairride:dispatch:drv:*")
	ctx := context.Background()

	_ = repo.UpdateLocation(ctx, "d_ttl", 10.0, 106.0, "", true, false)

	active, _ := repo.IsActive(ctx, "d_ttl")
	if !active {
		t.Fatal("expected active immediately after update")
	}

	time.Sleep(200 * time.Millisecond)

	active, err := repo.IsActive(ctx, "d_ttl")
	if err != nil {
		t.Fatalf("IsActive: %v", err)
	}
	if active {
		t.Error("expected driver to be inactive after TTL expiry")
	}
}

func TestFindNearby_ReportsServiceTypeAndCapability(t *testing.T) {
	repo := setupLocationTest(t)
	ctx := context.Background()

	_ = repo.UpdateLocation(ctx, "d_bike_plus", 10.762, 106.660, entity.ServiceTypeBikePlus, true, true)
	_ = repo.UpdateLocation(ctx, "d_unreported", 10.763, 106.661, "", true, false)

	drivers, err := repo.FindNearby(ctx, 10.762, 106.660, 10.0, 10)
	if err != nil {
		t.Fatalf("FindNearby: %v", err)
	}
	var gotBikePlus, gotUnreported bool
	for _, d := range drivers {
		switch d.DriverID {
		case "d_bike_plus":
			gotBikePlus = true
			if d.ServiceType != entity.ServiceTypeBikePlus {
				t.Errorf("d_bike_plus ServiceType = %q, want %q", d.ServiceType, entity.ServiceTypeBikePlus)
			}
			if !d.RideEnabled || !d.DeliveryEnabled {
				t.Errorf("d_bike_plus capability = ride:%v delivery:%v, want both true", d.RideEnabled, d.DeliveryEnabled)
			}
		case "d_unreported":
			gotUnreported = true
			if d.ServiceType != "" {
				t.Errorf("d_unreported ServiceType = %q, want empty", d.ServiceType)
			}
		}
	}
	if !gotBikePlus || !gotUnreported {
		t.Fatalf("expected both drivers in results, got %+v", drivers)
	}
}

// TestFindNearby_CapabilityDefaultsWhenNeverReported locks in backward
// compatibility: a driver whose capability was never explicitly set (e.g.
// RemoveLocation deleted the keys, or they simply expired) still shows up
// as RideEnabled=true/DeliveryEnabled=false — the migration 008 DB column
// defaults — not Go's zero-value false/false.
func TestFindNearby_CapabilityDefaultsWhenNeverReported(t *testing.T) {
	repo := setupLocationTest(t)
	ctx := context.Background()
	flushKeys(t, "fairride:dispatch:drv:ride:*")
	flushKeys(t, "fairride:dispatch:drv:delivery:*")

	// GeoAdd directly so no capability keys ever get written for this driver.
	_ = repo.UpdateLocation(ctx, "d_legacy", 10.762, 106.660, "", true, false)
	flushKeys(t, "fairride:dispatch:drv:ride:*")
	flushKeys(t, "fairride:dispatch:drv:delivery:*")

	drivers, err := repo.FindNearby(ctx, 10.762, 106.660, 10.0, 10)
	if err != nil {
		t.Fatalf("FindNearby: %v", err)
	}
	if len(drivers) != 1 {
		t.Fatalf("expected 1 driver, got %d", len(drivers))
	}
	if !drivers[0].RideEnabled || drivers[0].DeliveryEnabled {
		t.Errorf("capability defaults = ride:%v delivery:%v, want ride:true delivery:false", drivers[0].RideEnabled, drivers[0].DeliveryEnabled)
	}
}

func TestRemoveLocation(t *testing.T) {
	repo := setupLocationTest(t)
	ctx := context.Background()

	_ = repo.UpdateLocation(ctx, "d1", 10.0, 106.0, "", true, false)
	if err := repo.RemoveLocation(ctx, "d1"); err != nil {
		t.Fatalf("RemoveLocation: %v", err)
	}

	active, _ := repo.IsActive(ctx, "d1")
	if active {
		t.Error("expected driver to be inactive after removal")
	}

	drivers, _ := repo.FindNearby(ctx, 10.0, 106.0, 5.0, 10)
	for _, d := range drivers {
		if d.DriverID == "d1" {
			t.Error("d1 should not appear in FindNearby after removal")
		}
	}
}
