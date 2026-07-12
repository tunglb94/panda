package app_test

// Golden snapshot test — 105 frozen scenarios (>= 100 required), spanning
// every vehicle class x 35 distances (1-100km) x all 5 commission tiers,
// each computed once with FareCalculatorV3 against config.Default() and
// frozen here. A future change to the calculator or the default config that
// silently shifts any of these numbers will fail this test — the intended
// signal for "did I just change Pricing V3's output without meaning to."
//
// Regenerated for the P0-1/P0-3 implementation pass (van MinimumFare
// 48,000->40,000; commission ladder: only Bronze changed 20%->16%, Silver/
// Gold/Platinum/Diamond reverted to BRB §7.1's 18%/16%/14%/12% instead of
// the earlier full-ladder 16/15/14/13/12).
//
// To regenerate after an intentional config/formula change: temporarily
// restore a generator like this file's history (loop the scenario grid,
// call EstimateV3, fmt.Printf the struct literal), run it, and replace the
// goldenCases slice body below — never hand-edit individual frozen numbers
// to "make the test pass" without first confirming the numbers moved for
// the intended reason.

import (
	"testing"
	"time"

	"github.com/fairride/pricing/app"
	"github.com/fairride/pricing/config"
	"github.com/fairride/pricing/domain/entity"
)

type goldenCase struct {
	vehicle         entity.VehicleType
	tier            entity.CommissionTier
	km              float64
	wantBaseFare    int64
	wantDistFare    int64
	wantRideFare    int64
	wantCommission  int64
	wantPlatformRev int64
	wantFinalFare   int64
}

