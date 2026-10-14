package domain_test

import (
	"slices"
	"testing"

	"brillio/backend/internal/domain"
)

// listing returns a listing with just the fields ranking reads.
func listing(source, id, date string, price float64) domain.Listing {
	return domain.Listing{
		ID:         id,
		Source:     source,
		Price:      price,
		ListedDate: mustDate(date),
		Status:     domain.StatusActive,
	}
}

// key is a listing's identity, (Source, ID).
type key struct{ source, id string }

// listingKeys returns the identities of listings, in order.
func listingKeys(listings []domain.Listing) []key {
	out := make([]key, len(listings))
	for i, l := range listings {
		out[i] = key{l.Source, l.ID}
	}
	return out
}

// keys returns the identities of ranked, in order.
func keys(ranked []domain.Scored) []key {
	out := make([]key, len(ranked))
	for i, s := range ranked {
		out[i] = key{s.Source, s.ID}
	}
	return out
}

// tieFixture is six listings whose expected order (with fixedNow and no
// budget) exercises every tiebreak key, each deciding one adjacent pair. The
// slice is given in reverse of that order.
func tieFixture() []domain.Listing {
	want := []domain.Listing{
		// listedDate desc decides vs #2 (price alone would put #2 first).
		listing("MLS_B", "B7", "2026-10-20", 600000),
		// price asc decides vs #3 (source alone would put #3 first).
		listing("MLS_B", "B3", "2026-10-13", 300000),
		// source asc decides vs #4 (id alone would put #4 first).
		listing("MLS_A", "Z1", "2026-10-13", 450000),
		// id asc decides vs #5: same price and date.
		listing("MLS_B", "B1", "2026-10-13", 450000),
		// score desc decides vs #6 (price, source and id would put #6 first).
		// Without a budget the score tracks the listed date, so listedDate
		// desc agrees here; with_budget is the case only the score decides.
		listing("MLS_B", "B2", "2026-10-13", 450000),
		listing("MLS_A", "A1", "2026-09-13", 100000),
	}
	slices.Reverse(want)
	return want
}

// tieOrder is the expected rank order of tieFixture.
var tieOrder = []key{
	{"MLS_B", "B7"}, {"MLS_B", "B3"}, {"MLS_A", "Z1"}, {"MLS_B", "B1"}, {"MLS_B", "B2"}, {"MLS_A", "A1"},
}

// permutations calls visit with every ordering of s (Heap's algorithm). The
// slice passed to visit is reused between calls.
func permutations(s []domain.Listing, visit func([]domain.Listing)) {
	a := slices.Clone(s)
	c := make([]int, len(a))
	visit(a)
	for i := 0; i < len(a); {
		if c[i] < i {
			if i%2 == 0 {
				a[0], a[i] = a[i], a[0]
			} else {
				a[c[i]], a[i] = a[i], a[c[i]]
			}
			visit(a)
			c[i]++
			i = 0
		} else {
			c[i] = 0
			i++
		}
	}
}

