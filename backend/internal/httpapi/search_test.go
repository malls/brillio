package httpapi_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"brillio/backend/internal/domain"
	"brillio/backend/internal/httpapi"
	"brillio/backend/internal/store"
	"brillio/backend/internal/store/seed"
)

// The SQLite store is the production Searcher and CityLister.
var (
	_ httpapi.Searcher   = (*store.Store)(nil)
	_ httpapi.CityLister = (*store.Store)(nil)
)

// Repository files the tests read.
const (
	seedPath = "../../../data/sample_listings.json"
	specPath = "../../../docs/SPEC.md"
)

// testNow is the fixed request clock. Its UTC date, the asOf of every
// expected score here, is 2026-10-13.
var testNow = time.Date(2026, 10, 13, 15, 4, 5, 0, time.UTC)

// allStatuses is appliedFilters.status when no status is given.
var allStatuses = []string{"active", "pending", "sold"}

// The wire types mirror docs/SPEC.md sections 5.2, 5.4 and 5.5 by hand,
// independently of the handler's DTOs. Nothing is omitempty, so
// strictRoundTrip proves every key in both directions.

type wireSearchResponse struct {
	Items          []wireItem  `json:"items"`
	Page           wirePage    `json:"page"`
	Scoring        wireScoring `json:"scoring"`
	AppliedFilters wireFilters `json:"appliedFilters"`
}

type wireItem struct {
	Rank        int           `json:"rank"`
	ID          string        `json:"id"`
	Source      string        `json:"source"`
	Address     string        `json:"address"`
	City        string        `json:"city"`
	State       string        `json:"state"`
	Zip         string        `json:"zip"`
	Price       float64       `json:"price"`
	Bedrooms    int           `json:"bedrooms"`
	Bathrooms   float64       `json:"bathrooms"`
	Sqft        int           `json:"sqft"`
	Latitude    float64       `json:"latitude"`
	Longitude   float64       `json:"longitude"`
	ListedDate  string        `json:"listedDate"`
	Status      string        `json:"status"`
	Description string        `json:"description"`
	Relevance   wireRelevance `json:"relevance"`
}

type wireRelevance struct {
	Score     float64  `json:"score"`
	BudgetFit *float64 `json:"budgetFit"`
	Recency   float64  `json:"recency"`
}

type wirePage struct {
	Page       int `json:"page"`
	PageSize   int `json:"pageSize"`
	TotalItems int `json:"totalItems"`
	TotalPages int `json:"totalPages"`
}

type wireScoring struct {
	TargetBudget        *float64 `json:"targetBudget"`
	BudgetWeight        float64  `json:"budgetWeight"`
	RecencyWeight       float64  `json:"recencyWeight"`
	RecencyHalfLifeDays int      `json:"recencyHalfLifeDays"`
	AsOf                string   `json:"asOf"`
}

type wireFilters struct {
	MinPrice    *float64 `json:"minPrice"`
	MaxPrice    *float64 `json:"maxPrice"`
	MinBedrooms *int     `json:"minBedrooms"`
	City        *string  `json:"city"`
	Keyword     *string  `json:"keyword"`
	Status      []string `json:"status"`
}

// wireEnvelope is a 400 validation_error body; fields is always present.
type wireEnvelope struct {
	Error struct {
		Code    string           `json:"code"`
		Message string           `json:"message"`
		Fields  []wireFieldError `json:"fields"`
	} `json:"error"`
}

