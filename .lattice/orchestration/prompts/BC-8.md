# Ticket card — BC-8

```
TICKET=BC-8
TASK_UUID=task_01M4VXK2TV60YQ37FY66ZW2J2F
TITLE=Frontend core: API client and types, URL state and search hooks, BRANDING formatters
WORKTREE=/Users/forrest/Code/brillio-worktrees/bc-8-frontend-core
BRANCH=feat/BC-8-frontend-core
REVIEW_BASE=feat/BC-7-frontend-scaffold
ANCHOR=feat/BC-7-frontend-scaffold (BC-7 at review ~02:03Z, head 1d6008a; not yet landed)
PORTS=none (hooks and formatters are tested under jsdom)
AREA=area:11
```

Read `/Users/forrest/Code/brillio/.lattice/orchestration/prompts/delegator-brief.md` now and follow it.

## BC-8 notes (binding)

1. **Press-ahead anchor.** If `ANCHOR` is `feat/BC-7-frontend-scaffold`, your branch was cut from BC-7's in-review
   branch. Review your own diff against `REVIEW_BASE`. When BC-7 lands, `git merge master` (never rebase) and
   switch `REVIEW_BASE` to `master`. If BC-7 gains commits before it lands, `git merge feat/BC-7-frontend-scaffold`.
2. **You are the owner of the shared frontend helpers** that BC-9, BC-10 and BC-11 will build on in parallel:
   `src/api/*`, `src/hooks/*`, `src/format/*` (every BRANDING §7.1 row, plus anything §5.3/§5.6/§5.7 strings need:
   `12 results` / `1 result`, `ranked for $450,000`, `newest first`, `Page 1 of 2`, `No. 1`, two-decimal score,
   monogram initials) and `src/test/fixtures.ts` (three real listings — include A1, A7 (pending) and a Vienna or
   Fairfax listing — plus one full `SearchResponse`, typed). Make them complete enough that the component tickets
   never need to fork a parallel helper. Export them from obvious module paths and list the exports in your
   completion comment.
3. TypeScript is pinned at 6.0.3 with `exactOptionalPropertyTypes` and `noUncheckedIndexedAccess` (BC-7). No new
   runtime dependencies; no msw (stub `fetch` with `vi.fn`).
4. The current-year rule for `Listed Aug 29` vs `Listed Aug 29, 2025` takes the year as an injected parameter so
   tests are deterministic.
5. Gate: `make lint-web test-web` plus `npm run build` (full `make check` if BC-2 has landed on master by hand-off).
6. Validation: N/A until BC-12. The validator cites the vitest output listing the client, hook and formatter cases.
