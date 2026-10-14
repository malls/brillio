# Plan Review: BC-8 — Frontend core (API client and types, URL state and search hooks, BRANDING formatters)

Reviewer: agent:fable-5.1-plan-reviewer, 2026-10-13 (host clock; board dates read 2026-10-14Z).
Reviewed against: the BC-8 task card, `docs/SPEC.md` §5.3–5.5, §7.1–7.5, §8, `BRANDING.md` §5.3–5.7, §5.9, §7.1–7.2, the BC-7 toolchain actually on `master` (`frontend/package.json`, `tsconfig.app.json`, `eslint.config.js`, `vite.config.ts`, `src/test/setup.ts`), `data/sample_listings.json`, and `.lattice/orchestration/run-state.md`. The "card notes 1–6" the plan calls binding are not in the Lattice task, its events, or its notes file (they presumably live in the delegator's boot prompt), so they were not independently checked; notes 5 (gate) and 6 (validation N/A) as restated in the plan agree with the ticket and run-state.

## 1. Verdict

**PASS**

Implementation can proceed. Every issue below is minor; none changes the approach. Several are one-line amendments to the plan file or to the single SPEC edit the plan already makes.

## 2. Summary

The plan covers every bullet of the ticket's scope and all three acceptance criteria, maps each test to SPEC §8 and to every BRANDING §7.1 row, and names every file it creates. Its design decisions are driven by verified toolchain constraints rather than guesses: the derived-loading hook avoids react-hooks 7's `set-state-in-effect`, `ApiRequestError` satisfies `only-throw-error`, property-function members avoid `unbound-method`, `signal ?? null` satisfies `exactOptionalPropertyTypes`, and dates are parsed by parts because `new Date('YYYY-MM-DD')` is UTC midnight (I confirmed on this host: it prints Aug 28 in US Eastern). I recomputed the full 12-listing ranking from SPEC §6.2–6.4 at `targetBudget=450000`, `asOf 2026-10-13`: A1 is rank 1 (0.7414 / 1 / 0.3536), A3 rank 4 (0.6836 / 0.8867 / 0.3789), A7 rank 10 (0.4275 / 0.6 / 0.1688), exactly as the fixture table states. The key concern is small: the gate section is already stale (BC-2 and BC-7 have both landed and the worktree sits on `master` at 83952b0), and the one-sentence SPEC edit should absorb two of the plan's resolved ambiguities so they live in the contract rather than only in the plan file.

## 3. Issues

**[MINOR] Gate — precondition already resolved; anchor line stale**
The plan says the branch is anchored on `feat/BC-7-frontend-scaffold @1d6008a` and runs full `make check` "only if BC-2 has landed on `master` by hand-off". Both conditions have moved on: BC-2 landed at 02:09Z, BC-7 at 02:27Z, the run-state records "Half-gate exception ended; full `make check` now required for every merge", and `git worktree list` shows `bc-8-frontend-core` at `83952b0` (master). The ticket's acceptance is `make lint` green, which includes `lint-api`.
**Recommendation:** Replace the conditional with: run `make check` from the worktree root (both halves), then `cd frontend && npm run build`. Update the anchor line to `master @83952b0`. No other change.

**[MINOR] Decisions 5 and 8, Ambiguity 8 — silent refetch on revisiting a settled key deviates from SPEC §7.2 wording**
SPEC §7.2 says "on every params change, abort the previous in-flight request, set `loading`, call the client". With status derived from `settled.key !== requestKey`, navigating A → B → Back to A while B is in flight shows A's old result as `success` and refetches silently: StatusLine does not read `Loading`, LoadingLine is off, Search is enabled and `aria-busy` is false while a request is actually in flight. No wrong data can be shown (B's handler is aborted; A's refetch replaces in place), so this is safe, and it is a reasonable consequence of the lint-driven design. But it is observable behaviour that BC-12's e2e tests and the validator will meet, and today it is recorded only in the plan file.
**Recommendation:** Keep the design. Add a clause to the planned §7.2 sentence: "returning to a query whose result the hook already holds shows that result immediately and refetches without a loading state". If strict loading is preferred instead, that is a plan change, not an implementation detail; decide before `in_progress`.

**[MINOR] SPEC edits — the §7.2 edit should also fix the `success` comment and carry Ambiguity 1**
The plan's Ambiguity 2 (a page past the end is `success` with `items: []`) contradicts the §7.2 type comment `// data.items.length > 0`, which the one-sentence edit leaves in place. Ambiguity 1 (`clear()` keeps `pageSize`) resolves a genuine conflict inside §7.1 ("Clear writes an empty query string" vs "preserved if present") and the ticket text itself says "`clear()` pushes an empty query"; either reading is defensible, but the resolution belongs in the contract, not only in the plan, because BC-12's tests will assert one of them.
**Recommendation:** In the same in-branch SPEC edit, change the `success` comment to `// totalItems > 0; items may be empty on a page past the end` and add to §7.1 one clause: "Clear writes an empty query string except for `pageSize`, which is preserved like everywhere else." Cite both in the deviation comment.

**[MINOR] Decisions 2 and 10 — string normalization and mixed `status` values are unspecified**
The serializer omits `''` and the parser drops empty values, but neither says whether `city` and `keyword` are trimmed. `?city=%20` would parse as a non-empty filter, serialize back as `city=+`, and give an equality key the server treats as "absent" after its own trim (§5.3). Likewise "unknown `status` values are dropped" does not say whether `status=active,bogus` drops `bogus` or the whole param. Also, dropping unknown params on the UI side is the right call (no field could show the server's `unknown_param` error, and forwarding them would produce an invisible `invalid` state), but it is the mirror image of §5.3's "a typo must not silently widen the result set" rationale and should be stated so nobody later "fixes" it.
**Recommendation:** Add to decision 2: trim `city` and `keyword` in both directions, omit when empty after trim. Add to decision 10: a bad `status` value is dropped and the valid ones kept; unknown params are dropped deliberately because no field can display their error. Add one test each (`city=%20` parses to no city; `status=active,bogus` parses to `['active']`).

**[MINOR] Module contents — serializer and parser live in different layers**
`buildSearchQuery` is in `api/client.ts` and `parseSearchParams` in `hooks/useSearchParamsState.ts`, so the URL-state hook imports the HTTP client module, and the two halves of the round-trip property (`parseSearchParams(buildSearchQuery(p))` deep-equals `p`) sit in different files with different test suites. Low stakes, and SPEC §3 lists only `client.ts, types.ts, errors.ts` under `src/api/`.
**Recommendation:** Optional. Either put both in one module (`src/api/query.ts`, imported by `client.ts` and the hook, with the round-trip test beside them) and note the layout addition in the SPEC edit, or keep the plan as written and accept the import direction. Do not spend a rework cycle on this.

**[MINOR] Risks — `erasableSyntaxOnly` and the only class in the ticket**
The toolchain line lists `erasableSyntaxOnly`, but Risks does not mention its one practical trap here: `ApiRequestError` must not use constructor parameter properties (`constructor(readonly apiError: ApiError)`), which TypeScript rejects under that flag. Fields must be declared on the class and assigned in the constructor body, and `this.name` set so logs read `ApiRequestError`.
**Recommendation:** Add one line to Risks. `super(message, { cause })` is fine on the ES2023 lib.

**[MINOR] Tests — two mechanics to pin down**
(a) `restoreMocks: true` and `unstubGlobals: true` do not undo `vi.useFakeTimers`; the date suite's "default year from a faked clock" test needs `vi.useRealTimers()` in `afterEach` or the fake `Date` leaks into later tests in the file. (b) `searchResponse` carries three items under `page { pageSize: 10, totalItems: 12 }`; the plan rightly says to comment this, and it keeps BRANDING's own `12 results` / `Page 1 of 2` strings, but no test (in BC-8 or in the hand-off notes for BC-10/11) may assert `items.length` against `page.*`.
**Recommendation:** Add both to the Mechanics paragraph and to the completion comment's usage notes.

**[MINOR] Shared-file handling — appending to the same files from three parallel branches**
BC-9, BC-10 and BC-11 "may only add exports to `src/format/*` and `src/test/fixtures.ts`". Three branches each appending to the end of `format/money.ts` or `fixtures.ts` will conflict at landing even though the intent is a union; git cannot auto-merge adjacent appends.
**Recommendation:** In the completion comment, ask the three children to put new helpers in new files under `src/format/` where possible (for example BC-9's `$400,000` input parsing in `format/input.ts`) and to add fixtures inside a ticket-labelled block at the end of `fixtures.ts`. The orchestrator's landing order then merges cleanly.

## 4. Positive Observations

- **Fixtures are real and correct.** The three listing rows, their ranks and the four-decimal relevance components match an independent recomputation of SPEC §6.2–6.4 over all 12 sample rows. The fixture yields BRANDING's own `12 results · ranked for $450,000` and `Page 1 of 2`, so component tests downstream can assert the canonical strings on real data (BRANDING §8: no placeholder data).
- **Lint constraints were probed, not assumed.** Derived loading state (react-hooks 7 `set-state-in-effect`), `ApiRequestError` (`only-throw-error` / `prefer-promise-reject-errors`), property-function hook members (`unbound-method`), `const body: unknown` plus guards (`no-unsafe-*`), `.then(ok, fail)` with no async effects (`no-floating-promises`, `require-await`), string effect deps (`exhaustive-deps` at `--max-warnings 0`), and `signal ?? null` (`exactOptionalPropertyTypes`). Each is a real failure mode in this exact config and each has a stated fix.
- **One serializer, three uses.** `buildSearchQuery` produces the API URL, the `pushState` URL and the equality key, with canonical §5.3 order and deduped, enum-ordered `status`, so an equal-but-new params object never refetches and shared URLs are stable. The `String(NaN)` / Go `ParseFloat` observation is a good catch.
- **Timezone handling is right.** Dates parsed by parts with a literal month table, no `Date` constructed from `YYYY-MM-DD`, a pinned `TZ` in tests (Node's reassignment behaviour verified here as well), and an unparseable date left visible as `Listed <raw>`.
- **Abort semantics are explicit and tested.** Name-checked `AbortError` (jsdom and Node own separate `DOMException`s), `signal.aborted` checks in every settle handler (covers the mock's late resolve and a real abort mid-body-read), and StrictMode double-mount covered for both hooks.
- **Scope discipline.** A precise out-of-scope list (components, pages, CSS, config, lockfile, FilterBar input parsing, StatusLine mapping), no new dependencies, no msw, and the only SPEC change is a clarification routed through SPEC's own deviation rule.
- **Hand-off is complete.** The planned-exports block, usage notes for BC-9..12, the retry path and the "callers do not log twice" rule remove the most likely sources of duplicated or contradictory work in the three parallel component tickets.
- **Ambiguities are listed, numbered and cheap to flip**, which is exactly what a reviewer needs; the two that matter (1 and 8) are addressed above by moving their resolution into SPEC.
