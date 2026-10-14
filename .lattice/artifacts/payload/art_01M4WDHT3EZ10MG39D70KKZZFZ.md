Verdict: PASS-WITH-NITS
Reviewed HEAD: 0b4e2300181a524fdf96c7af7faf2422451c0955
Reviewer: fable sub-agent (decision 04:02Z)

## Findings
### Critical
none
### Major
none
### Minor
none
### Nit
- README.md:159 "`(source, id)` is unique, so the order is total and pages never shift between requests." The total-order claim is right (backend/internal/domain/rank.go:69-83), but "never shift between requests" holds only for requests scored on the same UTC date: recency decays with AgeDays(listedDate, now) (backend/internal/domain/scoring.go:61-77, :95), so listings with different fit/recency mixes can swap across a date boundary. README.md:150 already says scores drift daily; a hedge such as "pages never shift between requests on the same day" (or "for the same asOf") closes the gap. Optional.

## Checks run
- HEAD 0b4e230; clean before and after.
- N1 (listings.go:35,137-142), N2 (search.go:120,186-204; query.go:322-343; live echo confirmed), N3 (jsdom 30.1.2 engines; host node v24.21.0) all correct.
- Acceptance: quote diff empty (14 lines); '!' count 0; emoji scan empty; wc -l 245; voice grep empty.
- Live API on 18015 (:memory:, no Docker; killed, port free): healthz, cities, Springfield ranks A1/B7/A2/B8 with A1 0.7382 on 2026-10-14, §5.5 example fields/messages, %zz malformed_query, page past end 200 [], city= absent, Nowhere 200 totalItems 0, 404/405 envelopes, openapi.yaml served.
- Spot-checks vs query.go, respond.go, scoring.go, paginate.go, rank.go, search.go, router.go, migrate.go, main.go; versions vs go.mod/package.json/lockfile/Dockerfile.dev/compose.yaml/host; Makefile/compose/smoke behaviour; config defaults; layout tree (only docs/EXTENSIONS.md forward link absent); sample data duplicates; test-coverage claims with file:line.

## Summary
All three cycle-1 nits fixed correctly with no new inaccuracy. Full re-read against code, contract files and a live API found no factual error; all binding card notes met. One optional precision hedge on line 159; does not block in_validation.