# BC-6: OpenAPI 3.1 document, embedded and served at /api/openapi.yaml, with param sync test

Contract: docs/SPEC.md §3, §5.1–5.7, §8. Branch `feat/BC-6-openapi`, anchored on `feat/BC-5-search-api` @75530c0 + master @0e6730f.
Plan against the worktree code. No SPEC edit is needed.

## Scope
- `backend/openapi.yaml` (OpenAPI 3.1.1) describing `/healthz`, `/api/cities`, `/api/listings/search` and `/api/openapi.yaml`.
- Embed it with `embed.FS` and serve it at `GET /api/openapi.yaml` with `Content-Type: application/yaml` (RFC 9512 defines no charset parameter, so the value is exactly that).
- Tests: the document handler, the YAML parse, the param sync in both directions, examples equal to the SPEC, and the code enums.
- Run redocly once and record its output. It stays out of `make check`.

## Decisions
1. **Embed location: a package at the module root, `backend/openapi.go` (import path `brillio/backend`, package `backend`).**
   `go.mod` is `module brillio/backend` at `backend/`, so this package sits next to `openapi.yaml`, where SPEC §3 puts it. It imports only `embed`. `internal/httpapi` can import it and there is no import cycle, because the root package imports nothing internal. It exports `const OpenAPIPath = "openapi.yaml"` and `//go:embed openapi.yaml var FS embed.FS`, with a package comment saying why the package exists (go:embed cannot reach a parent directory).
   Rejected alternative: moving the yaml into `internal/httpapi`. That needs a SPEC §3 edit and buries the document the interviewer reads. BC-13's `.air.toml` already watches `yaml` with `root = "."` (backend/), so editing the doc hot-reloads.
2. **YAML parsing: stdlib only, through a strict YAML-subset reader that exists only in the tests.** SPEC §2 says "Stdlib only" and §8 says stdlib `testing`. go.mod/go.sum have no YAML library, and Go's stdlib has no YAML parser.
   - The document stays readable block YAML. The examples (and `type: ["number", "null"]`-style arrays) are written as **JSON flow nodes**. JSON flow is valid YAML, and the reader hands those nodes to `encoding/json`.
   - The reader goes in `internal/httpapi/yamlsubset_test.go`, about 150 lines, and produces `map[string]any` / `[]any` / `string` (plain scalars stay strings, with no type resolution, which the tests do not need). It accepts:
     - block mappings and block sequences (sequences indented deeper than their key);
     - `- key: value` compact items;
     - plain, `"double"` (decoded as a JSON string) and `'single'` scalars;
     - `|` / `|-` literal blocks;
     - full-line `#` comments;
     - JSON flow `[`/`{` values, which may span lines. Continuation lines and the closing bracket must be indented deeper than the key, the same rule as real YAML: redocly 2.63.0 rejects anything less with "deficient indentation", which I checked.
   - It **errors on everything else**: tabs, CR, `---`, anchors, aliases, tags, `?`, `>` folded scalars, duplicate keys, bad dedents, invalid JSON flow, inline ` #` comments, and plain scalars that start with an indicator or contain `: `.
   - Trade-off: 150 lines of hand-rolled test code to defend, in exchange for a readable contract with no new dependency.
   - What "parses as YAML" honestly means here: the strict reader accepts the document, every misreading would fail the content checks against independent sources (the handler allowlist and the SPEC JSON), and redocly's real YAML parser accepts it once.
   - Rejected: a test-only `go.yaml.in/yaml/v3`, which goes against SPEC §2 (see Risks). Also rejected: a document written entirely in JSON syntax. It needs no reader, but it is unreadable, with multi-paragraph descriptions packed into `\n` strings.
