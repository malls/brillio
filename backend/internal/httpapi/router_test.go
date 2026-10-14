package httpapi_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"brillio/backend/internal/httpapi"
)

const testOrigin = "http://localhost:5173"

// fakePinger is a test Pinger: it panics with panicWith when that is set,
// otherwise returns err. It records whether Ping saw a context deadline.
type fakePinger struct {
	err         error
	panicWith   any
	hadDeadline bool
}

func (p *fakePinger) Ping(ctx context.Context) error {
	_, p.hadDeadline = ctx.Deadline()
	if p.panicWith != nil {
		panic(p.panicWith)
	}
	return p.err
}

// newTestRouter returns a router whose logs are captured as JSON lines in the
// returned buffer.
func newTestRouter(pinger httpapi.Pinger) (http.Handler, *bytes.Buffer) {
	var buf bytes.Buffer
	h := httpapi.NewRouter(httpapi.Deps{
		Logger:      slog.New(slog.NewJSONHandler(&buf, nil)),
		CORSOrigins: []string{testOrigin, "https://app.test"},
		Pinger:      pinger,
	})
	return h, &buf
}

// serve sends one request through h with an httptest recorder. headers is a
// flat list of name, value pairs.
func serve(h http.Handler, method, target string, headers ...string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, target, nil)
	for i := 0; i+1 < len(headers); i += 2 {
		req.Header.Set(headers[i], headers[i+1])
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

// logLines decodes every JSON log line in buf.
func logLines(t *testing.T, buf *bytes.Buffer) []map[string]any {
	t.Helper()
	var lines []map[string]any
	for line := range strings.SplitSeq(strings.TrimSpace(buf.String()), "\n") {
		if line == "" {
			continue
		}
		var m map[string]any
		if err := json.Unmarshal([]byte(line), &m); err != nil {
			t.Fatalf("log line is not JSON: %v: %s", err, line)
		}
		lines = append(lines, m)
	}
	return lines
}

// linesWithMsg returns the log lines whose msg equals msg.
func linesWithMsg(lines []map[string]any, msg string) []map[string]any {
	var out []map[string]any
	for _, l := range lines {
		if l["msg"] == msg {
			out = append(out, l)
		}
	}
	return out
}

var hexID = regexp.MustCompile(`^[0-9a-f]{32}$`)

// assertRequestID checks that the response carries a request ID and returns it.
func assertRequestID(t *testing.T, rec *httptest.ResponseRecorder) string {
	t.Helper()
	id := rec.Header().Get("X-Request-ID")
	if id == "" {
		t.Fatalf("status %d response has no X-Request-ID header", rec.Code)
	}
	return id
}

// envelope mirrors the error response body for strict decoding.
type envelope struct {
	Error struct {
		Code    string               `json:"code"`
		Message string               `json:"message"`
		Fields  []httpapi.FieldError `json:"fields"`
	} `json:"error"`
}

// decodeEnvelope decodes an error body, rejecting unknown fields.
func decodeEnvelope(t *testing.T, body []byte) envelope {
	t.Helper()
	dec := json.NewDecoder(bytes.NewReader(body))
	dec.DisallowUnknownFields()
	var env envelope
	if err := dec.Decode(&env); err != nil {
		t.Fatalf("decode error envelope: %v: %s", err, body)
	}
	return env
}

func TestErrorEnvelopes(t *testing.T) {
	tests := []struct {
		name        string
		method      string
		target      string
		wantStatus  int
		wantCode    string
		wantMessage string
		wantAllow   string
	}{
		{name: "unknown path", method: http.MethodGet, target: "/nope",
			wantStatus: http.StatusNotFound, wantCode: "not_found", wantMessage: "Resource not found."},
		{name: "unbuilt api path", method: http.MethodGet, target: "/api/listings",
			wantStatus: http.StatusNotFound, wantCode: "not_found", wantMessage: "Resource not found."},
		{name: "POST on a GET route", method: http.MethodPost, target: "/healthz",
			wantStatus: http.StatusMethodNotAllowed, wantCode: "method_not_allowed",
			wantMessage: "Method not allowed.", wantAllow: "GET, HEAD"},
		{name: "DELETE on a GET route", method: http.MethodDelete, target: "/healthz",
			wantStatus: http.StatusMethodNotAllowed, wantCode: "method_not_allowed",
			wantMessage: "Method not allowed.", wantAllow: "GET, HEAD"},
		{name: "POST on the OpenAPI document", method: http.MethodPost, target: "/api/openapi.yaml",
			wantStatus: http.StatusMethodNotAllowed, wantCode: "method_not_allowed",
			wantMessage: "Method not allowed.", wantAllow: "GET, HEAD"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, _ := newTestRouter(nil)
			rec := serve(h, tt.method, tt.target)

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d (body %s)", rec.Code, tt.wantStatus, rec.Body)
			}
			if ct := rec.Header().Get("Content-Type"); ct != "application/json; charset=utf-8" {
				t.Errorf("Content-Type = %q, want application/json; charset=utf-8", ct)
			}
			assertRequestID(t, rec)
			if allow := rec.Header().Get("Allow"); allow != tt.wantAllow {
				t.Errorf("Allow = %q, want %q", allow, tt.wantAllow)
			}
			if strings.Contains(rec.Body.String(), `"fields"`) {
				t.Errorf("body has a fields key outside a validation error: %s", rec.Body)
			}
			env := decodeEnvelope(t, rec.Body.Bytes())
			if env.Error.Code != tt.wantCode || env.Error.Message != tt.wantMessage {
				t.Errorf("error = {%q, %q}, want {%q, %q}",
					env.Error.Code, env.Error.Message, tt.wantCode, tt.wantMessage)
			}
		})
	}
}

func TestNonCanonicalPathRedirects(t *testing.T) {
	h, _ := newTestRouter(nil)
	rec := serve(h, http.MethodGet, "/x/../nope")
	if rec.Code != http.StatusTemporaryRedirect {
		t.Fatalf("status = %d, want %d (body %s)", rec.Code, http.StatusTemporaryRedirect, rec.Body)
	}
	if loc := rec.Header().Get("Location"); loc != "/nope" {
		t.Errorf("Location = %q, want /nope", loc)
	}
	assertRequestID(t, rec)
}
