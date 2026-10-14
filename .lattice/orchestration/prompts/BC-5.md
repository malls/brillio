# Ticket card — BC-5

```
TICKET=BC-5
TASK_UUID=task_01M4VXK2E02WC334GKAWFQ5XDG
TITLE=Search API: strict query parsing, validation with field codes, pipeline handler, cities endpoint
WORKTREE=/Users/forrest/Code/brillio-worktrees/bc-5-search-api
BRANCH=feat/BC-5-search-api
REVIEW_BASE=master   (see note 1 — review only after BC-3 and BC-4 have both landed)
ANCHOR=master @83952b0 + feat/BC-3-domain-ranking @ee09340 (in validation) + feat/BC-4-sqlite-store @318ebc5 (in review), merged at creation
PORTS=API 18005 (PORT=18005, DB_PATH=:memory: or /tmp/brillio-bc5.db)
AREA=area:11
```

Read `/Users/forrest/Code/brillio/.lattice/orchestration/prompts/delegator-brief.md` now and follow it.

## BC-5 notes (binding)

1. **Two press-ahead anchors.** Your branch was created from master with both in-flight dependency branches merged
   in, so you can plan and implement against the real `domain.Rank`/`Paginate` (BC-3) and `store.Search`/`Cities`
   (BC-4) APIs. When either anchor gains commits before it lands, `git merge feat/BC-3-domain-ranking` /
   `git merge feat/BC-4-sqlite-store`. When each lands, `git merge master`. Never rebase. **Do not enter `review`
   until both BC-3 and BC-4 have landed on master** (check `git log --oneline master | grep -E 'BC-3|BC-4'`); if you
   reach that point first, mail the Orchestrator `HOLD BC-5 waiting for <ticket>` and wait. Then merge master,
   re-run the gate, and review against `master`.
2. **SPEC §5.4 example fix (run-state decision 2):** the example `relevance.score` 0.8731 is inconsistent with its own
   fields; at asOf 2026-10-13, A1 with targetBudget 450000 scores 0.6×1.0 + 0.4×0.3536 = **0.7414**. Correct the
   example in `docs/SPEC.md` in your branch and say so in a `lattice comment` (SPEC's own deviation rule).
3. **Shared domain rule (run-state shared-files row, BC-2 note):** `domain.Listing`/`Scored` are never JSON-encoded on
   the wire. The handler maps them to response DTOs; `listedDate` is formatted with `domain.DateLayout`. Rounding to
   four decimals happens here, not in domain.
4. **Router:** register `GET /api/listings/search` and `GET /api/cities` in `NewRouter` (BC-2), one line each, with the
   dependencies on `httpapi.Deps` (`Searcher`, `Now func() time.Time`, cities source). BC-6 adds the OpenAPI route after
   you; keep the param allowlist exported or otherwise readable by a test (BC-6's sync test reads it).
5. Field-error messages use BRANDING §7.2 strings verbatim where one exists; otherwise a plain sentence in the same
   voice. The response-shape test unmarshals with `DisallowUnknownFields` both ways (ticket acceptance).
6. **Gate:** full `make check` (both scaffolds are on master).
7. **Validation** per the ticket against the running API on port 18005: the SPEC §10 smoke queries plus
   `status=pending`, `keyword=%25`, `pageSize=0`, `minprice=1`; also confirm validation-plan rows 13–20 and 28/31 by
   curl (cities list, each field-code case, empty value, page past end, empty city, status repeat/comma, ranked
   Springfield response with A1 at 0.7414 for asOf 2026-10-13 — or the value your fixed clock gives, no-budget
   ordering, `pageSize=3&page=2` ranks 4–6). Kill the server afterwards.
