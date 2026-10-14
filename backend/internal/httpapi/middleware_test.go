package httpapi_test

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"

	"brillio/backend/internal/httpapi"
)

func TestRequestID(t *testing.T) {
	t.Run("echoed when sent", func(t *testing.T) {
		h, _ := newTestRouter(nil)
		rec := serve(h, http.MethodGet, "/healthz", "X-Request-ID", "abc-123")
		if got := assertRequestID(t, rec); got != "abc-123" {
			t.Errorf("X-Request-ID = %q, want %q", got, "abc-123")
		}
	})

	t.Run("generated when absent", func(t *testing.T) {
		h, _ := newTestRouter(nil)
		first := assertRequestID(t, serve(h, http.MethodGet, "/healthz"))
		second := assertRequestID(t, serve(h, http.MethodGet, "/healthz"))
		for _, id := range []string{first, second} {
			if !hexID.MatchString(id) {
				t.Errorf("generated X-Request-ID = %q, want 32 lowercase hex characters", id)
			}
		}
		if first == second {
			t.Errorf("two requests got the same generated ID %q", first)
		}
	})

	t.Run("present on every response kind", func(t *testing.T) {
		tests := []struct {
			name       string
			pinger     httpapi.Pinger
			method     string
			target     string
			headers    []string
			wantStatus int
		}{
			{name: "404", method: http.MethodGet, target: "/nope", wantStatus: http.StatusNotFound},
			{name: "405", method: http.MethodPost, target: "/healthz", wantStatus: http.StatusMethodNotAllowed},
			{name: "500", pinger: &fakePinger{panicWith: "boom"}, method: http.MethodGet, target: "/healthz",
				wantStatus: http.StatusInternalServerError},
			{name: "204 preflight", method: http.MethodOptions, target: "/healthz",
				headers:    []string{"Origin", testOrigin, "Access-Control-Request-Method", "GET"},
				wantStatus: http.StatusNoContent},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				h, _ := newTestRouter(tt.pinger)
				rec := serve(h, tt.method, tt.target, tt.headers...)
				if rec.Code != tt.wantStatus {
					t.Fatalf("status = %d, want %d", rec.Code, tt.wantStatus)
				}
				if id := assertRequestID(t, rec); !hexID.MatchString(id) {
					t.Errorf("X-Request-ID = %q, want a generated 32-hex ID", id)
				}
			})
		}
	})

	t.Run("RequestIDFrom without middleware", func(t *testing.T) {
		if got := httpapi.RequestIDFrom(context.Background()); got != "" {
			t.Errorf("RequestIDFrom(empty context) = %q, want empty", got)
		}
	})
}

func TestPanicRecovery(t *testing.T) {
	const secret = "boom: secret detail"
	h, logs := newTestRouter(&fakePinger{panicWith: secret})
	rec := serve(h, http.MethodGet, "/healthz")

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json; charset=utf-8" {
		t.Errorf("Content-Type = %q", ct)
	}
	body := rec.Body.String()
	const wantBody = `{"error":{"code":"internal_error","message":"Internal server error."}}` + "\n"
	if body != wantBody {
		t.Errorf("body = %q, want %q", body, wantBody)
	}
	for _, leak := range []string{"goroutine", ".go:", "boom", "secret"} {
		if strings.Contains(body, leak) {
			t.Errorf("body leaks %q: %s", leak, body)
		}
	}

	id := assertRequestID(t, rec)
	lines := logLines(t, logs)
	panics := linesWithMsg(lines, "panic recovered")
	if len(panics) != 1 {
		t.Fatalf("got %d panic log lines, want 1: %v", len(panics), lines)
	}
	p := panics[0]
	if p["level"] != "ERROR" {
		t.Errorf("panic log level = %v, want ERROR", p["level"])
	}
	if p["request_id"] != id {
		t.Errorf("panic log request_id = %v, want %q", p["request_id"], id)
	}
	if p["panic"] != secret {
		t.Errorf("panic log panic = %v, want %q", p["panic"], secret)
	}
	if stack, _ := p["stack"].(string); !strings.Contains(stack, "goroutine") {
		t.Errorf("panic log stack = %q, want a goroutine stack trace", stack)
	}
	requests := linesWithMsg(lines, "request")
	if len(requests) != 1 || requests[0]["status"] != float64(http.StatusInternalServerError) {
		t.Errorf("request log lines = %v, want one with status 500", requests)
	}
}

func TestPanicRecoveryRepanicsAbortHandler(t *testing.T) {
	h, logs := newTestRouter(&fakePinger{panicWith: http.ErrAbortHandler})
	defer func() {
		v := recover()
		if err, ok := v.(error); !ok || !errors.Is(err, http.ErrAbortHandler) {
			t.Errorf("recovered %v, want http.ErrAbortHandler to propagate", v)
		}
		if got := linesWithMsg(logLines(t, logs), "panic recovered"); len(got) != 0 {
			t.Errorf("ErrAbortHandler was logged as a recovered panic: %v", got)
		}
	}()
	serve(h, http.MethodGet, "/healthz")
	t.Error("ServeHTTP returned normally, want http.ErrAbortHandler to propagate")
}