var goldenCases = []goldenCase{
	{"car", "bronze", 1, 13000, 9500, 25000, 4000, 6300, 28000},
	{"car", "silver", 1.5, 13000, 14250, 27250, 4905, 7114, 30250},
	{"car", "gold", 2, 13000, 19000, 32000, 5120, 7308, 35000},
	{"car", "platinum", 2.5, 13000, 23300, 36300, 5082, 7274, 39300},
	{"car", "diamond", 3, 13000, 27600, 40600, 4872, 7085, 43600},
	{"car", "bronze", 4, 13000, 36200, 49200, 7872, 9785, 52200},
	{"car", "silver", 5, 13000, 44800, 57800, 10404, 12064, 60800},
	{"car", "gold", 6, 13000, 52600, 65600, 10496, 12146, 68600},
	{"car", "platinum", 7, 13000, 60400, 73400, 10276, 11948, 76400},
	{"car", "diamond", 8, 13000, 68200, 81200, 9744, 11470, 84200},
	{"car", "bronze", 9, 13000, 76000, 89000, 14240, 15516, 92000},
	{"car", "silver", 10, 13000, 83800, 96800, 17424, 18382, 99800},
	{"car", "gold", 11, 13000, 90800, 103800, 16608, 17647, 106800},
	{"car", "platinum", 12, 13000, 97800, 110800, 15512, 16661, 113800},
	{"car", "diamond", 13, 13000, 104800, 117800, 14136, 15422, 120800},
	{"car", "bronze", 14, 13000, 111800, 124800, 19968, 20671, 127800},
	{"car", "silver", 15, 13000, 118800, 131800, 23724, 24052, 134800},
	{"car", "gold", 16, 13000, 125800, 138800, 22208, 22687, 141800},
	{"car", "platinum", 18, 13000, 139800, 152800, 21392, 21953, 155800},
	{"car", "diamond", 20, 13000, 153800, 166800, 20016, 20714, 169800},
	{"car", "bronze", 22, 13000, 166200, 179200, 28672, 28505, 182200},
	{"car", "silver", 25, 13000, 184800, 197800, 35604, 34744, 200800},
	{"car", "gold", 28, 13000, 203400, 216400, 34624, 33862, 219400},
	{"car", "platinum", 30, 13000, 215800, 228800, 32032, 31529, 231800},
	{"car", "diamond", 35, 13000, 246800, 259800, 31176, 30758, 262800},
	{"car", "bronze", 40, 13000, 277800, 290800, 46528, 44575, 293800},
	{"car", "silver", 45, 13000, 304800, 317800, 57204, 54184, 320800},
	{"car", "gold", 50, 13000, 331800, 344800, 55168, 52351, 347800},
	{"car", "platinum", 55, 13000, 358800, 371800, 52052, 49547, 374800},
	{"car", "diamond", 60, 13000, 385800, 398800, 47856, 45770, 401800},
	{"car", "bronze", 65, 13000, 408800, 421800, 67488, 63439, 424800},
	{"car", "silver", 70, 13000, 431800, 444800, 80064, 74758, 447800},
	{"car", "gold", 80, 13000, 477800, 490800, 78528, 73375, 493800},
	{"car", "platinum", 90, 13000, 523800, 536800, 75152, 70337, 539800},
	{"car", "diamond", 100, 13000, 569800, 582800, 69936, 65642, 585800},
	{"motorcycle", "bronze", 1, 2500, 4200, 9000, 1440, 2196, 10000},
	{"motorcycle", "silver", 1.5, 2500, 6300, 9000, 1620, 2358, 10000},
	{"motorcycle", "gold", 2, 2500, 8400, 10900, 1744, 2470, 11900},
	{"motorcycle", "platinum", 2.5, 2500, 10250, 12750, 1785, 2506, 13750},
	{"motorcycle", "diamond", 3, 2500, 12100, 14600, 1752, 2477, 15600},
	{"motorcycle", "bronze", 4, 2500, 15800, 18300, 2928, 3535, 19300},
	{"motorcycle", "silver", 5, 2500, 19500, 22000, 3960, 4464, 23000},
	{"motorcycle", "gold", 6, 2500, 22800, 25300, 4048, 4543, 26300},
	{"motorcycle", "platinum", 7, 2500, 26100, 28600, 4004, 4504, 29600},
	{"motorcycle", "diamond", 8, 2500, 29400, 31900, 3828, 4345, 32900},
	{"motorcycle", "bronze", 9, 2500, 32700, 35200, 5632, 5969, 36200},
	{"motorcycle", "silver", 10, 2500, 36000, 38500, 6930, 7137, 39500},
	{"motorcycle", "gold", 11, 2500, 39000, 41500, 6640, 6876, 42500},
	{"motorcycle", "platinum", 12, 2500, 42000, 44500, 6230, 6507, 45500},
	{"motorcycle", "diamond", 13, 2500, 45000, 47500, 5700, 6030, 48500},
	{"motorcycle", "bronze", 14, 2500, 48000, 50500, 8080, 8172, 51500},
	{"motorcycle", "silver", 15, 2500, 51000, 53500, 9630, 9567, 54500},
	{"motorcycle", "gold", 16, 2500, 54000, 56500, 9040, 9036, 57500},
	{"motorcycle", "platinum", 18, 2500, 60000, 62500, 8750, 8775, 63500},
	{"motorcycle", "diamond", 20, 2500, 66000, 68500, 8220, 8298, 69500},
	{"motorcycle", "bronze", 22, 2500, 71400, 73900, 11824, 11542, 74900},
	{"motorcycle", "silver", 25, 2500, 79500, 82000, 14760, 14184, 83000},
	{"motorcycle", "gold", 28, 2500, 87600, 90100, 14416, 13874, 91100},
	{"motorcycle", "platinum", 30, 2500, 93000, 95500, 13370, 12933, 96500},
	{"motorcycle", "diamond", 35, 2500, 106500, 109000, 13080, 12672, 110000},
	{"motorcycle", "bronze", 40, 2500, 120000, 122500, 19600, 18540, 123500},
	{"motorcycle", "silver", 45, 2500, 132000, 134500, 24210, 22689, 135500},
	{"motorcycle", "gold", 50, 2500, 144000, 146500, 23440, 21996, 147500},
	{"motorcycle", "platinum", 55, 2500, 156000, 158500, 22190, 20871, 159500},
	{"motorcycle", "diamond", 60, 2500, 168000, 170500, 20460, 19314, 171500},
	{"motorcycle", "bronze", 65, 2500, 178500, 181000, 28960, 26964, 182000},
	{"motorcycle", "silver", 70, 2500, 189000, 191500, 34470, 31923, 192500},
	{"motorcycle", "gold", 80, 2500, 210000, 212500, 34000, 31500, 213500},
	{"motorcycle", "platinum", 90, 2500, 231000, 233500, 32690, 30321, 234500},
	{"motorcycle", "diamond", 100, 2500, 252000, 254500, 30540, 28386, 255500},
	{"van", "bronze", 1, 22000, 12500, 40000, 6400, 8460, 43000},
	{"van", "silver", 1.5, 22000, 18750, 40750, 7335, 9301, 43750},
	{"van", "gold", 2, 22000, 25000, 47000, 7520, 9468, 50000},
	{"van", "platinum", 2.5, 22000, 30650, 52650, 7371, 9334, 55650},
	{"van", "diamond", 3, 22000, 36300, 58300, 6996, 8996, 61300},
	{"van", "bronze", 4, 22000, 47600, 69600, 11136, 12722, 72600},
	{"van", "silver", 5, 22000, 58900, 80900, 14562, 15806, 83900},
	{"van", "gold", 6, 22000, 69100, 91100, 14576, 15818, 94100},
	{"van", "platinum", 7, 22000, 79300, 101300, 14182, 15464, 104300},
	{"van", "diamond", 8, 22000, 89500, 111500, 13380, 14742, 114500},
	{"van", "bronze", 9, 22000, 99700, 121700, 19472, 20225, 124700},
	{"van", "silver", 10, 22000, 109900, 131900, 23742, 24068, 134900},
	{"van", "gold", 11, 22000, 119000, 141000, 22560, 23004, 144000},
	{"van", "platinum", 12, 22000, 128100, 150100, 21014, 21613, 153100},
	{"van", "diamond", 13, 22000, 137200, 159200, 19104, 19894, 162200},
	{"van", "bronze", 14, 22000, 146300, 168300, 26928, 26935, 171300},
	{"van", "silver", 15, 22000, 155400, 177400, 31932, 31439, 180400},
	{"van", "gold", 16, 22000, 164500, 186500, 29840, 29556, 189500},
	{"van", "platinum", 18, 22000, 182700, 204700, 28658, 28492, 207700},
	{"van", "diamond", 20, 22000, 200900, 222900, 26748, 26773, 225900},
	{"van", "bronze", 22, 22000, 217100, 239100, 38256, 37130, 242100},
	{"van", "silver", 25, 22000, 241400, 263400, 47412, 45371, 266400},
	{"van", "gold", 28, 22000, 265700, 287700, 46032, 44129, 290700},
	{"van", "platinum", 30, 22000, 281900, 303900, 42546, 40991, 306900},
	{"van", "diamond", 35, 22000, 322400, 344400, 41328, 39895, 347400},
	{"van", "bronze", 40, 22000, 362900, 384900, 61584, 58126, 387900},
	{"van", "silver", 45, 22000, 397900, 419900, 75582, 70724, 422900},
	{"van", "gold", 50, 22000, 432900, 454900, 72784, 68206, 457900},
	{"van", "platinum", 55, 22000, 467900, 489900, 68586, 64427, 492900},
	{"van", "diamond", 60, 22000, 502900, 524900, 62988, 59389, 527900},
	{"van", "bronze", 65, 22000, 532900, 554900, 88784, 82606, 557900},
	{"van", "silver", 70, 22000, 562900, 584900, 105282, 97454, 587900},
	{"van", "gold", 80, 22000, 622900, 644900, 103184, 95566, 647900},
	{"van", "platinum", 90, 22000, 682900, 704900, 98686, 91517, 707900},
	{"van", "diamond", 100, 22000, 742900, 764900, 91788, 85309, 767900},
}

