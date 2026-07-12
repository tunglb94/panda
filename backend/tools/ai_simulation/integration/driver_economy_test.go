package integration

import (
	"testing"

	"github.com/fairride/ai_simulation/domain/entity"
)

func TestDriverEconomy_TierForDriver_BRB72Thresholds(t *testing.T) {
	e := NewDriverEconomy()
	cases := []struct {
		name       string
		trips      int
		rating     float64
		wantTier   entity.AccountType
	}{
		{"new driver", 0, 4.9, entity.AccountBronze},
		{"silver eligible, rating too low", 100, 4.4, entity.AccountBronze},
		{"exactly silver threshold", 100, 4.5, entity.AccountSilver},
		{"exactly gold threshold", 500, 4.7, entity.AccountGold},
		{"exactly platinum threshold", 1500, 4.8, entity.AccountPlatinum},
		{"exactly diamond threshold", 4000, 4.85, entity.AccountDiamond},
		{"high trips but rating just under diamond", 4000, 4.84, entity.AccountPlatinum},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := e.TierForDriver(c.trips, c.rating)
			if got != c.wantTier {
				t.Errorf("TierForDriver(%d, %v) = %q, want %q", c.trips, c.rating, got, c.wantTier)
			}
		})
	}
}

func TestDriverEconomy_CommissionRatePercent_BRB71Table(t *testing.T) {
	e := NewDriverEconomy()
	want := map[entity.AccountType]int64{
		entity.AccountBronze:   20,
		entity.AccountSilver:   18,
		entity.AccountGold:     16,
		entity.AccountPlatinum: 14,
		entity.AccountDiamond:  12,
	}
	for tier, rate := range want {
		if got := e.CommissionRatePercent(tier); got != rate {
			t.Errorf("CommissionRatePercent(%q) = %d, want %d", tier, got, rate)
		}
	}
}

func TestDriverEconomy_Split_BookingFeeGoesEntirelyToCommission(t *testing.T) {
	e := NewDriverEconomy()
	// Bronze = 20% commission on the metered fare; booking fee is excluded
	// from the metered fare per BRB §2.2.5 and flows entirely to the
	// platform, never to driver net.
	commission, driverNet := e.Split(entity.AccountBronze, 100_000, 5_000)

	wantCommission := int64(100_000*20/100) + 5_000 // 20,000 + 5,000 booking fee
	wantDriverNet := int64(100_000 - 100_000*20/100) // 80,000
	if commission != wantCommission {
		t.Errorf("commission = %d, want %d", commission, wantCommission)
	}
	if driverNet != wantDriverNet {
		t.Errorf("driverNet = %d, want %d", driverNet, wantDriverNet)
	}
	// Conservation check: commission + driverNet must equal meteredFare + bookingFee.
	if commission+driverNet != 100_000+5_000 {
		t.Errorf("commission+driverNet = %d, want meteredFare+bookingFee = %d", commission+driverNet, 100_000+5_000)
	}
}

func TestDriverEconomy_Split_HigherTierKeepsMoreNet(t *testing.T) {
	e := NewDriverEconomy()
	_, bronzeNet := e.Split(entity.AccountBronze, 100_000, 0)
	_, diamondNet := e.Split(entity.AccountDiamond, 100_000, 0)
	if diamondNet <= bronzeNet {
		t.Errorf("expected a diamond-tier driver to net more than bronze on the same fare, bronze=%d diamond=%d", bronzeNet, diamondNet)
	}
}