func TestRank(t *testing.T) {
	budget := ptr(450000)

	t.Run("no_matches", func(t *testing.T) {
		inputs := []struct {
			name     string
			listings []domain.Listing
			budget   *float64
		}{
			{name: "nil without budget", listings: nil},
			{name: "empty without budget", listings: []domain.Listing{}},
			{name: "nil with budget", listings: nil, budget: budget},
			{name: "empty with budget", listings: []domain.Listing{}, budget: budget},
		}
		for _, in := range inputs {
			got := domain.Rank(in.listings, domain.RankInput{TargetBudget: in.budget, Now: fixedNow})
			if got == nil || len(got) != 0 {
				t.Errorf("%s: Rank = %#v, want a non-nil empty slice", in.name, got)
			}
		}
	})

	t.Run("no_budget", func(t *testing.T) {
		in := []domain.Listing{
			listing("MLS_A", "OLD", "2026-08-14", 100000),
			listing("MLS_A", "MID", "2026-09-13", 200000),
			listing("MLS_A", "NEW", "2026-10-13", 300000),
		}
		got := domain.Rank(in, domain.RankInput{Now: fixedNow})
		wantIDs := []string{"NEW", "MID", "OLD"}
		wantRecency := []float64{1, 0.5, 0.25}
		if len(got) != len(wantIDs) {
			t.Fatalf("Rank returned %d items, want %d", len(got), len(wantIDs))
		}
		for i, s := range got {
			if s.ID != wantIDs[i] {
				t.Errorf("item %d: ID %q, want %q (newest first)", i, s.ID, wantIDs[i])
			}
			if s.BudgetFit != nil {
				t.Errorf("item %d: BudgetFit = %v, want nil without a budget", i, *s.BudgetFit)
			}
			if s.Recency != wantRecency[i] {
				t.Errorf("item %d: Recency %v, want %v", i, s.Recency, wantRecency[i])
			}
			if s.Score != s.Recency {
				t.Errorf("item %d: Score %v, want exactly Recency %v", i, s.Score, s.Recency)
			}
		}
	})

	t.Run("with_budget", func(t *testing.T) {
		in := []domain.Listing{
			listing("MLS_A", "W", "2026-09-13", 405000),
			listing("MLS_A", "X", "2026-08-29", 450000),
			listing("MLS_A", "Y", "2026-10-13", 495000),
			listing("MLS_A", "Z", "2026-10-13", 675000),
		}
		want := []struct {
			id              string
			fit, rec, score float64
			recExact        bool // whole half-lives are exact
		}{
			{id: "Y", fit: 0.8, rec: 1, score: 0.88, recExact: true},
			// SPEC section 5.4's A1 example: fit 1.0, recency 0.3536.
			{id: "X", fit: 1, rec: 0.35355339059327373, score: 0.7414213562373095},
			{id: "W", fit: 0.9, rec: 0.5, score: 0.74, recExact: true},
			{id: "Z", fit: 0, rec: 1, score: 0.4, recExact: true},
		}
		got := domain.Rank(in, domain.RankInput{TargetBudget: budget, Now: fixedNow})
		if len(got) != len(want) {
			t.Fatalf("Rank returned %d items, want %d", len(got), len(want))
		}
		seen := map[*float64]string{}
		for i, w := range want {
			s := got[i]
			if s.ID != w.id {
				t.Errorf("item %d: ID %q, want %q", i, s.ID, w.id)
				continue
			}
			if s.BudgetFit == nil {
				t.Errorf("%s: BudgetFit nil, want %v", w.id, w.fit)
				continue
			}
			if prev, ok := seen[s.BudgetFit]; ok {
				t.Errorf("%s: BudgetFit pointer shared with %s", w.id, prev)
			}
			seen[s.BudgetFit] = w.id
			if *s.BudgetFit != w.fit {
				t.Errorf("%s: BudgetFit %v, want %v", w.id, *s.BudgetFit, w.fit)
			}
			if (w.recExact && s.Recency != w.rec) || (!w.recExact && !near(s.Recency, w.rec)) {
				t.Errorf("%s: Recency %v, want %v", w.id, s.Recency, w.rec)
			}
			if !near(s.Score, w.score) {
				t.Errorf("%s: Score %v, want %v (within %g)", w.id, s.Score, w.score, eps)
			}
		}
	})

	t.Run("tied_scores_tiebreak_order", func(t *testing.T) {
		got := domain.Rank(tieFixture(), domain.RankInput{Now: fixedNow})
		if !slices.Equal(keys(got), tieOrder) {
			t.Fatalf("Rank order = %v, want %v", keys(got), tieOrder)
		}
		for i, s := range got[:5] {
			if s.Score != 1.0 {
				t.Errorf("item %d (%s): Score %v, want exactly 1", i, s.ID, s.Score)
			}
		}
		if got[5].Score != 0.5 {
			t.Errorf("item 5 (%s): Score %v, want exactly 0.5", got[5].ID, got[5].Score)
		}
	})

	t.Run("tied_scores_with_budget", func(t *testing.T) {
		// 405000 is 10% under and 472500 is 5% over: both fit exactly 0.9.
		// Source ascending alone would put A2 first; price ascending decides.
		in := []domain.Listing{
			listing("MLS_A", "A2", "2026-10-13", 472500),
			listing("MLS_B", "B9", "2026-10-13", 405000),
		}
		got := domain.Rank(in, domain.RankInput{TargetBudget: budget, Now: fixedNow})
		want := []key{{"MLS_B", "B9"}, {"MLS_A", "A2"}}
		if !slices.Equal(keys(got), want) {
			t.Fatalf("Rank order = %v, want %v", keys(got), want)
		}
		for _, s := range got {
			if s.BudgetFit == nil {
				t.Errorf("%s: BudgetFit nil, want exactly 0.9", s.ID)
			} else if *s.BudgetFit != 0.9 {
				t.Errorf("%s: BudgetFit %v, want exactly 0.9", s.ID, *s.BudgetFit)
			}
		}
		if got[0].Score != got[1].Score {
			t.Errorf("Scores %v and %v, want an exact tie", got[0].Score, got[1].Score)
		}
	})

	t.Run("input_order_independent", func(t *testing.T) {
		count := 0
		permutations(tieFixture(), func(p []domain.Listing) {
			count++
			if got := keys(domain.Rank(p, domain.RankInput{Now: fixedNow})); !slices.Equal(got, tieOrder) {
				t.Errorf("input %v: Rank order = %v, want %v", listingKeys(p), got, tieOrder)
			}
		})
		if count != 720 {
			t.Errorf("checked %d permutations, want 720", count)
		}
	})

	t.Run("does_not_mutate_input", func(t *testing.T) {
		in := tieFixture()
		before := slices.Clone(in)
		got := domain.Rank(in, domain.RankInput{TargetBudget: budget, Now: fixedNow})
		if !slices.Equal(in, before) {
			t.Errorf("Rank modified its input: got %v, want %v", in, before)
		}
		if len(got) != len(in) {
			t.Errorf("Rank returned %d items, want %d", len(got), len(in))
		}
	})
}
