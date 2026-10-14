package httpapi_test

import (
	"errors"
	"net/http"
	"strings"
	"testing"

	"brillio/backend/internal/httpapi"
)

func TestHealthz(t *testing.T) {
	tests := []struct {
		name       string
		pinger     *fakePinger // nil means Deps.Pinger is nil
		method     string
		wantStatus int
		// wantBody is the exact body; "" skips the check (HEAD bodies are
		// net/http's to suppress, and the recorder keeps them).
		wantBody string
		// wantLogErr, when set, must appear in a "health check failed" log line.
		wantLogErr string
	}{
		{name: "no pinger", method: http.MethodGet,
			wantStatus: http.StatusOK, wantBody: `{"status":"ok"}` + "\n"},
		{name: "pinger ok", pinger: &fakePinger{}, method: http.MethodGet,
			wantStatus: http.StatusOK, wantBody: `{"status":"ok"}` + "\n"},
		{name: "pinger fails", pinger: &fakePinger{err: errors.New("disk on fire")}, method: http.MethodGet,
			wantStatus: http.StatusServiceUnavailable,
			wantBody:   `{"status":"degraded","error":"database unavailable"}` + "\n",
			wantLogErr: "disk on fire"},
		{name: "HEAD", method: http.MethodHead, wantStatus: http.StatusOK},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Leave the interface nil when there is no fake, so the router
			// sees "no pinger" rather than a non-nil interface holding a nil
			// *fakePinger.
			var pinger httpapi.Pinger
			if tt.pinger != nil {
				pinger = tt.pinger
			}
			h, logs := newTestRouter(pinger)
			rec := serve(h, tt.method, "/healthz")

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d (body %s)", rec.Code, tt.wantStatus, rec.Body)
			}
			if ct := rec.Header().Get("Content-Type"); ct != "application/json; charset=utf-8" {
				t.Errorf("Content-Type = %q, want application/json; charset=utf-8", ct)
			}
			assertRequestID(t, rec)
			if tt.wantBody != "" && rec.Body.String() != tt.wantBody {
				t.Errorf("body = %q, want %q", rec.Body, tt.wantBody)
			}
			if tt.pinger != nil && !tt.pinger.hadDeadline {
				t.Error("Ping context has no deadline, want a bounded health check")
			}
			if tt.wantLogErr != "" {
				if strings.Contains(rec.Body.String(), tt.wantLogErr) {
					t.Errorf("body leaks the ping error: %s", rec.Body)
				}
				if !strings.Contains(logs.String(), `"msg":"health check failed"`) ||
					!strings.Contains(logs.String(), tt.wantLogErr) {
					t.Errorf("log lacks a health check failure naming %q: %s", tt.wantLogErr, logs)
				}
			}
		})
	}
}
