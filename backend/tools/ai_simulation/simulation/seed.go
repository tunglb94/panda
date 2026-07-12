package simulation

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/fairride/ai_simulation/domain/entity"
)

// seedDrivers populates n randomized DriverAgents and publishes their
// starting position to the Dispatch adapter so they're immediately
// matchable. Value ranges here are simulation-design choices, not sourced
// from any BRB number — there is no real driver population to sample.
func seedDrivers(w *World, n int) {
	zones := entity.AllZoneTypes()

	for i := 0; i < n; i++ {
		id := fmt.Sprintf("driver-%04d", i)
		zone := zones[w.Rand.Intn(len(zones))]
		vehicleType, serviceType := pickDriverVehicleAndServiceType(w.Rand)
		d := &entity.DriverAgent{
			ID:          id,
			Age:         22 + w.Rand.Intn(38), // 22-59
			Experience:  w.Rand.Intn(10),
			Rating:      4.2 + w.Rand.Float64()*0.75, // 4.2-4.95
			VehicleType: vehicleType,
			ServiceType: serviceType,
			// Every driver is Ride-capable by default (migration 008's DB
			// default); ~30% also opt into Delivery — a simulation-design
			// choice sized to give Delivery jobs real supply, not a BRB number.
			RideEnabled:     true,
			DeliveryEnabled: w.Rand.Float64() < 0.30,
			AccountType:     entity.AccountBronze,
			Satisfaction:    0.5 + w.Rand.Float64()*0.4,
			Fatigue:         w.Rand.Float64() * 0.2,
			PhoneBattery:    0.6 + w.Rand.Float64()*0.4,
			Fuel:            0.5 + w.Rand.Float64()*0.5,
			Cash:            int64(200_000 + w.Rand.Intn(2_000_000)),
			Zone:            zone,
			Online:          w.Rand.Float64() < 0.6,
			TotalTrips:      w.Rand.Intn(3000),
		}
		w.Drivers[id] = d
		if d.Online {
			pos := w.City.Zones[zone]
			_ = w.Dispatch.SetDriverPosition(context.Background(), id, pos.X, pos.Y, d.ServiceType, d.RideEnabled, d.DeliveryEnabled)
		}
	}
}

// pickDriverVehicleAndServiceType assigns a coherent (VehicleType,
// ServiceType) pair — ServiceType.RequiredVehicleType() always matches the
// returned VehicleType, mirroring driver.DriverProfile.SetServiceCapability's
// validation rule. Two-step pick: family (bike vs car), a simulation-design
// weight (not a BRB number — roughly preserves this simulation's original
// motorcycle-dominant SEA market mix); then tier within family, at the
// exact ratios the sprint brief specifies (Bike 65%/Bike Plus 35%, Car
// 75%/Car XL 25%).
func pickDriverVehicleAndServiceType(rnd *rand.Rand) (entity.DriverVehicleType, entity.ServiceType) {
	if rnd.Float64() < 0.55 { // bike family
		if rnd.Float64() < 0.65 {
			return entity.VehicleMotorcycle, entity.ServiceBike
		}
		return entity.VehicleMotorcycle, entity.ServiceBikePlus
	}
	// car family
	if rnd.Float64() < 0.75 {
		return entity.VehicleCar, entity.ServiceCar
	}
	return entity.VehicleVan, entity.ServiceCarXL
}

// seedRiders populates n randomized RiderAgents.
func seedRiders(w *World, n int) {
	zones := entity.AllZoneTypes()
	habits := []entity.CommuteHabit{entity.HabitCommuter, entity.HabitOccasional, entity.HabitNightOwl, entity.HabitBusinessTraveler}
	habitWeights := []float64{0.5, 0.3, 0.15, 0.05}
	memberships := []entity.Membership{entity.MembershipFree, entity.MembershipSilver, entity.MembershipGold, entity.MembershipDiamond}
	membershipWeights := []float64{0.7, 0.2, 0.08, 0.02}

	for i := 0; i < n; i++ {
		id := fmt.Sprintf("rider-%05d", i)
		habit := weightedHabit(w.Rand, habits, habitWeights)
		r := &entity.RiderAgent{
			ID:                     id,
			PriceSensitivity:       w.Rand.Float64(),
			Income:                 int64(4_000_000 + w.Rand.Intn(46_000_000)), // 4M-50M VND/month
			Habit:                  habit,
			WorkStartHour:          7 + w.Rand.Intn(3),  // 7-9
			WorkEndHour:            16 + w.Rand.Intn(4), // 16-19
			Patience:               0.3 + w.Rand.Float64()*0.6,
			Membership:             weightedMembership(w.Rand, memberships, membershipWeights),
			HasActiveVoucher:       w.Rand.Float64() < 0.25,
			VoucherDiscountPercent: 10,
			Zone:                   zones[w.Rand.Intn(len(zones))],
			TripCount:              w.Rand.Intn(200),
		}
		if r.HasActiveVoucher {
			w.voucherIssuedCount++
		}
		w.Riders[id] = r
	}
}

func weightedHabit(rnd *rand.Rand, habits []entity.CommuteHabit, weights []float64) entity.CommuteHabit {
	var total float64
	for _, wgt := range weights {
		total += wgt
	}
	r := rnd.Float64() * total
	for i, wgt := range weights {
		r -= wgt
		if r <= 0 {
			return habits[i]
		}
	}
	return habits[len(habits)-1]
}

func weightedMembership(rnd *rand.Rand, memberships []entity.Membership, weights []float64) entity.Membership {
	var total float64
	for _, wgt := range weights {
		total += wgt
	}
	r := rnd.Float64() * total
	for i, wgt := range weights {
		r -= wgt
		if r <= 0 {
			return memberships[i]
		}
	}
	return memberships[len(memberships)-1]
}
