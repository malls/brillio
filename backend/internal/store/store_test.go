package store_test

import (
	"cmp"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"brillio/backend/internal/domain"
	"brillio/backend/internal/store"
	"brillio/backend/internal/store/seed"
)

// samplePath is the repository's real seed file.
const samplePath = "../../../data/sample_listings.json"

// sampleCount is the number of listings in samplePath.
const sampleCount = 12

// openStore opens a migrated store at path and closes it at cleanup.
func openStore(t *testing.T, path string) *store.Store {
	t.Helper()
	s, err := store.Open(t.Context(), path)
	if err != nil {
		t.Fatalf("Open(%q) error = %v", path, err)
	}
	t.Cleanup(func() {
		if err := s.Close(); err != nil {
			t.Errorf("Close() error = %v", err)
		}
	})
	return s
}

// loadSample returns the listings in samplePath.
func loadSample(t *testing.T) []domain.Listing {
	t.Helper()
	listings, err := seed.Load(samplePath)
	if err != nil {
		t.Fatalf("seed.Load() error = %v", err)
	}
	return listings
}

// upsert writes listings into s and asserts the reported count.
func upsert(t *testing.T, s *store.Store, listings ...domain.Listing) {
	t.Helper()
	n, err := s.Upsert(t.Context(), listings)
	if err != nil {
		t.Fatalf("Upsert() error = %v", err)
	}
	if n != len(listings) {
		t.Fatalf("Upsert() = %d, want %d", n, len(listings))
	}
}

// newSeededStore returns a ":memory:" store seeded with the sample data.
func newSeededStore(t *testing.T) *store.Store {
	t.Helper()
	s := openStore(t, ":memory:")
	upsert(t, s, loadSample(t)...)
	return s
}

// search runs Search and fails the test on error.
func search(t *testing.T, s *store.Store, f domain.SearchFilter) []domain.Listing {
	t.Helper()
	got, err := s.Search(t.Context(), f)
	if err != nil {
		t.Fatalf("Search(%+v) error = %v", f, err)
	}
	return got
}

// ids returns each listing's ID, in order.
func ids(listings []domain.Listing) []string {
	out := make([]string, len(listings))
	for i, l := range listings {
		out[i] = l.ID
	}
	return out
}

// testListing returns a valid listing in source TEST.
func testListing(id, city, description string) domain.Listing {
	return domain.Listing{
		ID: id, Source: "TEST", Address: "1 Test Way", City: city, State: "VA", Zip: "22000",
		Price: 400000, Bedrooms: 3, Bathrooms: 2, Sqft: 1500, Latitude: 38.8, Longitude: -77.2,
		ListedDate: time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC), Status: domain.StatusActive,
		Description: description,
	}
}

// sameListing reports whether a and b are equal, comparing ListedDate as an
// instant.
func sameListing(a, b domain.Listing) bool {
	if !a.ListedDate.Equal(b.ListedDate) {
		return false
	}
	a.ListedDate, b.ListedDate = time.Time{}, time.Time{}
	return a == b
}

func TestSeedTwiceIdempotent(t *testing.T) {
	s := openStore(t, ":memory:")
	listings := loadSample(t)
	for round := 1; round <= 2; round++ {
		upsert(t, s, listings...)
		if got := len(search(t, s, domain.SearchFilter{})); got != sampleCount {
			t.Errorf("after seed %d: %d rows, want %d", round, got, sampleCount)
		}
	}
}

func TestRoundTrip(t *testing.T) {
	s := newSeededStore(t)
	want := loadSample(t)
	slices.SortFunc(want, func(a, b domain.Listing) int {
		return cmp.Or(cmp.Compare(a.Source, b.Source), cmp.Compare(a.ID, b.ID))
	})
	got := search(t, s, domain.SearchFilter{})
	if len(got) != len(want) {
		t.Fatalf("Search() = %d rows, want %d", len(got), len(want))
	}
	for i := range want {
		if !sameListing(got[i], want[i]) {
			t.Errorf("row %d =\n%+v\nwant\n%+v", i, got[i], want[i])
		}
	}
}

