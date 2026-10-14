Lattice-Reviewed-Commit: 43e88241c99be1cef02d5cf44659fc3f17abdff3
Lattice-Reviewed-Worktree: /Users/forrest/Code/brillio-worktrees/bc-8-frontend-core
Lattice-Reviewed-Base: master (c91fa7f79dfce9029dcc63f0db6b9dc4768fa6b4)
Lattice-Reviewed-Head: feat/BC-8-frontend-core (43e88241c99be1cef02d5cf44659fc3f17abdff3)

# Code Review: BC-8 — Frontend core (API client and types, URL state and search hooks, BRANDING formatters)

Reviewer: agent:fable-5.1-reviewer (cold read, fresh context)
Reviewed: `feat/BC-8-frontend-core` @ 43e8824 against `master` @ c91fa7f
Worktree: `/Users/forrest/Code/brillio-worktrees/bc-8-frontend-core`

## 1. Verdict

**PASS**

The implementation matches the plan (including every Plan-Review Cycle 1 resolution) and the ticket's acceptance criteria. All findings below are minor; none requires an implementation rework cycle. One process item (the missing completion comment) should be closed by the orchestrator before `in_validation`.

## 2. Summary

I reviewed 19 files (2,262 insertions): the §7.3 types, the client and error parser, three hooks, three formatter modules, shared fixtures, their co-located tests, and the three SPEC edits. The code is tight, well-commented at the decision level, and consistently defensive at the API boundary without over-engineering. I ran every gate the plan names and independently verified the fixture rows and relevance scores against the real sample data and the Go domain ranker; everything checks out. The key finding is a process gap, not a code defect: the plan required the completion comment to republish the exports block and the usage notes for BC-9/10/11, and that comment was never posted.

### Gate results (run by the reviewer in the worktree)

| Gate | Result |
|---|---|
| `eslint . --max-warnings 0` | No issues |
| `tsc -b --noEmit` | Clean |
| `prettier --check .` | All files formatted |
| `vitest run` | 10 files, 197 tests, all passed |
| `vite build` | Built (20 modules) |
| `make lint-api test-api` (api half of `make check`) | 0 issues; all packages pass with `-race` |
| New runtime dependencies | None (`package.json` and lockfile untouched) |
| `.lattice/` staged in BC-8 commits | None |
| Commit trailers | Present on all three commits |

### Independent fixture verification

- `listingA1`, `listingA3`, `listingA7` are byte-for-byte the A1/A3/A7 rows of `data/sample_listings.json`.
- I ran `domain.Rank` + `domain.Paginate` from the merged BC-3 code over the full 12-row sample with `targetBudget=450000`, `now=2026-10-13`: A1 → rank 1, 0.7414 / 1.0000 / 0.3536; A3 → rank 4, 0.6836 / 0.8867 / 0.3789; A7 → rank 10, 0.4275 / 0.6000 / 0.1688; `totalItems 12`, `totalPages 2`. All match `fixtures.ts` exactly. Arlington has zero listings, so `emptySearchResponse` is also a real response shape.
- The live `/api/listings/search` endpoint is not wired on `master` yet (404), so the HTTP envelope could not be checked end-to-end; the domain-level check above is the strongest available evidence.

### Acceptance criteria

| Criterion | Status |
|---|---|
| All tests green; `make lint` green | Met (both halves) |
| No new runtime dependencies | Met |
| Types compile against fixtures from real sample data (3 listings + full SearchResponse) | Met, and the data is verified real |
| `types.ts` exactly as §7.3 | Met (verbatim, plus the additive `LISTING_STATUSES`) |
| Client: undefined omitted, status repeated, numbers plain, non-2xx → ApiError, TypeError/AbortError → network | Met and tested |
| `useSearchParamsState` per §7.1 incl. popstate, page reset, pageSize preserved | Met and tested |
| `useListingSearch` per §7.2 incl. abort, previous, empty/invalid | Met and tested, incl. R2's held-result test |
| `useCities` load once, `[]` on failure, console.error | Met and tested |
| Every BRANDING §7.1 row | Met; every row has a test, plus §5.3–5.7 strings |
| Tests per §8 (400/500/TypeError/abort; hooks under jsdom with history/popstate; every formatter row) | Met |
| SPEC edits (R2, R3) with deviation comment | Met; comment recorded 2026-10-14T02:47Z |

## 3. Issues

**[minor] Lattice process (no file) — Completion comment with exports block and child usage notes was not posted**
The plan's "Planned exports" section says "republish this block in the completion comment", and resolutions R7(b) and R8 route the hand-off guidance for BC-9/10/11 through that same comment: items are trimmed in `searchResponse` (never compare `items.length` to `page.*`), new helpers go in new files under `src/format/`, fixture additions go in a labelled `// --- BC-N additions ---` block, components call date/meta formatters bare and pin the clock with `vi.useFakeTimers({ toFake: ['Date'], now: ... })`, BC-11 owns the state-to-halves mapping, SearchPage wires `retry` and must not log. Only the SPEC-deviation comment exists on BC-8. The in-code comments in `fixtures.ts` cover part of this, but the three child tickets read the Lattice record, not the file header.
**Fix:** Orchestrator posts one `lattice comment BC-8` containing the exports block and the usage notes verbatim from the plan. No code change; do this before `in_validation`.

