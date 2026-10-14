package httpapi_test

import (
	"fmt"
	"math"
	"net/url"
	"reflect"
	"slices"
	"strings"
	"testing"

	"brillio/backend/internal/domain"
	"brillio/backend/internal/httpapi"
)

// ptr returns a pointer to v.
func ptr[T any](v T) *T { return &v }

// parseQuery decodes raw as a query string, failing the test if it is not
// valid URL encoding, and runs ParseSearchQuery on it.
func parseQuery(t *testing.T, raw string) (httpapi.SearchQuery, []httpapi.FieldError) {
	t.Helper()
	v, err := url.ParseQuery(raw)
	if err != nil {
		t.Fatalf("url.ParseQuery(%q) error = %v", raw, err)
	}
	return httpapi.ParseSearchQuery(v)
}

// defaultQuery is the SearchQuery of a request with no parameters.
func defaultQuery() httpapi.SearchQuery {
	return httpapi.SearchQuery{
		Filter:   domain.SearchFilter{Statuses: domain.AllStatuses()},
		Page:     1,
		PageSize: 10,
	}
}

// fe builds a FieldError.
func fe(field, code, message string) httpapi.FieldError {
	return httpapi.FieldError{Field: field, Code: code, Message: message}
}

// Exact messages, so every rules-table string is asserted verbatim.
const (
	msgMinGtMax      = "Min price must be less than max price."
	msgPageSizeRange = "Page size must be between 1 and 100."
	msgUnknown       = "Unknown parameter."
	msgMalformed     = "Query string is not valid URL encoding."
)

// Strings at and over the city and keyword limits. "é" is two bytes, so the
// limits are counted in characters, not bytes.
var (
	city100        = strings.Repeat("c", 100)
	city101        = strings.Repeat("c", 101)
	cityWide100    = strings.Repeat("é", 100)
	cityWide101    = strings.Repeat("é", 101)
	keyword200     = strings.Repeat("k", 200)
	keyword201     = strings.Repeat("k", 201)
	keywordWide200 = strings.Repeat("é", 200)
	keywordWide201 = strings.Repeat("é", 201)
)

func TestParseSearchQueryDefaults(t *testing.T) {
	q, errs := httpapi.ParseSearchQuery(url.Values{})
	if errs != nil {
		t.Fatalf("errors = %v, want nil", errs)
	}
	if want := defaultQuery(); !reflect.DeepEqual(q, want) {
		t.Errorf("query = %+v, want %+v", q, want)
	}
	if q.Filter.KeywordTokens != nil {
		t.Errorf("KeywordTokens = %#v, want nil", q.Filter.KeywordTokens)
	}

	// A nil map is the same as an empty one.
	if q, errs := httpapi.ParseSearchQuery(nil); errs != nil || !reflect.DeepEqual(q, defaultQuery()) {
		t.Errorf("ParseSearchQuery(nil) = %+v, %v, want the defaults", q, errs)
	}
}

