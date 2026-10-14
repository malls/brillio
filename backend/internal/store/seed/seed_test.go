package seed_test

import (
	"encoding/json"
	"errors"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"brillio/backend/internal/domain"
	"brillio/backend/internal/store/seed"
)

// samplePath is the repository's real seed file.
const samplePath = "../../../../data/sample_listings.json"

// writeFile writes data to a file in a fresh temporary directory and returns
// its path.
func writeFile(t *testing.T, data string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "listings.json")
	if err := os.WriteFile(path, []byte(data), 0o600); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
	return path
}

// writeRecords writes elements as a JSON array and returns the file path.
// Elements are maps (or other plain values), never domain.Listing.
func writeRecords(t *testing.T, elements []any) string {
	t.Helper()
	data, err := json.Marshal(elements)
	if err != nil {
		t.Fatalf("marshal records: %v", err)
	}
	return writeFile(t, string(data))
}

// validRecord returns a valid wire record with the given id.
func validRecord(id string) map[string]any {
	return map[string]any{
		"id": id, "source": "MLS_T", "address": "1 Test Way", "city": "Testville",
		"state": "VA", "zip": "22000", "price": 400000.0, "bedrooms": 3, "bathrooms": 2.5,
		"sqft": 1500, "latitude": 38.8, "longitude": -77.2, "listedDate": "2026-09-01",
		"status": "active", "description": "A test listing.",
	}
}

