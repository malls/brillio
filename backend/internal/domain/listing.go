// Package domain holds the listing search service's core types and its pure
// business logic. It does no I/O and depends on nothing outside the standard
// library; the current time is always passed in, never read here.
package domain

import (
	"errors"
	"fmt"
	"slices"
	"time"
)

// DateLayout is the time layout of a listing's date on the wire and in
// storage: an ISO 8601 calendar date, YYYY-MM-DD.
const DateLayout = "2006-01-02"

// ListingStatus is a listing's market status.
type ListingStatus string

// The valid listing statuses. No other value is a ListingStatus.
const (
	StatusActive  ListingStatus = "active"
	StatusPending ListingStatus = "pending"
	StatusSold    ListingStatus = "sold"
)

// allStatuses is the single list of valid statuses, in canonical order.
var allStatuses = [...]ListingStatus{StatusActive, StatusPending, StatusSold}

// ErrInvalidStatus is wrapped by ParseStatus for a value that is not a valid
// ListingStatus. Check for it with errors.Is.
var ErrInvalidStatus = errors.New("invalid listing status")

// AllStatuses returns every valid status in canonical order: active, pending,
// sold. Each call returns a new slice, so callers may modify it. A search with
// no status filter means all of these.
func AllStatuses() []ListingStatus {
	return slices.Clone(allStatuses[:])
}

// ParseStatus returns the ListingStatus equal to s. The match is exact and
// case-sensitive, and s is not trimmed (callers trim). Any other value returns
// an error wrapping ErrInvalidStatus that quotes s.
func ParseStatus(s string) (ListingStatus, error) {
	for _, status := range allStatuses {
		if string(status) == s {
			return status, nil
		}
	}
	return "", fmt.Errorf("%w: %q", ErrInvalidStatus, s)
}

// Listing is one property listing from a source feed (docs/SPEC.md section 4).
// Identity is the pair (Source, ID): ID is unique only within its source, so
// never key anything on ID alone.
//
// Listing is the in-memory model; do not json.Marshal or json.Unmarshal it (or
// any struct embedding it) to the wire. ListedDate is a calendar date; parse
// and format it with DateLayout in a boundary record or DTO. The JSON tags give
// the canonical wire field names that those boundary types mirror.
type Listing struct {
	// ID identifies the listing within its Source.
	ID string `json:"id"`
	// Source is the feed the listing came from, such as MLS_A.
	Source string `json:"source"`
	// Address is the free-text street address, not normalized.
	Address string `json:"address"`
	// City is the city name, trimmed with internal whitespace collapsed.
	City string `json:"city"`
	// State is the uppercased state code.
	State string `json:"state"`
	// Zip is the postal code.
	Zip string `json:"zip"`
	// Price is the list price in US dollars.
	Price float64 `json:"price"`
	// Bedrooms is the bedroom count.
	Bedrooms int `json:"bedrooms"`
	// Bathrooms is the bathroom count; halves are allowed, such as 1.5.
	Bathrooms float64 `json:"bathrooms"`
	// Sqft is the interior area in square feet.
	Sqft int `json:"sqft"`
	// Latitude is the WGS 84 latitude in degrees.
	Latitude float64 `json:"latitude"`
	// Longitude is the WGS 84 longitude in degrees.
	Longitude float64 `json:"longitude"`
	// ListedDate is the calendar date the listing went on the market. It is a
	// date, not an instant: parse and format it with DateLayout in a boundary
	// record or DTO (time.Parse(DateLayout, s) yields midnight UTC), never
	// through encoding/json, whose time.Time form is RFC 3339 and rejects
	// "2026-08-29".
	ListedDate time.Time `json:"listedDate"`
	// Status is the listing's market status.
	Status ListingStatus `json:"status"`
	// Description is the free-text description that keyword search matches.
	Description string `json:"description"`
}