3. **Allowlist source:** `httpapi.SearchQueryParams()` in `internal/httpapi/query.go`. It is exported and returns a copy of the package-level `searchParams` array, which the parser reads directly, in §5.3 table order.
4. **Lint posture: the document passes redocly's built-in recommended ruleset with 0 errors and no config file.** I probed it empirically.
   - Root `servers: [{url: /, description: ...}]` is relative to the serving host, which is right because the API serves the doc. It satisfies `no-empty-servers`.
   - Root `security: []` (a public API) satisfies `security-defined`.
   - `info.license {name: Proprietary, identifier: LicenseRef-Proprietary}` satisfies `info-license-strict`. The repo has no LICENSE, so the default is all rights reserved.
   - Every operation gets a `summary` and an `operationId`: `getHealth`, `listCities`, `searchListings`, `getOpenAPIDocument`.
   - No tags.
   - Every component is referenced, so `no-unused-components` stays quiet.
   - Examples must conform to their schemas, because `no-invalid-media-type-examples` is on.
   - **Expected residue: 3 `operation-4xx-response` warnings** (healthz, cities, openapi.yaml). These GETs have no client error, since cities ignores query params. A `default` response does not satisfy the rule, and documenting 405 under `get` would be untrue. The warnings are justified in the lattice comment rather than suppressed.

## Files (worktree-relative)
- NEW `backend/openapi.yaml`: the document, about 480 lines.
- NEW `backend/openapi.go`: root embed package, about 18 lines.
- NEW `backend/internal/httpapi/openapi.go`:
  - `contentTypeYAML = "application/yaml"`;
  - a package var holding the bytes, read once via `fs.ReadFile(backend.FS, backend.OpenAPIPath)` (it panics only if the embed is missing, which the build already rules out);
  - handler `(a *api) openapi(w, _ *http.Request)`, which sets Content-Type and Content-Length and writes the bytes.
- EDIT `backend/internal/httpapi/router.go`: one line after the search route, `mux.HandleFunc("GET /api/openapi.yaml", a.openapi)`. The route inherits request ID, logging, recovery, CORS and the 404/405 fallbacks.
- NEW `backend/internal/httpapi/yamlsubset_test.go`: the reader and its table of reject cases.
- NEW `backend/internal/httpapi/openapi_test.go`: the tests below. It reuses `serve`, `newTestRouter`, `newSearchRouter`, `assertRequestID`, `specExample` and `testOrigin`.
- EDIT `backend/internal/httpapi/router_test.go`: one `TestErrorEnvelopes` row, POST `/api/openapi.yaml` → 405 `method_not_allowed`, Allow `GET, HEAD`.

## Document outline
- `info.description`:
  - JSON UTF-8 camelCase;
  - `X-Request-ID` echoed or generated on every response;
  - CORS (§5.6);
  - unknown route → 404 `not_found` and wrong method → 405 `method_not_allowed` with `Allow`, both in the error envelope;
  - the 500 message is generic.
- `components.headers.RequestID` is referenced from every response.
- `/healthz` → 200 and 503 (`Health`, with examples `{"status":"ok"}` and `{"status":"degraded","error":"database unavailable"}`).
- `/api/cities` → 200 (`CitiesResponse`, with the §5.2 example verbatim) and 500. The description says query params are ignored and the unknown-param rule applies to search only.
- `/api/listings/search`:
  - **9 inline `in: query` params in §5.3 order**, no path-level params. Each has type, constraints, default and a one-line description:
    - `minPrice`, `maxPrice`: number, min 0;
    - `minBedrooms`: integer, 0..9007199254740991;
    - `city`: maxLength 100;
    - `keyword`: maxLength 200;
    - `targetBudget`: exclusiveMinimum 0;
    - `status`: array of enum, `style: form`, `explode: true`, default `["active","pending","sold"]`, comma form noted;
    - `page`: default 1, 1..2^53-1;
    - `pageSize`: default 10, 1..100.
    - Params with no default say "absent: no filter".
  - The operation description covers:
    - the unknown-param 400 (whatever its value, `minprice=` included, with the rationale) and `malformed_query`;
    - empty values count as absent, and a repeated scalar uses its first value;
    - all errors are collected, with their ordering;
    - a page past the end is 200 with empty `items` and correct `page`;
    - an empty city is 200 with `totalItems: 0`;
    - the score formula, tie order, and rank counting across the whole set.
  - Responses:
    - 200 `SearchResponse`, with the §5.4 example verbatim (score 0.7414);
    - 400 `ErrorEnvelope`, with examples `spec` (§5.5 verbatim) and `malformedQuery` (`query`/`malformed_query`/"Query string is not valid URL encoding.");
    - 500.
