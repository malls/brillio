# BC-1: Repo bootstrap: initial commit, layout, Makefile, tooling baseline

Planner: agent:opus-5.5-planner. Sources read: `lattice show BC-1`, docs/SPEC.md §1-3, §8, §10-12,
BRANDING.md §8, .lattice/context.md, .lattice/.gitignore, .lattice/orchestration/run-state.md,
.lattice/orchestration/prompts/BC-1.md (binding card; overrides the ticket where they conflict),
.lattice/orchestration/prompts/delegator-brief.md.

Plan-time prototypes (scratchpad, not the repo): the Makefile below was run under the host's
`/usr/bin/make` (GNU Make 3.81) and the .gitignore under git 2.56.0. Every expected output quoted in
the Test list was observed there, including a simulated parallel BC-2/BC-7 edit that merged with no
conflict.

## Scope

In:
- Root `.gitignore`, `.editorconfig`, `Makefile` (stub skeleton), `README.md` (stub) at the repo root.
- `mv sample_listings.json data/sample_listings.json` and normalise its mode 0600 -> 0644.
- Empty `backend/.gitkeep` and `frontend/.gitkeep`.
- Exactly ONE commit on `master` (the repo has zero commits; HEAD is the unborn `refs/heads/master`)
  containing the 14 non-board paths listed in Approach step 6 plus the tracked `.lattice/` board.
  Message: `chore: bootstrap repository`, blank line, `Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>`.

