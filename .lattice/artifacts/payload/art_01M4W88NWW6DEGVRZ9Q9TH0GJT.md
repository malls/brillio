# Code Review: BC-5 (Search API) — cycle 2 re-review

Verdict: PASS
Reviewed HEAD: 1871b0127d1e942ca509a41e63a21be9493c4ec5
Base: master 0e6730fcfbd67c41afecd8c35bc346f22da4c6d9
Reviewer: agent:opus-5.5-reviewer (fresh-context own-reviewer fallback; diff exceeds 1500 changed lines, run-state decision 03:22Z)
Prior review: 75530c0 PASS-WITH-NITS (art_01M4W7VYKS1KWBVC8R1NX7T89M)

## Gate

- `make check` from the worktree root: **exit 0**.
  - `golangci-lint run`: `0 issues.`
  - eslint (`--max-warnings 0`), `tsc -b --noEmit` and `prettier --check .` are clean ("All matched files use Prettier code style!").
  - `go test -race ./...`: ok for config, domain, httpapi, store and store/seed. The results came from the Go test cache, so the uncached run below is the one that counts.
  - vitest: `Test Files 13 passed (13)`, `Tests 228 passed (228)`. Cycle 1 had 10 files and 197 tests. The extra 3 files and 31 tests are BC-11's StatusLine, Pager and LoadingLine, which came in with the merge.
- `cd backend && go test -race -count=1 ./...`: **exit 0**, all 5 test packages ok, uncached.
  - Counted with `-json`: 65 top-level tests and 367 subtests passed, 0 failed, 0 skipped.
  - These counts are identical to cycle 1, as expected for a byte-identical backend.
- `git diff --check 75530c0 fd15b4a` reports no whitespace errors. The SPEC lines fd15b4a adds are ASCII only.
- The worktree is still clean (`git status --porcelain` is empty).
  - Probes ran in a scratch copy made with `git archive HEAD`, never in the worktree.
  - The make-check log is at `scratchpad/make-check-1871b01.log`.

## Nit fixes (1-4: fixed / not fixed, evidence)

fd15b4a changes only docs/SPEC.md (+7 -3), and its parent is 75530c0. It has three hunks: two §5.3 bullets edited, one §5.3 sentence appended, and two HTML comments added.

1. **Nit 1 (the 10000-parameter cap gives `malformed_query`): fixed, accurately.**
   - New text: "The same error is returned when the query has more than 10000 `&`-separated parameters (Go's `url.ParseQuery` limit); then nothing decodes, so nothing else is validated."
   - Go 1.27.1 source (`net/url/url.go:979-982`): `parseQuery` checks `strings.Count(query, "&") + 1` against `defaultMaxParams = 10000`. On failure it returns before it decodes anything, so `ParseQuery` returns an empty `Values`.
     - Nothing in the repo sets `urlmaxqueryparams` (there is no `godebug` in go.mod, no `//go:debug` line and no GODEBUG setting), so the default applies.
   - `parseSearchRequest` (query.go:178-186) runs `ParseSearchQuery` on that empty map, which yields no errors. It then prepends the single `malformed_query`. So "nothing else is validated" is exactly what happens.
   - Probe through the router (scratch copy):
     - `minPrice=1&` x9999 then `minprice=1` is 10000 segments: 400 with only `{minprice, unknown_param}`. It decodes and validates normally.
     - `minPrice=1&` x10000 then `minprice=1` is 10001 segments: 400 with only `{query, malformed_query, "Query string is not valid URL encoding."}`.
     - The boundary sits exactly where the SPEC puts it.
   - Not a finding: 10000 real parameters plus a trailing `&` also count as 10001 by Go's rule, and the request is rejected. "`&`-separated parameters" still covers that case, since the empty piece is an `&`-separated piece. No real client is near the cap.
2. **Nit 2 (lengths count Unicode code points): fixed, accurately.**
   - New text: "Lengths count Unicode code points (characters): `city` after trimming, `keyword` as received."
   - This matches `utf8.RuneCountInString` in `city()` (after `strings.TrimSpace`) and in `keyword()` (on the raw value, after the whitespace-only check), at query.go:265-284.
   - Probe: 100 copies of U+1F3E0 (4 UTF-8 bytes, 2 UTF-16 units each) in `city` is accepted. 101 copies gives `too_long`. The limit is in code points, not bytes or UTF-16 units.
