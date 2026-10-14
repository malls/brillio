package httpapi

import (
	"context"
	"math"
	"net/http"
	"time"

	"brillio/backend/internal/domain"
)

// Searcher runs the filtered listing query (docs/SPEC.md section 6.1). It
// returns the listings matching f in any order, since ranking orders them.
// *store.Store implements it.
type Searcher interface {
	Search(ctx context.Context, f domain.SearchFilter) ([]domain.Listing, error)
}

// searchListings serves GET /api/listings/search. It is the search pipeline
// of docs/SPEC.md section 6 in one place: each stage below is one call and
// its early return, in order, so a new stage slots in between two others.
func (a *api) searchListings(w http.ResponseWriter, r *http.Request) {
	// Parse and validate (section 5.3): every problem is reported at once.
	q, errs := parseSearchRequest(r)
	if len(errs) > 0 {
		a.writeError(w, r, http.StatusBadRequest, codeValidation, msgValidation, errs)
		return
	}

	// Filter in SQL (section 6.1).
	listings, err := a.deps.Searcher.Search(r.Context(), q.Filter)
	if err != nil {
		a.internalError(w, r, err)
		return
	}

	// Dedupe slot (section 9.1): merging cross-source duplicates goes here,
	// between search and rank. Not built in v1.

	// Score and order (sections 6.2 and 6.3).
	now := a.deps.Now()
	ranked := domain.Rank(listings, domain.RankInput{TargetBudget: q.TargetBudget, Now: now})

	// Paginate and assign ranks (section 6.4). The parser already enforced
	// the page bounds, so an error here is a bug: a 500, not a 400.
	page, err := domain.Paginate(ranked, q.Page, q.PageSize)
	if err != nil {
		a.internalError(w, r, err)
		return
	}

	// Respond (section 5.4).
	a.writeJSON(w, r, http.StatusOK, newSearchResponse(page, q, now))
}

// searchResponse is the GET /api/listings/search body (docs/SPEC.md section
// 5.4). The response types are wire DTOs: domain.Listing and domain.Scored
// are never marshaled, and no field is omitempty, so absent filters and a
// missing budget fit encode as null.
type searchResponse struct {
	Items          []searchItem   `json:"items"`
	Page           pageMeta       `json:"page"`
	Scoring        scoringMeta    `json:"scoring"`
	AppliedFilters appliedFilters `json:"appliedFilters"`
}

// searchItem is one ranked listing.
type searchItem struct {
	Rank        int       `json:"rank"`
	ID          string    `json:"id"`
	Source      string    `json:"source"`
	Address     string    `json:"address"`
	City        string    `json:"city"`
	State       string    `json:"state"`
	Zip         string    `json:"zip"`
	Price       float64   `json:"price"`
	Bedrooms    int       `json:"bedrooms"`
	Bathrooms   float64   `json:"bathrooms"`
	Sqft        int       `json:"sqft"`
	Latitude    float64   `json:"latitude"`
	Longitude   float64   `json:"longitude"`
	ListedDate  string    `json:"listedDate"` // domain.DateLayout
	Status      string    `json:"status"`
	Description string    `json:"description"`
	Relevance   relevance `json:"relevance"`
}

// relevance is an item's score and its components, rounded to four
// decimals.
type relevance struct {
	Score     float64  `json:"score"`
	BudgetFit *float64 `json:"budgetFit"`
	Recency   float64  `json:"recency"`
}

// pageMeta is the page position within the whole ranked result set.
type pageMeta struct {
	Page       int `json:"page"`
	PageSize   int `json:"pageSize"`
	TotalItems int `json:"totalItems"`
	TotalPages int `json:"totalPages"`
}

// scoringMeta states how the items were scored and as of which UTC date.
type scoringMeta struct {
	TargetBudget        *float64 `json:"targetBudget"`
	BudgetWeight        float64  `json:"budgetWeight"`
	RecencyWeight       float64  `json:"recencyWeight"`
	RecencyHalfLifeDays int      `json:"recencyHalfLifeDays"`
	AsOf                string   `json:"asOf"`
}

// appliedFilters echoes the effective filters after parsing; nil is absent.
type appliedFilters struct {
	MinPrice    *float64 `json:"minPrice"`
	MaxPrice    *float64 `json:"maxPrice"`
	MinBedrooms *int     `json:"minBedrooms"`
	City        *string  `json:"city"`
	Keyword     *string  `json:"keyword"`
	Status      []string `json:"status"`
}

// newSearchResponse maps a page of ranked listings to the response body.
// now must be the instant the listings were ranked against; its UTC date is
// scoring.asOf.
func newSearchResponse(p domain.Page, q SearchQuery, now time.Time) searchResponse {
	items := make([]searchItem, 0, len(p.Items))
	for _, it := range p.Items {
		items = append(items, newSearchItem(it))
	}
	return searchResponse{
		Items:          items,
		Page:           pageMeta{Page: p.Page, PageSize: p.PageSize, TotalItems: p.TotalItems, TotalPages: p.TotalPages},
		Scoring:        newScoringMeta(q.TargetBudget, now),
		AppliedFilters: newAppliedFilters(q.Filter),
	}
}

// newSearchItem maps one ranked listing, rounding its relevance for display.
// The domain values stay unrounded; ordering already used them.
func newSearchItem(it domain.RankedItem) searchItem {
	rel := relevance{Score: round4(it.Score), Recency: round4(it.Recency)}
	if it.BudgetFit != nil {
		fit := round4(*it.BudgetFit)
		rel.BudgetFit = &fit
	}
	return searchItem{
		Rank:        it.Rank,
		ID:          it.ID,
		Source:      it.Source,
		Address:     it.Address,
		City:        it.City,
		State:       it.State,
		Zip:         it.Zip,
		Price:       it.Price,
		Bedrooms:    it.Bedrooms,
		Bathrooms:   it.Bathrooms,
		Sqft:        it.Sqft,
		Latitude:    it.Latitude,
		Longitude:   it.Longitude,
		ListedDate:  it.ListedDate.Format(domain.DateLayout),
		Status:      string(it.Status),
		Description: it.Description,
		Relevance:   rel,
	}
}

// newScoringMeta returns the scoring block. Without a target budget the
// score is recency alone, so the weights read 0 and 1 (docs/SPEC.md section
// 5.4).
func newScoringMeta(targetBudget *float64, now time.Time) scoringMeta {
	s := scoringMeta{
		BudgetWeight:        0,
		RecencyWeight:       1,
		RecencyHalfLifeDays: domain.RecencyHalfLifeDays,
		AsOf:                now.UTC().Format(domain.DateLayout),
	}
	if targetBudget != nil {
		budget := *targetBudget
		s.TargetBudget = &budget
		s.BudgetWeight, s.RecencyWeight = domain.BudgetWeight, domain.RecencyWeight
	}
	return s
}

// newAppliedFilters echoes f: city trimmed, keyword as received, and the
// effective statuses in canonical order.
func newAppliedFilters(f domain.SearchFilter) appliedFilters {
	af := appliedFilters{
		MinPrice:    f.MinPrice,
		MaxPrice:    f.MaxPrice,
		MinBedrooms: f.MinBedrooms,
		Status:      make([]string, 0, len(f.Statuses)),
	}
	if f.City != "" {
		af.City = &f.City
	}
	if f.Keyword != "" {
		af.Keyword = &f.Keyword
	}
	for _, s := range f.Statuses {
		af.Status = append(af.Status, string(s))
	}
	return af
}

// round4 rounds x to four decimals, half away from zero, for display.
func round4(x float64) float64 {
	return math.Round(x*1e4) / 1e4
}
