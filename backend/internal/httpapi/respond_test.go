package httpapi

import (
	"bytes"
	"log/slog"
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// No route returns an arbitrary value yet, so the encode-failure branch of
// writeJSON is exercised directly.
func TestWriteJSONEncodeFailure(t *testing.T) {
	var logs bytes.Buffer
	a := &api{log: slog.New(slog.NewJSONHandler(&logs, nil))}
	rec := httptest.NewRecorder()
	a.writeJSON(rec, httptest.NewRequest(http.MethodGet, "/", nil), http.StatusOK, map[string]float64{"x": math.NaN()})

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", rec.Code)
	}
	const wantBody = `{"error":{"code":"internal_error","message":"Internal server error."}}` + "\n"
	if got := rec.Body.String(); got != wantBody {
		t.Errorf("body = %q, want %q", got, wantBody)
	}
	if ct := rec.Header().Get("Content-Type"); ct != contentTypeJSON {
		t.Errorf("Content-Type = %q, want %q", ct, contentTypeJSON)
	}
	if !strings.Contains(logs.String(), `"msg":"response encoding failed"`) {
		t.Errorf("log lacks the encoding failure: %s", logs.String())
	}
}
