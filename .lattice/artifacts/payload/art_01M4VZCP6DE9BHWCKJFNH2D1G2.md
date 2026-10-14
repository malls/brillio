# Plan Review: BC-1 (Repo bootstrap)

Reviewer: agent:fable-5.1-plan-reviewer. Method: every factual claim in the plan was checked
against the live tree and host, and the plan's own Makefile, .gitignore, .editorconfig and README
blocks were extracted verbatim and executed in scratch directories outside the repo (then deleted).
The repository was not modified.

## 1. Verdict

**PASS**

## 2. Summary

The plan covers every scope bullet and acceptance criterion in the ticket and every ruling in the
BC-1 card and run-state, with an executable test list mapped one-to-one to the acceptance bullets.
Reproducing that test list on this host gave exactly the outputs the plan quotes (make exit codes
and messages, the 27/0 check-ignore split, the four-line commit-message shape, the `:!.lattice`
pathspec behaviour), the preconditions match the tree byte-for-byte, and the parallel BC-2/BC-7
Makefile merge claim re-simulated cleanly. The only concern worth fixing before implementation is
that `$SCRATCH` is used throughout the test list but never defined, which matters because Bash
calls do not share environment between invocations.

## 3. Issues

**[MINOR] Approach / Test list (T1.2, T3.6, C5, Validation) — `$SCRATCH` is never defined**
The plan says "`$SCRATCH` is the sub-agent's scratchpad dir, never the repo" but neither the plan
nor the delegator brief defines it, and the brief states explicitly that Bash calls do not keep
`export`s between calls. In zsh an unset variable expands to empty, so `"$SCRATCH/bc1-expected.txt"`
becomes `/bc1-expected.txt` and the redirect fails with permission denied, or the implementer
improvises a location per call. The plan is written to be followed mechanically, so this should
not be left to improvisation.
**Recommendation:** Add to step 0: `mkdir -p /tmp/bc1-scratch`, and replace every `$SCRATCH`
with the literal `/tmp/bc1-scratch` (or state once that each call using it must begin with
`SCRATCH=/tmp/bc1-scratch;`). Add `rm -rf /tmp/bc1-scratch` to step 8.

**[MINOR] Files / .gitignore — `/backend/api` also ignores a future `backend/api/` directory**
Verified: `git check-ignore` reports `backend/api/handlers.go` and `backend/api/` as ignored by
that line. SPEC §3 has no `backend/api/` directory today (`cmd/api/`, `internal/httpapi/`), and
where BC-2 writes its binary is BC-2's choice (air's default is `tmp/`, already covered by
`/backend/tmp/`). The rule guards a binary nothing yet produces while silently hiding a plausible
future package path, which is the class of bug that is hard to notice later.
**Recommendation:** Drop the `/backend/api` line (`/backend/tmp/` and `/backend/bin/` cover
build output) and let BC-2 add a rule for wherever it actually builds; update the C1 "MUST be
ignored" list to 26 entries and the .gitignore line count to 35. If kept, add a comment on the
line saying it must be removed if a `backend/api/` package is ever created.

**[MINOR] Risks vs Size estimate — inconsistent file counts**
Risks says the global excludes hide "none of the 111 real files that must be committed"; Size
estimate says the commit is 80-90 files (14 non-board + 65 board). Today the committed set is
79 files; 111 is roughly the whole working tree including ignored runtime paths. The sweep
result is still correct (re-run: nothing outside `.lattice` runtime dirs and `.claude` is
ignored), only the sentence is wrong.
**Recommendation:** Change the Risks sentence to "any file in the tree that must be committed
(79 at plan time)" so the two sections agree.

**[MINOR] Approach step 0 — `wc -l` output is space-padded on macOS**
`git rev-list --all | wc -l` prints `       0` on this host, not `0`; the same applies to the
`wc -l` checks in T1.1 and T2.1. Fine when read by eye, wrong if an implementer compares the
string.
**Recommendation:** Use `git rev-list --all --count` (prints `0` unpadded, verified on the
unborn branch) in step 0 and `git rev-list --count HEAD` alone in T1.1; for T2.1 keep `wc -l`
but note the padding, or pipe through `tr -d ' '`.

## 4. Positive Observations

- **Every quoted output reproduced.** T3.1-T3.7 (including the `-k` ordering and the
  ``Target `check' not remade`` line under GNU Make 3.81), the 10-tab/0-space/5-PHONY hygiene
  counts, the 27 ignored and 0 wrongly-ignored check-ignore probes, the 26/4-section
  .editorconfig, the 9-line README, and the four-line `cat -e` shape of the two-`-m` commit
  message all matched exactly.
- **Preconditions match the live tree.** Unborn `refs/heads/master`, zero commits, the exact
  `ls -A1` set, `.claude/` holding only the excluded lock file, no `.DS_Store`, README at 0
  bytes, data file at mode 0600 with the quoted SHA-256 and 12 listings, PDFs at 125 KB and
  66 KB, 65 non-runtime `.lattice` files, and the empty-directory list under `.lattice/`.
- **The shared-file design holds.** Re-simulated from the plan's Makefile: a BC-2 branch and a
  BC-7 branch each rewrote their three stubs with multi-line recipes, added a macro on their
  `.PHONY` line and appended a new phony target; both `--no-ff` merges were clean and
  `make -n check` printed all four real commands. The per-ticket block layout with its own
  `.PHONY` line is the right call, and the rule is written into the Makefile header where the
  next ticket will see it.
- **Alignment is tight.** Everything beyond the ticket's literal text (per-half targets, coverage
  ignores, Makefile tab rule in .editorconfig, the Co-Authored-By trailer) traces to the card,
  run-state decision 1, or the delegator brief. The README paragraph was checked against the data
  (sources are `MLS_A` and `MLS_B`) and SPEC §1 and is accurate.
- **Risk handling is specific, not generic.** Host-global excludes are neutralised with
  `-c core.excludesFile=/dev/null` and re-swept against the real tree; the exit-code 2 nuance,
  `.env.*` mode files, Markdown trailing-space semantics, the 0600 bind-mount problem, and the
  tab-vs-space recipe hazard each get a concrete mitigation. The explicit staging list plus the C5
  guard before the commit and T1.4 after it close the "runtime board paths committed" risk from
  both sides.
- **The contradictions section is honest** and resolves each conflict (exit 1 vs 2, 9 vs 13
  targets, "clean" narrowed to outside `.lattice`, `.claude/` new since the ticket) in favour of
  the binding card without silently rewriting the ticket.