Out:
- Any Go or TypeScript code, Docker/Compose files, CI, `docs/EXTENSIONS.md`, the real README (BC-15).
- Editing any existing file (CLAUDE.md, agents.md, BRANDING.md, docs/*, the PDFs, the data file's
  contents, anything in `.lattice/`). They are committed byte-for-byte as they are.
- `.claude/` (Claude Code runtime dir now present at the root; its only file,
  `.claude/scheduled_tasks.lock`, is excluded by `.git/info/exclude`). Never stage it.
- Anything under `/Users/forrest/Code/brillio-worktrees/`.
- Lattice status changes (the delegator owns them). Branch linking (card: skip for BC-1).

SPEC edits: none. SPEC §3 lists the Makefile targets `up, down, dev-api, dev-web, test, lint, check,
smoke, dist`. The per-half targets `test-api test-web lint-api lint-web` are an addition from the
orchestrator ruling in run-state, not a change to anything SPEC says, so no SPEC edit is needed.

## Decisions (with reasons)

1. **Makefile composite targets have no recipe of their own, only prerequisites.** `check: lint test`,
   `lint: lint-api lint-web`, `test: test-api test-web`. They have nothing of their own to run. A
   recipe such as `@echo ok` would only add noise. When a prerequisite fails, make stops with that
   prerequisite's own message, which is the clearest error. It also means BC-2 and BC-7 never need to
   touch the composite block.
2. **`make check` stops at the first failure** (make's default): it prints `not yet implemented: lint-api
   (lands in BC-2)` and exits. `make -k check` shows all four half stubs. Both exit non-zero.
3. **Exit code is 2, not 1.** Each stub recipe runs `exit 1`, as the card says. GNU make reports any
   failed target by exiting **2** itself. The tests therefore check for "non-zero (observed: 2)".
4. **Default goal is `check`.** It is the first target in the file, so plain `make` runs the gate. It has
   no side effects, unlike `up`, which would start Docker. `.DEFAULT_GOAL` is GNU-only and is not used.
5. **No `.POSIX:` special target.** GNU make 3.81 (host) and 4.x treat `.POSIX` differently (4.x passes
   `-e` to the shell). Leaving it out keeps behaviour the same everywhere. "Plain POSIX make" is met by
   syntax: only `.PHONY`, plain rules and (later) `NAME = value` macros. The file has no `ifeq`,
   `$(shell ...)`, `$(wildcard ...)`, `%` pattern rules, `:=`, `.ONESHELL` or `.DEFAULT_GOAL`.
6. **One block per owning ticket, each with its own `.PHONY:` line.** `.PHONY` may be declared more
   than once. This keeps BC-2's and BC-7's edits (including new phony targets) to separate lines; see
   "Shared-file handling".
7. **Stub message goes to stderr:** `@echo 'not yet implemented: <target> (lands in BC-N)' >&2; exit 1`.
   Single quotes, ASCII only.
8. **.gitignore patterns are anchored where a bare pattern could reach too far.** It uses `/dist/`, not
   `dist`, and `/frontend/dist/`, `/backend/tmp/`, `/backend/bin/`, `/backend/api`. It has no `*.json`,
   no `*.log` (this could one day hide a tracked `.lattice` file) and no `.lattice` entry. `.env.*` is NOT
   ignored wholesale, because Vite's `.env.development`-style mode files may legitimately be committed.
   Only `.env`, `.env.local` and `.env.*.local` are ignored (the Vite convention).
9. **.editorconfig sets 2-space as the `[*]` default and overrides it.** Go (`*.go`, `go.mod`,
   `go.work`) and Make (`Makefile`, `*.mk`) use tabs, with `indent_size = tab` so they do not inherit
   the width 2. Markdown sets `trim_trailing_whitespace = false`: two trailing spaces are a hard line
   break in Markdown, and CLAUDE.md already has a trailing-whitespace line that an auto-trimming editor
   would turn into a noise diff in a user-owned file.
10. **Data file mode -> 0644.** Git records the file as `100644` either way, because git tracks only
    the executable bit. The chmod is about the primary checkout's working tree: BC-13 bind-mounts
    `./data` into containers, and a 0600 file is unreadable to any non-root container user. It is
    harmless and recommended.

## Approach (ordered; follow mechanically)

All commands run in the primary checkout: begin every Bash call with `cd /Users/forrest/Code/brillio &&`.
The shell is zsh, so write path lists literally (zsh does not word-split `$VAR`). Do not `git add -A`
or `git add .`. Do not run `git reset`, `git stash` or `git clean` here: the live board is in this
tree.

0. **Preconditions (STOP and report to the delegator if any fails):**
   - `test "$(pwd)" = /Users/forrest/Code/brillio`
   - `git symbolic-ref HEAD` prints `refs/heads/master`. `git rev-list --all | wc -l` prints `0` (no
     commit exists yet; if one does, someone else bootstrapped: stop).
   - `/bin/ls -A1` prints exactly these entries, in any order: `.claude .git .lattice agents.md
     BRANDING.md CLAUDE.md Coding_Test_Question_Paper.pdf docs Instructions.pdf README.md
     sample_listings.json`. `.claude` may be absent. Any other entry is unexplained work by someone
     else: stop and ask, do not delete it.
   - `make --version | head -1` prints `GNU Make 3.81`, or any GNU Make (note the version in the report).
   - `find . -name .DS_Store -not -path './.git/*'` prints nothing. If it finds any, leave them alone:
     the new .gitignore and the explicit staging list keep them out.
   - Record the data checksum: `shasum -a 256 sample_listings.json`. Expected at plan time:
     `e785664fb4bde1c5534be07fb13cd6be8fc6c4673bce141508b34bb9f06efe07`. If it differs, use the
     recomputed value in T4.2.
1. **Move the data file:** `mkdir -p data && mv sample_listings.json data/sample_listings.json && chmod 644 data/sample_listings.json`
2. **Keep files:** `mkdir -p backend frontend && : > backend/.gitkeep && : > frontend/.gitkeep`. Both files are empty (0 bytes).
3. **Write `.gitignore`, `.editorconfig` and `README.md`** byte-exactly as given in "Files" below. Use a
   quoted heredoc (`cat > FILE <<'EOF'`) or the Write tool. README.md already exists at 0 bytes, so if
   you use the Write tool, Read it first. Each file ends with exactly one newline.
4. **Write `Makefile`.** Copy the Makefile block verbatim, including the literal marker `<TAB>` at the
   start of every recipe line. Then convert the markers: `perl -pi -e 's/^<TAB>/\t/' Makefile`. Verify
   with T3.7 before going on.
5. **Pre-commit checks:** run T3.1-T3.7, C1-C4 and T4.1-T4.3 (the working-tree halves). Then run
   `git status --porcelain --ignored`. Its `!!` lines may only be `.claude/` and paths under
   `.lattice/locks/`, `.lattice/review_state/`, `.lattice/tmp-prompts/` or `.lattice/.daemon/`.
   Anything else is an ignore-rule bug: stop.
6. **Stage explicit paths only:**
   ```
   git add -- .gitignore .editorconfig Makefile README.md BRANDING.md CLAUDE.md agents.md Coding_Test_Question_Paper.pdf Instructions.pdf docs data backend/.gitkeep frontend/.gitkeep .lattice
   ```
   `.lattice` is staged as a directory path. Its own `.gitignore` filters the runtime dirs. Run the C5
   staged-list guard. It must pass before committing.
7. **Commit** (no `--no-verify`; no hooks exist; the user's git identity is already configured):
   ```
   git commit -m 'chore: bootstrap repository' -m 'Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>'
   ```
   Two `-m` flags produce subject, blank line, trailer (verified at plan time).
8. **Post-commit checks:** T1.1-T1.6, T2.1, T3.1, T4.1-T4.3. Report: the commit sha, the full `git show
   --stat HEAD` file count, every test's observed output, and the `make --version` line.

**Amend procedure (review fixes before LAND-READY only, card item 3):** edit the file, re-run its
tests, `git add -- <that file>` (do not re-stage `.lattice`), `git commit --amend --no-edit`. Then
re-run the whole Test list. The sha changes, so report the new one. After LAND-READY, never amend.

## Files (exact contents)

### `.gitignore`

```
# Brillio Listings - root ignore rules.
# .lattice/ is tracked on purpose (board state); its runtime paths are
# ignored by .lattice/.gitignore. data/sample_listings.json is tracked.

# macOS
.DS_Store

# Environment files with local values or secrets
.env
.env.local
.env.*.local

# Go (backend): air build output, built binaries, test binaries
/backend/tmp/
/backend/bin/
/backend/api
*.exe
*.test

# Node (frontend)
node_modules/
/frontend/dist/
*.tsbuildinfo
npm-debug.log*

# SQLite databases and their journal / WAL / SHM side files
*.db
*.db-*

# Coverage output
coverage.out
*.coverprofile
/frontend/coverage/

# Release zip output (make dist)
/dist/
```

### `.editorconfig`

```
# EditorConfig - https://editorconfig.org
root = true

# Every file: UTF-8, LF line endings, final newline, no trailing whitespace,
# 2-space indent (ts, tsx, js, json, yaml, yml, md, css, html and the rest).
[*]
charset = utf-8
end_of_line = lf
insert_final_newline = true
trim_trailing_whitespace = true
indent_style = space
indent_size = 2

# Markdown: two trailing spaces are a hard line break, so keep them.
[*.md]
trim_trailing_whitespace = false

# Go: gofmt indents with tabs.
[{*.go,go.mod,go.work}]
indent_style = tab
indent_size = tab

# Make: recipe lines must start with a TAB.
[{Makefile,*.mk}]
indent_style = tab
indent_size = tab
```

### `Makefile`

Every line that starts with the literal text `<TAB>` must start with ONE real TAB character instead.
Step 4 converts the markers. There are exactly 10 such lines, and no line may start with a space.

```
# Brillio Listings - developer entry points (docs/SPEC.md sections 3, 8, 10, 12).
# `make` alone runs `make check` (lint, then tests, both halves).
# A stub prints "not yet implemented: <target> (lands in BC-N)" to stderr and
# fails, so the gate fails loudly until ticket BC-N replaces its recipe.
# One block per owning ticket: edit only your own block (including its own
# .PHONY line) so parallel branches merge cleanly. Plain POSIX make, no GNU
# extensions (host make is GNU Make 3.81). Recipe lines start with a TAB.

# --- Gate (BC-1): composite targets, prerequisites only, no recipes ---
.PHONY: check lint test
check: lint test
lint: lint-api lint-web
test: test-api test-web

# --- Backend (BC-2) ---
.PHONY: dev-api lint-api test-api
dev-api:
<TAB>@echo 'not yet implemented: dev-api (lands in BC-2)' >&2; exit 1
lint-api:
<TAB>@echo 'not yet implemented: lint-api (lands in BC-2)' >&2; exit 1
test-api:
<TAB>@echo 'not yet implemented: test-api (lands in BC-2)' >&2; exit 1

# --- Frontend (BC-7) ---
.PHONY: dev-web lint-web test-web
dev-web:
<TAB>@echo 'not yet implemented: dev-web (lands in BC-7)' >&2; exit 1
lint-web:
<TAB>@echo 'not yet implemented: lint-web (lands in BC-7)' >&2; exit 1
test-web:
<TAB>@echo 'not yet implemented: test-web (lands in BC-7)' >&2; exit 1

# --- Compose stack (BC-13) ---
.PHONY: up down smoke
up:
<TAB>@echo 'not yet implemented: up (lands in BC-13)' >&2; exit 1
down:
<TAB>@echo 'not yet implemented: down (lands in BC-13)' >&2; exit 1
smoke:
<TAB>@echo 'not yet implemented: smoke (lands in BC-13)' >&2; exit 1

# --- Release (BC-17) ---
.PHONY: dist
dist:
<TAB>@echo 'not yet implemented: dist (lands in BC-17)' >&2; exit 1
```

### `README.md`

```
# Brillio Listings

Brillio Listings is a full-stack listing search service built for the Brillio software
engineer take-home assessment: a Go REST API over the sample property listings
(`data/sample_listings.json`, two MLS feeds, stored in SQLite) and a React/TypeScript UI
that filters them, ranks them by a budget-and-recency relevance score, and pages through
the results. The engineering contract is `docs/SPEC.md`.

Status: scaffolding
```

### `backend/.gitkeep`, `frontend/.gitkeep`

Empty (0 bytes).

## Test list (mapped one-to-one to the acceptance bullets, then the card rulings)

Expected outputs below were observed at plan time. They come from running these exact commands under
zsh (the project shell) against the file contents extracted from this plan. "exit=" lines come from
appending `; echo exit=$?`.

### A1 - "git log shows exactly one commit on master containing the files above"

- **T1.1 one commit, one ref:** `git log --oneline | wc -l` -> `1`. `git rev-list --count HEAD` -> `1`.
  `git for-each-ref --format='%(refname)'` -> exactly `refs/heads/master`. `git symbolic-ref HEAD`
  -> `refs/heads/master`.
- **T1.2 exact non-board file set:** write the expected list to a scratch file, then diff it:
  ```
  cat > "$SCRATCH/bc1-expected.txt" <<'EOF'
  .editorconfig
  .gitignore
  BRANDING.md
  CLAUDE.md
  Coding_Test_Question_Paper.pdf
  Instructions.pdf
  Makefile
  README.md
  agents.md
  backend/.gitkeep
  data/sample_listings.json
  docs/SPEC.md
  docs/branding-demo.html
  frontend/.gitkeep
  EOF
  git ls-tree -r --name-only HEAD | grep -v '^\.lattice/' | LC_ALL=C sort | diff - "$SCRATCH/bc1-expected.txt"; echo exit=$?
  ```
  Expected: no diff output, `exit=0`. `$SCRATCH` is the sub-agent's scratchpad dir, never the repo.
- **T1.3 board is in the commit:**
  `for p in .lattice/.gitignore .lattice/config.json .lattice/context.md .lattice/ids.json .lattice/orchestration/run-state.md .lattice/orchestration/prompts/BC-1.md .lattice/tasks/task_01M4VX8DZ5P8G1NJGCS5ZRR3JJ.json .lattice/events/task_01M4VX8DZ5P8G1NJGCS5ZRR3JJ.jsonl .lattice/plans/task_01M4VX8DZ5P8G1NJGCS5ZRR3JJ.md; do git cat-file -e "HEAD:$p" || echo "MISSING $p"; done`
  -> no output. `git ls-tree -r --name-only HEAD | grep -c '^\.lattice/'` -> a positive count (65 non-lock
  `.lattice` files existed at plan time, and plan or code review adds a few; report the number).
- **T1.4 no runtime or junk paths committed:**
  `git ls-tree -r --name-only HEAD | grep -E '^\.lattice/(review_state|tmp-prompts|\.daemon|locks)/|(^|/)\.claude/|(^|/)\.DS_Store$|sample_listings\.json$' | grep -v '^data/sample_listings\.json$'; echo exit=$?`
  -> no output, `exit=1`. The last pattern also proves the root copy is not in the tree.
- **T1.5 commit message:**
  - `git log -1 --format=%s` -> `chore: bootstrap repository`
  - `git log -1 --format=%B | cat -e` -> exactly these lines:
    `chore: bootstrap repository$`, `$`, `Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>$`, `$`
  - `git log -1 --format='%(trailers:key=Co-Authored-By,valueonly)'` -> `Claude Opus 5.5 <noreply@anthropic.com>`
  - no emoji / non-ASCII: `git log -1 --format=%B | LC_ALL=C grep -n "$(printf '[^\t -~]')"; echo exit=$?` -> `exit=1`
- **T1.6 stat for the reviewer:** `git show --stat --format='%H %s' HEAD | tail -1` -> `N files changed, ...`.
  Report N and the sha (card item 7: the reviewer reads `git show --stat HEAD`).

### A2 - "git status is clean afterward" (card item 4: outside `.lattice/`)

- **T2.1:** `git status --porcelain -- . ':!.lattice' | wc -l` -> `0`. The exclude pathspec was
  verified at plan time: untracked `.lattice/*` files do not appear. Tracked `.lattice` files modified
  by the live board after the commit are expected and allowed.
- **T2.2 (informational):** `git status --porcelain --ignored -- . ':!.lattice'` -> empty, or `!! .claude/` only.

### A3 - "make check exits non-zero with a clear 'not yet implemented' message"

- **T3.1:** `make check; echo exit=$?` ->
  ```
  not yet implemented: lint-api (lands in BC-2)
  make: *** [lint-api] Error 1
  exit=2
  ```
- **T3.2 message is on stderr, stdout is silent:** `make check 2>/dev/null | wc -c` -> `0`.
  `make check 2>&1 >/dev/null | grep -cx 'not yet implemented: lint-api (lands in BC-2)'` -> `1`.
- **T3.3 default goal:** `make; echo exit=$?` -> same two lines as T3.1, `exit=2`.
- **T3.4 both halves are wired into check:** `make -k check; echo exit=$?` -> in order:
  `not yet implemented: lint-api (lands in BC-2)`, `make: *** [lint-api] Error 1`,
  `not yet implemented: lint-web (lands in BC-7)`, `make: *** [lint-web] Error 1`,
  `not yet implemented: test-api (lands in BC-2)`, `make: *** [test-api] Error 1`,
  `not yet implemented: test-web (lands in BC-7)`, `make: *** [test-web] Error 1`,
  ``make: Target `check' not remade because of errors.``, `exit=2`.
- **T3.5 dependency graph (dry run):** `make -n check; echo exit=$?` -> these four lines, in order, then `exit=0`:
  `echo 'not yet implemented: lint-api (lands in BC-2)' >&2; exit 1`, then the same for lint-web
  (BC-7), test-api (BC-2) and test-web (BC-7).
- **T3.6 every target individually:**
  ```
  for t in up down dev-api dev-web test-api test-web lint-api lint-web smoke dist test lint check; do make "$t" >"$SCRATCH/$t.out" 2>"$SCRATCH/$t.err"; printf '%-9s rc=%s stdout=%s %s\n' "$t" "$?" "$(wc -c <"$SCRATCH/$t.out" | tr -d ' ')" "$(head -1 "$SCRATCH/$t.err")"; done
  ```
  (`$?` here is make's status, because `printf` expands its arguments before it runs.) Expected rows:
  every one has `rc=2 stdout=0`, followed by the first stderr line:
  up -> `not yet implemented: up (lands in BC-13)`; down -> `... down (lands in BC-13)`;
  dev-api -> `... dev-api (lands in BC-2)`; dev-web -> `... dev-web (lands in BC-7)`;
  test-api -> `... test-api (lands in BC-2)`; test-web -> `... test-web (lands in BC-7)`;
  lint-api -> `... lint-api (lands in BC-2)`; lint-web -> `... lint-web (lands in BC-7)`;
  smoke -> `... smoke (lands in BC-13)`; dist -> `... dist (lands in BC-17)`;
  test -> `... test-api (lands in BC-2)`; lint -> `... lint-api (lands in BC-2)`;
  check -> `... lint-api (lands in BC-2)`.
- **T3.7 Makefile hygiene:**
  - `grep -c "$(printf '^\t')" Makefile` -> `10`
  - `grep -c '^ ' Makefile` -> `0`
  - `grep -c '<TAB>' Makefile` -> `0`
  - `LC_ALL=C grep -n "$(printf '[^\t -~]')" Makefile; echo exit=$?` -> `exit=1`
  - `grep -nE '^(ifeq|ifneq|ifdef|ifndef|define|include|override|export|vpath)|\$\((shell|wildcard|patsubst|foreach|call|eval)|:=|^\.(ONESHELL|DEFAULT_GOAL|POSIX)|%' Makefile; echo exit=$?` -> `exit=1`
  - `grep -c '^\.PHONY:' Makefile` -> `5`

### A4 - "data/sample_listings.json exists and the root copy is gone"

- **T4.1:** `test -f data/sample_listings.json && ! test -e sample_listings.json && echo ok` -> `ok`.
  `ls data/ sample_listings.json; echo exit=$?` -> `ls: sample_listings.json: No such file or directory`,
  `data/:`, `sample_listings.json`, `exit=1`.
- **T4.2 content unchanged:** `shasum -a 256 data/sample_listings.json` and (after the commit)
  `git show HEAD:data/sample_listings.json | shasum -a 256` -> both
  `e785664fb4bde1c5534be07fb13cd6be8fc6c4673bce141508b34bb9f06efe07` (or the value recorded in step 0).
  `python3 -c "import json;print(len(json.load(open('data/sample_listings.json'))))"` -> `12`.
- **T4.3 mode:** `stat -f %Lp data/sample_listings.json` -> `644`. After the commit,
  `git ls-files -s data/sample_listings.json | cut -c1-6` -> `100644`.

### Card rulings (6: .gitignore, 3: .editorconfig, 5: .gitkeep, README, 2/7: staging)

- **C1 .gitignore is correct on its own.** `-c core.excludesFile=/dev/null` disables the host's global
  ignore file (see Risks), so this tests only the repo's rules. `--no-index` tests the patterns even
  for tracked paths.
  - Must NOT be ignored (expect no output, `exit=1`):
    `git -c core.excludesFile=/dev/null check-ignore --no-index -v .lattice/.gitignore .lattice/config.json .lattice/ids.json .lattice/context.md .lattice/tasks/task_01M4VX8DZ5P8G1NJGCS5ZRR3JJ.json .lattice/events/_lifecycle.jsonl .lattice/plans/task_01M4VX8DZ5P8G1NJGCS5ZRR3JJ.md .lattice/orchestration/run-state.md .lattice/orchestration/delivery-receipts.jsonl data/sample_listings.json docs/SPEC.md docs/branding-demo.html README.md Makefile .editorconfig backend/.gitkeep frontend/.gitkeep backend/cmd/api/main.go backend/internal/store/migrations/0001_init.sql frontend/src/components/Header/Header.test.tsx frontend/src/api/client.test.ts frontend/.env.development .env.example frontend/package-lock.json compose.yaml docs/EXTENSIONS.md; echo exit=$?`
  - MUST be ignored (expect 27 lines, each sourced from `.gitignore:` or `.lattice/.gitignore:`, `exit=0`):
    `git -c core.excludesFile=/dev/null check-ignore --no-index -v frontend/node_modules/x/index.js node_modules/x backend/tmp/main backend/bin/api backend/api backend/internal/store/store.test listings.db data/listings.db data/listings.db-wal data/listings.db-shm data/listings.db-journal .env frontend/.env frontend/.env.local frontend/.env.production.local .DS_Store docs/.DS_Store dist/brillio-listings.zip frontend/dist/index.html coverage.out backend/coverage.out frontend/coverage/index.html frontend/tsconfig.tsbuildinfo .lattice/locks/x.lock .lattice/review_state/x.json .lattice/.daemon/auto.log .lattice/tmp-prompts/x.md; echo exit=$?`
    Add `| wc -l` (without `-v`) to confirm the count is `27`. `data/listings.db` matters most: SPEC §11's
    default `DB_PATH` is `./data/listings.db`, inside the tracked `data/` dir.
  - Real-tree sweep: `find . -not -path './.git/*' -not -path './.git' | sed 's|^\./||' | git check-ignore --stdin -v | grep -vE '^(\.lattice/\.gitignore:[0-9]+:(locks|review_state|tmp-prompts|\.daemon)/|\.git/info/exclude:[0-9]+:\*\*/\.claude/)'; echo exit=$?`
    -> no output, `exit=1`. Nothing real is ignored except board runtime dirs and `.claude` runtime files.
