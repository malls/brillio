# Plan Review: BC-5 (Search API)

Reviewer: agent:fable-5.1-plan-reviewer, 2026-10-14. Reviewed against the `feat/BC-5-search-api` worktree
(`5e0800d`, BC-3 and BC-4 merged in), docs/SPEC.md, BRANDING.md, `.lattice/orchestration/run-state.md`,
the three upstream comments on BC-5, and the BC-6/BC-9 task descriptions.

## 1. Verdict

**PASS**

Implementation can proceed. One finding (the `malformed_query` contract addition) is a genuine product
decision that the orchestrator must triage before `in_progress`; the plan already carries both outcomes,
so it does not block the verdict. Four minor findings should be folded into the plan file as obvious fixes.

## 2. Summary

The plan covers every item in the task description (parser, pipeline handler, cities endpoint, the full
§8 httpapi test list, the ticket's extra validation queries) and every acceptance criterion, with file-level
scope, exact signatures, exact error strings, and expected numbers. I verified its claims rather than
trusting them: every "real API" it cites exists on the branch with the stated signature; every expected
relevance value, rank order and page-metadata tuple was recomputed from the seed and §6.2 and matches at
both asOf dates; Go's `url.ParseQuery` and `strconv.ParseFloat` behave exactly as the plan asserts; no
existing test touches the two new routes, so the nil-dependency choice is safe. The key concern is that
SPEC edits 2 and 3 add a new field-error code (`malformed_query`) and a pseudo-field (`query`) to the
contract, which is a decision above the ticket's pay grade; the plan flags it correctly as an open question.

## 3. Issues

**[MAJOR] Open question 1 / SPEC edits 2-3 — `malformed_query` adds a code and a pseudo-field to the contract**
The ticket says "rules and codes per §5.3/§5.5"; `malformed_query` on field `query` is in neither, and it
is the one change here that downstream tickets must mirror (BC-6 freezes the code enum in OpenAPI; BC-9
maps errors by field and code). The motivation is sound and verified: `url.ParseQuery("keyword=100%&city=Reston")`
returns only `city` plus an error, and `r.URL.Query()` swallows that error, so a hand-typed bare `%` would
silently drop a filter, which is exactly the failure §5.3's unknown-param rule exists to prevent. Downstream
impact is contained: BC-9's description already routes unmapped fields to "a single line under the bar",
and BC-6 reads the enum from §5.5, which edit 3 updates. The plan also documents the fallback (keep
`r.URL.Query()`, record the silent drop as a known limitation).
**Recommendation:** Triage as a "complex" finding: the orchestrator records an explicit ack or decline on
BC-5 (`lattice comment`) before `in_progress`. If acked, note in SPEC edit 2 that only one `malformed_query`
error is ever emitted per request (Go reports the first bad pair only) and that pairs after the bad one are
still parsed. If declined, apply the plan's stated fallback and drop the `keyword=100%` test and curl.

**[MINOR] Parsing rules — precedence between "whitespace-only keyword is absent" and "raw length > 200 is too_long" is unstated**
City is trimmed first, then measured (101 spaces is absent, a 200). Keyword is measured raw, so 201 spaces
is either absent or `too_long` depending on which check runs first; the plan does not say, and the two
fields would behave differently for the same input.
**Recommendation:** State the order. The consistent choice is absence first for both fields (whitespace-only
is absent, full stop), then the length rule on the raw keyword / trimmed city. Add the 201-space keyword row
to `TestParseSearchQueryKeyword` and mirror the sentence in SPEC edit 2.

**[MINOR] Tests table, `TestSearchKeyword` — `NO+Pets` expected result is listed out of rank order**
The matching set is correct (A3, B9, A5, A7: "not" and "no" both contain "no"; B8's "renovated" contains
"no" but has no "pets"). But the plan writes it as `A3,B9,A5,A7`, which is seed order, not the no-budget
newest-first order the response will carry. An implementer who asserts the slice as written gets a failing
test and may "fix" it by weakening the assertion to a set.
**Recommendation:** Write the expectation in rank order: A5, A3, B9, A7. The `pets -> 7` count is right
(B7, A3, B9, A4, A5, A6, A7; A1 and B11 say "pet", not "pets").

**[MINOR] Parsing rules — "`-0` becomes 0" needs an explicit mechanism and a raw-body assertion**
Verified: `strconv.ParseFloat("-0", 64)` returns negative zero and `encoding/json` emits `-0`, so
`minPrice=-0` would echo `"minPrice":-0` unless normalized. The `strictRoundTrip` DeepEqual check cannot
catch this because `-0 == 0` is true for float64.
**Recommendation:** Add one line to the plan: after a successful parse, `if v == 0 { v = 0 }` (assigning the
literal yields positive zero). In the `-0` test row, assert on the raw body text (`"minPrice":0`), not only
on the decoded value.

**[MINOR] SPEC edit 2 — ParseFloat's grammar is wider than "a number" and should be documented once**
Verified: `strconv.ParseFloat` accepts `1_000` (1000), hex floats such as `0x1p4` (16), and `inf`/`nan` in any
case. The plan handles non-finite values, and accepting the rest is per §5.3 ("Numbers parse with
`strconv.ParseFloat`"), so this is not a defect. But BC-6's OpenAPI and BC-9's client-side mirror will
otherwise assume a stricter decimal grammar.
**Recommendation:** One sentence in SPEC edit 2: "number syntax is exactly Go's `strconv.ParseFloat`
(decimal or hex float literals, underscores allowed); only non-finite results are rejected." Optionally one
test row for `1_000` so the behavior is pinned rather than accidental.

**[MINOR] Validation (outside the plan file) — the master validation plan expects the wrong count for `status=active,pending`**
`.lattice/orchestration/validation-plan.md` item 18 expects "11 items" for `status=active,pending` and the
repeated-param form. The seed has 11 active plus 1 pending (A7), so the correct total is 12, which is what
this plan's `TestSearchStatusFilter` asserts.
**Recommendation:** The plan is right; the orchestrator should correct validation-plan item 18 to 12 before
BC-5 reaches the master validator, so the gate does not fail the task on a stale expectation.

## 4. Positive Observations

- **Every external claim checks out.** `store.(*Store).Search`/`Cities` (never nil, sorted, first-seen
  casing), `domain.Rank`/`Paginate`/`RankInput`/`Page`/`RankedItem`, the exported constants, `httpapi.FieldError`,
  `writeJSON`/`writeError`/`RequestIDFrom`, the `Deps` placeholder comment, and `testOrigin` all exist on the
  branch exactly as the plan describes. Nothing in the plan rests on an assumed API.
- **The expected numbers are correct to the last decimal.** I recomputed the Springfield table, the twelve-row
  no-budget order, the twelve-row budget order, the `TestSearchAsOfUTC` value (46 days, 0.3455), and the
  asOf 2026-10-14 live-server values independently; all match, including the fourth-decimal roundings
  (B8 0.538060 to 0.5381, B7 0.729702 to 0.7297). The §5.4 example correction (0.8731 to 0.7414) is right
  and was already sanctioned in the run-state decision log at 00:47Z.
- **Upstream constraints are all honored.** The BC-2 note (own DTO, never embed `Listing`/`Scored`, format
  `ListedDate` with `DateLayout`), the BC-3 plan note (parser enforces page and pageSize bounds first), and
  the BC-3 review note (`targetBudget <= 0` rejected as `not_positive` before `Rank`) each map to a concrete
  line in the plan.
- **Error ordering is deterministic by construction** (allowlist walk, then unknowns sorted) and reproduces
  the §5.5 example body verbatim for `minPrice=5&maxPrice=1&pageSize=0&minprice=1`; the 20-iteration
  map-order test is a cheap guard on top.
- **Acceptance 1 is genuinely implemented, not gestured at.** `strictRoundTrip` decodes with
  `DisallowUnknownFields` and then proves the reverse direction by re-marshaling independent mirror structs
  (no `omitempty`) and comparing as `any`, so a missing key or a stray `null` fails in both directions.
- **Shared-file discipline is explicit and minimal**: only the `Deps` fields, the `Now` default and two
  `HandleFunc` lines in router.go, three literal lines in main.go, so BC-6's one-line route merges cleanly.
  The review hold until BC-4 lands (decision 02:31Z) is acknowledged, and `.lattice/` is kept off the branch.
- **The pipeline function is specified as six commented stages and nothing else**, which is the acceptance
  criterion the interviewer will read; the dedupe slot is a comment between search and rank as §6 and §9.1
  describe.
- **Risks are honest and scoped**: ASCII-only SQLite `lower()`, city internal whitespace, invalid UTF-8
  echo, and the verified `url.ParseQuery` behavior that motivates edit 2 are each named with a disposition.
- **Validation is a real end-to-end run** on a dedicated port with before/after port checks, a recompute of
  every relevance value from `scoring.asOf`, and a named test run for the fixed-clock 0.7414.