type wireFieldError struct {
	Field   string `json:"field"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

type wireCities struct {
	Cities []string `json:"cities"`
}

// strictRoundTrip decodes body into v and proves the two match key for key:
// decoding rejects any key v lacks (and trailing data), and re-encoding v
// must give back exactly body's keys and values, so v has no key the body
// lacks. The second direction also catches a key that only matched v's tag
// case-insensitively.
func strictRoundTrip(t *testing.T, body []byte, v any) {
	t.Helper()
	dec := json.NewDecoder(bytes.NewReader(body))
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		t.Fatalf("strict decode into %T: %v: %s", v, err, body)
	}
	if err := dec.Decode(new(any)); err != io.EOF {
		t.Fatalf("data after the JSON value: %v: %s", err, body)
	}
	again, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("re-encode %T: %v", v, err)
	}
	var want, got any
	if err := json.Unmarshal(body, &want); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if err := json.Unmarshal(again, &got); err != nil {
		t.Fatalf("decode re-encoded body: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("round trip through %T differs:\n got %s\nwant %s", v, again, body)
	}
}

// fakeSearcher is a test Searcher that records its calls.
type fakeSearcher struct {
	listings []domain.Listing
	err      error
	calls    int
	got      domain.SearchFilter
}

func (f *fakeSearcher) Search(_ context.Context, filter domain.SearchFilter) ([]domain.Listing, error) {
	f.calls++
	f.got = filter
	return f.listings, f.err
}

// fakeCities is a test CityLister.
type fakeCities struct {
	cities []string
	err    error
}

func (f fakeCities) Cities(_ context.Context) ([]string, error) {
	return f.cities, f.err
}

// openStore opens an empty ":memory:" store and closes it at cleanup.
func openStore(t *testing.T) *store.Store {
	t.Helper()
	st, err := store.Open(t.Context(), ":memory:")
	if err != nil {
		t.Fatalf("store.Open() error = %v", err)
	}
	t.Cleanup(func() {
		if err := st.Close(); err != nil {
			t.Errorf("Close() error = %v", err)
		}
	})
	return st
}

// seededStore returns a ":memory:" store holding the sample listings.
func seededStore(t *testing.T) *store.Store {
	t.Helper()
	listings, err := seed.Load(seedPath)
	if err != nil {
		t.Fatalf("seed.Load() error = %v", err)
	}
	st := openStore(t)
	if _, err := st.Upsert(t.Context(), listings); err != nil {
		t.Fatalf("Upsert() error = %v", err)
	}
	return st
}

// newAPIRouter returns a router over deps with the clock fixed at now and
// logs captured as JSON lines in the returned buffer.
func newAPIRouter(deps httpapi.Deps, now time.Time) (http.Handler, *bytes.Buffer) {
	var buf bytes.Buffer
	deps.Logger = slog.New(slog.NewJSONHandler(&buf, nil))
	deps.CORSOrigins = []string{testOrigin}
	deps.Now = func() time.Time { return now }
	return httpapi.NewRouter(deps), &buf
}

// newSearchRouter returns a router over a seeded ":memory:" store with the
// clock fixed at now.
func newSearchRouter(t *testing.T, now time.Time) (http.Handler, *bytes.Buffer) {
	t.Helper()
	st := seededStore(t)
	return newAPIRouter(httpapi.Deps{Pinger: st, Searcher: st, Cities: st}, now)
}

// get sends GET target through h and checks what every JSON response
// shares: the status, the JSON content type and a request ID.
func get(t *testing.T, h http.Handler, target string, wantStatus int) *httptest.ResponseRecorder {
	t.Helper()
	rec := serve(h, http.MethodGet, target)
	if rec.Code != wantStatus {
		t.Fatalf("GET %s: status = %d, want %d (body %s)", target, rec.Code, wantStatus, rec.Body)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json; charset=utf-8" {
		t.Errorf("GET %s: Content-Type = %q, want application/json; charset=utf-8", target, ct)
	}
	assertRequestID(t, rec)
	return rec
}

// search runs a search that must succeed and returns its strictly decoded
// body and the raw bytes. Every item's listing fields must equal the seed
// record's.
func search(t *testing.T, h http.Handler, query string) (wireSearchResponse, []byte) {
	t.Helper()
	rec := get(t, h, "/api/listings/search?"+query, http.StatusOK)
	var resp wireSearchResponse
	strictRoundTrip(t, rec.Body.Bytes(), &resp)
	assertItemsMatchSeed(t, rec.Body.Bytes())
	return resp, rec.Body.Bytes()
}

// searchInvalid runs a search that must fail validation and returns the
// field errors and the raw body.
func searchInvalid(t *testing.T, h http.Handler, query string) ([]wireFieldError, []byte) {
	t.Helper()
	rec := get(t, h, "/api/listings/search?"+query, http.StatusBadRequest)
	var env wireEnvelope
	strictRoundTrip(t, rec.Body.Bytes(), &env)
	if env.Error.Code != "validation_error" || env.Error.Message != "Invalid search parameters." {
		t.Errorf("error = {%q, %q}, want {validation_error, Invalid search parameters.}",
			env.Error.Code, env.Error.Message)
	}
	if len(env.Error.Fields) == 0 {
		t.Error("validation_error has no fields, want at least one")
	}
	return env.Error.Fields, rec.Body.Bytes()
}

// seedRecords returns the seed file's records as decoded JSON objects, keyed
// by source and id. They are read straight from the file, not through the
// seed loader, so they are an independent copy of the data.
func seedRecords(t *testing.T) map[[2]string]map[string]any {
	t.Helper()
	data, err := os.ReadFile(seedPath)
	if err != nil {
		t.Fatalf("read seed: %v", err)
	}
	var records []map[string]any
	if err := json.Unmarshal(data, &records); err != nil {
		t.Fatalf("decode seed: %v", err)
	}
	out := make(map[[2]string]map[string]any, len(records))
	for _, r := range records {
		out[[2]string{r["source"].(string), r["id"].(string)}] = r
	}
	return out
}

// assertItemsMatchSeed checks that each item in a search body is its seed
// record, value for value, plus rank and relevance.
func assertItemsMatchSeed(t *testing.T, body []byte) {
	t.Helper()
	var resp struct {
		Items []map[string]any `json:"items"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		t.Fatalf("decode items: %v", err)
	}
	records := seedRecords(t)
	for _, item := range resp.Items {
		key := [2]string{item["source"].(string), item["id"].(string)}
		delete(item, "rank")
		delete(item, "relevance")
		if !reflect.DeepEqual(item, records[key]) {
			t.Errorf("item %v:\n got %v\nwant %v", key, item, records[key])
		}
	}
}

