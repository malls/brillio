package httpapi_test

import (
	"errors"
	"net/http"
	"slices"
	"strings"
	"testing"

	"brillio/backend/internal/httpapi"
)

func TestCities(t *testing.T) {
	h, _ := newSearchRouter(t, testNow)
	rec := get(t, h, "/api/cities", http.StatusOK)

	var resp wireCities
	strictRoundTrip(t, rec.Body.Bytes(), &resp)
	want := []string{"Chantilly", "Fairfax", "Manassas", "Reston", "Springfield", "Vienna"}
	if !slices.Equal(resp.Cities, want) {
		t.Errorf("cities = %v, want %v", resp.Cities, want)
	}

	// Query parameters are ignored: strictness is a search rule.
	if again := get(t, h, "/api/cities?minprice=1", http.StatusOK); again.Body.String() != rec.Body.String() {
		t.Errorf("with a query: body = %s, want %s", again.Body, rec.Body)
	}
}

func TestCitiesEmpty(t *testing.T) {
	const wantBody = `{"cities":[]}` + "\n"
	tests := []struct {
		name   string
		cities httpapi.CityLister
	}{
		{"unseeded store", openStore(t)},
		{"nil slice", fakeCities{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, _ := newAPIRouter(httpapi.Deps{Cities: tt.cities}, testNow)
			rec := get(t, h, "/api/cities", http.StatusOK)
			if rec.Body.String() != wantBody {
				t.Errorf("body = %q, want %q", rec.Body, wantBody)
			}
			var resp wireCities
			strictRoundTrip(t, rec.Body.Bytes(), &resp)
		})
	}
}

func TestCitiesFailure(t *testing.T) {
	const detail = "cities query: secret detail"
	h, logs := newAPIRouter(httpapi.Deps{Cities: fakeCities{err: errors.New(detail)}}, testNow)
	rec := get(t, h, "/api/cities", http.StatusInternalServerError)

	const wantBody = `{"error":{"code":"internal_error","message":"Internal server error."}}` + "\n"
	if rec.Body.String() != wantBody {
		t.Errorf("body = %q, want %q", rec.Body, wantBody)
	}
	if strings.Contains(rec.Body.String(), "secret") {
		t.Errorf("body leaks the error: %s", rec.Body)
	}
	failures := linesWithMsg(logLines(t, logs), "request failed")
	if len(failures) != 1 {
		t.Fatalf("got %d request failed log lines, want 1: %s", len(failures), logs)
	}
	if l := failures[0]; l["level"] != "ERROR" || l["err"] != detail || l["request_id"] != rec.Header().Get("X-Request-ID") {
		t.Errorf("request failed log = %v, want level ERROR, err %q and the response's request ID", l, detail)
	}
}
