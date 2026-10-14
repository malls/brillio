# BC-15: README: run instructions, architecture, scoring rationale and trade-offs

Write the README the handout asks for: run instructions, architecture, the scoring approach and its trade-offs. Spec: docs/SPEC.md §1, §2, §6, §9, §12; handout "Deliverables". Voice: BRANDING.md §7 (plain product, sentence case, no emoji).

Why: the handout grades "a short README explaining your scoring approach and any trade-offs". The interviewer reads this before the code. Every claim must be true of the code as merged.

Scope (README.md, under roughly 250 lines):
- What it is: two sentences.
- Run it: `docker compose up` then http://localhost:5173; native alternative with `make dev-api` and `make dev-web`; `make check`; `make smoke`; required tools and versions.
- Project layout: the §3 tree, trimmed to what exists.
- API: the three endpoints with one curl example each and a pointer to /api/openapi.yaml; the error envelope in five lines.
- Scoring: the formula in words and symbols, the asymmetric budget fit and why (over budget hurts twice as much), the 30-day half-life and why, the 0.6/0.4 weights and why, tie-breaking, the worked example from §6.2, what happens without a budget (newest first), and where the constants live. Quote the doc comment from BC-3.
- Handling invalid input: strict unknown params, all errors collected, page past the end is a 200, an empty city is a 200, and why each.
- Trade-offs: SQL filters with Go ranking (and the scaling path), SQLite over a server database, keeping cross-source duplicates visible, stdlib only, no dark mode by decision.
- Extending it: three sentences and a link to docs/EXTENSIONS.md.
- Testing: what is covered in each half and how to run it.

Acceptance:
- Every command in the README runs as written (BC-17 proves it from the zip).
- Every number and formula matches internal/domain.
- No emoji; no exclamation marks.

Validation: N/A (documentation). Record a one-line justification with --role validation; BC-17 performs the fresh-directory check.
