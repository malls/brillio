// Relevance scoring (docs/SPEC.md section 6.2).
//
// Every listing gets a relevance score from 0 to 1. With a target budget the
// score is 60 percent budget fit plus 40 percent recency; without one it is
// recency alone, so results run newest first.
//
// Budget fit is 1 when the price equals the target. Under the target it is the
// price divided by the target (10 percent under scores 0.9). Over the target it
// falls twice as fast (10 percent over scores 0.8) and reaches 0 at 50 percent
// over.
//
// Recency halves every 30 days: 1 on the listing date, 0.5 at 30 days, 0.25 at
// 60. Age is whole UTC calendar days from the listed date to now; a future
// listed date counts as age 0.

package domain

import (
	"math"
	"time"
)

// The scoring constants. They are untyped so they combine with float64 and
// int values without conversions.
const (
	// BudgetWeight is the share of the relevance score that budget fit carries
	// when the search has a target budget.
	BudgetWeight = 0.6
	// RecencyWeight is the share of the relevance score that recency carries
	// when the search has a target budget. Without one, recency is the whole
	// score.
	RecencyWeight = 0.4
	// RecencyHalfLifeDays is the listing age, in days, at which recency halves.
	RecencyHalfLifeDays = 30
)

// secondsPerDay is the length of a UTC calendar day. UTC has no daylight
// saving time and Go time has no leap seconds, so every UTC day is this long.
const secondsPerDay = 86400

// BudgetFit returns how well price matches the target budget, from 0 to 1.
// It is 1 when price equals target and price/target below it. Above target it
// is 1 - 2*(price-target)/target, clamped at 0, so going over budget costs
// twice as much as coming in under and the fit reaches 0 at 50 percent over.
//
// The precondition is target > 0 and price >= 0. Search requests reject a
// target budget that is not positive, and results outside the precondition
// are unspecified.
func BudgetFit(price, target float64) float64 {
	if price <= target {
		return price / target
	}
	return max(0, 1-2*((price-target)/target))
}

// AgeDays returns how many whole UTC calendar days lie between listedDate and
// now. Each instant is first converted to UTC and reduced to its calendar
// date, so the time of day and the location of either value do not matter: a
// listing dated yesterday is 1 day old at any time today. A listedDate after
// now counts as age 0.
func AgeDays(listedDate, now time.Time) int {
	l, n := listedDate.UTC(), now.UTC()
	from := time.Date(l.Year(), l.Month(), l.Day(), 0, 0, 0, 0, time.UTC)
	to := time.Date(n.Year(), n.Month(), n.Day(), 0, 0, 0, 0, time.UTC)
	// Both are UTC midnights, so the difference is an exact multiple of a day.
	// Unix seconds, unlike to.Sub(from), do not saturate for a zero listedDate.
	return max(0, int((to.Unix()-from.Unix())/secondsPerDay))
}

// Recency returns the listing's recency at now, from 0 to 1: 0.5 raised to
// AgeDays(listedDate, now) divided by RecencyHalfLifeDays. It is 1 at age 0
// (including a future listedDate), 0.5 at 30 days and 0.25 at 60 days.
func Recency(listedDate, now time.Time) float64 {
	return math.Pow(0.5, float64(AgeDays(listedDate, now))/RecencyHalfLifeDays)
}

// Score combines budget fit and recency into the relevance score. A nil
// budgetFit means the search has no target budget, and the score is recency
// unchanged. Otherwise it is BudgetWeight*budgetFit + RecencyWeight*recency.
// With both inputs in [0, 1] the result is in [0, 1]. The score is returned
// at full precision; rounding for display is the caller's job.
func Score(budgetFit *float64, recency float64) float64 {
	if budgetFit == nil {
		return recency
	}
	return BudgetWeight*(*budgetFit) + RecencyWeight*recency
}