// specExample returns the first fenced JSON block after the docs/SPEC.md
// heading that starts with heading.
func specExample(t *testing.T, heading string) []byte {
	t.Helper()
	data, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatalf("read SPEC: %v", err)
	}
	_, rest, ok := strings.Cut(string(data), "\n"+heading)
	if !ok {
		t.Fatalf("SPEC has no heading %q", heading)
	}
	_, rest, ok = strings.Cut(rest, "```json\n")
	if !ok {
		t.Fatalf("SPEC %q has no JSON example", heading)
	}
	block, _, ok := strings.Cut(rest, "```")
	if !ok {
		t.Fatalf("SPEC %q JSON example is not closed", heading)
	}
	return []byte(block)
}

// assertStrictBody strict-round-trips a 200 or 400 body into its wire type.
// Other statuses are left to the caller.
func assertStrictBody(t *testing.T, rec *httptest.ResponseRecorder, target string) {
	t.Helper()
	switch {
	case rec.Code == http.StatusBadRequest:
		strictRoundTrip(t, rec.Body.Bytes(), new(wireEnvelope))
	case rec.Code == http.StatusOK && strings.HasPrefix(target, "/api/cities"):
		strictRoundTrip(t, rec.Body.Bytes(), new(wireCities))
	case rec.Code == http.StatusOK:
		strictRoundTrip(t, rec.Body.Bytes(), new(wireSearchResponse))
	}
}

// wantItem is the expected rank, identity and relevance of one result.
type wantItem struct {
	rank      int
	id        string
	score     float64
	budgetFit *float64
	recency   float64
}

// assertItems compares items with want, exactly and in order.
func assertItems(t *testing.T, items []wireItem, want []wantItem) {
	t.Helper()
	if len(items) != len(want) {
		t.Fatalf("got %d items %v, want %d", len(items), itemIDs(items), len(want))
	}
	for i, w := range want {
		got := items[i]
		if got.Rank != w.rank || got.ID != w.id {
			t.Errorf("item %d = rank %d %s, want rank %d %s", i, got.Rank, got.ID, w.rank, w.id)
		}
		wantRel := wireRelevance{Score: w.score, BudgetFit: w.budgetFit, Recency: w.recency}
		if !reflect.DeepEqual(got.Relevance, wantRel) {
			t.Errorf("%s relevance = %s, want %s", got.ID, fmtRelevance(got.Relevance), fmtRelevance(wantRel))
		}
	}
}

// fmtRelevance formats r with its budget fit dereferenced.
func fmtRelevance(r wireRelevance) string {
	return fmt.Sprintf("{score %v, budgetFit %v, recency %v}", r.Score, deref(r.BudgetFit), r.Recency)
}

// itemIDs returns each item's id, in order.
func itemIDs(items []wireItem) []string {
	ids := make([]string, len(items))
	for i, it := range items {
		ids[i] = it.ID
	}
	return ids
}

// itemRanks returns each item's rank, in order.
func itemRanks(items []wireItem) []int {
	ranks := make([]int, len(items))
	for i, it := range items {
		ranks[i] = it.Rank
	}
	return ranks
}

