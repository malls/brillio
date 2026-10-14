package httpapi_test

import (
	"net/http"
	"testing"

	"brillio/backend/internal/httpapi"
	"brillio/backend/internal/store"
)

// The SQLite store is the production Pinger.
var _ httpapi.Pinger = (*store.Store)(nil)

func TestHealthzRealStore(t *testing.T) {
	st, err := store.Open(t.Context(), ":memory:")
	if err != nil {
		t.Fatalf("store.Open() error = %v", err)
	}
	h, _ := newTestRouter(st)

	rec := serve(h, http.MethodGet, "/healthz")
	if rec.Code != http.StatusOK || rec.Body.String() != `{"status":"ok"}`+"\n" {
		t.Errorf("open store: %d %q, want 200 {\"status\":\"ok\"}", rec.Code, rec.Body)
	}

	if err := st.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	rec = serve(h, http.MethodGet, "/healthz")
	wantBody := `{"status":"degraded","error":"database unavailable"}` + "\n"
	if rec.Code != http.StatusServiceUnavailable || rec.Body.String() != wantBody {
		t.Errorf("closed store: %d %q, want 503 %q", rec.Code, rec.Body, wantBody)
	}
}