func TestFareCalculatorV3_GoldenCases(t *testing.T) {
	if len(goldenCases) < 100 {
		t.Fatalf("golden case count = %d, want >= 100 (sprint brief PHẦN 16)", len(goldenCases))
	}

	cfg := config.Default()
	calc := app.NewFareCalculatorV3(cfg.Fare, cfg.Airport, cfg.Commission, cfg.VATRate, app.DefaultRuleConfigs())
	requestTime := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC) // fixed, neutral (not night/peak/weekend-sensitive)

	for _, tc := range goldenCases {
		tc := tc
		t.Run(string(tc.vehicle)+"_"+string(tc.tier)+"_"+formatKM(tc.km), func(t *testing.T) {
			fb, err := calc.EstimateV3(entity.RideInputV3{
				VehicleType:    tc.vehicle,
				DistanceKM:     tc.km,
				DurationMin:    tc.km * 2.2,
				RequestTime:    requestTime,
				CommissionTier: tc.tier,
			})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if fb.BaseFare != tc.wantBaseFare {
				t.Errorf("BaseFare = %d, want %d", fb.BaseFare, tc.wantBaseFare)
			}
			if fb.DistanceFare != tc.wantDistFare {
				t.Errorf("DistanceFare = %d, want %d", fb.DistanceFare, tc.wantDistFare)
			}
			if fb.RideFare != tc.wantRideFare {
				t.Errorf("RideFare = %d, want %d", fb.RideFare, tc.wantRideFare)
			}
			if fb.Commission != tc.wantCommission {
				t.Errorf("Commission = %d, want %d", fb.Commission, tc.wantCommission)
			}
			if fb.PlatformRevenue != tc.wantPlatformRev {
				t.Errorf("PlatformRevenue = %d, want %d", fb.PlatformRevenue, tc.wantPlatformRev)
			}
			if fb.FinalFare != tc.wantFinalFare {
				t.Errorf("FinalFare = %d, want %d", fb.FinalFare, tc.wantFinalFare)
			}
		})
	}
}

func formatKM(km float64) string {
	if km == float64(int64(km)) {
		return itoa(int64(km)) + "km"
	}
	return itoa(int64(km*10)) + "dkm" // e.g. 15 -> "1.5km" encoded as "15dkm"
}

func itoa(v int64) string {
	if v == 0 {
		return "0"
	}
	neg := v < 0
	if neg {
		v = -v
	}
	var buf [20]byte
	i := len(buf)
	for v > 0 {
		i--
		buf[i] = byte('0' + v%10)
		v /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
