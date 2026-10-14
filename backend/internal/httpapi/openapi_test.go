package httpapi_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"maps"
	"net/http"
	"os"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"testing"

	"brillio/backend/internal/httpapi"
)

// openAPIPath is the repository's OpenAPI document, which the server embeds.
const openAPIPath = "../../openapi.yaml"

// openAPIDoc fetches GET /api/openapi.yaml and parses the body with the YAML
// subset reader.
func openAPIDoc(t *testing.T) map[string]any {
	t.Helper()
	h, _ := newTestRouter(nil)
	rec := serve(h, http.MethodGet, "/api/openapi.yaml")
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /api/openapi.yaml: status = %d, want %d", rec.Code, http.StatusOK)
	}
	doc, err := parseYAMLSubset(rec.Body.Bytes())
	if err != nil {
		t.Fatalf("parse the OpenAPI document: %v", err)
	}
	return doc
}

// lookup walks from v through nested mappings along keys and returns the
// value it reaches.
func lookup(t *testing.T, v any, keys ...string) any {
	t.Helper()
	for i, key := range keys {
		m, ok := v.(map[string]any)
		if !ok {
			t.Fatalf("%q is a %T, want a mapping", keys[:i], v)
		}
		if v, ok = m[key]; !ok {
			t.Fatalf("the document has no %q", keys[:i+1])
		}
	}
	return v
}

// lookupMap is lookup for a value that must be a mapping.
func lookupMap(t *testing.T, v any, keys ...string) map[string]any {
	t.Helper()
	m, ok := lookup(t, v, keys...).(map[string]any)
	if !ok {
		t.Fatalf("%q is not a mapping", keys)
	}
	return m
}

// lookupStrings is lookup for a value that must be a sequence of strings.
func lookupStrings(t *testing.T, v any, keys ...string) []string {
	t.Helper()
	seq, ok := lookup(t, v, keys...).([]any)
	if !ok {
		t.Fatalf("%q is not a sequence", keys)
	}
	out := make([]string, len(seq))
	for i, item := range seq {
		if out[i], ok = item.(string); !ok {
			t.Fatalf("%q item %d is a %T, want a string", keys, i, item)
		}
	}
	return out
}

// exampleKeys returns the keys from the document root to the value of a
// named JSON example of an operation's response.
func exampleKeys(path, status, name string) []string {
	return []string{"paths", path, "get", "responses", status, "content", "application/json", "examples", name, "value"}
}

func TestOpenAPIDocument(t *testing.T) {
	want, err := os.ReadFile(openAPIPath)
	if err != nil {
		t.Fatalf("read the OpenAPI document: %v", err)
	}
	h, _ := newTestRouter(nil)
	rec := serve(h, http.MethodGet, "/api/openapi.yaml", "Origin", testOrigin)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d (body %s)", rec.Code, http.StatusOK, rec.Body)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/yaml" {
		t.Errorf("Content-Type = %q, want application/yaml", ct)
	}
	if cl := rec.Header().Get("Content-Length"); cl != strconv.Itoa(len(want)) {
		t.Errorf("Content-Length = %q, want %d", cl, len(want))
	}
	assertRequestID(t, rec)
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != testOrigin {
		t.Errorf("Access-Control-Allow-Origin = %q, want %q", got, testOrigin)
	}
	if !bytes.Equal(rec.Body.Bytes(), want) {
		t.Errorf("body (%d bytes) differs from %s (%d bytes), want the file byte for byte",
			rec.Body.Len(), openAPIPath, len(want))
	}
}

func TestOpenAPIParsesAsYAML(t *testing.T) {
	doc := openAPIDoc(t)
	if v, _ := doc["openapi"].(string); !strings.HasPrefix(v, "3.1.") {
		t.Errorf("openapi = %q, want 3.1.x", v)
	}

	paths := lookupMap(t, doc, "paths")
	wantPaths := []string{"/api/cities", "/api/listings/search", "/api/openapi.yaml", "/healthz"}
	if got := slices.Sorted(maps.Keys(paths)); !slices.Equal(got, wantPaths) {
		t.Fatalf("paths = %v, want %v", got, wantPaths)
	}

	// Every documented path is a GET route, and a plain GET gets one of the
	// statuses its operation documents.
	h, _ := newSearchRouter(t, testNow)
	for _, path := range wantPaths {
		if keys := slices.Sorted(maps.Keys(lookupMap(t, paths, path))); !slices.Equal(keys, []string{"get"}) {
			t.Errorf("%s has %v, want only get: the router serves GET alone", path, keys)
		}
		responses := lookupMap(t, paths, path, "get", "responses")
		rec := serve(h, http.MethodGet, path)
		if _, ok := responses[strconv.Itoa(rec.Code)]; !ok {
			t.Errorf("GET %s: status %d is not among the documented responses %v",
				path, rec.Code, slices.Sorted(maps.Keys(responses)))
		}
	}
}