- **C2 .editorconfig and authored-file hygiene:**
  - `grep -c '^root = true$' .editorconfig` -> `1`; `grep -c '^\[' .editorconfig` -> `4`.
  - For each of `.gitignore .editorconfig Makefile README.md`: `LC_ALL=C grep -c "$(printf '\r')"` -> `0`;
    `grep -c ' $'` -> `0`; `tail -c1 FILE | od -An -tx1 | tr -d ' '` -> `0a`;
    `LC_ALL=C grep -n "$(printf '[^\t -~]')" FILE; echo exit=$?` -> `exit=1` (ASCII only, no emoji).
- **C3 keep files:** `test -f backend/.gitkeep && ! test -s backend/.gitkeep && test -f frontend/.gitkeep && ! test -s frontend/.gitkeep && echo ok` -> `ok`.
- **C4 README:** `head -1 README.md` -> `# Brillio Listings`; `grep -cx 'Status: scaffolding' README.md` -> `1`;
  `grep -c '' README.md` -> `9` (line count).
- **C5 staged-list guard (between step 6 and step 7):**
  - `git diff --cached --name-only | grep -E '^\.lattice/(review_state|tmp-prompts|\.daemon|locks)/|(^|/)\.claude/|(^|/)\.DS_Store$'; echo exit=$?` -> `exit=1`
  - `git diff --cached --name-only | grep -v '^\.lattice/' | LC_ALL=C sort | diff - "$SCRATCH/bc1-expected.txt"; echo exit=$?` -> `exit=0`
  - `git status --porcelain -- . ':!.lattice' | grep -v '^A  '; echo exit=$?` -> `exit=1`. Every
    non-board entry is staged as an addition; there are no `??` entries.

