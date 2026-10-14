package httpapi

import (
	"context"
	"net/http"
)

// CityLister lists the distinct city names of the listings (docs/SPEC.md
// section 5.2): distinct ignoring case, sorted A to Z, each in its
// first-seen casing. *store.Store implements it.
type CityLister interface {
	Cities(ctx context.Context) ([]string, error)
}

// citiesResponse is the GET /api/cities body (docs/SPEC.md section 5.2).
type citiesResponse struct {
	Cities []string `json:"cities"`
}

// cities serves GET /api/cities. Query parameters are ignored: the strict
// unknown-parameter rule is a search rule (section 5.3).
func (a *api) cities(w http.ResponseWriter, r *http.Request) {
	names, err := a.deps.Cities.Cities(r.Context())
	if err != nil {
		a.internalError(w, r, err)
		return
	}
	if names == nil {
		names = []string{} // encode as [], never null
	}
	a.writeJSON(w, r, http.StatusOK, citiesResponse{Cities: names})
}
