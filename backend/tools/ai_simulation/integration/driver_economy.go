package integration

import "github.com/fairride/ai_simulation/domain/entity"

// DriverEconomy computes commission tier and driver net income from
// Business Rule Bible v1.0 Part 7 — Driver Economy. No production Go
// implementation of tiered commission exists anywhere in
// backend/services/driver today (confirmed by inspection: DriverProfile has
// no rating/earnings/trip-count fields, and no commission/tier code exists
// in that service or anywhere else in backend/). Per the sprint brief's
// "Nếu thiếu → chỉ ghi TODO. KHÔNG tự nghĩ rule mới" discipline (already
// established for the Promotion Engine sprint), this is new code — but
// every number in it is BRB §7.1's exact published commission table, not
// invented.
//
// Simplification, documented rather than silently assumed: BRB §7.2's full
// tier-upgrade criteria include trailing 30/90/180-day acceptance rate,
// completion rate, and violation-free windows this simulation does not
// track. Tier assignment here uses BRB's trip-count and rating thresholds
// only (the two factors DriverAgent actually models) — a driver who would
// be rating/trip-eligible in the real system might not qualify here if a
// real evaluation would also weigh acceptance/completion rate. This under-
// promotes relative to a fully faithful evaluation, never over-promotes.
type DriverEconomy struct{}

func NewDriverEconomy() *DriverEconomy { return &DriverEconomy{} }

// TierForDriver resolves BRB §7.2's trip-count + rating thresholds against
// a driver's simulated lifetime stats.
func (DriverEconomy) TierForDriver(totalTrips int, rating float64) entity.AccountType {
	switch {
	case totalTrips >= 4000 && rating >= 4.85:
		return entity.AccountDiamond
	case totalTrips >= 1500 && rating >= 4.8:
		return entity.AccountPlatinum
	case totalTrips >= 500 && rating >= 4.7:
		return entity.AccountGold
	case totalTrips >= 100 && rating >= 4.5:
		return entity.AccountSilver
	default:
		return entity.AccountBronze
	}
}

// CommissionRatePercent is BRB §7.1's exact published table.
func (DriverEconomy) CommissionRatePercent(tier entity.AccountType) int64 {
	switch tier {
	case entity.AccountDiamond:
		return 12
	case entity.AccountPlatinum:
		return 14
	case entity.AccountGold:
		return 16
	case entity.AccountSilver:
		return 18
	default: // Bronze / New
		return 20
	}
}

// Split applies BRB §7.1's commission rule: "Commission rate is applied to
// the metered fare (Base + Distance + Time + surcharges, excluding Booking
// Fee and Toll)." meteredFareVND must already exclude the booking fee —
// callers pass FareBreakdown.RideFare (see pricing_adapter.go), not .Total.
// Returns (commissionVND, driverNetVND); driverNetVND + commissionVND ==
// meteredFareVND + bookingFeeVND (the booking fee flows to the driver's net
// only in the sense that it's part of what the rider pays out-of-pocket but
// BRB §2.2.5 says booking fee "is NOT shared with the driver... flows
// entirely to FAIRRIDE" — so it is added to commission, not to driver net).
func (e *DriverEconomy) Split(tier entity.AccountType, meteredFareVND, bookingFeeVND int64) (commissionVND, driverNetVND int64) {
	rate := e.CommissionRatePercent(tier)
	commissionVND = meteredFareVND*rate/100 + bookingFeeVND
	driverNetVND = meteredFareVND - (meteredFareVND * rate / 100)
	return commissionVND, driverNetVND
}