func TestSearchRankedSpringfield(t *testing.T) {
	h, _ := newSearchRouter(t, testNow)
	resp, _ := search(t, h, "targetBudget=450000&city=Springfield")

	// Ages at 2026-10-13: A1 45 days, B7 47, A2 41, B8 44. Recency is
	// 0.5^(age/30); over budget the fit is 1 - 2*(price-450000)/450000; the
	// score is 0.6*fit + 0.4*recency, compared unrounded.
	assertItems(t, resp.Items, []wantItem{
		{rank: 1, id: "A1", score: 0.7414, budgetFit: ptr(1.0), recency: 0.3536},
		{rank: 2, id: "B7", score: 0.7297, budgetFit: ptr(0.9911), recency: 0.3376},
		{rank: 3, id: "A2", score: 0.5551, budgetFit: ptr(0.6667), recency: 0.3878},
		{rank: 4, id: "B8", score: 0.5381, budgetFit: ptr(0.6556), recency: 0.3618},
	})
	for _, it := range resp.Items {
		t.Logf("rank %d %s/%s price %v listed %s: score %v budgetFit %v recency %v",
			it.Rank, it.Source, it.ID, it.Price, it.ListedDate,
			it.Relevance.Score, deref(it.Relevance.BudgetFit), it.Relevance.Recency)
	}

	wantA1 := wireItem{
		Rank: 1, ID: "A1", Source: "MLS_A",
		Address: "123 Main St, Apt 4B", City: "Springfield", State: "VA", Zip: "22150",
		Price: 450000, Bedrooms: 2, Bathrooms: 1.5, Sqft: 980,
		Latitude: 38.7893, Longitude: -77.1873,
		ListedDate: "2026-08-29", Status: "active",
		Description: "Bright top-floor condo near shops and transit. Pet friendly.",
		Relevance:   wireRelevance{Score: 0.7414, BudgetFit: ptr(1.0), Recency: 0.3536},
	}
	if len(resp.Items) > 0 && !reflect.DeepEqual(resp.Items[0], wantA1) {
		t.Errorf("first item =\n%+v\nwant\n%+v", resp.Items[0], wantA1)
	}
	if want := (wirePage{Page: 1, PageSize: 10, TotalItems: 4, TotalPages: 1}); resp.Page != want {
		t.Errorf("page = %+v, want %+v", resp.Page, want)
	}
	wantScoring := wireScoring{TargetBudget: ptr(450000.0), BudgetWeight: 0.6, RecencyWeight: 0.4,
		RecencyHalfLifeDays: 30, AsOf: "2026-10-13"}
	if !reflect.DeepEqual(resp.Scoring, wantScoring) {
		t.Errorf("scoring = %+v, want %+v", resp.Scoring, wantScoring)
	}
	wantFilters := wireFilters{City: ptr("Springfield"), Status: allStatuses}
	if !reflect.DeepEqual(resp.AppliedFilters, wantFilters) {
		t.Errorf("appliedFilters = %+v, want %+v", resp.AppliedFilters, wantFilters)
	}
}

func TestSearchRankedAll(t *testing.T) {
	h, _ := newSearchRouter(t, testNow)

	t.Run("SPEC section 5.4 example", func(t *testing.T) {
		// The example is the first item of targetBudget=450000 with nothing
		// else, at asOf 2026-10-13.
		_, body := search(t, h, "targetBudget=450000")
		var got map[string]any
		if err := json.Unmarshal(body, &got); err != nil {
			t.Fatal(err)
		}
		got["items"] = got["items"].([]any)[:1]
		var want map[string]any
		if err := json.Unmarshal(specExample(t, "### 5.4"), &want); err != nil {
			t.Fatalf("decode SPEC example: %v", err)
		}
		if !reflect.DeepEqual(got, want) {
			trimmed, _ := json.Marshal(got)
			t.Errorf("response (first item only) =\n%s\nwant the SPEC section 5.4 example", trimmed)
		}
	})

	t.Run("full order", func(t *testing.T) {
		resp, _ := search(t, h, "targetBudget=450000&pageSize=12")
		assertItems(t, resp.Items, []wantItem{
			{rank: 1, id: "A1", score: 0.7414, budgetFit: ptr(1.0), recency: 0.3536},
			{rank: 2, id: "B7", score: 0.7297, budgetFit: ptr(0.9911), recency: 0.3376},
			{rank: 3, id: "A5", score: 0.7091, budgetFit: ptr(0.9111), recency: 0.4061},
			{rank: 4, id: "A3", score: 0.6836, budgetFit: ptr(0.8867), recency: 0.3789},
			{rank: 5, id: "B11", score: 0.6623, budgetFit: ptr(0.9333), recency: 0.2558},
			{rank: 6, id: "B9", score: 0.6616, budgetFit: ptr(0.8878), recency: 0.3223},
			{rank: 7, id: "A6", score: 0.6445, budgetFit: ptr(0.9222), recency: 0.2279},
			{rank: 8, id: "A2", score: 0.5551, budgetFit: ptr(0.6667), recency: 0.3878},
			{rank: 9, id: "B8", score: 0.5381, budgetFit: ptr(0.6556), recency: 0.3618},
			{rank: 10, id: "A7", score: 0.4275, budgetFit: ptr(0.6), recency: 0.1688},
			{rank: 11, id: "B10", score: 0.3549, budgetFit: ptr(0.4), recency: 0.2872},
			{rank: 12, id: "A4", score: 0.3321, budgetFit: ptr(0.2889), recency: 0.3969},
		})
	})
}

