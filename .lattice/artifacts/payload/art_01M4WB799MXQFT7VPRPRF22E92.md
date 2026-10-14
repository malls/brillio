# Code review BC-6 (reviewer: fable sub-agent, decision 04:02Z)

Verdict: PASS
Reviewed HEAD: 70518b51a3e3c9c655185be40e6bf98d527c0fe1
Base: master @bf36a3d13d12b2333f6190f4871191dada21656b

Merge-base of `master` and HEAD is `bf36a3d` (equal to master), so `git diff master...HEAD` is the exact branch delta: 7 files, all under `backend/` (`openapi.yaml`, `openapi.go`, `internal/httpapi/{openapi.go,openapi_test.go,yamlsubset_test.go,router.go,router_test.go}`), +1360 lines, nothing under `.lattice/`. `backend/go.mod` and `go.sum` are unchanged versus master. The BC-6 delta since the prior review (`git diff 5c4b993..70518b5 -- backend/internal backend/openapi.yaml backend/openapi.go`) is 2 files: one reworded assertion message and six new `description` lines on `Listing` properties; the rest of that range is BC-13's `.air.toml`, `.dockerignore` and `Dockerfile.dev` arriving via the master merge.

## Checks run

| Command | Result |
|---|---|
| `make check` (from worktree root) | exit 0. `golangci-lint run` 0 issues; eslint/tsc/prettier clean; `go test -race ./...` ok (7 packages); vitest 22 files, 382 tests passed |
| `cd backend && go test -race -count=1 ./internal/httpapi/` | exit 0, 274 tests passed; the 8 BC-6-relevant tests (`TestOpenAPIDocument`, `TestOpenAPIParsesAsYAML`, `TestOpenAPISearchParamsMatchHandler`, `TestOpenAPIExamplesMatchSPEC`, `TestOpenAPIExamplesMatchHandler`, `TestOpenAPIErrorCodeEnums`, `TestYAMLSubsetReader`, `TestErrorEnvelopes`) each PASS under `-v` |
| `cd backend && go build ./... && go vet ./... && go test -race -count=1 ./...` (uncached) | exit 0, all 5 test packages ok |
| `gofmt -l` on the 6 Go files in the diff | clean |
| `npx --yes @redocly/cli@latest --version` | 2.63.0 |
| `npx --yes @redocly/cli@latest lint backend/openapi.yaml` | exit 0. "Your API description is valid." **0 errors, 4 warnings**: `info-license` at `openapi.yaml:7:1` (#/info); `operation-4xx-response` at `:44:7` (/healthz), `:82:7` (/api/cities), `:318:7` (/api/openapi.yaml). Exactly the set the plan's Cycle-1 Resolution 6 predicts and justifies. Full output saved at `scratchpad/redocly.log` |
| `git diff master...HEAD --stat -- backend/go.mod backend/go.sum` | empty (stdlib only, SPEC section 2 honored) |
| `grep` of `$(go env GOROOT)/src/net/url/url.go` (go1.27.1) | `defaultMaxParams = 10000`, error "number of URL query parameters exceeded limit": the document's 10000-parameter `malformed_query` claim is grounded in the toolchain actually in use |

No servers were started.

## Findings

### Critical
None.

### Major
None.

### Minor
None.

### Nit
- `backend/internal/httpapi/openapi_test.go:132-137` — The per-path drift check accepts *any* documented status, so for `/api/cities` and `/api/listings/search` a 500 would pass it, and for `/healthz` a 503 would. Since `newSearchRouter` seeds a real store, all four plain GETs return 200 in practice, so asserting `rec.Code == http.StatusOK` and that `"200"` is among the documented responses would be strictly stronger at the same cost and still catch a renamed route (404 is documented nowhere). Optional hardening of a secondary assertion whose stated purpose (path drift) it already serves; no change requested before validation.

## Verification detail (independent pass)

**The two fixes from a0fd9bf.**
- `openapi_test.go:130` now reads `"%s has %v, want only get (HEAD is implicit and not documented)"`. True: Go's `ServeMux` `GET` patterns also match HEAD, the document lists only `get` under each path, and the 405 convention text advertises `Allow: GET, HEAD`.
- Every new `Listing` description traces to code with no invented claim:
  - `city` "trimmed with internal whitespace collapsed" = `domain/listing.go:68` and `seed.collapseSpace` (`seed.go:155,196-200`); "the `city` filter matches it ignoring case" = `store/listings.go:33` `city = ? COLLATE NOCASE`, mirroring SPEC section 5.3's "case-insensitive exact match"; "`GET /api/cities` lists the distinct non-empty names" = `selectCities` with `WHERE city <> ''` and `GROUP BY city COLLATE NOCASE` (`listings.go:49-51`).
  - `zip` "as the feed wrote it, trimmed. Its format is not checked" = `seed.go:157` `strings.TrimSpace`, and no zip validation anywhere in `parseRecord`.
  - `bedrooms` "Number of bedrooms" = `listing.go:76`.
  - `bathrooms` "Halves are allowed, such as 1.5" = `listing.go:78` verbatim; the seed accepts any float64, and the sentence does not claim exclusivity.
  - `latitude`/`longitude` "decimal degrees (WGS 84)" = `listing.go:82,84`.

**Acceptance criteria and plan.**
- Code enums: `FieldError.code` (`openapi.yaml:586-596`) is exactly SPEC section 5.5's ten values in order, including `malformed_query` (run-state 03:00Z); `ErrorEnvelope.error.code` (`:548`) is exactly the four-row section 5.5 table. Both pinned by `TestOpenAPIErrorCodeEnums` against hard-coded lists with section citations.
- Examples: the 200 search example carries `score: 0.7414` (run-state 00:47Z) and deep-equals SPEC section 5.4; the 400 `spec` example deep-equals section 5.5; the cities example deep-equals section 5.2 (`TestOpenAPIExamplesMatchSPEC`). The four examples the SPEC does not spell out (healthz ok/degraded, `malformedQuery`, `InternalError`) deep-equal live handler bodies (`TestOpenAPIExamplesMatchHandler`); the degraded case uses a pinger error of "disk on fire" and still matches `"database unavailable"`, which proves `health.go:29` emits a fixed reason and that `Health.error`'s description is true.
- Sync test: `TestOpenAPISearchParamsMatchHandler` computes both `accepted but not documented` and `documented but not accepted` against `httpapi.SearchQueryParams()`, rejects duplicates, requires `in: query`, forbids path-level `parameters`, and then checks section 5.3 order. `SearchQueryParams` clones the same `searchParams` array that `unknownParams` reads and whose constants every per-parameter reader uses (`query.go:17-43, 410-421`), so the allowlist is a single source of truth.
- Route: one line at `router.go:65`, inheriting request-ID, logging, recovery, CORS and the JSON 404/405 fallbacks; the new `TestErrorEnvelopes` row proves POST yields the 405 envelope with `Allow: GET, HEAD`.
- Handler: `Content-Type` is exactly `application/yaml` (no charset, RFC 9512); explicit `Content-Length` is the right call for a ~26 KB body that would otherwise be chunked; the ignored write error carries the same comment as `respond.go:75` (project idiom, not a BC-6 deviation); `_ *http.Request` satisfies revive. The embedded bytes are read once at init via `fs.ReadFile(backend.FS, backend.OpenAPIPath)` and `TestOpenAPIDocument` asserts the served body is byte-equal to `../../openapi.yaml` and that `Content-Length` equals the file length.
- Embed: root package `brillio/backend` (`backend/openapi.go`) with a package comment explaining the go:embed parent-directory limitation, doc comments on `OpenAPIPath` and `FS`, and no internal imports (no cycle). BC-13's `.dockerignore` is `*` with a bind-mounted `/app` and no `COPY`, so the embed is unaffected in the Compose dev image, and `.air.toml` watches `yaml`. The Vite proxy forwards `/api`, so the document is reachable through the dev server too.
- Document prose versus code (sampled widely, not just the merged-section additions): request ID echoed verbatim or 32 hex chars (`middleware.go:38-58`); CORS methods `GET, OPTIONS`, headers `Content-Type, X-Request-ID`, expose `X-Request-ID`, preflight 204 (`middleware.go:19-24, 120-138`); health two-second timeout (`health.go:11`); 404/405 fallbacks with `Allow` (`router.go:88-109`); `fields` `omitempty` so "present only when the code is validation_error" holds (`respond.go:45`); `appliedFilters.city` trimmed and `keyword` as received (`search.go:196-205`, `query.go` `city()`/`keyword()`); budget-fit formula, 30-day half-life, UTC calendar-day age with future dates at 0, 0.6/0.4 weights or recency alone (`scoring.go`); sort by unrounded score then listedDate desc, price asc, source, id (`rank.go:71-85`); `totalPages` ceil and 0 when empty, `page` echoed even past the end (`paginate.go:42,73`); `maximum: 9007199254740991` on `minBedrooms`/`page` matches `maxExactInt = 1<<53` with `>=` rejection; `pageSize` 1..100 matches `domain.MaxPageSize`.
- Lint posture per Resolutions 6 and 8: no `info.license`, no redocly config, no untrue 405s under `get`; exactly one `info-license` and three `operation-4xx-response` warnings.
- YAML-subset reader: I re-traced `mapping`, `sequence`, `value`, `splitKey`, `plainRule`, `literal` and `flow` against every construct the document uses (quoted keys, `$ref: "#/..."` compact items, `- url: /`, multi-line `required:` flow arrays, multi-line JSON examples, `|` blocks containing `- ` bullets and blank lines, `|-` blocks) and found no path where a misread yields a wrong value rather than a line-numbered rejection; the Resolution 2 and 3 accept cases and the 33-row reject table are present. The deviation from the plan's accept list (single-quoted scalars rejected, quoted keys not accepted in compact items) narrows the subset and is declared in the file header.

## Prior findings

Status of the four findings in art_01M4WAHCDBQ2EGN2X4P6V5HDTM (review of 5c4b993):

1. **[MINOR] `openapi_test.go:130` untrue assertion message** — Fixed in a0fd9bf. Verified the new wording is true (see above).
2. **[MINOR] `Listing.city`/`zip`/`bathrooms`/`latitude`/`longitude` without descriptions** — Fixed in a0fd9bf (and `bedrooms` gained one too). Every sentence verified against `domain/listing.go`, `store/seed/seed.go` and `store/listings.go`; nothing untrue or invented.
3. **[MINOR] Validator output summarized, not recorded verbatim** — Not a code issue; still owed at the validation stage. The verbatim 2.63.0 output for this HEAD is in the Checks table above and at `scratchpad/redocly.log`, but the validation sub-agent must still run it on the final committed document and attach it with the curl/diff/POST-405 transcript under `--role validation`.
4. **[MINOR, no change requested] Reader size over the plan's ~200-line fallback** — Upheld. Nothing in the delta touched `yamlsubset_test.go`; my independent trace agrees that the length buys strictness and line-numbered errors (Resolution 4), not permissiveness.

## Summary

PASS. Both prior minors are fixed and verified true against the code, the merge from master adds nothing under `backend/internal` or SPEC section 5 and the branch still differs from master by exactly the 7 BC-6 files, and my independent pass over the whole diff found no defect: the enums and examples match SPEC exactly, the sync test is set equality in both directions plus order against the parser's own allowlist, the route is one line with `Content-Type: application/yaml`, the embed is stdlib-only with `go.mod`/`go.sum` untouched, lint, race tests and redocly (0 errors, the 4 justified warnings) are all clean. The single nit is an optional strengthening of a secondary test assertion and does not need to land before validation. Ready for `in_validation`.