3. **Nit 3 (`minPrice=&minPrice=5` is absent): fixed, accurately.**
   - New text: "..., and an empty first value is absent (`minPrice=&minPrice=5` is absent, not 5)."
   - `scalar()` (query.go:201-207) returns `vs[0]`, where "" means absent. It never falls through to a later non-empty value.
   - It is pinned by the existing row `{"empty first value is absent", "minPrice=&minPrice=5", ...}` (query_test.go:211), which expects the default query.
   - Confirmed end to end in the probe: `search("minPrice=&minPrice=5")` returns 200 with `appliedFilters.minPrice` null.
   - The bullet keeps its "other than `status`" scope, so it does not contradict the status rule.
4. **Nit 4 (HTML comments marking the §5.4/§5.5 examples as asserted): fixed. The comments do not change which fence is read or what is compared.**
   - Each comment sits on its own line, between the `### 5.4` or `### 5.5` heading and the ```` ```json ```` fence.
   - `specExample` (search_test.go:312-333) works in three steps:
     - It cuts at `"\n" + heading`.
     - Then it cuts at the first ```` ```json\n ````.
     - Then it cuts at the next ```` ``` ````.
   - Neither comment contains a backtick, a `### ` heading or the string `json` followed by a fence. The same block is still selected, and nothing outside the fence reaches the compared bytes.
   - The comment text describes the tests correctly:
     - §5.4 is a `reflect.DeepEqual` on `map[string]any` against the `targetBudget=450000` response cut to its first item ("valid JSON and equal to ... first item only").
     - §5.5 is `bytes.Equal` against the `json.Compact` of the fence plus a trailing newline ("byte-for-byte after json.Compact, so keep key order and values in sync").
   - Both `SPEC section 5.4 example` and `SPEC section 5.5 example` pass in the uncached run.
   - `search_test.go` is the only code in the repo that reads docs/SPEC.md (`specPath`). No other parser depends on those lines.

Nit 5 was routed out of scope in cycle 1. It is tracked as task_01M4W7W5H5WJQ9EG3DZ3XJ6S2D ("Keyword match is case-insensitive only for ASCII ..."), which is `related_to` BC-5. Nothing is owed here.

## Merge check

- 1871b01 has parents fd15b4a (BC-5) and 0e6730f (master). It is a clean automatic merge:
  - `git merge-tree --write-tree fd15b4a 0e6730f` gives tree `e4eb385...`.
  - That is identical to `1871b01^{tree}`, so there is no hand resolution and no evil-merge content.
- What came in from master (`git log e5f49bb..0e6730f`): the BC-11 commits 34271a1, 8c0e891 and 0e70187, merge 0abb105, and the board sync 0e6730f.
  - Master changed nothing under `docs/` or `backend/` in that range (`git diff e5f49bb 0e6730f -- docs backend` is empty).
- `git diff 75530c0 1871b01 --stat` lists 49 files:
  - BC-5's own change is `docs/SPEC.md` (the fd15b4a hunks).
  - The other 48 are master's frontend StatusLine, Pager and LoadingLine files (9) and `.lattice/` files (39).
- `git diff fd15b4a HEAD -- docs backend` is empty. The merge did not touch BC-5's files.
- `git diff master HEAD -- frontend .lattice` and the root files are empty. Outside its own files, the branch now matches master exactly.
- `git diff master...HEAD --stat` is exactly BC-5's 10 files, +2069 -9:
  - `backend/cmd/api/main.go`
  - `backend/internal/httpapi/` cities.go, cities_test.go, query.go, query_test.go, respond.go, router.go, search.go, search_test.go
  - `docs/SPEC.md`
  - It has no `.lattice/` path and no frontend path. The merge-base is 0e6730f.

## Findings

### Critical
None.

### Major
None.

### Minor
None.

### Nit
None.

Regression check: `git diff 75530c0 HEAD -- backend/` is empty, so every backend file is byte-identical to the version cycle 1 passed. The uncached test counts match cycle 1 exactly (65/367). Nothing from cycle 1's PASS can have regressed.

## Acceptance (ticket's three bullets, still met?)

1. **Response bodies match §5.4 and §5.5 field for field: still met.**
   - The `strictRoundTrip` tests (DisallowUnknownFields, then re-marshal plus `reflect.DeepEqual`) are unchanged and pass.
   - The two SPEC-example tests still read the right fences after the nit 4 comments, and they pass.
2. **`make check` green: met at 1871b01.** Exit 0. Lint and tests pass on both halves, including the BC-11 frontend code merged in.
3. **Pipeline reads parse, search, [dedupe slot], rank, paginate, respond: still met.**
   - search.go is unchanged (L22-54).
   - It still has the six commented stages in order: parse, `Searcher.Search`, the §9.1 dedupe-slot comment, `now` with `domain.Rank`, `domain.Paginate`, and `writeJSON(newSearchResponse(...))`.

Validation against the running API (the curl smoke queries in the ticket) is the validation stage's job and is not reviewed here.