**[minor] frontend/src/hooks/useListingSearch.ts:37 — A held *error* outcome is also shown without a loading state on return**
`deriveState` treats any settled outcome for the current key as "held". Sequence: A settles with `http`/`network` error → params change to B (B in flight) → user navigates back to A. The hook immediately reports `error` for A (ErrorState, "Try again", Search enabled) while silently refetching A; if the refetch succeeds the UI flips to the list. SPEC §7.2's new clause says "a query whose result the hook already holds", which a reader will take to mean data. Distinguishing this case needs in-flight tracking, which the lint-driven design deliberately avoids (decision 5, R2), so this is a known cost rather than a bug. Clicking Try again during the transient still works (attempt++ aborts the refetch and restarts).
**Fix:** Either accept with a one-line justification comment on the task, or narrow the SWR clause by only treating `kind: 'data'` outcomes as held: when `settled.key === requestKey` and `outcome.kind === 'error'`, the effect that re-runs for that key cannot bump state, so the honest option is to document it in the hook's header comment ("a held error is shown while its refetch runs"). Recommend: accept and document.

**[minor] frontend/src/hooks/useSearchParamsState.ts:121 — Duplicate-write check compares the canonical query to the raw `location.search`**
`navigate` skips the push when `query === search.replace(/^\?/, '')`. A non-canonical URL that arrived by hand or from an external link (`?status=sold&status=active`, `?city=Spring%20field`, `?page=1`) fails the comparison on a same-params submit, so one extra history entry is pushed that rewrites the URL to canonical form. Harmless and arguably useful (the URL self-heals), but it contradicts the comment "A write equal to the current query adds no history entry".
**Fix (optional, one line):** compare against `buildSearchQuery(currentParams())` instead of the raw search, or amend the comment to say the comparison is textual and that a non-canonical URL is rewritten once.

**[minor] frontend/src/api/client.test.ts:223 — Test name overstates what it exercises**
"treats a body read cut short by an abort as an abort" resolves `fetch` with a syntactically invalid JSON body and a signal that is already aborted. It does not simulate a read interrupted by the abort; it exercises the same `catch` branch via a `SyntaxError` plus `signal.aborted === true`. Behaviour verified is correct and the branch is the one that matters.
**Fix:** Rename to something like "reports a body parse failure as an abort when the signal is aborted", or add a one-line comment.

**[minor] frontend/src/api/client.ts:59 and frontend/src/api/errors.ts:20 — `isRecord` is defined twice**
Identical helper in two sibling modules.
**Fix:** Export it from `errors.ts` (already imported by `client.ts`) or leave as is; cosmetic.

**[minor] frontend/src/format/date.ts:33 — Day check is not month-aware**
`2026-02-31` formats as `Listed Feb 31` rather than falling back to raw. The API only emits valid dates (parsed with Go's `DateLayout`), so this is unreachable with real data.
**Fix:** None required; note for completeness.

## 4. Positive Observations

- **One serializer, used three ways.** `buildSearchQuery` producing the API URL, the `pushState` URL and the fetch-equality key is the cleanest decision in the ticket. The round-trip property test (`parseSearchParams(buildSearchQuery(p))` deep-equals `p`) locks it in, and the enum-ordered, deduplicated `status` makes `['sold','active']` and `['active','sold']` the same search without any caller discipline.
- **Derived loading state is correct and survives StrictMode.** `{ settled, lastGood }` with a `${attempt}|${query}` key is a small, honest state model: no refs read during render, no set-state-in-effect, every setState inside a promise callback guarded by `controller.signal.aborted`. The StrictMode, unmount, mid-flight-abort, retry and held-result tests prove each transition rather than asserting on mock internals.
- **Error channel respects the typed lint without bending the SPEC.** `ApiRequestError` carries the unchanged `ApiError` union plus `requestId` and `cause`, declared as class fields (erasableSyntaxOnly-safe). `toApiError` keeps the hooks free of classification logic.
- **Boundary hardening is proportionate.** Shape guards on 2xx bodies (`items` array, `page`/`scoring` objects; cities filtered to strings), the HTML-502 fallback message, empty-body handling, and `signal ?? null` for exactOptionalPropertyTypes are each one or two lines and each has a test.
- **URL parsing policy is stated and tested as policy.** The "drop what no field can display, keep range errors for the API to judge" rule is written in one comment block, and the test tables mirror it row for row (`minPrice=-5` kept, `minBedrooms=2.5` dropped, `status=active,bogus` → `['active']`, `minprice=5` ignored).
- **Date formatting is time-zone proof by construction.** Parsing `YYYY-MM-DD` by regex groups and never constructing a `Date` sidesteps the UTC-midnight shift; the suite pins `TZ=America/Los_Angeles` and first asserts that the naive approach *would* shift, which makes the test self-validating.
- **Fixtures are real and provably so.** Rows are verbatim sample data; ranks and relevance match the merged BC-3 ranker to four decimals; `emptySearchResponse` corresponds to a city with zero listings. The "items trimmed" comment and `Deferred<T>` helper are both exactly what the child tickets need.
- **SPEC edits are surgical.** Three clarifications, each named in the deviation comment, none changing a type or endpoint. §7.1's `clear()` contradiction is resolved in the text rather than left in a plan file.
- **Commit hygiene.** Conventional Commit subjects, a body that explains the design choices, attribution trailers on all three commits, `.lattice/` untouched, no dependency drift.
