# BC-6 Validation: OpenAPI 3.1 document served at /api/openapi.yaml

- Ticket: BC-6 (task_01M4VXK2J03TXW4B4D9KGWQ145), "OpenAPI 3.1 document, embedded and served at /api/openapi.yaml, with param sync test"
- Validator: agent:opus-5.5-validator
- Branch / HEAD: `feat/BC-6-openapi` @ `70518b51a3e3c9c655185be40e6bf98d527c0fe1` (confirmed with `git rev-parse HEAD`; worktree clean before and after)
- Date/time (UTC): 2026-10-14, 04:38:08Z to 04:39:53Z
- Environment: macOS (darwin/arm64), go1.27.1, node v24.21.0, npx 11.19.0, @redocly/cli 2.63.0 (resolved from `@latest`)
- Server: real binary built from `./cmd/api` into a scratch directory outside the repo (no binary left in the worktree), started from the worktree root with `PORT=18006 DB_PATH=:memory:`. All other env at SPEC section 11 defaults (`LISTINGS_FILE=./data/sample_listings.json` resolves from the worktree root). Port 18006 was free before start.

## 1. Build and start the server: PASS

```
$ cd backend && go build -o $SCRATCH/bc6-api ./cmd/api      # BUILD_OK, 15.4M binary in scratch
$ PORT=18006 DB_PATH=:memory: nohup $SCRATCH/bc6-api > $SCRATCH/bc6-server.log 2>&1 &   # PID 67046
$ for i in 1..10: curl -s -m 1 localhost:18006/healthz && break; sleep 1
{"status":"ok"}
ready after try 2
```

Server log (first two lines):

```
{"time":"2026-10-14T00:38:29.085038-04:00","level":"INFO","msg":"seeded 12 listings","count":12,"listings_file":"./data/sample_listings.json","db_path":":memory:"}
{"time":"2026-10-14T00:38:29.085531-04:00","level":"INFO","msg":"listening","addr":"[::]:18006","log_level":"INFO","cors_origins":["http://localhost:5173"],"shutdown_timeout":"10s"}
```

## 2. GET /api/openapi.yaml headers: PASS

```
$ curl -si localhost:18006/api/openapi.yaml | head -20
HTTP/1.1 200 OK
Content-Length: 26360
Content-Type: application/yaml
Vary: Origin
X-Request-Id: 5d7ff6a83cecec386b241d57b935cdc5
Date: Wed, 14 Oct 2026 04:38:33 GMT

# OpenAPI 3.1 description of the listing search API (docs/SPEC.md section 5), served
# at GET /api/openapi.yaml. Written in a strict YAML subset: block mappings and
# sequences, plain, double-quoted and literal (|) scalars, and JSON flow values. No
# anchors, aliases, tags, single-quoted or folded scalars, inline comments or tabs.
# backend/internal/httpapi/yamlsubset_test.go enforces the subset.
openapi: 3.1.1
info:
  title: Brillio listing search API
  version: 1.0.0
  description: |
    A read-only API over the sample property listings. Filter the listings, rank them by how
    well they fit a target budget and how recently they were listed, and page through the
    results.
```

Observed: `HTTP/1.1 200 OK`, `Content-Type: application/yaml`, `X-Request-Id` present, `Content-Length: 26360` (equals the file size).

## 3. Served body is byte-identical to the committed file: PASS

```
$ curl -s localhost:18006/api/openapi.yaml | diff - backend/openapi.yaml && echo IDENTICAL
IDENTICAL
$ curl -s localhost:18006/api/openapi.yaml | wc -c ; wc -c < backend/openapi.yaml
   26360
   26360
$ curl -s localhost:18006/api/openapi.yaml | shasum -a 256 ; shasum -a 256 backend/openapi.yaml
d1111158cbb04b80af989c5f177a2b9eadb435ad1f670d5dd9c2a55fb64d1ff3  -
d1111158cbb04b80af989c5f177a2b9eadb435ad1f670d5dd9c2a55fb64d1ff3  backend/openapi.yaml
$ curl -s localhost:18006/api/openapi.yaml | head -12
# OpenAPI 3.1 description of the listing search API (docs/SPEC.md section 5), served
# at GET /api/openapi.yaml. Written in a strict YAML subset: block mappings and
# sequences, plain, double-quoted and literal (|) scalars, and JSON flow values. No
# anchors, aliases, tags, single-quoted or folded scalars, inline comments or tabs.
# backend/internal/httpapi/yamlsubset_test.go enforces the subset.
openapi: 3.1.1
info:
  title: Brillio listing search API
  version: 1.0.0
  description: |
    A read-only API over the sample property listings. Filter the listings, rank them by how
    well they fit a target budget and how recently they were listed, and page through the
```

