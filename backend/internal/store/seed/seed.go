// Package seed loads the listings seed file (docs/SPEC.md section 4.2): a
// JSON array of listing records, decoded strictly, normalized and validated
// into domain.Listing values. It does no database work; the store upserts
// what Load returns.
//
// Decoding rejects unknown fields and wrong JSON types. Normalization trims
// every string, collapses internal whitespace in city and state, and
// uppercases state. Validation requires every field to be present and
// non-null, a status in the enum, a YYYY-MM-DD listedDate, price, bedrooms
// and sqft >= 0, a non-empty id and source, and a (source, id) pair unique
// within the file. The first invalid record fails the whole load with a
// *RecordError naming its index and field.
package seed

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"brillio/backend/internal/domain"
)

// record is one seed file element as it appears on the wire. Every field is
// a pointer so a missing key or a JSON null (both nil) is told apart from a
// zero value; ListedDate and Status stay strings until validated.
type record struct {
	ID          *string  `json:"id"`
	Source      *string  `json:"source"`
	Address     *string  `json:"address"`
	City        *string  `json:"city"`
	State       *string  `json:"state"`
	Zip         *string  `json:"zip"`
	Price       *float64 `json:"price"`
	Bedrooms    *int     `json:"bedrooms"`
	Bathrooms   *float64 `json:"bathrooms"`
	Sqft        *int     `json:"sqft"`
	Latitude    *float64 `json:"latitude"`
	Longitude   *float64 `json:"longitude"`
	ListedDate  *string  `json:"listedDate"`
	Status      *string  `json:"status"`
	Description *string  `json:"description"`
}

// errMissing is the RecordError cause for an absent or null field.
var errMissing = errors.New("missing or null")

// RecordError reports the first invalid record in a seed file.
type RecordError struct {
	// Index is the record's 0-based position in the file's array.
	Index int
	// Field is the JSON name of the offending field, or "" when the problem
	// is not tied to one field (an unknown key, a non-object element, a
	// duplicate (source, id)).
	Field string
	// Err describes the problem.
	Err error
}

// Error returns `record index N: field "name": problem`, or
// `record index N: problem` when Field is empty.
func (e *RecordError) Error() string {
	if e.Field == "" {
		return fmt.Sprintf("record index %d: %v", e.Index, e.Err)
	}
	return fmt.Sprintf("record index %d: field %q: %v", e.Index, e.Field, e.Err)
}

// Unwrap returns Err.
func (e *RecordError) Unwrap() error {
	return e.Err
}

// Load reads the seed file at path and returns its listings in file order.
// A file that cannot be read returns the os error (its message names the
// path); any other failure is prefixed with the path, and an invalid record
// is a *RecordError.
func Load(path string) ([]domain.Listing, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if trimmed := bytes.TrimLeft(data, " \t\r\n"); len(trimmed) == 0 || trimmed[0] != '[' {
		return nil, fmt.Errorf("%s: expected a JSON array of listing records", path)
	}
	var raws []json.RawMessage
	if err := json.Unmarshal(data, &raws); err != nil {
		return nil, fmt.Errorf("%s: invalid JSON: %w", path, err)
	}

	listings := make([]domain.Listing, 0, len(raws))
	firstIndex := make(map[[2]string]int, len(raws)) // (source, id) -> record index
	for i, raw := range raws {
		l, err := parseRecord(i, raw)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", path, err)
		}
		key := [2]string{l.Source, l.ID}
		if j, dup := firstIndex[key]; dup {
			return nil, fmt.Errorf("%s: %w", path, &RecordError{Index: i,
				Err: fmt.Errorf("duplicate (source, id) (%q, %q), first at record index %d", l.Source, l.ID, j)})
		}
		firstIndex[key] = i
		listings = append(listings, l)
	}
	return listings, nil
}

// parseRecord decodes, checks, normalizes and validates the record at index i.
func parseRecord(i int, raw json.RawMessage) (domain.Listing, error) {
	fail := func(field string, err error) (domain.Listing, error) {
		return domain.Listing{}, &RecordError{Index: i, Field: field, Err: err}
	}

	if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return fail("", errors.New("expected a JSON object, got null"))
	}
	var r record
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&r); err != nil {
		if te, ok := errors.AsType[*json.UnmarshalTypeError](err); ok {
			if te.Field == "" {
				return fail("", fmt.Errorf("expected a JSON object, got %s", te.Value))
			}
			return fail(te.Field, fmt.Errorf("JSON %s where %s expected", te.Value, te.Type))
		}
		return fail("", err) // unknown field, or malformed element
	}

	// Presence, in docs/SPEC.md section 4 field order.
	presence := []struct {
		field   string
		present bool
	}{
		{"id", r.ID != nil}, {"source", r.Source != nil}, {"address", r.Address != nil},
		{"city", r.City != nil}, {"state", r.State != nil}, {"zip", r.Zip != nil},
		{"price", r.Price != nil}, {"bedrooms", r.Bedrooms != nil}, {"bathrooms", r.Bathrooms != nil},
		{"sqft", r.Sqft != nil}, {"latitude", r.Latitude != nil}, {"longitude", r.Longitude != nil},
		{"listedDate", r.ListedDate != nil}, {"status", r.Status != nil}, {"description", r.Description != nil},
	}
	for _, p := range presence {
		if !p.present {
			return fail(p.field, errMissing)
		}
	}

	l := domain.Listing{
		ID:          strings.TrimSpace(*r.ID),
		Source:      strings.TrimSpace(*r.Source),
		Address:     strings.TrimSpace(*r.Address),
		City:        collapseSpace(*r.City),
		State:       strings.ToUpper(collapseSpace(*r.State)),
		Zip:         strings.TrimSpace(*r.Zip),
		Price:       *r.Price,
		Bedrooms:    *r.Bedrooms,
		Bathrooms:   *r.Bathrooms,
		Sqft:        *r.Sqft,
		Latitude:    *r.Latitude,
		Longitude:   *r.Longitude,
		Description: strings.TrimSpace(*r.Description),
	}

	switch {
	case l.ID == "":
		return fail("id", errors.New("must not be empty"))
	case l.Source == "":
		return fail("source", errors.New("must not be empty"))
	case l.Price < 0:
		return fail("price", fmt.Errorf("must be >= 0, got %v", l.Price))
	case l.Bedrooms < 0:
		return fail("bedrooms", fmt.Errorf("must be >= 0, got %d", l.Bedrooms))
	case l.Sqft < 0:
		return fail("sqft", fmt.Errorf("must be >= 0, got %d", l.Sqft))
	}

	date := strings.TrimSpace(*r.ListedDate)
	listed, err := time.Parse(domain.DateLayout, date)
	if err != nil {
		return fail("listedDate", fmt.Errorf("not a YYYY-MM-DD date: %q", date))
	}
	l.ListedDate = listed

	status, err := domain.ParseStatus(strings.TrimSpace(*r.Status))
	if err != nil {
		return fail("status", err)
	}
	l.Status = status

	return l, nil
}

// collapseSpace trims s and collapses each internal run of whitespace to a
// single space.
func collapseSpace(s string) string {
	return strings.Join(strings.Fields(s), " ")
}