func TestLoadSample(t *testing.T) {
	listings, err := seed.Load(samplePath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	wantIDs := []string{"A1", "B7", "A2", "B8", "A3", "B9", "A4", "B10", "A5", "B11", "A6", "A7"}
	if len(listings) != len(wantIDs) {
		t.Fatalf("Load() returned %d listings, want %d", len(listings), len(wantIDs))
	}
	for i, l := range listings {
		if l.ID != wantIDs[i] {
			t.Errorf("listing %d ID = %q, want %q (file order)", i, l.ID, wantIDs[i])
		}
	}

	got := listings[0]
	if !got.ListedDate.Equal(time.Date(2026, 8, 29, 0, 0, 0, 0, time.UTC)) || got.ListedDate.Location() != time.UTC {
		t.Errorf("A1 ListedDate = %v, want 2026-08-29 00:00 UTC", got.ListedDate)
	}
	got.ListedDate = time.Time{}
	want := domain.Listing{
		ID: "A1", Source: "MLS_A", Address: "123 Main St, Apt 4B", City: "Springfield",
		State: "VA", Zip: "22150", Price: 450000, Bedrooms: 2, Bathrooms: 1.5, Sqft: 980,
		Latitude: 38.7893, Longitude: -77.1873, Status: domain.StatusActive,
		Description: "Bright top-floor condo near shops and transit. Pet friendly.",
	}
	if got != want {
		t.Errorf("A1 = %+v, want %+v", got, want)
	}
	if listings[11].Status != domain.StatusPending {
		t.Errorf("A7 Status = %q, want pending", listings[11].Status)
	}
}

func TestLoadNormalizes(t *testing.T) {
	r := validRecord(" X1 ")
	r["source"] = "\tMLS_T "
	r["address"] = "1  Main St "
	r["city"] = "  Falls \t  Church "
	r["state"] = " va "
	r["zip"] = " 22046 "
	r["listedDate"] = " 2026-08-29 "
	r["status"] = " pending "
	r["description"] = "  Roomy.  Quiet.  "

	listings, err := seed.Load(writeRecords(t, []any{r}))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	got := listings[0]
	checks := []struct{ field, got, want string }{
		{"ID", got.ID, "X1"},
		{"Source", got.Source, "MLS_T"},
		{"Address", got.Address, "1  Main St"}, // internal runs kept: only city and state collapse
		{"City", got.City, "Falls Church"},
		{"State", got.State, "VA"},
		{"Zip", got.Zip, "22046"},
		{"Status", string(got.Status), "pending"},
		{"Description", got.Description, "Roomy.  Quiet."},
	}
	for _, c := range checks {
		if c.got != c.want {
			t.Errorf("%s = %q, want %q", c.field, c.got, c.want)
		}
	}
	if !got.ListedDate.Equal(time.Date(2026, 8, 29, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("ListedDate = %v, want 2026-08-29 UTC", got.ListedDate)
	}
}

func TestLoadRecordErrors(t *testing.T) {
	// with returns a valid record (id X3) changed by set: a nil value is a
	// JSON null, and keys in drop are removed.
	with := func(set map[string]any, drop ...string) map[string]any {
		r := validRecord("X3")
		maps.Copy(r, set)
		for _, k := range drop {
			delete(r, k)
		}
		return r
	}
	tests := []struct {
		name      string
		bad       any    // the element at index 2
		wantField string // RecordError.Field
		wantMsg   string // substring of the error message
		wantIs    error  // optional errors.Is target
	}{
		{name: "missing bedrooms", bad: with(nil, "bedrooms"), wantField: "bedrooms", wantMsg: "missing or null"},
		{name: "null price", bad: with(map[string]any{"price": nil}), wantField: "price", wantMsg: "missing or null"},
		{name: "unknown key", bad: with(map[string]any{"garage": true}), wantMsg: `"garage"`},
		{name: "status gone", bad: with(map[string]any{"status": "gone"}), wantField: "status",
			wantMsg: `invalid listing status: "gone"`, wantIs: domain.ErrInvalidStatus},
		{name: "status wrong case", bad: with(map[string]any{"status": "Active"}), wantField: "status",
			wantIs: domain.ErrInvalidStatus},
		{name: "date month 13", bad: with(map[string]any{"listedDate": "2026-13-01"}), wantField: "listedDate"},
		{name: "date US format", bad: with(map[string]any{"listedDate": "08/29/2026"}), wantField: "listedDate"},
		{name: "date with time", bad: with(map[string]any{"listedDate": "2026-08-29T00:00:00Z"}), wantField: "listedDate"},
		{name: "negative price", bad: with(map[string]any{"price": -1}), wantField: "price", wantMsg: ">= 0"},
		{name: "negative bedrooms", bad: with(map[string]any{"bedrooms": -1}), wantField: "bedrooms", wantMsg: ">= 0"},
		{name: "negative sqft", bad: with(map[string]any{"sqft": -1}), wantField: "sqft", wantMsg: ">= 0"},
		{name: "fractional bedrooms", bad: with(map[string]any{"bedrooms": 2.5}), wantField: "bedrooms"},
		{name: "string bedrooms", bad: with(map[string]any{"bedrooms": "2"}), wantField: "bedrooms"},
		{name: "blank id", bad: with(map[string]any{"id": "   "}), wantField: "id", wantMsg: "empty"},
		{name: "empty source", bad: with(map[string]any{"source": ""}), wantField: "source", wantMsg: "empty"},
		{name: "string element", bad: "x", wantMsg: "expected a JSON object"},
		{name: "null element", bad: nil, wantMsg: "expected a JSON object"},
		{name: "duplicate (source, id)", bad: validRecord(" X1 "), wantMsg: "record index 0"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := writeRecords(t, []any{validRecord("X1"), validRecord("X2"), tt.bad})
			listings, err := seed.Load(path)
			if err == nil {
				t.Fatalf("Load() = %d listings, nil error; want a record error", len(listings))
			}
			t.Logf("error: %v", err)
			var re *seed.RecordError
			if !errors.As(err, &re) {
				t.Fatalf("error %q is not a *seed.RecordError", err)
			}
			if re.Index != 2 || re.Field != tt.wantField {
				t.Errorf("RecordError Index, Field = %d, %q; want 2, %q", re.Index, re.Field, tt.wantField)
			}
			msg := err.Error()
			if !strings.HasPrefix(msg, path+": record index 2: ") {
				t.Errorf("error %q does not start with the path and record index 2", msg)
			}
			if tt.wantField != "" && !strings.Contains(msg, `field "`+tt.wantField+`"`) {
				t.Errorf("error %q does not name field %q", msg, tt.wantField)
			}
			if !strings.Contains(msg, tt.wantMsg) {
				t.Errorf("error %q does not contain %q", msg, tt.wantMsg)
			}
			if tt.wantIs != nil && !errors.Is(err, tt.wantIs) {
				t.Errorf("errors.Is(%q, %v) = false", err, tt.wantIs)
			}
		})
	}
}

func TestLoadZeroValuesValid(t *testing.T) {
	r := validRecord("X1")
	r["price"], r["bedrooms"], r["sqft"] = 0, 0, 0
	listings, err := seed.Load(writeRecords(t, []any{r}))
	if err != nil {
		t.Fatalf("Load() error = %v, want price, bedrooms and sqft of 0 accepted", err)
	}
	if got := listings[0]; got.Price != 0 || got.Bedrooms != 0 || got.Sqft != 0 {
		t.Errorf("Price, Bedrooms, Sqft = %v, %d, %d; want 0, 0, 0", got.Price, got.Bedrooms, got.Sqft)
	}
}

// TestLoadStatusMessage pins the exact error text the startup log carries
// for an invalid status (main prefixes it with "seed: ").
func TestLoadStatusMessage(t *testing.T) {
	records := []any{validRecord("X0"), validRecord("X1"), validRecord("X2"), validRecord("X3")}
	records[3].(map[string]any)["status"] = "gone"
	path := writeRecords(t, records)
	_, err := seed.Load(path)
	want := path + `: record index 3: field "status": invalid listing status: "gone"`
	if err == nil || err.Error() != want {
		t.Errorf("Load() error = %v, want %q", err, want)
	}
}

func TestLoadFileErrors(t *testing.T) {
	t.Run("missing file", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "nope.json")
		_, err := seed.Load(path)
		if !errors.Is(err, fs.ErrNotExist) {
			t.Errorf("Load() error = %v, want fs.ErrNotExist", err)
		}
		if err != nil && !strings.Contains(err.Error(), path) {
			t.Errorf("Load() error %q does not name the path", err)
		}
	})

	for _, tt := range []struct{ name, data, want string }{
		{"object", `{}`, "expected a JSON array"},
		{"null", `null`, "expected a JSON array"},
		{"empty file", ``, "expected a JSON array"},
		{"malformed", `[{"id": `, "invalid JSON"},
		{"trailing data", `[] []`, "invalid JSON"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			path := writeFile(t, tt.data)
			listings, err := seed.Load(path)
			if err == nil {
				t.Fatalf("Load() = %d listings, nil error; want an error", len(listings))
			}
			if !strings.HasPrefix(err.Error(), path+": ") || !strings.Contains(err.Error(), tt.want) {
				t.Errorf("Load() error = %q, want the path and %q", err, tt.want)
			}
		})
	}

	t.Run("empty array", func(t *testing.T) {
		listings, err := seed.Load(writeFile(t, " [ ] "))
		if err != nil || len(listings) != 0 {
			t.Errorf("Load() = %v, %v; want no listings and nil error", listings, err)
		}
	})
}
