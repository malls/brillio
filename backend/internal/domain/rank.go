package domain

import (
	"cmp"
	"slices"
	"time"
)

// RankInput carries the per-request inputs to Rank other than the listings.
type RankInput struct {
	// TargetBudget is the search's target budget in US dollars; nil means the
	// search has none and listings are scored on recency alone. When set it
	// must be > 0.
	TargetBudget *float64
	// Now is the instant that listing ages are measured against. Now must be
	// the request clock; the zero time makes every listing look future-dated
	// (age 0).
	Now time.Time
}

// Relevance is a listing's relevance score and the components it was built
// from, at full precision. It is an in-memory type: map it to a response DTO
// (which rounds for display) rather than marshaling it.
type Relevance struct {
	// Score is the relevance score in [0, 1] that results are ordered by.
	Score float64
	// BudgetFit is how well the price matches the target budget, in [0, 1];
	// nil when the search has no target budget.
	BudgetFit *float64
	// Recency is the age-based recency in [0, 1].
	Recency float64
}

// Scored is a listing together with its relevance. Like Listing, it is an
// in-memory type; never json.Marshal it.
type Scored struct {
	// Listing is the scored listing.
	Listing
	// Relevance is the listing's relevance for the request.
	Relevance
}

// Rank scores every listing for the request (docs/SPEC.md section 6.2) and
// returns them in relevance order (section 6.3): score descending, then
// listedDate descending, price ascending, source ascending and id ascending.
// Scores are compared unrounded, so a tie means exact equality. Because
// (Source, ID) identifies a listing, the order is total and deterministic: the
// same listings give the same order whatever order they arrive in.
//
// Rank returns a new slice and never reorders or modifies listings. The result
// is never nil; nil or empty input gives an empty slice. Each listing with a
// budget fit gets its own BudgetFit pointer.
//
// in.Now must be the request clock; the zero time makes every listing look
// future-dated (age 0).
func Rank(listings []Listing, in RankInput) []Scored {
	out := make([]Scored, 0, len(listings))
	for _, l := range listings {
		rel := Relevance{Recency: Recency(l.ListedDate, in.Now)}
		if in.TargetBudget != nil {
			fit := BudgetFit(l.Price, *in.TargetBudget)
			rel.BudgetFit = &fit
		}
		rel.Score = Score(rel.BudgetFit, rel.Recency)
		out = append(out, Scored{Listing: l, Relevance: rel})
	}
	slices.SortStableFunc(out, byRelevance)
	return out
}

// byRelevance is the relevance order (docs/SPEC.md section 6.3) as a
// comparator for slices.SortStableFunc. It is the one place that ordering is
// defined; a new sort option would sit beside it.
func byRelevance(a, b Scored) int {
	return cmp.Or(
		// Higher score first, compared unrounded.
		cmp.Compare(b.Score, a.Score),
		// Then the more recently listed.
		b.ListedDate.Compare(a.ListedDate),
		// Then the cheaper.
		cmp.Compare(a.Price, b.Price),
		// Then by source, so the order does not depend on input order.
		cmp.Compare(a.Source, b.Source),
		// Then by id; (Source, ID) is unique, so this always decides.
		cmp.Compare(a.ID, b.ID),
	)
}