func TestParseSearchQueryRules(t *testing.T) {
	t.Run("invalid", func(t *testing.T) {
		tests := []struct {
			query string
			want  httpapi.FieldError
		}{
			// minPrice and maxPrice.
			{"minPrice=abc", fe("minPrice", "not_a_number", "Min price must be a number.")},
			{"maxPrice=abc", fe("maxPrice", "not_a_number", "Max price must be a number.")},
			{"minPrice=%205", fe("minPrice", "not_a_number", "Min price must be a number.")}, // untrimmed
			{"minPrice=5%20", fe("minPrice", "not_a_number", "Min price must be a number.")},
			{"minPrice=1__0", fe("minPrice", "not_a_number", "Min price must be a number.")},
			{"minPrice=NaN", fe("minPrice", "not_a_number", "Min price must be a number.")},
			{"minPrice=Inf", fe("minPrice", "not_a_number", "Min price must be a number.")},
			{"maxPrice=-Inf", fe("maxPrice", "not_a_number", "Max price must be a number.")},
			{"maxPrice=infinity", fe("maxPrice", "not_a_number", "Max price must be a number.")},
			{"minPrice=1e400", fe("minPrice", "not_a_number", "Min price must be a number.")},
			{"minPrice=-1", fe("minPrice", "negative", "Min price must be 0 or more.")},
			{"maxPrice=-0.01", fe("maxPrice", "negative", "Max price must be 0 or more.")},
			{"minPrice=5&maxPrice=1", fe("maxPrice", "min_gt_max", msgMinGtMax)},
			{"minPrice=450000.5&maxPrice=450000", fe("maxPrice", "min_gt_max", msgMinGtMax)},
			// minBedrooms.
			{"minBedrooms=two", fe("minBedrooms", "not_a_number", "Min bedrooms must be a number.")},
			{"minBedrooms=NaN", fe("minBedrooms", "not_a_number", "Min bedrooms must be a number.")},
			{"minBedrooms=2.5", fe("minBedrooms", "not_an_integer", "Min bedrooms must be a whole number.")},
			{"minBedrooms=-1", fe("minBedrooms", "negative", "Min bedrooms must be 0 or more.")},
			{"minBedrooms=-1.5", fe("minBedrooms", "not_an_integer", "Min bedrooms must be a whole number.")},
			{"minBedrooms=-1e300", fe("minBedrooms", "negative", "Min bedrooms must be 0 or more.")},          // sign first
			{"minBedrooms=9007199254740992", fe("minBedrooms", "out_of_range", "Min bedrooms is too large.")}, // 2^53
			{"minBedrooms=9007199254740993", fe("minBedrooms", "out_of_range", "Min bedrooms is too large.")}, // reads as 2^53
			{"minBedrooms=1e300", fe("minBedrooms", "out_of_range", "Min bedrooms is too large.")},
			// city and keyword.
			{"city=" + city101, fe("city", "too_long", "City must be 100 characters or fewer.")},
			{"city=" + cityWide101, fe("city", "too_long", "City must be 100 characters or fewer.")},
			{"keyword=" + keyword201, fe("keyword", "too_long", "Keyword must be 200 characters or fewer.")},
			// targetBudget.
			{"targetBudget=abc", fe("targetBudget", "not_a_number", "Target budget must be a number.")},
			{"targetBudget=0", fe("targetBudget", "not_positive", "Target budget must be more than $0.")},
			{"targetBudget=-0", fe("targetBudget", "not_positive", "Target budget must be more than $0.")},
			{"targetBudget=-5", fe("targetBudget", "not_positive", "Target budget must be more than $0.")},
			// status.
			{"status=foo", fe("status", "invalid_enum", "Status must be active, pending or sold.")},
			{"status=Active", fe("status", "invalid_enum", "Status must be active, pending or sold.")},
			// page.
			{"page=abc", fe("page", "not_a_number", "Page must be a number.")},
			{"page=1.5", fe("page", "not_an_integer", "Page must be a whole number.")},
			{"page=0", fe("page", "out_of_range", "Page must be 1 or more.")},
			{"page=-0", fe("page", "out_of_range", "Page must be 1 or more.")},
			{"page=-3", fe("page", "out_of_range", "Page must be 1 or more.")},
			{"page=9007199254740992", fe("page", "out_of_range", "Page is too large.")},
			{"page=1e300", fe("page", "out_of_range", "Page is too large.")},
			// pageSize.
			{"pageSize=abc", fe("pageSize", "not_a_number", "Page size must be a number.")},
			{"pageSize=2.5", fe("pageSize", "not_an_integer", "Page size must be a whole number.")},
			{"pageSize=0", fe("pageSize", "out_of_range", msgPageSizeRange)},
			{"pageSize=-1", fe("pageSize", "out_of_range", msgPageSizeRange)},
			{"pageSize=101", fe("pageSize", "out_of_range", msgPageSizeRange)},
			{"pageSize=1e300", fe("pageSize", "out_of_range", msgPageSizeRange)},
			// Unknown names, whatever the value; names are case-sensitive.
			{"minprice=1", fe("minprice", "unknown_param", msgUnknown)},
			{"minprice=", fe("minprice", "unknown_param", msgUnknown)},
			{"MinPrice=1", fe("MinPrice", "unknown_param", msgUnknown)},
			{"=1", fe("", "unknown_param", msgUnknown)},
			// Repeated scalars use the first value.
			{"minPrice=abc&minPrice=5", fe("minPrice", "not_a_number", "Min price must be a number.")},
			{"page=0&page=2", fe("page", "out_of_range", "Page must be 1 or more.")},
		}
		for _, tt := range tests {
			t.Run(short(tt.query), func(t *testing.T) {
				q, errs := parseQuery(t, tt.query)
				if want := []httpapi.FieldError{tt.want}; !reflect.DeepEqual(errs, want) {
					t.Fatalf("errors = %+v, want %+v", errs, want)
				}
				if !reflect.DeepEqual(q, httpapi.SearchQuery{}) {
					t.Errorf("query = %+v, want the zero SearchQuery alongside errors", q)
				}
				// BRANDING.md sections 7 and 8: a plain sentence, no
				// exclamation marks.
				if m := errs[0].Message; !strings.HasSuffix(m, ".") || strings.Contains(m, "!") {
					t.Errorf("message %q is not a plain sentence", m)
				}
			})
		}
	})

	t.Run("valid", func(t *testing.T) {
		tests := []struct {
			name  string
			query string
			edit  func(q *httpapi.SearchQuery) // changes from defaultQuery
		}{
			{"minPrice 0", "minPrice=0", func(q *httpapi.SearchQuery) { q.Filter.MinPrice = ptr(0.0) }},
			{"maxPrice 0", "maxPrice=0", func(q *httpapi.SearchQuery) { q.Filter.MaxPrice = ptr(0.0) }},
			{"min equals max", "minPrice=500000&maxPrice=500000", func(q *httpapi.SearchQuery) {
				q.Filter.MinPrice, q.Filter.MaxPrice = ptr(500000.0), ptr(500000.0)
			}},
			{"fractional price", "maxPrice=450000.75", func(q *httpapi.SearchQuery) { q.Filter.MaxPrice = ptr(450000.75) }},
			{"plus sign", "minPrice=%2B5", func(q *httpapi.SearchQuery) { q.Filter.MinPrice = ptr(5.0) }},
			{"underscores", "minPrice=1_000", func(q *httpapi.SearchQuery) { q.Filter.MinPrice = ptr(1000.0) }},
			{"hex float", "minPrice=0x1p4", func(q *httpapi.SearchQuery) { q.Filter.MinPrice = ptr(16.0) }},
			{"exponent", "maxPrice=4.5e5", func(q *httpapi.SearchQuery) { q.Filter.MaxPrice = ptr(450000.0) }},
			{"minBedrooms 0", "minBedrooms=0", func(q *httpapi.SearchQuery) { q.Filter.MinBedrooms = ptr(0) }},
			{"minBedrooms 2.0", "minBedrooms=2.0", func(q *httpapi.SearchQuery) { q.Filter.MinBedrooms = ptr(2) }},
			{"minBedrooms 1e2", "minBedrooms=1e2", func(q *httpapi.SearchQuery) { q.Filter.MinBedrooms = ptr(100) }},
			{"minBedrooms -0", "minBedrooms=-0", func(q *httpapi.SearchQuery) { q.Filter.MinBedrooms = ptr(0) }},
			{"minBedrooms 2^53-1", "minBedrooms=9007199254740991", func(q *httpapi.SearchQuery) {
				q.Filter.MinBedrooms = ptr(1<<53 - 1)
			}},
			{"page 1", "page=1", func(*httpapi.SearchQuery) {}},
			{"page 2^53-1", "page=9007199254740991", func(q *httpapi.SearchQuery) { q.Page = 1<<53 - 1 }},
			{"pageSize 1", "pageSize=1", func(q *httpapi.SearchQuery) { q.PageSize = 1 }},
			{"pageSize 100", "pageSize=100", func(q *httpapi.SearchQuery) { q.PageSize = 100 }},
			{"pageSize 1e2", "pageSize=1e2", func(q *httpapi.SearchQuery) { q.PageSize = 100 }},
			{"targetBudget", "targetBudget=450000", func(q *httpapi.SearchQuery) { q.TargetBudget = ptr(450000.0) }},
			{"tiny targetBudget", "targetBudget=0.01", func(q *httpapi.SearchQuery) { q.TargetBudget = ptr(0.01) }},
			{"city trimmed", "city=%20%20Reston%09", func(q *httpapi.SearchQuery) { q.Filter.City = "Reston" }},
			{"city case kept", "city=springFIELD", func(q *httpapi.SearchQuery) { q.Filter.City = "springFIELD" }},
			{"city 100", "city=" + city100, func(q *httpapi.SearchQuery) { q.Filter.City = city100 }},
			{"city 100 multibyte", "city=" + cityWide100, func(q *httpapi.SearchQuery) { q.Filter.City = cityWide100 }},
			{"city 100 after trim", "city=%20" + city100 + "%20", func(q *httpapi.SearchQuery) { q.Filter.City = city100 }},
			{"whitespace city absent", "city=%20%09%20", func(*httpapi.SearchQuery) {}},
			{"whitespace keyword absent", "keyword=%20%09%0A", func(*httpapi.SearchQuery) {}},
			{"every value empty", "minPrice=&maxPrice=&minBedrooms=&city=&keyword=&targetBudget=&status=&page=&pageSize=",
				func(*httpapi.SearchQuery) {}},
			{"name without value", "minPrice&page", func(*httpapi.SearchQuery) {}},
			{"repeated scalar uses first", "minPrice=1&minPrice=2&page=3&page=abc", func(q *httpapi.SearchQuery) {
				q.Filter.MinPrice, q.Page = ptr(1.0), 3
			}},
			{"empty first value is absent", "minPrice=&minPrice=5", func(*httpapi.SearchQuery) {}},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				q, errs := parseQuery(t, tt.query)
				if errs != nil {
					t.Fatalf("errors = %+v, want nil", errs)
				}
				want := defaultQuery()
				tt.edit(&want)
				if !reflect.DeepEqual(q, want) {
					t.Errorf("query = %s, want %s", dump(q), dump(want))
				}
			})
		}
	})

	t.Run("negative zero reads as zero", func(t *testing.T) {
		q, errs := parseQuery(t, "minPrice=-0&maxPrice=-0.0")
		if errs != nil {
			t.Fatalf("errors = %+v, want nil", errs)
		}
		for name, f := range map[string]*float64{"MinPrice": q.Filter.MinPrice, "MaxPrice": q.Filter.MaxPrice} {
			if f == nil || *f != 0 || math.Signbit(*f) {
				t.Errorf("%s = %v, want +0", name, f)
			}
		}
	})
}

