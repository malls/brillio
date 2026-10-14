package httpapi

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"
	"slices"
	"time"
)

// requestIDHeader carries the request ID in both directions. Go writes it on
// the wire in canonical form, X-Request-Id; header names are case-insensitive.
const requestIDHeader = "X-Request-ID"

// CORS response values (docs/SPEC.md section 5.6).
const (
	corsAllowMethods  = "GET, OPTIONS"
	corsAllowHeaders  = "Content-Type, " + requestIDHeader
	corsExposeHeaders = requestIDHeader
)

// requestIDKey is the context key for the request ID.
type requestIDKey struct{}

// RequestIDFrom returns the request ID that the router's middleware stored in
// ctx, or "" if there is none.
func RequestIDFrom(ctx context.Context) string {
	id, _ := ctx.Value(requestIDKey{}).(string)
	return id
}

// requestID echoes the client's X-Request-ID, or generates one when it is
// absent, sets it on the response before anything else runs, and stores it in
// the request context. The inbound value is echoed verbatim: net/http rejects
// CR and LF in header values, and the JSON log handler escapes the rest.
func requestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get(requestIDHeader)
		if id == "" {
			id = newRequestID()
		}
		w.Header().Set(requestIDHeader, id)
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), requestIDKey{}, id)))
	})
}

// newRequestID returns 16 random bytes, hex-encoded (32 characters).
func newRequestID() string {
	var b [16]byte
	// crypto/rand.Read never returns an error (since Go 1.24 it crashes the
	// program instead), so there is nothing to handle.
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}

// logRequests emits one info line per request once the response is done:
// method, path, status, duration_ms, request_id and remote_addr.
func (a *api) logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sw := newStatusWriter(w)
		next.ServeHTTP(sw, r)
		a.log.LogAttrs(r.Context(), slog.LevelInfo, "request",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Int("status", sw.status),
			slog.Float64("duration_ms", float64(time.Since(start).Microseconds())/1000),
			slog.String("request_id", RequestIDFrom(r.Context())),
			slog.String("remote_addr", r.RemoteAddr),
		)
	})
}

// recoverPanic turns a handler panic into a logged 500 internal_error
// envelope. The panic value and stack go to the log with the request ID, never
// to the client. http.ErrAbortHandler is re-panicked so net/http aborts the
// response as intended. If the handler had already started its response, the
// envelope is skipped because the status line is already sent.
func (a *api) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sw := newStatusWriter(w)
		defer func() {
			v := recover()
			if v == nil {
				return
			}
			if v == http.ErrAbortHandler {
				panic(v)
			}
			a.log.LogAttrs(r.Context(), slog.LevelError, "panic recovered",
				slog.String("request_id", RequestIDFrom(r.Context())),
				slog.String("panic", fmt.Sprint(v)),
				slog.String("stack", string(debug.Stack())),
			)
			if !sw.wrote {
				a.writeError(sw, r, http.StatusInternalServerError, codeInternal, msgInternal, nil)
			}
		}()
		next.ServeHTTP(sw, r)
	})
}

// cors applies docs/SPEC.md section 5.6. A request whose Origin exactly equals
// a configured origin gets Access-Control-Allow-Origin and
// Access-Control-Expose-Headers. A preflight (OPTIONS with Origin and
// Access-Control-Request-Method) is answered here with 204 and never reaches
// a route; only an allowed origin also gets the Allow-Methods and
// Allow-Headers that let the browser proceed. Requests from other origins, or
// with no Origin, are served normally without CORS headers.
func (a *api) cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Add("Vary", "Origin")
		origin := r.Header.Get("Origin")
		allowed := origin != "" && slices.Contains(a.deps.CORSOrigins, origin)
		if allowed {
			h.Set("Access-Control-Allow-Origin", origin)
			h.Set("Access-Control-Expose-Headers", corsExposeHeaders)
		}
		if r.Method == http.MethodOptions && origin != "" && r.Header.Get("Access-Control-Request-Method") != "" {
			if allowed {
				h.Set("Access-Control-Allow-Methods", corsAllowMethods)
				h.Set("Access-Control-Allow-Headers", corsAllowHeaders)
			}
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// statusWriter records the status code a handler sends. Unwrap lets
// http.ResponseController reach the underlying writer.
type statusWriter struct {
	http.ResponseWriter
	status int  // final status; 200 until the handler says otherwise
	wrote  bool // whether the final status has been sent
}

func newStatusWriter(w http.ResponseWriter) *statusWriter {
	return &statusWriter{ResponseWriter: w, status: http.StatusOK}
}

// WriteHeader records the first final (non-1xx) status and forwards it.
func (w *statusWriter) WriteHeader(code int) {
	if !w.wrote && code >= http.StatusOK {
		w.status, w.wrote = code, true
	}
	w.ResponseWriter.WriteHeader(code)
}

// Write marks the response as started (an implicit 200 when no status was
// set) and forwards b.
func (w *statusWriter) Write(b []byte) (int, error) {
	w.wrote = true
	return w.ResponseWriter.Write(b)
}

// Unwrap returns the wrapped ResponseWriter for http.ResponseController.
func (w *statusWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}