## 4. HEAD works; POST returns the 405 envelope: PASS

```
$ curl -sI localhost:18006/api/openapi.yaml
HTTP/1.1 200 OK
Content-Length: 26360
Content-Type: application/yaml
Vary: Origin
X-Request-Id: c7b0c072c76dd73aec81019cbe1539d9
Date: Wed, 14 Oct 2026 04:38:36 GMT

$ curl -si -X POST localhost:18006/api/openapi.yaml
HTTP/1.1 405 Method Not Allowed
Allow: GET, HEAD
Content-Type: application/json; charset=utf-8
Vary: Origin
X-Request-Id: 3e0cdf899c531d45e4729ea2e84b2c29
Date: Wed, 14 Oct 2026 04:38:36 GMT
Content-Length: 72

{"error":{"code":"method_not_allowed","message":"Method not allowed."}}
```

Extra check: `curl -si -X PUT localhost:18006/api/openapi.yaml` also returned `HTTP/1.1 405 Method Not Allowed` with `Allow: GET, HEAD`.

## 5. The document describes the live API: PASS

Live calls (bodies trimmed):

```
$ curl -s 'localhost:18006/api/listings/search?targetBudget=450000' | head -c 600      # HTTP 200
{"items":[{"rank":1,"id":"A1","source":"MLS_A","address":"123 Main St, Apt 4B","city":"Springfield","state":"VA","zip":"22150","price":450000,"bedrooms":2,"bathrooms":1.5,"sqft":980,"latitude":38.7893,"longitude":-77.1873,"listedDate":"2026-08-29","status":"active","description":"Bright top-floor condo near shops and transit. Pet friendly.","relevance":{"score":0.7382,"budgetFit":1,"recency":0.3455}},{"rank":2,"id":"B7","source":"MLS_B","address":"123 Main Street, Unit 4B","city":"Springfield","state":"VA","zip":"22150","price":452000,"bedrooms":2,"bathrooms":1.5,"sqft":980,"latitude":38.7893,
$ ... | jq -c '{page, scoring, appliedFilters}'
{"page":{"page":1,"pageSize":10,"totalItems":12,"totalPages":2},"scoring":{"targetBudget":450000,"budgetWeight":0.6,"recencyWeight":0.4,"recencyHalfLifeDays":30,"asOf":"2026-10-14"},"appliedFilters":{"minPrice":null,"maxPrice":null,"minBedrooms":null,"city":null,"keyword":null,"status":["active","pending","sold"]}}

$ curl -s -w '\nHTTP %{http_code}\n' 'localhost:18006/api/listings/search?minprice=1'
{"error":{"code":"validation_error","message":"Invalid search parameters.","fields":[{"field":"minprice","code":"unknown_param","message":"Unknown parameter."}]}}
HTTP 400

$ curl -s -w '\nHTTP %{http_code}\n' 'localhost:18006/api/listings/search?city=%'
{"error":{"code":"validation_error","message":"Invalid search parameters.","fields":[{"field":"query","code":"malformed_query","message":"Query string is not valid URL encoding."}]}}
HTTP 400
```

Comparison method: the document was converted to JSON in scratch with `npx --yes @redocly/cli@latest bundle backend/openapi.yaml --ext json -o $SCRATCH/openapi.json` (no YAML parser is installed locally), then the live bodies above (plus the POST 405 body) were checked against it with one `jq` program. Result, verbatim:

