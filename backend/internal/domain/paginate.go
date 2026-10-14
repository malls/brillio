package domain

import (
	"errors"
	"fmt"
)

// MaxPageSize is the largest allowed page size (docs/SPEC.md section 5.3). It
// is exported so that request parsing enforces the same bound as Paginate.
const MaxPageSize = 100

// ErrPageOutOfRange is wrapped in the error Paginate returns when page is
// less than 1. Check for it with errors.Is.
var ErrPageOutOfRange = errors.New("page out of range")

// ErrPageSizeOutOfRange is wrapped in the error Paginate returns when
// pageSize is outside 1..MaxPageSize. Check for it with errors.Is.
var ErrPageSizeOutOfRange = errors.New("pageSize out of range")

// RankedItem is a ranked listing on a page, with its rank.
type RankedItem struct {
	// Scored is the listing and its relevance.
	Scored
	// Rank is the 1-based position in the whole ranked result set, not
	// within the page: page 2 with page size 10 starts at rank 11.
	Rank int
}

// Page is one page of a ranked result set and the metadata to page through
// the rest of it (docs/SPEC.md section 6.4).
type Page struct {
	// Items are the page's listings in rank order. When Paginate returns no
	// error it is never nil: a page past the end, or a page of an empty
	// result set, has no items.
	Items []RankedItem
	// Page is the requested 1-based page number, echoed even past the end.
	Page int
	// PageSize is the requested page size.
	PageSize int
	// TotalItems is the size of the whole ranked result set.
	TotalItems int
	// TotalPages is ceil(TotalItems / PageSize), 0 when TotalItems is 0.
	TotalPages int
}

// Paginate returns page number page, of pageSize items each, from ranked,
// which must already be in rank order (see Rank). Each item's Rank is its
// 1-based position in the whole of ranked. A page past the end is not an
// error: it has no items and the same metadata, so a client can recover.
//
// page must be >= 1 and pageSize within 1..MaxPageSize. Otherwise Paginate
// returns the zero Page and an error that joins one wrapped
// ErrPageOutOfRange or ErrPageSizeOutOfRange per bad value, so callers can
// tell the two apart with errors.Is. Request parsing is expected to enforce
// the same bounds first, which makes these errors a defensive guard.
func Paginate(ranked []Scored, page, pageSize int) (Page, error) {
	var errs []error
	if page < 1 {
		errs = append(errs, fmt.Errorf("%w: got %d, want >= 1", ErrPageOutOfRange, page))
	}
	if pageSize < 1 || pageSize > MaxPageSize {
		errs = append(errs, fmt.Errorf("%w: got %d, want 1..%d", ErrPageSizeOutOfRange, pageSize, MaxPageSize))
	}
	if len(errs) > 0 {
		return Page{}, errors.Join(errs...)
	}

	total := len(ranked)
	totalPages := total / pageSize
	if total%pageSize != 0 {
		totalPages++
	}
	p := Page{Items: []RankedItem{}, Page: page, PageSize: pageSize, TotalItems: total, TotalPages: totalPages}
	// Checked before any multiplication: past here (page-1)*pageSize < total,
	// so a huge page number cannot overflow.
	if page > totalPages {
		return p, nil
	}
	start := (page - 1) * pageSize
	end := min(start+pageSize, total)
	p.Items = make([]RankedItem, end-start)
	for k := range p.Items {
		p.Items[k] = RankedItem{Scored: ranked[start+k], Rank: start + k + 1}
	}
	return p, nil
}
