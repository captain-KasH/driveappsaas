// Package fare is the server-authoritative fare calculator. Pure functions,
// integer minor units only (LKR cents), no floating point on money.
package fare

import (
	"math"
	"time"
)

type Rule struct {
	Currency           string
	BaseMinor          int64
	PerKmMinor         int64
	PerMinuteMinor     int64
	MinimumMinor       int64
	BookingFeeMinor    int64
	ScheduledFeeMinor  int64
	WaitingPerMinMinor int64
	FreeWaitingMinutes int64
	NightSurchargeBps  int64
	NightStart         *time.Duration // offset from local midnight; nil = no night window
	NightEnd           *time.Duration
	TaxBps             int64
	RoundToMinor       int64
}

type Input struct {
	DistanceM  int64
	DurationS  int64
	WaitingS   int64
	Scheduled  bool
	LocalStart time.Time // trip (or pickup) start in the tenant's timezone
}

type Line struct {
	Code        string `json:"code"`
	Label       string `json:"label"`
	AmountMinor int64  `json:"amountMinor"`
}

type Breakdown struct {
	Currency      string `json:"currency"`
	Lines         []Line `json:"lines"`
	SubtotalMinor int64  `json:"subtotalMinor"`
	TaxMinor      int64  `json:"taxMinor"`
	TotalMinor    int64  `json:"totalMinor"`
}

// mulDiv returns round-half-up(a*b/d) for non-negative inputs.
func mulDiv(a, b, d int64) int64 {
	if a <= 0 || b <= 0 {
		return 0
	}
	return (a*b + d/2) / d
}

func Compute(r Rule, in Input) Breakdown {
	b := Breakdown{Currency: r.Currency}
	add := func(code, label string, amount int64) {
		if amount != 0 {
			b.Lines = append(b.Lines, Line{code, label, amount})
		}
	}
	add("base", "Base fare", r.BaseMinor)
	add("distance", "Distance", mulDiv(r.PerKmMinor, in.DistanceM, 1000))
	add("time", "Time", mulDiv(r.PerMinuteMinor, in.DurationS, 60))
	add("waiting", "Waiting time", mulDiv(r.WaitingPerMinMinor, in.WaitingS-r.FreeWaitingMinutes*60, 60))

	metered := sum(b.Lines)
	if metered < r.MinimumMinor {
		add("minimum_adjustment", "Minimum fare adjustment", r.MinimumMinor-metered)
		metered = r.MinimumMinor
	}
	if r.NightSurchargeBps > 0 && inNightWindow(r, in.LocalStart) {
		add("night_surcharge", "Night surcharge", mulDiv(metered, r.NightSurchargeBps, 10000))
	}
	add("booking_fee", "Booking fee", r.BookingFeeMinor)
	if in.Scheduled {
		add("scheduled_fee", "Scheduled ride fee", r.ScheduledFeeMinor)
	}
	b.SubtotalMinor = sum(b.Lines)
	b.TaxMinor = mulDiv(b.SubtotalMinor, r.TaxBps, 10000)
	total := b.SubtotalMinor + b.TaxMinor
	if r.RoundToMinor > 1 {
		rounded := (total + r.RoundToMinor/2) / r.RoundToMinor * r.RoundToMinor
		add("rounding", "Rounding", rounded-total)
		total = rounded
	}
	b.TotalMinor = total
	if b.Lines == nil {
		b.Lines = []Line{}
	}
	return b
}

func sum(lines []Line) int64 {
	var s int64
	for _, l := range lines {
		s += l.AmountMinor
	}
	return s
}

// inNightWindow handles windows that wrap midnight (e.g. 22:00–05:00).
func inNightWindow(r Rule, t time.Time) bool {
	if r.NightStart == nil || r.NightEnd == nil {
		return false
	}
	tod := time.Duration(t.Hour())*time.Hour + time.Duration(t.Minute())*time.Minute + time.Duration(t.Second())*time.Second
	s, e := *r.NightStart, *r.NightEnd
	if s <= e {
		return tod >= s && tod < e
	}
	return tod >= s || tod < e
}

// EstimateMethod names the heuristic so stored quotes say how they were made.
const EstimateMethod = "straight_line_x1.35_at_22kmh_v1"

// EstimateRoute approximates road distance and duration without a routing
// provider. ponytail: haversine × 1.35 at 22 km/h; replace with a routing API (T-013b) when one is chosen.
func EstimateRoute(lat1, lng1, lat2, lng2 float64) (meters, seconds int64) {
	const r = 6371000.0
	toRad := math.Pi / 180
	dLat, dLng := (lat2-lat1)*toRad, (lng2-lng1)*toRad
	h := math.Sin(dLat/2)*math.Sin(dLat/2) + math.Cos(lat1*toRad)*math.Cos(lat2*toRad)*math.Sin(dLng/2)*math.Sin(dLng/2)
	straight := 2 * r * math.Asin(math.Sqrt(h))
	meters = int64(math.Round(straight * 1.35))
	seconds = int64(math.Round(float64(meters) / (22000.0 / 3600.0)))
	return meters, seconds
}