```
{
  "search200_top_keys_eq_SearchResponse_props": true,
  "search200_item_keys_eq_Listing_plus_rank_relevance": true,
  "search200_every_item_has_all_required": true,
  "relevance_keys_eq": true,
  "page_keys_eq": true,
  "scoring_keys_eq": true,
  "appliedFilters_keys_eq": true,
  "item_count": 10,
  "unknown_param_body_code_in_FieldError_enum": true,
  "unknown_param_envelope_code_in_enum": true,
  "unknown_param_fieldError_keys_eq": true,
  "malformed_body_equals_doc_example_malformedQuery": true,
  "method_not_allowed_code_in_enum": true
}
```

What this shows:
- 200: top-level keys (`appliedFilters, items, page, scoring`) equal the `SearchResponse` properties; every item's keys equal `Listing` properties plus `rank` and `relevance` (`ListingResult` = allOf Listing + {rank, relevance}), and every item carries all required fields; `relevance`, `page`, `scoring` and `appliedFilters` keys equal the `Relevance`, `PageMeta`, `ScoringMeta` and `AppliedFilters` properties.
- 400 `unknown_param`: envelope code `validation_error` is in the `ErrorEnvelope.error.code` enum; `unknown_param` is in the `FieldError.code` enum (enum is exactly the 10 section 5.5 codes); the field-error keys equal `field, code, message`.
- 400 `malformed_query`: the live body is deep-equal to the document's `malformedQuery` example on the search 400 response (field `query`, code `malformed_query`).
- 405: `method_not_allowed` is in the envelope code enum.
- The search operation documents exactly the nine parameters `minPrice, maxPrice, minBedrooms, city, keyword, targetBudget, status, page, pageSize`.

## 6. Stop the server; port 18006 free: PASS

```
$ kill 67046
process 67046 exited
$ lsof -nP -iTCP:18006 -sTCP:LISTEN
lsof exit 1 (1 = nothing listening)
$ pgrep -fl bc6-api
pgrep exit 1
```

Server log tail: `"msg":"shutting down","timeout":"10s"` then `"msg":"server stopped"` (graceful shutdown on SIGTERM). Every request in the log matches the calls above (200 for the GETs and HEAD, 405 for POST/PUT, 400 for the two bad searches).

## 7. External validator, redocly (review finding 3; plan Resolution 5): PASS

Run from the worktree root. Output verbatim except that one emoji was stripped from the final banner line.

```
$ npx --yes @redocly/cli@latest --version
2.63.0

$ npx --yes @redocly/cli@latest lint backend/openapi.yaml 2>&1      # exit 0
No configurations were provided -- using built in recommended configuration by default.

validating backend/openapi.yaml...
[1] backend/openapi.yaml:7:1 at #/info

Info object should contain `license` field.

5 | # backend/internal/httpapi/yamlsubset_test.go enforces the subset.
6 | openapi: 3.1.1
7 | info:
  | ^^^^
8 |   title: Brillio listing search API
9 |   version: 1.0.0

Warning was generated by the info-license rule.

Reference: https://redocly.com/docs/cli/rules/oas/info-license


[2] backend/openapi.yaml:44:7 at #/paths/~1healthz/get/responses

Operation must have at least one `4XX` response.

42 |   Reports whether the service can reach its database, for liveness and readiness checks.
43 |   It sits outside `/api` because it is about the service, not the listings.
44 | responses:
   | ^^^^^^^^^
45 |   "200":
46 |     description: The service and its database are up.

Warning was generated by the operation-4xx-response rule.

Reference: https://redocly.com/docs/cli/rules/oas/operation-4xx-response


[3] backend/openapi.yaml:82:7 at #/paths/~1api~1cities/get/responses

Operation must have at least one `4XX` response.

80 |   data. Query parameters are ignored: the rule that rejects unknown parameters applies
81 |   to search only.
82 | responses:
   | ^^^^^^^^^
83 |   "200":
84 |     description: The city names.

Warning was generated by the operation-4xx-response rule.

Reference: https://redocly.com/docs/cli/rules/oas/operation-4xx-response


[4] backend/openapi.yaml:318:7 at #/paths/~1api~1openapi.yaml/get/responses

Operation must have at least one `4XX` response.

316 | summary: Get this OpenAPI document
317 | description: Returns this document, byte for byte as it is embedded in the server.
318 | responses:
    | ^^^^^^^^^
319 |   "200":
320 |     description: The OpenAPI 3.1 document.

Warning was generated by the operation-4xx-response rule.

Reference: https://redocly.com/docs/cli/rules/oas/operation-4xx-response


backend/openapi.yaml: validated in 56ms

Woohoo! Your API description is valid.
You have 4 warnings.

```

