package httpapi

import (
	"fmt"
	"math"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"unicode/utf8"

	"brillio/backend/internal/domain"
)

// Search query parameter names (docs/SPEC.md section 5.3). Names are
// case-sensitive.
const (
	paramMinPrice     = "minPrice"
	paramMaxPrice     = "maxPrice"
	paramMinBedrooms  = "minBedrooms"
	paramCity         = "city"
	paramKeyword      = "keyword"
	paramTargetBudget = "targetBudget"
	paramStatus       = "status"
	paramPage         = "page"
	paramPageSize     = "pageSize"
)

// searchParams is the search parameter allowlist in the section 5.3 table
// order. SearchQueryParams hands out copies; the parser reads it directly, so
// the two cannot drift.
var searchParams = [...]string{
	paramMinPrice, paramMaxPrice, paramMinBedrooms, paramCity, paramKeyword,
	paramTargetBudget, paramStatus, paramPage, paramPageSize,
}

// SearchQueryParams returns the names GET /api/listings/search accepts, in
// the docs/SPEC.md section 5.3 table order. Any other name is a 400
// unknown_param. Each call returns a new slice, so callers may modify it.
func SearchQueryParams() []string {
	return slices.Clone(searchParams[:])
}

// fieldQuery is the pseudo-field a malformed_query error is reported on: the
// problem is the query string as a whole, not one parameter.
const fieldQuery = "query"

// Field error codes (docs/SPEC.md section 5.5).
const (
	codeNotANumber     = "not_a_number"
	codeNotAnInteger   = "not_an_integer"
	codeNegative       = "negative"
	codeNotPositive    = "not_positive"
	codeOutOfRange     = "out_of_range"
	codeTooLong        = "too_long"
	codeInvalidEnum    = "invalid_enum"
	codeMinGtMax       = "min_gt_max"
	codeUnknownParam   = "unknown_param"
	codeMalformedQuery = "malformed_query"
)

// Parameter bounds and defaults (docs/SPEC.md section 5.3).
const (
	maxCityLen      = 100 // characters, after trimming
	maxKeywordLen   = 200 // characters, as received
	defaultPage     = 1
	defaultPageSize = 10
	// maxExactInt is 2^53. Every whole number below it is exact in a
	// float64; from here on some are not ("9007199254740993" parses as
	// 2^53), so an integer parameter this large or larger cannot be read
	// exactly and is out_of_range.
	maxExactInt = 1 << 53
)

// Field labels for messages: the BRANDING.md section 7.2 field labels, with
// plain words for the parameters the UI has no field for.
const (
	labelMinPrice     = "Min price"
	labelMaxPrice     = "Max price"
	labelMinBedrooms  = "Min bedrooms"
	labelCity         = "City"
	labelKeyword      = "Keyword"
	labelTargetBudget = "Target budget"
	labelPage         = "Page"
	labelPageSize     = "Page size"
)

// Field error messages that are fixed sentences. The first two are
// BRANDING.md section 7.2 verbatim; the rest are in the same voice. The
// bedrooms string in section 7.2 comes out of msgNegative.
const (
	msgMinGtMax       = "Min price must be less than max price."
	msgNotPositive    = "Target budget must be more than $0."
	msgInvalidStatus  = "Status must be active, pending or sold."
	msgPageTooSmall   = "Page must be 1 or more."
	msgUnknownParam   = "Unknown parameter."
	msgMalformedQuery = "Query string is not valid URL encoding."
)

// msgPageSizeRange is the pageSize out_of_range message: "Page size must be
// between 1 and 100."
var msgPageSizeRange = fmt.Sprintf("%s must be between 1 and %d.", labelPageSize, domain.MaxPageSize)

// Field error messages built from a field label.
func msgNotANumber(label string) string   { return label + " must be a number." }
func msgNotAnInteger(label string) string { return label + " must be a whole number." }
func msgNegative(label string) string     { return label + " must be 0 or more." }
func msgTooLarge(label string) string     { return label + " is too large." }

func msgTooLong(label string, limit int) string {
	return fmt.Sprintf("%s must be %d characters or fewer.", label, limit)
}