func TestUpsertUpdates(t *testing.T) {
	s := newSeededStore(t)
	changed := domain.Listing{
		ID: "A1", Source: "MLS_A", Address: "9 New Rd", City: "SPRINGFIELD", State: "MD",
		Zip: "20001", Price: 1, Bedrooms: 7, Bathrooms: 3.5, Sqft: 4000, Latitude: 1.5,
		Longitude: -2.5, ListedDate: time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC),
		Status: domain.StatusSold, Description: "Changed.",
	}
	upsert(t, s, changed)

	all := search(t, s, domain.SearchFilter{})
	if len(all) != sampleCount {
		t.Fatalf("after re-upserting A1: %d rows, want %d", len(all), sampleCount)
	}
	idx := slices.IndexFunc(all, func(l domain.Listing) bool { return l.Source == "MLS_A" && l.ID == "A1" })
	if idx < 0 {
		t.Fatal("MLS_A/A1 missing after update")
	}
	if !sameListing(all[idx], changed) {
		t.Errorf("A1 =\n%+v\nwant every non-key column updated:\n%+v", all[idx], changed)
	}
}

func TestUpsertAtomic(t *testing.T) {
	s := openStore(t, ":memory:")
	bad := testListing("BAD", "Vienna", "Bad status.")
	bad.Status = "gone"
	_, err := s.Upsert(t.Context(), []domain.Listing{testListing("OK", "Vienna", "Fine."), bad})
	if err == nil || !strings.Contains(err.Error(), "upsert TEST/BAD") {
		t.Fatalf("Upsert() error = %v, want one naming upsert TEST/BAD", err)
	}
	if got := search(t, s, domain.SearchFilter{}); len(got) != 0 {
		t.Errorf("after a failed Upsert: rows %q, want none (rolled back)", ids(got))
	}
}