- `/api/openapi.yaml` → 200 `application/yaml`, schema `type: string`.
- Schemas: `Listing`, `ListingResult` (allOf Listing + `rank` (int ≥1) + `relevance`), `Relevance`, `PageMeta`, `ScoringMeta`, `AppliedFilters`, `SearchResponse`, `ErrorEnvelope`, `FieldError`, `Health`, `CitiesResponse`.
  - Types follow the DTOs in `search.go`, and every key is `required` because none is omitempty. The exceptions are envelope `fields` (only on `validation_error`, minItems 1) and health `error`.
  - Nullable fields use `["<type>", "null"]`: `scoring.targetBudget`, `relevance.budgetFit`, and the 5 scalar `appliedFilters`. `appliedFilters.status` is a non-null array, minItems 1, uniqueItems.
  - Integer fields: bedrooms, sqft, rank, all of PageMeta, recencyHalfLifeDays, minBedrooms. Number fields: price, bathrooms, lat/long, scores, weights, budgets.
  - Constraints are added only where the seed or parser enforces them: price, bedrooms and sqft ≥0; id and source minLength 1; scores 0..1. There is no lat/long range, because the seed does not check one.
  - `listedDate` and `asOf` are `format: date`.
  - `FieldError.code` enum, exactly §5.5: `not_a_number, not_an_integer, negative, not_positive, out_of_range, too_long, invalid_enum, min_gt_max, unknown_param, malformed_query`. `field` is described as the parameter name as sent, or `query` for `malformed_query`.
  - `ErrorEnvelope.error.code` enum: `validation_error, not_found, method_not_allowed, internal_error`.
- YAML footguns:
  - quote `"null"` inside type arrays (use JSON flow);
  - quote response codes (`"200":`, which OAS requires);
  - quote zip-like strings;
  - no inline comments.

## Tests (each maps to an acceptance bullet or a §8 httpapi case)
1. `TestOpenAPIDocument` (acceptance: curl returns the document; §8: status, request ID, CORS). GET `/api/openapi.yaml` with Origin `testOrigin` returns:
   - 200;
   - Content-Type exactly `application/yaml`;
   - `X-Request-ID` present;
   - `Access-Control-Allow-Origin` set;
   - a body byte-equal to `os.ReadFile("../../openapi.yaml")`, which proves the embed serves the repo file.
2. `TestOpenAPIParsesAsYAML` (ticket: "body parses as YAML"). The served body parses under the subset reader, `openapi` starts with `3.1.`, and the `paths` keys are exactly the four SPEC paths. Each path's `get` is served by `newSearchRouter(t, testNow)` with a status listed in its `responses`, which catches doc/router path drift.
3. `TestOpenAPISearchParamsMatchHandler` (acceptance: sync test). Every search parameter is `in: query`, and there are no path-level params and no duplicates. The names are compared with `httpapi.SearchQueryParams()` as **sets, in both directions**, and the failure message lists what is missing from the doc and what is missing from the handler. A secondary assertion checks that the doc order matches the table order.
4. `TestOpenAPIExamplesMatchSPEC` ("examples lifted from §5.4 and §5.5"). The 200 search example deep-equals `specExample("### 5.4")`, the 400 `spec` example deep-equals `specExample("### 5.5")`, and the cities example deep-equals `specExample("### 5.2")`.
   - BC-5's tests already pin handler output to those SPEC blocks, so the doc examples are pinned to the handler by transitivity.
