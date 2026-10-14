package domain_test

import (
	"fmt"
	"math"
	"testing"
	"time"

	"brillio/backend/internal/domain"
)

// fixedNow is the clock for every domain test. The time of day is deliberate:
// age is counted in calendar days, so it must not matter.
var fixedNow = time.Date(2026, 10, 13, 15, 30, 0, 0, time.UTC)

// eps is the tolerance for values that are not exact in float64: fractional
// powers and weighted sums (which the compiler may fuse into an FMA).
const eps = 1e-12

// near reports whether a and b differ by at most eps.
func near(a, b float64) bool {
	return math.Abs(a-b) <= eps
}

// mustDate parses a YYYY-MM-DD date the way listings are parsed: midnight UTC.
func mustDate(s string) time.Time {
	t, err := time.Parse(domain.DateLayout, s)
	if err != nil {
		panic(err)
	}
	return t
}

// ptr returns a pointer to a copy of f.
func ptr(f float64) *float64 {
	return &f
}

func TestBudgetFit(t *testing.T) {
	const target = 450000
	// Every value here is exact in float64, so the comparison is ==.
	tests := []struct {
		name  string
		price float64
		want  float64
	}{
		{name: "budget_equal_to_price", price: 450000, want: 1.0},
		{name: "worked_405000_10pct_under", price: 405000, want: 0.9},
		{name: "worked_495000_10pct_over", price: 495000, want: 0.8},
		{name: "worked_562500_25pct_over", price: 562500, want: 0.5},
		{name: "worked_675000_50pct_over", price: 675000, want: 0},
		{name: "beyond_50pct_over_clamps_to_0", price: 900000, want: 0},
		{name: "price_0", price: 0, want: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := domain.BudgetFit(tt.price, target); got != tt.want {
				t.Errorf("BudgetFit(%v, %v) = %v, want %v", tt.price, float64(target), got, tt.want)
			}
		})
	}

	t.Run("asymmetry_at_matching_percentages", func(t *testing.T) {
		// Prices are integer literals: 450000*1.1 is 495000.00000000006.
		pairs := []struct {
			pct          string
			under, over  float64
			wantUnderFit float64
		}{
			{pct: "5%", under: 427500, over: 472500, wantUnderFit: 0.95},
			{pct: "10%", under: 405000, over: 495000, wantUnderFit: 0.9},
			{pct: "25%", under: 337500, over: 562500, wantUnderFit: 0.75},
			{pct: "40%", under: 270000, over: 630000, wantUnderFit: 0.6},
		}
		for _, p := range pairs {
			under := domain.BudgetFit(p.under, target)
			over := domain.BudgetFit(p.over, target)
			if !near(under, p.wantUnderFit) {
				t.Errorf("%s under: BudgetFit(%v) = %v, want %v", p.pct, p.under, under, p.wantUnderFit)
			}
			if over >= under {
				t.Errorf("%s: over fit %v >= under fit %v, want over < under", p.pct, over, under)
			}
			if !near(1-over, 2*(1-under)) {
				t.Errorf("%s: over penalty %v, want twice the under penalty %v", p.pct, 1-over, 2*(1-under))
			}
		}
	})
}

