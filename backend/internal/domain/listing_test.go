package domain_test

import (
	"encoding/json"
	"errors"
	"slices"
	"strings"
	"testing"

	"brillio/backend/internal/domain"
)

func TestParseStatus(t *testing.T) {
	tests := []struct {
		in      string
		want    domain.ListingStatus
		wantErr bool
	}{
		{in: "active", want: domain.StatusActive},
		{in: "pending", want: domain.StatusPending},
		{in: "sold", want: domain.StatusSold},
		{in: "Active", wantErr: true},
		{in: " active", wantErr: true},
		{in: "", wantErr: true},
		{in: "closed", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got, err := domain.ParseStatus(tt.in)
			if tt.wantErr {
				if !errors.Is(err, domain.ErrInvalidStatus) {
					t.Fatalf("ParseStatus(%q) error = %v, want ErrInvalidStatus", tt.in, err)
				}
				if !strings.Contains(err.Error(), `"`+tt.in+`"`) {
					t.Errorf("ParseStatus(%q) error = %q, want it to quote the input", tt.in, err)
				}
				if got != "" {
					t.Errorf("ParseStatus(%q) = %q, want empty status on error", tt.in, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseStatus(%q) error = %v, want nil", tt.in, err)
			}
			if got != tt.want {
				t.Errorf("ParseStatus(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestAllStatuses(t *testing.T) {
	want := []domain.ListingStatus{domain.StatusActive, domain.StatusPending, domain.StatusSold}
	got := domain.AllStatuses()
	if !slices.Equal(got, want) {
		t.Fatalf("AllStatuses() = %v, want %v", got, want)
	}

	got[0] = "mutated"
	if again := domain.AllStatuses(); !slices.Equal(again, want) {
		t.Errorf("AllStatuses() after mutating an earlier result = %v, want %v", again, want)
	}
	for _, s := range want {
		if _, err := domain.ParseStatus(string(s)); err != nil {
			t.Errorf("ParseStatus(%q) after mutating an AllStatuses result: %v", s, err)
		}
	}
}

func TestListingJSONKeys(t *testing.T) {
	want := []string{
		"address", "bathrooms", "bedrooms", "city", "description", "id", "latitude", "listedDate",
		"longitude", "price", "source", "sqft", "state", "status", "zip",
	}
	b, err := json.Marshal(domain.Listing{})
	if err != nil {
		t.Fatalf("json.Marshal(Listing{}): %v", err)
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(b, &fields); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	got := make([]string, 0, len(fields))
	for k := range fields {
		got = append(got, k)
	}
	slices.Sort(got)
	if !slices.Equal(got, want) {
		t.Errorf("Listing JSON keys = %v, want %v", got, want)
	}
}
