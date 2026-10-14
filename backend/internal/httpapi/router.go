// Package httpapi is the service's HTTP layer: routing, middleware, request
// parsing and JSON responses (docs/SPEC.md section 5). NewRouter is the only
// place routes are registered.
package httpapi

import (
	"context"
	"log/slog"
	"net/http"
	"slices"
	"time"
)

// Pinger reports whether a backing dependency, such as the database, is
// reachable. GET /healthz calls it.
type Pinger interface {
	Ping(ctx context.Context) error
}

// Deps are the dependencies NewRouter wires into the handlers.
type Deps struct {
	// Logger receives request, panic and error logs. Nil discards them.
	Logger *slog.Logger
	// CORSOrigins are the exact origins allowed cross-origin access.
	CORSOrigins []string
	// Pinger is checked by GET /healthz. Nil means there is nothing to check.
	Pinger Pinger
	// Searcher runs the query behind GET /api/listings/search. It must be set
	// when that route is served: nil is a wiring bug, and the route panics
	// into the recovery middleware's 500.
	Searcher Searcher
	// Cities lists the names GET /api/cities returns. Like Searcher, it must
	// be set when its route is served.
	Cities CityLister
	// Now is the request clock that search ages listings against. Nil means
	// time.Now.
	Now func() time.Time
}

// api holds the dependencies shared by every handler and middleware.
type api struct {
	deps Deps
	log  *slog.Logger
}

// NewRouter returns the service's root handler: every route, wrapped (from
// the outside in) in request ID, request logging, panic recovery and CORS
// middleware, with JSON error envelopes for unknown routes (404) and
// unsupported methods (405).
func NewRouter(deps Deps) http.Handler {
	deps.CORSOrigins = slices.Clone(deps.CORSOrigins)
	if deps.Now == nil {
		deps.Now = time.Now
	}
	log := deps.Logger
	if log == nil {
		log = slog.New(slog.DiscardHandler)
	}
	a := &api{deps: deps, log: log}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", a.healthz)
	mux.HandleFunc("GET /api/cities", a.cities)
	mux.HandleFunc("GET /api/listings/search", a.searchListings)

	h := a.fallbacks(mux)
	h = a.cors(h)
	h = a.recoverPanic(h)
	h = a.logRequests(h)
	h = requestID(h)
	return h
}

// fallbacks serves mux but replaces its plain-text 404 and 405 responses with
// the JSON error envelope.
//
// ServeMux.Handler returns an empty pattern exactly when no registered route
// applies, and the handler it then returns is the mux's own "not found",
// "method not allowed" or redirect handler, never one of ours. Running that
// handler against a probe writer is therefore free of side effects and tells
// us which case it is, including the Allow header the mux computed for a 405.
// Matched requests go through mux.ServeHTTP so r.Pattern and r.PathValue are
// populated as usual.
func (a *api) fallbacks(mux *http.ServeMux) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h, pattern := mux.Handler(r)
		if pattern != "" {
			mux.ServeHTTP(w, r)
			return
		}
		probe := &probeWriter{header: http.Header{}}
		h.ServeHTTP(probe, r)
		switch probe.status {
		case http.StatusNotFound:
			a.writeError(w, r, http.StatusNotFound, codeNotFound, msgNotFound, nil)
		case http.StatusMethodNotAllowed:
			w.Header().Set("Allow", probe.header.Get("Allow"))
			a.writeError(w, r, http.StatusMethodNotAllowed, codeMethodNotAllowed, msgMethodNotAllowed, nil)
		default:
			h.ServeHTTP(w, r)
		}
	})
}

// probeWriter discards a response, keeping only its headers and status.
type probeWriter struct {
	header http.Header
	status int
}

func (p *probeWriter) Header() http.Header { return p.header }

func (p *probeWriter) WriteHeader(code int) {
	if p.status == 0 {
		p.status = code
	}
}

func (p *probeWriter) Write(b []byte) (int, error) {
	if p.status == 0 {
		p.status = http.StatusOK
	}
	return len(b), nil
}