func TestAgeDays(t *testing.T) {
	edt := time.FixedZone("EDT", -4*3600)
	tests := []struct {
		name   string
		listed time.Time
		nows   []time.Time
		want   int
	}{
		{name: "age_0_days", listed: mustDate("2026-10-13"), nows: []time.Time{fixedNow}, want: 0},
		{name: "age_30_days", listed: mustDate("2026-09-13"), nows: []time.Time{fixedNow}, want: 30},
		{name: "age_60_days", listed: mustDate("2026-08-14"), nows: []time.Time{fixedNow}, want: 60},
		{name: "age_90_days", listed: mustDate("2026-07-15"), nows: []time.Time{fixedNow}, want: 90},
		{name: "future_listed_date", listed: mustDate("2026-10-20"), nows: []time.Time{fixedNow}, want: 0},
		{
			// 12 hours apart but across midnight: a calendar day, not a 24h period.
			name:   "calendar_days_not_24h_periods",
			listed: time.Date(2026, 10, 12, 18, 0, 0, 0, time.UTC),
			nows:   []time.Time{time.Date(2026, 10, 13, 6, 0, 0, 0, time.UTC)},
			want:   1,
		},
		{
			name:   "now_time_of_day_ignored",
			listed: mustDate("2026-09-13"),
			nows: []time.Time{
				time.Date(2026, 10, 13, 0, 0, 0, 0, time.UTC),
				time.Date(2026, 10, 13, 23, 59, 59, 999_000_000, time.UTC),
			},
			want: 30,
		},
		{
			// 22:00 EDT on Oct 13 is 02:00 UTC on Oct 14.
			name:   "now_non_utc_location",
			listed: mustDate("2026-09-13"),
			nows:   []time.Time{time.Date(2026, 10, 13, 22, 0, 0, 0, edt)},
			want:   31,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, now := range tt.nows {
				if got := domain.AgeDays(tt.listed, now); got != tt.want {
					t.Errorf("AgeDays(%v, %v) = %d, want %d", tt.listed, now, got, tt.want)
				}
			}
		})
	}
}

func TestRecency(t *testing.T) {
	tests := []struct {
		name   string
		listed string
		want   float64
		exact  bool // whole half-lives are exact powers of two
	}{
		{name: "age_0_days", listed: "2026-10-13", want: 1, exact: true},
		{name: "age_30_days", listed: "2026-09-13", want: 0.5, exact: true},
		{name: "age_60_days", listed: "2026-08-14", want: 0.25, exact: true},
		{name: "age_90_days", listed: "2026-07-15", want: 0.125, exact: true},
		{name: "future_listed_date", listed: "2026-10-20", want: 1, exact: true},
		// Fractional exponents catch integer division of age by the half-life.
		{name: "age_15_days", listed: "2026-09-28", want: math.Sqrt(0.5)},
		// SPEC section 5.4 shows this recency as 0.3536.
		{name: "age_45_days_spec_example", listed: "2026-08-29", want: 0.35355339059327373},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := domain.Recency(mustDate(tt.listed), fixedNow)
			if tt.exact && got != tt.want {
				t.Errorf("Recency(%s) = %v, want exactly %v", tt.listed, got, tt.want)
			}
			if !tt.exact && !near(got, tt.want) {
				t.Errorf("Recency(%s) = %v, want %v (within %g)", tt.listed, got, tt.want, eps)
			}
		})
	}
}

func TestScore(t *testing.T) {
	t.Run("no_budget", func(t *testing.T) {
		for _, rec := range []float64{0, 0.125, 0.35355339059327373, 0.5, 1} {
			if got := domain.Score(nil, rec); got != rec {
				t.Errorf("Score(nil, %v) = %v, want recency unchanged", rec, got)
			}
		}
	})

	t.Run("with_budget_worked_values", func(t *testing.T) {
		rows := []struct {
			fit, rec, want float64
		}{
			{fit: 1, rec: 1, want: 1.0},
			{fit: 0.9, rec: 0.5, want: 0.74},
			{fit: 0.8, rec: 0.25, want: 0.58},
			{fit: 0.5, rec: 0.125, want: 0.35},
			{fit: 0, rec: 1, want: 0.4},
			{fit: 1, rec: 0.35355339059327373, want: 0.7414213562373095},
		}
		for _, r := range rows {
			if got := domain.Score(ptr(r.fit), r.rec); !near(got, r.want) {
				t.Errorf("Score(%v, %v) = %v, want %v (within %g)", r.fit, r.rec, got, r.want, eps)
			}
		}
	})

	t.Run("score_in_0_1", func(t *testing.T) {
		fits := []*float64{nil, ptr(0), ptr(0.5), ptr(0.8), ptr(0.9), ptr(1)}
		recencies := []float64{0, 0.125, 0.25, 0.5, 1}
		for _, fit := range fits {
			label := "nil"
			if fit != nil {
				label = fmt.Sprint(*fit)
			}
			for _, rec := range recencies {
				if s := domain.Score(fit, rec); s < 0 || s > 1 {
					t.Errorf("Score(%s, %v) = %v, want within [0, 1]", label, rec, s)
				}
			}
		}
	})
}
