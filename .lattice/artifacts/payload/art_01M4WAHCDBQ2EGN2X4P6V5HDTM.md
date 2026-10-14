Lattice-Reviewed-Commit: 5c4b9939052f14b34309916df83f3b856898cf16
Lattice-Reviewed-Worktree: /Users/forrest/Code/brillio-worktrees.noindex/bc-6-openapi
Lattice-Reviewed-Base: master (76d6a97bf7f18da629e9294b139ae65147212ca3)
Lattice-Reviewed-Head: feat/BC-6-openapi (5c4b9939052f14b34309916df83f3b856898cf16)

# Code Review: BC-6 — OpenAPI 3.1 document, embedded and served at /api/openapi.yaml

Reviewed: feat/BC-6-openapi @5c4b993 against master @76d6a97 (merge-base confirmed; diff is the 7 BC-6 files, +1352, nothing under `.lattice/`).

## 1. Verdict

**PASS**

Evidence gathered independently in the worktree:

| Check | Result |
|---|---|
| `go build ./... && go vet ./...` | clean |
| `go test -race -count=1 ./...` | 481 passed, 7 packages |
| `golangci-lint run` (standard + errcheck/govet/staticcheck/revive, gofmt) | 0 issues |
| `make check` (both halves; vitest 269/269) | exit 0 |
| `npx --yes @redocly/cli@latest lint backend/openapi.yaml` (resolved 2.63.0) | **0 errors**, 4 warnings: `info-license` ×1, `operation-4xx-response` ×3 (/healthz, /api/cities, /api/openapi.yaml) — exactly the set the plan's Cycle-1 Resolution 6 predicted and justified |
| Served document == `backend/openapi.yaml` | asserted byte-for-byte by `TestOpenAPIDocument`; file is 25817 bytes, matching the implementer's smoke Content-Length |
| SPEC §5.3 "more than 10000 parameters is malformed_query" claim mirrored by the doc | verified against Go: `url.ParseQuery` with 10001 pairs returns "number of URL query parameters exceeded limit" and decodes nothing; 10000 succeeds |

No finding requires rework. Three minor items below are polish; none blocks `in_validation`.

## 2. Summary

I reviewed the document (`backend/openapi.yaml`), the root embed package (`backend/openapi.go`), the handler and one-line route in `internal/httpapi`, the 405 table row, the strict YAML-subset reader, and the six OpenAPI tests against the plan (including its authoritative Cycle-1 resolutions) and the task's acceptance criteria. Quality is high: the search operation description mirrors the merged §5.3 rule for rule (including the three Cycle-1 additions), every schema constraint traces to something the seed loader or query parser actually enforces, every example in the document is pinned by a test to either a SPEC JSON block or a live handler body, and the parameter sync test runs in both directions plus order. The one deviation from plan — the test-only reader is ~247 code lines, past the plan's ~200-line "switch to a JSON-syntax document" fallback — was flagged honestly by the implementer and accepted by the delegator; having read the reader line by line, I uphold that call (reasoning under Issue 4).

## 3. Issues

**[MINOR] backend/internal/httpapi/openapi_test.go:130 — Assertion message misstates what the router does**
The check that each path documents only `get` is correct (HEAD is conventionally left implicit in OpenAPI), but the message says "the router serves GET alone". Go's `ServeMux` `GET` patterns also answer HEAD, and the document's own 405 text advertises `Allow: GET, HEAD`. A future reader of a failure would be told something untrue about the server.
**Fix:** Reword to e.g. `"%s has %v, want only get (HEAD is implicit and not documented)"`.

**[MINOR] backend/openapi.yaml:372 — `Listing.city` (and `zip`, `bathrooms`, `latitude`, `longitude`) have no description**
Sibling properties carry one-line descriptions; these five do not. `city` is the one that matters: the domain type documents that it is trimmed with internal whitespace collapsed, and both the `city` filter and `/api/cities` key on it, so a consumer comparing values benefits from knowing the normalization. The plan only required descriptions on query parameters, so this is optional polish, not a plan miss.
**Fix:** Add one-liners, e.g. `city: City name, trimmed with internal whitespace collapsed; /api/cities lists the distinct values.`; `bathrooms: Bathroom count; halves such as 1.5 are allowed.`; `latitude`/`longitude`: `WGS 84 degrees.`

**[MINOR] Validation hand-off — validator output is summarized, not recorded verbatim**
The task asks that the external validator's output be "recorded in a task comment" and that the curl and validator output be attached `--role validation`. The implementer's 04:15 comment records the resolved version (2.63.0), 0 errors, and names all four warnings with justification, which satisfies the spirit; the verbatim redocly output and curl transcript are not yet attached. Not a code defect and not the review stage's gate.
**Fix:** The validation sub-agent should paste the full `npx --yes @redocly/cli@latest lint backend/openapi.yaml` output (with `--version`) and the `curl -si` / `diff` / POST-405 transcript from plan Validation steps 2–6 as the `--role validation` attachment.

