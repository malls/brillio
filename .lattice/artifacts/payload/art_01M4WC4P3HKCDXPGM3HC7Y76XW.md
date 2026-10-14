Verdict: FAIL
Reviewed plan: task_01M4VXK3QPSMGSSGYBJ82HRQ4Z.md (167 lines)
Reviewer: fable sub-agent (decision 04:02Z)

## Findings
### Critical
- none

### Major
- [M1] "Facts the README asserts" > Quote, and "Acceptance checks" > A2 — The plan says the scoring doc comment is "scoring.go lines 1-13" and the A2 check diffs `sed -n 1,13p` against the quoted block. The comment is 14 lines: `backend/internal/domain/scoring.go:14` is `// listed date counts as age 0.`, line 15 is blank, line 16 is `package domain`. As written, the README would quote a comment that ends mid-sentence and the acceptance check would bless the truncation, which breaks card note 1. — Recommendation: change both to lines 1-14, and make A2 independent of the count: `awk 'NF==0{exit} {print}' backend/internal/domain/scoring.go | diff - <quoted block>`. Amendment-level.
- [M2] "Validation" step 4 — The plan runs `make up`, `env -u DOCKER_DEFAULT_PLATFORM docker compose up -d --wait`, `make down` (twice) and `make smoke` on 8080/5173. The run-state's approval scope is "Docker Compose on 8080/5173 for BC-13/BC-17 only", and card note 4 scopes BC-15's command runs to the native API on PORT=18015 / DB_PATH=:memory:. `make down` is `docker compose down -v` (Makefile:47-48) against the fixed project name `brillio` (compose.yaml:25), so a race with a BC-17 start would delete its containers and cache volumes. — Recommendation: drop step 4; for the Docker blocks cite BC-13's landed evidence (make smoke 8/8, art_01M4WAG45G8GAP1ATBBVZHAVYC) and BC-17's fresh-directory re-check; a live run needs explicit orchestrator approval recorded first. Amendment-level.

### Minor
- [m1] "README outline" — no architecture beyond the §3 tree and the SQL-vs-Go trade-off. Recommendation: 6-8 lines on the request path (parse/validate -> store.Search -> dedupe slot (not built) -> domain.Rank -> domain.Paginate -> DTOs with 4-decimal rounding), mirroring backend/internal/httpapi/search.go:22-54, plus one line for the frontend.
- [m2] Frontend test-file count (23) may drift with BC-14 in flight. Recommendation: no test or test-file counts in the README; describe coverage by area.
- [m3] `make down` absent from Run it. Recommendation: one line for `make down` (removes containers, network and volumes) and a note that the first `make up` is slow (start_period 180 s).

### Nit
- [n1] Line 3 anchor 261760f; HEAD now c4d50a9 (no non-.lattice diff). Update.
- [n2] Spell out the SPEC §5.5 example request: `minPrice=5&maxPrice=1&pageSize=0&minprice=1` (search_test.go:725).
- [n3] BRANDING §7 also bans "please", apologies and cheer; add to the Voice line.
- [n4] Validation step 3 `npm run dev` needs frontend/node_modules, which `make check` installs first; state the ordering.
- [n5] Project layout tree must omit the handout PDFs and `.lattice/` (SPEC §12 excludes them from the zip).

## Summary
Fact dossier thorough and almost entirely correct (constants, formulas, tie-break order, error codes incl. malformed_query, params, Makefile recipes, compose health gating, versions). Two Majors block implementing as written: the quote range is off by one line, and validation step 4 runs Docker Compose outside the run-state's approved scope with a down -v on a shared project name. Both are amendments, not plan-level rework.