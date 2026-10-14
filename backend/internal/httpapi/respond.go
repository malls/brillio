package httpapi

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

// contentTypeJSON is the Content-Type of every JSON response.
const contentTypeJSON = "application/json; charset=utf-8"

// Error codes and messages for the error envelope (docs/SPEC.md section 5.5).
const (
	codeValidation       = "validation_error"
	codeNotFound         = "not_found"
	codeMethodNotAllowed = "method_not_allowed"
	codeInternal         = "internal_error"

	msgValidation       = "Invalid search parameters."
	msgNotFound         = "Resource not found."
	msgMethodNotAllowed = "Method not allowed."
	msgInternal         = "Internal server error."
)

// FieldError describes one invalid request parameter in a 400
// validation_error response.
type FieldError struct {
	// Field is the parameter name exactly as the client sent it.
	Field string `json:"field"`
	// Code is a stable machine-readable reason, such as out_of_range.
	Code string `json:"code"`
	// Message is a plain sentence a person can act on.
	Message string `json:"message"`
}

// errorEnvelope is the body of every error response:
// {"error":{"code":...,"message":...,"fields":[...]}}.
type errorEnvelope struct {
	Error errorBody `json:"error"`
}

type errorBody struct {
	Code    string       `json:"code"`
	Message string       `json:"message"`
	Fields  []FieldError `json:"fields,omitempty"`
}

// internalErrorBody is the pre-encoded 500 envelope that writeJSON falls back
// to when a response cannot be encoded.
var internalErrorBody = func() []byte {
	// Encoding two plain strings cannot fail.
	b, _ := json.Marshal(errorEnvelope{Error: errorBody{Code: codeInternal, Message: msgInternal}})
	return append(b, '\n')
}()

// writeJSON writes v as a JSON response with the given status. It encodes
// before writing anything, so an unencodable value (a NaN, say) becomes a
// logged 500 internal_error rather than a truncated body.
func (a *api) writeJSON(w http.ResponseWriter, r *http.Request, status int, v any) {
	body, err := json.Marshal(v)
	if err != nil {
		a.log.LogAttrs(r.Context(), slog.LevelError, "response encoding failed",
			slog.String("request_id", RequestIDFrom(r.Context())),
			slog.String("err", err.Error()),
		)
		status, body = http.StatusInternalServerError, internalErrorBody
	} else {
		body = append(body, '\n')
	}
	w.Header().Set("Content-Type", contentTypeJSON)
	w.WriteHeader(status)
	// A failed write means the client has gone away; there is nothing to do.
	_, _ = w.Write(body)
}

// writeError writes the error envelope. fields is set only for a 400
// validation_error and is omitted from the body when empty.
func (a *api) writeError(w http.ResponseWriter, r *http.Request, status int, code, message string, fields []FieldError) {
	a.writeJSON(w, r, status, errorEnvelope{Error: errorBody{Code: code, Message: message, Fields: fields}})
}

// internalError logs err with the request ID and writes the generic 500
// internal_error envelope. The error text goes to the log, never to the
// client.
func (a *api) internalError(w http.ResponseWriter, r *http.Request, err error) {
	a.log.LogAttrs(r.Context(), slog.LevelError, "request failed",
		slog.String("request_id", RequestIDFrom(r.Context())),
		slog.String("err", err.Error()),
	)
	a.writeError(w, r, http.StatusInternalServerError, codeInternal, msgInternal, nil)
}