## Validation steps

N/A: no running system; this is a repo bootstrap only (ticket Validation block, card item 8). The
validator sub-agent runs these four commands at the reviewed HEAD and records the real output:

```
git log --oneline
git status --porcelain -- . ':!.lattice'
make check; echo exit=$?
ls data/ sample_listings.json
```

It writes `$EVIDENCE_DIR/BC-1-validation.md` and attaches it with `--role validation`, in this shape:

`N/A (no running system: repository bootstrap). At HEAD <sha>: git log --oneline -> "<short> chore:
bootstrap repository" (1 commit); git status --porcelain -- . ':!.lattice' -> (empty); make check ->
"not yet implemented: lint-api (lands in BC-2)", "make: *** [lint-api] Error 1", exit=2; ls data/
sample_listings.json -> "data/: sample_listings.json" and "ls: sample_listings.json: No such file or
directory".`

The evidence file lives in `.lattice/orchestration/evidence/` and stays untracked after the commit.
That is fine under card item 4.

## Shared-file handling (run-state shared-files table)

- **Makefile.** BC-1 owns and creates it. Later writers, in landing order: BC-2 (api halves + dev-api),
  BC-7 (web halves + dev-web), BC-13 (up/down/smoke), BC-17 (dist). Conflict rule: union, ticket order.
  - BC-2 and BC-7 run in parallel, so their edits must never touch the same or adjacent lines. A git
    three-way merge conflicts when changed regions overlap or touch. The layout keeps them apart:
    - Each owning ticket gets a block that starts with a `# --- <Name> (BC-N) ---` header and its own
      `.PHONY:` line.
    - A blank line and the next block's header separate the blocks. Neither BC-2 nor BC-7 edits those
      separator lines.
  - Rule for later tickets (it is in the Makefile header): replace your stub recipes and add new
    targets, macros (`GO = go`, `NPM = npm`) and `.PHONY` names only on lines strictly inside your
    own block. Never edit the Gate block, the header, the separators or another ticket's block. Keep
    to plain POSIX syntax that GNU Make 3.81 accepts: no `!=`, `::=`, `.ONESHELL`, `$(shell)`,
    pattern rules or `.DEFAULT_GOAL`.
  - Verified at plan time: from the BC-1 Makefile as base, a BC-2 branch and a BC-7 branch each
    rewrote their 3 stubs (with multi-line recipes), added a macro and appended a new phony target
    after their last stub. `git merge --no-ff` of both into master merged with no conflict, and
    `make -n check` printed both halves' real commands.
  - The per-half targets also serve the run-state Gate exception. Until both scaffolds land, the
    landing ticket proves its own half (`make lint-api test-api` for BC-2, `make lint-web test-web`
    for BC-7) while the other half is still a stub.
