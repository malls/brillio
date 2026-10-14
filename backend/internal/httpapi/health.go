package httpapi

import (
	"context"
	"log/slog"
	"net/http"
	"time"
)

// healthPingTimeout bounds the dependency check in GET /healthz.
const healthPingTimeout = 2 * time.Second

// healthResponse is the GET /healthz body (docs/SPEC.md section 5.1).
type healthResponse struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

// healthz answers 200 {"status":"ok"}, or 503 {"status":"degraded",...} when
// the Pinger fails. The failure detail goes to the log, not the client.
func (a *api) healthz(w http.ResponseWriter, r *http.Request) {
	if a.deps.Pinger != nil {
		ctx, cancel := context.WithTimeout(r.Context(), healthPingTimeout)
		defer cancel()
		if err := a.deps.Pinger.Ping(ctx); err != nil {
			a.log.LogAttrs(r.Context(), slog.LevelError, "health check failed",
				slog.String("request_id", RequestIDFrom(r.Context())),
				slog.String("err", err.Error()),
			)
			a.writeJSON(w, r, http.StatusServiceUnavailable,
				healthResponse{Status: "degraded", Error: "database unavailable"})
			return
		}
	}
	a.writeJSON(w, r, http.StatusOK, healthResponse{Status: "ok"})
}