func TestOpenAPISearchParamsMatchHandler(t *testing.T) {
	doc := openAPIDoc(t)
	if _, ok := lookupMap(t, doc, "paths", "/api/listings/search")["parameters"]; ok {
		t.Error("the search path has path-level parameters, want them all on its get")
	}
	params, ok := lookup(t, doc, "paths", "/api/listings/search", "get", "parameters").([]any)
	if !ok {
		t.Fatal("the search parameters are not a sequence")
	}

	var documented []string
	for i, p := range params {
		param, _ := p.(map[string]any)
		name, _ := param["name"].(string)
		if name == "" {
			t.Fatalf("search parameter %d has no name, want an inline parameter object", i)
		}
		if in := param["in"]; in != "query" {
			t.Errorf("search parameter %s is in %v, want query", name, in)
		}
		if slices.Contains(documented, name) {
			t.Errorf("search parameter %s is documented twice", name)
		}
		documented = append(documented, name)
	}

	// The handler's allowlist and the document must name the same parameters.
	accepted := httpapi.SearchQueryParams()
	var undocumented, unaccepted []string
	for _, name := range accepted {
		if !slices.Contains(documented, name) {
			undocumented = append(undocumented, name)
		}
	}
	for _, name := range documented {
		if !slices.Contains(accepted, name) {
			unaccepted = append(unaccepted, name)
		}
	}
	if len(undocumented) > 0 || len(unaccepted) > 0 {
		t.Fatalf("search parameters out of sync with httpapi.SearchQueryParams():\n"+
			"  accepted but not documented: %v\n  documented but not accepted: %v", undocumented, unaccepted)
	}
	if !slices.Equal(documented, accepted) {
		t.Errorf("documented order = %v, want %v (docs/SPEC.md section 5.3 table order)", documented, accepted)
	}
}

func TestOpenAPIExamplesMatchSPEC(t *testing.T) {
	doc := openAPIDoc(t)
	tests := []struct {
		heading string
		path    string
		status  string
	}{
		{"### 5.2", "/api/cities", "200"},
		{"### 5.4", "/api/listings/search", "200"},
		{"### 5.5", "/api/listings/search", "400"},
	}
	for _, tt := range tests {
		t.Run("section "+strings.TrimPrefix(tt.heading, "### "), func(t *testing.T) {
			got := lookup(t, doc, exampleKeys(tt.path, tt.status, "spec")...)
			var want any
			if err := json.Unmarshal(specExample(t, tt.heading), &want); err != nil {
				t.Fatalf("decode SPEC example: %v", err)
			}
			if !reflect.DeepEqual(got, want) {
				js, _ := json.Marshal(got)
				t.Errorf("GET %s %s example =\n%s\nwant the SPEC %q example", tt.path, tt.status, js, tt.heading)
			}
		})
	}
}

// The examples that the SPEC does not spell out must equal what the handlers
// send.
func TestOpenAPIExamplesMatchHandler(t *testing.T) {
	doc := openAPIDoc(t)
	healthy, _ := newTestRouter(nil)
	degraded, _ := newTestRouter(&fakePinger{err: errors.New("disk on fire")})
	searching, _ := newSearchRouter(t, testNow)
	failing, _ := newAPIRouter(httpapi.Deps{Searcher: &fakeSearcher{err: errors.New("search down")}}, testNow)
	tests := []struct {
		name       string
		keys       []string
		h          http.Handler
		target     string
		wantStatus int
	}{
		{"healthy", exampleKeys("/healthz", "200", "ok"), healthy, "/healthz", http.StatusOK},
		{"degraded", exampleKeys("/healthz", "503", "degraded"), degraded, "/healthz", http.StatusServiceUnavailable},
		{"malformed query", exampleKeys("/api/listings/search", "400", "malformedQuery"),
			searching, "/api/listings/search?keyword=100%", http.StatusBadRequest},
		{"internal error",
			[]string{"components", "responses", "InternalError", "content", "application/json", "examples", "internal", "value"},
			failing, "/api/listings/search", http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := serve(tt.h, http.MethodGet, tt.target)
			if rec.Code != tt.wantStatus {
				t.Fatalf("GET %s: status = %d, want %d (body %s)", tt.target, rec.Code, tt.wantStatus, rec.Body)
			}
			var want any
			if err := json.Unmarshal(rec.Body.Bytes(), &want); err != nil {
				t.Fatalf("decode body: %v", err)
			}
			if got := lookup(t, doc, tt.keys...); !reflect.DeepEqual(got, want) {
				js, _ := json.Marshal(got)
				t.Errorf("example %q =\n%s\nwant the GET %s body\n%s", tt.keys, js, tt.target, rec.Body)
			}
		})
	}
}

func TestOpenAPIErrorCodeEnums(t *testing.T) {
	schemas := lookupMap(t, openAPIDoc(t), "components", "schemas")
	tests := []struct {
		schema string
		keys   []string
		want   []string
	}{
		// docs/SPEC.md section 5.5, "Field error code values".
		{"FieldError", []string{"FieldError", "properties", "code", "enum"}, []string{
			"not_a_number", "not_an_integer", "negative", "not_positive", "out_of_range",
			"too_long", "invalid_enum", "min_gt_max", "unknown_param", "malformed_query",
		}},
		// docs/SPEC.md section 5.5, the status and code table.
		{"ErrorEnvelope", []string{"ErrorEnvelope", "properties", "error", "properties", "code", "enum"}, []string{
			"validation_error", "not_found", "method_not_allowed", "internal_error",
		}},
	}
	for _, tt := range tests {
		t.Run(tt.schema, func(t *testing.T) {
			if got := lookupStrings(t, schemas, tt.keys...); !slices.Equal(got, tt.want) {
				t.Errorf("%s code enum = %v, want %v", tt.schema, got, tt.want)
			}
		})
	}
}