func TestCORS(t *testing.T) {
	preflight := func(origin string) []string {
		return []string{
			"Origin", origin,
			"Access-Control-Request-Method", "GET",
			"Access-Control-Request-Headers", "X-Request-ID",
		}
	}
	tests := []struct {
		name       string
		method     string
		target     string
		headers    []string
		wantStatus int
		// want maps header names to exact values; "" means the header must be absent.
		want      map[string]string
		wantEmpty bool
	}{
		{
			name: "allowed preflight", method: http.MethodOptions, target: "/healthz",
			headers: preflight(testOrigin), wantStatus: http.StatusNoContent, wantEmpty: true,
			want: map[string]string{
				"Access-Control-Allow-Origin":   testOrigin,
				"Access-Control-Allow-Methods":  "GET, OPTIONS",
				"Access-Control-Allow-Headers":  "Content-Type, X-Request-ID",
				"Access-Control-Expose-Headers": "X-Request-ID",
				"Vary":                          "Origin",
			},
		},
		{
			name: "allowed preflight to an unknown path", method: http.MethodOptions, target: "/nope",
			headers: preflight(testOrigin), wantStatus: http.StatusNoContent, wantEmpty: true,
			want: map[string]string{"Access-Control-Allow-Origin": testOrigin},
		},
		{
			name: "second configured origin", method: http.MethodOptions, target: "/healthz",
			headers: preflight("https://app.test"), wantStatus: http.StatusNoContent, wantEmpty: true,
			want: map[string]string{
				"Access-Control-Allow-Origin":  "https://app.test",
				"Access-Control-Allow-Methods": "GET, OPTIONS",
			},
		},
		{
			name: "disallowed preflight", method: http.MethodOptions, target: "/healthz",
			headers: preflight("http://evil.test"), wantStatus: http.StatusNoContent, wantEmpty: true,
			want: map[string]string{
				"Access-Control-Allow-Origin":   "",
				"Access-Control-Allow-Methods":  "",
				"Access-Control-Allow-Headers":  "",
				"Access-Control-Expose-Headers": "",
				"Vary":                          "Origin",
			},
		},
		{
			name: "origin match is exact", method: http.MethodOptions, target: "/healthz",
			headers: preflight(testOrigin + "/"), wantStatus: http.StatusNoContent, wantEmpty: true,
			want: map[string]string{"Access-Control-Allow-Origin": "", "Access-Control-Allow-Methods": ""},
		},
		{
			name: "allowed GET", method: http.MethodGet, target: "/healthz",
			headers: []string{"Origin", testOrigin}, wantStatus: http.StatusOK,
			want: map[string]string{
				"Access-Control-Allow-Origin":   testOrigin,
				"Access-Control-Expose-Headers": "X-Request-ID",
				"Access-Control-Allow-Methods":  "",
				"Vary":                          "Origin",
			},
		},
		{
			name: "disallowed GET is served without CORS headers", method: http.MethodGet, target: "/healthz",
			headers: []string{"Origin", "http://evil.test"}, wantStatus: http.StatusOK,
			want: map[string]string{"Access-Control-Allow-Origin": "", "Access-Control-Expose-Headers": ""},
		},
		{
			name: "GET without Origin", method: http.MethodGet, target: "/healthz", wantStatus: http.StatusOK,
			want: map[string]string{"Access-Control-Allow-Origin": "", "Vary": "Origin"},
		},
		{
			name: "OPTIONS that is not a preflight", method: http.MethodOptions, target: "/healthz",
			headers: []string{"Origin", testOrigin}, wantStatus: http.StatusMethodNotAllowed,
			want: map[string]string{"Access-Control-Allow-Origin": testOrigin, "Allow": "GET, HEAD"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, _ := newTestRouter(nil)
			rec := serve(h, tt.method, tt.target, tt.headers...)
			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d (body %s)", rec.Code, tt.wantStatus, rec.Body)
			}
			for name, want := range tt.want {
				if got := rec.Header().Get(name); got != want {
					t.Errorf("%s = %q, want %q", name, got, want)
				}
			}
			if tt.wantEmpty && rec.Body.Len() != 0 {
				t.Errorf("body = %q, want empty", rec.Body)
			}
			if tt.wantStatus == http.StatusMethodNotAllowed {
				if env := decodeEnvelope(t, rec.Body.Bytes()); env.Error.Code != "method_not_allowed" {
					t.Errorf("error code = %q, want method_not_allowed", env.Error.Code)
				}
			}
		})
	}
}

func TestRequestLog(t *testing.T) {
	tests := []struct {
		target     string
		wantStatus int
	}{
		{target: "/healthz", wantStatus: http.StatusOK},
		{target: "/nope", wantStatus: http.StatusNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.target, func(t *testing.T) {
			h, logs := newTestRouter(nil)
			rec := serve(h, http.MethodGet, tt.target)
			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", rec.Code, tt.wantStatus)
			}

			requests := linesWithMsg(logLines(t, logs), "request")
			if len(requests) != 1 {
				t.Fatalf("got %d request log lines, want 1", len(requests))
			}
			l := requests[0]
			want := map[string]any{
				"level":       "INFO",
				"method":      http.MethodGet,
				"path":        tt.target,
				"status":      float64(tt.wantStatus),
				"request_id":  rec.Header().Get("X-Request-ID"),
				"remote_addr": "192.0.2.1:1234", // httptest.NewRequest's default
			}
			for key, w := range want {
				if l[key] != w {
					t.Errorf("request log %s = %v, want %v", key, l[key], w)
				}
			}
			if d, ok := l["duration_ms"].(float64); !ok || d < 0 {
				t.Errorf("request log duration_ms = %v, want a non-negative number", l["duration_ms"])
			}
		})
	}
}