- **README.md.** BC-1 owns the stub; BC-15 writes the real file. The stub is deliberately 9 lines so
  BC-15 can replace it wholesale. No other ticket edits it.
- **.gitignore and .editorconfig** are not in the shared table. Later tickets may append rules (for
  example, BC-2 for golangci-lint caches). Keep each addition inside its own commented section so
  appends from parallel branches merge as a union.
- **`.lattice/`** is committed here only (card item 2). Every later feature branch keeps its
  pre-commit `.lattice` guard.

## Risks

- **Runtime `.lattice` paths committed.** `locks/` exists now. `review_state/`, `tmp-prompts/` and
  `.daemon/` appear once plan-review or code-review runs. `.lattice/.gitignore` covers all four.
  Mitigation: the C5 guard before the commit and T1.4 after it.
- **Host-global ignore rules hide files or mask a weak repo .gitignore.** The global excludes file
  `/Users/forrest/.gitignore` (set by `core.excludesFile` in `~/.gitconfig`) applies to this repo. It
  carries `*.out`, `*.log`, `*.zip`, `*.env`, `*.exe`, `*.d`, `*.[ao]`, `*.tmp`, `*.bak`,
  `notes.md`, `node_modules`, `.DS_Store` and others. `.git/info/exclude` also ignores several
  `**/.claude/...` runtime files. Checked at plan time: none of these hides any of the 111 real files
  that must be committed. The C1 real-tree sweep re-checks this at implementation time.
  - Because of the global file, checks against the repo's own .gitignore must run with
    `-c core.excludesFile=/dev/null`, so a clone on another machine (or the BC-17 zip) behaves the same.
  - Note for the orchestrator, not BC-1 scope: later tickets on this host will see `*.log`/`*.out`/`*.zip`
    ignored even where the repo does not ignore them. This is a candidate run-time footgun row.