// SearchQuery is a parsed and validated search request (docs/SPEC.md
// section 5.3).
type SearchQuery struct {
	// Filter is what the store searches with. Statuses is never empty, and
	// KeywordTokens is nil when the keyword is absent.
	Filter domain.SearchFilter
	// TargetBudget is the target budget; nil means absent. When set it is
	// > 0, as domain.Rank requires.
	TargetBudget *float64
	// Page is the 1-based page number, 1 when absent.
	Page int
	// PageSize is the page size, within 1..domain.MaxPageSize; 10 when
	// absent.
	PageSize int
}

// ParseSearchQuery validates the decoded query parameters of a search
// request against docs/SPEC.md section 5.3. It collects every problem rather
// than stopping at the first, and returns nil errors exactly when the query
// is valid; otherwise the SearchQuery is the zero value.
//
// The rules: an empty value is absent. Scalar parameters use their first
// value. Numbers use the strconv.ParseFloat syntax, untrimmed, and must be
// finite; -0 reads as 0. Integers must also be whole and below 2^53. city is
// trimmed; city and keyword are absent when only whitespace, and otherwise
// limited to 100 and 200 characters. status values may repeat and hold
// comma-separated pieces, each trimmed and matched case-sensitively; the set
// comes back deduplicated in domain.AllStatuses order, or all of them when
// none is given. minPrice may equal maxPrice but not exceed it. A name
// outside SearchQueryParams is an unknown_param whatever its value.
//
// Errors are ordered: at most one per parameter, in SearchQueryParams order
// (min_gt_max is maxPrice's), then the unknown parameters sorted by name.
func ParseSearchQuery(v url.Values) (SearchQuery, []FieldError) {
	p := &parser{values: v}
	var q SearchQuery

	// Walk the parameters in table order, so the errors come out in it.
	q.Filter.MinPrice = p.nonNegative(paramMinPrice, labelMinPrice)
	q.Filter.MaxPrice = p.nonNegative(paramMaxPrice, labelMaxPrice)
	if lo, hi := q.Filter.MinPrice, q.Filter.MaxPrice; lo != nil && hi != nil && *lo > *hi {
		p.fail(paramMaxPrice, codeMinGtMax, msgMinGtMax)
	}
	q.Filter.MinBedrooms = p.minBedrooms()
	q.Filter.City = p.city()
	q.Filter.Keyword, q.Filter.KeywordTokens = p.keyword()
	q.TargetBudget = p.targetBudget()
	q.Filter.Statuses = p.statuses()
	q.Page = p.page()
	q.PageSize = p.pageSize()
	p.unknownParams()

	if len(p.errs) > 0 {
		return SearchQuery{}, p.errs
	}
	return q, nil
}

// parser reads parameters from values and collects their field errors.
// Every read method records at most one error, and returns the zero value
// (nil, "" or a default) for an absent or invalid parameter.
type parser struct {
	values url.Values
	errs   []FieldError
}

// fail records a field error.
func (p *parser) fail(field, code, message string) {
	p.errs = append(p.errs, FieldError{Field: field, Code: code, Message: message})
}

// scalar returns the first value of name; "" means absent.
func (p *parser) scalar(name string) string {
	if vs := p.values[name]; len(vs) > 0 {
		return vs[0]
	}
	return ""
}