// dump formats a SearchQuery with its pointers dereferenced.
func dump(q httpapi.SearchQuery) string {
	f := q.Filter
	return fmt.Sprintf("{minPrice:%v maxPrice:%v minBedrooms:%v city:%q keyword:%q tokens:%q "+
		"statuses:%v targetBudget:%v page:%d pageSize:%d}",
		deref(f.MinPrice), deref(f.MaxPrice), deref(f.MinBedrooms), f.City, f.Keyword, f.KeywordTokens,
		f.Statuses, deref(q.TargetBudget), q.Page, q.PageSize)
}

// deref returns *p, or nil when p is nil.
func deref[T any](p *T) any {
	if p == nil {
		return nil
	}
	return *p
}

// short returns s cut to 40 characters, for a subtest name.
func short(s string) string {
	if r := []rune(s); len(r) > 40 {
		return string(r[:40]) + "..."
	}
	return s
}

func TestParseSearchQueryCollectsAll(t *testing.T) {
	tests := []struct {
		name  string
		query string
		want  []httpapi.FieldError
	}{
		{
			name: "one error per parameter, table order, then unknowns by name",
			// Deliberately out of table order on the wire.
			query: "zeta=1&pageSize=0&page=0&status=foo&targetBudget=0&keyword=" + keyword201 +
				"&city=" + city101 + "&minBedrooms=2.5&maxPrice=-1&minPrice=abc&alpha=2&Beta=3",
			want: []httpapi.FieldError{
				fe("minPrice", "not_a_number", "Min price must be a number."),
				fe("maxPrice", "negative", "Max price must be 0 or more."),
				fe("minBedrooms", "not_an_integer", "Min bedrooms must be a whole number."),
				fe("city", "too_long", "City must be 100 characters or fewer."),
				fe("keyword", "too_long", "Keyword must be 200 characters or fewer."),
				fe("targetBudget", "not_positive", "Target budget must be more than $0."),
				fe("status", "invalid_enum", "Status must be active, pending or sold."),
				fe("page", "out_of_range", "Page must be 1 or more."),
				fe("pageSize", "out_of_range", msgPageSizeRange),
				// Byte order: uppercase sorts before lowercase.
				fe("Beta", "unknown_param", msgUnknown),
				fe("alpha", "unknown_param", msgUnknown),
				fe("zeta", "unknown_param", msgUnknown),
			},
		},
		{
			name:  "SPEC section 5.5 example",
			query: "minprice=1&pageSize=0&maxPrice=1&minPrice=5",
			want: []httpapi.FieldError{
				fe("maxPrice", "min_gt_max", msgMinGtMax),
				fe("pageSize", "out_of_range", msgPageSizeRange),
				fe("minprice", "unknown_param", msgUnknown),
			},
		},
		{
			name:  "min_gt_max sits in maxPrice's place",
			query: "minBedrooms=x&maxPrice=1&minPrice=5",
			want: []httpapi.FieldError{
				fe("maxPrice", "min_gt_max", msgMinGtMax),
				fe("minBedrooms", "not_a_number", "Min bedrooms must be a number."),
			},
		},
		{
			name:  "no min_gt_max when a bound is itself invalid",
			query: "minPrice=-5&maxPrice=-10",
			want: []httpapi.FieldError{
				fe("minPrice", "negative", "Min price must be 0 or more."),
				fe("maxPrice", "negative", "Max price must be 0 or more."),
			},
		},
		{
			name:  "a repeated unknown name is reported once",
			query: "foo=1&foo=2&bar=",
			want: []httpapi.FieldError{
				fe("bar", "unknown_param", msgUnknown),
				fe("foo", "unknown_param", msgUnknown),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Map iteration order varies per run; the output must not.
			for range 20 {
				_, errs := parseQuery(t, tt.query)
				if !reflect.DeepEqual(errs, tt.want) {
					t.Fatalf("errors =\n%+v\nwant\n%+v", errs, tt.want)
				}
			}
		})
	}
}