func TestSearchNoBudget(t *testing.T) {
	h, _ := newSearchRouter(t, testNow)
	resp, body := search(t, h, "pageSize=12")

	// Without a budget the score is recency, so the order is newest first.
	recency := []struct {
		id  string
		rec float64
	}{
		{"A5", 0.4061}, {"A4", 0.3969}, {"A2", 0.3878}, {"A3", 0.3789}, {"B8", 0.3618}, {"A1", 0.3536},
		{"B7", 0.3376}, {"B9", 0.3223}, {"B10", 0.2872}, {"B11", 0.2558}, {"A6", 0.2279}, {"A7", 0.1688},
	}
	want := make([]wantItem, len(recency))
	for i, r := range recency {
		want[i] = wantItem{rank: i + 1, id: r.id, score: r.rec, recency: r.rec}
	}
	assertItems(t, resp.Items, want)

	if got := strings.Count(string(body), `"budgetFit":null`); got != 12 {
		t.Errorf(`body has %d "budgetFit":null, want 12`, got)
	}
	wantScoring := wireScoring{BudgetWeight: 0, RecencyWeight: 1, RecencyHalfLifeDays: 30, AsOf: "2026-10-13"}
	if !reflect.DeepEqual(resp.Scoring, wantScoring) {
		t.Errorf("scoring = %+v, want %+v", resp.Scoring, wantScoring)
	}
	if !strings.Contains(string(body), `"targetBudget":null,"budgetWeight":0,"recencyWeight":1,`) {
		t.Errorf("scoring is not encoded as null, 0, 1: %s", body)
	}
	if want := (wireFilters{Status: allStatuses}); !reflect.DeepEqual(resp.AppliedFilters, want) {
		t.Errorf("appliedFilters = %+v, want %+v", resp.AppliedFilters, want)
	}

	// No parameters at all: the first page of the same order.
	resp, _ = search(t, h, "")
	if got, want := itemIDs(resp.Items), []string{"A5", "A4", "A2", "A3", "B8", "A1", "B7", "B9", "B10", "B11"}; !slices.Equal(got, want) {
		t.Errorf("default page ids = %v, want %v", got, want)
	}
	if want := (wirePage{Page: 1, PageSize: 10, TotalItems: 12, TotalPages: 2}); resp.Page != want {
		t.Errorf("default page = %+v, want %+v", resp.Page, want)
	}
}

func TestSearchAsOfUTC(t *testing.T) {
	// 23:30 on the 13th at UTC-7 is 06:30 on the 14th in UTC: ages and asOf
	// follow the UTC date.
	now := time.Date(2026, 10, 13, 23, 30, 0, 0, time.FixedZone("UTC-7", -7*60*60))
	h, _ := newSearchRouter(t, now)
	resp, _ := search(t, h, "targetBudget=450000&city=Springfield")

	if resp.Scoring.AsOf != "2026-10-14" {
		t.Errorf("asOf = %q, want 2026-10-14", resp.Scoring.AsOf)
	}
	// A1 is 46 days old: recency 0.5^(46/30) = 0.3455, score 0.6 + 0.4*0.3455.
	assertItems(t, resp.Items[:1], []wantItem{{rank: 1, id: "A1", score: 0.7382, budgetFit: ptr(1.0), recency: 0.3455}})
}

func TestSearchPaging(t *testing.T) {
	h, _ := newSearchRouter(t, testNow)
	tests := []struct {
		query     string
		wantPage  wirePage
		wantRanks []int
		wantIDs   []string
	}{
		{"pageSize=3&page=2", wirePage{2, 3, 12, 4}, []int{4, 5, 6}, []string{"A3", "B8", "A1"}},
		{"page=2", wirePage{2, 10, 12, 2}, []int{11, 12}, []string{"A6", "A7"}},
		{"pageSize=1&page=12", wirePage{12, 1, 12, 12}, []int{12}, []string{"A7"}},
		{"pageSize=100", wirePage{1, 100, 12, 1}, []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12},
			[]string{"A5", "A4", "A2", "A3", "B8", "A1", "B7", "B9", "B10", "B11", "A6", "A7"}},
		// Past the end: 200 with no items and the real totals.
		{"page=99", wirePage{99, 10, 12, 2}, []int{}, []string{}},
		{"city=Springfield&page=2", wirePage{2, 10, 4, 1}, []int{}, []string{}},
	}
	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			resp, body := search(t, h, tt.query)
			if resp.Page != tt.wantPage {
				t.Errorf("page = %+v, want %+v", resp.Page, tt.wantPage)
			}
			if got := itemRanks(resp.Items); !slices.Equal(got, tt.wantRanks) {
				t.Errorf("ranks = %v, want %v", got, tt.wantRanks)
			}
			if got := itemIDs(resp.Items); !slices.Equal(got, tt.wantIDs) {
				t.Errorf("ids = %v, want %v", got, tt.wantIDs)
			}
			if len(tt.wantIDs) == 0 && !bytes.Contains(body, []byte(`"items":[]`)) {
				t.Errorf(`body lacks "items":[]: %s`, body)
			}
		})
	}
}