// number reads name as a finite number. It returns nil when name is absent
// and records not_a_number when the value is not one.
func (p *parser) number(name, label string) *float64 {
	s := p.scalar(name)
	if s == "" {
		return nil
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil || math.IsNaN(f) || math.IsInf(f, 0) {
		p.fail(name, codeNotANumber, msgNotANumber(label))
		return nil
	}
	if f == 0 {
		f = 0 // -0 becomes 0, so appliedFilters never echoes -0
	}
	return &f
}

// whole reads name as a finite whole number, still as a float64 so the
// caller range-checks it before converting. It records not_an_integer for a
// fractional value.
func (p *parser) whole(name, label string) *float64 {
	f := p.number(name, label)
	if f != nil && math.Trunc(*f) != *f {
		p.fail(name, codeNotAnInteger, msgNotAnInteger(label))
		return nil
	}
	return f
}

// nonNegative reads a price bound: a number >= 0.
func (p *parser) nonNegative(name, label string) *float64 {
	f := p.number(name, label)
	if f != nil && *f < 0 {
		p.fail(name, codeNegative, msgNegative(label))
		return nil
	}
	return f
}

// minBedrooms reads minBedrooms: a whole number >= 0, below 2^53.
func (p *parser) minBedrooms() *int {
	f := p.whole(paramMinBedrooms, labelMinBedrooms)
	switch {
	case f == nil:
		return nil
	case *f < 0:
		p.fail(paramMinBedrooms, codeNegative, msgNegative(labelMinBedrooms))
		return nil
	case *f >= maxExactInt:
		p.fail(paramMinBedrooms, codeOutOfRange, msgTooLarge(labelMinBedrooms))
		return nil
	}
	n := int(*f)
	return &n
}

// city reads city, trimmed; "" means absent.
func (p *parser) city() string {
	s := strings.TrimSpace(p.scalar(paramCity))
	if utf8.RuneCountInString(s) > maxCityLen {
		p.fail(paramCity, codeTooLong, msgTooLong(labelCity, maxCityLen))
		return ""
	}
	return s
}

// keyword reads keyword as received, and its lowercased whitespace-separated
// tokens. A keyword that is only whitespace is absent, before the length
// rule applies. "%" and "_" are ordinary characters.
func (p *parser) keyword() (string, []string) {
	s := p.scalar(paramKeyword)
	if strings.TrimSpace(s) == "" {
		return "", nil
	}
	if utf8.RuneCountInString(s) > maxKeywordLen {
		p.fail(paramKeyword, codeTooLong, msgTooLong(labelKeyword, maxKeywordLen))
		return "", nil
	}
	return s, strings.Fields(strings.ToLower(s))
}

// targetBudget reads targetBudget: a number > 0, so domain.Rank's
// precondition holds.
func (p *parser) targetBudget() *float64 {
	f := p.number(paramTargetBudget, labelTargetBudget)
	if f != nil && *f <= 0 {
		p.fail(paramTargetBudget, codeNotPositive, msgNotPositive)
		return nil
	}
	return f
}

// statuses reads every status value, each split on commas. Pieces are
// trimmed and empty ones skipped; any piece that is not a status gives one
// invalid_enum error. The result is in domain.AllStatuses order without
// duplicates, and is all statuses when no piece is given.
func (p *parser) statuses() []domain.ListingStatus {
	want := make(map[domain.ListingStatus]bool)
	invalid := false
	for _, value := range p.values[paramStatus] {
		for piece := range strings.SplitSeq(value, ",") {
			piece = strings.TrimSpace(piece)
			if piece == "" {
				continue
			}
			status, err := domain.ParseStatus(piece)
			if err != nil {
				invalid = true
				continue
			}
			want[status] = true
		}
	}
	if invalid {
		p.fail(paramStatus, codeInvalidEnum, msgInvalidStatus)
		return nil
	}
	all := domain.AllStatuses()
	if len(want) == 0 {
		return all
	}
	return slices.DeleteFunc(all, func(s domain.ListingStatus) bool { return !want[s] })
}

// page reads page: a whole number >= 1, below 2^53; 1 when absent.
func (p *parser) page() int {
	f := p.whole(paramPage, labelPage)
	switch {
	case f == nil:
		return defaultPage
	case *f < 1:
		p.fail(paramPage, codeOutOfRange, msgPageTooSmall)
		return defaultPage
	case *f >= maxExactInt:
		p.fail(paramPage, codeOutOfRange, msgTooLarge(labelPage))
		return defaultPage
	}
	return int(*f)
}

// pageSize reads pageSize: a whole number in 1..domain.MaxPageSize; 10 when
// absent.
func (p *parser) pageSize() int {
	f := p.whole(paramPageSize, labelPageSize)
	switch {
	case f == nil:
		return defaultPageSize
	case *f < 1 || *f > domain.MaxPageSize:
		p.fail(paramPageSize, codeOutOfRange, msgPageSizeRange)
		return defaultPageSize
	}
	return int(*f)
}

// unknownParams records one unknown_param per name outside searchParams,
// sorted by name (byte order), so the output never depends on map order.
func (p *parser) unknownParams() {
	var unknown []string
	for name := range p.values {
		if !slices.Contains(searchParams[:], name) {
			unknown = append(unknown, name)
		}
	}
	slices.Sort(unknown)
	for _, name := range unknown {
		p.fail(name, codeUnknownParam, msgUnknownParam)
	}
}