func TestParseSearchQueryStatus(t *testing.T) {
	const (
		active  = domain.StatusActive
		pending = domain.StatusPending
		sold    = domain.StatusSold
	)
	tests := []struct {
		query string
		want  []domain.ListingStatus // nil means one invalid_enum error
	}{
		{"status=pending", []domain.ListingStatus{pending}},
		{"status=sold,active", []domain.ListingStatus{active, sold}},
		{"status=active&status=pending", []domain.ListingStatus{active, pending}},
		{"status=pending,active", []domain.ListingStatus{active, pending}},
		{"status=sold&status=pending,active", []domain.ListingStatus{active, pending, sold}},
		{"status=%20pending%20,%09active", []domain.ListingStatus{active, pending}},
		{"status=active,active&status=active", []domain.ListingStatus{active}},
		{"status=,active,,", []domain.ListingStatus{active}},
		{"status=active&status=", []domain.ListingStatus{active}},
		{"status=,", []domain.ListingStatus{active, pending, sold}},
		{"status=%20", []domain.ListingStatus{active, pending, sold}},
		{"status=active,foo", nil},
		{"status=foo&status=bar,baz", nil},
		{"status=ACTIVE", nil},
	}
	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			q, errs := parseQuery(t, tt.query)
			if tt.want == nil {
				want := []httpapi.FieldError{fe("status", "invalid_enum", "Status must be active, pending or sold.")}
				if !reflect.DeepEqual(errs, want) {
					t.Errorf("errors = %+v, want %+v", errs, want)
				}
				return
			}
			if errs != nil {
				t.Fatalf("errors = %+v, want nil", errs)
			}
			if !slices.Equal(q.Filter.Statuses, tt.want) {
				t.Errorf("Statuses = %v, want %v", q.Filter.Statuses, tt.want)
			}
		})
	}

	t.Run("each call gets its own slice", func(t *testing.T) {
		q, _ := parseQuery(t, "")
		q.Filter.Statuses[0] = "mutated"
		if again, _ := parseQuery(t, ""); again.Filter.Statuses[0] != active {
			t.Errorf("Statuses[0] = %q after mutating an earlier result, want active", again.Filter.Statuses[0])
		}
	})
}