func TestSearchEmptyCity(t *testing.T) {
	h, _ := newSearchRouter(t, testNow)

	resp, body := search(t, h, "city=Atlantis")
	if want := (wirePage{Page: 1, PageSize: 10, TotalItems: 0, TotalPages: 0}); resp.Page != want {
		t.Errorf("Atlantis page = %+v, want %+v", resp.Page, want)
	}
	if !bytes.Contains(body, []byte(`"items":[]`)) {
		t.Errorf(`Atlantis body lacks "items":[]: %s`, body)
	}
	if c := resp.AppliedFilters.City; c == nil || *c != "Atlantis" {
		t.Errorf("Atlantis appliedFilters.city = %v, want Atlantis", deref(c))
	}

	// Trimmed, matched ignoring case, and echoed as trimmed.
	resp, _ = search(t, h, "city=%20springfield%20")
	if got, want := itemIDs(resp.Items), []string{"A2", "B8", "A1", "B7"}; !slices.Equal(got, want) {
		t.Errorf("springfield ids = %v, want %v", got, want)
	}
	if c := resp.AppliedFilters.City; c == nil || *c != "springfield" {
		t.Errorf("springfield appliedFilters.city = %v, want springfield", deref(c))
	}
}

func TestSearchStatusFilter(t *testing.T) {
	h, _ := newSearchRouter(t, testNow)
	tests := []struct {
		query      string
		wantTotal  int
		wantIDs    []string // checked when set
		wantStatus []string
	}{
		{"status=pending", 1, []string{"A7"}, []string{"pending"}},
		{"status=sold", 0, []string{}, []string{"sold"}},
		{"status=active", 11, nil, []string{"active"}},
		{"status=active&status=pending", 12, nil, []string{"active", "pending"}},
		{"status=pending,active", 12, nil, []string{"active", "pending"}},
	}
	bodies := map[string][]byte{}
	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			resp, body := search(t, h, tt.query)
			bodies[tt.query] = body
			if resp.Page.TotalItems != tt.wantTotal {
				t.Errorf("totalItems = %d, want %d", resp.Page.TotalItems, tt.wantTotal)
			}
			if got := itemIDs(resp.Items); tt.wantIDs != nil && !slices.Equal(got, tt.wantIDs) {
				t.Errorf("ids = %v, want %v", got, tt.wantIDs)
			}
			if !slices.Equal(resp.AppliedFilters.Status, tt.wantStatus) {
				t.Errorf("appliedFilters.status = %v, want %v", resp.AppliedFilters.Status, tt.wantStatus)
			}
		})
	}
	if a, b := bodies["status=active&status=pending"], bodies["status=pending,active"]; !bytes.Equal(a, b) {
		t.Errorf("repeated and comma-separated status differ:\n%s\n%s", a, b)
	}
}

func TestSearchKeyword(t *testing.T) {
	h, _ := newSearchRouter(t, testNow)
	tests := []struct {
		query       string
		wantIDs     []string
		wantKeyword string
	}{
		// No description holds "%" or "_": they are literal, not wildcards.
		{"keyword=%25", []string{}, "%"},
		{"keyword=_", []string{}, "_"},
		{"keyword=pets", []string{"A5", "A4", "A3", "B7", "B9", "A6", "A7"}, "pets"},
		// Every token must appear, ignoring case; results in rank order.
		{"keyword=NO+Pets", []string{"A5", "A3", "B9", "A7"}, "NO Pets"},
	}
	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			resp, _ := search(t, h, tt.query)
			if got := itemIDs(resp.Items); !slices.Equal(got, tt.wantIDs) {
				t.Errorf("ids = %v, want %v", got, tt.wantIDs)
			}
			if k := resp.AppliedFilters.Keyword; k == nil || *k != tt.wantKeyword {
				t.Errorf("appliedFilters.keyword = %v, want %q", deref(k), tt.wantKeyword)
			}
		})
	}
}

func TestSearchNegativeZeroEcho(t *testing.T) {
	h, _ := newSearchRouter(t, testNow)
	resp, body := search(t, h, "minPrice=-0&maxPrice=-0")
	if !bytes.Contains(body, []byte(`"minPrice":0,"maxPrice":0,`)) || bytes.Contains(body, []byte("-0")) {
		t.Errorf(`body does not echo -0 as 0: %s`, body)
	}
	if f := resp.AppliedFilters; f.MinPrice == nil || *f.MinPrice != 0 || f.MaxPrice == nil || *f.MaxPrice != 0 {
		t.Errorf("appliedFilters prices = %v, %v, want 0, 0", deref(f.MinPrice), deref(f.MaxPrice))
	}
	if resp.Page.TotalItems != 0 {
		t.Errorf("totalItems = %d, want 0 (no listing costs $0)", resp.Page.TotalItems)
	}
}