func TestSearchFilters(t *testing.T) {
	s := newSeededStore(t)
	fp := func(v float64) *float64 { return &v }
	ip := func(v int) *int { return &v }
	statuses := func(ss ...domain.ListingStatus) []domain.ListingStatus { return ss }
	all := []string{"A1", "A2", "A3", "A4", "A5", "A6", "A7", "B10", "B11", "B7", "B8", "B9"}
	active := []string{"A1", "A2", "A3", "A4", "A5", "A6", "B10", "B11", "B7", "B8", "B9"}

	tests := []struct {
		name   string
		filter domain.SearchFilter
		want   []string
	}{
		{"no filters", domain.SearchFilter{}, all},

		{"minPrice", domain.SearchFilter{MinPrice: fp(500000)}, []string{"A2", "A4", "A7", "B10", "B8"}},
		{"minPrice inclusive", domain.SearchFilter{MinPrice: fp(450000)},
			[]string{"A1", "A2", "A4", "A5", "A7", "B10", "B11", "B7", "B8"}},
		{"maxPrice inclusive", domain.SearchFilter{MaxPrice: fp(450000)}, []string{"A1", "A3", "A6", "B9"}},
		{"price range", domain.SearchFilter{MinPrice: fp(450000), MaxPrice: fp(470000)},
			[]string{"A1", "A5", "B11", "B7"}},
		{"maxPrice zero is a bound", domain.SearchFilter{MaxPrice: fp(0)}, []string{}},
		{"minPrice zero", domain.SearchFilter{MinPrice: fp(0)}, all},

		{"minBedrooms", domain.SearchFilter{MinBedrooms: ip(4)}, []string{"A4", "A7"}},
		{"minBedrooms zero", domain.SearchFilter{MinBedrooms: ip(0)}, all},

		{"city lowercase", domain.SearchFilter{City: "springfield"}, []string{"A1", "A2", "B7", "B8"}},
		{"city padded uppercase", domain.SearchFilter{City: "  FAIRFAX  "}, []string{"A3", "B9"}},
		{"city blank is absent", domain.SearchFilter{City: "   "}, all},
		{"city is exact, not prefix", domain.SearchFilter{City: "Spring"}, []string{}},

		{"keyword", domain.SearchFilter{KeywordTokens: []string{"pets"}},
			[]string{"A3", "A4", "A5", "A6", "A7", "B7", "B9"}},
		{"keyword uppercase token", domain.SearchFilter{KeywordTokens: []string{"PETS"}},
			[]string{"A3", "A4", "A5", "A6", "A7", "B7", "B9"}},
		{"keyword hyphenated", domain.SearchFilter{KeywordTokens: []string{"top-floor"}}, []string{"A1"}},
		{"keyword tokens AND", domain.SearchFilter{KeywordTokens: []string{"fenced", "kitchen"}},
			[]string{"A2", "B8"}},
		{"keyword tokens AND, any order", domain.SearchFilter{KeywordTokens: []string{"metro", "quiet"}},
			[]string{"A5"}},
		{"keyword tokens AND, no listing has both", domain.SearchFilter{KeywordTokens: []string{"fenced", "pool"}},
			[]string{}},
		{"keyword percent is literal", domain.SearchFilter{KeywordTokens: []string{"%"}}, []string{}},
		{"keyword underscore is literal", domain.SearchFilter{KeywordTokens: []string{"_"}}, []string{}},
		{"keyword blank tokens skipped", domain.SearchFilter{KeywordTokens: []string{"", "  "}}, all},

		{"status pending", domain.SearchFilter{Statuses: statuses(domain.StatusPending)}, []string{"A7"}},
		{"status sold", domain.SearchFilter{Statuses: statuses(domain.StatusSold)}, []string{}},
		{"status active", domain.SearchFilter{Statuses: statuses(domain.StatusActive)}, active},
		{"status active, pending", domain.SearchFilter{Statuses: statuses(domain.StatusActive, domain.StatusPending)}, all},
		{"status pending, sold", domain.SearchFilter{Statuses: statuses(domain.StatusPending, domain.StatusSold)},
			[]string{"A7"}},
		{"status all", domain.SearchFilter{Statuses: domain.AllStatuses()}, all},
		{"status nil means all", domain.SearchFilter{Statuses: nil}, all},

		{"combined city, bedrooms, maxPrice", domain.SearchFilter{
			City: "Springfield", MinBedrooms: ip(3), MaxPrice: fp(526000),
		}, []string{"A2"}},
		{"combined range, bedrooms, keyword, active", domain.SearchFilter{
			MinPrice: fp(400000), MaxPrice: fp(550000), MinBedrooms: ip(3),
			Keyword: "pets", KeywordTokens: []string{"pets"}, Statuses: statuses(domain.StatusActive),
		}, []string{"A5", "A6"}},
		{"combined range, bedrooms, keyword, active or pending", domain.SearchFilter{
			MinPrice: fp(400000), MaxPrice: fp(550000), MinBedrooms: ip(3),
			Keyword: "pets", KeywordTokens: []string{"pets"},
			Statuses: statuses(domain.StatusActive, domain.StatusPending),
		}, []string{"A5", "A6", "A7"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := search(t, s, tt.filter)
			if got == nil {
				t.Fatal("Search() returned a nil slice, want non-nil")
			}
			if !slices.Equal(ids(got), tt.want) {
				t.Errorf("Search() = %q, want %q", ids(got), tt.want)
			}
		})
	}
}

func TestSearchLiteralWildcards(t *testing.T) {
	s := newSeededStore(t)
	upsert(t, s, testListing("T1", "Vienna", "100% move-in ready, pet_friendly"))
	// With LIKE, "%" and "_" would be wildcards and "pet_" would also match
	// the sample's "pets".
	for _, tokens := range [][]string{{"100%"}, {"%"}, {"_"}, {"pet_"}, {"PET_FRIENDLY", "100%"}} {
		t.Run(strings.Join(tokens, " "), func(t *testing.T) {
			got := search(t, s, domain.SearchFilter{KeywordTokens: tokens})
			if len(got) != 1 || got[0].Source != "TEST" || got[0].ID != "T1" {
				t.Errorf("Search(%q) = %q, want exactly TEST/T1", tokens, ids(got))
			}
		})
	}
}