- **`.claude/` at the root.** It is Claude Code runtime, and it is not in the staging list. If a
  non-excluded file appears in it before the commit, T2.1 would show `?? .claude/`. In that case stop
  and report to the delegator. Do not add a root ignore rule or edit `.git/info/exclude` without a
  decision.
- **Concurrent writes.** The live board writes `.lattice/` while you stage, and other agents share this
  checkout. The commit snapshots whatever is staged; later board writes show as modified `.lattice`
  files, which card item 4 allows. Never revert, stash or clean unexplained files. The step 0 snapshot
  check catches unexpected root entries.
- **macOS make is GNU Make 3.81 (2006)**, at `/usr/bin/make`. No `gmake` or `bmake` is installed.
  Error text uses 3.81 style (``make: *** [x] Error 1``, ``Target `check' not remade``). Tests match the
  `not yet implemented` lines exactly and treat make's own lines as informational, apart from T3.1/T3.4,
  which quote what 3.81 printed. The exit code is 2 (decision 3).
- **TAB vs spaces in recipes.** The Write tool or a model may emit spaces. Mitigation: the `<TAB>`
  marker plus the perl conversion, then the T3.7 counts (10 TAB lines, 0 lines starting with a space).
- **File mode 0600.** It is normalised to 0644 (decision 10). Git stores 100644 regardless.
  `.lattice/*` files are also 0600 locally, which is irrelevant to git and left alone.
- **Empty `.lattice` dirs are not tracked** (`archive/*`, `notes`, `resources`, `templates`,
  `artifacts/meta`, `artifacts/payload`, `sessions/archive`, `orchestration/evidence`): git does not
  track empty directories. That is harmless, because every lattice call uses
  `LATTICE_ROOT=/Users/forrest/Code/brillio`, where they still exist. Do not add `.gitkeep`s inside
  `.lattice/`.
- **PDF sizes** are 125 KB and 66 KB. That is fine for git. SPEC §12's dist exclusion of the PDFs is
  BC-17's concern.
- **Emoji.** Plan time found none in the files being committed (no 4-byte UTF-8 sequences and no
  U+2600-27BF characters in the root docs or the data). BC-1's own files and the commit message are
  checked ASCII-only by C2/T1.5.

## Contradictions and clarifications found

1. The card says each stub "exits 1". The recipe does, but `make` itself exits 2. The tests assert
   non-zero (2).
2. SPEC §3 lists 9 Makefile targets; the card adds 4 per-half targets. This is additive, so there is
   no SPEC edit (see Scope).
3. Ticket: "git status is clean afterward". Card item 4 narrows this to outside `.lattice/`; T2.1
   uses that definition.
4. Ticket: "Commit everything present". Card item 2 lists the set. `.claude/` (new since the ticket
   was written) is runtime and is excluded. The ticket's "docs/ (already has SPEC.md)" also holds
   `docs/branding-demo.html`, which is committed.
5. BRANDING §8 refers to `sample_listings.json` by bare filename. It is still accurate after the move,
   and BRANDING.md is not edited by builders.

## Size estimate

Authored: 116 lines across 4 files (Makefile 45, .gitignore 36, .editorconfig 26, README 9) plus
2 empty `.gitkeep`s and one file move. The commit itself is about 80-90 files: 14 non-board paths plus
65 or more `.lattice` board files. Most of its bytes are existing docs, PDFs and board state that BC-1
did not author.

## Plan-Review Cycle 1 Resolutions (AUTHORITATIVE — overrides earlier text on conflict)

Plan review artifact: art_01M4VZCP6DE9BHWCKJFNH2D1G2 (verdict PASS, 4 minor findings). All four triaged
as obvious and resolved here.

1. **`$SCRATCH` undefined** (Approach step 0, T1.2, T3.6, C5, Validation).
   Resolution: `SCRATCH` is the literal path `/tmp/bc1-scratch`. Step 0 adds `mkdir -p /tmp/bc1-scratch`.
   Every Bash call that uses it begins with `SCRATCH=/tmp/bc1-scratch;` (Bash calls do not share env).
   Step 8 adds `rm -rf /tmp/bc1-scratch`. Sections: Approach step 0 and 8, Test list T1.2/T3.6/C5.
2. **`/backend/api` in `.gitignore` would hide a future `backend/api/` package.**
   Resolution: drop the `/backend/api` line from `.gitignore` (and from the "Decisions" .gitignore
   summary). `/backend/tmp/` and `/backend/bin/` cover build output; BC-2 adds a rule for wherever it
   actually builds. The check-ignore "MUST be ignored" probe list (C1) drops `backend/api`, so the
   expected count is **26** (not 27); `.gitignore` line count drops by one. Sections: Files/.gitignore,
   Decisions, Test list C1, Size estimate.
3. **Inconsistent file counts (111 vs 80-90).** Resolution: the Risks sentence reads "none of these hides
   any file in the tree that must be committed (79 at plan time)". The Size estimate's ~80-90 stands;
   the implementer reports the real count from `git ls-tree -r --name-only HEAD | wc -l`. Section: Risks.
4. **`wc -l` is space-padded on macOS.** Resolution: step 0 uses `git rev-list --all --count` (expects `0`);
   T1.1 uses `git rev-list --count HEAD` (expects `1`) as the authoritative check; any remaining
   `wc -l` comparison is piped through `tr -d ' '`. Sections: Approach step 0, T1.1, T2.1.