5. `TestOpenAPIErrorCodeEnums`. The `FieldError.code` enum and the envelope `error.code` enum equal the §5.5 lists. The 10 codes are hard-coded in the test with a §5.5 citation, and the list includes `malformed_query`.
6. `TestYAMLSubsetReader`. A small accept case (nested mapping, sequence of mappings, literal block, multi-line JSON flow) plus one reject case for each rejected construct listed in Decision 2, including a closing bracket that is not indented deeper.
7. A `TestErrorEnvelopes` row: POST `/api/openapi.yaml` → 405 envelope.
- Gate: `make check` (golangci-lint with revive: package comment plus doc comments on `FS` and `OpenAPIPath`; unused params named `_`) and `go test -race ./...`.

## Validation (port per run-state: BC-6 API `PORT=18006`, `DB_PATH=:memory:`)
1. From the repo root, build with `cd backend && go build -o bin/api ./cmd/api && cd ..`, then start the server in the background with `PORT=18006 DB_PATH=:memory: ./backend/bin/api &` (the default `LISTINGS_FILE` resolves from the root).
2. `curl -si localhost:18006/api/openapi.yaml | head -15` shows 200, `Content-Type: application/yaml` and `X-Request-Id`.
3. `curl -s localhost:18006/api/openapi.yaml | diff - backend/openapi.yaml` is empty, so the served doc is byte-identical to the file.
4. `curl -si -X POST localhost:18006/api/openapi.yaml` returns the 405 envelope.
5. Kill the server and confirm with `lsof -nP -iTCP:18006 -sTCP:LISTEN`, which should show nothing.
6. Run `npx --yes @redocly/cli@latest lint backend/openapi.yaml` from the worktree root and expect 0 errors and the 3 justified `operation-4xx-response` warnings. Iterating with it while writing the doc is fine. The run that counts is the one on the final committed doc: paste its output into a `lattice comment`, and attach it with steps 2–4 using `--role validation`. Do not add it to the Makefile.

## Shared files and sequencing (run-state)
- `httpapi` router: BC-6 is the later writer after BC-5 (serial by dependency) and adds a one-line route. `router_test.go` gets one table row.
- `docs/SPEC.md`: no edit.
- BC-5 is not on master. Implementation proceeds now, but **`review` waits until BC-5 lands**. Then merge `master` into the branch (never rebase) and re-run `make check`.
- Never commit `.lattice/`.
- For the implementer: `lattice file-link` `backend/openapi.go` (the root embed package decision) and `internal/httpapi/yamlsubset_test.go` (the stdlib-only YAML reader decision).

## Risks and open questions
- **Human, non-blocking: the hand-rolled reader vs. a YAML test dependency.** The plan honors SPEC §2 "Stdlib only". If the human prefers `go.yaml.in/yaml/v3` as a test-only dependency, it would replace the reader (about −150 lines) and need "SPEC edit: §2 — test-only YAML parser exception".
- **Human, non-blocking: two lint and metadata choices.** One: accept the 3 `operation-4xx-response` warnings instead of adding a redocly config or untrue 405s. Two: the `Proprietary` license entry. If the human prefers no license claim, drop it and accept one `info-license` warning.
- Bugs in the reader: covered by the strict rejection, the reject-table test, the content checks against independent sources (tests 3–5), and the redocly run once. Fallback if the reader grows past about 200 lines: a document written entirely in JSON syntax.
- Drift between the response schemas and the DTOs beyond the examples is not tested continuously, because that would need a JSON Schema validator in Go. The examples are checked against the schemas by redocly and pinned to the SPEC and handler by tests 4 and BC-5. Accepted.
- Merge with master after BC-5 lands: an overlap is possible only on the router line.

## Size
About 900 LOC added: YAML about 480, Go production about 55, Go tests about 360. No production dependencies change.