func TestSearchValidationErrors(t *testing.T) {
	h, _ := newSearchRouter(t, testNow)
	long := func(c string, n int) string { return strings.Repeat(c, n) }
	malformed := wireFieldError{"query", "malformed_query", "Query string is not valid URL encoding."}
	tests := []struct {
		name  string
		query string
		want  []wireFieldError
	}{
		{"not_a_number", "minPrice=abc", []wireFieldError{{"minPrice", "not_a_number", "Min price must be a number."}}},
		{"not_an_integer", "minBedrooms=2.5",
			[]wireFieldError{{"minBedrooms", "not_an_integer", "Min bedrooms must be a whole number."}}},
		{"negative", "minPrice=-1", []wireFieldError{{"minPrice", "negative", "Min price must be 0 or more."}}},
		{"not_positive", "targetBudget=0",
			[]wireFieldError{{"targetBudget", "not_positive", "Target budget must be more than $0."}}},
		{"out_of_range page", "page=0", []wireFieldError{{"page", "out_of_range", "Page must be 1 or more."}}},
		{"out_of_range pageSize", "pageSize=0",
			[]wireFieldError{{"pageSize", "out_of_range", "Page size must be between 1 and 100."}}},
		{"too_long city", "city=" + long("c", 101),
			[]wireFieldError{{"city", "too_long", "City must be 100 characters or fewer."}}},
		{"too_long keyword", "keyword=" + long("k", 201),
			[]wireFieldError{{"keyword", "too_long", "Keyword must be 200 characters or fewer."}}},
		{"invalid_enum", "status=foo", []wireFieldError{{"status", "invalid_enum", "Status must be active, pending or sold."}}},
		{"min_gt_max", "minPrice=5&maxPrice=1",
			[]wireFieldError{{"maxPrice", "min_gt_max", "Min price must be less than max price."}}},
		{"unknown_param", "minprice=1", []wireFieldError{{"minprice", "unknown_param", "Unknown parameter."}}},
		{"malformed_query bare percent", "keyword=100%", []wireFieldError{malformed}},
		{"malformed_query bad escape in a name", "%zz=1", []wireFieldError{malformed}},
		{"malformed_query semicolon", "minPrice=1;maxPrice=2", []wireFieldError{malformed}},
		{"one malformed_query per request", "keyword=%&city=%zz", []wireFieldError{malformed}},
		{"decoded pairs are still validated", "keyword=100%&city=" + long("c", 101) + "&minprice=1&minPrice=abc",
			[]wireFieldError{
				malformed,
				{"minPrice", "not_a_number", "Min price must be a number."},
				{"city", "too_long", "City must be 100 characters or fewer."},
				{"minprice", "unknown_param", "Unknown parameter."},
			}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := searchInvalid(t, h, tt.query)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("fields =\n%+v\nwant\n%+v", got, tt.want)
			}
		})
	}

	t.Run("SPEC section 5.5 example", func(t *testing.T) {
		_, body := searchInvalid(t, h, "minPrice=5&maxPrice=1&pageSize=0&minprice=1")
		var want bytes.Buffer
		if err := json.Compact(&want, specExample(t, "### 5.5")); err != nil {
			t.Fatalf("compact SPEC example: %v", err)
		}
		want.WriteByte('\n')
		if !bytes.Equal(body, want.Bytes()) {
			t.Errorf("body =\n%s\nwant the SPEC section 5.5 example\n%s", body, want.Bytes())
		}
	})
}

func TestSearchStoreFailure(t *testing.T) {
	const detail = "disk on fire: secret detail"
	fake := &fakeSearcher{err: errors.New(detail)}
	h, logs := newAPIRouter(httpapi.Deps{Searcher: fake, Cities: fakeCities{}}, testNow)

	rec := get(t, h, "/api/listings/search?city=Reston&status=sold", http.StatusInternalServerError)
	const wantBody = `{"error":{"code":"internal_error","message":"Internal server error."}}` + "\n"
	if rec.Body.String() != wantBody {
		t.Errorf("body = %q, want %q", rec.Body, wantBody)
	}
	if fake.calls != 1 || fake.got.City != "Reston" ||
		!slices.Equal(fake.got.Statuses, []domain.ListingStatus{domain.StatusSold}) {
		t.Errorf("Search calls = %d with %+v, want one with the parsed filter", fake.calls, fake.got)
	}

	failures := linesWithMsg(logLines(t, logs), "request failed")
	if len(failures) != 1 {
		t.Fatalf("got %d request failed log lines, want 1: %s", len(failures), logs)
	}
	if l := failures[0]; l["level"] != "ERROR" || l["err"] != detail || l["request_id"] != rec.Header().Get("X-Request-ID") {
		t.Errorf("request failed log = %v, want level ERROR, err %q and the response's request ID", l, detail)
	}

	// An invalid request is answered before the store is asked.
	get(t, h, "/api/listings/search?page=0", http.StatusBadRequest)
	if fake.calls != 1 {
		t.Errorf("Search calls = %d after a 400, want still 1", fake.calls)
	}
}

