# Ticket card — BC-1

```
TICKET=BC-1
TASK_UUID=task_01M4VX8DZ5P8G1NJGCS5ZRR3JJ
TITLE=Repo bootstrap: initial commit, layout, Makefile, tooling baseline
WORKTREE=/Users/forrest/Code/brillio        # the primary checkout itself (see exceptions below)
BRANCH=master                               # commits directly on master; the only ticket that does
REVIEW_BASE=(none — initial commit)
ANCHOR=none
PORTS=none
AREA=area:8
```

Read `/Users/forrest/Code/brillio/.lattice/orchestration/prompts/delegator-brief.md` now and follow it,
with the BC-1 exceptions below (they override the brief where they conflict).

## BC-1 exceptions to the brief

1. **You work in the primary checkout on `master`.** The repo has zero commits, so there is nothing to branch
   from. The "primary checkout is off limits" rule is lifted for BC-1 only: the implementer edits and commits in
   `/Users/forrest/Code/brillio`. The worktree assertion compares against that path. Skip `lattice branch-link`.
2. **The single commit includes `.lattice/`** (board state, plans, orchestration files — the runtime paths in
   `.lattice/.gitignore` stay ignored). So the pre-commit `.lattice` guard does NOT apply to BC-1. Every other
   file present in the repo root is committed too (CLAUDE.md, agents.md, BRANDING.md, docs/, data/, both PDFs).
   Do not commit anything under `/Users/forrest/Code/brillio-worktrees/` (it is outside the repo anyway).
3. **Exactly one commit**, message `chore: bootstrap repository` plus the Co-Authored-By trailer. Review fixes
   before hand-off may `git commit --amend` that one commit (nobody stands on master yet). After LAND-READY, no
   amend.
4. **"git status is clean afterward"** means clean outside `.lattice/`: the live board writes to tracked
   `.lattice/` files on every status change after the commit (run-state decision log).
5. **Makefile — orchestrator ruling (run-state decision 1):** besides the phony targets the ticket lists
   (`up down dev-api dev-web test lint check smoke dist`), add per-half targets `test-api test-web lint-api lint-web`.
   `test` = `test-api test-web`, `lint` = `lint-api lint-web`, `check` = `lint test`. Every target with nothing to
   run yet prints `not yet implemented: <target> (lands in BC-N)` to stderr and exits 1, so `make check` fails
   loudly. This lets BC-2 (api halves) and BC-7 (web halves) fill disjoint lines in parallel. Keep it plain
   POSIX make with a short comment header; no emoji.
6. **Root `.gitignore`** must not ignore `.lattice/` or `data/sample_listings.json`. It covers Go (`/backend/tmp/`,
   binaries), Node (`node_modules/`, `/frontend/dist/`), `*.db` and `*.db-*`, `.DS_Store`, `.env`, `/dist/`, coverage output.
7. **Code review:** skip `lattice code-review` (the initial-commit diff includes PDFs and the whole board and exceeds
   `review_max_diff_chars`). Go straight to the own-reviewer path: a fresh-context reviewer sub-agent reviews the
   files BC-1 authored (`.gitignore`, `.editorconfig`, `Makefile`, `README.md`, the `data/` move, the `.gitkeep`s)
   plus `git show --stat HEAD` against the ticket's acceptance list, and its review names the commit sha.
8. **Validation:** N/A (no running system). The validator sub-agent records a one-line justification citing the real
   output of `git log --oneline`, `git status --porcelain -- . ':!.lattice'`, `make check; echo exit=$?`, and
   `ls data/ sample_listings.json`.
9. Hand-off: send `LAND-READY BC-1 BRANCH master HEAD <sha> ...` as usual. The Orchestrator verifies and moves
   BC-1 to `pr_open` then `done` (no merge needed). Then BC-2 and BC-7 start from your commit.