func TestSearchInjection(t *testing.T) {
	s := newSeededStore(t)
	tests := []struct {
		name   string
		filter domain.SearchFilter
	}{
		{"city", domain.SearchFilter{City: "Springfield' OR '1'='1"}},
		{"keyword token", domain.SearchFilter{KeywordTokens: []string{"') or 1=1 --"}}},
		{"keyword drop table", domain.SearchFilter{KeywordTokens: []string{"'); DROP TABLE listings; --"}}},
		{"status", domain.SearchFilter{Statuses: []domain.ListingStatus{"active' OR '1'='1"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := search(t, s, tt.filter)
			if got == nil || len(got) != 0 {
				t.Errorf("Search() = %#v, want a non-nil empty slice", ids(got))
			}
		})
	}
	if got := len(search(t, s, domain.SearchFilter{})); got != sampleCount {
		t.Errorf("after injection attempts: %d rows, want %d", got, sampleCount)
	}
}

func TestCities(t *testing.T) {
	sample := []string{"Chantilly", "Fairfax", "Manassas", "Reston", "Springfield", "Vienna"}
	tests := []struct {
		name   string
		seeded bool             // start from the sample data
		add    []domain.Listing // upserted in order after any seeding
		want   []string
	}{
		{name: "sample", seeded: true, want: sample},
		{name: "later casing variant keeps first-seen casing", seeded: true,
			add: []domain.Listing{testListing("T2", "SPRINGFIELD", "x")}, want: sample},
		{name: "first seen lowercase",
			add:  []domain.Listing{testListing("T1", "vienna", "x"), testListing("T2", "Vienna", "x")},
			want: []string{"vienna"}},
		{name: "first seen capitalized",
			add:  []domain.Listing{testListing("T1", "Vienna", "x"), testListing("T2", "vienna", "x")},
			want: []string{"Vienna"}},
		{name: "sorted ignoring case",
			add:  []domain.Listing{testListing("T1", "Vienna", "x"), testListing("T2", "arlington", "x")},
			want: []string{"arlington", "Vienna"}},
		{name: "empty city excluded",
			add:  []domain.Listing{testListing("T1", "", "x"), testListing("T2", "Reston", "x")},
			want: []string{"Reston"}},
		{name: "empty store", want: []string{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := openStore(t, ":memory:")
			if tt.seeded {
				upsert(t, s, loadSample(t)...)
			}
			if len(tt.add) > 0 {
				upsert(t, s, tt.add...)
			}
			got, err := s.Cities(t.Context())
			if err != nil {
				t.Fatalf("Cities() error = %v", err)
			}
			if got == nil {
				t.Fatal("Cities() returned a nil slice, want non-nil")
			}
			if !slices.Equal(got, tt.want) {
				t.Errorf("Cities() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPing(t *testing.T) {
	s, err := store.Open(t.Context(), ":memory:")
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	if err := s.Ping(t.Context()); err != nil {
		t.Errorf("Ping() on an open store = %v, want nil", err)
	}
	if err := s.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	if err := s.Ping(t.Context()); err == nil {
		t.Error("Ping() after Close = nil, want an error")
	}
}

func TestReopenFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "listings.db")
	listings := loadSample(t)

	first, err := store.Open(t.Context(), path)
	if err != nil {
		t.Fatalf("first Open() error = %v", err)
	}
	upsert(t, first, listings...)
	if err := first.Close(); err != nil {
		t.Fatalf("first Close() error = %v", err)
	}

	second := openStore(t, path) // migrations already applied: a no-op
	if got := len(search(t, second, domain.SearchFilter{})); got != sampleCount {
		t.Errorf("after reopening: %d rows, want %d", got, sampleCount)
	}
	upsert(t, second, listings...)
	if got := len(search(t, second, domain.SearchFilter{})); got != sampleCount {
		t.Errorf("after reseeding the reopened file: %d rows, want %d", got, sampleCount)
	}
}

func TestOpenErrors(t *testing.T) {
	missingDir := filepath.Join(t.TempDir(), "missing", "listings.db")
	tests := []struct {
		name string
		path string
		want string // substring of the error
	}{
		{name: "parent directory missing", path: missingDir, want: missingDir},
		{name: "query string", path: "x.db?mode=ro", want: "DB_PATH must be a file path or :memory:"},
		{name: "file URI", path: "file:x.db", want: "DB_PATH must be a file path or :memory:"},
		{name: "empty", path: "", want: "DB_PATH must be a file path or :memory:"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, err := store.Open(t.Context(), tt.path)
			if err == nil {
				_ = s.Close()
				t.Fatalf("Open(%q) error = nil, want one containing %q", tt.path, tt.want)
			}
			t.Logf("Open(%q) error: %v", tt.path, err)
			if !strings.Contains(err.Error(), tt.want) {
				t.Errorf("Open(%q) error = %q, want it to contain %q", tt.path, err, tt.want)
			}
		})
	}
}