Result: 0 errors, exactly 4 warnings: 1 `info-license` and 3 `operation-4xx-response` (on `/healthz`, `/api/cities`, `/api/openapi.yaml`), as expected by plan Resolution 6.

Justification for each warning (0 errors; 4 warnings, all accepted per plan Decision 4 and Resolutions 5, 6 and 8; no redocly config file is added):
1. info-license (#/info): deliberately no license claim. The repo has no LICENSE file, and asserting one (for example LicenseRef-Proprietary) is a legal claim the author is not placed to make (plan Resolution 6).
2. operation-4xx-response on GET /healthz: the operation has no client-error response (it is a liveness/readiness check with no parameters). A `default` response does not satisfy the rule, and documenting 405 under `get` would be untrue, since a 405 is never the response to a GET (plan Decision 4).
3. operation-4xx-response on GET /api/cities: no client-error response; query parameters are ignored there (the unknown-parameter rule applies to search only). A `default` response does not satisfy the rule, and a 405 under `get` would be untrue (plan Decision 4).
4. operation-4xx-response on GET /api/openapi.yaml: no client-error response; it returns the embedded document byte for byte. A `default` response does not satisfy the rule, and a 405 under `get` would be untrue (plan Decision 4).

## 8. Sync tests, for the record: PASS

```
$ cd backend && go test -race -count=1 -run 'OpenAPI|YAMLSubset' -v ./internal/httpapi/ 2>&1 | grep -E '^(=== RUN|--- (PASS|FAIL)|ok|FAIL)' | grep -v '/'
=== RUN   TestOpenAPIDocument
--- PASS: TestOpenAPIDocument (0.00s)
=== RUN   TestOpenAPIParsesAsYAML
--- PASS: TestOpenAPIParsesAsYAML (0.03s)
=== RUN   TestOpenAPISearchParamsMatchHandler
--- PASS: TestOpenAPISearchParamsMatchHandler (0.00s)
=== RUN   TestOpenAPIExamplesMatchSPEC
--- PASS: TestOpenAPIExamplesMatchSPEC (0.00s)
=== RUN   TestOpenAPIExamplesMatchHandler
--- PASS: TestOpenAPIExamplesMatchHandler (0.02s)
=== RUN   TestOpenAPIErrorCodeEnums
--- PASS: TestOpenAPIErrorCodeEnums (0.00s)
=== RUN   TestYAMLSubsetReader
--- PASS: TestYAMLSubsetReader (0.00s)
```

The `grep -v '/'` in the prescribed filter also drops the package summary line because it contains the import path. From the same run, unfiltered: `PASS` and `ok  	brillio/backend/internal/httpapi	1.378s` (go test exit 0). The run had 41 passing subtests and 0 `--- FAIL` lines. `TestOpenAPISearchParamsMatchHandler` passes.

## Overall: PASS

The real binary serves `/api/openapi.yaml` with 200, `application/yaml`, `X-Request-Id` and `Content-Length`. The body is byte-identical to `backend/openapi.yaml` (same SHA-256). HEAD works, and POST/PUT return the 405 `method_not_allowed` envelope with `Allow: GET, HEAD`. The live search 200, `unknown_param` and `malformed_query` responses match the document's schemas, enums and the `malformedQuery` example. redocly 2.63.0 reports 0 errors and the 4 expected, justified warnings. The race-enabled sync tests, including `TestOpenAPISearchParamsMatchHandler`, pass. The server was stopped and port 18006 is free.
