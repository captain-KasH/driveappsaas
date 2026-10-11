package fare

import (
	"testing"
	"time"
)

func dur(h, m int) *time.Duration {
	d := time.Duration(h)*time.Hour + time.Duration(m)*time.Minute
	return &d
}

var colombo, _ = time.LoadLocation("Asia/Colombo")

func at(h, m int) time.Time { return time.Date(2026, 10, 12, h, m, 0, 0, colombo) }

// A typical Colombo tuk/taxi rule: LKR 100 base, LKR 85/km, LKR 5/min, LKR 250 minimum.
var rule = Rule{
	Currency: "LKR", BaseMinor: 10000, PerKmMinor: 8500, PerMinuteMinor: 500, MinimumMinor: 25000,
	BookingFeeMinor: 2000, ScheduledFeeMinor: 5000, WaitingPerMinMinor: 300, FreeWaitingMinutes: 3,
	NightSurchargeBps: 2000, NightStart: dur(22, 0), NightEnd: dur(5, 0), TaxBps: 0, RoundToMinor: 100,
}

func TestCompute(t *testing.T) {
	cases := []struct {
		name  string
		rule  Rule
		in    Input
		total int64
		lines map[string]int64
	}{
		{
			name: "day trip 5.2km 14min",
			rule: rule, in: Input{DistanceM: 5200, DurationS: 840, LocalStart: at(10, 0)},
			// 10000 + 8500*5.2=44200 + 500*14=7000 = 61200; +2000 fee = 63200
			total: 63200, lines: map[string]int64{"base": 10000, "distance": 44200, "time": 7000, "booking_fee": 2000},
		},
		{
			name: "short trip lifted to minimum",
			rule: rule, in: Input{DistanceM: 800, DurationS: 180, LocalStart: at(10, 0)},
			// 10000 + 6800 + 1500 = 18300 → minimum 25000 (+6700); +2000
			total: 27000, lines: map[string]int64{"minimum_adjustment": 6700},
		},
		{
			name: "night surcharge applies across midnight on the metered amount",
			rule: rule, in: Input{DistanceM: 5200, DurationS: 840, LocalStart: at(1, 30)},
			// metered 61200 * 20% = 12240; 61200+12240+2000 = 75440 → rounds to 75400
			total: 75400, lines: map[string]int64{"night_surcharge": 12240, "rounding": -40},
		},
		{
			name: "night window end is exclusive",
			rule: rule, in: Input{DistanceM: 5200, DurationS: 840, LocalStart: at(5, 0)},
			total: 63200,
		},
		{
			name: "waiting beyond free minutes and scheduled fee",
			rule: rule, in: Input{DistanceM: 5200, DurationS: 840, WaitingS: 7 * 60, Scheduled: true, LocalStart: at(10, 0)},
			// waiting (7-3)=4min*300=1200; 61200+1200+2000+5000 = 69400
			total: 69400, lines: map[string]int64{"waiting": 1200, "scheduled_fee": 5000},
		},
		{
			name: "waiting inside free minutes is free",
			rule: rule, in: Input{DistanceM: 5200, DurationS: 840, WaitingS: 120, LocalStart: at(10, 0)},
			total: 63200, lines: map[string]int64{"waiting": 0},
		},
		{
			name: "tax on subtotal, rounded half up to the rupee",
			rule: Rule{Currency: "LKR", BaseMinor: 10050, TaxBps: 1800, RoundToMinor: 100},
			in:   Input{LocalStart: at(10, 0)},
			// 10050 * 18% = 1809; 11859 → 11900
			total: 11900, lines: map[string]int64{"rounding": 41},
		},
		{
			name:  "zero distance and no minimum = base only",
			rule:  Rule{Currency: "LKR", BaseMinor: 5000, PerKmMinor: 8000, RoundToMinor: 1},
			in:    Input{LocalStart: at(10, 0)},
			total: 5000,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			b := Compute(c.rule, c.in)
			if b.TotalMinor != c.total {
				t.Fatalf("total = %d, want %d (%+v)", b.TotalMinor, c.total, b.Lines)
			}
			got := map[string]int64{}
			for _, l := range b.Lines {
				got[l.Code] = l.AmountMinor
			}
			for code, want := range c.lines {
				if got[code] != want {
					t.Errorf("line %s = %d, want %d", code, got[code], want)
				}
			}
			if sum(b.Lines)+b.TaxMinor != b.TotalMinor {
				t.Errorf("lines + tax (%d) != total (%d)", sum(b.Lines)+b.TaxMinor, b.TotalMinor)
			}
		})
	}
}

func TestEstimateRoute(t *testing.T) {
	// Colombo Fort → Bambalapitiya is ~4.6 km straight line.
	m, s := EstimateRoute(6.9344, 79.8428, 6.8906, 79.8562)
	if m < 6000 || m > 7000 || s < 900 || s > 1200 {
		t.Fatalf("estimate = %dm %ds, outside expected band", m, s)
	}
}