**[MINOR, decision confirmed — no change requested] backend/internal/httpapi/yamlsubset_test.go — Reader size vs. the plan's fallback threshold**
The reader is ~247 non-blank, non-comment lines before its test (405 lines total), versus the plan's estimate of ~150 and a fallback to a JSON-syntax document "past about 200". The delegator accepted the deviation and asked the reviewer to re-weigh. I would keep the reader:
- Every accepted construct (block mappings/sequences, `- key: value`, plain and double-quoted scalars, double-quoted keys, `|`/`|-`, full-line comments, multi-line JSON flow) is used by the document; the implementer also *narrowed* the subset (no single-quoted scalars, no quoted keys in compact items), which is the right direction.
- Roughly a third of the length is line-numbered, rule-naming error paths — which Cycle-1 Resolution 4 explicitly asked for — and the 33-case reject table.
- I traced the reader for places it could silently diverge from YAML 1.2 and found it strictly narrower everywhere: tabs, CR, trailing whitespace, document markers, anchors/aliases/tags, folded and `|+` scalars, inline comments after plain/quoted/flow values, multi-line double-quoted scalars, non-JSON flow, YAML-only escapes, `: ` in plain scalars, bad dedents and same-column sequences are all hard errors with the offending line. A misread therefore surfaces as a rejection, never a wrong value.
- Correctness is cross-checked three ways on the real bytes: redocly's real YAML parser accepts the same file; the parsed examples deep-equal the SPEC JSON blocks and live handler bodies; the implementer's one-off js-yaml diff was clean.
- The fallback would sacrifice exactly the readability the ticket exists to provide.
**Fix:** None. If the human prefers the threshold be honored literally, that is a plan-level choice, not an implementation defect.

## 4. Positive Observations

- **Contract fidelity.** The search description carries every §5.3 rule, including the three Cycle-1 additions (10000-parameter `malformed_query`, code-point lengths on `city`/`keyword`, empty-first-value-is-absent). The scoring/ordering prose (asymmetric budget fit, 30-day half-life, UTC calendar-day age with future dates at 0, unrounded-score sort, listedDate → price → source → id tie-break, rank across the whole set, four-decimal rounding) matches `domain/scoring.go` and `domain/rank.go` exactly. CORS, request-ID (32 hex chars), 404/405/500 conventions and the two-second health timeout all match `middleware.go`/`health.go`.
- **Constraints only where enforced.** `minLength: 1` on id/source and `minimum: 0` on price/bedrooms/sqft trace to `seed.go`; `maximum: 9007199254740991` on `minBedrooms`/`page` and `1..100` on `pageSize` trace to `query.go`; no invented lat/long range. Nullable fields map one-for-one to the pointer DTO fields in `search.go`; `appliedFilters.status` is correctly non-null with `minItems: 1`.
- **Tests pin to independent sources, not to the document itself.** `httpapi.SearchQueryParams()` for the allowlist (both directions, plus §5.3 order); `specExample` for §5.2/§5.4/§5.5; and, beyond the plan, `TestOpenAPIExamplesMatchHandler` pins the four examples the SPEC does not spell out (healthz ok/degraded, malformedQuery, InternalError) to live handler bodies — so all six examples in the document are tested. `TestOpenAPIParsesAsYAML` also catches doc/router path drift by hitting each documented path and checking the status is a documented response.
- **Minimal, correct serving path.** Root package `brillio/backend` with a package comment explaining the `go:embed` parent-directory limitation; no import cycle; document read once at init; `Content-Type` exactly `application/yaml` (RFC 9512, no charset) and `Content-Length` set; `_ *http.Request` satisfies revive; the route inherits request-ID, logging, recovery, CORS and the JSON 404/405 fallbacks for free, and the new `TestErrorEnvelopes` row proves the 405 with `Allow: GET, HEAD`.
- **Honest lint posture.** No redocly config to hide warnings, no untrue 405 responses under `get` to appease `operation-4xx-response`, and no invented license claim; the four warnings are each justified in the task comment.
- **Reader discipline.** Errors name the line and the rule; the accept case exercises a quoted key, a quoted value containing `: `, a multi-paragraph `|` block and a `|-` block with exact chomping strings, a compact `- key: value` item, and multi-line JSON flow; the document's header comment names the subset and points at the enforcing file, as Resolution 4 required.
- **Process hygiene.** Deviations were declared in the implementation comment rather than buried; file-links were recorded for the two design decisions; nothing under `.lattice/` was committed; the branch was merged from master (not rebased) after BC-5 landed, as the plan sequenced.