func TestSearchMethodAndCORS(t *testing.T) {
	h, _ := newSearchRouter(t, testNow)
	preflight := []string{
		"Origin", testOrigin,
		"Access-Control-Request-Method", "GET",
		"Access-Control-Request-Headers", "X-Request-ID",
	}
	tests := []struct {
		name       string
		method     string
		target     string
		headers    []string
		wantStatus int
		want       map[string]string // exact header values
	}{
		{name: "POST search", method: http.MethodPost, target: "/api/listings/search",
			wantStatus: http.StatusMethodNotAllowed, want: map[string]string{"Allow": "GET, HEAD"}},
		{name: "DELETE cities", method: http.MethodDelete, target: "/api/cities",
			wantStatus: http.StatusMethodNotAllowed, want: map[string]string{"Allow": "GET, HEAD"}},
		{name: "preflight search", method: http.MethodOptions, target: "/api/listings/search?city=Reston",
			headers: preflight, wantStatus: http.StatusNoContent, want: map[string]string{
				"Access-Control-Allow-Origin":  testOrigin,
				"Access-Control-Allow-Methods": "GET, OPTIONS",
				"Access-Control-Allow-Headers": "Content-Type, X-Request-ID",
			}},
		{name: "cross-origin search", method: http.MethodGet, target: "/api/listings/search",
			headers: []string{"Origin", testOrigin}, wantStatus: http.StatusOK, want: map[string]string{
				"Access-Control-Allow-Origin":   testOrigin,
				"Access-Control-Expose-Headers": "X-Request-ID",
			}},
		{name: "cross-origin cities", method: http.MethodGet, target: "/api/cities",
			headers: []string{"Origin", testOrigin}, wantStatus: http.StatusOK, want: map[string]string{
				"Access-Control-Allow-Origin":   testOrigin,
				"Access-Control-Expose-Headers": "X-Request-ID",
			}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := serve(h, tt.method, tt.target, tt.headers...)
			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d (body %s)", rec.Code, tt.wantStatus, rec.Body)
			}
			for name, want := range tt.want {
				if got := rec.Header().Get(name); got != want {
					t.Errorf("%s = %q, want %q", name, got, want)
				}
			}
			assertStrictBody(t, rec, tt.target)
			switch tt.wantStatus {
			case http.StatusMethodNotAllowed:
				if env := decodeEnvelope(t, rec.Body.Bytes()); env.Error.Code != "method_not_allowed" {
					t.Errorf("error code = %q, want method_not_allowed", env.Error.Code)
				}
			case http.StatusNoContent:
				if rec.Body.Len() != 0 {
					t.Errorf("preflight body = %q, want empty", rec.Body)
				}
			}
		})
	}
}

func TestSearchRequestID(t *testing.T) {
	h, _ := newSearchRouter(t, testNow)
	failing, _ := newAPIRouter(httpapi.Deps{
		Searcher: &fakeSearcher{err: errors.New("search down")},
		Cities:   fakeCities{err: errors.New("cities down")},
	}, testNow)
	tests := []struct {
		name       string
		h          http.Handler
		method     string
		target     string
		wantStatus int
	}{
		{"search 200", h, http.MethodGet, "/api/listings/search", http.StatusOK},
		{"search 400", h, http.MethodGet, "/api/listings/search?minprice=1", http.StatusBadRequest},
		{"search 405", h, http.MethodPost, "/api/listings/search", http.StatusMethodNotAllowed},
		{"search 500", failing, http.MethodGet, "/api/listings/search", http.StatusInternalServerError},
		{"cities 200", h, http.MethodGet, "/api/cities", http.StatusOK},
		{"cities 500", failing, http.MethodGet, "/api/cities", http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := serve(tt.h, tt.method, tt.target)
			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d (body %s)", rec.Code, tt.wantStatus, rec.Body)
			}
			if id := assertRequestID(t, rec); !hexID.MatchString(id) {
				t.Errorf("X-Request-ID = %q, want a generated 32-hex ID", id)
			}
			assertStrictBody(t, rec, tt.target)
			rec = serve(tt.h, tt.method, tt.target, "X-Request-ID", "bc5-check")
			if id := rec.Header().Get("X-Request-ID"); id != "bc5-check" {
				t.Errorf("inbound X-Request-ID echoed as %q, want bc5-check", id)
			}
		})
	}
}