func TestParseSearchQueryKeyword(t *testing.T) {
	tests := []struct {
		name        string
		query       string
		wantKeyword string
		wantTokens  []string
		wantErr     bool
	}{
		{name: "one word", query: "keyword=Pets", wantKeyword: "Pets", wantTokens: []string{"pets"}},
		{name: "plus is a space", query: "keyword=NO+Pets", wantKeyword: "NO Pets", wantTokens: []string{"no", "pets"}},
		{name: "any whitespace splits, raw kept", query: "keyword=%20a%09B%0A%0Dc%20%20",
			wantKeyword: " a\tB\n\rc  ", wantTokens: []string{"a", "b", "c"}},
		{name: "percent is literal", query: "keyword=%25", wantKeyword: "%", wantTokens: []string{"%"}},
		{name: "underscore is literal", query: "keyword=_", wantKeyword: "_", wantTokens: []string{"_"}},
		{name: "mixed wildcards", query: "keyword=100%25_OFF", wantKeyword: "100%_OFF", wantTokens: []string{"100%_off"}},
		{name: "unicode lowercase", query: "keyword=%C3%89COLE", wantKeyword: "ÉCOLE", wantTokens: []string{"école"}},
		{name: "200 characters", query: "keyword=" + keyword200, wantKeyword: keyword200, wantTokens: []string{keyword200}},
		{name: "201 characters", query: "keyword=" + keyword201, wantErr: true},
		{name: "200 two-byte characters", query: "keyword=" + keywordWide200,
			wantKeyword: keywordWide200, wantTokens: []string{keywordWide200}},
		{name: "201 two-byte characters", query: "keyword=" + keywordWide201, wantErr: true},
		{name: "length counts the raw value", query: "keyword=%20" + keyword200, wantErr: true},
		// Absence is decided before the length rule.
		{name: "201 spaces is absent", query: "keyword=" + strings.Repeat("%20", 201)},
		{name: "empty", query: "keyword="},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q, errs := parseQuery(t, tt.query)
			if tt.wantErr {
				want := []httpapi.FieldError{fe("keyword", "too_long", "Keyword must be 200 characters or fewer.")}
				if !reflect.DeepEqual(errs, want) {
					t.Errorf("errors = %+v, want %+v", errs, want)
				}
				return
			}
			if errs != nil {
				t.Fatalf("errors = %+v, want nil", errs)
			}
			if q.Filter.Keyword != tt.wantKeyword {
				t.Errorf("Keyword = %q, want %q", q.Filter.Keyword, tt.wantKeyword)
			}
			if !reflect.DeepEqual(q.Filter.KeywordTokens, tt.wantTokens) {
				t.Errorf("KeywordTokens = %#v, want %#v", q.Filter.KeywordTokens, tt.wantTokens)
			}
		})
	}
}

func TestSearchQueryParams(t *testing.T) {
	want := []string{
		"minPrice", "maxPrice", "minBedrooms", "city", "keyword",
		"targetBudget", "status", "page", "pageSize",
	}
	got := httpapi.SearchQueryParams()
	if !slices.Equal(got, want) {
		t.Fatalf("SearchQueryParams() = %v, want %v (docs/SPEC.md section 5.3 order)", got, want)
	}

	got[0] = "mutated"
	if again := httpapi.SearchQueryParams(); !slices.Equal(again, want) {
		t.Errorf("SearchQueryParams() = %v after mutating an earlier result, want %v", again, want)
	}

	// The parser accepts exactly these names.
	v := url.Values{}
	for _, name := range want {
		v.Set(name, "")
	}
	if _, errs := httpapi.ParseSearchQuery(v); errs != nil {
		t.Errorf("every listed name, empty: errors = %+v, want nil", errs)
	}
}
